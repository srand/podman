package images

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/containers/common/pkg/completion"
	"github.com/containers/podman/v4/cmd/podman/common"
	"github.com/containers/podman/v4/cmd/podman/parse"
	"github.com/containers/podman/v4/cmd/podman/registry"
	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	containerConfig = registry.PodmanConfig()
)

var (
	saveDescription = `Save an image to docker-archive or oci-archive on the local machine. Default is docker-archive.`

	saveCommand = &cobra.Command{
		Use:   "save [options] IMAGE [IMAGE...]",
		Short: "Save image(s) to an archive",
		Long:  saveDescription,
		RunE:  save,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("need at least 1 argument")
			}
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}
			if !util.StringInSlice(format, common.ValidSaveFormats) {
				return fmt.Errorf("format value must be one of %s", strings.Join(common.ValidSaveFormats, " "))
			}
			return nil
		},
		ValidArgsFunction: common.AutocompleteImages,
		Example: `podman save --quiet -o myimage.tar imageID
  podman save --format docker-dir -o ubuntu-dir ubuntu
  podman save > alpine-all.tar alpine:latest`,
	}

	imageSaveCommand = &cobra.Command{
		Args:              saveCommand.Args,
		Use:               saveCommand.Use,
		Short:             saveCommand.Short,
		Long:              saveCommand.Long,
		RunE:              saveCommand.RunE,
		ValidArgsFunction: saveCommand.ValidArgsFunction,
		Example: `podman image save --quiet -o myimage.tar imageID
  podman image save --format docker-dir -o ubuntu-dir ubuntu
  podman image save > alpine-all.tar alpine:latest`,
	}
)

// saveOptionsWrapper wraps entities.ImageSaveOptions and prevents leaking
// CLI-only fields into the API types.
type saveOptionsWrapper struct {
	entities.ImageSaveOptions
	EncryptionKeys []string
	EncryptLayers  []int
}

var (
	saveOpts saveOptionsWrapper
)

func init() {
	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: saveCommand,
	})
	saveFlags(saveCommand)

	registry.Commands = append(registry.Commands, registry.CliCommand{
		Command: imageSaveCommand,
		Parent:  imageCmd,
	})
	saveFlags(imageSaveCommand)
}

func saveFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.BoolVar(&saveOpts.Compress, "compress", false, "Compress tarball image layers when saving to a directory using the 'dir' transport. (default is same compression type as source)")

	flags.BoolVar(&saveOpts.OciAcceptUncompressedLayers, "uncompressed", false, "Accept uncompressed layers when copying OCI images")

	formatFlagName := "format"
	flags.StringVar(&saveOpts.Format, formatFlagName, define.V2s2Archive, "Save image to oci-archive, oci-dir (directory with oci manifest type), docker-archive, docker-dir (directory with v2s2 manifest type)")
	_ = cmd.RegisterFlagCompletionFunc(formatFlagName, common.AutocompleteImageSaveFormat)

	outputFlagName := "output"
	flags.StringVarP(&saveOpts.Output, outputFlagName, "o", "", "Write to a specified file (default: stdout, which must be redirected)")
	_ = cmd.RegisterFlagCompletionFunc(outputFlagName, completion.AutocompleteDefault)

	flags.BoolVarP(&saveOpts.Quiet, "quiet", "q", false, "Suppress the output")
	flags.BoolVarP(&saveOpts.MultiImageArchive, "multi-image-archive", "m", containerConfig.ContainersConfDefaultsRO.Engine.MultiImageArchive, "Interpret additional arguments as images not tags and create a multi-image-archive (only for docker-archive)")

	if !registry.IsRemote() {
		flags.StringVar(&saveOpts.SignaturePolicy, "signature-policy", "", "Path to a signature-policy file")
		_ = flags.MarkHidden("signature-policy")
	}

	encryptionKeysFlagName := "encryption-key"
	flags.StringSliceVar(&saveOpts.EncryptionKeys, encryptionKeysFlagName, nil, "Key with the encryption protocol to use to encrypt the image (e.g. jwe:/path/to/key.pem)")
	_ = cmd.RegisterFlagCompletionFunc(encryptionKeysFlagName, completion.AutocompleteDefault)

	encryptLayersFlagName := "encrypt-layer"
	flags.IntSliceVar(&saveOpts.EncryptLayers, encryptLayersFlagName, nil, "Layers to encrypt, 0-indexed layer indices with support for negative indexing (e.g. 0 is the first layer, -1 is the last layer). If not defined, will encrypt all layers if encryption-key flag is specified")
	_ = cmd.RegisterFlagCompletionFunc(encryptLayersFlagName, completion.AutocompleteDefault)

}

func save(cmd *cobra.Command, args []string) (finalErr error) {
	var (
		tags      []string
		succeeded = false
	)
	if cmd.Flag("compress").Changed && saveOpts.Format != define.V2s2ManifestDir {
		return errors.New("--compress can only be set when --format is 'docker-dir'")
	}
	if len(saveOpts.Output) == 0 {
		saveOpts.Quiet = true
		fi := os.Stdout
		if term.IsTerminal(int(fi.Fd())) {
			return errors.New("refusing to save to terminal. Use -o flag or redirect")
		}
		pipePath, cleanup, err := setupPipe()
		if err != nil {
			return err
		}
		if cleanup != nil {
			defer func() {
				errc := cleanup()
				if succeeded {
					writeErr := <-errc
					if writeErr != nil && finalErr == nil {
						finalErr = writeErr
					}
				}
			}()
		}
		saveOpts.Output = pipePath
	}
	if err := parse.ValidateFileName(saveOpts.Output); err != nil {
		return err
	}
	if len(args) > 1 {
		tags = args[1:]
	}

	encConfig, encLayers, err := util.EncryptConfig(saveOpts.EncryptionKeys, saveOpts.EncryptLayers)
	if err != nil {
		return fmt.Errorf("unable to obtain encryption config: %w", err)
	}
	saveOpts.OciEncryptConfig = encConfig
	saveOpts.OciEncryptLayers = encLayers

	err = registry.ImageEngine().Save(context.Background(), args[0], tags, saveOpts.ImageSaveOptions)
	if err == nil {
		succeeded = true
	}
	return err
}

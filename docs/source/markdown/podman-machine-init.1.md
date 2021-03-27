% podman-machine-init(1)

## NAME
podman\-machine\-init - Initialize a new virtual machine

## SYNOPSIS
**podman machine init** [*options*] [*name*]

## DESCRIPTION

Initialize a new virtual machine for Podman.

Podman on MacOS requires a virtual machine. This is because containers are Linux -
containers do not run on any other OS because containers' core functionality are
tied to the Linux kernel.

**podman machine init** initializes a new Linux virtual machine where containers are run.

## OPTIONS

#### **--cpus**=*number*

Number of CPUs.

#### **--ignition-path**

Fully qualified path of the ignition file

#### **--image-path**

Fully qualified path of the uncompressed image file

#### **--memory**, **-m**=*number*

Memory (in MB).

#### **--help**

Print usage statement.

## EXAMPLES

```
$ podman machine init myvm
$ podman machine init --device=/dev/xvdc:rw myvm
$ podman machine init --memory=1024 myvm
```

## SEE ALSO
podman-machine (1)

## HISTORY
March 2021, Originally compiled by Ashley Cui <acui@redhat.com>
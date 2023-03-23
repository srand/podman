####> This option file is used in:
####>   podman push, save
####> If file is edited, make sure the changes
####> are applicable to all of those.
#### **--encryption-key**=*key*

The [protocol:keyfile] specifies the encryption protocol, which can be JWE (RFC7516), PGP (RFC4880), and PKCS7 (RFC2315) and the key material required for image encryption. For instance, jwe:/path/to/key.pem or pgp:admin@example.com or pkcs7:/path/to/x509-file.

####> This option file is used in:
####>   podman push, save
####> If file is edited, make sure the changes
####> are applicable to all of those.
#### **--encrypt-layer**=*layer(s)*

Layer(s) to encrypt: 0-indexed layer indices with support for negative indexing (e.g. 0 is the first layer, -1 is the last layer). If not defined, will encrypt all layers if encryption-key flag is specified.

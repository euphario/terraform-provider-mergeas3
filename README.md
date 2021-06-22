# Terraform Provider Merge AS3

Run the following command to build the provider

```shell
make
```

The purpose of this provider is to assemble fragments of AS3 declarations, the provider is given a root folder and will procede to scan all subfolders.

The structure of the folder looks like this;
<root folder> / <bigip device> / <tenant>

The final output can be checked against a schema.

## Test sample configuration

First, build and install the provider.

```shell
make install
```

Then, run the following command to initialize the workspace and apply the sample configuration.

```shell
terraform init && terraform apply
```
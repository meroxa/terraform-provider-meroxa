# terraform-provider-meroxa

Terraform provider for the Meroxa Platform

- Website: https://www.terraform.io

[![Support Server](https://img.shields.io/discord/828680256877363200.svg?label=Meroxa%20Community&logo=Discord&colorB=7289da&style=for-the-badge)](https://discord.meroxa.com)

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) v1.0.0
-	[Go](https://golang.org/doc/install) 1.18 (to build the provider plugin)


## Build provider

Run the following command to build the provider

```shell
$ go build -o terraform-provider-meroxa
```

## Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory.

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```

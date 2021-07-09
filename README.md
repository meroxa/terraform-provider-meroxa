# terraform-provider-meroxa
Terraform provider for the Meroxa Platform

- Website: https://www.terraform.io
<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">
  
[![Support Server](https://img.shields.io/discord/591914197219016707.svg?label=Meroxa%20Community&logo=Discord&colorB=7289da&style=for-the-badge)](https://discord.meroxa.com/channels/828680256877363200/828680256877363206)

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) v1.0.0
-	[Go](https://golang.org/doc/install) 1.16 (to build the provider plugin)


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

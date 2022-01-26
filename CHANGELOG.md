## v1.5.0

ENHANCEMENTS:

* Updating for latest meroxa-go interface; no longer exposing Connector and Pipeline metadata. ([#26]https://github.com/meroxa/terraform-provider-meroxa/issues/26))

BUG FIXES:

* Add connector name validation ([#29](https://github.com/meroxa/terraform-provider-meroxa/issues/29))

## v1.4.0

ENHANCEMENTS:

* Add support for optional `refresh_token` to help renew an expired `access_token`. ([#21](https://github.com/meroxa/terraform-provider-meroxa/issues/21))

BUG FIXES:

* Fix `config` attribute handling in the connector resource ([#21](https://github.com/meroxa/terraform-provider-meroxa/issues/21))

## v1.3.0

ENHANCEMENTS:

* Enable state monitoring for connector and resource resources.
Resource creation will error after 10 minutes timeout if the desired
state has not been reached. ([#20](https://github.com/meroxa/terraform-provider-meroxa/issues/20))

## v1.2.0

BUG FIXES:

* Ensure `private_key` attribute is submitted ([#17](https://github.com/meroxa/terraform-provider-meroxa/issues/17))

## v1.1.0

ENHANCEMENTS:

* Add new `private_key` attribute to `meroxa_resource.ssh_tunnel`.
Make attribute `address` on `meroxa_resource.ssh_tunnel` required. ([#11](https://github.com/meroxa/terraform-provider-meroxa/issues/11))

## v0.1.0

FEATURES:

* **New Resource:** `meroxa_connector`
* **New Resource:** `meroxa_endpoint`
* **New Resource:** `meroxa_pipeline`
* **New Resource:** `meroxa_resource`


* **New Data Source:** `meroxa_connector`
* **New Data Source:** `meroxa_endpoint`
* **New Data Source:** `meroxa_pipeline`
* **New Data Source:** `meroxa_resource`
* **New Data Source:** `meroxa_resource_types`
* **New Data Source:** `meroxa_transforms`

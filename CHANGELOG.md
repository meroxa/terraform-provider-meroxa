## 1.3.0

ENHANCEMENTS:

* Enable state monitoring for connector and resource resources.
Resource creation will error after 10 minutes timeout if the desired
state has not been reached. ([#20](https://github.com/meroxa/terraform-provider-meroxa/issues/20))

## 1.2.0

BUG FIXES:

* Ensure `private_key` attribute is submitted ([#17](https://github.com/meroxa/terraform-provider-meroxa/issues/17))

## 1.1.0

ENHANCEMENTS:

* Add new `private_key` attribute to `meroxa_resource.ssh_tunnel`.
Make attribute `address` on `meroxa_resource.ssh_tunnel` required. ([#11](https://github.com/meroxa/terraform-provider-meroxa/issues/11))

## 0.1.0

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

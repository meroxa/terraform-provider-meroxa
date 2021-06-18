terraform {
  required_providers {
    meroxa = {
      version = "0.1"
      source  = "meroxa.io/dev/meroxa"
    }
  }
}

provider "meroxa" {
  access_token=""
  refresh_token=""
  api_url=""
}

data "meroxa_resource_types" "all" {}

resource "meroxa_resource" "without_creds" {
  name = "test4"
  type = "postgres"
  url = "postgres://foo:bar@example:5432/example"
}

resource "meroxa_resource" "with_creds" {
  name = "test4"
  type = "postgres"
  url = "postgres://test1.cm5kbe8ybs1j.us-east-1.rds.amazonaws.com:5432/meroxa"
  credentials {
    username = ""
    password = ""
  }
}

//resource "meroxa_resource" "import_test" {}

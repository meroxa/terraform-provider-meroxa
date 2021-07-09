resource "meroxa_resource" "inline" {
  name = "inline"
  type = "postgres"
  url  = "postgres://foo:bar@example:5432/db"
}

resource "meroxa_resource" "credential_block" {
  name = "credential-block"
  type = "postgres"
  url  = "postgres://example:5432/db"
  credentials {
    username = "foo"
    password = "bar"
  }
}

resource "meroxa_resource" "inline" {
  name = "example"
  type = "postgres"
  url  = "postgres://foo:bar@example:5432/db"
}

resource "meroxa_connector" "basic" {
  name      = "http"
  source_id = meroxa_resource.inline.id
  input     = "public"
}

resource "meroxa_endpoint" "http" {
  name     = "myendpoint"
  protocol = "HTTP"
  stream   = meroxa_connector.basic.streams[0].output[0]
}

resource "meroxa_resource" "inline" {
  name = "inline"
  type = "postgres"
  url  = "postgres://foo:bar@example:5432/db"
}

resource "meroxa_pipeline" "basic" {
  name = "pipeline"
}

resource "meroxa_connector" "basic" {
  name        = "basic"
  source_id   = meroxa_resource.inline.id
  input       = "public.Users"
  pipeline_id = meroxa_pipeline.basic.id
}

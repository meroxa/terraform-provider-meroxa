# Configure the Meroxa provider
provider "meroxa" {
  access_token = var.access_token # optionally use MEROXA_ACCESS_TOKEN env var
  api_url      = var.api_url      # optionally use MEROXA_API_URL env var
  timeout      = var.timeout      # optionally use MEROXA_TIMEOUT env var

  # To enable debug
  debug = false # optionally use MEROXA_DEBUG env var
}

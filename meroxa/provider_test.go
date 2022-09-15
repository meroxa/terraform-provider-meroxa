package meroxa

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// testAccProviderFactories is a static map containing only the main provider instance.
var testAccProviderFactories map[string]func() (*schema.Provider, error)

// testAccProvider is the "main" provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheck(t) must be called before using this provider instance.
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider("testacc")()

	// Always allocate a new provider instance each invocation, otherwise gRPC
	// ProviderConfigure() can overwrite configuration during concurrent testing.
	//nolint:unparam
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"meroxa": func() (*schema.Provider, error) {
			return Provider("testacc")(), nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	if err := os.Getenv("MEROXA_ACCESS_TOKEN"); err == "" {
		t.Fatal("MEROXA_ACCESS_TOKEN must be set for acceptance tests")
	}

	if err := os.Getenv("MEROXA_API_URL"); err == "" {
		t.Fatal("MEROXA_API_URL must be set for acceptance tests")
	}

	if err := os.Getenv("MEROXA_POSTGRES_URL"); err == "" {
		t.Fatal("MEROXA_POSTGRES_URL must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

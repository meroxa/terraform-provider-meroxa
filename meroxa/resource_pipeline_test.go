package meroxa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/meroxa/meroxa-go"
	"os"
	"strings"
	"testing"
)

func init() {
	driver, rest := splitUrlSchema(os.Getenv("MEROXA_POSTGRES_URL"))
	_, base := splitUrlCreds(rest)
	postgresqlUrl = strings.Join([]string{driver, base}, "")
}

func TestAccMeroxaPipeline_basic(t *testing.T) {
	testAccMeroxaPipelineBasic := `
	resource "meroxa_pipeline" "basic" {
	  name = "basic"
	}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaPipelineBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_pipeline.basic"),
					resource.TestCheckResourceAttr("meroxa_pipeline.basic", "name", "basic"),
				),
			},
		},
	})
}

func testAccCheckMeroxaPipelineDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*meroxa.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "meroxa_pipeline" {
			continue
		}

		pipelineName := rs.Primary.Attributes["name"]

		r, err := c.GetPipelineByName(context.Background(), pipelineName)
		if err == nil && r != nil {
			return fmt.Errorf("connector still exists")
		}
	}
	return nil
}

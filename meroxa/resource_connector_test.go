package meroxa

import (
	"context"
	"fmt"
	"github.com/meroxa/meroxa-go"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	driver, rest := splitUrlSchema(os.Getenv("MEROXA_POSTGRES_URL"))
	_, base := splitUrlCreds(rest)
	postgresqlUrl = strings.Join([]string{driver, base}, "")
}

func TestAccMeroxaConnector_basic(t *testing.T) {
	testAccMeroxaConnectionBasic := fmt.Sprintf(`
	resource "meroxa_resource" "inline" {
	  name = "connector_inline"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_connector" "basic" {
		name = "basic"
        source_id = meroxa_resource.inline.id
        input = "public"
	}
	`, os.Getenv("MEROXA_POSTGRES_URL"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaConnectionBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_connector.basic"),
					resource.TestCheckResourceAttr("meroxa_connector.basic", "name", "basic"),
					resource.TestCheckResourceAttr("meroxa_connector.basic", "type", "jdbc-source"),
				),
			},
		},
	})
}

func testAccCheckMeroxaConnectorDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*meroxa.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "meroxa_connector" {
			continue
		}

		connectorID := rs.Primary.ID
		rID, err := strconv.Atoi(connectorID)
		if err != nil {
			return err
		}

		r, err := c.GetConnector(context.Background(), rID)
		if err == nil && r != nil {
			return fmt.Errorf("connector still exists")
		}
	}

	return nil
}

package meroxa

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/meroxa/meroxa-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMeroxaConnector_basic(t *testing.T) {
	testAccMeroxaConnectionBasic := fmt.Sprintf(`
	resource "meroxa_resource" "inline" {
	  name = "connector-inline"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_connector" "basic" {
		name = "connector-basic"
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
					resource.TestCheckResourceAttr("meroxa_connector.basic", "name", "connector-basic"),
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

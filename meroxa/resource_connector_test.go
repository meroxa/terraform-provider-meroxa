package meroxa

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/meroxa/meroxa-go/pkg/meroxa"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccMeroxaConnector_basic(t *testing.T) {
	testAccMeroxaConnectionBasic := fmt.Sprintf(`
	resource "meroxa_resource" "connector_test" {
	  name = "connector-inline"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_pipeline" "connector_test" {
	  name = "connector-test"
	}
	resource "meroxa_connector" "basic" {
		name = "connector-basic"
		pipeline_id = meroxa_pipeline.connector_test.id
        source_id = meroxa_resource.connector_test.id
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
					resource.TestCheckResourceAttr("meroxa_connector.basic", "state", "running"),
				),
			},
		},
	})
}

func TestAccMeroxaConnector_WithoutPipeline(t *testing.T) {
	testAccMeroxaConnectionBasic := fmt.Sprintf(`
	resource "meroxa_resource" "connector_test" {
	  name = "connector-inline"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_connector" "basic" {
		name = "connector-basic"
        source_id = meroxa_resource.connector_test.id
        input = "public"
	}
	`, os.Getenv("MEROXA_POSTGRES_URL"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccMeroxaConnectionBasic,
				ExpectError: regexp.MustCompile("The argument \"pipeline_id\" is required, but no definition was found."),
			},
		},
	})
}

func TestAccMeroxaConnector_WithConfig(t *testing.T) {
	testAccMeroxaConnectionWithConfig := func(k, v string) string {
		return fmt.Sprintf(`
			resource "meroxa_resource" "connector_test" {
	  			name = "connector-inline"
	  			type = "postgres"
	  			url = "%s"
			}
			resource "meroxa_pipeline" "connector_test" {
	  			name = "connector-test"
			}
			resource "meroxa_connector" "with_config" {
				name = "connector-basic"
				pipeline_id = meroxa_pipeline.connector_test.id
        		source_id = meroxa_resource.connector_test.id
        		input = "public"
        		config = {
        			%q = %q
        		}
			}
		`, os.Getenv("MEROXA_POSTGRES_URL"), k, v)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaConnectorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaConnectionWithConfig("key1", "val1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_connector.with_config"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "name", "connector-basic"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "type", "jdbc-source"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "state", "running"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "config.key1", "val1"),
				),
			},
			{
				Config:             testAccMeroxaConnectionWithConfig("key1", "val1"),
				Check:              testAccCheckMeroxaResourceExists("meroxa_connector.with_config"),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
			{
				Config: testAccMeroxaConnectionWithConfig("key1", "val2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_connector.with_config"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "name", "connector-basic"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "type", "jdbc-source"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "state", "running"),
					resource.TestCheckResourceAttr("meroxa_connector.with_config", "config.key1", "val2"),
				),
			},
		},
	})
}

func testAccCheckMeroxaConnectorDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(meroxa.Client)

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

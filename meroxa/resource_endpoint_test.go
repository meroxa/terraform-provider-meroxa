package meroxa

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/meroxa/meroxa-go"
)

func TestAccMeroxaEndpoint_http(t *testing.T) {
	testAccMeroxaEndpointBasic := fmt.Sprintf(`
	resource "meroxa_resource" "inline" {
	  name = "http-acceptance"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_connector" "basic" {
		name = "http-acceptance"
        source_id = meroxa_resource.inline.id
        input = "public"
	}
	resource "meroxa_endpoint" "http" {
		name = "http"
        protocol = "HTTP"
		stream = meroxa_connector.basic.streams[0].output[0]
	}
	`, os.Getenv("MEROXA_POSTGRES_URL"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaEndpointBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_endpoint.http"),
					resource.TestCheckResourceAttr("meroxa_endpoint.http", "name", "http"),
					resource.TestCheckResourceAttr("meroxa_endpoint.http", "protocol", "HTTP"),
				),
			},
		},
	})
}

func TestAccMeroxaEndpoint_grpc(t *testing.T) {
	testAccMeroxaEndpointBasic := fmt.Sprintf(`
	resource "meroxa_resource" "inline" {
	  name = "grpc-acceptance"
	  type = "postgres"
	  url = "%s"
	}
	resource "meroxa_connector" "basic" {
		name = "grpc-acceptance"
        source_id = meroxa_resource.inline.id
        input = "public"
	}
	resource "meroxa_endpoint" "grpc" {
		name = "grpc"
        protocol = "GRPC"
		stream = meroxa_connector.basic.streams[0].output[0]
	}
	`, os.Getenv("MEROXA_POSTGRES_URL"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaEndpointBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_endpoint.grpc"),
					resource.TestCheckResourceAttr("meroxa_endpoint.grpc", "name", "grpc"),
					resource.TestCheckResourceAttr("meroxa_endpoint.grpc", "protocol", "GRPC"),
				),
			},
		},
	})
}

func testAccCheckMeroxaEndpointDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*meroxa.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "meroxa_endpoint" {
			continue
		}

		eName := rs.Primary.ID

		r, err := c.GetEndpoint(context.Background(), eName)
		if err == nil && r != nil {
			return fmt.Errorf("endpoint still exists")
		}
	}
	return nil
}

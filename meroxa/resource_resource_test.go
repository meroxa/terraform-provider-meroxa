package meroxa

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/meroxa/meroxa-go"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var (
	postgresqlURL      string
	postgresqlUsername string
	postgresqlPassword string
)

func init() {
	driver, rest := splitURLSchema(os.Getenv("MEROXA_POSTGRES_URL"))
	creds, base := splitURLCreds(rest)
	postgresqlUsername = strings.Split(creds, ":")[0]
	postgresqlPassword = strings.Split(creds, ":")[1]
	postgresqlURL = strings.Join([]string{driver, base}, "")
}

func TestAccMeroxaResource_basic(t *testing.T) {
	testAccMeroxaResourceBasic := fmt.Sprintf(`
	resource "meroxa_resource" "basic" {
	  name = "resource-basic"
	  type = "postgres"
	  url = "%s"
	  credentials {
		username = "%s"
		password = "%s"
	  }
	}
	`, postgresqlURL, postgresqlUsername, postgresqlPassword)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaResourceBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_resource.basic"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "name", "resource-basic"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "type", "postgres"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "url", postgresqlURL),
				),
			},
		},
	})
}

func TestAccMeroxaResource_inline(t *testing.T) {
	testAccMeroxaResourceInline := fmt.Sprintf(`
	resource "meroxa_resource" "inline" {
	  name = "inline"
	  type = "postgres"
	  url = "%s"
	}
	`, os.Getenv("MEROXA_POSTGRES_URL"))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaResourceInline,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_resource.inline"),
					resource.TestCheckResourceAttr("meroxa_resource.inline", "name", "inline"),
					resource.TestCheckResourceAttr("meroxa_resource.inline", "type", "postgres"),
					resource.TestCheckResourceAttr("meroxa_resource.inline", "url", postgresqlURL),
				),
			},
		},
	})
}

func testAccCheckMeroxaResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*meroxa.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "meroxa_resource" {
			continue
		}

		resourceID := rs.Primary.ID
		rID, err := strconv.Atoi(resourceID)
		if err != nil {
			return err
		}

		r, err := c.GetResource(context.Background(), rID)
		if err == nil && r != nil {
			return fmt.Errorf("resource still exists")
		}
	}

	return nil
}

func testAccCheckMeroxaResourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ResourceID set")
		}

		return nil
	}
}

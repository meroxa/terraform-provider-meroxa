package meroxa

import (
	"context"
	"fmt"
	"github.com/meroxa/meroxa-go"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	postgresqlUsername = "user"
	postgresqlPassword = "password"
	postgresqlDatabase = "postgres"
)

var postgresqlUrl string

func init() {
	s := stripUrlSchema(os.Getenv("MEROXA_POSTGRES_URL"))
	postgresqlUrl = stripUrlSchema(s)
}

func TestAccMeroxaResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaResourceBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_resource.basic"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "name", "basic"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "type", "postgres"),
					resource.TestCheckResourceAttr("meroxa_resource.basic", "url", postgresqlUrl),
				),
			},
		},
	})
}

func TestAccMeroxaResource_inline(t *testing.T) {
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
					resource.TestCheckResourceAttr("meroxa_resource.inline", "url", postgresqlUrl),
				),
			},
		},
	})
}

var testAccMeroxaResourceBasic = fmt.Sprintf(`
resource "meroxa_resource" "basic" {
  name = "basic"
  type = "postgres"
  url = "%s"
  credentials {
    username = "%s"
    password = "%s"
  }
}
`, postgresqlUrl, postgresqlUsername, postgresqlPassword)

var testAccMeroxaResourceInline = fmt.Sprintf(`
resource "meroxa_resource" "inline" {
  name = "inline"
  type = "postgres"
  url = "postgres://%s:%s@%s"
}
`, postgresqlUsername, postgresqlPassword, postgresqlUrl)

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
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ResourceID set")
		}

		return nil
	}
}

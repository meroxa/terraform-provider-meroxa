package meroxa

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/meroxa/meroxa-go/pkg/meroxa"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/joeshaw/envdecode"
	"golang.org/x/crypto/ssh"
)

type TestsConfig struct {
	PostgresURL        string `env:"MEROXA_POSTGRES_URL,required"`
	BastionUser        string `env:"MEROXA_BASTION_USER,default=ec2-user"`
	BastionHost        string `env:"MEROXA_BASTION_HOST,required"`
	BastionKey         string `env:"MEROXA_BASTION_KEY,required"`
	PrivatePostgresURL string `env:"MEROXA_PRIVATE_POSTGRES_URL,required"`
}

var Config TestsConfig
var (
	postgresqlURL      string
	postgresqlUsername string
	postgresqlPassword string
)

func init() {
	err := envdecode.Decode(&Config)
	if err != nil {
		log.Fatal(err)
	}

	driver, rest := splitURLSchema(Config.PostgresURL)
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
					resource.TestCheckResourceAttr("meroxa_resource.basic", "status", "ready"),
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
	`, Config.PostgresURL)
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
					resource.TestCheckResourceAttr("meroxa_resource.inline", "status", "ready"),
				),
			},
		},
	})
}

func TestAccMeroxaResource_sshTunnel(t *testing.T) {
	bastionAddr := fmt.Sprintf("%s@%s:22", Config.BastionUser, Config.BastionHost)
	privatePostgresURL, err := URLWithoutCredentials(Config.PrivatePostgresURL)
	if err != nil {
		t.Error(err)
	}
	b, _ := pem.Decode([]byte(Config.BastionKey))
	if err != nil {
		t.Error(err)
	}
	privKey, err := x509.ParsePKCS1PrivateKey(b.Bytes)
	if err != nil {
		t.Error(err)
	}
	pubKey, err := ssh.NewPublicKey(privKey.Public())
	if err != nil {
		t.Error(err)
	}
	sshPubKey := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pubKey)), "\n")

	testAccMeroxaResourceSSHTunnel := fmt.Sprintf(
		`resource "meroxa_resource" "with_tunnel" {
	  		name = "with_ssh_tunnel"
	  		type = "postgres"
	  		url = %q
	  		ssh_tunnel {
	  			address = %q
	  			private_key = %s
	  		}
		}`,
		Config.PrivatePostgresURL,
		withSSHURL(bastionAddr),
		fmt.Sprintf("<<-EOT\n%s\nEOT\n", Config.BastionKey),
	)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMeroxaResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMeroxaResourceSSHTunnel,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMeroxaResourceExists("meroxa_resource.with_tunnel"),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "name", "with_ssh_tunnel"),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "type", "postgres"),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "url", privatePostgresURL),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "status", "ready"),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "ssh_tunnel.0.address", withSSHURL(bastionAddr)),
					resource.TestCheckResourceAttr("meroxa_resource.with_tunnel", "ssh_tunnel.0.public_key", sshPubKey),
				),
			},
			{
				Config:             testAccMeroxaResourceSSHTunnel,
				Check:              testAccCheckMeroxaResourceExists("meroxa_resource.with_tunnel"),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckMeroxaResourceDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(meroxa.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "meroxa_resource" {
			continue
		}

		resourceID := rs.Primary.ID
		rID, err := strconv.Atoi(resourceID)
		if err != nil {
			return err
		}

		r, err := c.GetResourceByNameOrID(context.Background(), fmt.Sprint(rID))
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

func URLWithoutCredentials(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	parsed.User = nil
	return parsed.String(), nil
}

func withSSHURL(addr string) string {
	return fmt.Sprintf("ssh://%s", addr)
}

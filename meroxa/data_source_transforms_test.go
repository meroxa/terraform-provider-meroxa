package meroxa

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataAPGroup_default(t *testing.T) {
	datasourceAddress := "data.meroxa_transforms.default"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataMeroxaTransforms,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceAddress, "transforms.%", "1"),
				),
			},
		},
	})
}

const testAccDataMeroxaTransforms = `
data "meroxa_transforms" "default" {}
`

package meroxa

import (
	"context"
	"github.com/meroxa/meroxa-go"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceResourceTypes() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceResourceTypesRead,
		Schema: map[string]*schema.Schema{
			"resource_types": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of support resource types",
				Computed:    true,
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Description: "Meroxa Resource Types",
				},
			},
		},
	}
}

func dataSourceResourceTypesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*meroxa.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	rTypes, err := c.ListResourceTypes(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("resource_types", rTypes); err != nil {
		return diag.FromErr(err)
	}
	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

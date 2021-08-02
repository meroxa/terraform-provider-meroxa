package meroxa

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
)

func dataSourceTransforms() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTransformsRead,
		Schema: map[string]*schema.Schema{
			"transforms": {
				Type:        schema.TypeList,
				Description: "List of Transforms",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Transform ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Transform Name",
						},
						"required": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Transform Required",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Transform Description",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Transform Type",
						},
						"properties": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Transform Properties",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Property Name",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Property Type",
									},
									"required": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Property Required",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceTransformsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*meroxa.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	transforms, err := c.ListTransforms(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("transforms", flattenTransform(transforms)); err != nil {
		return diag.FromErr(err)
	}
	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenTransform(transforms []*meroxa.Transform) []interface{} {
	if transforms != nil {
		tMap := make([]interface{}, len(transforms))
		for i, t := range transforms {
			ti := make(map[string]interface{})
			ti["id"] = t.ID
			ti["name"] = t.Name
			ti["required"] = t.Required
			ti["description"] = t.Description
			ti["type"] = t.Type
			ti["properties"] = flattenProperties(t.Properties)

			tMap[i] = ti
		}
		return tMap
	}
	return make([]interface{}, 0)
}

func flattenProperties(properties []meroxa.Property) []interface{} {
	if properties != nil {
		pMap := make([]interface{}, len(properties))
		for i, p := range properties {
			pi := make(map[string]interface{})
			pi["name"] = p.Name
			pi["required"] = p.Required
			pi["type"] = p.Type
			pMap[i] = pi
		}
		return pMap
	}
	return make([]interface{}, 0)
}

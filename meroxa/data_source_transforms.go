package meroxa

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"strconv"
	"time"
)

/*
 {
                "id": 27,
                "name": "Flatten",
                "required": false,
                "description": "Flatten a nested data structure, generating names for each field by concatenating the field names at each level with a configurable delimiter character. Applies to a Struct when a schema is present, or a Map in the case of schemaless data.",
                "type": "builtin",
                "properties": [
                        {
                                "name": "delimiter",
                                "required": false,
                                "type": "string"
                        }
                ]
        }

*/

func dataSourceTransforms() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceTransformsRead,
		Schema: map[string]*schema.Schema{
			"transforms": &schema.Schema{
				Type:        schema.TypeList,
				Description: "List of Transforms",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
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
							Elem:        schema.TypeMap,
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

	if err = d.Set("transforms", transforms); err != nil {
		return diag.FromErr(err)
	}
	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

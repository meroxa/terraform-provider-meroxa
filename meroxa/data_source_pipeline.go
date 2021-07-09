package meroxa

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"strconv"
)

func dataSourcePipeline() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePipelineRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Pipeline ID",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Pipeline name",
				Required:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Pipeline state",
				Computed:    true,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Description: "Pipeline metadata",
				Computed:    true,
			},
		},
	}
}

func dataSourcePipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var p *meroxa.Pipeline
	var err error

	c := m.(*meroxa.Client)

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		p, err = c.GetPipelineByName(ctx, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("id", strconv.Itoa(p.ID))
	_ = d.Set("name", p.Name)
	_ = d.Set("state", p.State)
	_ = d.Set("metadata", p.Metadata)

	return diags
}

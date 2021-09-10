package meroxa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
)

func dataSourceConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConnectorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Connector ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Connector Name",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Connector Type",
			},
			"streams": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Connector Streams",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dynamic": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"input": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     schema.TypeString,
						},
						"output": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     schema.TypeString,
						},
					},
				},
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Connector state",
			},
			"config": {
				Type:        schema.TypeMap,
				Description: "Connector configuration",
				Computed:    true,
				Elem:        schema.TypeString,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Description: "Connector metadata",
				Computed:    true,
				Elem:        schema.TypeString,
			},
			"pipeline_id": {
				Type:        schema.TypeInt,
				Description: "Connector's Pipeline ID",
				Computed:    true,
			},
			"pipeline_name": {
				Type:        schema.TypeString,
				Description: "Connector's Pipeline Name",
				Computed:    true,
			},
			"source_id": {
				Type:        schema.TypeString,
				Description: "The resource ID for a source connector",
				Computed:    true,
			},
			"destination_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource ID for a destination connector",
			},
		},
	}
}

func dataSourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var conn *meroxa.Connector
	var err error

	c := m.(*meroxa.Client)

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		conn, err = c.GetConnectorByName(ctx, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("id", strconv.Itoa(conn.ID))
	_ = d.Set("type", conn.Type)
	_ = d.Set("name", conn.Name)
	_ = d.Set("config", conn.Configuration)
	_ = d.Set("metadata", conn.Metadata)
	err = d.Set("streams", flattenStreams(conn))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting streams: %s", err))
	}
	_ = d.Set("state", conn.State)
	_ = d.Set("pipeline_id", conn.PipelineID)
	_ = d.Set("pipeline_name", conn.PipelineName)

	return diags
}

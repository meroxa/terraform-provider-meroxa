package meroxa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go/pkg/meroxa"
)

func dataSourceEndpoint() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceEndpointRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Endpoint ID - Matches Name",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Endpoint Name",
				Required:    true,
			},
			"protocol": {
				Type:        schema.TypeString,
				Description: "Protocol. Must be HTTP or GRPC",
				Computed:    true,
			},
			"stream": {
				Type:        schema.TypeString,
				Description: "The Endpoint's stream",
				Computed:    true,
			},
			"host": {
				Type:        schema.TypeString,
				Description: "The Endpoint's host",
				Computed:    true,
			},
			"ready": {
				Type:        schema.TypeBool,
				Description: "The Endpoint's ready state",
				Computed:    true,
			},
			"basic_auth_username": {
				Type:        schema.TypeString,
				Description: "Endpoint's username name credential",
				Computed:    true,
			},
			"basic_auth_password": {
				Type:        schema.TypeString,
				Description: "Endpoint's password credential",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	var e *meroxa.Endpoint
	var err error

	c := m.(meroxa.Client)

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		e, err = c.GetEndpoint(ctx, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("id", e.Name)
	_ = d.Set("protocol", string(e.Protocol))
	_ = d.Set("stream", e.Stream)
	_ = d.Set("ready", e.Ready)
	_ = d.Set("host", e.Host)
	_ = d.Set("basic_auth_username", e.BasicAuthUsername)
	_ = d.Set("basic_auth_password", e.BasicAuthPassword)

	return diags
}

package meroxa

import (
	"context"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go/pkg/meroxa"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEndpointCreate,
		ReadContext:   resourceEndpointRead,
		DeleteContext: resourceEndpointDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Endpoint Name",
				Required:    true,
				ForceNew:    true,
			},
			"protocol": {
				Type:        schema.TypeString,
				Description: "Protocol. Must be HTTP or GRPC",
				Optional:    true,
				ForceNew:    true,
				ValidateDiagFunc: func(val interface{}, path cty.Path) diag.Diagnostics {
					var diags diag.Diagnostics
					protocol := val.(string)
					switch protocol {
					case "HTTP", "GRPC":
					default:
						diags = append(diags, diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Protocol not supported",
							Detail:   "Please use \"HTTP\" or \"GRPC\"",
						})
					}
					return diags
				},
			},
			"stream": {
				Type:        schema.TypeString,
				Description: "The Endpoint's stream",
				Required:    true,
				ForceNew:    true,
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(meroxa.Client)
	name := d.Get("name").(string)
	input := &meroxa.CreateEndpointInput{
		Name:     name,
		Protocol: meroxa.EndpointProtocol(d.Get("protocol").(string)),
		Stream:   d.Get("stream").(string),
	}
	err := c.CreateEndpoint(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return resourceEndpointRead(ctx, d, m)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(meroxa.Client)

	name := d.Get("name").(string)

	e, err := c.GetEndpoint(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("protocol", string(e.Protocol))
	_ = d.Set("stream", e.Stream)
	_ = d.Set("ready", e.Ready)
	_ = d.Set("host", e.Host)
	_ = d.Set("basic_auth_username", e.BasicAuthUsername)
	_ = d.Set("basic_auth_password", e.BasicAuthPassword)

	return diags
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(meroxa.Client)
	eName := d.Id()

	err := c.DeleteEndpoint(ctx, eName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

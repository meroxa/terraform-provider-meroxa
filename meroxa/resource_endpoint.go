package meroxa

import (
	"context"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
)

func resourceEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEndpointCreate,
		ReadContext:   resourceEndpointRead,
		DeleteContext: resourceEndpointDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ready": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"basic_auth_username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"basic_auth_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceEndpointCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*meroxa.Client)
	name := d.Get("name").(string)
	protocol := d.Get("protocol").(string)
	stream := d.Get("stream").(string)
	err := c.CreateEndpoint(ctx, name, protocol, stream)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return resourceEndpointRead(ctx, d, m)
}

func resourceEndpointRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	name := d.Get("name").(string)

	e, err := c.GetEndpoint(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("protocol", e.Protocol)
	d.Set("stream", e.Stream)
	d.Set("ready", e.Ready)
	d.Set("host", e.Host)
	d.Set("basic_auth_username", e.BasicAuthUsername)
	d.Set("basic_auth_password", e.BasicAuthPassword)

	return diags
}

func resourceEndpointDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(*meroxa.Client)
	eName := d.Id()

	err := c.DeleteEndpoint(ctx, eName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

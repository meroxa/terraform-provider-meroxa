package meroxa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
)

func dataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceResourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Resource ID",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Resource Name. (Required)",
				Required:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Resource Type. Must be one of the supported resource types.",
				Computed:    true,
			},
			"url": {
				Type:        schema.TypeString,
				Description: "Resource URL. Warning will be thrown if credentials are placed inline. Using the credentials block is highly encouraged",
				Computed:    true,
				Sensitive:   false,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Description: "Resource Metadata",
				Computed:    true,
			},
			"ssh_tunnel": {
				Type:        schema.TypeList,
				Description: "Resource SSH tunnel configuration",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": { // todo fix state in API
				Type:        schema.TypeString,
				Description: "Resource Status",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: "Resource Created at timestamp",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Resource Updated at timestamp",
				Computed:    true,
			},
			"credentials": {
				Type:        schema.TypeList,
				Description: "Resource Credentials configuration",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:         schema.TypeString,
							Description:  "Resource username",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    false,
							Computed:     true,
						},
						"password": {
							Type:         schema.TypeString,
							Description:  "Resource password",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    true,
							Computed:     true,
						},
						"cacert": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource CACert. Trusted certificates for verifying resource",
						},
						"clientcert": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource Client Cert. Certificate for authenticating to the resource",
						},
						"clientkey": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Resource Client key. private key for authenticating to the resource",
							Sensitive:   true,
						},
						"ssl": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Resource SSL. Set Resource SSL option",
						},
					},
				},
			},
		},
	}
}

func dataSourceResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*meroxa.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var r *meroxa.Resource
	var err error

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		r, err = c.GetResourceByName(ctx, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	_ = d.Set("id", r.ID)
	_ = d.Set("name", r.Name)
	_ = d.Set("type", r.Type)
	_ = d.Set("url", r.URL)
	_ = d.Set("metadata", r.Metadata)
	_ = d.Set("status", r.Status.State) //todo flatten
	_ = d.Set("created_at", r.CreatedAt.String())
	_ = d.Set("updated_at", r.UpdatedAt.String())

	err = d.Set("credentials", flattenCredentials(r.Credentials))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting credentials: %s", err))
	}

	err = d.Set("ssh_tunnel", flattenSSHTunnel(r.SSHTunnel))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting ssh tunnel: %s", err))
	}

	// always run
	d.SetId(strconv.Itoa(r.ID))
	return diags
}

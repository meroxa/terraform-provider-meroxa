package meroxa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"strconv"
)

func dataSourceResource() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceResourceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Optional: true,
				Computed: true,
				Type:     schema.TypeString,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": {
				Type:        schema.TypeString,
				Description: "Resource URL",
				Computed:    true,
				Sensitive:   false, //if we contain secrets
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
			},
			"ssh_tunnel": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"status": &schema.Schema{ // todo fix state in API
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "resource username",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    false,
							Computed:     true,
						},
						"password": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "resource password",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    true,
							Computed:     true,
						},
						"cacert": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "trusted certificates for verifying resource",
						},
						"clientcert": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "client certificate for authenticating to the resource",
						},
						"clientkey": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "client private key for authenticating to the resource",
							Sensitive:   true,
						},
						"ssl": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "use SSL",
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
	if v, ok := d.GetOk("id"); ok && v.(string) != "" {
		id, err := strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		r, err = c.GetResource(ctx, id)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if v, ok := d.GetOk("name"); ok && v.(string) != "" {
		r, err = c.GetResourceByName(ctx, v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.Set("id", r.ID)
	d.Set("name", r.Name)
	d.Set("type", r.Type)
	d.Set("url", r.URL)
	d.Set("metadata", r.Metadata)
	d.Set("status", r.Status.State) //todo flatten
	d.Set("created_at", r.CreatedAt.String())
	d.Set("updated_at", r.UpdatedAt.String())

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

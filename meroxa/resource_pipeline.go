package meroxa

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go/pkg/meroxa"
)

func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePipelineCreate,
		ReadContext:   resourcePipelineRead,
		UpdateContext: resourcePipelineUpdate,
		DeleteContext: resourcePipelineDelete,
		Schema: map[string]*schema.Schema{
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
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	c := m.(meroxa.Client)
	pipeline := &meroxa.CreatePipelineInput{
		Name: d.Get("name").(string),
	}

	p, err := c.CreatePipeline(ctx, pipeline)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(p.ID))
	resourcePipelineRead(ctx, d, m)

	return diags
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(meroxa.Client)

	name := d.Get("name").(string)

	p, err := c.GetPipelineByName(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("state", p.State)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(meroxa.Client)
	input := &meroxa.UpdatePipelineInput{
		Name: d.Get("name").(string),
	}

	dID := d.Id()
	pID, err := strconv.Atoi(dID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = c.UpdatePipeline(ctx, pID, input)
	if err != nil {
		return diag.FromErr(err)
	}

	resourcePipelineRead(ctx, d, m)

	return diags
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(meroxa.Client)
	dID := d.Id()
	pID, err := strconv.Atoi(dID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeletePipeline(ctx, pID)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

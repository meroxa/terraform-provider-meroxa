package meroxa

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"strconv"
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
				Description: "The pipeline's name",
				Required:    true,
				ForceNew:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The pipeline's state",
				Computed:    true,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Description: "The pipeline's metadata",
				Optional:    true,
				Computed:    true,
				Elem:        schema.TypeString,
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

	c := m.(*meroxa.Client)
	pipeline := &meroxa.Pipeline{
		Name: d.Get("name").(string),
	}

	if v, ok := d.GetOk("metadata"); ok {
		pipeline.Metadata = v.(map[string]interface{})
	} else {
		meta := make(map[string]interface{})
		pipeline.Metadata = meta
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

	c := m.(*meroxa.Client)

	name := d.Get("name").(string)

	p, err := c.GetPipelineByName(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("metadata", p.Metadata)
	d.Set("state", p.State)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)
	input := meroxa.UpdatePipelineInput{
		Name: d.Get("name").(string),
	}

	dID := d.Id()
	pID, err := strconv.Atoi(dID)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("metadata") {
		input.Metadata = d.Get("metadata").(map[string]interface{})
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
	c := m.(*meroxa.Client)
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

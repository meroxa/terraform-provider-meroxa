package meroxa

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Connector Name",
			},
			"input": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Input stream",
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
				Optional:    true,
				Elem:        schema.TypeString,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Description: "Connector metadata",
				Optional:    true,
				Computed:    true,
				Elem:        schema.TypeString,
			},
			"pipeline_id": {
				Type:         schema.TypeInt,
				Description:  "Connector's Pipeline ID",
				Optional: 	  false,
				ExactlyOneOf: []string{"pipeline_name", "pipeline_id"},
			},
			"pipeline_name": {
				Type:         schema.TypeString,
				Description:  "Connector's Pipeline Name",
				Optional: 	  false,
			},
			"source_id": {
				Type:          schema.TypeString,
				Description:   "The resource ID for a source connector",
				Optional:      true,
				ConflictsWith: []string{"destination_id"},
			},
			"destination_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The resource ID for a destination connector",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var resourceID int
	var err error

	c := m.(*meroxa.Client)
	input := meroxa.CreateConnectorInput{
		Name:       d.Get("name").(string),
		ResourceID: resourceID,
	}
	if v, ok := d.GetOk("pipeline_id"); ok {
		input.PipelineID = v.(int)
	}

	if v, ok := d.GetOk("pipeline_name"); ok {
		input.PipelineName = v.(string)
	}

	if v, ok := d.GetOk("config"); ok {
		input.Configuration = v.(map[string]interface{})
	}

	if v, ok := d.GetOk("input"); ok {
		if input.Configuration == nil {
			input.Configuration = make(map[string]interface{})
		}
		input.Configuration["input"] = v.(string)
	}

	if v, ok := d.GetOk("metadata"); ok {
		input.Metadata = v.(map[string]interface{})
	} else {
		meta := make(map[string]interface{})
		input.Metadata = meta
	}

	if v, ok := d.GetOk("source_id"); ok && v.(string) != "" {
		resourceID, err = strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Metadata["mx:connectorType"] = "source"
	}

	if v, ok := d.GetOk("destination_id"); ok && v.(string) != "" {
		resourceID, err = strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Metadata["mx:connectorType"] = "destination"
	}
	input.ResourceID = resourceID

	conn, err := c.CreateConnector(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(conn.ID))
	resourceConnectorRead(ctx, d, m)

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	cID := d.Id()
	id, err := strconv.Atoi(cID)
	if err != nil {
		return diag.FromErr(err)
	}

	conn, err := c.GetConnector(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

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

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var state string
	c := m.(*meroxa.Client)

	name := d.Get("name").(string)
	if d.HasChange("state") {
		state = d.Get("state").(string)
		_, err := c.UpdateConnectorStatus(ctx, name, state)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	resourceResourceRead(ctx, d, m)

	return diags
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(*meroxa.Client)
	rID := d.Id()
	id, err := strconv.Atoi(rID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteConnector(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func flattenStreams(conn *meroxa.Connector) []interface{} {
	s := make(map[string]interface{})
	s["dynamic"] = conn.Streams["dynamic"].(bool)
	s["output"] = conn.Streams["output"]
	s["input"] = conn.Streams["input"]
	return []interface{}{s}
}

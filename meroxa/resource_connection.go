package meroxa

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"strconv"
)

func resourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, //TODO check if we can change name
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"streams": {
				Type:     schema.TypeList,
				Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},
			//"input": {
			//	Type:        schema.TypeList,
			//	Required:    true,
			//	Description: "comma delimited list of input streams",
			//	Elem: &schema.Schema{
			//		Type: schema.TypeString,
			//	},
			//},
			"config": {
				Type:        schema.TypeMap,
				Description: "connector configuration",
				Optional:    true,
				Elem:        schema.TypeString,
			},
			"metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
			"pipeline_id": &schema.Schema{
				Type:        schema.TypeInt,
				Description: "pipeline id to attach connector to",
				Optional:    true,
			},
			"pipeline_name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "pipeline name connector is attach to",
				Computed:    true,
			},
			"source_id": &schema.Schema{ // todo fix state in API
				Type:          schema.TypeString,
				Description:   "resource id to use as source",
				Optional:      true,
				ConflictsWith: []string{"destination_id"},
			},
			"destination_id": &schema.Schema{ // todo fix state in API
				Type:        schema.TypeString,
				Optional:    true,
				Description: "resource id to use as destination",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		//Timeouts: &schema.ResourceTimeout{
		//	Create: schema.DefaultTimeout(30 * time.Second),
		//	Update: schema.DefaultTimeout(30 * time.Second),
		//},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if v, ok := d.GetOk("config"); ok {
		input.Configuration = v.(map[string]interface{})
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
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Warning,
		Summary:  "connection connection",
		Detail:   fmt.Sprintf("%+v\n", conn),
	})

	d.SetId(strconv.Itoa(conn.ID))
	resourceConnectionRead(ctx, d, m)

	return diags
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	d.Set("id", strconv.Itoa(conn.ID))
	d.Set("type", conn.Type)
	d.Set("name", conn.Name)
	d.Set("config", conn.Configuration)
	d.Set("metadata", conn.Metadata)
	err = d.Set("streams", flattenStreams(conn))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting streams: %s", err))
	}
	d.Set("state", conn.State)
	d.Set("pipeline_id", conn.PipelineID)
	d.Set("pipeline_name", conn.PipelineName)

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	return diags
}

func flattenStreams(conn *meroxa.Connector) []interface{} {
	s := make(map[string]interface{})
	s["dynamic"] = conn.Streams["dynamic"].(bool)
	s["output"] = conn.Streams["output"]
	s["input"] = conn.Streams["input"]
	return []interface{}{s}
}

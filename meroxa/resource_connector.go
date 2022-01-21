package meroxa

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go/pkg/meroxa"
)

// TODO: DRY this up & move, doesn't quite belong here
const (
	connectorNameMin int = 3
	connectorNameMax int = 64
)

var connectorNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*[a-zA-Z0-9]$`)

func resourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConnectorCreate,
		ReadContext:   resourceConnectorRead,
		UpdateContext: resourceConnectorUpdate,
		DeleteContext: resourceConnectorDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				Description:      "Connector Name",
				ValidateDiagFunc: validateConnectorName(), // todo add validation
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
			"pipeline_id": {
				Type:        schema.TypeInt,
				Description: "Connector's Pipeline ID",
				Required:    true,
			},
			"pipeline_name": {
				Type:        schema.TypeString,
				Description: "Connector's Pipeline Name",
				Computed:    true,
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

	c := m.(meroxa.Client)
	input := &meroxa.CreateConnectorInput{
		Name:          d.Get("name").(string),
		ResourceID:    resourceID,
		Configuration: resourceConnectorConfig(d),
	}

	if v, ok := d.GetOk("pipeline_id"); ok {
		input.PipelineID = v.(int)
	}

	if v, ok := d.GetOk("pipeline_name"); ok {
		input.PipelineName = v.(string)
	}

	if v, ok := d.GetOk("source_id"); ok && v.(string) != "" {
		resourceID, err = strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Type = meroxa.ConnectorTypeSource
	}

	if v, ok := d.GetOk("destination_id"); ok && v.(string) != "" {
		resourceID, err = strconv.Atoi(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Type = meroxa.ConnectorTypeDestination
	}

	if v, ok := d.GetOk("input"); ok && v.(string) != "" {
		input.Input = v.(string)
	}

	input.ResourceID = resourceID

	conn, err := c.CreateConnector(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(conn.ID))

	createStateConf := &resource.StateChangeConf{
		Pending: []string{
			string(meroxa.ConnectorStatePending),
		},
		Target: []string{
			string(meroxa.ConnectorStateRunning),
		},
		Refresh:    resourceConnectorStateFunc(ctx, c, conn.ID),
		Timeout:    10 * time.Minute,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	_, err = createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(
			fmt.Errorf("error waiting for connector (%s) to be created: %s", d.Id(), err),
		)
	}

	resourceConnectorRead(ctx, d, m)

	return diags
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(meroxa.Client)

	cID := d.Id()
	id, err := strconv.Atoi(cID)
	if err != nil {
		return diag.FromErr(err)
	}

	conn, err := c.GetConnectorByNameOrID(ctx, fmt.Sprint(id))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("type", string(conn.Type))
	_ = d.Set("name", conn.Name)

	err = d.Set("streams", flattenStreams(conn))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting streams: %s", err))
	}
	_ = d.Set("state", string(conn.State))
	_ = d.Set("pipeline_id", conn.PipelineID)
	_ = d.Set("pipeline_name", conn.PipelineName)

	// N.B. Configuration is write-only attribute where the platform API
	//      returns empty map. Configuration is persisted in the state only.
	// _ = d.Set("config", conn.Configuration)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(meroxa.Client)

	name := d.Get("name").(string)
	if d.HasChange("state") {
		state := d.Get("state").(string)
		if _, err := c.UpdateConnectorStatus(ctx, name, meroxa.Action(state)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("config") {
		input := &meroxa.UpdateConnectorInput{
			Configuration: resourceConnectorConfig(d),
		}
		if _, err := c.UpdateConnector(ctx, name, input); err != nil {
			return diag.FromErr(err)
		}
	}

	resourceResourceRead(ctx, d, m)

	return diags
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	c := m.(meroxa.Client)
	rID := d.Id()
	id, err := strconv.Atoi(rID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.DeleteConnector(ctx, fmt.Sprint(id))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourceConnectorStateFunc(ctx context.Context, c meroxa.Client, id int) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := c.GetConnectorByNameOrID(ctx, fmt.Sprint(id))
		if err != nil {
			return nil, "", err
		}

		return resp, string(resp.State), nil
	}
}

func flattenStreams(conn *meroxa.Connector) []interface{} {
	s := make(map[string]interface{})
	s["dynamic"] = conn.Streams["dynamic"].(bool)
	s["output"] = conn.Streams["output"]
	s["input"] = conn.Streams["input"]
	return []interface{}{s}
}

func resourceConnectorConfig(d *schema.ResourceData) map[string]interface{} {
	config := make(map[string]interface{})

	if v, ok := d.GetOk("config"); ok {
		for k, v := range v.(map[string]interface{}) {
			config[k] = v
		}
	}

	return config
}

func validateConnectorName() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		name := val.(string)

		if len(name) > connectorNameMax {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid connector name",
				Detail:   fmt.Sprintf("connector name should not be longer than %d characters", connectorNameMax),
			})
			return diags
		}

		if len(name) < connectorNameMin {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid connector name",
				Detail:   fmt.Sprintf("connector name should be at least %d characters long", connectorNameMin),
			})
			return diags
		}

		if name != strings.ToLower(name) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid connector name",
				Detail:   "connector name should only contain lowercase letters",
			})
			return diags
		}

		if !connectorNamePattern.MatchString(name) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Invalid connector name",
				Detail:   "connector name should start with a letter and contain only alphanumeric characters or dashes",
			})
		}

		return diags
	}
}

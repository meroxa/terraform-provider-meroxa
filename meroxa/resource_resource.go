package meroxa

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"log"
	"strconv"
	"strings"
)

func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourceCreate,
		ReadContext:   resourceResourceRead,
		UpdateContext: resourceResourceUpdate,
		DeleteContext: resourceResourceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, //TODO check if we can change name
				//StateFunc: func(i interface{}) string {
				//	return strings.ToLower(i.(string)) // TODO: did we address lowercase name?
				//},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"url": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Resource URL",
				ValidateDiagFunc: validateURL(), // can validate url
				Sensitive:        false,         //if we contain secrets
			},
			"metadata": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_tunnel": &schema.Schema{
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "resource username",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    false,
						},
						"password": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "resource password",
							InputDefault: "",
							ValidateFunc: nil, // todo add validation
							Sensitive:    true,
						},
						"cacert": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "trusted certificates for verifying resource",
						},
						"clientcert": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "client certificate for authenticating to the resource",
						},
						"clientkey": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "client private key for authenticating to the resource",
							Sensitive:   true,
						},
						"ssl": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "use SSL",
						},
					},
				},
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

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	input := meroxa.CreateResourceInput{
		Type: d.Get("type").(string),
		Name: d.Get("name").(string),
		URL:  d.Get("url").(string),
	}
	if v, ok := d.GetOk("credentials"); ok {
		input.Credentials = expandCredentials(v.([]interface{}))
	}

	if v, ok := d.GetOk("metadata"); ok && v != "" {
		err := json.Unmarshal([]byte(v.(string)), &input.Metadata)
		if err != nil {
			return diag.FromErr(err) //may need to add better information here
		}
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		input.SSHTunnel = expandSSHTunnel(v.([]interface{}))
	}

	res, err := c.CreateResource(ctx, &input)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.Itoa(res.ID))
	if tun := res.SSHTunnel; tun != nil {
		detail := fmt.Sprintf(
			"Resource %q is successfully created but is pending for validation!\n"+
				"Paste the following public key on your host:\n"+
				tun.PublicKey+
				"Meroxa will try to connect to the resource for 60 minutes and send an email confirmation after a successful resource validation.", res.Name)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Validate SSH Tunnel",
			Detail:   detail,
		})
	}

	// should we wait for resource state?

	resourceResourceRead(ctx, d, m)

	return diags
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	rID := d.Id()
	id, err := strconv.Atoi(rID)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.GetResource(ctx, id)
	if err != nil {
		return diag.FromErr(err)
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

	return diags
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	//rID := d.Id()
	//id, err := strconv.Atoi(rID)
	//if err != nil {
	//	return diag.FromErr(err)
	//}
	req := meroxa.UpdateResourceInput{
		Name: d.Get("name").(string),
		URL:  d.Get("url").(string),
	}
	if d.HasChange("metadata") {
		err := json.Unmarshal([]byte(d.Get("metadata").(string)), &req.Metadata)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("credentials") {
		req.Credentials = expandCredentials(d.Get("credentials").([]interface{}))
	}
	if d.HasChange("ssh_tunnel") {
		req.SSHTunnel = expandSSHTunnel(d.Get("ssh_tunnel").([]interface{}))
	}
	log.Printf("[DEBUG] Updating meroxa resource: %v", req)
	_, err := c.UpdateResource(ctx, req.Name, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating API Gateway v2 integration: %s", err))
	}
	resourceResourceRead(ctx, d, m)
	return diags
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	c := m.(*meroxa.Client)

	dID := d.Id()
	rID, err := strconv.Atoi(dID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = c.DeleteResource(ctx, rID)
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}

func expandCredentials(vCredentials []interface{}) *meroxa.Credentials {
	credentials := &meroxa.Credentials{}
	if len(vCredentials) == 0 || vCredentials[0] == nil {
		return credentials
	}
	mCredentials := vCredentials[0].(map[string]interface{})

	if vUsername, ok := mCredentials["username"].(string); ok && vUsername != "" {
		credentials.Username = vUsername
	}

	if vPassword, ok := mCredentials["password"].(string); ok && vPassword != "" {
		credentials.Password = vPassword
	}

	if vCacert, ok := mCredentials["cacert"].(string); ok && vCacert != "" {
		credentials.CACert = vCacert
	}

	if vClientCert, ok := mCredentials["clientcert"].(string); ok && vClientCert != "" {
		credentials.ClientCert = vClientCert
	}

	if vClientKey, ok := mCredentials["clientkey"].(string); ok && vClientKey != "" {
		credentials.ClientCertKey = vClientKey
	}

	if vSSL, ok := mCredentials["ssl"].(bool); ok {
		credentials.UseSSL = vSSL
	}
	return credentials
}

func flattenCredentials(credentials *meroxa.Credentials) []interface{} {
	if credentials == nil {
		return nil
	}
	c := make(map[string]interface{})

	c["username"] = credentials.Username
	c["password"] = credentials.Password
	c["cacert"] = credentials.CACert
	c["clientcert"] = credentials.ClientCert
	c["clientkey"] = credentials.ClientCertKey
	c["ssl"] = credentials.UseSSL

	return []interface{}{c}
}

func expandSSHTunnel(vSSHTunnel []interface{}) *meroxa.ResourceSSHTunnelInput {
	sshTunnel := &meroxa.ResourceSSHTunnelInput{}
	if len(vSSHTunnel) == 0 || vSSHTunnel[0] == nil {
		return sshTunnel
	}
	mSSHTunnel := vSSHTunnel[0].(map[string]interface{})

	if vAddress, ok := mSSHTunnel["address"].(string); ok && vAddress != "" {
		sshTunnel.Address = vAddress
	}
	return sshTunnel
}

func flattenSSHTunnel(sshTunnel *meroxa.ResourceSSHTunnel) []interface{} {
	if sshTunnel == nil {
		return nil
	}
	c := make(map[string]interface{})
	c["address"] = sshTunnel.Address
	c["public_key"] = sshTunnel.PublicKey
	return []interface{}{c}
}

func validateURL() schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		urlStr := val.(string)
		s := strings.SplitAfter(urlStr, "://")
		// ensure schema
		if len(s) == 1 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "URL missing Schema",
				Detail:   "Please add correct URL Schema",
			})
		}
		rest := strings.Split(s[1], "@")
		if len(rest) == 2 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "URL includes credentials",
				Detail:   "The apply will fail if username and password are also set",
			})
		}
		return diags
	}
}

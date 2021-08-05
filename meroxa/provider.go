package meroxa

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/meroxa/meroxa-go"
	"golang.org/x/oauth2"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func Provider(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"access_token": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("MEROXA_ACCESS_TOKEN", nil),
				},
				"debug": {
					Type:     schema.TypeBool,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						v := os.Getenv("AUTH0_DEBUG")
						if v == "" {
							return false, nil
						}
						return v == "1" || v == "true" || v == "on", nil
					},
				},
				"timeout": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MEROXA_TIMEOUT", nil),
				},
				"api_url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MEROXA_API_URL", nil),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"meroxa_connector": resourceConnector(),
				"meroxa_endpoint":  resourceEndpoint(),
				"meroxa_pipeline":  resourcePipeline(),
				"meroxa_resource":  resourceResource(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"meroxa_connector":      dataSourceConnector(),
				"meroxa_endpoint":       dataSourceEndpoint(),
				"meroxa_pipeline":       dataSourcePipeline(),
				"meroxa_resource_types": dataSourceResourceTypes(),
				"meroxa_resource":       dataSourceResource(),
				"meroxa_transforms":     dataSourceTransforms(),
			},
		}
		p.ConfigureContextFunc = configure(version)
		return p
	}
}

func configure(version string) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		accessToken := d.Get("access_token").(string)

		// Warning or errors can be collected in a slice type
		var diags diag.Diagnostics

		options := []meroxa.Option{
			meroxa.WithUserAgent(fmt.Sprintf("Meroxa Terraform Provider %s", version)),
		}
		debug := d.Get("debug")
		if debug != "" {
			options = append(options, meroxa.WithDumpTransport(os.Stdout))
		}

		timeoutInt := d.Get("timeout")

		if timeoutInt != "" {
			timeout := int64(timeoutInt.(int))
			options = append(options, meroxa.WithClientTimeout(time.Second*time.Duration(timeout)))
		}

		apiURL := d.Get("api_url")
		if apiURL != "" {
			options = append(options, meroxa.WithBaseURL(apiURL.(string)))
		}

		// WithAuthentication needs to be added after WithDumpTransport
		// to catch requests to auth0
		options = append(options, meroxa.WithAuthentication(
			&oauth2.Config{
				Endpoint: meroxa.OAuth2Endpoint,
			},
			accessToken,
			"",
		))

		c, err := meroxa.New(options...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return c, diags
	}
}
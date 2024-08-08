package roxywi

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	Client *Client
}

const (
	ProviderBaseURL = "base_url"
	LoginField      = "login"
	PasswordField   = "password"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			LoginField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for Roxy-WI.",
				DefaultFunc: schema.EnvDefaultFunc("ROXYWI_USERNAME", nil),
			},
			PasswordField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Password for Roxy-WI.",
				DefaultFunc: schema.EnvDefaultFunc("ROXYWI_PASSWORD", nil),
			},
			ProviderBaseURL: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL to connect for Roxy-WI.",
				DefaultFunc: schema.EnvDefaultFunc("ROXYWI_BASE_URL", "https://demo.roxy-wi.org"),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"roxywi_group":             resourceGroup(),
			"roxywi_udp_listener":      resourceUdpListener(),
			"roxywi_user":              resourceUser(),
			"roxywi_user_role_binding": resourceUserRoleBinding(),
			"roxywi_server":            resourceServer(),
			"roxywi_channel":           resourceChannel(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"roxywi_group":        dataSourceGroup(),
			"roxywi_udp_listener": dataSourceUdpListener(),
			"roxywi_user_role":    dataSourceUserRole(),
		},
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "1.0+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}

	return p
}

func providerConfigure(
	_ context.Context,
	d *schema.ResourceData,
	terraformVersion string,
) (interface{}, diag.Diagnostics) {
	username := d.Get(LoginField).(string)
	password := d.Get(PasswordField).(string)
	apiEndpoint := d.Get(ProviderBaseURL).(string)

	userAgent := fmt.Sprintf("terraform/%s", terraformVersion)

	var diags diag.Diagnostics

	client, err := NewClient(apiEndpoint, username, password, userAgent)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	config := &Config{
		Client: client,
	}

	return config, diags
}

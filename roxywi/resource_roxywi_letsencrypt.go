package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	DomainsField  = "domains"
	ApiKeyField   = "api_key"
	ApiTokenField = "api_token"
	EmailField    = "email"
)

func resourceLetsencrypt() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLetsencryptCreate,
		ReadWithoutTimeout:   resourceLetsencryptRead,
		UpdateWithoutTimeout: resourceLetsencryptUpdate,
		DeleteWithoutTimeout: resourceLetsencryptDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage Let's Encrypt certificates.",

		Schema: map[string]*schema.Schema{
			DescriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the certificate.",
			},
			DomainsField: {
				Type:        schema.TypeList,
				Description: "A list of user groups.",
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			ServerIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the server to deploy to.",
			},
			ApiTokenField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Token to use for authentication in DNS API. For Route53 it is the access key.",
				Sensitive:   true,
			},
			ApiKeyField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Secret key to use for authentication in DNS API. For Route53.",
				Sensitive:   true,
			},
			EmailField: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Email address to use for registration with Let's Encrypt.",
				ValidateFunc: validateEmail,
			},
			TypeField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "What challenge should be used. Available: 'standalone', 'route53', 'digitalocean', 'cloudflare', 'linode'",
				ValidateFunc: validation.StringInSlice([]string{
					"standalone", "route53", "digitalocean", "cloudflare", "linode",
				}, false),
			},
		},
	}
}

func resourceLetsencryptCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	requestBody := map[string]interface{}{
		DescriptionField: d.Get(DescriptionField),
		DomainsField:     d.Get(DomainsField),
		ServerIdField:    d.Get(ServerIdField),
		ApiTokenField:    d.Get(ApiTokenField),
		ApiKeyField:      d.Get(ApiKeyField),
		EmailField:       d.Get(EmailField),
		TypeField:        d.Get(TypeField),
	}

	resp, err := client.doRequest("POST", "api/service/letsencrypt", requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}

	d.SetId(fmt.Sprintf("%d", int(id)))
	return resourceLetsencryptRead(ctx, d, m)
}

func resourceLetsencryptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/letsencrypt/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(DescriptionField, result[DescriptionField])
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(DomainsField, result[DomainsField])
	d.Set(ApiTokenField, result[ApiTokenField])
	d.Set(ApiKeyField, result[ApiKeyField])
	d.Set(EmailField, result[EmailField])
	d.Set(TypeField, result[TypeField])

	return nil
}

func resourceLetsencryptUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	requestBody := map[string]interface{}{
		DescriptionField: d.Get(DescriptionField),
		DomainsField:     d.Get(DomainsField),
		ServerIdField:    d.Get(ServerIdField),
		ApiTokenField:    d.Get(ApiTokenField),
		ApiKeyField:      d.Get(ApiKeyField),
		EmailField:       d.Get(EmailField),
		TypeField:        d.Get(TypeField),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/letsencrypt/%s", id), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLetsencryptRead(ctx, d, m)
}

func resourceLetsencryptDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/letsencrypt/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

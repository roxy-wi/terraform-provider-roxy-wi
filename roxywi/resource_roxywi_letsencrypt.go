package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	DomainsField  = "domains"
	ApiKeyField   = "api_key"
	ApiTokenField = "api_token"
	EmailField    = "email"
)

func resourceLetsencrypt() *schema.Resource {
	fields := map[string]*schema.Schema{
		DescriptionField: {Type: schema.TypeString, Optional: true, Description: "Description of the certificate."},
		DomainsField: {Type: schema.TypeList, Required: true, ForceNew: true, MinItems: 1, MaxItems: 100,
			Elem: &schema.Schema{Type: schema.TypeString}, Description: "Certificate domains; the first determines the PEM name."},
		ServerIdField: {Type: schema.TypeInt, Required: true, ForceNew: true, ValidateFunc: validation.IntAtLeast(1), Description: "Server to deploy to."},
		ApiTokenField: {Type: schema.TypeString, Optional: true, Sensitive: true, ConflictsWith: []string{"dns_profile_id"}, Description: "DNS API token; Route53 secret access key."},
		ApiKeyField:   {Type: schema.TypeString, Optional: true, Sensitive: true, ConflictsWith: []string{"dns_profile_id"}, Description: "Route53 access key ID."},
		EmailField:    {Type: schema.TypeString, Optional: true, ValidateFunc: validateEmail, Description: "ACME account email; required for standalone."},
		TypeField: {Type: schema.TypeString, Required: true, ForceNew: true,
			ValidateFunc: validation.StringInSlice([]string{"standalone", "route53", "digitalocean", "cloudflare", "linode"}, false),
			Description:  "Challenge provider: standalone, route53, digitalocean, cloudflare or linode."},
		"dns_profile_id": {Type: schema.TypeInt, Optional: true, ValidateFunc: validation.IntAtLeast(1), ConflictsWith: []string{ApiKeyField, ApiTokenField}, Description: "Existing DNS profile in the same group (Roxy-WI 9.1+)."},
		"draft":          {Type: schema.TypeBool, Optional: true, Default: false, Description: "Save without issuing. Change to false to run preflight and issue the draft (Roxy-WI 9.1+)."},
		"last_task_id":   {Type: schema.TypeInt, Computed: true, Description: "Latest Operations task ID."},
		"legacy_pending": {Type: schema.TypeBool, Computed: true, Description: "Whether legacy cron migration is required."},
	}
	for key, description := range map[string]string{
		"status": "Certificate lifecycle status.", "pem_name": "Managed PEM filename.",
		"not_after": "Certificate expiry in UTC.", "next_run_at": "Next renewal check in UTC.",
		"retry_at": "Next automatic retry in UTC.", "last_error_code": "Latest certificate error code.",
	} {
		fields[key] = &schema.Schema{Type: schema.TypeString, Computed: true, Description: description}
	}
	return &schema.Resource{
		CreateContext: resourceLetsencryptCreate, ReadContext: resourceLetsencryptRead,
		UpdateContext: resourceLetsencryptUpdate, DeleteContext: resourceLetsencryptDelete,
		Importer:    &schema.ResourceImporter{StateContext: schema.ImportStatePassthroughContext},
		Timeouts:    &schema.ResourceTimeout{Create: schema.DefaultTimeout(30 * time.Minute), Update: schema.DefaultTimeout(30 * time.Minute), Delete: schema.DefaultTimeout(30 * time.Minute)},
		Description: "Manage Let's Encrypt certificates and wait for certificate operations to complete.",
		Schema:      fields,
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			if d.Id() != "" && d.HasChange("draft") {
				old, next := d.GetChange("draft")
				if old == false && next == true {
					return fmt.Errorf("an issued certificate cannot be changed back to a draft")
				}
			}
			return nil
		},
	}
}

func letsencryptPayload(d *schema.ResourceData) map[string]interface{} {
	data := map[string]interface{}{
		DescriptionField: d.Get(DescriptionField), DomainsField: d.Get(DomainsField),
		ServerIdField: d.Get(ServerIdField), EmailField: d.Get(EmailField), TypeField: d.Get(TypeField),
		"draft": d.Get("draft"),
	}
	if profile, ok := d.GetOk("dns_profile_id"); ok {
		data["dns_profile_id"] = profile
	} else {
		if d.HasChange("dns_profile_id") {
			data["dns_profile_id"] = nil
		}
		for _, key := range []string{ApiKeyField, ApiTokenField} {
			// Imported resources have no credentials in state. Omit them to retain the server's secret.
			if value, ok := d.GetOk(key); ok {
				data[key] = value
			} else if d.HasChange(key) {
				data[key] = ""
			}
		}
	}
	return data
}

func resourceLetsencryptCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutCreate))
	defer cancel()
	client := m.(*Config).Client
	resp, err := client.doRequest(ctx, http.MethodPost, "api/service/letsencrypt", letsencryptPayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}
	id, err := apiInt(result, "id")
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprint(id))
	if err := client.waitForTasks(ctx, resp); err != nil {
		return diag.FromErr(err)
	}
	return resourceLetsencryptRead(ctx, d, m)
}

func resourceLetsencryptRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*Config).Client.doRequest(ctx, http.MethodGet, "api/service/letsencrypt/"+d.Id(), nil)
	if err != nil {
		return readDiagnostics(d, err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}
	for _, key := range []string{DescriptionField, DomainsField, ServerIdField, EmailField, TypeField, "draft", "dns_profile_id"} {
		if value, present := result[key]; present {
			if err := d.Set(key, value); err != nil {
				return diag.Errorf("set %s: %v", key, err)
			}
		}
	}
	// Secret values returned as null (or echoed by older APIs) never replace configured credentials.
	state, ok := result["state"].(map[string]interface{})
	if result["state"] != nil && !ok {
		return diag.Errorf("unexpected certificate state format")
	}
	for _, key := range []string{"status", "pem_name", "not_after", "next_run_at", "retry_at", "last_task_id", "last_error_code", "legacy_pending"} {
		if err := d.Set(key, state[key]); err != nil {
			return diag.Errorf("set %s: %v", key, err)
		}
	}
	return nil
}

func resourceLetsencryptUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	client := m.(*Config).Client
	endpoint := "api/service/letsencrypt/" + d.Id()
	resp, err := client.doRequest(ctx, http.MethodPut, endpoint, letsencryptPayload(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.waitForTasks(ctx, resp); err != nil {
		return diag.FromErr(err)
	}
	oldDraft, newDraft := d.GetChange("draft")
	if oldDraft == true && newDraft == false {
		for _, action := range []string{"preflight", "issue"} {
			resp, err := client.doRequest(ctx, http.MethodPatch, endpoint, map[string]string{"action": action})
			if err != nil {
				return diag.FromErr(err)
			}
			if err := client.waitForTasks(ctx, resp); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return resourceLetsencryptRead(ctx, d, m)
}

func resourceLetsencryptDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutDelete))
	defer cancel()
	client := m.(*Config).Client
	resp, err := client.doRequest(ctx, http.MethodDelete, "api/service/letsencrypt/"+d.Id(), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.waitForTasks(ctx, resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

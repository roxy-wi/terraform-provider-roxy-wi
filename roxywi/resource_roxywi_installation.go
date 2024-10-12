package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	AutoStart = "auto_start"
	Checker   = "checker"
	Metrics   = "metrics"
	Docker    = "docker"
	Service   = "service"
)

func resourceServiceInstallation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceInstallationCreate,
		ReadWithoutTimeout:   resourceServiceInstallationRead,
		UpdateWithoutTimeout: resourceServiceInstallationUpdate,
		DeleteWithoutTimeout: resourceServiceInstallationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Description: "Manages service installation and Tools settings.",

		Schema: map[string]*schema.Schema{
			"service": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service name.",
				ForceNew:    true,
			},
			"server_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server ID.",
				ForceNew:    true,
			},
			"auto_start": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is Auto start tool enabled for this service.",
			},
			"checker": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is Checker tool enabled for this service.",
			},
			"metrics": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is Metrics tool enabled for this service.",
			},
			"docker": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Is this service should be run in Docker container.",
			},
		},
	}
}

func resourceServiceInstallationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	service, ok := d.Get("service").(string)
	if !ok || service == "" {
		return diag.Errorf("service is required and must be a string")
	}

	serverID, ok := d.Get("server_id").(int)
	if !ok {
		return diag.Errorf("server_id is required and must be an int")
	}

	autoStart := boolToInt(d.Get("auto_start").(bool))
	checker := boolToInt(d.Get("checker").(bool))
	metrics := boolToInt(d.Get("metrics").(bool))
	docker := boolToInt(d.Get("docker").(bool))

	payload := map[string]interface{}{
		"auto_start": autoStart,
		"checker":    checker,
		"metrics":    metrics,
		"docker":     docker,
	}

	url := fmt.Sprintf("/api/service/%s/%d/install", service, serverID)
	resp, err := client.doRequest(http.MethodPost, url, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.Errorf("unexpected response format, could not unmarshal: %s", string(resp))
	}

	id, ok := result["id"].(string)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}
	d.SetId(id)

	return resourceServiceInstallationRead(ctx, d, m)
}

func resourceServiceInstallationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	service, ok := d.Get("service").(string)
	if !ok || service == "" {
		return diag.Errorf("service is required and must be a string")
	}

	serverID, ok := d.Get("server_id").(int)
	if !ok {
		return diag.Errorf("server_id is required and must be an int")
	}

	autoStart := boolToInt(d.Get("auto_start").(bool))
	checker := boolToInt(d.Get("checker").(bool))
	metrics := boolToInt(d.Get("metrics").(bool))
	docker := boolToInt(d.Get("docker").(bool))

	payload := map[string]interface{}{
		"auto_start": autoStart,
		"checker":    checker,
		"metrics":    metrics,
		"docker":     docker,
	}

	url := fmt.Sprintf("/api/service/%s/%d/install", service, serverID)
	resp, err := client.doRequest(http.MethodPut, url, payload)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.Errorf("unexpected response format, could not unmarshal: %s", string(resp))
	}

	id, ok := result["id"].(string)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}
	d.SetId(id)

	return resourceServiceInstallationRead(ctx, d, m)
}

func resourceServiceInstallationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	// Получаем идентификатор ресурса
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'resourceID-serverID', got: %s", fullID)
	}

	// Используем первую часть как идентификатор ресурса
	id := parts[0]
	service := parts[1]

	url := fmt.Sprintf("/api/service/%s/%s/install", service, id)
	resp, err := client.doRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.Errorf("unexpected response format, could not unmarshal: %s", string(resp))
	}

	// Extracting the data and ensuring they are set correctly
	d.Set(AutoStart, intToBool(result[AutoStart].(float64)))
	d.Set(Checker, intToBool(result[Checker].(float64)))
	d.Set(Metrics, intToBool(result[Metrics].(float64)))
	d.Set(Docker, intToBool(result[Docker].(float64)))
	d.Set(ServerField, result[ServerField].(float64))
	d.Set(Service, result[Service].(string))

	return nil
}

func resourceServiceInstallationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	service := d.Get("service").(string)
	serverID := d.Get("server_id").(int)

	url := fmt.Sprintf("/api/service/%s/%d/install", service, serverID)
	_, err := client.doRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Удаляем ресурс из состояния Terraform
	d.SetId("")

	return nil
}

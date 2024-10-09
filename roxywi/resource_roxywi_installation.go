package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	AutoStart   = "auto_start"
	Checker     = "checker"
	Metrics     = "metrics"
	SynFlood    = "syn_flood"
	Servers     = "servers"
	Services    = "services"
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
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     0,
                Description: "Is Auto start tool enabled for this service.",
            },
            "checker": {
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     0,
                Description: "Is Checker tool enabled for this service.",
            },
            "metrics": {
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     0,
                Description: "Is Metrics tool enabled for this service.",
            },
            "syn_flood": {
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     0,
                Description: "SYN flood setting.",
            },
            "docker": {
                Type:        schema.TypeInt,
                Optional:    true,
                Default:     0,
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

    autoStart := d.Get("auto_start").(int)
    checker := d.Get("checker").(int)
    metrics := d.Get("metrics").(int)
    synFlood := d.Get("syn_flood").(int)
    docker := d.Get("docker").(int)

    payload := map[string]interface{}{
        "auto_start": autoStart,
        "checker":    checker,
        "metrics":    metrics,
        "syn_flood":  synFlood,
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

    autoStart := d.Get("auto_start").(int)
    checker := d.Get("checker").(int)
    metrics := d.Get("metrics").(int)
    synFlood := d.Get("syn_flood").(int)
    docker := d.Get("docker").(int)

    payload := map[string]interface{}{
        "auto_start": autoStart,
        "checker":    checker,
        "metrics":    metrics,
        "syn_flood":  synFlood,
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
    service, ok := d.Get("service").(string)
    if !ok || service == "" {
        return diag.Errorf("service must be set and must be a string")
    }

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
    if autoStart, ok := result["auto_start"].(int); ok {
        d.Set("auto_start", autoStart)
    }
    if checker, ok := result["checker"].(int); ok {
        d.Set("checker", checker)
    }
    if metrics, ok := result["metrics"].(int); ok {
        d.Set("metrics", metrics)
    }
    if synFlood, ok := result["syn_flood"].(int); ok {
        d.Set("syn_flood", synFlood)
    }
    if docker, ok := result["docker"].(int); ok {
        d.Set("docker", docker)
    }

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

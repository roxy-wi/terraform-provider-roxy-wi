package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ClusterIdField     = "cluster_id"
	ConfigField        = "config"
	BackendIPField     = "backend_ip"
	BackendPortField   = "port"
	BackendWeightField = "weight"
	GroupIdField       = "group_id"
	LbAlgorithmField   = "lb_algo"
	PortField          = "port"
	ServerIdField      = "server_id"
	VIPField           = "vip"
)

var validLbAlgorithms = []string{
	"rr",
	"wrr",
	"lc",
	"wlc",
	"sh",
	"dh",
	"lblc",
}

func resourceUdpListener() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUdpListenerCreate,
		ReadWithoutTimeout:   resourceUdpListenerRead,
		UpdateWithoutTimeout: resourceUdpListenerUpdate,
		DeleteWithoutTimeout: resourceUdpListenerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage UDP listeners in Roxy-WI. All servers managed via Roxy-WI can be included in groups.",

		Schema: map[string]*schema.Schema{
			ClusterIdField: {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{ClusterIdField, ServerIdField},
				Description:  "ID of the cluster.",
			},
			ConfigField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Configuration for the backend servers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						BackendIPField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "IP address of the backend server.",
						},
						BackendPortField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port of the backend server.",
						},
						BackendWeightField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Weight of the backend server.",
						},
					},
				},
			},
			DescriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the UDP listener.",
			},
			GroupIdField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the group.",
			},
			LbAlgorithmField: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      fmt.Sprintf("Load balancing algorithm. Available values are: %v", validLbAlgorithms),
				ValidateDiagFunc: validateLbAlgorithm(),
			},
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the UDP listener.",
			},
			PortField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Port on which the UDP listener will listen.",
			},
			ServerIdField: {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{ClusterIdField, ServerIdField},
				Description:  "ID of the server.",
			},
			VIPField: {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Virtual IP address.",
				ValidateDiagFunc: validateVip(),
			},
		},
	}
}

func resourceUdpListenerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	clusterID := d.Get(ClusterIdField).(int)
	serverID := d.Get(ServerIdField).(int)
	vip := d.Get(VIPField).(string)

	if err := checkVipExists(client, clusterID, serverID, vip); err != nil {
		return diag.FromErr(err)
	}

	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")
	name := strings.ReplaceAll(d.Get(NameField).(string), "'", "")
	groupID := d.Get(GroupIdField).(string)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	port := d.Get(PortField).(string)

	configs := parseConfigList(d.Get(ConfigField).([]interface{}))

	requestBody := map[string]interface{}{
		ClusterIdField:   clusterID,
		ConfigField:      configs,
		DescriptionField: description,
		GroupIdField:     groupID,
		LbAlgorithmField: lbAlgo,
		NameField:        name,
		PortField:        port,
		ServerIdField:    serverID,
		VIPField:         vip,
	}

	resp, err := client.doRequest("POST", "/api/udp/listener", requestBody)
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
	return resourceUdpListenerRead(ctx, d, m)
}

func resourceUdpListenerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/udp/listener/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(ClusterIdField, intFromInterface(result[ClusterIdField]))
	description := strings.ReplaceAll(result[DescriptionField].(string), "'", "")
	name := strings.ReplaceAll(result[NameField].(string), "'", "")
	d.Set(DescriptionField, description)
	d.Set(NameField, name)
	d.Set(GroupIdField, intFromInterface(result[GroupIdField]))
	d.Set(LbAlgorithmField, result[LbAlgorithmField])
	d.Set(PortField, intFromInterface(result[PortField]))
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(VIPField, result[VIPField])

	config, err := parseConfig(result["config"])
	if err != nil {
		return diag.FromErr(err)
	}

	configList := parseConfigResult(config)
	d.Set(ConfigField, configList)

	return nil
}

func resourceUdpListenerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	clusterID := d.Get(ClusterIdField).(int)
	serverID := d.Get(ServerIdField).(int)
	vip := d.Get(VIPField).(string)

	if err := checkVipExists(client, clusterID, serverID, vip); err != nil {
		return diag.FromErr(err)
	}

	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")
	name := strings.ReplaceAll(d.Get(NameField).(string), "'", "")
	groupID := d.Get(GroupIdField).(string)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	port := d.Get(PortField).(string)

	configs := parseConfigList(d.Get(ConfigField).([]interface{}))

	requestBody := map[string]interface{}{
		ClusterIdField:   clusterID,
		ConfigField:      configs,
		DescriptionField: description,
		GroupIdField:     groupID,
		LbAlgorithmField: lbAlgo,
		NameField:        name,
		PortField:        port,
		ServerIdField:    serverID,
		VIPField:         vip,
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/udp/listener/%s", id), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceUdpListenerRead(ctx, d, m)
}

func resourceUdpListenerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/udp/listener/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func parseConfigList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			BackendIPField:     configDetails[BackendIPField].(string),
			BackendPortField:   intFromInterface(configDetails[BackendPortField]),
			BackendWeightField: intFromInterface(configDetails[BackendWeightField]),
		})
	}
	return configs
}

func parseConfig(config interface{}) ([]map[string]interface{}, error) {
	switch v := config.(type) {
	case string:
		var parsedConfig []map[string]interface{}
		log.Printf("[DEBUG] Config string before parsing: %s", v)
		v = strings.ReplaceAll(v, "'", "\"")
		if err := json.Unmarshal([]byte(v), &parsedConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config field: %v", err)
		}
		return parsedConfig, nil
	case []interface{}:
		var parsedConfig []map[string]interface{}
		for _, item := range v {
			parsedConfig = append(parsedConfig, item.(map[string]interface{}))
		}
		return parsedConfig, nil
	default:
		return nil, fmt.Errorf("invalid config format: %v", config)
	}
}

func parseConfigResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			BackendIPField:     c[BackendIPField].(string),
			BackendPortField:   intFromInterface(c[BackendPortField]),
			BackendWeightField: intFromInterface(c[BackendWeightField]),
		})
	}
	return configList
}

func intFromInterface(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func validateLbAlgorithm() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", path)
		}

		for _, str := range validLbAlgorithms {
			if v == str {
				return nil
			}
		}

		return diag.Errorf("invalid value for %s: %s. Valid values are: %v", path, v, validLbAlgorithms)
	}
}

func validateVip() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		vip, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", path)
		}

		if vip == "" {
			return diag.Errorf("VIP cannot be empty")
		}

		return nil
	}
}

func checkVipExists(client *Client, clusterID, serverID int, vip string) error {
	var url string
	if clusterID != 0 {
		url = fmt.Sprintf("/api/ha/cluster/%d/vips", clusterID)
	} else if serverID != 0 {
		url = fmt.Sprintf("/api/server/%d/ip", serverID)
	} else {
		return fmt.Errorf("either cluster_id or server_id must be specified")
	}

	log.Printf("[DEBUG] Checking VIP existence with URL: %s", url)

	resp, err := client.doRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	log.Printf("[DEBUG] Response from VIP check: %s", string(resp))

	if clusterID != 0 {
		var result []map[string]interface{}
		if err := json.Unmarshal(resp, &result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}

		for _, item := range result {
			if itemVip, ok := item["vip"].(string); ok && itemVip == vip {
				return nil
			}
		}
	} else {
		var ips []string
		if err := json.Unmarshal(resp, &ips); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}

		log.Printf("[DEBUG] Parsed response: %v", ips)

		for _, itemVip := range ips {
			if itemVip == vip {
				return nil
			}
		}
	}

	return fmt.Errorf("VIP %s not found", vip)
}

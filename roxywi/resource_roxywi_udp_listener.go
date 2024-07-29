package roxywi

import (
	"context"
	"encoding/json"
	"fmt"

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
	"Round robin",
	"Weighted Round Robin",
	"Least Connection",
	"Weighted Least Connection",
	"Source Hashing",
	"Destination Hashing",
	"Locality-Based Least Connection",
}

func resourceUdpListener() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUdpListenerCreate,
		ReadContext:   resourceUdpListenerRead,
		UpdateContext: resourceUdpListenerUpdate,
		DeleteContext: resourceUdpListenerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				Type:        schema.TypeSet,
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "Port of the backend server.",
						},
						BackendWeightField: {
							Type:        schema.TypeString,
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

	description := d.Get(DescriptionField).(string)
	groupID := d.Get(GroupIdField).(string)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	name := d.Get(NameField).(string)
	port := d.Get(PortField).(string)

	configs := parseConfigSet(d.Get(ConfigField).(*schema.Set))

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

	d.Set(ClusterIdField, result[ClusterIdField])
	d.Set(DescriptionField, result[DescriptionField])
	d.Set(GroupIdField, result[GroupIdField])
	d.Set(LbAlgorithmField, result[LbAlgorithmField])
	d.Set(NameField, result[NameField])
	d.Set(PortField, result[PortField])
	d.Set(ServerIdField, result[ServerIdField])
	d.Set(VIPField, result[VIPField])

	if config, ok := result["config"].([]interface{}); ok {
		configSet := parseConfigResult(config)
		d.Set(ConfigField, configSet)
	}

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

	description := d.Get(DescriptionField).(string)
	groupID := d.Get(GroupIdField).(string)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	name := d.Get(NameField).(string)
	port := d.Get(PortField).(string)

	configs := parseConfigSet(d.Get(ConfigField).(*schema.Set))

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

func parseConfigSet(configSet *schema.Set) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configSet.List() {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			BackendIPField:     configDetails[BackendIPField].(string),
			BackendPortField:   configDetails[BackendPortField].(string),
			BackendWeightField: configDetails[BackendWeightField].(string),
		})
	}
	return configs
}

func parseConfigResult(config []interface{}) *schema.Set {
	configSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			BackendIPField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP address of the backend server.",
			},
			BackendPortField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Port of the backend server.",
			},
			BackendWeightField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Weight of the backend server.",
			},
		},
	}), nil)

	for _, c := range config {
		configDetails := c.(map[string]interface{})
		configSet.Add(map[string]interface{}{
			BackendIPField:     configDetails[BackendIPField].(string),
			BackendPortField:   configDetails[BackendPortField].(string),
			BackendWeightField: configDetails[BackendWeightField].(string),
		})
	}

	return configSet
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

	resp, err := client.doRequest("GET", url, nil)
	if err != nil {
		return err
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	for _, item := range result {
		if itemVip, ok := item["vip"].(string); ok && itemVip == vip {
			return nil
		}
	}

	return fmt.Errorf("VIP %s not found", vip)
}

package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ClusterIdField                    = "cluster_id"
	ConfigField                       = "config"
	BackendIPField                    = "backend_ip"
	BackendPortField                  = "port"
	BackendWeightField                = "weight"
	GroupIdField                      = "group_id"
	LbAlgorithmField                  = "lb_algo"
	PortField                         = "port"
	ServerIdField                     = "server_id"
	VIPField                          = "vip"
	ReconfigureField                  = "reconfigure"
	LbAlgorithmRoundRobin             = "rr"
	LbAlgorithmWeightRoundRobin       = "wrr"
	LbAlgorithmLeastConn              = "lc"
	LbAlgorithmWeightLeastConn        = "wlc"
	LbAlgorithmSourceHash             = "sh"
	LbAlgorithmDestinationHash        = "dh"
	LbAlgorithmLocalityBasedLeastConn = "lblc"
	IsCheckerFileld                   = "is_checker"
)

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
				Description:  fmt.Sprintf("Cluster ID where the UDP listener is located. Must be determined if `%s` empty.", ServerIdField),
			},
			ConfigField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Configuration for the backend servers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						BackendIPField: {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "IP address of the backend server.",
							ValidateFunc: validation.IsIPAddress,
						},
						BackendPortField: {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Port number on which the backend server listens for requests.",
							ValidateFunc: validation.IsPortNumber,
						},
						BackendWeightField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Weight assigned to the backend server.",
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
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the group.",
			},
			LbAlgorithmField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Load balancing algorithm. Available values are: `%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`.", LbAlgorithmRoundRobin, LbAlgorithmWeightRoundRobin, LbAlgorithmLeastConn, LbAlgorithmWeightLeastConn, LbAlgorithmSourceHash, LbAlgorithmDestinationHash, LbAlgorithmLocalityBasedLeastConn),
				ValidateFunc: validation.StringInSlice([]string{
					LbAlgorithmRoundRobin,
					LbAlgorithmWeightRoundRobin,
					LbAlgorithmLeastConn,
					LbAlgorithmWeightLeastConn,
					LbAlgorithmSourceHash,
					LbAlgorithmDestinationHash,
					LbAlgorithmLocalityBasedLeastConn,
				}, false),
			},
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the UDP listener.",
			},
			PortField: {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Port on which the UDP listener will listen.",
				ValidateFunc: validation.IsPortNumber,
			},
			ServerIdField: {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{ClusterIdField, ServerIdField},
				Description:  fmt.Sprintf("Server ID where the UDP listener is located. Must be determined if `%s` empty", ClusterIdField),
			},
			VIPField: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  fmt.Sprintf("IP address of the UDP listener binding, if `%s` specified. VIP address of the UDP listener binding, if `%s` specified. Must be a valid IPv4 and exists.", ServerIdField, ClusterIdField),
				ValidateFunc: validation.IsIPAddress,
			},
			IsCheckerFileld: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Should be Checker service check this UDP listener?",
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
	groupID := d.Get(GroupIdField).(int)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	port := d.Get(PortField).(int)

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
		ReconfigureField: true,
		IsCheckerFileld:  boolToInt(d.Get(IsCheckerFileld).(bool)),
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
	d.Set(IsCheckerFileld, intToBool(result[IsCheckerFileld].(float64)))

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
	groupID := d.Get(GroupIdField).(int)
	lbAlgo := d.Get(LbAlgorithmField).(string)
	port := d.Get(PortField).(int)

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
		IsCheckerFileld:  boolToInt(d.Get(IsCheckerFileld).(bool)),
	}

	if d.HasChange(ConfigField) || d.HasChange(LbAlgorithmField) || d.HasChange(PortField) || d.HasChange(VIPField) {
		requestBody[ReconfigureField] = true
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

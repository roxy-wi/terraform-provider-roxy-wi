package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"time"
)

func resourceHaClusterVip() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaClusterVipCreate,
		ReadWithoutTimeout:   resourceHaClusterVipRead,
		UpdateWithoutTimeout: resourceHaClusterVipUpdate,
		DeleteWithoutTimeout: resourceHaClusterVipDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage additional VIP for HA cluster.",

		Schema: map[string]*schema.Schema{
			ReturnToMasterField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Return to master setting for the HA Cluster.",
			},
			ServersField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of servers in the HA Cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						IDField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Server ID.",
						},
						MasterField: {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Master setting for the server.",
						},
						EthField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Ethernet interface for the server.",
						},
					},
				},
			},
			UseSrcField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use source setting for the HA Cluster.",
			},
			VIPField: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Virtual IP address for the HA Cluster.",
				ValidateFunc: validation.IsIPAddress,
			},
			VirtServerField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Virtual server setting for the HA Cluster.",
			},
			ClusterIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "Cluster ID.",
			},
		},
	}
}

func resourceHaClusterVipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	clusterId := d.Get(ClusterIdField).(int)

	servers := parseServersList(d.Get(ServersField).([]interface{}))
	fmt.Printf("Servers: %+v\n", servers)

	haCluster := map[string]interface{}{
		ClusterIdField:      clusterId,
		ReturnToMasterField: boolToInt(d.Get(ReturnToMasterField).(bool)),
		ServersField:        servers,
		UseSrcField:         boolToInt(d.Get(UseSrcField).(bool)),
		VIPField:            d.Get(VIPField).(string),
		VirtServerField:     boolToInt(d.Get(VirtServerField).(bool)),
		ReconfigureField:    true,
	}

	jsonData, _ := json.Marshal(haCluster)
	fmt.Printf("HA Cluster VIPData: %s\n", string(jsonData))

	resp, err := client.doRequest("POST", fmt.Sprintf("/api/ha/cluster/%d/vip", clusterId), haCluster)
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

	d.SetId(fmt.Sprintf("%d-vip-%d", clusterId, int(id)))
	return resourceHaClusterVipRead(ctx, d, m)
}

func resourceHaClusterVipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	fullId := d.Id()
	clusterId, vipId, err := resourceParseId(fullId, "-vip-")
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	servers, err := parseConfig(result[ServersField])
	if err != nil {
		return diag.FromErr(err)
	}
	serversResult := parseServersResult(servers)

	d.Set(ClusterIdField, clusterId)
	d.Set(ReturnToMasterField, intToBool(result[ReturnToMasterField].(float64)))
	d.Set(ServersField, serversResult)
	d.Set(UseSrcField, intToBool(result[UseSrcField].(float64)))
	d.Set(VIPField, result[VIPField])
	d.Set(VirtServerField, intToBool(result[VirtServerField].(float64)))

	return nil
}

func resourceHaClusterVipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullId := d.Id()
	clusterId, vipId, err1 := resourceParseId(fullId, "-vip-")
	if err1 != nil {
		return diag.FromErr(err1)
	}

	servers := parseServersList(d.Get(ServersField).([]interface{}))

	haCluster := map[string]interface{}{
		ClusterIdField:      clusterId,
		ReturnToMasterField: boolToInt(d.Get(ReturnToMasterField).(bool)),
		ServersField:        servers,
		VIPField:            d.Get(VIPField).(string),
		VirtServerField:     boolToInt(d.Get(VirtServerField).(bool)),
		UseSrcField:         boolToInt(d.Get(UseSrcField).(bool)),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), haCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaClusterVipRead(ctx, d, m)
}

func resourceHaClusterVipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullId := d.Id()
	clusterId, vipId, err1 := resourceParseId(fullId, "-vip-")
	if err1 != nil {
		return diag.FromErr(err1)
	}

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

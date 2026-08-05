package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHaClusterVip() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHaClusterVipCreate,
		ReadContext:   resourceHaClusterVipRead,
		UpdateContext: resourceHaClusterVipUpdate,
		DeleteContext: resourceHaClusterVipDelete,

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

	haCluster := map[string]interface{}{
		ClusterIdField:      clusterId,
		ReturnToMasterField: boolToInt(d.Get(ReturnToMasterField).(bool)),
		ServersField:        servers,
		UseSrcField:         boolToInt(d.Get(UseSrcField).(bool)),
		VIPField:            d.Get(VIPField).(string),
		VirtServerField:     boolToInt(d.Get(VirtServerField).(bool)),
		ReconfigureField:    true,
	}

	resp, err := client.doRequest(ctx, "POST", fmt.Sprintf("/api/ha/cluster/%d/vip", clusterId), haCluster)
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

	d.SetId(fmt.Sprintf("%d-vip-%d", clusterId, id))
	return resourceHaClusterVipRead(ctx, d, m)
}

func resourceHaClusterVipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	fullId := d.Id()
	clusterId, vipId, err := resourceParseId(fullId, "-vip-")
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.doRequest(ctx, "GET", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), nil)
	if err != nil {
		return readDiagnostics(d, err)
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

	clusterIDValue, err := strconv.Atoi(clusterId)
	if err != nil {
		return diag.Errorf("invalid cluster ID %q: %v", clusterId, err)
	}
	returnToMaster, err := apiBool(result, ReturnToMasterField)
	if err != nil {
		return diag.FromErr(err)
	}
	useSource, err := apiBool(result, UseSrcField)
	if err != nil {
		return diag.FromErr(err)
	}
	virtualServer, err := apiBool(result, VirtServerField)
	if err != nil {
		return diag.FromErr(err)
	}
	vip, err := apiString(result, VIPField)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(ClusterIdField, clusterIDValue); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(ReturnToMasterField, returnToMaster); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(ServersField, serversResult); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(UseSrcField, useSource); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(VIPField, vip); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(VirtServerField, virtualServer); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}

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

	_, err := client.doRequest(ctx, "PUT", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), haCluster)
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

	_, err := client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/ha/cluster/%s/vip/%s", clusterId, vipId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

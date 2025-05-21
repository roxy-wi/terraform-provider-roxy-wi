package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNginxSectionUpstream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNginxSectionUpstreamCreate,
		ReadWithoutTimeout:   resourceNginxSectionUpstreamRead,
		UpdateWithoutTimeout: resourceNginxSectionUpstreamUpdate,
		DeleteWithoutTimeout: resourceNginxSectionUpstreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage Nginx upstream sections. Please note that changes may cause Nginx to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Upstream section.",
			},
			BackendServersField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of backend servers configuration.",
				Elem: &schema.Resource{
					Schema: nginxBackendServerSchema(),
				},
			},
			BalanceField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Load balancing algorithm. Available values are: `%s`,`%s`,`%s`,`%s`.", NginxBalanceRoundRobin, NginxBalanceIpHash, NginxBalanceLeastConn, NginxBalanceRandom),
				ValidateFunc: validation.StringInSlice([]string{
					NginxBalanceRoundRobin,
					NginxBalanceIpHash,
					NginxBalanceLeastConn,
					NginxBalanceRandom,
				}, false),
			},
			ServerIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the server to deploy to.",
			},
			ActionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "What action should be taken after changing the config. Available: save, reload, restart.",
				Default:     "save",
				ValidateFunc: validation.StringInSlice([]string{
					"save",
					"reload",
					"restart",
				}, false),
			},
			NginxKeepAlive: {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Connection keep alive timeout in seconds. Default: 32.",
				Default:     32,
			},
		},
	}
}

func resourceNginxSectionUpstreamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	backends := parseNginxBackendsServerList(d.Get(BackendServersField).([]interface{}))

	requestBody := map[string]interface{}{
		BackendServersField: backends,
		NameField:           d.Get(NameField),
		ServerIdField:       d.Get(ServerIdField),
		ActionField:         d.Get(ActionField),
		BalanceField:        d.Get(BalanceField),
		NginxKeepAlive:      d.Get(NginxKeepAlive),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/nginx/%d/section/upstream", d.Get(ServerIdField)), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"].(string)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}

	d.SetId(fmt.Sprintf("%s", id))
	return resourceNginxSectionUpstreamRead(ctx, d, m)
}

func resourceNginxSectionUpstreamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId, sectionName, err := resourceSectionParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/nginx/%s/section/upstream/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(NameField, result[NameField])
	d.Set(BalanceField, result[BalanceField])
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(NginxKeepAlive, intFromInterface(result[NginxKeepAlive]))

	backendServers, err := parseConfig(result[BackendServersField])
	if err != nil {
		return diag.FromErr(err)
	}

	backendServersList := parseNginxBackendServerResult(backendServers)
	d.Set(BackendServersField, backendServersList)

	return nil
}

func resourceNginxSectionUpstreamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	backends := parseNginxBackendsServerList(d.Get(BackendServersField).([]interface{}))

	requestBody := map[string]interface{}{
		BackendServersField: backends,
		NameField:           d.Get(NameField),
		ServerIdField:       d.Get(ServerIdField),
		ActionField:         d.Get(ActionField),
		BalanceField:        d.Get(BalanceField),
		NginxKeepAlive:      d.Get(NginxKeepAlive),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/nginx/%d/section/upstream/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceNginxSectionUpstreamRead(ctx, d, m)
}

func resourceNginxSectionUpstreamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/nginx/%d/section/upstream/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

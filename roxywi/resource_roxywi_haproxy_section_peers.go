package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	PeersField    = "peers"
	PeerNameField = "name"
)

func resourceHaproxySectionPeers() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxySectionPeersCreate,
		ReadWithoutTimeout:   resourceHaproxySectionPeersRead,
		UpdateWithoutTimeout: resourceHaproxySectionPeersUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionPeersDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Peers sections. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Peers section.",
			},
			PeersField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of peers configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						IPField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "IP address of a peer server.",
						},
						PeerNameField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Peer name. Must be the same as a server hostname",
						},
						PortField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Peer port.",
						},
					},
				},
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
			},
		},
	}
}

func resourceHaproxySectionPeersCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	configs := parsePeersConfigList(d.Get(PeersField).([]interface{}))

	requestBody := map[string]interface{}{
		PeersField:    configs,
		NameField:     d.Get(NameField),
		TypeField:     "peers",
		ServerIdField: d.Get(ServerIdField),
		ActionField:   d.Get(ActionField),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/haproxy/%d/section/peers", d.Get(ServerIdField)), requestBody)
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
	return resourceHaproxySectionPeersRead(ctx, d, m)
}

func resourceHaproxySectionPeersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]
	sectionName := parts[1]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/peers/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(NameField, result[NameField])
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(PeersField, result[PeersField])

	config, err := parseConfig(result["peers"])
	if err != nil {
		return diag.FromErr(err)
	}

	configList := parsePeersConfigListResult(config)
	d.Set(ConfigField, configList)

	return nil
}

func resourceHaproxySectionPeersUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	configs := parsePeersConfigList(d.Get(PeersField).([]interface{}))

	requestBody := map[string]interface{}{
		PeersField:    configs,
		NameField:     d.Get(NameField),
		TypeField:     "peers",
		ServerIdField: d.Get(ServerIdField),
		ActionField:   d.Get(ActionField),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/peers/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionPeersRead(ctx, d, m)
}

func resourceHaproxySectionPeersDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/peers/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

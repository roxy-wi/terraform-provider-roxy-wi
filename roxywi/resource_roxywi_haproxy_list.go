package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ServerIpField = "server_ip"
	ColorField    = "color"
	ContentField  = "content"
)

func resourceHaproxyList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxyListCreate,
		ReadWithoutTimeout:   resourceHaproxyListRead,
		UpdateWithoutTimeout: resourceHaproxyListUpdate,
		DeleteWithoutTimeout: resourceHaproxyListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy white and black lists. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the List.",
			},
			ServerIpField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The IP of the server to deploy to.",
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
			ColorField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The white or black list. Available: white, black.",
				ValidateFunc: validation.StringInSlice([]string{
					"white",
					"black",
				}, false),
			},
			ContentField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The content of the list whit `\n` separator.",
			},
			GroupIDField: {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The ID of the group to deploy to. Only for superAdmin role.",
			},
		},
	}
}

func resourceHaproxyListCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	requestBody := map[string]interface{}{
		NameField:     d.Get(NameField),
		ColorField:    d.Get(ColorField),
		ContentField:  d.Get(ContentField),
		ServerIpField: d.Get(ServerIpField),
		ActionField:   d.Get(ActionField),
		GroupIDField:  d.Get(GroupIDField),
	}

	resp, err := client.doRequest("POST", "api/service/haproxy/list", requestBody)
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
	return resourceHaproxyListRead(ctx, d, m)
}

func resourceHaproxyListRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	parts := strings.Split(d.Id(), "-")
	if len(parts) < 3 {
		return diag.FromErr(fmt.Errorf("expected ID in the format 'group_id-color-list_name.lst', got: %s", d.Id()))
	}
	color := parts[1]
	listName := parts[2]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/list/%s/%s", listName, color), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(NameField, result[NameField])
	d.Set(ServerIpField, result[ServerIpField])
	d.Set(ActionField, result[ActionField])
	d.Set(ColorField, result[ColorField])
	d.Set(ContentField, result[ContentField])
	d.Set(GroupIDField, intFromInterface(result[GroupIDField]))

	return nil
}

func resourceHaproxyListUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	requestBody := map[string]interface{}{
		NameField:     d.Get(NameField),
		ColorField:    d.Get(ColorField),
		ContentField:  d.Get(ContentField),
		ServerIpField: d.Get(ServerIpField),
		ActionField:   d.Get(ActionField),
		GroupIDField:  d.Get(GroupIDField),
	}

	_, err := client.doRequest("PUT", "api/service/haproxy/list", requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxyListRead(ctx, d, m)
}

func resourceHaproxyListDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	requestBody := map[string]interface{}{
		NameField:    d.Get(NameField),
		ColorField:   d.Get(ColorField),
		GroupIDField: d.Get(GroupIDField),
	}

	_, err := client.doRequest("DELETE", "api/service/haproxy/list", requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

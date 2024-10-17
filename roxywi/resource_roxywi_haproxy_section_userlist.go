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
	UserListField  = "userlist_users"
	GroupNameField = "group"
	UserListGroup  = "userlist_groups"
	UserFiled      = "user"
)

func resourceHaproxySectionUserlist() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxySectionUserlistCreate,
		ReadWithoutTimeout:   resourceHaproxySectionUserlistRead,
		UpdateWithoutTimeout: resourceHaproxySectionUserlistUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionUserlistDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy User list sections. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Userlist section.",
			},
			UserListGroup: {
				Type:        schema.TypeList,
				Description: "A list of user groups.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			UserListField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of userlist configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						UserFiled: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User name.",
						},
						PasswordField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "User password.",
						},
						GroupNameField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "User group.",
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

func resourceHaproxySectionUserlistCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	configs := parseUserListConfigList(d.Get(UserListField).([]interface{}))

	requestBody := map[string]interface{}{
		UserListField: configs,
		NameField:     d.Get(NameField),
		TypeField:     "userlist",
		ServerIdField: d.Get(ServerIdField),
		UserListGroup: d.Get(UserListGroup),
		ActionField:   d.Get(ActionField),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/haproxy/%d/section/userlist", d.Get(ServerIdField)), requestBody)
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
	return resourceHaproxySectionUserlistRead(ctx, d, m)
}

func resourceHaproxySectionUserlistRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]
	sectionName := parts[1]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/userlist/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(NameField, result[NameField])
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(UserListField, result[UserListField])
	d.Set(UserListGroup, result[UserListGroup])

	config, err := parseConfig(result["userlist_users"])
	if err != nil {
		return diag.FromErr(err)
	}

	configList := parseUserListConfigListResult(config)
	d.Set(ConfigField, configList)

	return nil
}

func resourceHaproxySectionUserlistUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	configs := parseUserListConfigList(d.Get(UserListField).([]interface{}))

	requestBody := map[string]interface{}{
		UserListField: configs,
		NameField:     d.Get(NameField),
		UserListGroup: d.Get(UserListGroup),
		TypeField:     "userlist",
		ServerIdField: d.Get(ServerIdField),
		ActionField:   d.Get(ActionField),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/userlist/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionUserlistRead(ctx, d, m)
}

func resourceHaproxySectionUserlistDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/userlist/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

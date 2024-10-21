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
	RetriesFiled       = "retries"
	TimeoutField       = "timeout"
	CheckField         = "check"
	ClientField        = "client"
	ConnectField       = "connect"
	HttpKeepAliveField = "http_keep_alive"
	HttpRequestField   = "http_request"
	QueueField         = "queue"
	ServerTimeoutField = "server"
)

func resourceHaproxySectionDefaults() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout:   resourceHaproxySectionDefaultsRead,
		UpdateWithoutTimeout: resourceHaproxySectionDefaultsUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionDefaultsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Defaults sections. Since this section exists by default and cannot be deleted, it can only be imported and edited. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			LogField: {
				Type:        schema.TypeString,
				Description: "A list loging settings.",
				Optional:    true,
				Default:     "global",
			},
			TimeoutField: {
				Type:        schema.TypeSet,
				Description: "A Set of timeout settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						CheckField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "IP address of the backend server.",
						},
						ClientField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     60,
							Description: "Port number on which the backend server listens for requests.",
						},
						ConnectField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "Weight assigned to the backend server.",
						},
						HttpKeepAliveField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "Weight assigned to the backend server.",
						},
						HttpRequestField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     10,
							Description: "Weight assigned to the backend server.",
						},
						QueueField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     60,
							Description: "Weight assigned to the backend server.",
						},
						ServerTimeoutField: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     60,
							Description: "Weight assigned to the backend server.",
						},
					},
				},
			},
			ServerIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The ID of the server to deploy to.",
			},
			MaxconnFiled: {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5000,
				Description: "Limits the per-process connection limit.",
			},
			ActionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "What action should be taken after changing the config. Available: save, reload, restart.",
				Default:     "save",
			},
			OptionFiled: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Here you can put addinional options separeted by '\n'.",
			},
			RetriesFiled: {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Set the number of retries to perform on a server after a failure.",
			},
		},
	}
}

func resourceHaproxySectionDefaultsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/defaults", serverId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	if err = setTimeoutField(d, "timeout", result["timeout"]); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Success:", d.Get("timeout"))
	}

	d.Set(MaxconnFiled, intFromInterface(result[MaxconnFiled]))
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(RetriesFiled, intFromInterface(result[RetriesFiled]))
	d.Set(LogField, result[LogField])
	d.Set(OptionFiled, result[OptionFiled])
	d.Set(ActionField, result[ActionField])

	return nil
}

func resourceHaproxySectionDefaultsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	timeouts, errs := getTimeoutMap(d, "timeout")
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		MaxconnFiled:  d.Get(MaxconnFiled),
		LogField:      d.Get(LogField),
		TypeField:     "defaults",
		ServerIdField: d.Get(ServerIdField),
		OptionFiled:   d.Get(OptionFiled),
		RetriesFiled:  d.Get(RetriesFiled),
		ActionField:   d.Get(ActionField),
		TimeoutField:  timeouts,
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/defaults", serverId), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionDefaultsRead(ctx, d, m)
}

func resourceHaproxySectionDefaultsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

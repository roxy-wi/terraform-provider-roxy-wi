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
	DaemonField  = "daemon"
	LogField     = "log"
	MaxconnFiled = "maxconn"
	OptionFiled  = "option"
	PidFileFiled = "pidfile"
	SocketFiled  = "socket"
	ChrootField  = "chroot"
	ActionField  = "action"
)

func resourceHaproxySectionGlobal() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout:   resourceHaproxySectionGlobalRead,
		UpdateWithoutTimeout: resourceHaproxySectionGlobalUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionGlobalDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Global sections. Since this section exists by default and cannot be deleted, it can only be imported and edited. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			LogField: {
				Type:        schema.TypeList,
				Description: "A list loging settings.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "['127.0.0.1 local1','127.0.0.1 local1 notice']",
				},
			},
			SocketFiled: {
				Type:        schema.TypeList,
				Description: "A list socket settings.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:    schema.TypeString,
					Default: "['*:1999 level admin','/var/run/haproxy.sock mode 600 level admin','/var/lib/haproxy/stats']",
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
			PidFileFiled: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "/var/run/haproxy.pid",
				Description: "Path to the pid file.",
			},
			UserFiled: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "haproxy",
				Description: "A user with what HAProxy will be started.",
			},
			GroupNameField: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "haproxy",
				Description: "A group with what HAProxy will be started.",
			},
			ChrootField: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "haproxy",
				Description: "HAProxy is designed to isolate itself into a chroot jail during startup, where\nit cannot perform any file-system access at all.",
			},
			DaemonField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Start as a daemon. The process detaches from the current terminal after forking, and errors are not reported anymore in the terminal.",
			},
		},
	}
}

func resourceHaproxySectionGlobalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/global", serverId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(MaxconnFiled, intFromInterface(result[MaxconnFiled]))
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(LogField, result[LogField])
	d.Set(SocketFiled, result[SocketFiled])
	d.Set(OptionFiled, result[OptionFiled])
	d.Set(PidFileFiled, result[PidFileFiled])
	d.Set(DaemonField, result[DaemonField])
	d.Set(UserFiled, result[UserFiled])
	d.Set(GroupNameField, result[GroupNameField])
	d.Set(ChrootField, result[ChrootField])

	return nil
}

func resourceHaproxySectionGlobalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)

	requestBody := map[string]interface{}{
		MaxconnFiled:   d.Get(MaxconnFiled),
		LogField:       d.Get(LogField),
		TypeField:      "global",
		ServerIdField:  d.Get(ServerIdField),
		SocketFiled:    d.Get(SocketFiled),
		OptionFiled:    d.Get(OptionFiled),
		PidFileFiled:   d.Get(PidFileFiled),
		DaemonField:    d.Get(DaemonField),
		UserFiled:      d.Get(UserFiled),
		GroupNameField: d.Get(GroupNameField),
		ChrootField:    d.Get(ChrootField),
		ActionField:    d.Get(ActionField),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/global", serverId), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionGlobalRead(ctx, d, m)
}

func resourceHaproxySectionGlobalDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/global", serverId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

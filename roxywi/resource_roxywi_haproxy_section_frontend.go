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

func resourceHaproxySectionFrontend() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxySectionFrontendCreate,
		ReadWithoutTimeout:   resourceHaproxySectionFrontendRead,
		UpdateWithoutTimeout: resourceHaproxySectionFrontendUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionFrontendDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Frontend sections. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Frontend section.",
			},
			UseBackendField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default backend to use.",
			},
			AclsField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of ACLs configuration.",
				Elem: &schema.Resource{
					Schema: aclSchema(),
				},
			},
			ModeField: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ModeHTTP,
				Description: fmt.Sprintf("Load balancing mode. Available values are: `%s`,`%s`.", ModeTCP, ModeHTTP),
				ValidateFunc: validation.StringInSlice([]string{
					ModeTCP,
					ModeHTTP,
				}, false),
			},
			BindsField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of backend servers configuration.",
				Elem: &schema.Resource{
					Schema: bindSchema(),
				},
			},
			ServerIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the server to deploy to.",
			},
			MaxconnFiled: {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2000,
				Description: "Limits the per-process connection limit.",
			},
			ActionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "What action should be taken after changing the config. Available: save, reload, restart.",
				Default:     "save",
			},
			BlacklistField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a blacklist.",
			},
			WhitelistField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Path to a whitelist.",
			},
			CacheField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Cache enabling.",
			},
			SlowAttackField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "In a Slow POST attack, an attacker begins by sending a legitimate HTTP POST header to a Web server, exactly as they would under normal circumstances. The header specifies the exact size of the message body that will then follow. However, that message body is then sent at an alarmingly low rate – sometimes as slow as 1 byte per approximately two minutes.",
			},
			ForwardForField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When HAProxy Enterprise proxies a TCP connection, it overwrites the client's source IP address with its own when communicating with the backend server. However, when relaying HTTP messages, it can store the client's address in the HTTP header X-Forwarded-For. The backend server can then be configured to read the value from that header to retrieve the client's IP address.",
			},
			SslOffloadingField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enable redirection from HTTP scheme to HTTPS scheme.",
			},
			CompressionField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "HTTP compression allows you to shrink the body of a response before it is relayed to a client, which results in using less network bandwidth per request. From a client's perspective, this reduces latency.",
			},
			AntiBotField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Add Anti Bot settings.",
			},
			DdosField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "DDOS attack protect.",
			},
			WafField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Add WAF settings.",
			},
			SslField: {
				Type:        schema.TypeSet,
				Description: "SSL settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: sslSchema(),
				},
			},
			HeadersField: {
				Type:        schema.TypeList,
				Description: "Set custom check parameters.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: headersSchema(),
				},
			},
		},
	}
}

func resourceHaproxySectionFrontendCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	//if err := validateModeAndOptions(d); err != nil {
	//	return diag.FromErr(err)
	//}

	binds := parseUserBindsList(d.Get(BindsField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
	headers := parseHeaderList(d.Get(HeadersField).([]interface{}))
	ssl, errs := getSetMap(d, SslField)
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		BindsField:         binds,
		UseBackendField:    d.Get(UseBackendField),
		AclsField:          acls,
		HeadersField:       headers,
		NameField:          d.Get(NameField),
		TypeField:          "frontend",
		ServerIdField:      d.Get(ServerIdField),
		ActionField:        d.Get(ActionField),
		BlacklistField:     d.Get(BlacklistField),
		WhitelistField:     d.Get(WhitelistField),
		ModeField:          d.Get(ModeField),
		SslField:           ssl,
		CacheField:         d.Get(CacheField),
		CompressionField:   d.Get(CompressionField),
		ForwardForField:    d.Get(ForwardForField),
		SslOffloadingField: d.Get(SslOffloadingField),
		SlowAttackField:    d.Get(SlowAttackField),
		AntiBotField:       d.Get(AntiBotField),
		DdosField:          d.Get(DdosField),
		WafField:           d.Get(WafField),
		MaxconnFiled:       d.Get(MaxconnFiled),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/haproxy/%d/section/frontend", d.Get(ServerIdField)), requestBody)
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
	return resourceHaproxySectionFrontendRead(ctx, d, m)
}

func resourceHaproxySectionFrontendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]
	sectionName := strings.Join(parts[1:], "-")

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/frontend/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	d.Set(NameField, result[NameField])
	d.Set(UseBackendField, result[UseBackendField])
	d.Set(ServerIdField, intFromInterface(result[ServerIdField]))
	d.Set(BlacklistField, result[BlacklistField])
	d.Set(WhitelistField, result[WhitelistField])
	d.Set(ModeField, result[ModeField])
	d.Set(CacheField, result[CacheField])
	d.Set(CompressionField, result[CompressionField])
	d.Set(ForwardForField, result[ForwardForField])
	d.Set(SslOffloadingField, result[SslOffloadingField])
	d.Set(SlowAttackField, result[SlowAttackField])
	d.Set(AntiBotField, result[AntiBotField])
	d.Set(DdosField, result[DdosField])
	d.Set(WafField, result[WafField])
	d.Set(MaxconnFiled, result[MaxconnFiled])

	if err = setTimeoutField(d, SslField, result[SslField]); err != nil {
		fmt.Println("Error:", err)
	}

	binds, err := parseConfig(result[BindsField])
	if err != nil {
		return diag.FromErr(err)
	}

	acl, err := parseConfig(result[AclsField])
	if err != nil {
		return diag.FromErr(err)
	}
	header, err := parseConfig(result[HeadersField])
	if err != nil {
		return diag.FromErr(err)
	}

	bindsList := parseBindsResult(binds)
	acls := parseAclsServerResult(acl)
	headers := parseHeadersResult(header)
	d.Set(BindsField, bindsList)
	d.Set(AclsField, acls)
	d.Set(HeadersField, headers)

	return nil
}

func resourceHaproxySectionFrontendUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	//if err := validateModeAndOptions(d); err != nil {
	//	return diag.FromErr(err)
	//}
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)
	binds := parseUserBindsList(d.Get(BindsField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
	header := parseHeaderList(d.Get(HeadersField).([]interface{}))
	ssl, errs := getSetMap(d, SslField)
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		BindsField:         binds,
		UseBackendField:    d.Get(UseBackendField),
		AclsField:          acls,
		HeadersField:       header,
		NameField:          d.Get(NameField),
		TypeField:          "frontend",
		ServerIdField:      d.Get(ServerIdField),
		ActionField:        d.Get(ActionField),
		BlacklistField:     d.Get(BlacklistField),
		WhitelistField:     d.Get(WhitelistField),
		ModeField:          d.Get(ModeField),
		SslField:           ssl,
		CacheField:         d.Get(CacheField),
		CompressionField:   d.Get(CompressionField),
		ForwardForField:    d.Get(ForwardForField),
		SslOffloadingField: d.Get(SslOffloadingField),
		SlowAttackField:    d.Get(SlowAttackField),
		AntiBotField:       d.Get(AntiBotField),
		DdosField:          d.Get(DdosField),
		WafField:           d.Get(WafField),
		MaxconnFiled:       d.Get(MaxconnFiled),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/frontend/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionFrontendRead(ctx, d, m)
}

func resourceHaproxySectionFrontendDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/frontend/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

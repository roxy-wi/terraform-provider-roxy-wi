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

func resourceHaproxySectionBackend() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxySectionBackendCreate,
		ReadWithoutTimeout:   resourceHaproxySectionBackendRead,
		UpdateWithoutTimeout: resourceHaproxySectionBackendUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionBackendDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Backend sections. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Backend section.",
			},
			AclsField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of ACLs configuration.",
				Elem: &schema.Resource{
					Schema: aclSchema(),
				},
			},
			BackendServersField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of backend servers configuration.",
				Elem: &schema.Resource{
					Schema: backendServerSchema(),
				},
			},
			BalanceField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Load balancing algorithm. Available values are: `%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`.", RoundRobinAlgorithm, SourceAlgorithm, LeastconnAlgorithm, FirstAlgorithm, RdpCookieAlgorithm, UriAlgorithmAlgorithm, UriWholeAlgorithm, StaticRrAlgorithm, StaticRrAlgorithm, UrlParamUseridAlgorithm),
				ValidateFunc: validation.StringInSlice([]string{
					RoundRobinAlgorithm,
					SourceAlgorithm,
					LeastconnAlgorithm,
					FirstAlgorithm,
					RdpCookieAlgorithm,
					UriAlgorithmAlgorithm,
					UriWholeAlgorithm,
					StaticRrAlgorithm,
					StaticRrAlgorithm,
					UrlParamUseridAlgorithm,
				}, false),
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
			CacheField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Cache enabling.",
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
			RedisPatchField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "In HTTP mode, if a server designated by a cookie is down, clients may definitely stick to it because they cannot flush the cookie, so they will not be able to access the service anymore. Specifying 'option redispatch' will allow the proxy to break their persistence and redistribute them to a working server. It also allows to retry connections to another server in case of multiple connection failures. Of course, it requires having 'retries' set to a nonzero value.",
			},
			CompressionField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "HTTP compression allows you to shrink the body of a response before it is relayed to a client, which results in using less network bandwidth per request. From a client's perspective, this reduces latency.",
			},
			CircuitBreakingField: {
				Type:        schema.TypeSet,
				Description: "A Set of timeout settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: circuitBreakingSchema(),
				},
			},
			ServersCheckField: {
				Type:        schema.TypeSet,
				Description: "Set custom check parameters.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: serversCheckSchema(),
				},
			},
			SslField: {
				Type:        schema.TypeSet,
				Description: "SSL settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: sslSchema(),
				},
			},
			CookieField: {
				Type:        schema.TypeSet,
				Description: "To send a client to the same server where they were sent previously in order to reuse a session on that server, you can enable cookie-based session persistence. Add a cookie directive to the backend section and set the cookie parameter to a unique value on each server line.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: cookieSchema(),
				},
			},
			HealthCheckField: {
				Type:        schema.TypeSet,
				Description: "Set custom check parameters.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: healthCheckSchema(),
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

func resourceHaproxySectionBackendCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	if err := validateModeAndOptions(d); err != nil {
		return diag.FromErr(err)
	}
	backends := parseBackendsServerList(d.Get(BackendServersField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
	headers := parseHeaderList(d.Get(HeadersField).([]interface{}))
	circuitBreaking, errs := getSetMap(d, CircuitBreakingField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	serversCheck, errs := getSetMap(d, ServersCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	ssl, errs := getSetMap(d, SslField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	healthCheck, errs := getSetMap(d, HealthCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	cookie, errs := getSetMap(d, CookieField)
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		BackendServersField:  backends,
		AclsField:            acls,
		HeadersField:         headers,
		NameField:            d.Get(NameField),
		TypeField:            "backend",
		ServerIdField:        d.Get(ServerIdField),
		ActionField:          d.Get(ActionField),
		BalanceField:         d.Get(BalanceField),
		BlacklistField:       d.Get(BlacklistField),
		WhitelistField:       d.Get(WhitelistField),
		ModeField:            d.Get(ModeField),
		CircuitBreakingField: circuitBreaking,
		ServersCheckField:    serversCheck,
		SslField:             ssl,
		HealthCheckField:     healthCheck,
		CookieField:          cookie,
		CacheField:           d.Get(CacheField),
		CompressionField:     d.Get(CompressionField),
		ForwardForField:      d.Get(ForwardForField),
		SslOffloadingField:   d.Get(SslOffloadingField),
		RedisPatchField:      d.Get(RedisPatchField),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/haproxy/%d/section/backend", d.Get(ServerIdField)), requestBody)
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
	return resourceHaproxySectionBackendRead(ctx, d, m)
}

func resourceHaproxySectionBackendRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]
	sectionName := strings.Join(parts[1:], "-")

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/backend/%s", serverId, sectionName), nil)
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
	d.Set(BlacklistField, result[BlacklistField])
	d.Set(WhitelistField, result[WhitelistField])
	d.Set(ModeField, result[ModeField])
	d.Set(CacheField, result[CacheField])
	d.Set(CompressionField, result[CompressionField])
	d.Set(ForwardForField, result[ForwardForField])
	d.Set(SslOffloadingField, result[SslOffloadingField])
	d.Set(RedisPatchField, result[RedisPatchField])

	if err = setTimeoutField(d, CircuitBreakingField, result[CircuitBreakingField]); err != nil {
		fmt.Println("Error:", err)
	}
	if err = setTimeoutField(d, ServersCheckField, result[ServersCheckField]); err != nil {
		fmt.Println("Error:", err)
	}
	if err = setTimeoutField(d, SslField, result[SslField]); err != nil {
		fmt.Println("Error:", err)
	}
	if err = setTimeoutField(d, HealthCheckField, result[HealthCheckField]); err != nil {
		fmt.Println("Error:", err)
	}

	if err = setTimeoutField(d, CookieField, result[CookieField]); err != nil {
		fmt.Println("Error:", err)
	}

	backendServers, err := parseConfig(result[BackendServersField])
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

	backendServersList := parseBackendServerResult(backendServers)
	acls := parseAclsServerResult(acl)
	headers := parseHeadersResult(header)
	d.Set(BackendServersField, backendServersList)
	d.Set(AclsField, acls)
	d.Set(HeadersField, headers)

	_ = setTimeoutField(d, CircuitBreakingField, result[CircuitBreakingField])

	return nil
}

func resourceHaproxySectionBackendUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	if err := validateModeAndOptions(d); err != nil {
		return diag.FromErr(err)
	}
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)
	backends := parseBackendsServerList(d.Get(BackendServersField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
	header := parseHeaderList(d.Get(HeadersField).([]interface{}))
	circuitBreaking, errs := getSetMap(d, CircuitBreakingField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	serversCheck, errs := getSetMap(d, ServersCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	ssl, errs := getSetMap(d, SslField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	healthCheck, errs := getSetMap(d, HealthCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	cookie, errs := getSetMap(d, CookieField)
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		BackendServersField:  backends,
		AclsField:            acls,
		HeadersField:         header,
		NameField:            d.Get(NameField),
		TypeField:            "backend",
		ServerIdField:        d.Get(ServerIdField),
		ActionField:          d.Get(ActionField),
		ModeField:            d.Get(ModeField),
		CircuitBreakingField: circuitBreaking,
		ServersCheckField:    serversCheck,
		SslField:             ssl,
		HealthCheckField:     healthCheck,
		CookieField:          cookie,
		CacheField:           d.Get(CacheField),
		CompressionField:     d.Get(CompressionField),
		ForwardForField:      d.Get(ForwardForField),
		SslOffloadingField:   d.Get(SslOffloadingField),
		RedisPatchField:      d.Get(RedisPatchField),
		BalanceField:         d.Get(BalanceField),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/backend/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionBackendRead(ctx, d, m)
}

func resourceHaproxySectionBackendDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/backend/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

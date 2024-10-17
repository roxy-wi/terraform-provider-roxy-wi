package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	AclsField                      = "acls"
	AclIfField                     = "acl_if"
	AclThenField                   = "acl_then"
	AclThenValueField              = "acl_then_value"
	AclValueField                  = "acl_value"
	BackendServersField            = "backend_servers"
	BackendServersBackupField      = "backup"
	BackendServersPortCheckField   = "port_check"
	BackendServersSendProxyField   = "send_proxy"
	BalanceField                   = "balance"
	BindsField                     = "binds"
	BlacklistField                 = "blacklist"
	WhitelistField                 = "whitelist"
	CacheField                     = "cache"
	CircuitBreakingField           = "circuit_breaking"
	CircuitBreakingErrorLimitField = "error_limit"
	CircuitBreakingObserveField    = "observe"
	CircuitBreakingOnerrorField    = "on_error"
	CompressionField               = "compression"
	CookieField                    = "cookie"
	CookieDomainField              = "domain"
	CookieDynamicField             = "dynamic"
	CookieDynamicKeyField          = "dynamic_key"
	CookieNameField                = "name"
	CookieNocacheField             = "nocache"
	CookiePostonlyField            = "postonly"
	CookiePrefixField              = "prefix"
	ForwardForField                = "forward_for"
	HeadersField                   = "headers"
	HealthCheckField               = "health_check"
	HealthCheckTypeField           = "check"
	HealthCheckDomainField         = "domain"
	HealthCheckPathField           = "path"
	ModeField                      = "mode"
	ModeTCP                        = "tcp"
	ModeHTTP                       = "http"
	ServersCheckField              = "servers_check"
	ServersCheckFallField          = "fall"
	ServersCheckInterField         = "inter"
	ServersCheckRiseField          = "rise"
	SlowAttackField                = "slow_attack"
	SslField                       = "ssl"
	SslCertField                   = "cert"
	SslCheckField                  = "ssl_check_backend"
	SslOffloadingField             = "ssl_offloading"
	RedisPatchField                = "redispatch"
	RoundRobinAlgorithm            = "roundrobin"
	SourceAlgorithm                = "source"
	LeastconnAlgorithm             = "leastconn"
	FirstAlgorithm                 = "first"
	RdpCookieAlgorithm             = "rdp-cookie"
	UriAlgorithmAlgorithm          = "uri"
	UriWholeAlgorithm              = "uri whole"
	StaticRrAlgorithm              = "static-rr"
	UrlParamUseridAlgorithm        = "url_param userid"
)

func resourceHaproxySectionListen() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHaproxySectionListenCreate,
		ReadWithoutTimeout:   resourceHaproxySectionListenRead,
		UpdateWithoutTimeout: resourceHaproxySectionListenUpdate,
		DeleteWithoutTimeout: resourceHaproxySectionListenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage HAProxy Listen sections. Please note that changes may cause HAProxy to restart.",

		Schema: map[string]*schema.Schema{
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the Listen section.",
			},
			AclsField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of ACLs configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						AclIfField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "If statment: 1: 'hdr_beg(host) -i', 2: 'hdr_end(host) -i', 3: 'path_beg -i', 4: 'path_end -i', 6: 'src ip'.",
						},
						AclValueField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "If value.",
						},
						AclThenField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Then statment: 2: 'http-request redirect location', 3:'http-request allow', 4: 'http-request deny', 5: 'use_backend'.",
						},
						AclThenValueField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Then value.",
						},
					},
				},
			},
			BackendServersField: {
				Type:        schema.TypeList,
				Optional:    true,
				Default:     nil,
				Description: "List of backend servers configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						ServerTimeoutField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Backend server address.",
						},
						BackendPortField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Backend server port.",
						},
						BackendServersPortCheckField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Backend port for check. Usually the same as the backend_port.",
						},
						MaxconnFiled: {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     200,
							Description: "Maximun connection to the server.",
						},
						BackendServersSendProxyField: {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable Send proxy option for this backend server. Only for HTTP mode.",
						},
						BackendServersBackupField: {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Is this server backup server?",
						},
					},
				},
			},
			BalanceField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Load balancing algorithm. Available values are: `%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`,`%s`.", RoundRobinAlgorithm, SourceAlgorithm, LeastconnAlgorithm, FirstAlgorithm, RdpCookieAlgorithm, UriAlgorithmAlgorithm, UriWholeAlgorithm, StaticRrAlgorithm, StaticRrAlgorithm, UrlParamUseridAlgorithm),
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					v, ok := i.(string)
					if !ok {
						return diag.Errorf("Invalid type for %s: expected string", path)
					}
					algorithms := []string{RoundRobinAlgorithm, SourceAlgorithm, LeastconnAlgorithm, FirstAlgorithm, RdpCookieAlgorithm, UriAlgorithmAlgorithm, UriWholeAlgorithm, StaticRrAlgorithm, StaticRrAlgorithm, UrlParamUseridAlgorithm}
					for _, alg := range algorithms {
						if v == alg {
							return nil
						}
					}
					return diag.Errorf("Invalid value for %s: %s. Valid values are: %s", path, v, strings.Join(algorithms, ", "))
				},
			},
			ModeField: {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ModeHTTP,
				Description: fmt.Sprintf("Load balancing mode. Available values are: `%s`,`%s`.", ModeTCP, ModeHTTP),
				ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
					v, ok := i.(string)
					if !ok {
						return diag.Errorf("Invalid type for %s: expected string", path)
					}
					algorithms := []string{ModeTCP, ModeHTTP}
					for _, alg := range algorithms {
						if v == alg {
							return nil
						}
					}
					return diag.Errorf("Invalid value for %s: %s. Valid values are: %s", path, v, strings.Join(algorithms, ", "))
				},
			},
			BindsField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of backend servers configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						IPField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "IP for binding listener.",
						},
						PortField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Port for binding listener.",
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
				Description: "Cache enabling.",
			},
			SlowAttackField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "In a Slow POST attack, an attacker begins by sending a legitimate HTTP POST header to a Web server, exactly as they would under normal circumstances. The header specifies the exact size of the message body that will then follow. However, that message body is then sent at an alarmingly low rate – sometimes as slow as 1 byte per approximately two minutes.",
			},
			ForwardForField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "When HAProxy Enterprise proxies a TCP connection, it overwrites the client's source IP address with its own when communicating with the backend server. However, when relaying HTTP messages, it can store the client's address in the HTTP header X-Forwarded-For. The backend server can then be configured to read the value from that header to retrieve the client's IP address.",
			},
			SslOffloadingField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable redirection from HTTP scheme to HTTPS scheme.",
			},
			RedisPatchField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "In HTTP mode, if a server designated by a cookie is down, clients may definitely stick to it because they cannot flush the cookie, so they will not be able to access the service anymore. Specifying 'option redispatch' will allow the proxy to break their persistence and redistribute them to a working server. It also allows to retry connections to another server in case of multiple connection failures. Of course, it requires having 'retries' set to a nonzero value.",
			},
			CompressionField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "HTTP compression allows you to shrink the body of a response before it is relayed to a client, which results in using less network bandwidth per request. From a client's perspective, this reduces latency.",
			},
			CircuitBreakingField: {
				Type:        schema.TypeSet,
				Description: "A Set of timeout settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						CircuitBreakingErrorLimitField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of errors considered as the threshold.",
						},
						CircuitBreakingObserveField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Monitor live traffic for HTTP or TCP. Available values are: `layer7`,`layer4`.",
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								v, ok := i.(string)
								if !ok {
									return diag.Errorf("Invalid type for %s: expected string", path)
								}
								algorithms := []string{"layer7", "layer4"}
								for _, alg := range algorithms {
									if v == alg {
										return nil
									}
								}
								return diag.Errorf("Invalid value for %s: %s. Valid values are: %s", path, v, strings.Join(algorithms, ", "))
							},
						},
						CircuitBreakingOnerrorField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Actions. Available values are: `mark-down`,`fastinter`, `fail-check`, `sudden-death`.",
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								v, ok := i.(string)
								if !ok {
									return diag.Errorf("Invalid type for %s: expected string", path)
								}
								algorithms := []string{"mark-down", "fastinter", "fail-check", "sudden-death"}
								for _, alg := range algorithms {
									if v == alg {
										return nil
									}
								}
								return diag.Errorf("Invalid value for %s: %s. Valid values are: %s", path, v, strings.Join(algorithms, ", "))
							},
						},
					},
				},
			},
			ServersCheckField: {
				Type:        schema.TypeSet,
				Description: "Set custom check parameters.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						ServersCheckFallField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The 'fall' parameter states that a server will be considered as dead after <count> consecutive unsuccessful health checks. This value defaults to 5 if unspecified.",
						},
						ServersCheckInterField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The \"inter\" parameter sets the interval between two consecutive health checks to <delay> milliseconds. If left unspecified, the delay defaults to 2000 ms.",
						},
						ServersCheckRiseField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The 'rise' parameter states that a server will be considered as operational after <count> consecutive successful health checks. This value defaults to 2 if unspecified.",
						},
					},
				},
			},
			SslField: {
				Type:        schema.TypeSet,
				Description: "SSL settings.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						SslCertField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path to the pem file.",
						},
						SslCheckField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Disable SSL verify on servers.",
						},
					},
				},
			},
			CookieField: {
				Type:        schema.TypeSet,
				Description: "To send a client to the same server where they were sent previously in order to reuse a session on that server, you can enable cookie-based session persistence. Add a cookie directive to the backend section and set the cookie parameter to a unique value on each server line.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						CookieDomainField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "This option allows to specify the domain at which a cookie is inserted. It requires exactly one parameter: a valid domain name.",
						},
						CookieDynamicField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Set the dynamic cookie secret key for a backend.",
						},
						CookieDynamicKeyField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Activate dynamic cookies. When used, a session cookie is dynamically created for each server, based on the IP and port of the server, and a secret key, specified in the \"dynamic-cookie-key\" backend directive.",
						},
						CookieNameField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Is the name of the cookie which will be monitored, modified or inserted in order to bring persistence.",
						},
						CookieNocacheField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "This option is recommended in conjunction with the insert mode when there is a cache between the client and HAProxy, as it ensures that a cacheable response will be tagged non-cacheable if a cookie needs to be inserted.",
						},
						CookiePostonlyField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "This option ensures that cookie insertion will only be performed on responses to POST requests. It is an alternative to the \"nocache\" option, because POST responses are not cacheable, so this ensures that the persistence cookie will never get cached.",
						},
						CookiePrefixField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "This keyword indicates that instead of relying on a dedicated cookie for the persistence, an existing one will be completed.",
						},
					},
				},
			},
			HealthCheckField: {
				Type:        schema.TypeSet,
				Description: "Set custom check parameters.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						HealthCheckTypeField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Custome check type. Could be: tcp-check, ssl-hello-chk, httpchk, ldap-check, mysql-check, pgsql-check, redis-check, smtpchk",
							ValidateDiagFunc: func(i interface{}, path cty.Path) diag.Diagnostics {
								v, ok := i.(string)
								if !ok {
									return diag.Errorf("Invalid type for %s: expected string", path)
								}
								algorithms := []string{"tcp-check", "ssl-hello-chk", "httpchk", "ldap-check", "mysql-check", "pgsql-check", "redis-check", "smtpchk"}
								for _, alg := range algorithms {
									if v == alg {
										return nil
									}
								}
								return diag.Errorf("Invalid value for %s: %s. Valid values are: %s", path, v, strings.Join(algorithms, ", "))
							},
						},
						HealthCheckDomainField: {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Domain name. Only for HTTP check.",
						},
						HealthCheckPathField: {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "/",
							Description: "URI path for checking. Only for HTTP check.",
						},
					},
				},
			},
		},
	}
}

func resourceHaproxySectionListenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	binds := parseUserBindsList(d.Get(BindsField).([]interface{}))
	backends := parseBackendsServerList(d.Get(BackendServersField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
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
		BindsField:           binds,
		BackendServersField:  backends,
		AclsField:            acls,
		NameField:            d.Get(NameField),
		TypeField:            "listen",
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
		CacheField:           boolToInt(d.Get(CacheField).(bool)),
		CompressionField:     boolToInt(d.Get(CompressionField).(bool)),
		ForwardForField:      boolToInt(d.Get(ForwardForField).(bool)),
		SslOffloadingField:   boolToInt(d.Get(SslOffloadingField).(bool)),
		SlowAttackField:      boolToInt(d.Get(SlowAttackField).(bool)),
	}

	resp, err := client.doRequest("POST", fmt.Sprintf("api/service/haproxy/%d/section/listen", d.Get(ServerIdField)), requestBody)
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
	return resourceHaproxySectionListenRead(ctx, d, m)
}

func resourceHaproxySectionListenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	fullID := d.Id()
	parts := strings.Split(fullID, "-")
	if len(parts) < 2 {
		return diag.Errorf("expected ID in the format 'server_id-section_name', got: %s", fullID)
	}
	serverId := parts[0]
	sectionName := parts[1]

	resp, err := client.doRequest("GET", fmt.Sprintf("api/service/haproxy/%s/section/listen/%s", serverId, sectionName), nil)
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
	d.Set(CacheField, intToBool(result[ModeField].(float64)))
	d.Set(CompressionField, intToBool(result[CompressionField].(float64)))
	d.Set(ForwardForField, intToBool(result[ForwardForField].(float64)))
	d.Set(SslOffloadingField, intToBool(result[SslOffloadingField].(float64)))
	d.Set(SlowAttackField, intToBool(result[SlowAttackField].(float64)))

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

	binds, err := parseConfig(result[BindsField])
	if err != nil {
		return diag.FromErr(err)
	}

	backendServers, err := parseConfig(result[BackendServersField])
	if err != nil {
		return diag.FromErr(err)
	}

	bindsList := parseBindsResult(binds)
	backendServersList := parseBackendServerResult(backendServers)
	d.Set(BindsField, bindsList)
	d.Set(BackendServersField, backendServersList)

	_ = setTimeoutField(d, CircuitBreakingField, result[CircuitBreakingField])

	return nil
}

func resourceHaproxySectionListenUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)
	binds := parseUserBindsList(d.Get(BindsField).([]interface{}))
	backends := parseBackendsServerList(d.Get(BackendServersField).([]interface{}))
	acls := parseAclsList(d.Get(AclsField).([]interface{}))
	circuitBreaking, errs := getTimeoutMap(d, CircuitBreakingField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	serversCheck, errs := getTimeoutMap(d, ServersCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	ssl, errs := getTimeoutMap(d, SslField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	healthCheck, errs := getTimeoutMap(d, HealthCheckField)
	if errs != nil {
		return diag.FromErr(errs)
	}
	cookie, errs := getTimeoutMap(d, CookieField)
	if errs != nil {
		return diag.FromErr(errs)
	}

	requestBody := map[string]interface{}{
		BindsField:           binds,
		BackendServersField:  backends,
		AclsField:            acls,
		NameField:            d.Get(NameField),
		TypeField:            "listen",
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
		CacheField:           boolToInt(d.Get(CacheField).(bool)),
		CompressionField:     boolToInt(d.Get(CompressionField).(bool)),
		ForwardForField:      boolToInt(d.Get(ForwardForField).(bool)),
		SslOffloadingField:   boolToInt(d.Get(SslOffloadingField).(bool)),
		SlowAttackField:      boolToInt(d.Get(SlowAttackField).(bool)),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("api/service/haproxy/%d/section/listen/%s", serverId, sectionName), requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaproxySectionListenRead(ctx, d, m)
}

func resourceHaproxySectionListenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	serverId := d.Get(ServerIdField)
	sectionName := d.Get(NameField)

	_, err := client.doRequest("DELETE", fmt.Sprintf("api/service/haproxy/%d/section/listen/%s", serverId, sectionName), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

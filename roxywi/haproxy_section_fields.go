package roxywi

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	PathField                      = "path"
	MethodField                    = "method"
	HeaderNameField                = "name"
	ValueField                     = "value"
	HealthCheckField               = "health_check"
	HealthCheckTypeField           = "check"
	HealthCheckDomainField         = "domain"
	HealthCheckPathField           = "path"
	ModeField                      = "mode"
	ModeLog                        = "log"
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
	AntiBotField                   = "antibot"
	DdosField                      = "ddos"
	WafField                       = "waf"
	UseBackendField                = "backends"
)

func backendServerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		ServerTimeoutField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Backend server address.",
		},
		BackendPortField: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Backend server port.",
			ValidateFunc: validation.IsPortNumber,
		},
		BackendServersPortCheckField: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Backend port for check. Usually the same as the backend_port.",
			ValidateFunc: validation.IsPortNumber,
		},
		MaxconnFiled: {
			Type:        schema.TypeInt,
			Optional:    true,
			Default:     2000,
			Description: "Maximum connection to the server.",
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
	}
}

func aclSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AclIfField: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "If statement: 1: 'hdr_beg(host) -i', 2: 'hdr_end(host) -i', 3: 'path_beg -i', 4: 'path_end -i', 6: 'src ip'.",
			ValidateFunc: validation.IntBetween(1, 6),
		},
		AclValueField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "If value.",
		},
		AclThenField: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Then statement: 2: 'http-request redirect location', 3:'http-request allow', 4: 'http-request deny', 5: 'use_backend'.",
			ValidateFunc: validation.IntBetween(2, 5),
		},
		AclThenValueField: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Then value.",
		},
	}
}

func bindSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		IPField: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "IP for binding frontender.",
		},
		PortField: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "Port for binding frontender.",
			ValidateFunc: validation.IsPortNumber,
		},
	}
}

func sslSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func headersSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		PathField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Header. Could be: http-response, http-request",
			ValidateFunc: validation.StringInSlice([]string{
				"http-response",
				"http-request",
			}, false),
		},
		MethodField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Header method. Could be: add-header, set-header, del-header",
			ValidateFunc: validation.StringInSlice([]string{
				"add-header",
				"set-header",
				"del-header",
			}, false),
		},
		HeaderNameField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Header name",
		},
		ValueField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Header value. Leave blank if using del-header.",
		},
	}
}

func healthCheckSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		HealthCheckTypeField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Custom check type. Could be: tcp-check, ssl-hello-chk, httpchk, ldap-check, mysql-check, pgsql-check, redis-check, smtpchk",
			ValidateFunc: validation.StringInSlice([]string{
				"tcp-check",
				"ssl-hello-chk",
				"httpchk",
				"ldap-check",
				"mysql-check",
				"pgsql-check",
				"redis-check",
				"smtpchk",
			}, false),
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
	}
}

func cookieSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func serversCheckSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func circuitBreakingSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		CircuitBreakingErrorLimitField: {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of errors considered as the threshold.",
		},
		CircuitBreakingObserveField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Monitor live traffic for HTTP or TCP. Available values are: `layer7`,`layer4`.",
			ValidateFunc: validation.StringInSlice([]string{
				"layer7",
				"layer4",
			}, false),
		},
		CircuitBreakingOnerrorField: {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Actions. Available values are: `mark-down`,`fastinter`, `fail-check`, `sudden-death`.",
			ValidateFunc: validation.StringInSlice([]string{
				"mark-down",
				"fastinter",
				"sudden-death",
				"fail-check",
			}, false),
		},
	}
}

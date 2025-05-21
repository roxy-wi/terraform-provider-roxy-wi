package roxywi

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	MaxFails               = "max_fails"
	FailTimeout            = "fail_timeout"
	NginxBalanceRoundRobin = "round_robin"
	NginxBalanceIpHash     = "ip_hash"
	NginxBalanceLeastConn  = "least_conn"
	NginxBalanceRandom     = "random"
	NginxKeepAlive         = "keepalive"
)

func nginxBackendServerSchema() map[string]*schema.Schema {
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
		MaxFails: {
			Type:         schema.TypeInt,
			Required:     true,
			Description:  "How many unsuccessful connection attempts (connect, send, read, or timeout) must occur in a row before the server is considered temporarily unavailable.",
			ValidateFunc: validation.IntBetween(0, 10000),
		},
		FailTimeout: {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      2000,
			Description:  "The time during which max_fails are taken into account, and the server will be excluded from the pool after exceeding this threshold.\nAfter fail_timeout the server is considered available again.",
			ValidateFunc: validation.IntBetween(0, 10000),
		},
	}
}

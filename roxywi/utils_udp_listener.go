package roxywi

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
)

func parseConfigList(configList []interface{}) []map[string]interface{} {
	var configs []map[string]interface{}
	for _, config := range configList {
		configDetails := config.(map[string]interface{})
		configs = append(configs, map[string]interface{}{
			BackendIPField:     configDetails[BackendIPField].(string),
			BackendPortField:   intFromInterface(configDetails[BackendPortField]),
			BackendWeightField: intFromInterface(configDetails[BackendWeightField]),
		})
	}
	return configs
}

func parseConfig(config interface{}) ([]map[string]interface{}, error) {
	switch v := config.(type) {
	case string:
		var parsedConfig []map[string]interface{}
		log.Printf("[DEBUG] Config string before parsing: %s", v)
		v = strings.ReplaceAll(v, "'", "\"")
		if err := json.Unmarshal([]byte(v), &parsedConfig); err != nil {
			return nil, fmt.Errorf("failed to parse config field: %v", err)
		}
		return parsedConfig, nil
	case []interface{}:
		var parsedConfig []map[string]interface{}
		for _, item := range v {
			parsedConfig = append(parsedConfig, item.(map[string]interface{}))
		}
		return parsedConfig, nil
	default:
		return nil, fmt.Errorf("invalid config format: %v", config)
	}
}

func parseConfigResult(config []map[string]interface{}) []interface{} {
	var configList []interface{}
	for _, c := range config {
		configList = append(configList, map[string]interface{}{
			BackendIPField:     c[BackendIPField].(string),
			BackendPortField:   intFromInterface(c[BackendPortField]),
			BackendWeightField: intFromInterface(c[BackendWeightField]),
		})
	}
	return configList
}

func intFromInterface(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func validateLbAlgorithm() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		v, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", path)
		}

		for _, str := range validLbAlgorithms {
			if v == str {
				return nil
			}
		}

		return diag.Errorf("invalid value for %s: %s. Valid values are: %v", path, v, validLbAlgorithms)
	}
}

func validateVip() schema.SchemaValidateDiagFunc {
	return func(i interface{}, path cty.Path) diag.Diagnostics {
		vip, ok := i.(string)
		if !ok {
			return diag.Errorf("expected type of %s to be string", path)
		}

		if vip == "" {
			return diag.Errorf("VIP cannot be empty")
		}

		return nil
	}
}

func checkVipExists(client *Client, clusterID, serverID int, vip string) error {
	var url string
	if clusterID != 0 {
		url = fmt.Sprintf("/api/ha/cluster/%d/vips", clusterID)
	} else if serverID != 0 {
		url = fmt.Sprintf("/api/server/%d/ip", serverID)
	} else {
		return fmt.Errorf("either cluster_id or server_id must be specified")
	}

	log.Printf("[DEBUG] Checking VIP existence with URL: %s", url)

	resp, err := client.doRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}

	log.Printf("[DEBUG] Response from VIP check: %s", string(resp))

	if clusterID != 0 {
		var result []map[string]interface{}
		if err := json.Unmarshal(resp, &result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}

		for _, item := range result {
			if itemVip, ok := item["vip"].(string); ok && itemVip == vip {
				return nil
			}
		}
	} else {
		var ips []string
		if err := json.Unmarshal(resp, &ips); err != nil {
			return fmt.Errorf("failed to unmarshal response: %v", err)
		}

		log.Printf("[DEBUG] Parsed response: %v", ips)

		for _, itemVip := range ips {
			if itemVip == vip {
				return nil
			}
		}
	}

	return fmt.Errorf("VIP %s not found", vip)
}

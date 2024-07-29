package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ListenerIdField       = "id"
	CheckEnabledField     = "check_enabled"
	DelayBeforeRetryField = "delay_before_retry"
	DelayLoopField        = "delay_loop"
	RetryField            = "retry"
)

func dataSourceUdpListener() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUdpListenerRead,
		Description: "Data source for retrieving information about a UDP listener in Roxy-WI.",

		Schema: map[string]*schema.Schema{
			ListenerIdField: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "ID of the UDP listener.",
				ExactlyOneOf: []string{ListenerIdField, NameField},
			},
			CheckEnabledField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Check enabled field.",
			},
			ClusterIdField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the cluster.",
			},
			ConfigField: {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Configuration for the backend servers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						BackendIPField: {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address of the backend server.",
						},
						BackendPortField: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Port number on which the backend server listens for requests.",
						},
						BackendWeightField: {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Weight assigned to the backend server.",
						},
					},
				},
			},
			DelayBeforeRetryField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Delay before retry field.",
			},
			DelayLoopField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Delay loop field.",
			},
			DescriptionField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the UDP listener.",
			},
			GroupIdField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group.",
			},
			LbAlgorithmField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Load balancing algorithm.",
			},
			NameField: {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The name of the UDP listener.",
				ExactlyOneOf: []string{ListenerIdField, NameField},
			},
			PortField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port on which the UDP listener will listen.",
			},
			RetryField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Retry field.",
			},
			ServerIdField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the server.",
			},
			VIPField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual IP address.",
			},
		},
	}
}
func dataSourceUdpListenerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id, idExists := d.GetOk(ListenerIdField)
	name, nameExists := d.GetOk(NameField)

	var result map[string]interface{}
	var err error

	switch {
	case idExists:
		result, err = getListenerByID(client, id.(string))
	case nameExists:
		result, err = getListenerByName(client, name.(string))
	default:
		return diag.Errorf("Either %s or %s must be specified", ListenerIdField, NameField)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if err := setResourceDataFromResult(d, result); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getListenerByID(client *Client, id string) (map[string]interface{}, error) {
	resp, err := client.doRequest("GET", fmt.Sprintf("/api/udp/listener/%s", id), nil)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func getListenerByName(client *Client, name string) (map[string]interface{}, error) {
	resp, err := client.doRequest("GET", "/api/udp/listeners", nil)
	if err != nil {
		return nil, err
	}

	var listeners []map[string]interface{}
	if err := json.Unmarshal(resp, &listeners); err != nil {
		return nil, err
	}

	trimmedName := strings.TrimSpace(name)
	for _, listener := range listeners {
		listenerName := strings.TrimSpace(fmt.Sprintf("%v", listener[NameField]))
		if strings.Trim(listenerName, "'\"") == trimmedName {
			return listener, nil
		}
	}

	return nil, fmt.Errorf("No UDP listener found with name %s", name)
}

func setResourceDataFromResult(d *schema.ResourceData, result map[string]interface{}) error {
	id, ok := result[ListenerIdField].(float64)
	if !ok {
		return fmt.Errorf("Invalid ID type for listener")
	}
	d.SetId(fmt.Sprintf("%d", int(id)))

	setField(d, CheckEnabledField, result[CheckEnabledField])
	setField(d, ClusterIdField, result[ClusterIdField])
	setField(d, DelayBeforeRetryField, result[DelayBeforeRetryField])
	setField(d, DelayLoopField, result[DelayLoopField])
	setField(d, DescriptionField, result[DescriptionField])
	setField(d, RetryField, result[RetryField])
	setField(d, ServerIdField, result[ServerIdField])
	setField(d, VIPField, result[VIPField])
	setField(d, LbAlgorithmField, result[LbAlgorithmField])
	setField(d, NameField, strings.Trim(fmt.Sprintf("%v", result[NameField]), "'\""))
	setField(d, PortField, result[PortField])
	setField(d, GroupIdField, result[GroupIdField])

	configStr, ok := result["config"].(string)
	if !ok || configStr == "" {
		return d.Set(ConfigField, nil)
	}

	configStr = strings.ReplaceAll(configStr, "'", "\"")

	var config []map[string]interface{}
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return fmt.Errorf("Failed to parse config field: %v", err)
	}

	if len(config) == 0 {
		return d.Set(ConfigField, nil)
	}

	configSet := schema.NewSet(schema.HashResource(&schema.Resource{
		Schema: map[string]*schema.Schema{
			BackendIPField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of the backend server.",
			},
			BackendPortField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number on which the backend server listens for requests.",
			},
			BackendWeightField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Weight assigned to the backend server.",
			},
		},
	}), nil)

	for _, c := range config {
		configSet.Add(map[string]interface{}{
			BackendIPField:     getStringValue(c["backend_ip"]),
			BackendPortField:   convertToInt(c["port"]),
			BackendWeightField: convertToInt(c["weight"]),
		})
	}

	return d.Set(ConfigField, configSet)
}

func setField(d *schema.ResourceData, field string, value interface{}) {
	if value == nil {
		d.Set(field, "")
		return
	}
	switch v := value.(type) {
	case float64:
		d.Set(field, int(v))
	case int, int32, int64:
		d.Set(field, v)
	case string:
		d.Set(field, v)
	default:
		d.Set(field, v)
	}
}

func convertToInt(value interface{}) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int, int32, int64:
		return v.(int)
	default:
		return 0
	}
}

func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the UDP listener.",
			},
			CheckEnabledField: {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Check enabled field.",
			},
			ClusterIdField: {
				Type:        schema.TypeString,
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "ID of the group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the group.",
						},
						"group_id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "ID of the group.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the group.",
						},
					},
				},
			},
			LbAlgorithmField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Load balancing algorithm.",
			},
			NameField: {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the UDP listener.",
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
	id := d.Get(ListenerIdField).(string)

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/udp/listener/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Received API response: %v", result)

	d.SetId(id)
	setField(d, CheckEnabledField, result[CheckEnabledField])
	setField(d, ClusterIdField, result[ClusterIdField])
	setField(d, DelayBeforeRetryField, result[DelayBeforeRetryField])
	setField(d, DelayLoopField, result[DelayLoopField])
	setField(d, DescriptionField, result[DescriptionField])
	setField(d, RetryField, result[RetryField])
	setField(d, ServerIdField, result[ServerIdField])
	setField(d, VIPField, result[VIPField])
	setField(d, LbAlgorithmField, result[LbAlgorithmField])
	setField(d, NameField, result[NameField])
	setField(d, PortField, result[PortField])

	if group, ok := result["group_id"].(map[string]interface{}); ok {
		groupList := []interface{}{
			map[string]interface{}{
				"description": group["description"],
				"group_id":    convertToInt(group["group_id"]),
				"name":        group["name"],
			},
		}
		d.Set(GroupIdField, groupList)
	}

	if config, ok := result["config"].([]interface{}); ok {
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
			configDetails := c.(map[string]interface{})
			configSet.Add(map[string]interface{}{
				BackendIPField:     configDetails[BackendIPField].(string),
				BackendPortField:   convertToInt(configDetails[BackendPortField]),
				BackendWeightField: convertToInt(configDetails[BackendWeightField]),
			})
		}

		d.Set(ConfigField, configSet)
	}

	return nil
}

func setField(d *schema.ResourceData, field string, value interface{}) {
	if value == nil {
		return
	}
	switch v := value.(type) {
	case float64:
		d.Set(field, int(v))
	case int:
		d.Set(field, v)
	case int32:
		d.Set(field, int(v))
	case int64:
		d.Set(field, int(v))
	default:
		d.Set(field, v)
	}
}

func convertToInt(value interface{}) int {
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

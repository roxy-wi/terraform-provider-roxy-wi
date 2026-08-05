package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ReturnToMasterField = "return_master"
	ServersField        = "servers"
	MasterField         = "master"
	ServicesField       = "services"
	DockerField         = "docker"
	SynFloodField       = "syn_flood"
	UseSrcField         = "use_src"
	VirtServerField     = "virt_server"
	EthField            = "eth"
)

func resourceHaCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHaClusterCreate,
		ReadContext:   resourceHaClusterRead,
		UpdateContext: resourceHaClusterUpdate,
		DeleteContext: resourceHaClusterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Managing HA cluster resources.",

		Schema: map[string]*schema.Schema{
			DescriptionField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of the HA Cluster.",
			},
			NameField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the HA Cluster.",
			},
			ReturnToMasterField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Return to master setting for the HA Cluster.",
			},
			ServersField: {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of servers in the HA Cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						IDField: {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Server ID.",
						},
						MasterField: {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Master setting for the server.",
						},
						EthField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Ethernet interface for the server.",
						},
					},
				},
			},
			ServicesField: {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Services configuration for the HA Cluster.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						NameField: {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the service.",
							ValidateFunc: validation.StringInSlice([]string{
								"haproxy",
								"nginx",
								"apache",
							}, false),
						},
						DockerField: {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Docker setting for the service.",
						},
						EnabledField: {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Enabled status for the service.",
						},
					},
				},
			},
			SynFloodField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "SYN flood protection setting for the HA Cluster.",
			},
			UseSrcField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Use source setting for the HA Cluster.",
			},
			VIPField: {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Virtual IP address for the HA Cluster.",
				ValidateFunc: validation.IsIPAddress,
			},
			VirtServerField: {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Virtual server setting for the HA Cluster.",
			},
		},
	}
}

func resourceHaClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")
	name := strings.ReplaceAll(d.Get(NameField).(string), "'", "")

	services := d.Get(ServicesField).([]interface{})
	servicesMap := make(map[string]map[string]interface{})

	for _, service := range services {
		serviceData := service.(map[string]interface{})
		serviceName := serviceData[NameField].(string)

		servicesMap[serviceName] = map[string]interface{}{
			DockerField:  boolToInt(serviceData[DockerField].(bool)),
			EnabledField: boolToInt(serviceData[EnabledField].(bool)),
		}
	}

	servers := parseServersList(d.Get(ServersField).([]interface{}))

	haCluster := map[string]interface{}{
		DescriptionField:    description,
		NameField:           name,
		ReturnToMasterField: boolToInt(d.Get(ReturnToMasterField).(bool)),
		ServersField:        servers,
		ServicesField:       servicesMap,
		SynFloodField:       boolToInt(d.Get(SynFloodField).(bool)),
		UseSrcField:         boolToInt(d.Get(UseSrcField).(bool)),
		VIPField:            d.Get(VIPField).(string),
		VirtServerField:     boolToInt(d.Get(VirtServerField).(bool)),
		ReconfigureField:    true,
	}

	resp, err := client.doRequest(ctx, "POST", "/api/ha/cluster", haCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}

	d.SetId(fmt.Sprintf("%d", int(id)))
	return resourceHaClusterRead(ctx, d, m)
}

func resourceHaClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	id := d.Id()

	resp, err := client.doRequest(ctx, "GET", fmt.Sprintf("/api/ha/cluster/%s", id), nil)
	if err != nil {
		return readDiagnostics(d, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	servicesList, err := flattenHAServices(result[ServicesField])
	if err != nil {
		return diag.FromErr(err)
	}

	servers, err := parseConfig(result[ServersField])
	if err != nil {
		return diag.FromErr(err)
	}
	serversResult := parseServersResult(servers)

	description := strings.ReplaceAll(result[DescriptionField].(string), "'", "")
	name := strings.ReplaceAll(result[NameField].(string), "'", "")

	state := map[string]interface{}{
		DescriptionField:    description,
		NameField:           name,
		ReturnToMasterField: intFromInterface(result[ReturnToMasterField]) == 1,
		ServersField:        serversResult,
		ServicesField:       servicesList,
		SynFloodField:       intFromInterface(result[SynFloodField]) == 1,
		UseSrcField:         intFromInterface(result[UseSrcField]) == 1,
		VIPField:            result[VIPField],
		VirtServerField:     intFromInterface(result[VirtServerField]) == 1,
	}
	for field, value := range state {
		if err := d.Set(field, value); err != nil {
			return diag.Errorf("set HA cluster field %q: %v", field, err)
		}
	}

	return nil
}

func flattenHAServices(value interface{}) ([]map[string]interface{}, error) {
	servicesMap, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected %s value in HA cluster response: %T", ServicesField, value)
	}

	servicesList := make([]map[string]interface{}, 0, len(servicesMap))
	for serviceName, serviceDetails := range servicesMap {
		serviceData, ok := serviceDetails.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected service %q value in HA cluster response: %T", serviceName, serviceDetails)
		}
		servicesList = append(servicesList, map[string]interface{}{
			NameField:    serviceName,
			DockerField:  intFromInterface(serviceData[DockerField]) == 1,
			EnabledField: intFromInterface(serviceData[EnabledField]) == 1,
		})
	}
	sort.Slice(servicesList, func(i, j int) bool {
		return servicesList[i][NameField].(string) < servicesList[j][NameField].(string)
	})
	return servicesList, nil
}

func resourceHaClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")
	name := strings.ReplaceAll(d.Get(NameField).(string), "'", "")

	services := d.Get(ServicesField).([]interface{})
	servicesMap := make(map[string]map[string]interface{})

	for _, service := range services {
		serviceData := service.(map[string]interface{})
		serviceName := serviceData[NameField].(string)

		servicesMap[serviceName] = map[string]interface{}{
			DockerField:  boolToInt(serviceData[DockerField].(bool)),
			EnabledField: boolToInt(serviceData[EnabledField].(bool)),
		}
	}

	servers := parseServersList(d.Get(ServersField).([]interface{}))

	haCluster := map[string]interface{}{
		DescriptionField:    description,
		NameField:           name,
		ReturnToMasterField: boolToInt(d.Get(ReturnToMasterField).(bool)),
		ServersField:        servers,
		ServicesField:       servicesMap,
		SynFloodField:       boolToInt(d.Get(SynFloodField).(bool)),
		UseSrcField:         boolToInt(d.Get(UseSrcField).(bool)),
		VIPField:            d.Get(VIPField).(string),
		VirtServerField:     boolToInt(d.Get(VirtServerField).(bool)),
	}

	if d.HasChange(ReturnToMasterField) || d.HasChange(ServersField) || d.HasChange(ServicesField) || d.HasChange(UseSrcField) || d.HasChange(VIPField) {
		haCluster[ReconfigureField] = true
	}

	_, err := client.doRequest(ctx, "PUT", fmt.Sprintf("/api/ha/cluster/%s", id), haCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaClusterRead(ctx, d, m)
}

func resourceHaClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	_, err := client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/ha/cluster/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

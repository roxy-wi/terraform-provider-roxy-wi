package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
	"time"
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
		CreateWithoutTimeout: resourceHaClusterCreate,
		ReadWithoutTimeout:   resourceHaClusterRead,
		UpdateWithoutTimeout: resourceHaClusterUpdate,
		DeleteWithoutTimeout: resourceHaClusterDelete,

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
	fmt.Printf("Servers: %+v\n", servers)

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

	jsonData, _ := json.Marshal(haCluster)
	fmt.Printf("HA Cluster Data: %s\n", string(jsonData))

	resp, err := client.doRequest("POST", "/api/ha/cluster", haCluster)
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

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/ha/cluster/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	servicesMap := result[ServicesField].(map[string]interface{})
	var servicesList []map[string]interface{}

	for serviceName, serviceDetails := range servicesMap {
		serviceData := serviceDetails.(map[string]interface{})
		servicesList = append(servicesList, map[string]interface{}{
			NameField:    serviceName,
			DockerField:  intToBool(serviceData[DockerField].(float64)),
			EnabledField: intToBool(serviceData[EnabledField].(float64)),
		})
	}

	servers, err := parseConfig(result[ServersField])
	if err != nil {
		return diag.FromErr(err)
	}
	serversResult := parseServersResult(servers)

	description := strings.ReplaceAll(result[DescriptionField].(string), "'", "")
	name := strings.ReplaceAll(result[NameField].(string), "'", "")

	d.Set(DescriptionField, description)
	d.Set(NameField, name)
	d.Set(ReturnToMasterField, intToBool(result[ReturnToMasterField].(float64)))
	d.Set(ServersField, serversResult)
	d.Set(ServicesField, result[ServicesField])
	d.Set(SynFloodField, intToBool(result[SynFloodField].(float64)))
	d.Set(UseSrcField, intToBool(result[UseSrcField].(float64)))
	d.Set(VIPField, result[VIPField])
	d.Set(VirtServerField, intToBool(result[VirtServerField].(float64)))

	return nil
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
	fmt.Printf("Servers: %+v\n", servers)

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

	jsonData, _ := json.Marshal(haCluster)
	fmt.Printf("HA Cluster Data: %s\n", string(jsonData))

	if d.HasChange(ReturnToMasterField) || d.HasChange(ServersField) || d.HasChange(ServicesField) || d.HasChange(UseSrcField) || d.HasChange(VIPField) {
		haCluster[ReconfigureField] = true
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/ha/cluster/%s", id), haCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHaClusterRead(ctx, d, m)
}

func resourceHaClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/ha/cluster/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

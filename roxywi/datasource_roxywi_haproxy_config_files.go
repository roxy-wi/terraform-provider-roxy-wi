package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	ConfigFilesField   = "files"
	MultipleFilesField = "multiple_files"
)

func dataSourceHaproxyConfigFiles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceHaproxyConfigFilesRead,
		Description: "Lists HAProxy configuration files detected by Roxy-WI 9.2 or later.",
		Schema: map[string]*schema.Schema{
			ServerIdField: {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Server ID whose HAProxy configuration files should be listed.",
			},
			ConfigFilesField: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Ordered HAProxy configuration files returned by Roxy-WI.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			MultipleFilesField: {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether multiple HAProxy configuration files are enabled for the server.",
			},
		},
	}
}

func dataSourceHaproxyConfigFilesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Config).Client
	serverID := d.Get(ServerIdField).(int)
	response, err := client.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/service/haproxy/%d/config/list", serverID), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result struct {
		Files         []string `json:"data"`
		MultipleFiles *bool    `json:"multiple_files"`
	}
	if err := json.Unmarshal(response, &result); err != nil {
		return diag.Errorf("decode HAProxy configuration file list: %v", err)
	}
	if result.MultipleFiles == nil {
		return diag.Errorf("HAProxy configuration file response is missing %q", MultipleFilesField)
	}
	if err := d.Set(ConfigFilesField, result.Files); err != nil {
		return diag.Errorf("set HAProxy configuration files: %v", err)
	}
	if err := d.Set(MultipleFilesField, *result.MultipleFiles); err != nil {
		return diag.Errorf("set HAProxy multiple-files state: %v", err)
	}
	d.SetId(fmt.Sprintf("%d", serverID))
	return nil
}

package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	BranchField    = "branch"
	RepoField      = "repo"
	ServiceIdField = "service_id"
)

func resourceBackupGit() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBackupGitCreate,
		ReadContext:   resourceBackupGitRead,
		UpdateContext: resourceBackupGitUpdate,
		DeleteContext: resourceBackupGitDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage Git backups",

		Schema: map[string]*schema.Schema{
			CredIDField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Credential ID.",
			},
			DescriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the backup.",
			},
			BranchField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Git branch to push.",
			},
			RepoField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Git repository to push.",
			},
			ServerField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server ID for the backup.",
			},
			TimeS3Field: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Time for the backup. Could be: dayli, weekly, monthly.",
			},
			ServiceIdField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Service ID: 1: HAProxy, 2: NGINX, 3: Keepalived, 4: Apache.",
			},
		},
	}
}

func resourceBackupGitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	backup := map[string]interface{}{
		CredIDField:      d.Get(CredIDField).(int),
		DescriptionField: description,
		BranchField:      d.Get(BranchField).(string),
		TimeS3Field:      d.Get(TimeS3Field).(string),
		ServiceIdField:   d.Get(ServiceIdField).(int),
		ServerField:      d.Get(ServerField).(int),
		RepoField:        d.Get(RepoField).(string),
	}

	resp, err := client.doRequest(ctx, "POST", "/api/server/backup/git", backup)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	id, ok := result["id"]
	if !ok {
		return diag.Errorf("unable to find ID in response: %v", result)
	}

	// Проверка типа id и приведение к строке
	var idStr string
	switch v := id.(type) {
	case string:
		idStr = v
	case int:
		idStr = strconv.Itoa(v)
	case float64:
		idStr = strconv.FormatFloat(v, 'f', 0, 64)
	default:
		return diag.Errorf("unsupported type for ID: %T", v)
	}

	d.SetId(idStr)
	return resourceBackupGitRead(ctx, d, m)
}

func resourceBackupGitRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	resp, err := client.doRequest(ctx, "GET", fmt.Sprintf("/api/server/backup/git/%s", id), nil)
	if err != nil {
		return readDiagnostics(d, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	descriptionValue, err := apiString(result, DescriptionField)
	if err != nil {
		return diag.FromErr(err)
	}
	description := strings.ReplaceAll(descriptionValue, "'", "")

	if err := d.Set(CredIDField, result[CredIDField]); err != nil {

		return diag.Errorf("set Terraform state: %v", err)

	}
	if err := d.Set(DescriptionField, description); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(BranchField, result[BranchField]); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(TimeS3Field, result[TimeS3Field]); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(ServerField, result[ServerField]); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(ServiceIdField, result[ServiceIdField]); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}
	if err := d.Set(RepoField, result[RepoField]); err != nil {
		return diag.Errorf("set Terraform state: %v", err)
	}

	return nil
}

func resourceBackupGitUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	backup := map[string]interface{}{
		CredIDField:      d.Get(CredIDField).(int),
		DescriptionField: description,
		BranchField:      d.Get(BranchField).(string),
		TimeS3Field:      d.Get(TimeS3Field).(string),
		ServiceIdField:   d.Get(ServiceIdField).(int),
		ServerField:      d.Get(ServerField).(int),
		RepoField:        d.Get(RepoField).(string),
	}

	_, err := client.doRequest(ctx, "PUT", fmt.Sprintf("/api/server/backup/git/%s", id), backup)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBackupGitRead(ctx, d, m)
}

func resourceBackupGitDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	// Подготовка данных для удаления
	deleteData := map[string]interface{}{
		ServerField: d.Get(ServerField).(int),
		CredIDField: d.Get(CredIDField).(int),
	}

	_, err := client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/server/backup/git/%s", id), deleteData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

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
	RPathField   = "rpath"
	RServerField = "rserver"
	ServerField  = "server_id"
	TimeField    = "time"
	TypeField    = "type"
)

func resourceBackupFs() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackupFsCreate,
		ReadWithoutTimeout:   resourceBackupFsRead,
		UpdateWithoutTimeout: resourceBackupFsUpdate,
		DeleteWithoutTimeout: resourceBackupFsDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage backups to remote File system",

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
			RPathField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote path for the backup.",
			},
			RServerField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Remote server for the backup.",
			},
			ServerField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server ID for the backup.",
			},
			TimeField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Time for the backup. Could be: dayli, weekly, monthly",
			},
			TypeField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Type of the backup. Could be: backup, synchronization",
			},
		},
	}
}

func resourceBackupFsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	backup := map[string]interface{}{
		CredIDField:      d.Get(CredIDField).(int),
		DescriptionField: description,
		RPathField:       d.Get(RPathField).(string),
		RServerField:     d.Get(RServerField).(string),
		ServerField:      d.Get(ServerField).(int),
		TimeField:        d.Get(TimeField).(string),
		TypeField:        d.Get(TypeField).(string),
	}

	resp, err := client.doRequest("POST", "/api/server/backup/fs", backup)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	// Добавляем больше логирования для отладки
	fmt.Printf("Response result: %v\n", result)

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
	return resourceBackupFsRead(ctx, d, m)
}

func resourceBackupFsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/server/backup/fs/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	description := strings.ReplaceAll(result[DescriptionField].(string), "'", "")

	d.Set(CredIDField, result[CredIDField])
	d.Set(DescriptionField, description)
	d.Set(RPathField, result[RPathField])
	d.Set(RServerField, result[RServerField])
	d.Set(ServerField, result[ServerField])
	d.Set(TimeField, result[TimeField])
	d.Set(TypeField, result[TypeField])

	return nil
}

func resourceBackupFsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	backup := map[string]interface{}{
		CredIDField:      d.Get(CredIDField).(int),
		DescriptionField: description,
		RPathField:       d.Get(RPathField).(string),
		RServerField:     d.Get(RServerField).(string),
		ServerField:      d.Get(ServerField).(int),
		TimeField:        d.Get(TimeField).(string),
		TypeField:        d.Get(TypeField).(string),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/server/backup/fs/%s", id), backup)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBackupFsRead(ctx, d, m)
}

func resourceBackupFsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	// Подготовка данных для удаления
	deleteData := map[string]interface{}{
		ServerField: d.Get(ServerField).(int),
		CredIDField: d.Get(CredIDField).(int),
	}

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/server/backup/fs/%s", id), deleteData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

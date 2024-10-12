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
	S3Server    = "s3_server"
	AccessKey   = "access_key"
	SecretKey   = "secret_key"
	Bucket      = "bucket"
	TimeS3Field = "time"
)

func resourceBackupS3() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackupS3Create,
		ReadWithoutTimeout:   resourceBackupS3Read,
		UpdateWithoutTimeout: resourceBackupS3Update,
		DeleteWithoutTimeout: resourceBackupS3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Description: "Manage creating backup to S3.",

		Schema: map[string]*schema.Schema{
			S3Server: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "S3 server endpoint.",
			},
			DescriptionField: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the backup.",
			},
			AccessKey: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "S3 Access key.",
			},
			SecretKey: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "S3 Secret key.",
			},
			Bucket: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "S3 bucket.",
			},
			TimeField: {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Time for the backup. Could be: dayli, weekly, monthly",
			},
			ServerField: {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Server ID for backup",
			},
		},
	}
}

func resourceBackupS3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client

	// Получение значений из схемы
	s3Server := d.Get(S3Server).(string)
	accessKey := d.Get(AccessKey).(string)
	secretKey := d.Get(SecretKey).(string)
	bucket := d.Get(Bucket).(string)
	serverID := d.Get(ServerField).(int)
	backupTime := d.Get(TimeS3Field).(string)
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	// Создаём резервные копии S3 на основе данных схемы
	backup := map[string]interface{}{
		S3Server:         s3Server,
		AccessKey:        accessKey,
		SecretKey:        secretKey,
		Bucket:           bucket,
		ServerField:      serverID,
		TimeS3Field:      backupTime,
		DescriptionField: description,
	}

	resp, err := client.doRequest("POST", "/api/server/backup/s3", backup)
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
	return resourceBackupS3Read(ctx, d, m)
}

func resourceBackupS3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	resp, err := client.doRequest("GET", fmt.Sprintf("/api/server/backup/s3/%s", id), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return diag.FromErr(err)
	}

	description := strings.ReplaceAll(result[DescriptionField].(string), "'", "")

	d.Set(S3Server, result[S3Server])
	d.Set(DescriptionField, description)
	d.Set(AccessKey, result[AccessKey])
	d.Set(SecretKey, result[SecretKey])
	d.Set(Bucket, result[Bucket])
	d.Set(ServerField, result[ServerField])
	d.Set(TimeField, result[TimeField])
	d.Set(DescriptionField, result[DescriptionField])

	return nil
}

func resourceBackupS3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()
	description := strings.ReplaceAll(d.Get(DescriptionField).(string), "'", "")

	backup := map[string]interface{}{
		S3Server:         d.Get(S3Server).(string),
		DescriptionField: description,
		AccessKey:        d.Get(AccessKey).(string),
		SecretKey:        d.Get(SecretKey).(string),
		Bucket:           d.Get(Bucket).(string),
		ServerField:      d.Get(ServerField).(int),
		TimeField:        d.Get(TimeField).(string),
	}

	_, err := client.doRequest("PUT", fmt.Sprintf("/api/server/backup/s3/%s", id), backup)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceBackupS3Read(ctx, d, m)
}

func resourceBackupS3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	// Подготовка данных для удаления
	deleteData := map[string]interface{}{
		Bucket:      d.Get(Bucket).(int),
		ServerField: d.Get(ServerField).(int),
	}

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/server/backup/s3/%s", id), deleteData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

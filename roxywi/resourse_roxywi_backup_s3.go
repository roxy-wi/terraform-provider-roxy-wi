package roxywi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	S3Server           = "s3_server"
	AccessKey          = "access_key"
	SecretKey          = "secret_key"
	Bucket             = "bucket"
	ServerIdS3Field    = "server_id"
	TimeS3Field        = "time"
	DescriptionS3Field = "description"
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

		Description: "",

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
			ServerIdS3Field: {
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
	serverID := d.Get(ServerIdS3Field).(int)
	backupTime := d.Get(TimeS3Field).(string)
	description := d.Get(DescriptionS3Field).(string)

	// Создаём резервные копии S3 на основе данных схемы
	backup := map[string]interface{}{
		S3Server:           s3Server,
		AccessKey:          accessKey,
		SecretKey:          secretKey,
		Bucket:             bucket,
		ServerIdS3Field:    serverID,
		TimeS3Field:        backupTime,
		DescriptionS3Field: description,
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

	d.Set(S3Server, result[S3Server])
	d.Set(DescriptionS3Field, result[DescriptionS3Field])
	d.Set(AccessKey, result[AccessKey])
	d.Set(SecretKey, result[SecretKey])
	d.Set(Bucket, result[Bucket])
	d.Set(ServerIdS3Field, result[ServerIdS3Field])
	d.Set(TimeField, result[TimeField])
	d.Set(DescriptionS3Field, result[DescriptionS3Field])

	return nil
}

func resourceBackupS3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Config).Client
	id := d.Id()

	backup := map[string]interface{}{
		S3Server:           d.Get(S3Server).(string),
		DescriptionS3Field: d.Get(DescriptionS3Field).(string),
		AccessKey:          d.Get(AccessKey).(string),
		SecretKey:          d.Get(SecretKey).(string),
		Bucket:             d.Get(Bucket).(string),
		ServerIdS3Field:    d.Get(ServerIdS3Field).(int),
		TimeField:          d.Get(TimeField).(string),
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
		Bucket:          d.Get(Bucket).(int),
		ServerIdS3Field: d.Get(ServerIdS3Field).(int),
	}

	_, err := client.doRequest("DELETE", fmt.Sprintf("/api/server/backup/s3/%s", id), deleteData)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

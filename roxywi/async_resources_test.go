package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestBackupScheduleReadAndClear(t *testing.T) {
	for _, kind := range []string{"fs", "s3"} {
		t.Run(kind, func(t *testing.T) {
			resource, read := resourceBackupFs(), resourceBackupFsRead
			if kind == "s3" {
				resource, read = resourceBackupS3(), resourceBackupS3Read
			}
			d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
			d.SetId("12")
			for _, schedule := range []interface{}{
				map[string]interface{}{"timezone": "Europe/Moscow", "next_run_at": "2026-10-05T00:00:00Z", "retry_at": nil, "last_task_id": nil, "last_status": nil, "migration_required": true},
				nil,
			} {
				config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
					writeTestJSON(t, w, map[string]interface{}{"description": "backup", "cred_id": 1, "server_id": 10, "rpath": "/backup", "rserver": "backup.example.com", "type": "backup", "time": "hourly", "s3_server": "https://s3.example.com", "bucket": "backups", "schedule": schedule})
				})
				if result := read(context.Background(), d, config); result.HasError() {
					t.Fatal(result)
				}
				got := d.Get("schedule").([]interface{})
				if schedule == nil {
					if len(got) != 0 {
						t.Fatal("stale schedule not cleared")
					}
				} else {
					if len(got) != 1 {
						t.Fatal("missing schedule")
					}
					row := got[0].(map[string]interface{})
					if row["migration_required"] != true || row["last_task_id"] != 0 || row["timezone"] != "Europe/Moscow" {
						t.Fatalf("unexpected schedule: %v", row)
					}
				}
			}
		})
	}
}

func TestAsyncInstallationAndHAKeepIDOnFailure(t *testing.T) {
	for _, kind := range []string{"service", "ha"} {
		for _, update := range []bool{false, true} {
			t.Run(kind+map[bool]string{false: " create", true: " update"}[update], func(t *testing.T) {
				resource, create, change, id := resourceServiceInstallation(), resourceServiceInstallationCreate, resourceServiceInstallationUpdate, "10-haproxy"
				raw := map[string]interface{}{"service": "haproxy", "server_id": 10}
				if kind == "ha" {
					resource, create, change, id = resourceHaCluster(), resourceHaClusterCreate, resourceHaClusterUpdate, "12"
					raw = map[string]interface{}{"name": "edge", "description": "cluster", "vip": "192.0.2.1", "servers": []interface{}{map[string]interface{}{"id": 10, "master": true, "eth": "eth0"}}, "services": []interface{}{map[string]interface{}{"name": "haproxy", "enabled": true, "docker": true}}}
				}
				config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/api/operations/7" {
						writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": "failed"})
						return
					}
					if r.Method == "GET" {
						t.Error("resource was read before operation completed")
					}
					var responseID interface{} = id
					if kind == "ha" {
						responseID = 12
					}
					w.WriteHeader(202)
					writeTestJSON(t, w, map[string]interface{}{"id": responseID, "tasks_ids": []int{7}})
				})
				d := schema.TestResourceDataRaw(t, resource.Schema, raw)
				fn := create
				if update {
					d.SetId(id)
					fn = change
				}
				if result := fn(context.Background(), d, config); !result.HasError() {
					t.Fatal("failed operation reported success")
				}
				if d.Id() != id {
					t.Fatalf("lost resource ID: %s", d.Id())
				}
			})
		}
	}
}

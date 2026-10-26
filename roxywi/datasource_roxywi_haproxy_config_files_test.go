package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceHaproxyConfigFilesRead(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/service/haproxy/7/config/list" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		writeTestJSON(t, w, map[string]interface{}{
			"data":           []string{"/etc/haproxy/haproxy.cfg", "/etc/haproxy/conf.d/apps.cfg"},
			"multiple_files": true,
		})
	})

	dataSource := dataSourceHaproxyConfigFiles()
	d := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{ServerIdField: 7})
	if diags := dataSourceHaproxyConfigFilesRead(context.Background(), d, config); diags.HasError() {
		t.Fatalf("read HAProxy config files: %v", diags)
	}
	if d.Id() != "7" {
		t.Fatalf("unexpected data source ID %q", d.Id())
	}
	if !d.Get(MultipleFilesField).(bool) {
		t.Fatal("expected multiple_files to be true")
	}
	files := d.Get(ConfigFilesField).([]interface{})
	if len(files) != 2 || files[1] != "/etc/haproxy/conf.d/apps.cfg" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestDataSourceHaproxyConfigFilesRejectsMissingMode(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]interface{}{"data": []string{}})
	})

	dataSource := dataSourceHaproxyConfigFiles()
	d := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]interface{}{ServerIdField: 7})
	if diags := dataSourceHaproxyConfigFilesRead(context.Background(), d, config); !diags.HasError() {
		t.Fatal("expected missing multiple_files to fail")
	}
}

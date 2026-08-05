package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceGroupReadByID(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/group/7" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		writeTestJSON(t, w, map[string]interface{}{
			NameField:        "edge",
			DescriptionField: "Edge servers",
		})
	})
	d := schema.TestResourceDataRaw(t, dataSourceGroup().Schema, map[string]interface{}{IDField: "7"})

	diagnostics := dataSourceGroupRead(context.Background(), d, config)
	if diagnostics.HasError() {
		t.Fatalf("dataSourceGroupRead returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "7" || d.Get(NameField) != "edge" || d.Get(DescriptionField) != "Edge servers" {
		t.Fatalf("unexpected group state: id=%q name=%#v description=%#v", d.Id(), d.Get(NameField), d.Get(DescriptionField))
	}
}

func TestDataSourceGroupReadByName(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/groups" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		writeTestJSON(t, w, []map[string]interface{}{
			{"group_id": 5, NameField: "other", DescriptionField: "Other"},
			{"group_id": 8, NameField: "database", DescriptionField: "Database servers"},
		})
	})
	d := schema.TestResourceDataRaw(t, dataSourceGroup().Schema, map[string]interface{}{NameField: "database"})

	diagnostics := dataSourceGroupRead(context.Background(), d, config)
	if diagnostics.HasError() {
		t.Fatalf("dataSourceGroupRead returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "8" || d.Get(DescriptionField) != "Database servers" {
		t.Fatalf("unexpected group state: id=%q description=%#v", d.Id(), d.Get(DescriptionField))
	}
}

func TestDataSourceGroupReadReportsMissingSelectionAndName(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("API should not be called without a selector")
		http.Error(w, "unexpected request", http.StatusBadRequest)
	})
	d := schema.TestResourceDataRaw(t, dataSourceGroup().Schema, nil)
	if diagnostics := dataSourceGroupRead(context.Background(), d, config); !diagnostics.HasError() {
		t.Fatal("expected a missing-selector diagnostic")
	}

	config = newTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(t, w, []map[string]interface{}{{"group_id": 1, NameField: "default"}})
	})
	d = schema.TestResourceDataRaw(t, dataSourceGroup().Schema, map[string]interface{}{NameField: "missing"})
	if diagnostics := dataSourceGroupRead(context.Background(), d, config); !diagnostics.HasError() {
		t.Fatal("expected a not-found diagnostic")
	}
}

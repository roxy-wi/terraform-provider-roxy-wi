package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestHaproxySectionIDRoundTrip(t *testing.T) {
	const filePath = "/etc/haproxy/conf.d/apps & edge.cfg"
	if id := haproxySectionID("7", "api-backend", ""); id != "7-api-backend" {
		t.Fatalf("default configuration ID changed: %q", id)
	}
	id := haproxySectionID("7", "api-backend", filePath)
	serverID, sectionName, parsedPath, err := resourceHaproxySectionParseID(id)
	if err != nil {
		t.Fatal(err)
	}
	if serverID != "7" || sectionName != "api-backend" || parsedPath != filePath {
		t.Fatalf("unexpected parsed ID: %q %q %q", serverID, sectionName, parsedPath)
	}

	serverID, sectionName, parsedPath, err = resourceHaproxySectionParseID("7-api-backend")
	if err != nil {
		t.Fatal(err)
	}
	if serverID != "7" || sectionName != "api-backend" || parsedPath != "" {
		t.Fatalf("unexpected legacy ID: %q %q %q", serverID, sectionName, parsedPath)
	}
}

func TestHaproxySectionImport(t *testing.T) {
	resource := resourceHaproxySectionBackend()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
	d.SetId("7|/etc/haproxy/conf.d/apps.cfg|api-backend")
	states, err := resource.Importer.StateContext(context.Background(), d, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 {
		t.Fatalf("unexpected imported states: %d", len(states))
	}
	if got := states[0].Get(ServerIdField).(int); got != 7 {
		t.Fatalf("unexpected server_id %d", got)
	}
	if got := states[0].Get(NameField).(string); got != "api-backend" {
		t.Fatalf("unexpected name %q", got)
	}
	if got := states[0].Get(FilePathField).(string); got != "/etc/haproxy/conf.d/apps.cfg" {
		t.Fatalf("unexpected file_path %q", got)
	}
	if states[0].Id() == "7-api-backend" {
		t.Fatal("multi-file import must use an unambiguous canonical ID")
	}
}

func TestHaproxySectionResourcesExposeFilePath(t *testing.T) {
	resources := map[string]struct {
		resource *schema.Resource
		forceNew bool
	}{
		"backend":  {resourceHaproxySectionBackend(), true},
		"defaults": {resourceHaproxySectionDefaults(), false},
		"frontend": {resourceHaproxySectionFrontend(), true},
		"global":   {resourceHaproxySectionGlobal(), false},
		"listen":   {resourceHaproxySectionListen(), true},
		"peers":    {resourceHaproxySectionPeers(), true},
		"userlist": {resourceHaproxySectionUserlist(), true},
	}
	for name, test := range resources {
		field, ok := test.resource.Schema[FilePathField]
		if !ok {
			t.Errorf("%s resource does not expose file_path", name)
			continue
		}
		if !field.Optional || field.ForceNew != test.forceNew {
			t.Errorf("%s file_path: Optional=%t ForceNew=%t", name, field.Optional, field.ForceNew)
		}
	}
}

func TestSingletonHaproxySectionImportRejectsWrongSection(t *testing.T) {
	resource := resourceHaproxySectionGlobal()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{})
	d.SetId("7-defaults")
	if _, err := resource.Importer.StateContext(context.Background(), d, nil); err == nil {
		t.Fatal("global import accepted defaults section ID")
	}
}

func TestSingletonHaproxySectionIDTracksFilePath(t *testing.T) {
	resource := resourceHaproxySectionGlobal()
	d := schema.TestResourceDataRaw(t, resource.Schema, map[string]interface{}{
		ServerIdField: 7,
		FilePathField: "/etc/haproxy/conf.d/apps.cfg",
	})
	d.SetId("7-global")
	if err := setSingletonHaproxySectionID(d, "global"); err != nil {
		t.Fatal(err)
	}
	serverID, sectionName, filePath, err := resourceHaproxySectionParseID(d.Id())
	if err != nil {
		t.Fatal(err)
	}
	if serverID != "7" || sectionName != "global" || filePath != "/etc/haproxy/conf.d/apps.cfg" {
		t.Fatalf("unexpected singleton ID: %q %q %q", serverID, sectionName, filePath)
	}
}

func TestNamedHaproxySectionDeletesUseFilePath(t *testing.T) {
	type deleteFunc func(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics
	tests := map[string]struct {
		resource *schema.Resource
		delete   deleteFunc
		path     string
	}{
		"backend":  {resourceHaproxySectionBackend(), resourceHaproxySectionBackendDelete, "/api/service/haproxy/7/section/backend/example"},
		"frontend": {resourceHaproxySectionFrontend(), resourceHaproxySectionFrontendDelete, "/api/service/haproxy/7/section/frontend/example"},
		"listen":   {resourceHaproxySectionListen(), resourceHaproxySectionListenDelete, "/api/service/haproxy/7/section/listen/example"},
		"peers":    {resourceHaproxySectionPeers(), resourceHaproxySectionPeersDelete, "/api/service/haproxy/7/section/peers/example"},
		"userlist": {resourceHaproxySectionUserlist(), resourceHaproxySectionUserlistDelete, "/api/service/haproxy/7/section/userlist/example"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete || r.URL.Path != test.path {
					http.Error(w, "unexpected request", http.StatusBadRequest)
					return
				}
				if got := r.URL.Query().Get(FilePathField); got != "/etc/haproxy/conf.d/apps.cfg" {
					t.Errorf("unexpected file_path %q", got)
				}
				w.WriteHeader(http.StatusNoContent)
			})

			d := schema.TestResourceDataRaw(t, test.resource.Schema, map[string]interface{}{
				ServerIdField: 7,
				NameField:     "example",
				FilePathField: "/etc/haproxy/conf.d/apps.cfg",
			})
			d.SetId(haproxySectionID("7", "example", "/etc/haproxy/conf.d/apps.cfg"))
			if diags := test.delete(context.Background(), d, config); diags.HasError() {
				t.Fatalf("delete section: %v", diags)
			}
		})
	}
}

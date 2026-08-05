package roxywi

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceUserRoleRead(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/user/roles" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		writeTestJSON(t, w, []map[string]interface{}{
			{RoleIDField: 1, RoleNameField: "Admin", RoleDescriptionField: "All permissions"},
			{RoleIDField: 2, RoleNameField: "Viewer", RoleDescriptionField: "Read only"},
		})
	})
	d := schema.TestResourceDataRaw(t, dataSourceUserRole().Schema, nil)

	diagnostics := dataSourceUserRoleRead(context.Background(), d, config)
	if diagnostics.HasError() {
		t.Fatalf("dataSourceUserRoleRead returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "roles" {
		t.Fatalf("unexpected ID %q", d.Id())
	}
	roles := d.Get(RolesField).([]interface{})
	if len(roles) != 2 || roles[0].(map[string]interface{})[RoleIDField] != "1" {
		t.Fatalf("unexpected roles state: %#v", roles)
	}
}

func TestDataSourceUserRoleReadRejectsEmptyResult(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, _ *http.Request) {
		writeTestJSON(t, w, []map[string]interface{}{})
	})
	d := schema.TestResourceDataRaw(t, dataSourceUserRole().Schema, nil)
	if diagnostics := dataSourceUserRoleRead(context.Background(), d, config); !diagnostics.HasError() {
		t.Fatal("expected an empty roles diagnostic")
	}
}

func TestConvertRolesRejectsMissingField(t *testing.T) {
	role := cty.ObjectVal(map[string]cty.Value{
		RoleIDField:   cty.NumberIntVal(1),
		RoleNameField: cty.StringVal("Admin"),
	})
	if _, err := convertRoles([]cty.Value{role}); err == nil {
		t.Fatal("expected a missing description error")
	}
}

package roxywi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceGroupLifecycle(t *testing.T) {
	var mu sync.Mutex
	group := map[string]interface{}{NameField: "edge", DescriptionField: "Edge servers"}
	var methods []string

	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods = append(methods, r.Method+" "+r.URL.Path)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/group":
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode group create: %v", err)
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload[NameField] != "edge" || payload[DescriptionField] != "Edge servers" {
				t.Errorf("unexpected create payload: %#v", payload)
				http.Error(w, "unexpected payload", http.StatusBadRequest)
				return
			}
			writeTestJSON(t, w, map[string]interface{}{IDField: 12})
		case r.Method == http.MethodGet && r.URL.Path == "/api/group/12":
			writeTestJSON(t, w, group)
		case r.Method == http.MethodPut && r.URL.Path == "/api/group/12":
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode group update: %v", err)
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			group[NameField] = payload[NameField]
			if description, ok := payload[DescriptionField]; ok {
				group[DescriptionField] = description
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/group/12":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceGroup().Schema, map[string]interface{}{
		NameField:        "edge",
		DescriptionField: "Edge servers",
	})
	if diagnostics := resourceGroupCreate(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("create returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "12" {
		t.Fatalf("unexpected group ID %q", d.Id())
	}

	if err := d.Set(NameField, "edge-updated"); err != nil {
		t.Fatal(err)
	}
	if diagnostics := resourceGroupUpdate(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("update returned diagnostics: %v", diagnostics)
	}
	if d.Get(NameField) != "edge-updated" {
		t.Fatalf("unexpected updated name %#v", d.Get(NameField))
	}

	if diagnostics := resourceGroupDelete(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("delete returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "" {
		t.Fatalf("group ID was not cleared: %q", d.Id())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(methods) != 5 {
		t.Fatalf("unexpected request sequence: %#v", methods)
	}
}

func TestResourceGroupReadClearsMissingState(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/group/404" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		http.NotFound(w, r)
	})
	d := schema.TestResourceDataRaw(t, resourceGroup().Schema, nil)
	d.SetId("404")

	if diagnostics := resourceGroupRead(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("read returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "" {
		t.Fatalf("missing group remained in state: %q", d.Id())
	}
}

func TestResourceUserLifecycleDoesNotRefreshPassword(t *testing.T) {
	var mu sync.Mutex
	user := map[string]interface{}{
		UserEmailField:    "user@example.com",
		UserEnabledField:  1,
		UserUsernameField: "user",
		UserPasswordField: "echoed-api-password",
	}
	var methods []string

	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods = append(methods, r.Method+" "+r.URL.Path)

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/user":
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode user create: %v", err)
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload[UserPasswordField] != "configured-password" || payload[UserEnabledField] != float64(1) {
				t.Errorf("unexpected create payload: %#v", payload)
				http.Error(w, "unexpected payload", http.StatusBadRequest)
				return
			}
			writeTestJSON(t, w, map[string]interface{}{"id": "21"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/user/21":
			writeTestJSON(t, w, user)
		case r.Method == http.MethodPut && r.URL.Path == "/api/user/21":
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("decode user update: %v", err)
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			user[UserEmailField] = payload[UserEmailField]
			user[UserEnabledField] = payload[UserEnabledField]
			user[UserUsernameField] = payload[UserUsernameField]
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/api/user/21":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
		}
	})

	d := schema.TestResourceDataRaw(t, resourceUser().Schema, map[string]interface{}{
		UserEmailField:    "user@example.com",
		UserEnabledField:  true,
		UserPasswordField: "configured-password",
		UserUsernameField: "user",
	})
	if diagnostics := resourceUserCreate(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("create returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "21" || d.Get(UserPasswordField) != "configured-password" {
		t.Fatalf("unexpected user state: id=%q password=%#v", d.Id(), d.Get(UserPasswordField))
	}

	if err := d.Set(UserEmailField, "updated@example.com"); err != nil {
		t.Fatal(err)
	}
	if diagnostics := resourceUserUpdate(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("update returned diagnostics: %v", diagnostics)
	}
	if d.Get(UserEmailField) != "updated@example.com" || d.Get(UserPasswordField) != "configured-password" {
		t.Fatalf("unexpected updated user state: %#v", d.State())
	}

	if diagnostics := resourceUserDelete(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("delete returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "" {
		t.Fatalf("user ID was not cleared: %q", d.Id())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(methods) != 5 {
		t.Fatalf("unexpected request sequence: %#v", methods)
	}
}

func TestResourceUserReadClearsMissingState(t *testing.T) {
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user/404" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		http.NotFound(w, r)
	})
	d := schema.TestResourceDataRaw(t, resourceUser().Schema, nil)
	d.SetId("404")

	if diagnostics := resourceUserRead(context.Background(), d, config); diagnostics.HasError() {
		t.Fatalf("read returned diagnostics: %v", diagnostics)
	}
	if d.Id() != "" {
		t.Fatalf("missing user remained in state: %q", d.Id())
	}
}

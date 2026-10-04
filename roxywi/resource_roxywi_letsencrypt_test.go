package roxywi

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func leTestData(t *testing.T) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, resourceLetsencrypt().Schema, map[string]interface{}{
		DomainsField: []interface{}{"example.com"}, ServerIdField: 10, TypeField: "cloudflare", ApiTokenField: "configured-token",
	})
}

func TestLetsencryptLifecyclePreservesCredentials(t *testing.T) {
	var mu sync.Mutex
	methods := []string{}
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		methods = append(methods, r.Method+" "+r.URL.Path)
		if r.URL.Path == "/api/operations/7" {
			writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": "completed"})
			return
		}
		if r.Method == "GET" {
			writeTestJSON(t, w, map[string]interface{}{"id": 12, "domains": []string{"example.com"}, "server_id": 10, "type": "cloudflare", "api_token": nil, "api_key": nil, "dns_profile_id": nil, "draft": false, "state": map[string]interface{}{"status": "active", "last_task_id": 7, "not_after": "2026-12-01T00:00:00Z"}})
			return
		}
		if r.Method == "POST" || r.Method == "PUT" {
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
			}
			if payload[ApiTokenField] != "configured-token" {
				t.Error("configured token was lost")
			}
		} else if r.Method != "DELETE" {
			t.Errorf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusAccepted)
		writeTestJSON(t, w, map[string]interface{}{"id": 12, "status": "accepted", "tasks_ids": []int{7}})
	})
	d := leTestData(t)
	for _, action := range []schema.CreateContextFunc{resourceLetsencryptCreate, resourceLetsencryptUpdate} {
		if result := action(context.Background(), d, config); result.HasError() {
			t.Fatal(result)
		}
		if d.Get(ApiTokenField) != "configured-token" || d.Get("status") != "active" || d.Get("last_task_id") != 7 {
			t.Fatal("incorrect certificate state")
		}
	}
	if result := resourceLetsencryptDelete(context.Background(), d, config); result.HasError() {
		t.Fatal(result)
	}
	if d.Id() != "" {
		t.Fatal("resource not removed")
	}
	mu.Lock()
	defer mu.Unlock()
	want := []string{"POST /api/service/letsencrypt", "GET /api/operations/7", "GET /api/service/letsencrypt/12", "PUT /api/service/letsencrypt/12", "GET /api/operations/7", "GET /api/service/letsencrypt/12", "DELETE /api/service/letsencrypt/12", "GET /api/operations/7"}
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("unexpected requests: %v", methods)
	}
}

func TestLetsencryptFailuresRetainID(t *testing.T) {
	for _, action := range []string{"create", "update", "delete"} {
		t.Run(action, func(t *testing.T) {
			config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": "failed"})
					return
				}
				w.WriteHeader(http.StatusAccepted)
				writeTestJSON(t, w, map[string]interface{}{"id": 12, "tasks_ids": []int{7}})
			})
			d := leTestData(t)
			fn := resourceLetsencryptCreate
			if action != "create" {
				d.SetId("12")
				fn = resourceLetsencryptUpdate
			}
			if action == "delete" {
				fn = resourceLetsencryptDelete
			}
			if result := fn(context.Background(), d, config); !result.HasError() {
				t.Fatal("failure was ignored")
			}
			if d.Id() != "12" {
				t.Fatal("lost resource ID after failure")
			}
		})
	}
}

func TestLetsencryptDraftPromotion(t *testing.T) {
	for _, failPreflight := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "preflight failure"}[failPreflight], func(t *testing.T) {
			r := resourceLetsencrypt()
			d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"domains": []interface{}{"example.com"}, "server_id": 10, "type": "cloudflare", "dns_profile_id": 4, "draft": true})
			d.SetId("12")
			var err error
			d, err = schema.InternalMap(r.Schema).Data(d.State(), &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{"draft": {Old: "true", New: "false"}}})
			if err != nil {
				t.Fatal(err)
			}
			var mu sync.Mutex
			actions := []string{}
			config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				defer mu.Unlock()
				switch r.Method {
				case "PUT":
					var payload map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if payload["dns_profile_id"] != float64(4) {
						t.Error("DNS profile not sent")
					}
					if _, found := payload["api_token"]; found {
						t.Error("credentials sent with a profile")
					}
					writeTestJSON(t, w, map[string]interface{}{"id": 12, "status": "draft", "tasks_ids": []int{}})
				case "PATCH":
					var payload map[string]string
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					actions = append(actions, payload["action"])
					w.WriteHeader(202)
					writeTestJSON(t, w, map[string]interface{}{"id": 12, "tasks_ids": []int{7}})
				case "GET":
					if r.URL.Path == "/api/operations/7" {
						state := "completed"
						if failPreflight {
							state = "failed"
						}
						writeTestJSON(t, w, map[string]interface{}{"task_id": 7, "status": state})
					} else {
						writeTestJSON(t, w, map[string]interface{}{"draft": false, "dns_profile_id": 4, "state": map[string]interface{}{"status": "active"}})
					}
				default:
					t.Errorf("unexpected request %s", r.Method)
				}
			})
			if result := resourceLetsencryptUpdate(context.Background(), d, config); result.HasError() != failPreflight {
				t.Fatal(result)
			}
			mu.Lock()
			defer mu.Unlock()
			want := []string{"preflight", "issue"}
			if failPreflight {
				want = []string{"preflight"}
			}
			if !reflect.DeepEqual(actions, want) {
				t.Fatalf("actions=%v", actions)
			}
			if d.Id() != "12" {
				t.Fatal("resource ID lost")
			}
		})
	}
}

func TestLetsencryptImportedSecretsAndMissingResource(t *testing.T) {
	d := resourceLetsencrypt().Data(&terraform.InstanceState{ID: "12", Attributes: map[string]string{"type": "cloudflare", "server_id": "10", "domains.#": "1", "domains.0": "example.com"}})
	payload := letsencryptPayload(d)
	if _, ok := payload[ApiTokenField]; ok {
		t.Fatal("imported secret must be omitted")
	}
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })
	if result := resourceLetsencryptDelete(context.Background(), d, config); result.HasError() {
		t.Fatal(result)
	}
	if d.Id() != "" {
		t.Fatal("404 delete must clear ID")
	}
	d.SetId("12")
	if result := resourceLetsencryptRead(context.Background(), d, config); result.HasError() {
		t.Fatal(result)
	}
	if d.Id() != "" {
		t.Fatal("404 read must clear ID")
	}
}

func TestLetsencryptRefreshHasNoCredentialDiff(t *testing.T) {
	resource := resourceLetsencrypt()
	raw := map[string]interface{}{"domains": []interface{}{"example.com"}, "server_id": 10, "type": "cloudflare", "api_token": "configured-token"}
	d := schema.TestResourceDataRaw(t, resource.Schema, raw)
	d.SetId("12")
	config := newTestConfig(t, func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]interface{}{"domains": []string{"example.com"}, "server_id": 10, "type": "cloudflare", "api_token": nil, "api_key": nil, "dns_profile_id": nil, "draft": false, "state": map[string]interface{}{"status": "active"}})
	})
	if result := resourceLetsencryptRead(context.Background(), d, config); result.HasError() {
		t.Fatal(result)
	}
	diff, err := resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(raw), config)
	if err != nil {
		t.Fatal(err)
	}
	if diff != nil && !diff.Empty() {
		t.Fatalf("refresh created a diff: %#v", diff.Attributes)
	}
}

func TestLetsencryptRejectsReturningToDraft(t *testing.T) {
	resource := resourceLetsencrypt()
	raw := map[string]interface{}{"domains": []interface{}{"example.com"}, "server_id": 10, "type": "cloudflare", "dns_profile_id": 4, "draft": false}
	d := schema.TestResourceDataRaw(t, resource.Schema, raw)
	d.SetId("12")
	raw["draft"] = true
	if _, err := resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(raw), nil); err == nil {
		t.Fatal("active certificate could become a draft")
	}
}

package roxywi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceBackupS3DeleteSendsStringBucket(t *testing.T) {
	payloads := make(chan map[string]interface{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
			return
		}
		if r.Method != http.MethodDelete || r.URL.Path != "/api/server/backup/s3/7" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode DELETE payload: %v", err)
		}
		payloads <- payload
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	d := schema.TestResourceDataRaw(t, resourceBackupS3().Schema, map[string]interface{}{
		Bucket:      "backups",
		ServerField: 9,
	})
	d.SetId("7")

	if diags := resourceBackupS3Delete(context.Background(), d, &Config{Client: client}); len(diags) != 0 {
		t.Fatalf("resourceBackupS3Delete returned diagnostics: %v", diags)
	}
	payload := <-payloads
	if got, ok := payload[Bucket].(string); !ok || got != "backups" {
		t.Fatalf("expected string bucket in DELETE payload, got %#v", payload[Bucket])
	}
	if d.Id() != "" {
		t.Fatalf("expected resource ID to be cleared, got %q", d.Id())
	}
}

package roxywi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceServiceInstallationReadRejectsMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
			return
		}
		_, _ = w.Write([]byte(`{"auto_start":1,"checker":0,"metrics":"invalid","docker":0,"server_id":1,"service":"haproxy"}`))
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	d := schema.TestResourceDataRaw(t, resourceServiceInstallation().Schema, nil)
	d.SetId("1-haproxy")

	diagnostics := resourceServiceInstallationRead(context.Background(), d, &Config{Client: client})
	if !diagnostics.HasError() {
		t.Fatal("expected malformed API data to produce a diagnostic")
	}
	if !strings.Contains(diagnostics[0].Summary, `field "metrics"`) {
		t.Fatalf("unexpected diagnostic: %s", diagnostics[0].Summary)
	}
}

package roxywi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestConfig(t *testing.T, handler http.HandlerFunc) *Config {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			writeTestJSON(t, w, map[string]interface{}{"access_token": "token"})
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatalf("create test client: %v", err)
	}
	return &Config{Client: client}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Errorf("encode test response: %v", err)
	}
}

package roxywi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestClientJoinsBaseURLAndEndpoint(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		if strings.HasSuffix(r.URL.Path, "/api/login") {
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("unexpected Authorization header %q", got)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	for _, baseURL := range []string{server.URL + "/roxy-wi", server.URL + "/roxy-wi/"} {
		client, err := NewClient(context.Background(), baseURL, "user", "password", "test-agent")
		if err != nil {
			t.Fatalf("NewClient(%q): %v", baseURL, err)
		}
		if _, err := client.doRequest(context.Background(), http.MethodGet, "/api/group/1", nil); err != nil {
			t.Fatalf("request with leading slash: %v", err)
		}
		if _, err := client.doRequest(context.Background(), http.MethodGet, "api/service/haproxy/1/section/global", nil); err != nil {
			t.Fatalf("request without leading slash: %v", err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	for _, path := range paths {
		if !strings.HasPrefix(path, "/roxy-wi/api/") {
			t.Errorf("URL path was not joined correctly: %q", path)
		}
	}
}

func TestClientRefreshesAuthenticationOnce(t *testing.T) {
	var mu sync.Mutex
	loginCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path == "/api/login" {
			loginCount++
			_, _ = fmt.Fprintf(w, `{"access_token":"token-%d"}`, loginCount)
			return
		}
		if r.Header.Get("Authorization") == "Bearer token-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.doRequest(context.Background(), http.MethodGet, "/api/group/1", nil); err != nil {
		t.Fatalf("request after token refresh: %v", err)
	}
	if loginCount != 2 {
		t.Fatalf("expected two authentication requests, got %d", loginCount)
	}
}

func TestClientReturnsTypedRedactedAPIError(t *testing.T) {
	const secret = "must-not-appear-in-error"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"access_token":"` + secret + `"}`))
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.doRequest(context.Background(), http.MethodGet, "/api/group/404", nil)
	if err == nil {
		t.Fatal("expected API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected typed 404 API error, got %T: %v", err, err)
	}
	if !isNotFound(err) {
		t.Fatal("expected isNotFound to recognize APIError")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("API error leaked response body: %v", err)
	}
}

func TestClientHonorsCancelledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			_, _ = w.Write([]byte(`{"access_token":"token"}`))
			return
		}
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := NewClient(context.Background(), server.URL, "user", "password", "test-agent")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.doRequest(ctx, http.MethodGet, "/api/group/1", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestNewClientRejectsInvalidBaseURL(t *testing.T) {
	for _, baseURL := range []string{"", "example.com", "ftp://example.com", "https://example.com?token=secret"} {
		if _, err := NewClient(context.Background(), baseURL, "user", "password", "test-agent"); err == nil {
			t.Errorf("expected invalid base URL %q to fail", baseURL)
		}
	}
}

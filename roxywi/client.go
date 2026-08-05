package roxywi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	login      string
	password   string
	userAgent  string
	token      string
	tokenMu    sync.RWMutex
	authMu     sync.Mutex
}

type APIError struct {
	StatusCode int
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Roxy-WI API request %s %s failed with status %d (%s)", e.Method, e.URL, e.StatusCode, http.StatusText(e.StatusCode))
}

func NewClient(ctx context.Context, baseURL, login, password, userAgent string) (*Client, error) {
	parsedBaseURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse Roxy-WI base URL: %w", err)
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, fmt.Errorf("Roxy-WI base URL must use http or https")
	}
	if parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("Roxy-WI base URL must include a host")
	}
	if parsedBaseURL.RawQuery != "" || parsedBaseURL.Fragment != "" {
		return nil, fmt.Errorf("Roxy-WI base URL must not include a query or fragment")
	}

	client := &Client{
		baseURL:    strings.TrimRight(parsedBaseURL.String(), "/"),
		httpClient: &http.Client{},
		login:      login,
		password:   password,
		userAgent:  userAgent,
	}

	if err := client.authenticate(ctx); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) authenticate(ctx context.Context) error {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	authURL, err := url.JoinPath(c.baseURL, "api/login")
	if err != nil {
		return fmt.Errorf("build authentication URL: %w", err)
	}
	authData := map[string]string{
		"login":    c.login,
		"password": c.password,
	}

	reqBody, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("encode authentication request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create authentication request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authenticate with Roxy-WI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read authentication response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &APIError{StatusCode: resp.StatusCode, Method: http.MethodPost, URL: authURL}
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("decode authentication response: %w", err)
	}

	token, ok := result["access_token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("authentication response does not contain an access token")
	}

	c.tokenMu.Lock()
	c.token = token
	c.tokenMu.Unlock()
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}) ([]byte, error) {
	return c.doRequestWithAuthRetry(ctx, method, endpoint, body, true)
}

func (c *Client) doRequestWithAuthRetry(ctx context.Context, method, endpoint string, body interface{}, retryAuth bool) ([]byte, error) {
	requestURL, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return nil, fmt.Errorf("build Roxy-WI API URL: %w", err)
	}

	var reqBody []byte
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode Roxy-WI API request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create Roxy-WI API request: %w", err)
	}

	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Roxy-WI API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read Roxy-WI API response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized && retryAuth {
		if err := c.authenticate(ctx); err != nil {
			return nil, fmt.Errorf("refresh Roxy-WI authentication: %w", err)
		}
		return c.doRequestWithAuthRetry(ctx, method, endpoint, body, false)
	}
	if resp.StatusCode == http.StatusNotFound && method == http.MethodDelete {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Method: method, URL: requestURL}
	}

	return respBody, nil
}

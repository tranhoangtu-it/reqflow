package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestBuildTokenRequest_ClientCredentials(t *testing.T) {
	config := domain.OAuth2Config{
		TokenURL:     "https://auth.example.com/token",
		ClientID:     "my-client-id",
		ClientSecret: "my-client-secret",
		Scope:        "read write",
		GrantType:    "client_credentials",
	}

	req := BuildTokenRequest(config)

	// Method should be POST.
	if req.Method != domain.MethodPost {
		t.Errorf("Method = %s, want POST", req.Method)
	}

	// URL should be the token URL.
	if req.URL != config.TokenURL {
		t.Errorf("URL = %s, want %s", req.URL, config.TokenURL)
	}

	// Content-Type should be application/x-www-form-urlencoded.
	var contentType string
	for _, h := range req.Headers {
		if h.Key == "Content-Type" {
			contentType = h.Value
			break
		}
	}
	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/x-www-form-urlencoded")
	}

	// Body should contain form parameters.
	body := string(req.Body)
	for _, want := range []string{"grant_type=client_credentials", "client_id=my-client-id", "client_secret=my-client-secret", "scope=read+write"} {
		if !containsStr(body, want) {
			t.Errorf("body missing %q, got: %s", want, body)
		}
	}
}

func TestBuildTokenRequest_PasswordGrant(t *testing.T) {
	config := domain.OAuth2Config{
		TokenURL:     "https://auth.example.com/token",
		ClientID:     "my-client-id",
		ClientSecret: "my-client-secret",
		Scope:        "read",
		GrantType:    "password",
		Username:     "user@example.com",
		Password:     "s3cret",
	}

	req := BuildTokenRequest(config)

	body := string(req.Body)
	for _, want := range []string{"grant_type=password", "username=user%40example.com", "password=s3cret"} {
		if !containsStr(body, want) {
			t.Errorf("body missing %q, got: %s", want, body)
		}
	}
}

func TestApplyOAuth2(t *testing.T) {
	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://api.example.com/resource",
		Headers: []domain.Header{
			{Key: "Accept", Value: "application/json"},
		},
	}

	got := ApplyOAuth2(req, "my-access-token")

	// Should have Authorization: Bearer header.
	var authValue string
	for _, h := range got.Headers {
		if h.Key == "Authorization" {
			authValue = h.Value
			break
		}
	}
	if authValue != "Bearer my-access-token" {
		t.Errorf("Authorization = %q, want %q", authValue, "Bearer my-access-token")
	}

	// Verify existing headers preserved.
	found := false
	for _, h := range got.Headers {
		if h.Key == "Accept" && h.Value == "application/json" {
			found = true
			break
		}
	}
	if !found {
		t.Error("existing Accept header was not preserved")
	}

	// Verify original not modified.
	for _, h := range req.Headers {
		if h.Key == "Authorization" {
			t.Error("original request was modified")
		}
	}
}

func TestFetchOAuth2Token_WithMockServer(t *testing.T) {
	// Create a mock OAuth2 token endpoint.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("grant_type = %q, want %q", r.FormValue("grant_type"), "client_credentials")
		}
		if r.FormValue("client_id") != "test-client" {
			t.Errorf("client_id = %q, want %q", r.FormValue("client_id"), "test-client")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-token-12345",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	config := domain.OAuth2Config{
		TokenURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Scope:        "read",
		GrantType:    "client_credentials",
	}

	token, err := FetchOAuth2Token(context.Background(), config, &testHTTPDoer{client: server.Client()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token != "mock-token-12345" {
		t.Errorf("token = %q, want %q", token, "mock-token-12345")
	}
}

func TestFetchOAuth2Token_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	config := domain.OAuth2Config{
		TokenURL:     server.URL,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		GrantType:    "client_credentials",
	}

	_, err := FetchOAuth2Token(context.Background(), config, &testHTTPDoer{client: server.Client()})
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

// testHTTPDoer is a simple adapter that wraps an *http.Client for testing
// the FetchOAuth2Token function.
type testHTTPDoer struct {
	client *http.Client
}

func (d *testHTTPDoer) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, string(req.Method), req.URL, nil)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	if req.Body != nil {
		httpReq.Body = http.NoBody
		httpReq.GetBody = nil
		// Re-create with body.
		bodyReader := strings.NewReader(string(req.Body))
		httpReq, err = http.NewRequestWithContext(ctx, string(req.Method), req.URL, bodyReader)
		if err != nil {
			return domain.HTTPResponse{}, err
		}
	}

	for _, h := range req.Headers {
		httpReq.Header.Set(h.Key, h.Value)
	}

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return domain.HTTPResponse{}, err
	}
	defer resp.Body.Close()

	bodyBytes := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if readErr != nil {
			break
		}
	}

	return domain.HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       bodyBytes,
	}, nil
}

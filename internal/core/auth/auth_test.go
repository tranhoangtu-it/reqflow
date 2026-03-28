package auth

import (
	"errors"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestApply(t *testing.T) {
	baseReq := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com",
		Headers: []domain.Header{
			{Key: "Accept", Value: "application/json"},
		},
	}

	tests := []struct {
		name      string
		req       domain.HTTPRequest
		config    *domain.AuthConfig
		wantErr   error
		checkFunc func(t *testing.T, got domain.HTTPRequest)
	}{
		{
			name:   "nil config returns unchanged request",
			req:    baseReq,
			config: nil,
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				if len(got.Headers) != len(baseReq.Headers) {
					t.Errorf("headers count = %d, want %d", len(got.Headers), len(baseReq.Headers))
				}
			},
		},
		{
			name: "type none returns unchanged request",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthNone,
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				if len(got.Headers) != len(baseReq.Headers) {
					t.Errorf("headers count = %d, want %d", len(got.Headers), len(baseReq.Headers))
				}
			},
		},
		{
			name: "basic auth with valid config",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthBasic,
				Basic: &domain.BasicAuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				var authValue string
				for _, h := range got.Headers {
					if h.Key == "Authorization" {
						authValue = h.Value
						break
					}
				}
				if authValue == "" {
					t.Error("expected Authorization header, got none")
				}
				if len(authValue) < len("Basic ") {
					t.Errorf("unexpected auth value: %q", authValue)
				}
			},
		},
		{
			name: "bearer auth with valid config",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthBearer,
				Bearer: &domain.BearerAuthConfig{
					Token: "mytoken",
				},
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				var authValue string
				for _, h := range got.Headers {
					if h.Key == "Authorization" {
						authValue = h.Value
						break
					}
				}
				if authValue != "Bearer mytoken" {
					t.Errorf("Authorization = %q, want %q", authValue, "Bearer mytoken")
				}
			},
		},
		{
			name: "apikey auth with valid config",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthAPIKey,
				APIKey: &domain.APIKeyAuthConfig{
					Key:      "X-API-Key",
					Value:    "secret123",
					Location: domain.APIKeyInHeader,
				},
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				found := false
				for _, h := range got.Headers {
					if h.Key == "X-API-Key" && h.Value == "secret123" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected X-API-Key header not found")
				}
			},
		},
		{
			name: "digest auth with valid config",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthDigest,
				Digest: &domain.DigestAuthConfig{
					Username: "admin",
					Password: "secret",
				},
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				// Digest auth through the dispatcher stores credentials
				// but doesn't add Authorization header (needs challenge first).
				// The dispatcher just validates the config is present.
				if len(got.Headers) < len(baseReq.Headers) {
					t.Errorf("headers count = %d, want at least %d", len(got.Headers), len(baseReq.Headers))
				}
			},
		},
		{
			name: "aws auth with valid config",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthAWS,
				AWS: &domain.AWSAuthConfig{
					AccessKey: "AKIDEXAMPLE",
					SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY",
					Region:    "us-east-1",
					Service:   "execute-api",
				},
			},
			checkFunc: func(t *testing.T, got domain.HTTPRequest) {
				var authValue string
				for _, h := range got.Headers {
					if h.Key == "Authorization" {
						authValue = h.Value
						break
					}
				}
				if authValue == "" {
					t.Error("expected Authorization header, got none")
				}
			},
		},
		{
			name: "oauth2 type returns error from dispatcher",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthOAuth2,
				OAuth2: &domain.OAuth2Config{
					TokenURL: "https://auth.example.com/token",
				},
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "unknown type returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: "unknown-type",
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "digest type with nil Digest field returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type:   domain.AuthDigest,
				Digest: nil,
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "aws type with nil AWS field returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type: domain.AuthAWS,
				AWS:  nil,
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "basic type with nil Basic field returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type:  domain.AuthBasic,
				Basic: nil,
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "bearer type with nil Bearer field returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type:   domain.AuthBearer,
				Bearer: nil,
			},
			wantErr: domain.ErrInvalidAuth,
		},
		{
			name: "apikey type with nil APIKey field returns error",
			req:  baseReq,
			config: &domain.AuthConfig{
				Type:   domain.AuthAPIKey,
				APIKey: nil,
			},
			wantErr: domain.ErrInvalidAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Apply(tt.req, tt.config)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	}
}

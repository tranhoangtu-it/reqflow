package auth

import (
	"encoding/base64"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestApplyBasic(t *testing.T) {
	tests := []struct {
		name           string
		req            domain.HTTPRequest
		config         domain.BasicAuthConfig
		wantAuthHeader string
	}{
		{
			name: "standard credentials",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BasicAuthConfig{Username: "admin", Password: "secret"},
			wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
		},
		{
			name: "empty password",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BasicAuthConfig{Username: "admin", Password: ""},
			wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:")),
		},
		{
			name: "empty username",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BasicAuthConfig{Username: "", Password: "secret"},
			wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte(":secret")),
		},
		{
			name: "special characters",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BasicAuthConfig{Username: "user@domain.com", Password: "p@ss:word"},
			wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("user@domain.com:p@ss:word")),
		},
		{
			name: "preserves existing headers",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "X-Custom", Value: "value"},
				},
			},
			config:         domain.BasicAuthConfig{Username: "admin", Password: "secret"},
			wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secret")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyBasic(tt.req, tt.config)

			// Verify auth header is present
			var authValue string
			for _, h := range got.Headers {
				if h.Key == "Authorization" {
					authValue = h.Value
					break
				}
			}
			if authValue != tt.wantAuthHeader {
				t.Errorf("Authorization header = %q, want %q", authValue, tt.wantAuthHeader)
			}

			// Verify existing headers are preserved
			for _, orig := range tt.req.Headers {
				found := false
				for _, h := range got.Headers {
					if h.Key == orig.Key && h.Value == orig.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("existing header %q: %q was not preserved", orig.Key, orig.Value)
				}
			}

			// Verify original request was not modified
			for _, h := range tt.req.Headers {
				if h.Key == "Authorization" {
					t.Error("original request was modified: Authorization header found")
				}
			}
		})
	}
}

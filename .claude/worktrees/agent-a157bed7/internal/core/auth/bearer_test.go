package auth

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestApplyBearer(t *testing.T) {
	tests := []struct {
		name           string
		req            domain.HTTPRequest
		config         domain.BearerAuthConfig
		wantAuthHeader string
	}{
		{
			name: "standard token",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BearerAuthConfig{Token: "mytoken123"},
			wantAuthHeader: "Bearer mytoken123",
		},
		{
			name: "custom prefix",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BearerAuthConfig{Token: "abc", Prefix: "Token"},
			wantAuthHeader: "Token abc",
		},
		{
			name: "empty prefix defaults to Bearer",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BearerAuthConfig{Token: "abc", Prefix: ""},
			wantAuthHeader: "Bearer abc",
		},
		{
			name: "long JWT token",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config:         domain.BearerAuthConfig{Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
			wantAuthHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
		{
			name: "preserves existing headers",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
					{Key: "Accept", Value: "text/html"},
				},
			},
			config:         domain.BearerAuthConfig{Token: "mytoken123"},
			wantAuthHeader: "Bearer mytoken123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyBearer(tt.req, tt.config)

			// Verify auth header
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

			// Verify existing headers preserved
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

			// Verify original not modified
			for _, h := range tt.req.Headers {
				if h.Key == "Authorization" {
					t.Error("original request was modified: Authorization header found")
				}
			}
		})
	}
}

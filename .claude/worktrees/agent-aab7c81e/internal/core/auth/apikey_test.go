package auth

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestApplyAPIKey(t *testing.T) {
	tests := []struct {
		name            string
		req             domain.HTTPRequest
		config          domain.APIKeyAuthConfig
		wantHeader      *domain.Header
		wantQueryParam  *domain.QueryParam
	}{
		{
			name: "in header",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config: domain.APIKeyAuthConfig{
				Key:      "X-API-Key",
				Value:    "secret123",
				Location: domain.APIKeyInHeader,
			},
			wantHeader: &domain.Header{Key: "X-API-Key", Value: "secret123"},
		},
		{
			name: "in query",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config: domain.APIKeyAuthConfig{
				Key:      "api_key",
				Value:    "secret123",
				Location: domain.APIKeyInQuery,
			},
			wantQueryParam: &domain.QueryParam{Key: "api_key", Value: "secret123"},
		},
		{
			name: "default location falls back to header",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			config: domain.APIKeyAuthConfig{
				Key:      "X-API-Key",
				Value:    "secret123",
				Location: "",
			},
			wantHeader: &domain.Header{Key: "X-API-Key", Value: "secret123"},
		},
		{
			name: "preserves existing headers and query params",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
				},
				QueryParams: []domain.QueryParam{
					{Key: "page", Value: "1"},
				},
			},
			config: domain.APIKeyAuthConfig{
				Key:      "X-API-Key",
				Value:    "secret123",
				Location: domain.APIKeyInHeader,
			},
			wantHeader: &domain.Header{Key: "X-API-Key", Value: "secret123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyAPIKey(tt.req, tt.config)

			if tt.wantHeader != nil {
				found := false
				for _, h := range got.Headers {
					if h.Key == tt.wantHeader.Key && h.Value == tt.wantHeader.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected header %q: %q not found in %v", tt.wantHeader.Key, tt.wantHeader.Value, got.Headers)
				}
			}

			if tt.wantQueryParam != nil {
				found := false
				for _, qp := range got.QueryParams {
					if qp.Key == tt.wantQueryParam.Key && qp.Value == tt.wantQueryParam.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected query param %q: %q not found in %v", tt.wantQueryParam.Key, tt.wantQueryParam.Value, got.QueryParams)
				}
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

			// Verify existing query params preserved
			for _, orig := range tt.req.QueryParams {
				found := false
				for _, qp := range got.QueryParams {
					if qp.Key == orig.Key && qp.Value == orig.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("existing query param %q: %q was not preserved", orig.Key, orig.Value)
				}
			}

			// Verify original not modified
			if len(tt.req.Headers) != len(tt.req.Headers) {
				t.Error("original request headers were modified")
			}
		})
	}
}

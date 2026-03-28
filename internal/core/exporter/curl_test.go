package exporter

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/importer"
	"github.com/ye-kart/reqflow/internal/domain"
)

// parseForRoundTrip delegates to the importer for round-trip fidelity tests.
func parseForRoundTrip(curlCmd string) (domain.RequestConfig, error) {
	return importer.ParseCurl(curlCmd)
}

func TestExportCurl(t *testing.T) {
	tests := []struct {
		name   string
		config domain.RequestConfig
		want   string
	}{
		{
			name: "simple GET",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			want: "curl https://example.com",
		},
		{
			name: "POST with body",
			config: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"key":"val"}`),
			},
			want: "curl -X POST \\\n  -d '{\"key\":\"val\"}' \\\n  https://example.com",
		},
		{
			name: "with headers",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept", Value: "application/json"},
					{Key: "X-Custom", Value: "value"},
				},
			},
			want: "curl \\\n  -H 'Accept: application/json' \\\n  -H 'X-Custom: value' \\\n  https://example.com",
		},
		{
			name: "with basic auth",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Auth: &domain.AuthConfig{
					Type: domain.AuthBasic,
					Basic: &domain.BasicAuthConfig{
						Username: "admin",
						Password: "secret",
					},
				},
			},
			want: "curl \\\n  -u 'admin:secret' \\\n  https://example.com",
		},
		{
			name: "with bearer auth",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Auth: &domain.AuthConfig{
					Type: domain.AuthBearer,
					Bearer: &domain.BearerAuthConfig{
						Token: "mytoken123",
					},
				},
			},
			want: "curl \\\n  -H 'Authorization: Bearer mytoken123' \\\n  https://example.com",
		},
		{
			name: "PUT with body and headers",
			config: domain.RequestConfig{
				Method: domain.MethodPut,
				URL:    "https://example.com/resource",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
				},
				Body: []byte(`{"update":true}`),
			},
			want: "curl -X PUT \\\n  -H 'Content-Type: application/json' \\\n  -d '{\"update\":true}' \\\n  https://example.com/resource",
		},
		{
			name: "DELETE method",
			config: domain.RequestConfig{
				Method: domain.MethodDelete,
				URL:    "https://example.com/resource/1",
			},
			want: "curl -X DELETE https://example.com/resource/1",
		},
		{
			name: "bearer auth with custom prefix",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Auth: &domain.AuthConfig{
					Type: domain.AuthBearer,
					Bearer: &domain.BearerAuthConfig{
						Token:  "mytoken",
						Prefix: "Token",
					},
				},
			},
			want: "curl \\\n  -H 'Authorization: Token mytoken' \\\n  https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExportCurl(tt.config)
			if got != tt.want {
				t.Errorf("ExportCurl() =\n%s\n\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		config domain.RequestConfig
	}{
		{
			name: "simple GET round-trip",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
		},
		{
			name: "POST with body round-trip",
			config: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com/api",
				Body:   []byte(`{"key":"value"}`),
			},
		},
		{
			name: "with headers round-trip",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept", Value: "application/json"},
				},
			},
		},
		{
			name: "with basic auth round-trip",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Auth: &domain.AuthConfig{
					Type: domain.AuthBasic,
					Basic: &domain.BasicAuthConfig{
						Username: "user",
						Password: "pass",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exported := ExportCurl(tt.config)
			reimported, err := parseForRoundTrip(exported)
			if err != nil {
				t.Fatalf("round-trip parse failed: %v", err)
			}

			if reimported.Method != tt.config.Method {
				t.Errorf("Method = %q, want %q", reimported.Method, tt.config.Method)
			}
			if reimported.URL != tt.config.URL {
				t.Errorf("URL = %q, want %q", reimported.URL, tt.config.URL)
			}
			if string(reimported.Body) != string(tt.config.Body) {
				t.Errorf("Body = %q, want %q", string(reimported.Body), string(tt.config.Body))
			}
			if len(reimported.Headers) != len(tt.config.Headers) {
				t.Errorf("Headers len = %d, want %d", len(reimported.Headers), len(tt.config.Headers))
			}
			if tt.config.Auth != nil && reimported.Auth != nil {
				if reimported.Auth.Type != tt.config.Auth.Type {
					t.Errorf("Auth.Type = %q, want %q", reimported.Auth.Type, tt.config.Auth.Type)
				}
			}
		})
	}
}

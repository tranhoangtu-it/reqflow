package request

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestBuildRequest(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.RequestConfig
		vars    map[string]string
		want    domain.HTTPRequest
		wantErr bool
	}{
		{
			name: "simple GET request",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "POST with JSON body auto-detects content-type",
			config: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com/api",
				Body:   []byte(`{"name":"test"}`),
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method:      domain.MethodPost,
				URL:         "https://example.com/api",
				Body:        []byte(`{"name":"test"}`),
				ContentType: "application/json",
			},
			wantErr: false,
		},
		{
			name: "request with headers",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept", Value: "application/json"},
					{Key: "X-Custom", Value: "value"},
				},
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept", Value: "application/json"},
					{Key: "X-Custom", Value: "value"},
				},
			},
			wantErr: false,
		},
		{
			name: "request with query params",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com/search",
				QueryParams: []domain.QueryParam{
					{Key: "q", Value: "golang"},
					{Key: "page", Value: "1"},
				},
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com/search",
				QueryParams: []domain.QueryParam{
					{Key: "q", Value: "golang"},
					{Key: "page", Value: "1"},
				},
			},
			wantErr: false,
		},
		{
			name: "variable substitution in URL",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://{{host}}/api/v1",
			},
			vars: map[string]string{"host": "api.example.com"},
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://api.example.com/api/v1",
			},
			wantErr: false,
		},
		{
			name: "variable substitution in headers",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Authorization", Value: "Bearer {{token}}"},
				},
			},
			vars: map[string]string{"token": "abc123"},
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Authorization", Value: "Bearer abc123"},
				},
			},
			wantErr: false,
		},
		{
			name: "variable substitution in query params",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com/search",
				QueryParams: []domain.QueryParam{
					{Key: "q", Value: "{{query}}"},
				},
			},
			vars: map[string]string{"query": "golang"},
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com/search",
				QueryParams: []domain.QueryParam{
					{Key: "q", Value: "golang"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple variables in one string",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://{{host}}:{{port}}/api",
			},
			vars: map[string]string{"host": "localhost", "port": "8080"},
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://localhost:8080/api",
			},
			wantErr: false,
		},
		{
			name: "missing variable left as-is",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://{{host}}/api",
			},
			vars:    map[string]string{},
			want: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://{{host}}/api",
			},
			wantErr: false,
		},
		{
			name: "explicit content-type is preserved",
			config: domain.RequestConfig{
				Method:      domain.MethodPost,
				URL:         "https://example.com/api",
				Body:        []byte(`{"name":"test"}`),
				ContentType: "text/plain",
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method:      domain.MethodPost,
				URL:         "https://example.com/api",
				Body:        []byte(`{"name":"test"}`),
				ContentType: "text/plain",
			},
			wantErr: false,
		},
		{
			name: "empty body produces no content-type",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			vars: nil,
			want: domain.HTTPRequest{
				Method:      domain.MethodGet,
				URL:         "https://example.com",
				ContentType: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildRequest(tt.config, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Method != tt.want.Method {
				t.Errorf("Method = %v, want %v", got.Method, tt.want.Method)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %v, want %v", got.URL, tt.want.URL)
			}
			if got.ContentType != tt.want.ContentType {
				t.Errorf("ContentType = %q, want %q", got.ContentType, tt.want.ContentType)
			}
			if string(got.Body) != string(tt.want.Body) {
				t.Errorf("Body = %q, want %q", got.Body, tt.want.Body)
			}

			if len(got.Headers) != len(tt.want.Headers) {
				t.Errorf("Headers length = %d, want %d", len(got.Headers), len(tt.want.Headers))
			} else {
				for i, h := range got.Headers {
					if h.Key != tt.want.Headers[i].Key || h.Value != tt.want.Headers[i].Value {
						t.Errorf("Header[%d] = {%q, %q}, want {%q, %q}",
							i, h.Key, h.Value, tt.want.Headers[i].Key, tt.want.Headers[i].Value)
					}
				}
			}

			if len(got.QueryParams) != len(tt.want.QueryParams) {
				t.Errorf("QueryParams length = %d, want %d", len(got.QueryParams), len(tt.want.QueryParams))
			} else {
				for i, qp := range got.QueryParams {
					if qp.Key != tt.want.QueryParams[i].Key || qp.Value != tt.want.QueryParams[i].Value {
						t.Errorf("QueryParam[%d] = {%q, %q}, want {%q, %q}",
							i, qp.Key, qp.Value, tt.want.QueryParams[i].Key, tt.want.QueryParams[i].Value)
					}
				}
			}
		})
	}
}

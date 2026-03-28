package importer

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestParseCurl(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    domain.RequestConfig
		wantErr bool
	}{
		{
			name:  "simple GET",
			input: "curl https://example.com",
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
		},
		{
			name:  "GET with explicit method",
			input: "curl -X GET https://example.com",
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
		},
		{
			name:  "POST with data",
			input: `curl -X POST -d '{"key":"val"}' https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"key":"val"}`),
			},
		},
		{
			name:  "POST implied by -d",
			input: `curl -d '{"key":"val"}' https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"key":"val"}`),
			},
		},
		{
			name:  "multiple headers",
			input: `curl -H "Accept: json" -H "Auth: Bearer tok" https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept", Value: "json"},
					{Key: "Auth", Value: "Bearer tok"},
				},
			},
		},
		{
			name:  "basic auth",
			input: "curl -u admin:secret https://example.com",
			want: domain.RequestConfig{
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
		},
		{
			name:  "user agent",
			input: `curl -A "MyApp/1.0" https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "User-Agent", Value: "MyApp/1.0"},
				},
			},
		},
		{
			name:  "multiline with backslash",
			input: "curl \\\n  -X POST \\\n  https://example.com",
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
			},
		},
		{
			name:  "URL without curl prefix",
			input: "https://example.com",
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
		},
		{
			name:  "single-quoted body",
			input: `curl -d '{"a":"b"}' https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"a":"b"}`),
			},
		},
		{
			name:  "double-quoted body",
			input: `curl -d "{\"a\":\"b\"}" https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"a":"b"}`),
			},
		},
		{
			name:    "empty input returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid input no URL",
			input:   "curl -X GET",
			wantErr: true,
		},
		{
			name:  "data-raw flag",
			input: `curl --data-raw '{"key":"val"}' https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Body:   []byte(`{"key":"val"}`),
			},
		},
		{
			name:  "long form flags",
			input: `curl --request POST --header "Content-Type: application/json" --data '{"x":1}' https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodPost,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
				},
				Body: []byte(`{"x":1}`),
			},
		},
		{
			name:  "cookie flag",
			input: `curl -b "session=abc123" https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Cookie", Value: "session=abc123"},
				},
			},
		},
		{
			name:  "compressed flag adds Accept-Encoding",
			input: `curl --compressed https://example.com`,
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
				Headers: []domain.Header{
					{Key: "Accept-Encoding", Value: "gzip, deflate, br"},
				},
			},
		},
		{
			name:  "URL with quotes",
			input: `curl "https://example.com/path?q=1"`,
			want: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com/path?q=1",
			},
		},
		{
			name:  "PUT method",
			input: `curl -X PUT -d '{"update":true}' https://example.com/resource`,
			want: domain.RequestConfig{
				Method: domain.MethodPut,
				URL:    "https://example.com/resource",
				Body:   []byte(`{"update":true}`),
			},
		},
		{
			name:  "DELETE method",
			input: `curl -X DELETE https://example.com/resource/1`,
			want: domain.RequestConfig{
				Method: domain.MethodDelete,
				URL:    "https://example.com/resource/1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCurl(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Method != tt.want.Method {
				t.Errorf("Method = %q, want %q", got.Method, tt.want.Method)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
			if string(got.Body) != string(tt.want.Body) {
				t.Errorf("Body = %q, want %q", string(got.Body), string(tt.want.Body))
			}

			if len(got.Headers) != len(tt.want.Headers) {
				t.Errorf("Headers len = %d, want %d; got %+v", len(got.Headers), len(tt.want.Headers), got.Headers)
			} else {
				for i, h := range got.Headers {
					if h.Key != tt.want.Headers[i].Key || h.Value != tt.want.Headers[i].Value {
						t.Errorf("Header[%d] = {%q, %q}, want {%q, %q}", i, h.Key, h.Value, tt.want.Headers[i].Key, tt.want.Headers[i].Value)
					}
				}
			}

			if tt.want.Auth != nil {
				if got.Auth == nil {
					t.Fatalf("Auth = nil, want %+v", tt.want.Auth)
				}
				if got.Auth.Type != tt.want.Auth.Type {
					t.Errorf("Auth.Type = %q, want %q", got.Auth.Type, tt.want.Auth.Type)
				}
				if tt.want.Auth.Basic != nil {
					if got.Auth.Basic == nil {
						t.Fatalf("Auth.Basic = nil, want %+v", tt.want.Auth.Basic)
					}
					if got.Auth.Basic.Username != tt.want.Auth.Basic.Username {
						t.Errorf("Auth.Basic.Username = %q, want %q", got.Auth.Basic.Username, tt.want.Auth.Basic.Username)
					}
					if got.Auth.Basic.Password != tt.want.Auth.Basic.Password {
						t.Errorf("Auth.Basic.Password = %q, want %q", got.Auth.Basic.Password, tt.want.Auth.Basic.Password)
					}
				}
			} else if got.Auth != nil {
				t.Errorf("Auth = %+v, want nil", got.Auth)
			}
		})
	}
}

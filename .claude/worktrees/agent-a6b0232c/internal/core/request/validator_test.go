package request

import (
	"errors"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     domain.HTTPRequest
		wantErr error
	}{
		{
			name: "valid GET request",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			wantErr: nil,
		},
		{
			name: "valid URL with query params",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "https://example.com/api?key=value",
			},
			wantErr: nil,
		},
		{
			name: "valid URL with http scheme",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "http://localhost:8080/api",
			},
			wantErr: nil,
		},
		{
			name: "valid URL with https scheme",
			req: domain.HTTPRequest{
				Method: domain.MethodPost,
				URL:    "https://api.example.com/v1/users",
			},
			wantErr: nil,
		},
		{
			name: "empty URL returns ErrEmptyURL",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "",
			},
			wantErr: domain.ErrEmptyURL,
		},
		{
			name: "invalid method returns ErrInvalidMethod",
			req: domain.HTTPRequest{
				Method: "INVALID",
				URL:    "https://example.com",
			},
			wantErr: domain.ErrInvalidMethod,
		},
		{
			name: "unparseable URL returns ErrInvalidURL",
			req: domain.HTTPRequest{
				Method: domain.MethodGet,
				URL:    "://missing-scheme",
			},
			wantErr: domain.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.req)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateRequest() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateRequest() expected error %v, got nil", tt.wantErr)
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateRequest() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.RequestConfig
		wantErr error
	}{
		{
			name: "valid config",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "https://example.com",
			},
			wantErr: nil,
		},
		{
			name: "empty URL returns ErrEmptyURL",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "",
			},
			wantErr: domain.ErrEmptyURL,
		},
		{
			name: "invalid method returns ErrInvalidMethod",
			config: domain.RequestConfig{
				Method: "BADMETHOD",
				URL:    "https://example.com",
			},
			wantErr: domain.ErrInvalidMethod,
		},
		{
			name: "unparseable URL returns ErrInvalidURL",
			config: domain.RequestConfig{
				Method: domain.MethodGet,
				URL:    "://no-scheme",
			},
			wantErr: domain.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("ValidateConfig() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateConfig() expected error %v, got nil", tt.wantErr)
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateConfig() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

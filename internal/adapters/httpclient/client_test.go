package httpclient_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/httpclient"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestDo_BasicRequests(t *testing.T) {
	tests := []struct {
		name           string
		method         domain.HTTPMethod
		serverHandler  http.HandlerFunc
		body           []byte
		wantStatusCode int
		wantBody       string
	}{
		{
			name:   "GET request returns correct status code and body",
			method: domain.MethodGet,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("hello world"))
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "hello world",
		},
		{
			name:   "POST request sends body correctly",
			method: domain.MethodPost,
			body:   []byte(`{"key":"value"}`),
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST, got %s", r.Method)
				}
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()
				if string(body) != `{"key":"value"}` {
					t.Errorf("expected body %q, got %q", `{"key":"value"}`, string(body))
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created"))
			},
			wantStatusCode: http.StatusCreated,
			wantBody:       "created",
		},
		{
			name:   "PUT request with JSON body",
			method: domain.MethodPut,
			body:   []byte(`{"updated":true}`),
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("expected PUT, got %s", r.Method)
				}
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()
				if string(body) != `{"updated":true}` {
					t.Errorf("expected body %q, got %q", `{"updated":true}`, string(body))
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("updated"))
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "updated",
		},
		{
			name:   "DELETE request works",
			method: domain.MethodDelete,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(http.StatusNoContent)
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       "",
		},
		{
			name:   "HEAD request returns no body",
			method: domain.MethodHead,
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "HEAD" {
					t.Errorf("expected HEAD, got %s", r.Method)
				}
				w.Header().Set("X-Custom", "present")
				w.WriteHeader(http.StatusOK)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.serverHandler)
			defer srv.Close()

			client := httpclient.New()
			resp, err := client.Do(context.Background(), domain.HTTPRequest{
				Method: tt.method,
				URL:    srv.URL,
				Body:   tt.body,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("status code: want %d, got %d", tt.wantStatusCode, resp.StatusCode)
			}

			if string(resp.Body) != tt.wantBody {
				t.Errorf("body: want %q, got %q", tt.wantBody, string(resp.Body))
			}
		})
	}
}

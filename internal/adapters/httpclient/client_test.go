package httpclient_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
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

func TestDo_Headers(t *testing.T) {
	t.Run("request headers are sent to server", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("X-Custom-Header"); got != "custom-value" {
				t.Errorf("X-Custom-Header: want %q, got %q", "custom-value", got)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer token123" {
				t.Errorf("Authorization: want %q, got %q", "Bearer token123", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
			Headers: []domain.Header{
				{Key: "X-Custom-Header", Value: "custom-value"},
				{Key: "Authorization", Value: "Bearer token123"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("response headers are captured", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Response-Id", "abc123")
			w.Header().Set("X-Request-Duration", "42ms")
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		headerMap := make(map[string]string)
		for _, h := range resp.Headers {
			headerMap[h.Key] = h.Value
		}
		if headerMap["X-Response-Id"] != "abc123" {
			t.Errorf("X-Response-Id: want %q, got %q", "abc123", headerMap["X-Response-Id"])
		}
		if headerMap["X-Request-Duration"] != "42ms" {
			t.Errorf("X-Request-Duration: want %q, got %q", "42ms", headerMap["X-Request-Duration"])
		}
	})

	t.Run("Content-Type is set from req.ContentType", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type: want %q, got %q", "application/json", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method:      domain.MethodPost,
			URL:         srv.URL,
			Body:        []byte(`{"a":1}`),
			ContentType: "application/json",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple headers with same key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values := r.Header.Values("X-Multi")
			sort.Strings(values)
			want := "value1,value2"
			got := strings.Join(values, ",")
			if got != want {
				t.Errorf("X-Multi values: want %q, got %q", want, got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
			Headers: []domain.Header{
				{Key: "X-Multi", Value: "value1"},
				{Key: "X-Multi", Value: "value2"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestDo_QueryParams(t *testing.T) {
	t.Run("query params are appended to URL", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("foo"); got != "bar" {
				t.Errorf("foo: want %q, got %q", "bar", got)
			}
			if got := r.URL.Query().Get("baz"); got != "qux" {
				t.Errorf("baz: want %q, got %q", "qux", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
			QueryParams: []domain.QueryParam{
				{Key: "foo", Value: "bar"},
				{Key: "baz", Value: "qux"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("query params are URL-encoded", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("msg"); got != "hello world&more" {
				t.Errorf("msg: want %q, got %q", "hello world&more", got)
			}
			if got := r.URL.Query().Get("path"); got != "/a/b?c=d" {
				t.Errorf("path: want %q, got %q", "/a/b?c=d", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
			QueryParams: []domain.QueryParam{
				{Key: "msg", Value: "hello world&more"},
				{Key: "path", Value: "/a/b?c=d"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("existing URL query params are preserved and merged", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("existing"); got != "yes" {
				t.Errorf("existing: want %q, got %q", "yes", got)
			}
			if got := r.URL.Query().Get("added"); got != "new" {
				t.Errorf("added: want %q, got %q", "new", got)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL + "?existing=yes",
			QueryParams: []domain.QueryParam{
				{Key: "added", Value: "new"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

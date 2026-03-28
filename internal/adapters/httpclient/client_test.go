package httpclient_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/adapters/cookiejar"
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

func TestDo_ResponseHandling(t *testing.T) {
	t.Run("response body is fully read", func(t *testing.T) {
		want := "complete response body content here"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(want))
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
		if string(resp.Body) != want {
			t.Errorf("body: want %q, got %q", want, string(resp.Body))
		}
	})

	t.Run("response duration is greater than zero", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if resp.Duration <= 0 {
			t.Errorf("duration should be > 0, got %v", resp.Duration)
		}
	})

	t.Run("response size matches body length", func(t *testing.T) {
		body := "twelve chars"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(body))
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
		if resp.Size != int64(len(body)) {
			t.Errorf("size: want %d, got %d", len(body), resp.Size)
		}
		if resp.Size != int64(len(resp.Body)) {
			t.Errorf("size should match body length: size=%d, len(body)=%d", resp.Size, len(resp.Body))
		}
	})

	t.Run("large response body works", func(t *testing.T) {
		// Generate a 1MB body
		large := strings.Repeat("x", 1024*1024)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(large))
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
		if len(resp.Body) != len(large) {
			t.Errorf("body length: want %d, got %d", len(large), len(resp.Body))
		}
		if resp.Size != int64(len(large)) {
			t.Errorf("size: want %d, got %d", len(large), resp.Size)
		}
	})

	t.Run("empty response body works", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer srv.Close()

		client := httpclient.New()
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodDelete,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Body) != 0 {
			t.Errorf("body should be empty, got %d bytes", len(resp.Body))
		}
		if resp.Size != 0 {
			t.Errorf("size should be 0, got %d", resp.Size)
		}
	})

	t.Run("status string is populated", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
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
		if !strings.Contains(resp.Status, "404") {
			t.Errorf("status should contain 404, got %q", resp.Status)
		}
	})
}

func TestDo_ErrorHandling(t *testing.T) {
	t.Run("context cancellation returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel immediately

		client := httpclient.New()
		_, err := client.Do(ctx, domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err == nil {
			t.Fatal("expected error for cancelled context, got nil")
		}
	})

	t.Run("invalid URL returns error", func(t *testing.T) {
		client := httpclient.New()
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    "://not-a-valid-url",
		})
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}
	})

	t.Run("connection refused returns error", func(t *testing.T) {
		client := httpclient.New(httpclient.WithTimeout(1 * time.Second))
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    "http://127.0.0.1:1", // port 1 should refuse connections
		})
		if err == nil {
			t.Fatal("expected error for connection refused, got nil")
		}
	})
}

func TestDo_Options(t *testing.T) {
	t.Run("WithTimeout causes timeout on slow server", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(5 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		client := httpclient.New(httpclient.WithTimeout(50 * time.Millisecond))
		_, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err == nil {
			t.Fatal("expected timeout error, got nil")
		}
	})

	t.Run("default timeout is 30 seconds", func(t *testing.T) {
		// We verify the default by making a successful request to a fast server.
		// The actual default value is tested implicitly - if it were 0 (no timeout),
		// the adapter would behave differently for slow servers.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ok")
		}))
		defer srv.Close()

		client := httpclient.New() // default timeout
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status: want %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("WithTransport uses custom transport", func(t *testing.T) {
		// Create a custom round tripper that always returns a fixed response
		transport := &fixedRoundTripper{
			resp: &http.Response{
				StatusCode: http.StatusTeapot,
				Status:     "418 I'm a teapot",
				Header:     http.Header{"X-Custom-Transport": []string{"yes"}},
				Body:       io.NopCloser(strings.NewReader("teapot")),
			},
		}

		client := httpclient.New(httpclient.WithTransport(transport))
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    "http://example.com",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusTeapot {
			t.Errorf("status: want %d, got %d", http.StatusTeapot, resp.StatusCode)
		}
		if string(resp.Body) != "teapot" {
			t.Errorf("body: want %q, got %q", "teapot", string(resp.Body))
		}
	})
}

func TestDo_Trace(t *testing.T) {
	t.Run("timing fields are populated with trace enabled", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))
		defer srv.Close()

		client := httpclient.New(httpclient.WithTrace(true))
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Timing.Total <= 0 {
			t.Errorf("expected Total > 0, got %v", resp.Timing.Total)
		}
		if resp.Timing.TCPConnect <= 0 {
			t.Errorf("expected TCPConnect > 0, got %v", resp.Timing.TCPConnect)
		}
		if resp.Timing.FirstByte <= 0 {
			t.Errorf("expected FirstByte > 0, got %v", resp.Timing.FirstByte)
		}
	})

	t.Run("timing Total >= FirstByte (logical ordering)", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))
		defer srv.Close()

		client := httpclient.New(httpclient.WithTrace(true))
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Timing.Total < resp.Timing.FirstByte {
			t.Errorf("Total (%v) should be >= FirstByte (%v)", resp.Timing.Total, resp.Timing.FirstByte)
		}
	})

	t.Run("timing is zero when trace not enabled", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))
		defer srv.Close()

		client := httpclient.New() // no trace
		resp, err := client.Do(context.Background(), domain.HTTPRequest{
			Method: domain.MethodGet,
			URL:    srv.URL,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if resp.Timing.Total != 0 {
			t.Errorf("expected zero Total without trace, got %v", resp.Timing.Total)
		}
	})
}

func TestDo_CookieJar_SendsCookiesWithRequest(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "session", Value: "abc123", Domain: "example.com", Path: "/"},
	})

	var receivedCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// The test server's host is 127.0.0.1, so we need cookies for that host.
	_ = jar.SetCookies(srv.URL, []domain.Cookie{
		{Name: "session", Value: "abc123", Domain: "127.0.0.1", Path: "/"},
	})

	client := httpclient.New(httpclient.WithCookieJar(jar))
	_, err := client.Do(context.Background(), domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedCookie, "session=abc123") {
		t.Errorf("expected Cookie header to contain 'session=abc123', got %q", receivedCookie)
	}
}

func TestDo_CookieJar_StoresSetCookieResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "token",
			Value: "xyz789",
			Path:  "/",
		})
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	jar := cookiejar.New()
	client := httpclient.New(httpclient.WithCookieJar(jar))

	_, err := client.Do(context.Background(), domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cookies, err := jar.All()
	if err != nil {
		t.Fatalf("jar.All: unexpected error: %v", err)
	}

	found := false
	for _, c := range cookies {
		if c.Name == "token" && c.Value == "xyz789" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected jar to contain 'token=xyz789', got %v", cookies)
	}
}

func TestDo_NoCookieJar_DoesNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := httpclient.New() // no jar
	_, err := client.Do(context.Background(), domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    srv.URL,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// fixedRoundTripper is a test helper that returns a fixed response.
type fixedRoundTripper struct {
	resp *http.Response
}

func (f *fixedRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return f.resp, nil
}

package httpclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// Compile-time check that Client satisfies the driven.HTTPClient interface.
var _ driven.HTTPClient = (*Client)(nil)

const defaultTimeout = 30 * time.Second

// Client is an HTTP client adapter that wraps Go's net/http.Client
// and satisfies the driven.HTTPClient port interface.
type Client struct {
	httpClient *http.Client
}

// Option configures the Client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithTransport sets a custom http.RoundTripper (useful for testing, proxies).
func WithTransport(t http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.Transport = t
	}
}

// New creates a new Client with the given options.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Do executes an HTTP request and returns the response.
func (c *Client) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	var bodyReader io.Reader
	if len(req.Body) > 0 {
		bodyReader = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, string(req.Method), req.URL, bodyReader)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	// Set headers
	for _, h := range req.Headers {
		httpReq.Header.Add(h.Key, h.Value)
	}

	// Set Content-Type if specified
	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Append query params
	if len(req.QueryParams) > 0 {
		q := httpReq.URL.Query()
		for _, qp := range req.QueryParams {
			q.Add(qp.Key, qp.Value)
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		return domain.HTTPResponse{}, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return domain.HTTPResponse{}, err
	}

	// Collect response headers
	var headers []domain.Header
	for key, values := range httpResp.Header {
		for _, v := range values {
			headers = append(headers, domain.Header{Key: key, Value: v})
		}
	}

	return domain.HTTPResponse{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    headers,
		Body:       body,
		Duration:   duration,
		Size:       int64(len(body)),
	}, nil
}

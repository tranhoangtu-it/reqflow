package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

// Client is a minimal HTTP client adapter that implements driven.HTTPClient.
type Client struct {
	http *http.Client
}

// New creates a new Client with sensible defaults.
func New() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

// Do sends an HTTP request and returns the response.
func (c *Client) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	// Build the URL with query parameters.
	reqURL, err := buildURL(req.URL, req.QueryParams)
	if err != nil {
		return domain.HTTPResponse{}, fmt.Errorf("building URL: %w", err)
	}

	// Create the stdlib request.
	var body io.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, string(req.Method), reqURL, body)
	if err != nil {
		return domain.HTTPResponse{}, fmt.Errorf("creating request: %w", err)
	}

	// Set headers.
	for _, h := range req.Headers {
		httpReq.Header.Set(h.Key, h.Value)
	}
	if req.ContentType != "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	// Send the request.
	start := time.Now()
	resp, err := c.http.Do(httpReq)
	duration := time.Since(start)

	if err != nil {
		return domain.HTTPResponse{}, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.HTTPResponse{}, fmt.Errorf("reading response body: %w", err)
	}

	// Convert response headers.
	var headers []domain.Header
	for key, values := range resp.Header {
		for _, v := range values {
			headers = append(headers, domain.Header{Key: key, Value: v})
		}
	}

	return domain.HTTPResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headers,
		Body:       respBody,
		Duration:   duration,
		Size:       int64(len(respBody)),
	}, nil
}

// buildURL appends query parameters to the base URL.
func buildURL(baseURL string, params []domain.QueryParam) (string, error) {
	if len(params) == 0 {
		return baseURL, nil
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for _, p := range params {
		q.Add(p.Key, p.Value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

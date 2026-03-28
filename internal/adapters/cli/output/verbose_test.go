package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestFormatVerbose_GETRequest_ShowsMethodPathHost(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://api.example.com/api/users",
		Headers: []domain.Header{
			{Key: "Accept", Value: "application/json"},
		},
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"users":[]}`),
		Duration: 100 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Request line: > GET /api/users HTTP/1.1
	if !strings.Contains(output, "> GET /api/users HTTP/1.1") {
		t.Errorf("expected request line '> GET /api/users HTTP/1.1', got:\n%s", output)
	}

	// Host header
	if !strings.Contains(output, "> Host: api.example.com") {
		t.Errorf("expected host header '> Host: api.example.com', got:\n%s", output)
	}
}

func TestFormatVerbose_RequestHeaders_ShownWithPrefix(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://api.example.com/api/users",
		Headers: []domain.Header{
			{Key: "Authorization", Value: "Bearer token123"},
			{Key: "Accept", Value: "application/json"},
		},
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   50 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "> Authorization: Bearer token123") {
		t.Errorf("expected '> Authorization: Bearer token123', got:\n%s", output)
	}
	if !strings.Contains(output, "> Accept: application/json") {
		t.Errorf("expected '> Accept: application/json', got:\n%s", output)
	}
}

func TestFormatVerbose_ResponseStatus_ShownWithPrefix(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "< HTTP/1.1 200 OK") {
		t.Errorf("expected '< HTTP/1.1 200 OK', got:\n%s", output)
	}
}

func TestFormatVerbose_ResponseHeaders_ShownWithPrefix(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Content-Length", Value: "42"},
		},
		Body:     []byte("ok"),
		Duration: 10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "< Content-Type: application/json") {
		t.Errorf("expected '< Content-Type: application/json', got:\n%s", output)
	}
	if !strings.Contains(output, "< Content-Length: 42") {
		t.Errorf("expected '< Content-Length: 42', got:\n%s", output)
	}
}

func TestFormatVerbose_JSONBody_PrettyPrinted(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte(`{"users":[{"id":1}]}`),
		Duration:   10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Pretty-printed JSON should have indentation
	if !strings.Contains(output, "\"users\"") {
		t.Errorf("expected pretty-printed JSON body, got:\n%s", output)
	}
	if !strings.Contains(output, "  ") {
		t.Errorf("expected indentation in JSON body, got:\n%s", output)
	}
}

func TestFormatVerbose_NoColor_NoANSICodes(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "\033[") {
		t.Errorf("expected no ANSI escape codes with noColor=true, got:\n%q", output)
	}
}

func TestFormatVerbose_WithColor_HasANSICodes(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
		Headers: []domain.Header{
			{Key: "Accept", Value: "application/json"},
		},
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte("ok"),
		Duration: 10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Cyan for request lines: \033[36m
	if !strings.Contains(output, "\033[36m") {
		t.Errorf("expected cyan ANSI code for request lines, got:\n%q", output)
	}
	// Yellow for response lines: \033[33m
	if !strings.Contains(output, "\033[33m") {
		t.Errorf("expected yellow ANSI code for response lines, got:\n%q", output)
	}
}

func TestFormatVerbose_EmptyRequestLine(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}
	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should have a blank request separator line "> " followed by response
	if !strings.Contains(output, ">\n") {
		t.Errorf("expected blank request separator line '>', got:\n%q", output)
	}
	// Should have a blank response separator line "< " followed by body
	if !strings.Contains(output, "<\n") {
		t.Errorf("expected blank response separator line '<', got:\n%q", output)
	}
}

func TestFormatVerbose_RequestLineOrder(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodPost,
		URL:    "https://api.example.com/data",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
	}
	resp := domain.HTTPResponse{
		StatusCode: 201,
		Status:     "201 Created",
		Headers:    []domain.Header{},
		Body:       []byte("created"),
		Duration:   10 * time.Millisecond,
	}

	err := FormatVerbose(&buf, req, resp, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Request lines should come before response lines
	reqIdx := strings.Index(output, "> POST")
	respIdx := strings.Index(output, "< HTTP/1.1")
	if reqIdx == -1 {
		t.Fatalf("expected request line not found")
	}
	if respIdx == -1 {
		t.Fatalf("expected response line not found")
	}
	if reqIdx >= respIdx {
		t.Errorf("request lines should appear before response lines")
	}
}

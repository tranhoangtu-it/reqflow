package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestPrettyFormatter_200OK_JSONBody(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"id":1,"name":"John"}`),
		Duration: 123 * time.Millisecond,
		Size:     21,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Status line
	if !strings.Contains(output, "HTTP/1.1 200 OK") {
		t.Errorf("expected status line 'HTTP/1.1 200 OK', got:\n%s", output)
	}

	// Header
	if !strings.Contains(output, "Content-Type: application/json") {
		t.Errorf("expected header 'Content-Type: application/json', got:\n%s", output)
	}

	// Pretty-printed JSON body
	if !strings.Contains(output, `"id": 1`) {
		t.Errorf("expected pretty-printed JSON with '\"id\": 1', got:\n%s", output)
	}
	if !strings.Contains(output, `"name": "John"`) {
		t.Errorf("expected pretty-printed JSON with '\"name\": \"John\"', got:\n%s", output)
	}

	// Duration
	if !strings.Contains(output, "(took 123ms)") {
		t.Errorf("expected duration '(took 123ms)', got:\n%s", output)
	}
}

func TestPrettyFormatter_404Response(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 404,
		Status:     "404 Not Found",
		Headers:    []domain.Header{},
		Body:       []byte("not found"),
		Duration:   50 * time.Millisecond,
		Size:       9,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "HTTP/1.1 404 Not Found") {
		t.Errorf("expected status line 'HTTP/1.1 404 Not Found', got:\n%s", output)
	}
}

func TestPrettyFormatter_MultipleHeaders(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "text/plain"},
			{Key: "X-Request-Id", Value: "abc-123"},
			{Key: "Cache-Control", Value: "no-cache"},
		},
		Body:     []byte("hello"),
		Duration: 10 * time.Millisecond,
		Size:     5,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Content-Type: text/plain") {
		t.Errorf("missing header Content-Type")
	}
	if !strings.Contains(output, "X-Request-Id: abc-123") {
		t.Errorf("missing header X-Request-Id")
	}
	if !strings.Contains(output, "Cache-Control: no-cache") {
		t.Errorf("missing header Cache-Control")
	}
}

func TestPrettyFormatter_NonJSONBody(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("plain text body"),
		Duration:   5 * time.Millisecond,
		Size:       15,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "plain text body") {
		t.Errorf("expected body 'plain text body', got:\n%s", output)
	}
}

func TestPrettyFormatter_EmptyBody(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 204,
		Status:     "204 No Content",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   5 * time.Millisecond,
		Size:       0,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "HTTP/1.1 204 No Content") {
		t.Errorf("expected status line, got:\n%s", output)
	}
	// Should have status line and duration but no body section between them
	// The blank separator line + body content should not be present
	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]
	if !strings.Contains(lastLine, "(took") {
		t.Errorf("expected last line to be duration, got: %s", lastLine)
	}
}

func TestPrettyFormatter_NoColor_NoANSICodes(t *testing.T) {
	f := &PrettyFormatter{noColor: true}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   1 * time.Millisecond,
		Size:       2,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("expected no ANSI codes when noColor=true, got:\n%q", output)
	}
}

func TestPrettyFormatter_WithColor_HasANSICodes(t *testing.T) {
	f := &PrettyFormatter{noColor: false}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   1 * time.Millisecond,
		Size:       2,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\033[") {
		t.Errorf("expected ANSI codes when noColor=false, got:\n%q", output)
	}
}

func TestPrettyFormatter_Color_2xx_Green(t *testing.T) {
	f := &PrettyFormatter{noColor: false}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   1 * time.Millisecond,
	}

	f.FormatResponse(&buf, resp)
	output := buf.String()
	// Green ANSI code
	if !strings.Contains(output, "\033[32m") {
		t.Errorf("expected green ANSI code for 2xx, got:\n%q", output)
	}
}

func TestPrettyFormatter_Color_3xx_Yellow(t *testing.T) {
	f := &PrettyFormatter{noColor: false}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 301,
		Status:     "301 Moved Permanently",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   1 * time.Millisecond,
	}

	f.FormatResponse(&buf, resp)
	output := buf.String()
	if !strings.Contains(output, "\033[33m") {
		t.Errorf("expected yellow ANSI code for 3xx, got:\n%q", output)
	}
}

func TestPrettyFormatter_Color_4xx_Red(t *testing.T) {
	f := &PrettyFormatter{noColor: false}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 404,
		Status:     "404 Not Found",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   1 * time.Millisecond,
	}

	f.FormatResponse(&buf, resp)
	output := buf.String()
	if !strings.Contains(output, "\033[31m") {
		t.Errorf("expected red ANSI code for 4xx, got:\n%q", output)
	}
}

func TestPrettyFormatter_Color_5xx_Red(t *testing.T) {
	f := &PrettyFormatter{noColor: false}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 500,
		Status:     "500 Internal Server Error",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   1 * time.Millisecond,
	}

	f.FormatResponse(&buf, resp)
	output := buf.String()
	if !strings.Contains(output, "\033[31m") {
		t.Errorf("expected red ANSI code for 5xx, got:\n%q", output)
	}
}

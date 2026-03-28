package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestMinimalFormatter_StatusLineAndBody(t *testing.T) {
	f := &MinimalFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"id":1,"name":"John"}`),
		Duration: 100 * time.Millisecond,
		Size:     21,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.SplitN(output, "\n", 3)

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %q", len(lines), output)
	}

	// First line: status
	if lines[0] != "200 OK" {
		t.Errorf("expected first line '200 OK', got %q", lines[0])
	}

	// Second line: body
	if lines[1] != `{"id":1,"name":"John"}` {
		t.Errorf("expected second line to be body, got %q", lines[1])
	}
}

func TestMinimalFormatter_EmptyBody_OnlyStatusLine(t *testing.T) {
	f := &MinimalFormatter{}
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

	output := strings.TrimRight(buf.String(), "\n")
	if output != "204 No Content" {
		t.Errorf("expected only '204 No Content', got %q", output)
	}
}

func TestMinimalFormatter_NoHeaders_InOutput(t *testing.T) {
	f := &MinimalFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "text/plain"},
			{Key: "X-Custom", Value: "value"},
		},
		Body:     []byte("hello"),
		Duration: 1 * time.Millisecond,
		Size:     5,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Content-Type") {
		t.Errorf("minimal output should not contain headers, got:\n%s", output)
	}
	if strings.Contains(output, "X-Custom") {
		t.Errorf("minimal output should not contain headers, got:\n%s", output)
	}
}

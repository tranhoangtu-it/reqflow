package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestFormatDryRun_ShowsDryRunIndicator(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://api.example.com/users",
	}

	err := FormatDryRun(&buf, req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "DRY RUN") {
		t.Errorf("expected 'DRY RUN' indicator, got:\n%s", output)
	}
}

func TestFormatDryRun_ShowsRequestDetails(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodPost,
		URL:    "https://api.example.com/data",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
			{Key: "Authorization", Value: "Bearer token123"},
		},
	}

	err := FormatDryRun(&buf, req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "> POST /data HTTP/1.1") {
		t.Errorf("expected request line '> POST /data HTTP/1.1', got:\n%s", output)
	}
	if !strings.Contains(output, "> Host: api.example.com") {
		t.Errorf("expected host header, got:\n%s", output)
	}
	if !strings.Contains(output, "> Content-Type: application/json") {
		t.Errorf("expected Content-Type header, got:\n%s", output)
	}
	if !strings.Contains(output, "> Authorization: Bearer token123") {
		t.Errorf("expected Authorization header, got:\n%s", output)
	}
}

func TestFormatDryRun_DoesNotShowResponseLines(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}

	err := FormatDryRun(&buf, req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "< ") {
		t.Errorf("dry run should not contain response lines, got:\n%s", output)
	}
}

func TestFormatDryRun_NoColor_NoANSICodes(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
	}

	err := FormatDryRun(&buf, req, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("expected no ANSI codes with noColor=true, got:\n%q", output)
	}
}

func TestFormatDryRun_WithColor_HasANSICodes(t *testing.T) {
	var buf bytes.Buffer

	req := domain.HTTPRequest{
		Method: domain.MethodGet,
		URL:    "https://example.com/",
		Headers: []domain.Header{
			{Key: "Accept", Value: "*/*"},
		},
	}

	err := FormatDryRun(&buf, req, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\033[") {
		t.Errorf("expected ANSI codes with noColor=false, got:\n%q", output)
	}
}

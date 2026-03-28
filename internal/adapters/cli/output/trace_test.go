package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestFormatTrace_ShowsAllTimingFields(t *testing.T) {
	var buf bytes.Buffer

	timing := domain.TimingInfo{
		DNSLookup:    12 * time.Millisecond,
		TCPConnect:   25 * time.Millisecond,
		TLSHandshake: 45 * time.Millisecond,
		FirstByte:    89 * time.Millisecond,
		Total:        142 * time.Millisecond,
	}

	err := FormatTrace(&buf, timing, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "DNS Lookup:") {
		t.Errorf("expected 'DNS Lookup:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "12ms") {
		t.Errorf("expected '12ms' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "TCP Connect:") {
		t.Errorf("expected 'TCP Connect:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "25ms") {
		t.Errorf("expected '25ms' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "TLS Handshake:") {
		t.Errorf("expected 'TLS Handshake:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "45ms") {
		t.Errorf("expected '45ms' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "First Byte:") {
		t.Errorf("expected 'First Byte:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "89ms") {
		t.Errorf("expected '89ms' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Total:") {
		t.Errorf("expected 'Total:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "142ms") {
		t.Errorf("expected '142ms' in output, got:\n%s", output)
	}
}

func TestFormatTrace_ZeroTLSHandshake(t *testing.T) {
	var buf bytes.Buffer

	timing := domain.TimingInfo{
		DNSLookup:    5 * time.Millisecond,
		TCPConnect:   10 * time.Millisecond,
		TLSHandshake: 0, // HTTP, no TLS
		FirstByte:    20 * time.Millisecond,
		Total:        30 * time.Millisecond,
	}

	err := FormatTrace(&buf, timing, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// TLS Handshake should still appear, showing 0ms
	if !strings.Contains(output, "TLS Handshake:") {
		t.Errorf("expected 'TLS Handshake:' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "0ms") {
		t.Errorf("expected '0ms' for TLS Handshake, got:\n%s", output)
	}
}

func TestFormatTrace_NoColor_NoANSICodes(t *testing.T) {
	var buf bytes.Buffer

	timing := domain.TimingInfo{
		Total: 100 * time.Millisecond,
	}

	err := FormatTrace(&buf, timing, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("expected no ANSI codes with noColor=true, got:\n%q", output)
	}
}

package commands_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestIntegration_RequestUsesConfigDefaults(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_output: json
default_timeout: 5s
default_headers:
  User-Agent: "reqflow/0.1"
  Accept: "application/json"
`
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"ok":true}`)}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify default headers from config are applied.
	headerMap := make(map[string]string)
	for _, h := range capturedReq.Headers {
		headerMap[h.Key] = h.Value
	}

	if headerMap["User-Agent"] != "reqflow/0.1" {
		t.Errorf("expected default User-Agent header 'reqflow/0.1', got %q", headerMap["User-Agent"])
	}
	if headerMap["Accept"] != "application/json" {
		t.Errorf("expected default Accept header 'application/json', got %q", headerMap["Accept"])
	}
}

func TestIntegration_RequestWithFlagsOverridesConfig(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `default_timeout: 5s
default_headers:
  Accept: "application/json"
`
	if err := os.WriteFile(filepath.Join(globalDir, "config.yaml"), []byte(globalContent), 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	var capturedReq domain.HTTPRequest
	var capturedCtx context.Context
	mock := &mockHTTPClient{
		doFunc: func(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			capturedCtx = ctx
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a,
		commands.WithGlobalConfigDir(globalDir),
		commands.WithProjectConfigDir(projectDir),
	)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api",
		"-H", "Accept: text/html",
		"-t", "15s",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CLI flag header should override config default.
	headerMap := make(map[string]string)
	for _, h := range capturedReq.Headers {
		headerMap[h.Key] = h.Value
	}

	if headerMap["Accept"] != "text/html" {
		t.Errorf("expected CLI flag Accept 'text/html' to override config, got %q", headerMap["Accept"])
	}

	// Timeout flag should override config.
	deadline, ok := capturedCtx.Deadline()
	if !ok {
		t.Fatal("expected context to have a deadline")
	}
	remaining := time.Until(deadline)
	// The timeout was set to 15s via flag. Allow some slack for test execution.
	if remaining < 14*time.Second || remaining > 16*time.Second {
		t.Errorf("expected timeout ~15s from flag override, got %v remaining", remaining)
	}
}

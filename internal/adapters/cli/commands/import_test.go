package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestImportCurlCommand_ExecutesRequest(t *testing.T) {
	var capturedReq domain.HTTPRequest
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			capturedReq = req
			return domain.HTTPResponse{StatusCode: 200, Status: "200 OK", Body: []byte(`{"ok":true}`)}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"import", "curl", `curl -X POST -H "Content-Type: application/json" -d '{"key":"val"}' https://example.com/api`})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Method != domain.MethodPost {
		t.Errorf("Method = %q, want POST", capturedReq.Method)
	}
	if capturedReq.URL != "https://example.com/api" {
		t.Errorf("URL = %q, want https://example.com/api", capturedReq.URL)
	}
}

func TestImportCurlCommand_DryRun(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			t.Fatal("request should not be executed in dry-run mode")
			return domain.HTTPResponse{}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"import", "curl", "--dry-run", `curl -X POST -d '{"key":"val"}' https://example.com/api`})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "POST") {
		t.Errorf("dry-run output should contain method POST, got: %s", output)
	}
	if !strings.Contains(output, "https://example.com/api") {
		t.Errorf("dry-run output should contain URL, got: %s", output)
	}
}

func TestImportCurlCommand_InvalidInput(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"import", "curl", ""})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for empty curl command, got nil")
	}
}

func TestImportCurlCommand_MissingArg(t *testing.T) {
	a := newTestApp(&mockHTTPClient{
		doFunc: func(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{}, nil
		},
	})
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"import", "curl"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing curl argument, got nil")
	}
}

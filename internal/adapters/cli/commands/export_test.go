package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestExportCurlCommand_PrintsCurl(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Status: "200 OK", Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--curl"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "curl") {
		t.Errorf("expected output to contain 'curl', got: %s", output)
	}
	if !strings.Contains(output, "https://example.com/api") {
		t.Errorf("expected output to contain URL, got: %s", output)
	}
}

func TestExportCurlCommand_PostWithData(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			t.Fatal("request should not be executed with --curl flag")
			return domain.HTTPResponse{}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"post", "https://example.com/api", "-d", `{"key":"val"}`, "--curl"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "curl") {
		t.Errorf("expected output to contain 'curl', got: %s", output)
	}
	if !strings.Contains(output, "-X POST") {
		t.Errorf("expected output to contain '-X POST', got: %s", output)
	}
	if !strings.Contains(output, "-d") {
		t.Errorf("expected output to contain '-d', got: %s", output)
	}
}

func TestExportCurlSubcommand_Registered(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	// Verify "export" command exists.
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "export" {
			found = true
			// Verify "curl" subcommand exists under "export".
			hasCurl := false
			for _, sub := range cmd.Commands() {
				if sub.Name() == "curl" {
					hasCurl = true
				}
			}
			if !hasCurl {
				t.Error("export command missing 'curl' subcommand")
			}
		}
	}
	if !found {
		t.Error("root command missing 'export' subcommand")
	}
}

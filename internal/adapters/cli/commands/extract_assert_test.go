package commands_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestExtractFlag_PrintsExtractedValue(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"url":"https://httpbin.org/get","headers":{"Host":"httpbin.org"}}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://httpbin.org/get", "--extract", "$.url"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "https://httpbin.org/get" {
		t.Errorf("output = %q, want %q", output, "https://httpbin.org/get")
	}
}

func TestExtractFlag_WithLabel(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"url":"https://httpbin.org/get","headers":{"Host":"httpbin.org"}}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://httpbin.org/get", "--extract", "host=$.headers.Host"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "host=httpbin.org" {
		t.Errorf("output = %q, want %q", output, "host=httpbin.org")
	}
}

func TestExtractFlag_JSONOutput(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"url":"https://httpbin.org/get","id":42}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://httpbin.org/get", "--extract", "url=$.url", "--extract", "id=$.id", "-o", "json"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if result["url"] != "https://httpbin.org/get" {
		t.Errorf("url = %q, want %q", result["url"], "https://httpbin.org/get")
	}
	if result["id"] != "42" {
		t.Errorf("id = %q, want %q", result["id"], "42")
	}
}

func TestAssertFlag_Passes(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"name":"John"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--assert", "status == 200"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status == 200") {
		t.Errorf("expected assertion result in output, got:\n%s", output)
	}
}

func TestAssertFlag_Fails_ExitCode4(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--assert", "status == 404", "--no-fail-on-error"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for failed assertion, got nil")
	}

	var exitErr *domain.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *domain.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != domain.ExitAssertionFailed {
		t.Errorf("Code = %d, want %d", exitErr.Code, domain.ExitAssertionFailed)
	}
}

func TestAssertFlag_MultipleAsserts_AllPass(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"name":"John"}`),
				Headers: []domain.Header{
					{Key: "Content-Type", Value: "application/json"},
				},
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api",
		"--assert", "status == 200",
		"--assert", "body.name == John",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestAssertFlag_MultipleAsserts_OneFails(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"name":"John"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api",
		"--assert", "status == 200",
		"--assert", "status == 404",
		"--no-fail-on-error",
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for failed assertion, got nil")
	}

	var exitErr *domain.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *domain.ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != domain.ExitAssertionFailed {
		t.Errorf("Code = %d, want %d", exitErr.Code, domain.ExitAssertionFailed)
	}
}

func TestExtractAndAssert_Together(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"id":42,"name":"John"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api",
		"--extract", "id=$.id",
		"--assert", "status == 200",
	})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	// Extract output should come before assert output
	if !strings.Contains(output, "id=42") {
		t.Errorf("expected extract output 'id=42', got:\n%s", output)
	}
	if !strings.Contains(output, "status == 200") {
		t.Errorf("expected assertion result in output, got:\n%s", output)
	}
}

func TestExtractFlag_SkipsNormalResponseFormatting(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{
				StatusCode: 200,
				Body:       []byte(`{"id":42,"name":"John"}`),
			}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--extract", "$.id"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	// Should only print the extracted value, not the full response body
	if output != "42" {
		t.Errorf("output = %q, want %q", output, "42")
	}
}

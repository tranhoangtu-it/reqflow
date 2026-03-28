package commands_test

import (
	"context"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/domain"

	"bytes"
)

func TestRetryFlag_SetsRetryCount(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--retry", "3"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryBackoffFlag_AcceptsExponential(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--retry", "3", "--retry-backoff", "exponential"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOnFlag_ParsesCommaSeparatedStatusCodes(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--retry", "3", "--retry-on", "502,503"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryFlags_InvalidRetryOn_ReturnsError(t *testing.T) {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--retry", "3", "--retry-on", "abc,503"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid status code, got nil")
	}
}

func TestRetryFlag_UsedInRetryLogic(t *testing.T) {
	callCount := 0
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
			callCount++
			if callCount < 3 {
				return domain.HTTPResponse{StatusCode: 503, Body: []byte("unavailable")}, nil
			}
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}

	a := newTestApp(mock)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"get", "https://example.com/api", "--retry", "5", "--retry-on", "503", "--no-fail-on-error"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("HTTP calls = %d, want 3 (2 retries + 1 success)", callCount)
	}
}

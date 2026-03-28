package runner_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

func TestWebhookListener_CapturesBody(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0, // ephemeral port
		Path:    "/webhook",
		Timeout: 5 * time.Second,
		Capture: "callback_body",
	}

	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	// Get the actual port assigned
	port := listener.Port()

	// Send a request to the webhook
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/webhook", port),
		"application/json",
		bytes.NewBufferString(`{"status":"completed","id":42}`),
	)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	resp.Body.Close()

	// Wait for the result
	result := <-resultCh
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if string(result.Body) != `{"status":"completed","id":42}` {
		t.Errorf("Body = %q, want %q", string(result.Body), `{"status":"completed","id":42}`)
	}
}

func TestWebhookListener_Timeout(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0,
		Path:    "/webhook",
		Timeout: 100 * time.Millisecond,
		Capture: "callback_body",
	}

	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	// Don't send any request - should timeout
	result := <-resultCh
	if result.Error == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWebhookListener_CorrectPath(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0,
		Path:    "/callback",
		Timeout: 5 * time.Second,
		Capture: "result",
	}

	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	port := listener.Port()

	// Send to the correct path
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/callback", port),
		"text/plain",
		bytes.NewBufferString("hello"),
	)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	resp.Body.Close()

	result := <-resultCh
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if string(result.Body) != "hello" {
		t.Errorf("Body = %q, want %q", string(result.Body), "hello")
	}
}

func TestWebhookListener_IgnoresWrongPath(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0,
		Path:    "/webhook",
		Timeout: 500 * time.Millisecond,
		Capture: "callback_body",
	}

	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	port := listener.Port()

	// Send to the wrong path - should be ignored
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/wrong", port),
		"application/json",
		bytes.NewBufferString(`{"data":"ignored"}`),
	)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong path status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	resp.Body.Close()

	// Should timeout since the correct path was never hit
	result := <-resultCh
	if result.Error == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestWebhookListener_ContextCancellation(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0,
		Path:    "/webhook",
		Timeout: 30 * time.Second,
		Capture: "callback_body",
	}

	ctx, cancel := context.WithCancel(context.Background())
	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	// Cancel the context
	cancel()

	// Should receive an error result
	result := <-resultCh
	if result.Error == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

func TestWebhookListener_CapturesHeaders(t *testing.T) {
	cfg := domain.ListenConfig{
		Port:    0,
		Path:    "/webhook",
		Timeout: 5 * time.Second,
		Capture: "callback_body",
	}

	listener := runner.NewWebhookListener(cfg)
	resultCh, err := listener.Start(context.Background())
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer listener.Stop()

	port := listener.Port()

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/webhook", port), bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatalf("NewRequest error: %v", err)
	}
	req.Header.Set("X-Custom-Header", "test-value")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	resp.Body.Close()

	result := <-resultCh
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Headers["X-Custom-Header"] != "test-value" {
		t.Errorf("Headers[X-Custom-Header] = %q, want %q", result.Headers["X-Custom-Header"], "test-value")
	}
}

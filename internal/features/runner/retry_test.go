package runner_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

func TestRetry_SuccessOnFirstTry_NoRetry(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"ok":true}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "retry-test",
		Steps: []domain.Step{
			{
				Name:   "first try success",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/ok",
				Retry: &domain.RetryConfig{
					Max:          3,
					Backoff:      "fixed",
					InitialDelay: 10 * time.Millisecond,
					RetryOn:      []int{502, 503, 504},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].Error != nil {
		t.Errorf("unexpected step error: %v", result.Steps[0].Error)
	}
	if client.calls != 1 {
		t.Errorf("HTTP calls = %d, want 1 (no retry needed)", client.calls)
	}
}

func TestRetry_503Then200_RetriesOnceSuccessfully(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
			{StatusCode: 200, Body: []byte(`{"ok":true}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "retry-test",
		Steps: []domain.Step{
			{
				Name:   "retry once",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/flaky",
				Retry: &domain.RetryConfig{
					Max:          3,
					Backoff:      "fixed",
					InitialDelay: 10 * time.Millisecond,
					RetryOn:      []int{502, 503, 504},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps[0].Error != nil {
		t.Errorf("unexpected step error: %v", result.Steps[0].Error)
	}
	if result.Steps[0].Response.StatusCode != 200 {
		t.Errorf("status = %d, want 200", result.Steps[0].Response.StatusCode)
	}
	if client.calls != 2 {
		t.Errorf("HTTP calls = %d, want 2 (1 initial + 1 retry)", client.calls)
	}
}

func TestRetry_503ThreeTimes_MaxExceeded_ReturnsError(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "retry-test",
		Steps: []domain.Step{
			{
				Name:   "always fails",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/down",
				Retry: &domain.RetryConfig{
					Max:          2,
					Backoff:      "fixed",
					InitialDelay: 10 * time.Millisecond,
					RetryOn:      []int{503},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps[0].Error == nil {
		t.Fatal("expected step error after max retries exceeded")
	}
	// 1 initial + 2 retries = 3 calls
	if client.calls != 3 {
		t.Errorf("HTTP calls = %d, want 3 (1 initial + 2 retries)", client.calls)
	}
}

func TestRetry_NetworkError_WithRetryOnError_Retries(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{},
			{StatusCode: 200, Body: []byte(`{"ok":true}`)},
		},
		errors: []error{
			fmt.Errorf("connection refused"),
			nil,
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "retry-test",
		Steps: []domain.Step{
			{
				Name:   "network error then success",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/flaky",
				Retry: &domain.RetryConfig{
					Max:          3,
					Backoff:      "fixed",
					InitialDelay: 10 * time.Millisecond,
					RetryOnError: true,
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps[0].Error != nil {
		t.Errorf("unexpected step error: %v", result.Steps[0].Error)
	}
	if result.Steps[0].Response.StatusCode != 200 {
		t.Errorf("status = %d, want 200", result.Steps[0].Response.StatusCode)
	}
	if client.calls != 2 {
		t.Errorf("HTTP calls = %d, want 2", client.calls)
	}
}

func TestRetry_NonRetryableStatus_NoRetry(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 404, Body: []byte(`{"error":"not found"}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "retry-test",
		Steps: []domain.Step{
			{
				Name:   "not retryable",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/missing",
				Retry: &domain.RetryConfig{
					Max:          3,
					Backoff:      "fixed",
					InitialDelay: 10 * time.Millisecond,
					RetryOn:      []int{502, 503, 504},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps[0].Response.StatusCode != 404 {
		t.Errorf("status = %d, want 404", result.Steps[0].Response.StatusCode)
	}
	if client.calls != 1 {
		t.Errorf("HTTP calls = %d, want 1 (no retry for 404)", client.calls)
	}
}

func TestRetry_NoRetryConfig_ExecutesNormally(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "no-retry-test",
		Steps: []domain.Step{
			{
				Name:   "no retry configured",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/down",
				// No Retry config
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Steps[0].Response.StatusCode != 503 {
		t.Errorf("status = %d, want 503", result.Steps[0].Response.StatusCode)
	}
	if client.calls != 1 {
		t.Errorf("HTTP calls = %d, want 1", client.calls)
	}
}

func TestRetry_ContextCancelled_StopsRetrying(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
			{StatusCode: 503, Body: []byte(`{"error":"unavailable"}`)},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "ctx-cancel-test",
		Steps: []domain.Step{
			{
				Name:   "cancelled",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/down",
				Retry: &domain.RetryConfig{
					Max:          5,
					Backoff:      "fixed",
					InitialDelay: 1 * time.Second,
					RetryOn:      []int{503},
				},
			},
		},
	}

	result, err := r.Run(ctx, wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have stopped early due to context cancellation
	if result.Steps[0].Error == nil {
		t.Error("expected error from cancelled context")
	}
}

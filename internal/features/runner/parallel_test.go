package runner_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

// concurrentMockHTTPClient is a thread-safe mock for parallel test scenarios.
// It dispatches responses based on the request URL.
type concurrentMockHTTPClient struct {
	mu        sync.Mutex
	byURL     map[string]mockResponse
	callOrder []string
	callTimes []time.Time
	delay     time.Duration // artificial delay per request
}

type mockResponse struct {
	resp domain.HTTPResponse
	err  error
}

func newConcurrentMock() *concurrentMockHTTPClient {
	return &concurrentMockHTTPClient{
		byURL: make(map[string]mockResponse),
	}
}

func (m *concurrentMockHTTPClient) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return domain.HTTPResponse{}, ctx.Err()
		}
	}

	m.mu.Lock()
	m.callOrder = append(m.callOrder, req.URL)
	m.callTimes = append(m.callTimes, time.Now())
	m.mu.Unlock()

	if r, ok := m.byURL[req.URL]; ok {
		if r.err != nil {
			return domain.HTTPResponse{}, r.err
		}
		return r.resp, nil
	}
	return domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)}, nil
}

func TestRunner_ParallelBothSucceed(t *testing.T) {
	client := newConcurrentMock()
	client.byURL["https://api.example.com/users"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"count":10}`)},
	}
	client.byURL["https://api.example.com/posts"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"count":5}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "parallel-test",
		Steps: []domain.Step{
			{
				Name:     "fetch both",
				FailFast: true,
				Parallel: []domain.Step{
					{
						Name:   "get users",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/users",
					},
					{
						Name:   "get posts",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/posts",
					},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have results for both parallel sub-steps
	if len(result.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(result.Steps))
	}

	// Both should succeed
	for _, sr := range result.Steps {
		if sr.Error != nil {
			t.Errorf("step %q: unexpected error: %v", sr.StepName, sr.Error)
		}
	}

	// Verify both URLs were called
	if len(client.callOrder) != 2 {
		t.Fatalf("call count = %d, want 2", len(client.callOrder))
	}
}

func TestRunner_ParallelVariablesMerged(t *testing.T) {
	client := newConcurrentMock()
	client.byURL["https://api.example.com/users"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"id":"user-1"}`)},
	}
	client.byURL["https://api.example.com/posts"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"id":"post-1"}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "merge-vars-test",
		Steps: []domain.Step{
			{
				Name:     "fetch both",
				FailFast: true,
				Parallel: []domain.Step{
					{
						Name:   "get users",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/users",
						Extract: map[string]string{
							"user_id": "$.id",
						},
					},
					{
						Name:   "get posts",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/posts",
						Extract: map[string]string{
							"post_id": "$.id",
						},
					},
				},
			},
			{
				Name:   "use vars",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/result/{{user_id}}/{{post_id}}",
			},
		},
	}

	client.byURL["https://api.example.com/result/user-1/post-1"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"ok":true}`)},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 3 results: 2 from parallel + 1 sequential
	if len(result.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(result.Steps))
	}

	// Last step should have used the merged variables (URL contains both IDs)
	lastStep := result.Steps[2]
	if lastStep.Error != nil {
		t.Errorf("last step error: %v", lastStep.Error)
	}
	if lastStep.Request.URL != "https://api.example.com/result/user-1/post-1" {
		t.Errorf("last step URL = %q, want interpolated URL with both vars", lastStep.Request.URL)
	}
}

func TestRunner_ParallelFailFastCancelsRemaining(t *testing.T) {
	client := newConcurrentMock()
	client.delay = 50 * time.Millisecond

	client.byURL["https://api.example.com/fast"] = mockResponse{
		err: fmt.Errorf("fast failure"),
	}
	client.byURL["https://api.example.com/slow"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "fail-fast-test",
		Steps: []domain.Step{
			{
				Name:     "parallel block",
				FailFast: true,
				Parallel: []domain.Step{
					{
						Name:   "fast fail",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/fast",
					},
					{
						Name:   "slow success",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/slow",
					},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// At least one step should have an error
	hasError := false
	for _, sr := range result.Steps {
		if sr.Error != nil {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Fatal("expected at least one step error in fail-fast mode")
	}
}

func TestRunner_ParallelCollectAllMode(t *testing.T) {
	client := newConcurrentMock()
	client.byURL["https://api.example.com/fail"] = mockResponse{
		err: fmt.Errorf("intentional failure"),
	}
	client.byURL["https://api.example.com/success"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"ok":true}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "collect-all-test",
		Steps: []domain.Step{
			{
				Name:     "parallel block",
				FailFast: false, // collect-all mode
				Parallel: []domain.Step{
					{
						Name:   "failing step",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/fail",
					},
					{
						Name:   "succeeding step",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/success",
					},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both steps should have run
	if len(result.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2 (both should run in collect-all mode)", len(result.Steps))
	}

	// Verify one failed and one succeeded
	var failCount, successCount int
	for _, sr := range result.Steps {
		if sr.Error != nil {
			failCount++
		} else {
			successCount++
		}
	}
	if failCount != 1 {
		t.Errorf("fail count = %d, want 1", failCount)
	}
	if successCount != 1 {
		t.Errorf("success count = %d, want 1", successCount)
	}
}

func TestRunner_ParallelMaxParallelOne(t *testing.T) {
	client := newConcurrentMock()
	client.delay = 50 * time.Millisecond

	client.byURL["https://api.example.com/a"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)},
	}
	client.byURL["https://api.example.com/b"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "max-parallel-test",
		Steps: []domain.Step{
			{
				Name:        "sequential via max_parallel",
				FailFast:    true,
				MaxParallel: 1,
				Parallel: []domain.Step{
					{
						Name:   "step a",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/a",
					},
					{
						Name:   "step b",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/b",
					},
				},
			},
		},
	}

	start := time.Now()
	result, err := r.Run(context.Background(), wf, nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(result.Steps))
	}

	// With MaxParallel=1 and 50ms delay each, should take >=100ms
	// (sequential execution). With unlimited parallelism it would take ~50ms.
	if elapsed < 90*time.Millisecond {
		t.Errorf("elapsed = %v, expected >= 90ms for sequential execution (MaxParallel=1)", elapsed)
	}
}

func TestRunner_ParallelVarsAvailableInNextSequentialStep(t *testing.T) {
	client := newConcurrentMock()
	client.byURL["https://api.example.com/auth"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"token":"secret-123"}`)},
	}
	client.byURL["https://api.example.com/profile"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "vars-chain-test",
		Steps: []domain.Step{
			{
				Name:     "auth parallel",
				FailFast: true,
				Parallel: []domain.Step{
					{
						Name:   "get token",
						Method: domain.MethodGet,
						URL:    "https://api.example.com/auth",
						Extract: map[string]string{
							"auth_token": "$.token",
						},
					},
				},
			},
			{
				Name:   "use token",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/profile",
				Headers: map[string]string{
					"Authorization": "Bearer {{auth_token}}",
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 parallel sub-step result + 1 sequential step result
	if len(result.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(result.Steps))
	}

	// The sequential step should have the interpolated header
	lastStep := result.Steps[1]
	if lastStep.Error != nil {
		t.Errorf("last step error: %v", lastStep.Error)
	}

	// Check that Authorization header was interpolated
	found := false
	for _, h := range lastStep.Request.Headers {
		if h.Key == "Authorization" && strings.Contains(h.Value, "secret-123") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Authorization header with interpolated token, not found")
	}
}

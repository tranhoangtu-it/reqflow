package runner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

// mockHTTPClient implements driven.HTTPClient for testing.
type mockHTTPClient struct {
	responses []domain.HTTPResponse
	errors    []error
	calls     int
}

func (m *mockHTTPClient) Do(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.errors) && m.errors[idx] != nil {
		return domain.HTTPResponse{}, m.errors[idx]
	}
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)}, nil
}

func TestRunner_SingleStep(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{
				StatusCode: 200,
				Body:       []byte(`{"message":"ok"}`),
				Headers:    []domain.Header{{Key: "Content-Type", Value: "application/json"}},
			},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "test",
		Steps: []domain.Step{
			{
				Name:   "get users",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/users",
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" {
		t.Errorf("name = %q, want %q", result.Name, "test")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].StepName != "get users" {
		t.Errorf("step name = %q, want %q", result.Steps[0].StepName, "get users")
	}
	if result.Steps[0].Error != nil {
		t.Errorf("unexpected step error: %v", result.Steps[0].Error)
	}
}

func TestRunner_MultiStepChainsVariables(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{
				StatusCode: 201,
				Body:       []byte(`{"id":"user-123","token":"abc"}`),
			},
			{
				StatusCode: 200,
				Body:       []byte(`{"profile":"complete"}`),
			},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "chain-test",
		Steps: []domain.Step{
			{
				Name:   "create user",
				Method: domain.MethodPost,
				URL:    "https://api.example.com/users",
				Body:   `{"name":"John"}`,
				Extract: map[string]string{
					"user_id":    "$.id",
					"auth_token": "$.token",
				},
			},
			{
				Name:   "get profile",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/users/{{user_id}}",
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
	if len(result.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(result.Steps))
	}

	// First step should have extracted variables
	step0 := result.Steps[0]
	if step0.Extracted["user_id"] != "user-123" {
		t.Errorf("extracted user_id = %q, want %q", step0.Extracted["user_id"], "user-123")
	}
	if step0.Extracted["auth_token"] != "abc" {
		t.Errorf("extracted auth_token = %q, want %q", step0.Extracted["auth_token"], "abc")
	}

	// Second step should have no error
	if result.Steps[1].Error != nil {
		t.Errorf("step 1 error: %v", result.Steps[1].Error)
	}
}

func TestRunner_AssertionResults(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{
				StatusCode: 200,
				Body:       []byte(`{"status":"ok"}`),
			},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "assert-test",
		Steps: []domain.Step{
			{
				Name:   "check status",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/health",
				Assert: []domain.Assertion{
					{Field: "status", Operator: "==", Expected: 200},
					{Field: "body.status", Operator: "==", Expected: "ok"},
				},
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPassed != 2 {
		t.Errorf("total passed = %d, want 2", result.TotalPassed)
	}
	if result.TotalFailed != 0 {
		t.Errorf("total failed = %d, want 0", result.TotalFailed)
	}
}

func TestRunner_FailedAssertionStopsExecution(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{
				StatusCode: 404,
				Body:       []byte(`{"error":"not found"}`),
			},
			{
				StatusCode: 200,
				Body:       []byte(`{}`),
			},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "fail-test",
		Steps: []domain.Step{
			{
				Name:   "failing step",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/missing",
				Assert: []domain.Assertion{
					{Field: "status", Operator: "==", Expected: 200},
				},
			},
			{
				Name:   "should not run",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/other",
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the first step should have executed
	if len(result.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1 (should stop on failure)", len(result.Steps))
	}
	if result.TotalFailed != 1 {
		t.Errorf("total failed = %d, want 1", result.TotalFailed)
	}
}

func TestRunner_HTTPErrorCaptured(t *testing.T) {
	client := &mockHTTPClient{
		errors: []error{fmt.Errorf("connection refused")},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "error-test",
		Steps: []domain.Step{
			{
				Name:   "failing request",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/down",
			},
		},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].Error == nil {
		t.Fatal("expected step error, got nil")
	}
}

func TestRunner_InitialVarsUsed(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "vars-test",
		Steps: []domain.Step{
			{
				Name:   "use vars",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/users/{{user_id}}",
			},
		},
	}

	initialVars := map[string]string{"user_id": "42"}
	result, err := r.Run(context.Background(), wf, initialVars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("steps count = %d, want 1", len(result.Steps))
	}
	if result.Steps[0].Error != nil {
		t.Errorf("unexpected error: %v", result.Steps[0].Error)
	}
}

func TestRunner_BodyAsMapMarshaledToJSON(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 201, Body: []byte(`{"id":1}`)},
		},
	}

	r := runner.New(client)
	wf := domain.Workflow{
		Name: "map-body-test",
		Steps: []domain.Step{
			{
				Name:   "post with map body",
				Method: domain.MethodPost,
				URL:    "https://api.example.com/users",
				Body:   map[string]interface{}{"name": "John", "age": 30},
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
}

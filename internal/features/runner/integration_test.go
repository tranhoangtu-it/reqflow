package runner_test

import (
	"context"
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

func TestIntegration_ParallelThenSequentialWithVars(t *testing.T) {
	yamlData := []byte(`
name: integration-parallel
steps:
  - name: fetch resources
    parallel:
      - name: get user
        method: GET
        url: https://api.example.com/users/1
        extract:
          user_name: $.name
      - name: get config
        method: GET
        url: https://api.example.com/config
        extract:
          api_version: $.version
  - name: create report
    method: POST
    url: https://api.example.com/reports
    body: '{"user":"{{user_name}}","version":"{{api_version}}"}'
`)

	wf, err := workflow.Parse(yamlData)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Verify parsed structure
	if len(wf.Steps) != 2 {
		t.Fatalf("parsed steps = %d, want 2", len(wf.Steps))
	}
	if len(wf.Steps[0].Parallel) != 2 {
		t.Fatalf("parallel sub-steps = %d, want 2", len(wf.Steps[0].Parallel))
	}

	// Default FailFast should be true
	if !wf.Steps[0].FailFast {
		t.Error("expected FailFast to default to true")
	}

	// Set up mock HTTP client
	client := newConcurrentMock()
	client.byURL["https://api.example.com/users/1"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"name":"Alice"}`)},
	}
	client.byURL["https://api.example.com/config"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 200, Body: []byte(`{"version":"v2"}`)},
	}
	client.byURL["https://api.example.com/reports"] = mockResponse{
		resp: domain.HTTPResponse{StatusCode: 201, Body: []byte(`{"id":"report-1"}`)},
	}

	r := runner.New(client)
	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("run error: %v", err)
	}

	// 2 parallel results + 1 sequential result = 3
	if len(result.Steps) != 3 {
		t.Fatalf("result steps = %d, want 3", len(result.Steps))
	}

	// All steps should succeed
	for _, sr := range result.Steps {
		if sr.Error != nil {
			t.Errorf("step %q failed: %v", sr.StepName, sr.Error)
		}
	}

	// The sequential step should have received interpolated body
	reportStep := result.Steps[2]
	if reportStep.StepName != "create report" {
		t.Errorf("last step name = %q, want %q", reportStep.StepName, "create report")
	}

	// Verify the request body was interpolated with parallel-extracted vars
	bodyStr := string(reportStep.Request.Body)
	if bodyStr != `{"user":"Alice","version":"v2"}` {
		t.Errorf("body = %q, want interpolated JSON with vars from parallel steps", bodyStr)
	}
}

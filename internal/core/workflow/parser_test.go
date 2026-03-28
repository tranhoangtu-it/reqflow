package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestParse_ValidWorkflow(t *testing.T) {
	yaml := []byte(`
name: test-workflow
env: dev
steps:
  - name: get users
    method: GET
    url: https://api.example.com/users
    headers:
      Accept: application/json
  - name: create user
    method: POST
    url: https://api.example.com/users
    content_type: application/json
    body: '{"name": "John"}'
    extract:
      user_id: $.id
    assert:
      - field: status
        operator: "=="
        expected: 201
`)

	wf, err := workflow.Parse(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Name != "test-workflow" {
		t.Errorf("name = %q, want %q", wf.Name, "test-workflow")
	}
	if wf.Env != "dev" {
		t.Errorf("env = %q, want %q", wf.Env, "dev")
	}
	if len(wf.Steps) != 2 {
		t.Fatalf("steps count = %d, want 2", len(wf.Steps))
	}

	step0 := wf.Steps[0]
	if step0.Name != "get users" {
		t.Errorf("step[0].Name = %q, want %q", step0.Name, "get users")
	}
	if step0.Method != domain.MethodGet {
		t.Errorf("step[0].Method = %q, want %q", step0.Method, domain.MethodGet)
	}
	if step0.URL != "https://api.example.com/users" {
		t.Errorf("step[0].URL = %q, want %q", step0.URL, "https://api.example.com/users")
	}
	if step0.Headers["Accept"] != "application/json" {
		t.Errorf("step[0].Headers[Accept] = %q, want %q", step0.Headers["Accept"], "application/json")
	}

	step1 := wf.Steps[1]
	if step1.Name != "create user" {
		t.Errorf("step[1].Name = %q, want %q", step1.Name, "create user")
	}
	if step1.ContentType != "application/json" {
		t.Errorf("step[1].ContentType = %q, want %q", step1.ContentType, "application/json")
	}
	if step1.Extract["user_id"] != "$.id" {
		t.Errorf("step[1].Extract[user_id] = %q, want %q", step1.Extract["user_id"], "$.id")
	}
	if len(step1.Assert) != 1 {
		t.Fatalf("step[1].Assert count = %d, want 1", len(step1.Assert))
	}
	if step1.Assert[0].Field != "status" {
		t.Errorf("step[1].Assert[0].Field = %q, want %q", step1.Assert[0].Field, "status")
	}
	if step1.Assert[0].Operator != "==" {
		t.Errorf("step[1].Assert[0].Operator = %q, want %q", step1.Assert[0].Operator, "==")
	}
}

func TestParse_MissingName(t *testing.T) {
	yaml := []byte(`
steps:
  - name: step1
    method: GET
    url: https://example.com
`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestParse_NoSteps(t *testing.T) {
	yaml := []byte(`
name: empty-workflow
steps: []
`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for empty steps, got nil")
	}
}

func TestParse_StepWithoutURL(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - name: bad step
    method: GET
`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for step without URL, got nil")
	}
}

func TestParse_StepWithoutMethod(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - name: bad step
    url: https://example.com
`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for step without method, got nil")
	}
}

func TestParse_StepWithoutName(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - method: GET
    url: https://example.com
`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for step without name, got nil")
	}
}

func TestParse_BodyAsString(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - name: string body
    method: POST
    url: https://example.com
    body: "hello world"
`)

	wf, err := workflow.Parse(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, ok := wf.Steps[0].Body.(string)
	if !ok {
		t.Fatalf("body type = %T, want string", wf.Steps[0].Body)
	}
	if body != "hello world" {
		t.Errorf("body = %q, want %q", body, "hello world")
	}
}

func TestParse_BodyAsMap(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - name: map body
    method: POST
    url: https://example.com
    body:
      name: John
      age: 30
`)

	wf, err := workflow.Parse(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// When body is a map, it should remain as a map (marshaled to JSON later by the runner).
	bodyMap, ok := wf.Steps[0].Body.(map[string]interface{})
	if !ok {
		t.Fatalf("body type = %T, want map[string]interface{}", wf.Steps[0].Body)
	}
	if bodyMap["name"] != "John" {
		t.Errorf("body[name] = %v, want %q", bodyMap["name"], "John")
	}
}

func TestParse_ExtractAndAssertFields(t *testing.T) {
	yaml := []byte(`
name: test
steps:
  - name: step1
    method: GET
    url: https://example.com
    extract:
      token: $.data.token
      id: $.data.id
    assert:
      - field: status
        operator: "=="
        expected: 200
      - field: body.data.active
        operator: "=="
        expected: true
      - field: header.Content-Type
        operator: contains
        expected: json
`)

	wf, err := workflow.Parse(yaml)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	step := wf.Steps[0]
	if len(step.Extract) != 2 {
		t.Fatalf("extract count = %d, want 2", len(step.Extract))
	}
	if step.Extract["token"] != "$.data.token" {
		t.Errorf("extract[token] = %q, want %q", step.Extract["token"], "$.data.token")
	}

	if len(step.Assert) != 3 {
		t.Fatalf("assert count = %d, want 3", len(step.Assert))
	}
	if step.Assert[1].Field != "body.data.active" {
		t.Errorf("assert[1].Field = %q, want %q", step.Assert[1].Field, "body.data.active")
	}
	if step.Assert[2].Operator != "contains" {
		t.Errorf("assert[2].Operator = %q, want %q", step.Assert[2].Operator, "contains")
	}
}

func TestParse_InvalidYAML(t *testing.T) {
	yaml := []byte(`{{{invalid yaml`)

	_, err := workflow.Parse(yaml)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

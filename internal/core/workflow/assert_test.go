package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
)

func makeResponse(status int, body string, headers ...domain.Header) domain.HTTPResponse {
	return domain.HTTPResponse{
		StatusCode: status,
		Body:       []byte(body),
		Headers:    headers,
	}
}

func TestEvaluateAssertions_StatusEquals_Pass(t *testing.T) {
	resp := makeResponse(200, `{}`)
	assertions := []domain.Assertion{
		{Field: "status", Operator: "==", Expected: 200},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	if !results[0].Passed {
		t.Errorf("expected assertion to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_StatusEquals_Fail(t *testing.T) {
	resp := makeResponse(404, `{}`)
	assertions := []domain.Assertion{
		{Field: "status", Operator: "==", Expected: 200},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	if results[0].Passed {
		t.Error("expected assertion to fail")
	}
	if results[0].Actual != 404 {
		t.Errorf("actual = %v, want 404", results[0].Actual)
	}
}

func TestEvaluateAssertions_StatusNotEquals(t *testing.T) {
	resp := makeResponse(200, `{}`)
	assertions := []domain.Assertion{
		{Field: "status", Operator: "!=", Expected: 404},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected != 404 to pass for 200: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_StatusLessThan(t *testing.T) {
	resp := makeResponse(200, `{}`)
	assertions := []domain.Assertion{
		{Field: "status", Operator: "<", Expected: 300},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected 200 < 300 to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_StatusGreaterThan(t *testing.T) {
	resp := makeResponse(404, `{}`)
	assertions := []domain.Assertion{
		{Field: "status", Operator: ">", Expected: 300},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected 404 > 300 to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_BodyFieldEquals_Pass(t *testing.T) {
	resp := makeResponse(200, `{"name":"John"}`)
	assertions := []domain.Assertion{
		{Field: "body.name", Operator: "==", Expected: "John"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected body.name == John to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_BodyFieldEquals_Fail(t *testing.T) {
	resp := makeResponse(200, `{"name":"Jane"}`)
	assertions := []domain.Assertion{
		{Field: "body.name", Operator: "==", Expected: "John"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if results[0].Passed {
		t.Error("expected body.name == John to fail for Jane")
	}
}

func TestEvaluateAssertions_BodyContains_Pass(t *testing.T) {
	resp := makeResponse(200, `{"message":"hello world"}`)
	assertions := []domain.Assertion{
		{Field: "body", Operator: "contains", Expected: "hello"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected body contains 'hello' to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_BodyContains_Fail(t *testing.T) {
	resp := makeResponse(200, `{"message":"goodbye"}`)
	assertions := []domain.Assertion{
		{Field: "body", Operator: "contains", Expected: "hello"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if results[0].Passed {
		t.Error("expected body contains 'hello' to fail")
	}
}

func TestEvaluateAssertions_HeaderEquals_Pass(t *testing.T) {
	resp := makeResponse(200, `{}`,
		domain.Header{Key: "Content-Type", Value: "application/json"},
	)
	assertions := []domain.Assertion{
		{Field: "header.Content-Type", Operator: "==", Expected: "application/json"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected header check to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_HeaderEquals_Fail(t *testing.T) {
	resp := makeResponse(200, `{}`,
		domain.Header{Key: "Content-Type", Value: "text/html"},
	)
	assertions := []domain.Assertion{
		{Field: "header.Content-Type", Operator: "==", Expected: "application/json"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if results[0].Passed {
		t.Error("expected header check to fail")
	}
}

func TestEvaluateAssertions_HeaderContains(t *testing.T) {
	resp := makeResponse(200, `{}`,
		domain.Header{Key: "Content-Type", Value: "application/json; charset=utf-8"},
	)
	assertions := []domain.Assertion{
		{Field: "header.Content-Type", Operator: "contains", Expected: "json"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected header contains to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_ExistsOperator_Pass(t *testing.T) {
	resp := makeResponse(200, `{"token":"abc123"}`)
	assertions := []domain.Assertion{
		{Field: "body.token", Operator: "exists"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected exists to pass: %s", results[0].Message)
	}
}

func TestEvaluateAssertions_ExistsOperator_Fail(t *testing.T) {
	resp := makeResponse(200, `{"name":"John"}`)
	assertions := []domain.Assertion{
		{Field: "body.token", Operator: "exists"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if results[0].Passed {
		t.Error("expected exists to fail for missing field")
	}
}

func TestEvaluateAssertions_MultipleAssertions(t *testing.T) {
	resp := makeResponse(200, `{"name":"John"}`,
		domain.Header{Key: "Content-Type", Value: "application/json"},
	)
	assertions := []domain.Assertion{
		{Field: "status", Operator: "==", Expected: 200},
		{Field: "body.name", Operator: "==", Expected: "John"},
		{Field: "header.Content-Type", Operator: "contains", Expected: "json"},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if len(results) != 3 {
		t.Fatalf("results count = %d, want 3", len(results))
	}
	for i, r := range results {
		if !r.Passed {
			t.Errorf("assertion %d failed: %s", i, r.Message)
		}
	}
}

func TestEvaluateAssertions_EmptyAssertions(t *testing.T) {
	resp := makeResponse(200, `{}`)
	results := workflow.EvaluateAssertions(nil, resp)
	if len(results) != 0 {
		t.Errorf("expected empty results, got %d", len(results))
	}
}

func TestEvaluateAssertions_BodyNestedFieldEquals(t *testing.T) {
	resp := makeResponse(200, `{"data":{"user":{"active":true}}}`)
	assertions := []domain.Assertion{
		{Field: "body.data.user.active", Operator: "==", Expected: true},
	}

	results := workflow.EvaluateAssertions(assertions, resp)
	if !results[0].Passed {
		t.Errorf("expected nested body field check to pass: %s", results[0].Message)
	}
}

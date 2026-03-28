package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
)

func TestEvaluateCondition_StatusEqualsCompleted(t *testing.T) {
	body := []byte(`{"status":"completed"}`)
	result, err := workflow.EvaluateCondition(body, "$.status == 'completed'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for matching condition, got false")
	}
}

func TestEvaluateCondition_StatusNotCompleted(t *testing.T) {
	body := []byte(`{"status":"pending"}`)
	result, err := workflow.EvaluateCondition(body, "$.status == 'completed'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for non-matching condition, got true")
	}
}

func TestEvaluateCondition_CountGreaterThan(t *testing.T) {
	body := []byte(`{"count":10}`)
	result, err := workflow.EvaluateCondition(body, "$.count > 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for count 10 > 5, got false")
	}
}

func TestEvaluateCondition_CountNotGreaterThan(t *testing.T) {
	body := []byte(`{"count":3}`)
	result, err := workflow.EvaluateCondition(body, "$.count > 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for count 3 > 5, got true")
	}
}

func TestEvaluateCondition_NotEquals(t *testing.T) {
	body := []byte(`{"status":"running"}`)
	result, err := workflow.EvaluateCondition(body, "$.status != 'failed'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for 'running' != 'failed', got false")
	}
}

func TestEvaluateCondition_NotEquals_WhenEqual(t *testing.T) {
	body := []byte(`{"status":"failed"}`)
	result, err := workflow.EvaluateCondition(body, "$.status != 'failed'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result {
		t.Error("expected false for 'failed' != 'failed', got true")
	}
}

func TestEvaluateCondition_LessThan(t *testing.T) {
	body := []byte(`{"score":3}`)
	result, err := workflow.EvaluateCondition(body, "$.score < 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for 3 < 5, got false")
	}
}

func TestEvaluateCondition_LessThanOrEqual(t *testing.T) {
	body := []byte(`{"score":5}`)
	result, err := workflow.EvaluateCondition(body, "$.score <= 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for 5 <= 5, got false")
	}
}

func TestEvaluateCondition_GreaterThanOrEqual(t *testing.T) {
	body := []byte(`{"score":5}`)
	result, err := workflow.EvaluateCondition(body, "$.score >= 5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for 5 >= 5, got false")
	}
}

func TestEvaluateCondition_InvalidSyntax(t *testing.T) {
	body := []byte(`{"status":"ok"}`)
	_, err := workflow.EvaluateCondition(body, "invalid condition")
	if err == nil {
		t.Error("expected error for invalid condition syntax, got nil")
	}
}

func TestEvaluateCondition_MissingField(t *testing.T) {
	body := []byte(`{"other":"value"}`)
	result, err := workflow.EvaluateCondition(body, "$.status == 'completed'")
	if err != nil {
		t.Fatalf("expected no error for missing field, got: %v", err)
	}
	if result {
		t.Error("expected false for missing field, got true")
	}
}

func TestEvaluateCondition_NestedField(t *testing.T) {
	body := []byte(`{"data":{"status":"done"}}`)
	result, err := workflow.EvaluateCondition(body, "$.data.status == 'done'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for nested field match, got false")
	}
}

func TestEvaluateCondition_NumericEquality(t *testing.T) {
	body := []byte(`{"code":200}`)
	result, err := workflow.EvaluateCondition(body, "$.code == 200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result {
		t.Error("expected true for numeric equality, got false")
	}
}

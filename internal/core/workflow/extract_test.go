package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
)

func TestExtractValues_SimpleField(t *testing.T) {
	body := []byte(`{"name":"John"}`)
	exprs := map[string]string{"result": "$.name"}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["result"] != "John" {
		t.Errorf("result = %q, want %q", vals["result"], "John")
	}
}

func TestExtractValues_NestedField(t *testing.T) {
	body := []byte(`{"user":{"name":"Jane","email":"jane@example.com"}}`)
	exprs := map[string]string{
		"name":  "$.user.name",
		"email": "$.user.email",
	}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["name"] != "Jane" {
		t.Errorf("name = %q, want %q", vals["name"], "Jane")
	}
	if vals["email"] != "jane@example.com" {
		t.Errorf("email = %q, want %q", vals["email"], "jane@example.com")
	}
}

func TestExtractValues_NumberField(t *testing.T) {
	body := []byte(`{"id":42}`)
	exprs := map[string]string{"id": "$.id"}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["id"] != "42" {
		t.Errorf("id = %q, want %q", vals["id"], "42")
	}
}

func TestExtractValues_BoolField(t *testing.T) {
	body := []byte(`{"active":true}`)
	exprs := map[string]string{"active": "$.active"}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["active"] != "true" {
		t.Errorf("active = %q, want %q", vals["active"], "true")
	}
}

func TestExtractValues_ArrayIndex(t *testing.T) {
	body := []byte(`{"items":[{"id":"first"},{"id":"second"}]}`)
	exprs := map[string]string{"first_id": "$.items[0].id"}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["first_id"] != "first" {
		t.Errorf("first_id = %q, want %q", vals["first_id"], "first")
	}
}

func TestExtractValues_ArrayIndexTopLevel(t *testing.T) {
	body := []byte(`{"tags":["alpha","beta","gamma"]}`)
	exprs := map[string]string{"second": "$.tags[1]"}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vals["second"] != "beta" {
		t.Errorf("second = %q, want %q", vals["second"], "beta")
	}
}

func TestExtractValues_MissingField(t *testing.T) {
	body := []byte(`{"name":"John"}`)
	exprs := map[string]string{"missing": "$.nonexistent"}

	_, err := workflow.ExtractValues(body, exprs)
	if err == nil {
		t.Fatal("expected error for missing field, got nil")
	}
}

func TestExtractValues_InvalidJSON(t *testing.T) {
	body := []byte(`not json`)
	exprs := map[string]string{"x": "$.field"}

	_, err := workflow.ExtractValues(body, exprs)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestExtractValues_EmptyExpressions(t *testing.T) {
	body := []byte(`{"name":"John"}`)
	exprs := map[string]string{}

	vals, err := workflow.ExtractValues(body, exprs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 0 {
		t.Errorf("expected empty map, got %v", vals)
	}
}

func TestExtractValues_NilExpressions(t *testing.T) {
	body := []byte(`{"name":"John"}`)

	vals, err := workflow.ExtractValues(body, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) != 0 {
		t.Errorf("expected empty map, got %v", vals)
	}
}

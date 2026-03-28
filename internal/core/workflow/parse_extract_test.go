package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
)

func TestParseExtractString_BareJSONPath(t *testing.T) {
	varName, jsonPath, err := workflow.ParseExtractString("$.id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if varName != "" {
		t.Errorf("varName = %q, want empty", varName)
	}
	if jsonPath != "$.id" {
		t.Errorf("jsonPath = %q, want %q", jsonPath, "$.id")
	}
}

func TestParseExtractString_WithLabel(t *testing.T) {
	varName, jsonPath, err := workflow.ParseExtractString("user_id=$.id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if varName != "user_id" {
		t.Errorf("varName = %q, want %q", varName, "user_id")
	}
	if jsonPath != "$.id" {
		t.Errorf("jsonPath = %q, want %q", jsonPath, "$.id")
	}
}

func TestParseExtractString_Empty(t *testing.T) {
	_, _, err := workflow.ParseExtractString("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestParseExtractString_NestedPath(t *testing.T) {
	varName, jsonPath, err := workflow.ParseExtractString("host=$.headers.Host")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if varName != "host" {
		t.Errorf("varName = %q, want %q", varName, "host")
	}
	if jsonPath != "$.headers.Host" {
		t.Errorf("jsonPath = %q, want %q", jsonPath, "$.headers.Host")
	}
}

func TestParseExtractString_BareNestedPath(t *testing.T) {
	varName, jsonPath, err := workflow.ParseExtractString("$.data.items[0].name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if varName != "" {
		t.Errorf("varName = %q, want empty", varName)
	}
	if jsonPath != "$.data.items[0].name" {
		t.Errorf("jsonPath = %q, want %q", jsonPath, "$.data.items[0].name")
	}
}

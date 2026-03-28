package workflow_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/core/workflow"
)

func TestParseAssertionString_StatusEquals(t *testing.T) {
	a, err := workflow.ParseAssertionString("status == 200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "status" {
		t.Errorf("Field = %q, want %q", a.Field, "status")
	}
	if a.Operator != "==" {
		t.Errorf("Operator = %q, want %q", a.Operator, "==")
	}
	// Expected should be numeric 200
	if expected, ok := a.Expected.(int); !ok || expected != 200 {
		t.Errorf("Expected = %v (%T), want 200 (int)", a.Expected, a.Expected)
	}
}

func TestParseAssertionString_BodyFieldEquals(t *testing.T) {
	a, err := workflow.ParseAssertionString("body.name == John")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "body.name" {
		t.Errorf("Field = %q, want %q", a.Field, "body.name")
	}
	if a.Operator != "==" {
		t.Errorf("Operator = %q, want %q", a.Operator, "==")
	}
	if a.Expected != "John" {
		t.Errorf("Expected = %v, want %q", a.Expected, "John")
	}
}

func TestParseAssertionString_BodyContains(t *testing.T) {
	a, err := workflow.ParseAssertionString("body contains 'hello'")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "body" {
		t.Errorf("Field = %q, want %q", a.Field, "body")
	}
	if a.Operator != "contains" {
		t.Errorf("Operator = %q, want %q", a.Operator, "contains")
	}
	if a.Expected != "hello" {
		t.Errorf("Expected = %v, want %q", a.Expected, "hello")
	}
}

func TestParseAssertionString_DurationLessThan(t *testing.T) {
	a, err := workflow.ParseAssertionString("duration < 500")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "duration" {
		t.Errorf("Field = %q, want %q", a.Field, "duration")
	}
	if a.Operator != "<" {
		t.Errorf("Operator = %q, want %q", a.Operator, "<")
	}
	if expected, ok := a.Expected.(int); !ok || expected != 500 {
		t.Errorf("Expected = %v (%T), want 500 (int)", a.Expected, a.Expected)
	}
}

func TestParseAssertionString_HeaderEquals(t *testing.T) {
	a, err := workflow.ParseAssertionString("header.Content-Type == application/json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "header.Content-Type" {
		t.Errorf("Field = %q, want %q", a.Field, "header.Content-Type")
	}
	if a.Operator != "==" {
		t.Errorf("Operator = %q, want %q", a.Operator, "==")
	}
	if a.Expected != "application/json" {
		t.Errorf("Expected = %v, want %q", a.Expected, "application/json")
	}
}

func TestParseAssertionString_Empty(t *testing.T) {
	_, err := workflow.ParseAssertionString("")
	if err == nil {
		t.Fatal("expected error for empty string, got nil")
	}
}

func TestParseAssertionString_InvalidFormat(t *testing.T) {
	_, err := workflow.ParseAssertionString("invalid")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseAssertionString_StatusNotEquals(t *testing.T) {
	a, err := workflow.ParseAssertionString("status != 404")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Field != "status" {
		t.Errorf("Field = %q, want %q", a.Field, "status")
	}
	if a.Operator != "!=" {
		t.Errorf("Operator = %q, want %q", a.Operator, "!=")
	}
	if expected, ok := a.Expected.(int); !ok || expected != 404 {
		t.Errorf("Expected = %v (%T), want 404 (int)", a.Expected, a.Expected)
	}
}

func TestParseAssertionString_StatusGreaterThan(t *testing.T) {
	a, err := workflow.ParseAssertionString("status > 199")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Operator != ">" {
		t.Errorf("Operator = %q, want %q", a.Operator, ">")
	}
	if expected, ok := a.Expected.(int); !ok || expected != 199 {
		t.Errorf("Expected = %v (%T), want 199 (int)", a.Expected, a.Expected)
	}
}

func TestParseAssertionString_BodyContainsWithoutQuotes(t *testing.T) {
	a, err := workflow.ParseAssertionString("body contains hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Expected != "hello" {
		t.Errorf("Expected = %v, want %q", a.Expected, "hello")
	}
}

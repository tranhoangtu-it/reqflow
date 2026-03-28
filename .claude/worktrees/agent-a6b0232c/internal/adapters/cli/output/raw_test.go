package output

import (
	"bytes"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestRawFormatter_OutputsOnlyBody(t *testing.T) {
	f := &RawFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"id":1,"name":"John"}`),
		Duration: 100 * time.Millisecond,
		Size:     21,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{"id":1,"name":"John"}`
	if buf.String() != expected {
		t.Errorf("expected exactly %q, got %q", expected, buf.String())
	}
}

func TestRawFormatter_EmptyBody_OutputsNothing(t *testing.T) {
	f := &RawFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 204,
		Status:     "204 No Content",
		Headers:    []domain.Header{},
		Body:       nil,
		Duration:   5 * time.Millisecond,
		Size:       0,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty body, got %q", buf.String())
	}
}

func TestRawFormatter_DoesNotAddNewline(t *testing.T) {
	f := &RawFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("no trailing newline"),
		Duration:   1 * time.Millisecond,
		Size:       19,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if output != "no trailing newline" {
		t.Errorf("expected exact body without added newline, got %q", output)
	}
}

func TestRawFormatter_PreservesExistingNewline(t *testing.T) {
	f := &RawFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("has newline\n"),
		Duration:   1 * time.Millisecond,
		Size:       12,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != "has newline\n" {
		t.Errorf("expected body with original newline, got %q", buf.String())
	}
}

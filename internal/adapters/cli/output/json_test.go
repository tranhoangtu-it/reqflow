package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestJSONFormatter_JSONBody_ParsedAsObject(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"id":1,"name":"John"}`),
		Duration: 123 * time.Millisecond,
		Size:     42,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	// status_code
	if sc, ok := result["status_code"].(float64); !ok || int(sc) != 200 {
		t.Errorf("expected status_code=200, got %v", result["status_code"])
	}

	// status
	if s, ok := result["status"].(string); !ok || s != "200 OK" {
		t.Errorf("expected status='200 OK', got %v", result["status"])
	}

	// body should be a parsed object, not a string
	body, ok := result["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected body to be an object, got %T: %v", result["body"], result["body"])
	}
	if id, ok := body["id"].(float64); !ok || int(id) != 1 {
		t.Errorf("expected body.id=1, got %v", body["id"])
	}
	if name, ok := body["name"].(string); !ok || name != "John" {
		t.Errorf("expected body.name='John', got %v", body["name"])
	}

	// duration_ms should be integer
	if dm, ok := result["duration_ms"].(float64); !ok || int(dm) != 123 {
		t.Errorf("expected duration_ms=123, got %v", result["duration_ms"])
	}

	// size
	if sz, ok := result["size"].(float64); !ok || int(sz) != 42 {
		t.Errorf("expected size=42, got %v", result["size"])
	}

	// headers
	headers, ok := result["headers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected headers to be an object, got %T", result["headers"])
	}
	if ct, ok := headers["Content-Type"].(string); !ok || ct != "application/json" {
		t.Errorf("expected headers.Content-Type='application/json', got %v", headers["Content-Type"])
	}
}

func TestJSONFormatter_NonJSONBody_BodyIsString(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("plain text"),
		Duration:   10 * time.Millisecond,
		Size:       10,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	body, ok := result["body"].(string)
	if !ok {
		t.Fatalf("expected body to be a string for non-JSON, got %T: %v", result["body"], result["body"])
	}
	if body != "plain text" {
		t.Errorf("expected body='plain text', got %q", body)
	}
}

func TestJSONFormatter_EmptyBody(t *testing.T) {
	f := &JSONFormatter{}
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

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// body should be null
	if result["body"] != nil {
		t.Errorf("expected body to be null for empty body, got %v", result["body"])
	}
}

func TestJSONFormatter_HeadersFlattened(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "text/plain"},
			{Key: "X-Dup", Value: "first"},
			{Key: "X-Dup", Value: "second"},
		},
		Body:     []byte("ok"),
		Duration: 1 * time.Millisecond,
		Size:     2,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(buf.Bytes(), &result)

	headers := result["headers"].(map[string]interface{})
	// First value wins for duplicate keys
	if v := headers["X-Dup"].(string); v != "first" {
		t.Errorf("expected first value for duplicate key, got %q", v)
	}
}

func TestJSONFormatter_DurationIsMilliseconds(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers:    []domain.Header{},
		Body:       []byte("ok"),
		Duration:   2500 * time.Millisecond,
		Size:       2,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(buf.Bytes(), &result)

	dm := result["duration_ms"].(float64)
	if int(dm) != 2500 {
		t.Errorf("expected duration_ms=2500, got %v", dm)
	}
}

func TestJSONFormatter_OutputIsValidJSON(t *testing.T) {
	f := &JSONFormatter{}
	var buf bytes.Buffer

	resp := domain.HTTPResponse{
		StatusCode: 200,
		Status:     "200 OK",
		Headers: []domain.Header{
			{Key: "Content-Type", Value: "application/json"},
		},
		Body:     []byte(`{"key":"value"}`),
		Duration: 100 * time.Millisecond,
		Size:     15,
	}

	err := f.FormatResponse(&buf, resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Errorf("output is not valid JSON: %s", buf.String())
	}
}

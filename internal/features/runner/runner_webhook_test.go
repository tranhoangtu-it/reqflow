package runner_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

// webhookTriggerClient is a mock HTTP client that sends a POST to the webhook
// listener URL embedded in the request body, simulating an async callback.
type webhookTriggerClient struct {
	callbackBody string
	calls        int
	responses    []domain.HTTPResponse
}

func (m *webhookTriggerClient) Do(_ context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
	idx := m.calls
	m.calls++

	// If the request body contains a webhook URL, extract it and POST to it
	if req.Body != nil {
		bodyStr := string(req.Body)
		if start := strings.Index(bodyStr, "http://localhost:"); start >= 0 {
			// Extract the URL up to the next quote
			end := strings.Index(bodyStr[start:], `"`)
			if end > 0 {
				webhookURL := bodyStr[start : start+end]
				go func() {
					time.Sleep(20 * time.Millisecond)
					http.Post(webhookURL, "application/json",
						bytes.NewBufferString(m.callbackBody))
				}()
			}
		}
	}

	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return domain.HTTPResponse{StatusCode: 200, Body: []byte(`{}`)}, nil
}

func TestRunner_ListenStepCapturesCallback(t *testing.T) {
	// This test verifies the full flow:
	// 1. Listen step starts a webhook server
	// 2. Next step triggers an API that calls back to the webhook
	// 3. The captured variable is available in subsequent steps

	client := &webhookTriggerClient{
		callbackBody: `{"result":"success","id":"abc-123"}`,
		responses: []domain.HTTPResponse{
			// Response to the trigger step
			{StatusCode: 202, Body: []byte(`{"queued":true}`)},
			// Response to the verification step
			{StatusCode: 200, Body: []byte(`{"verified":true}`)},
		},
	}

	wf := domain.Workflow{
		Name: "webhook-test",
		Steps: []domain.Step{
			{
				Name: "wait for callback",
				Listen: &domain.ListenConfig{
					Port:    0, // ephemeral port
					Path:    "/callback",
					Timeout: 5 * time.Second,
					Capture: "callback_data",
				},
			},
			{
				Name:   "trigger async job",
				Method: domain.MethodPost,
				URL:    "https://api.example.com/trigger",
				Body:   `{"webhook":"http://localhost:{{listen_port}}/callback"}`,
			},
			{
				Name:   "use captured data",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/verify/{{callback_data}}",
			},
		},
	}

	r := runner.New(client)
	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// All three steps should have completed
	if len(result.Steps) != 3 {
		t.Fatalf("steps count = %d, want 3", len(result.Steps))
	}

	// Listen step should succeed
	if result.Steps[0].Error != nil {
		t.Errorf("listen step error: %v", result.Steps[0].Error)
	}

	// The captured variable should contain the callback body
	if result.Steps[0].Extracted["callback_data"] != `{"result":"success","id":"abc-123"}` {
		t.Errorf("captured data = %q, want %q",
			result.Steps[0].Extracted["callback_data"],
			`{"result":"success","id":"abc-123"}`)
	}

	// Trigger step should succeed
	if result.Steps[1].Error != nil {
		t.Errorf("trigger step error: %v", result.Steps[1].Error)
	}

	// Verify step should succeed
	if result.Steps[2].Error != nil {
		t.Errorf("verify step error: %v", result.Steps[2].Error)
	}
}

func TestRunner_ListenStepTimeout(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{}`)},
		},
	}

	wf := domain.Workflow{
		Name: "timeout-test",
		Steps: []domain.Step{
			{
				Name: "wait for callback",
				Listen: &domain.ListenConfig{
					Port:    0,
					Path:    "/webhook",
					Timeout: 100 * time.Millisecond,
					Capture: "callback_body",
				},
			},
			{
				Name:   "this should not run",
				Method: domain.MethodGet,
				URL:    "https://api.example.com/check",
			},
		},
	}

	r := runner.New(client)
	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Listen step should have failed
	if len(result.Steps) < 1 {
		t.Fatal("expected at least 1 step result")
	}
	if result.Steps[0].Error == nil {
		t.Error("expected listen step to have an error (timeout)")
	}

	// The next step should not have executed since listen failed
	if len(result.Steps) > 1 {
		t.Errorf("steps count = %d, want 1 (should stop on listen failure)", len(result.Steps))
	}
}

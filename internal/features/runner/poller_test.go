package runner_test

import (
	"context"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/features/runner"
)

func TestPollStep_ConditionMetOnFirstTry(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{
				StatusCode: 200,
				Body:       []byte(`{"status":"completed"}`),
			},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 100 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Steps) != 1 {
		t.Fatalf("expected 1 step result, got %d", len(result.Steps))
	}
	sr := result.Steps[0]
	if sr.Error != nil {
		t.Fatalf("unexpected step error: %v", sr.Error)
	}
	if sr.Response.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", sr.Response.StatusCode)
	}
	if len(sr.PollAttempts) != 1 {
		t.Errorf("expected 1 poll attempt, got %d", len(sr.PollAttempts))
	}
	if !sr.PollAttempts[0].ConditionMet {
		t.Error("expected first attempt condition to be met")
	}
}

func TestPollStep_ConditionMetOnThirdTry(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"running"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"completed"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error != nil {
		t.Fatalf("unexpected step error: %v", sr.Error)
	}
	if len(sr.PollAttempts) != 3 {
		t.Errorf("expected 3 poll attempts, got %d", len(sr.PollAttempts))
	}
	// Final response should be the completed one
	if string(sr.Response.Body) != `{"status":"completed"}` {
		t.Errorf("expected final body with completed status, got %s", string(sr.Response.Body))
	}
}

func TestPollStep_TimeoutExceeded(t *testing.T) {
	// Always returns pending
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  50 * time.Millisecond,
			Until:    "$.status == 'completed'",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if len(sr.PollAttempts) == 0 {
		t.Error("expected at least one poll attempt")
	}
}

func TestPollStep_ContextCancelled(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately after a short delay
	go func() {
		time.Sleep(25 * time.Millisecond)
		cancel()
	}()

	result, err := r.Run(ctx, wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
}

func TestPollStep_ExtractsFromFinalResponse(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending","result":""}`)},
			{StatusCode: 200, Body: []byte(`{"status":"completed","result":"data-123"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
		},
		Extract: map[string]string{
			"result": "$.result",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error != nil {
		t.Fatalf("unexpected step error: %v", sr.Error)
	}
	if sr.Extracted["result"] != "data-123" {
		t.Errorf("expected extracted result='data-123', got %q", sr.Extracted["result"])
	}
}

func TestPollStep_LinearBackoff(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"completed"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
			Backoff:  "linear",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error != nil {
		t.Fatalf("unexpected step error: %v", sr.Error)
	}
	if len(sr.PollAttempts) != 3 {
		t.Errorf("expected 3 poll attempts with linear backoff, got %d", len(sr.PollAttempts))
	}
}

func TestPollStep_ExponentialBackoff(t *testing.T) {
	client := &mockHTTPClient{
		responses: []domain.HTTPResponse{
			{StatusCode: 200, Body: []byte(`{"status":"pending"}`)},
			{StatusCode: 200, Body: []byte(`{"status":"completed"}`)},
		},
	}

	r := runner.New(client)
	step := domain.Step{
		Name:   "poll job",
		Method: domain.MethodGet,
		URL:    "https://api.example.com/jobs/1",
		Poll: &domain.PollConfig{
			Interval: 10 * time.Millisecond,
			Timeout:  5 * time.Second,
			Until:    "$.status == 'completed'",
			Backoff:  "exponential",
		},
	}

	wf := domain.Workflow{
		Name:  "poll-test",
		Steps: []domain.Step{step},
	}

	result, err := r.Run(context.Background(), wf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sr := result.Steps[0]
	if sr.Error != nil {
		t.Fatalf("unexpected step error: %v", sr.Error)
	}
	if len(sr.PollAttempts) != 2 {
		t.Errorf("expected 2 poll attempts with exponential backoff, got %d", len(sr.PollAttempts))
	}
}

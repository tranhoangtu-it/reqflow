package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/ye-kart/reqflow/internal/core/variable"
	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// Runner executes workflows by stepping through each HTTP request sequentially.
type Runner struct {
	httpClient driven.HTTPClient
}

// New creates a new Runner with the given HTTP client.
func New(httpClient driven.HTTPClient) *Runner {
	return &Runner{httpClient: httpClient}
}

// Run executes a workflow sequentially, chaining extracted variables between
// steps and evaluating assertions. Stops on first assertion failure.
func (r *Runner) Run(ctx context.Context, wf domain.Workflow, initialVars map[string]string) (domain.WorkflowResult, error) {
	start := time.Now()

	vars := make(map[string]string)
	for k, v := range initialVars {
		vars[k] = v
	}

	result := domain.WorkflowResult{
		Name: wf.Name,
	}

	for i := 0; i < len(wf.Steps); i++ {
		step := wf.Steps[i]

		// Handle listen steps: start webhook listener and run next step concurrently
		if step.Listen != nil {
			listenResult, triggerResult, skip := r.executeListenStep(ctx, step, wf.Steps, i, vars)

			if stop := accumulateResult(&result, listenResult, vars); stop {
				break
			}

			// If we ran the next step concurrently, record it too
			if triggerResult != nil {
				if stop := accumulateResult(&result, *triggerResult, vars); stop {
					break
				}
				i += skip
			}

			continue
		}

		// Handle parallel steps
		if len(step.Parallel) > 0 {
			subResults, mergedVars, err := r.runParallel(ctx, step, vars)
			if err != nil {
				break
			}
			hasError := false
			for _, sr := range subResults {
				if stop := accumulateResult(&result, sr, vars); stop {
					hasError = true
				}
			}
			// Merge parallel extracted vars
			for k, v := range mergedVars {
				vars[k] = v
			}
			if hasError {
				break
			}
			continue
		}

		// Handle poll steps
		if step.Poll != nil {
			sr := r.pollStep(ctx, step, vars)
			if stop := accumulateResult(&result, sr, vars); stop {
				break
			}
			continue
		}

		// Handle retry steps
		if step.Retry != nil {
			sr := r.retryStep(ctx, step, vars)
			if stop := accumulateResult(&result, sr, vars); stop {
				break
			}
			continue
		}

		// Default: execute step normally
		stepResult := r.executeStep(ctx, step, vars)
		if stop := accumulateResult(&result, stepResult, vars); stop {
			break
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// accumulateResult appends a step result to the workflow result, counts assertions,
// merges extracted variables, and returns true if execution should stop.
func accumulateResult(result *domain.WorkflowResult, sr domain.StepResult, vars map[string]string) bool {
	result.Steps = append(result.Steps, sr)

	for _, ar := range sr.Assertions {
		if ar.Passed {
			result.TotalPassed++
		} else {
			result.TotalFailed++
		}
	}

	if sr.Error != nil {
		return true
	}

	for k, v := range sr.Extracted {
		vars[k] = v
	}

	return result.TotalFailed > 0
}

// executeListenStep handles a step with Listen config. It starts the webhook
// listener, runs the next step concurrently (if available), and waits for
// the callback. Returns the listen step result, an optional trigger step
// result, and the number of extra steps consumed.
func (r *Runner) executeListenStep(ctx context.Context, step domain.Step, steps []domain.Step, idx int, vars map[string]string) (domain.StepResult, *domain.StepResult, int) {
	start := time.Now()

	listenResult := domain.StepResult{
		StepName:  step.Name,
		Extracted: make(map[string]string),
	}

	listener := NewWebhookListener(*step.Listen)
	resultCh, err := listener.Start(ctx)
	if err != nil {
		listenResult.Error = fmt.Errorf("starting webhook listener: %w", err)
		listenResult.Duration = time.Since(start)
		return listenResult, nil, 0
	}

	// Store the listener port so the next step can reference it
	port := listener.Port()
	vars["listen_port"] = strconv.Itoa(port)

	// If there's a next step, run it concurrently (it's the trigger)
	var triggerResult *domain.StepResult
	skip := 0
	if idx+1 < len(steps) {
		nextStep := steps[idx+1]
		skip = 1

		// Run the next step in a goroutine
		triggerCh := make(chan domain.StepResult, 1)
		go func() {
			triggerCh <- r.executeStep(ctx, nextStep, vars)
		}()

		// Wait for the webhook callback
		webhookResult := <-resultCh
		listener.Stop()

		if webhookResult.Error != nil {
			listenResult.Error = fmt.Errorf("webhook listener: %w", webhookResult.Error)
			listenResult.Duration = time.Since(start)
			return listenResult, nil, skip
		}

		// Store captured body as the configured variable
		listenResult.Extracted[step.Listen.Capture] = string(webhookResult.Body)
		listenResult.Duration = time.Since(start)

		// Wait for the trigger step to complete
		tr := <-triggerCh
		triggerResult = &tr

		return listenResult, triggerResult, skip
	}

	// No next step - just wait for the callback
	webhookResult := <-resultCh
	listener.Stop()

	if webhookResult.Error != nil {
		listenResult.Error = fmt.Errorf("webhook listener: %w", webhookResult.Error)
		listenResult.Duration = time.Since(start)
		return listenResult, nil, 0
	}

	listenResult.Extracted[step.Listen.Capture] = string(webhookResult.Body)
	listenResult.Duration = time.Since(start)
	return listenResult, nil, 0
}

func (r *Runner) executeStep(ctx context.Context, step domain.Step, vars map[string]string) domain.StepResult {
	start := time.Now()

	sr := domain.StepResult{
		StepName:  step.Name,
		Extracted: make(map[string]string),
	}

	// Build the request config from the step definition
	config, err := buildRequestConfig(step, vars)
	if err != nil {
		sr.Error = fmt.Errorf("building request: %w", err)
		sr.Duration = time.Since(start)
		return sr
	}

	// Build the HTTP request (apply variable interpolation)
	req := buildHTTPRequest(config, vars)
	sr.Request = req

	// Execute the HTTP request
	resp, err := r.httpClient.Do(ctx, req)
	if err != nil {
		sr.Error = fmt.Errorf("executing request: %w", err)
		sr.Duration = time.Since(start)
		return sr
	}
	sr.Response = resp

	// Extract values from response body
	if len(step.Extract) > 0 {
		extracted, err := workflow.ExtractValues(resp.Body, step.Extract)
		if err != nil {
			sr.Error = fmt.Errorf("extracting values: %w", err)
			sr.Duration = time.Since(start)
			return sr
		}
		sr.Extracted = extracted
	}

	// Evaluate assertions
	if len(step.Assert) > 0 {
		sr.Assertions = workflow.EvaluateAssertions(step.Assert, resp)
	}

	sr.Duration = time.Since(start)
	return sr
}

func buildRequestConfig(step domain.Step, vars map[string]string) (domain.RequestConfig, error) {
	config := domain.RequestConfig{
		Method:      step.Method,
		URL:         variable.Interpolate(step.URL, vars),
		ContentType: step.ContentType,
		Auth:        step.Auth,
	}

	// Convert headers
	if len(step.Headers) > 0 {
		headers := make([]domain.Header, 0, len(step.Headers))
		for k, v := range step.Headers {
			headers = append(headers, domain.Header{
				Key:   k,
				Value: variable.Interpolate(v, vars),
			})
		}
		config.Headers = headers
	}

	// Convert body
	if step.Body != nil {
		bodyBytes, err := resolveBody(step.Body, vars)
		if err != nil {
			return domain.RequestConfig{}, err
		}
		config.Body = bodyBytes
	}

	return config, nil
}

func resolveBody(body interface{}, vars map[string]string) ([]byte, error) {
	switch b := body.(type) {
	case string:
		return []byte(variable.Interpolate(b, vars)), nil
	case map[string]interface{}:
		data, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("marshaling body map: %w", err)
		}
		// Interpolate variables in the JSON string
		return []byte(variable.Interpolate(string(data), vars)), nil
	default:
		data, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("marshaling body: %w", err)
		}
		return data, nil
	}
}

func buildHTTPRequest(config domain.RequestConfig, vars map[string]string) domain.HTTPRequest {
	_ = vars // already interpolated in config building
	return domain.HTTPRequest{
		Method:      config.Method,
		URL:         config.URL,
		Headers:     config.Headers,
		Body:        config.Body,
		ContentType: config.ContentType,
	}
}

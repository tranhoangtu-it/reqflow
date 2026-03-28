package runner

import (
	"context"
	"encoding/json"
	"fmt"
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

	for _, step := range wf.Steps {
		var stepResults []domain.StepResult
		var extractedVars map[string]string

		if len(step.Parallel) > 0 {
			subResults, mergedVars, err := r.runParallel(ctx, step, vars)
			if err != nil {
				break
			}
			stepResults = subResults
			extractedVars = mergedVars
		} else if step.Poll != nil {
			sr := r.pollStep(ctx, step, vars)
			stepResults = []domain.StepResult{sr}
			extractedVars = sr.Extracted
		} else if step.Retry != nil {
			sr := r.retryStep(ctx, step, vars)
			stepResults = []domain.StepResult{sr}
			extractedVars = sr.Extracted
		} else {
			sr := r.executeStep(ctx, step, vars)
			stepResults = []domain.StepResult{sr}
			extractedVars = sr.Extracted
		}

		// Collect results and count assertions
		hasError := false
		for _, sr := range stepResults {
			result.Steps = append(result.Steps, sr)
			if sr.Error != nil {
				hasError = true
			}
			for _, ar := range sr.Assertions {
				if ar.Passed {
					result.TotalPassed++
				} else {
					result.TotalFailed++
				}
			}
		}

		if hasError {
			break
		}

		// Merge extracted variables for subsequent steps
		for k, v := range extractedVars {
			vars[k] = v
		}

		if result.TotalFailed > 0 {
			break
		}
	}

	result.Duration = time.Since(start)
	return result, nil
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

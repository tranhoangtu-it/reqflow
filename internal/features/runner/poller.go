package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
)

// pollStep executes a step with poll-until-ready behavior. It repeatedly sends
// the request and evaluates the poll condition against the response body. If the
// condition is met, it returns the final response. If the timeout is exceeded or
// the context is cancelled, it returns an error.
func (r *Runner) pollStep(ctx context.Context, step domain.Step, vars map[string]string) domain.StepResult {
	start := time.Now()
	poll := step.Poll

	sr := domain.StepResult{
		StepName:  step.Name,
		Extracted: make(map[string]string),
	}

	// Create a timeout context for the entire polling operation
	pollCtx, cancel := context.WithTimeout(ctx, poll.Timeout)
	defer cancel()

	attempt := 0
	for {
		attempt++
		attemptStart := time.Now()

		// Build and execute request
		config, err := buildRequestConfig(step, vars)
		if err != nil {
			sr.Error = fmt.Errorf("building request: %w", err)
			sr.Duration = time.Since(start)
			return sr
		}

		req := buildHTTPRequest(config, vars)
		sr.Request = req

		resp, err := r.httpClient.Do(pollCtx, req)
		if err != nil {
			// Check if it's a context error
			if pollCtx.Err() != nil {
				sr.Error = fmt.Errorf("polling cancelled: %w", pollCtx.Err())
				sr.Duration = time.Since(start)
				return sr
			}
			sr.Error = fmt.Errorf("executing poll request: %w", err)
			sr.Duration = time.Since(start)
			return sr
		}

		// Evaluate condition
		conditionMet, condErr := workflow.EvaluateCondition(resp.Body, poll.Until)

		pa := domain.PollAttempt{
			Number:       attempt,
			StatusCode:   resp.StatusCode,
			Body:         resp.Body,
			Duration:     time.Since(attemptStart),
			ConditionMet: conditionMet,
		}
		sr.PollAttempts = append(sr.PollAttempts, pa)

		if condErr != nil {
			sr.Error = fmt.Errorf("evaluating poll condition: %w", condErr)
			sr.Response = resp
			sr.Duration = time.Since(start)
			return sr
		}

		if conditionMet {
			sr.Response = resp

			// Extract values from the final successful response
			if len(step.Extract) > 0 {
				extracted, err := workflow.ExtractValues(resp.Body, step.Extract)
				if err != nil {
					sr.Error = fmt.Errorf("extracting values: %w", err)
					sr.Duration = time.Since(start)
					return sr
				}
				sr.Extracted = extracted
			}

			// Evaluate assertions against the final response
			if len(step.Assert) > 0 {
				sr.Assertions = workflow.EvaluateAssertions(step.Assert, resp)
			}

			sr.Duration = time.Since(start)
			return sr
		}

		// Wait for the next interval or context cancellation
		waitDuration := calculateWaitDuration(poll.Interval, poll.Backoff, attempt)

		select {
		case <-pollCtx.Done():
			sr.Response = resp
			sr.Error = fmt.Errorf("polling timed out after %d attempts: %w", attempt, pollCtx.Err())
			sr.Duration = time.Since(start)
			return sr
		case <-time.After(waitDuration):
			// Continue to next attempt
		}
	}
}

// calculateWaitDuration computes the wait duration based on backoff strategy.
func calculateWaitDuration(baseInterval time.Duration, backoff string, attempt int) time.Duration {
	switch backoff {
	case "linear":
		return baseInterval * time.Duration(attempt)
	case "exponential":
		multiplier := time.Duration(1)
		for i := 1; i < attempt; i++ {
			multiplier *= 2
		}
		return baseInterval * multiplier
	default: // "fixed" or empty
		return baseInterval
	}
}

package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/ye-kart/reqflow/internal/core/workflow"
	"github.com/ye-kart/reqflow/internal/domain"
)

// retryStep executes a step with retry logic based on the step's RetryConfig.
func (r *Runner) retryStep(ctx context.Context, step domain.Step, vars map[string]string) domain.StepResult {
	cfg := step.Retry
	if cfg == nil {
		return r.executeStep(ctx, step, vars)
	}

	var lastResult domain.StepResult

	// Attempt 0 is the initial attempt, then up to cfg.Max retries
	for attempt := 0; attempt <= cfg.Max; attempt++ {
		// Wait before retry (skip delay on first attempt)
		if attempt > 0 {
			delay := workflow.CalculateDelay(cfg.Backoff, attempt, cfg.InitialDelay)
			if cfg.Jitter {
				delay = workflow.ApplyJitter(delay)
			}

			select {
			case <-ctx.Done():
				lastResult.Error = fmt.Errorf("retry cancelled: %w", ctx.Err())
				return lastResult
			case <-time.After(delay):
			}
		}

		lastResult = r.executeStep(ctx, step, vars)

		// If there was a network/execution error
		if lastResult.Error != nil {
			if cfg.RetryOnError && attempt < cfg.Max {
				continue
			}
			return lastResult
		}

		// Check if status code is in the retry list
		if shouldRetryStatus(lastResult.Response.StatusCode, cfg.RetryOn) {
			if attempt < cfg.Max {
				continue
			}
			// Exhausted all retries with a retryable status
			lastResult.Error = fmt.Errorf("max retries (%d) exceeded, last status: %d", cfg.Max, lastResult.Response.StatusCode)
			return lastResult
		}

		// Success or non-retryable status
		return lastResult
	}

	return lastResult
}

// shouldRetryStatus checks if the given status code is in the retryable list.
func shouldRetryStatus(statusCode int, retryOn []int) bool {
	for _, code := range retryOn {
		if statusCode == code {
			return true
		}
	}
	return false
}

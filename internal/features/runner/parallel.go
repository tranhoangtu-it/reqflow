package runner

import (
	"context"
	"sync"

	"github.com/ye-kart/reqflow/internal/domain"
)

// parallelResult holds the result of a single parallel sub-step along with
// its original index to preserve ordering.
type parallelResult struct {
	index  int
	result domain.StepResult
}

// runParallel executes the sub-steps in step.Parallel concurrently.
// It respects MaxParallel for concurrency limiting and FailFast for early cancellation.
// Returns all sub-step results (ordered by original index) and merged extracted variables.
func (r *Runner) runParallel(ctx context.Context, step domain.Step, vars map[string]string) ([]domain.StepResult, map[string]string, error) {
	subSteps := step.Parallel

	// Create a cancellable context for fail-fast support
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Semaphore for concurrency limiting
	var sem chan struct{}
	if step.MaxParallel > 0 {
		sem = make(chan struct{}, step.MaxParallel)
	}

	// Snapshot vars so parallel steps share the same initial variable state
	snapVars := make(map[string]string, len(vars))
	for k, v := range vars {
		snapVars[k] = v
	}

	results := make([]parallelResult, len(subSteps))
	var wg sync.WaitGroup

	for i, sub := range subSteps {
		wg.Add(1)
		go func(idx int, s domain.Step) {
			defer wg.Done()

			// Acquire semaphore slot if limited
			if sem != nil {
				select {
				case sem <- struct{}{}:
					defer func() { <-sem }()
				case <-runCtx.Done():
					results[idx] = parallelResult{
						index: idx,
						result: domain.StepResult{
							StepName:  s.Name,
							Error:     runCtx.Err(),
							Extracted: make(map[string]string),
						},
					}
					return
				}
			}

			// Check context before executing
			if runCtx.Err() != nil {
				results[idx] = parallelResult{
					index: idx,
					result: domain.StepResult{
						StepName:  s.Name,
						Error:     runCtx.Err(),
						Extracted: make(map[string]string),
					},
				}
				return
			}

			sr := r.executeStep(runCtx, s, snapVars)
			results[idx] = parallelResult{index: idx, result: sr}

			// In fail-fast mode, cancel remaining on first error
			if step.FailFast && sr.Error != nil {
				cancel()
			}
		}(i, sub)
	}

	wg.Wait()

	// Collect results in original order and merge extracted variables
	orderedResults := make([]domain.StepResult, len(subSteps))
	mergedVars := make(map[string]string)

	for i, pr := range results {
		orderedResults[i] = pr.result
		for k, v := range pr.result.Extracted {
			mergedVars[k] = v
		}
	}

	return orderedResults, mergedVars, nil
}

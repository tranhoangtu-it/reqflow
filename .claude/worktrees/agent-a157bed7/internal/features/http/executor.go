package http

import (
	"context"
	"fmt"

	"github.com/ye-kart/reqflow/internal/core/auth"
	"github.com/ye-kart/reqflow/internal/core/request"
	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// Executor orchestrates HTTP request execution by tying core logic to adapters.
type Executor struct {
	httpClient driven.HTTPClient
}

// NewExecutor creates a new Executor with the given HTTP client.
func NewExecutor(hc driven.HTTPClient) *Executor {
	return &Executor{httpClient: hc}
}

// Execute builds a request from config, applies auth, sends it, and returns the result.
func (e *Executor) Execute(ctx context.Context, config domain.RequestConfig, vars map[string]string) (domain.ExecutionResult, error) {
	// Build the HTTP request with variable substitution.
	req, err := request.BuildRequest(config, vars)
	if err != nil {
		return domain.ExecutionResult{}, fmt.Errorf("building request: %w", err)
	}

	// Apply authentication if configured.
	if config.Auth != nil {
		req, err = auth.Apply(req, config.Auth)
		if err != nil {
			return domain.ExecutionResult{}, fmt.Errorf("applying auth: %w", err)
		}
	}

	// Send the request via the HTTP client adapter.
	resp, err := e.httpClient.Do(ctx, req)
	if err != nil {
		return domain.ExecutionResult{}, fmt.Errorf("executing request: %w", err)
	}

	return domain.ExecutionResult{
		Request:  req,
		Response: resp,
	}, nil
}

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

// BuildRequest builds a fully resolved HTTP request from config without sending it.
// This applies variable substitution and auth but does not make an HTTP call.
func (e *Executor) BuildRequest(config domain.RequestConfig, vars map[string]string) (domain.HTTPRequest, error) {
	req, err := request.BuildRequest(config, vars)
	if err != nil {
		return domain.HTTPRequest{}, fmt.Errorf("building request: %w", err)
	}

	if config.Auth != nil {
		req, err = auth.Apply(req, config.Auth)
		if err != nil {
			return domain.HTTPRequest{}, fmt.Errorf("applying auth: %w", err)
		}
	}

	return req, nil
}

// Execute builds a request from config, applies auth, sends it, and returns the result.
func (e *Executor) Execute(ctx context.Context, config domain.RequestConfig, vars map[string]string) (domain.ExecutionResult, error) {
	req, err := e.BuildRequest(config, vars)
	if err != nil {
		return domain.ExecutionResult{}, err
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

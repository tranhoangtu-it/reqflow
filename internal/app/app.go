package app

import (
	featurehttp "github.com/ye-kart/reqflow/internal/features/http"
	"github.com/ye-kart/reqflow/internal/features/runner"
	"github.com/ye-kart/reqflow/internal/ports/driven"
)

// App is the main application coordinator that holds all feature executors.
type App struct {
	HTTPExecutor *featurehttp.Executor
	Runner       *runner.Runner
	Storage      driven.Storage
	CookieJar    driven.CookieJar
	httpClient   driven.HTTPClient
}

// New creates a new App wired with the given adapters.
func New(httpClient driven.HTTPClient, storage driven.Storage) *App {
	return &App{
		HTTPExecutor: featurehttp.NewExecutor(httpClient),
		Runner:       runner.New(httpClient),
		Storage:      storage,
		httpClient:   httpClient,
	}
}

// EnableTrace enables detailed timing instrumentation on the HTTP client,
// if the client supports it.
func (a *App) EnableTrace() {
	if ts, ok := a.httpClient.(driven.TraceSetter); ok {
		ts.SetTrace(true)
	}
}

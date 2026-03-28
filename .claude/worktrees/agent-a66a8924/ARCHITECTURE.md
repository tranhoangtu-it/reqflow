# Architecture Design

## Approach: Hybrid Architecture

After analyzing Hexagonal, Clean, Functional Core/Imperative Shell, Vertical Slice, and DDD patterns in the Go context, the recommendation is a **hybrid of three complementary patterns**:

| Pattern | Role | Where It Applies |
|---------|------|-----------------|
| **Hexagonal (Ports & Adapters)** | Primary structure | CLI/TUI as driving adapters, HTTP client/storage/JS engine as driven adapters |
| **Functional Core, Imperative Shell** | Core logic design | Pure functions for request building, variable resolution, assertions |
| **Vertical Slices** | Feature organization | Independent protocol features (HTTP, WebSocket, gRPC, mock server) |

### Why Not a Single Pattern?

- **Hexagonal alone** doesn't guide how to structure business logic internally (pure vs impure).
- **Functional Core alone** doesn't provide the adapter abstraction needed for CLI/TUI swappability.
- **Vertical Slices alone** can lead to duplication without shared ports and core logic.
- **Clean Architecture** adds too many layers (Entities, Use Cases, Gateways, Adapters) for a CLI tool — the overhead isn't justified.
- **Full DDD** is overkill — our domain model is relatively straightforward (requests, responses, assertions).

The hybrid gives us: swappable interfaces (Hexagonal), testable pure logic (Functional Core), and independent features (Vertical Slices).

---

## How It Maps to the Project

```
┌─────────────────────────────────────────────────────────────────┐
│                      DRIVING ADAPTERS                           │
│                                                                 │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐                 │
│   │   CLI    │    │   TUI    │    │  Daemon  │                 │
│   │ (Cobra)  │    │(Bubble   │    │ (mock    │                 │
│   │          │    │  Tea)    │    │  server) │                 │
│   └────┬─────┘    └────┬─────┘    └────┬─────┘                 │
│        │               │               │                        │
├────────┼───────────────┼───────────────┼────────────────────────┤
│        ▼               ▼               ▼                        │
│   ┌─────────────────────────────────────────────┐               │
│   │              DRIVING PORTS                   │               │
│   │  (interfaces that adapters call into)        │               │
│   └──────────────────┬──────────────────────────┘               │
│                      │                                          │
│   ┌──────────────────▼──────────────────────────┐               │
│   │              APP COORDINATOR                 │               │
│   │  (wires core + features + adapters)          │               │
│   └──────────────────┬──────────────────────────┘               │
│                      │                                          │
│   ┌──────────────────▼──────────────────────────┐               │
│   │           FUNCTIONAL CORE (pure)             │               │
│   │                                              │               │
│   │  request/builder    variable/interpolate      │               │
│   │  request/validator  script/parser             │               │
│   │  script/assertion   auth/signer               │               │
│   │                                              │               │
│   │  No side effects. No I/O. Fully testable.    │               │
│   └──────────────────────────────────────────────┘               │
│                                                                 │
│   ┌──────────────────────────────────────────────┐               │
│   │        FEATURE SLICES (vertical)              │               │
│   │                                              │               │
│   │  ┌────────┐ ┌───────────┐ ┌──────┐ ┌──────┐ │               │
│   │  │  HTTP  │ │ WebSocket │ │ gRPC │ │ Mock │ │               │
│   │  └────────┘ └───────────┘ └──────┘ └──────┘ │               │
│   │                                              │               │
│   │  Each feature orchestrates core + adapters   │               │
│   └──────────────────┬──────────────────────────┘               │
│                      │                                          │
│   ┌──────────────────▼──────────────────────────┐               │
│   │              DRIVEN PORTS                    │               │
│   │  (interfaces the core/features depend on)    │               │
│   └──────────────────┬──────────────────────────┘               │
│                      │                                          │
├──────────────────────┼──────────────────────────────────────────┤
│                      ▼                                          │
│   ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│   │  HTTP    │ │   File   │ │    JS    │ │  Logger  │          │
│   │  Client  │ │  Storage │ │  Engine  │ │  (slog)  │          │
│   │(net/http)│ │   (os)   │ │  (Goja)  │ │          │          │
│   └──────────┘ └──────────┘ └──────────┘ └──────────┘          │
│                      DRIVEN ADAPTERS                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Package Layout

```
http-cli/
│
├── cmd/                                # Entry points (one per binary)
│   ├── http-cli/
│   │   └── main.go                    # CLI entry point
│   └── http-cli-tui/
│       └── main.go                    # TUI entry point (future)
│
├── internal/                           # Private application code
│   │
│   ├── core/                          # FUNCTIONAL CORE — pure, no I/O
│   │   ├── request/
│   │   │   ├── builder.go            # Build HTTPRequest from config + variables
│   │   │   ├── builder_test.go
│   │   │   ├── validator.go          # Validate request structure
│   │   │   └── validator_test.go
│   │   │
│   │   ├── variable/
│   │   │   ├── interpolate.go        # {{var}} substitution in strings
│   │   │   ├── interpolate_test.go
│   │   │   ├── resolve.go            # Scope-aware variable resolution
│   │   │   ├── resolve_test.go
│   │   │   └── dynamic.go            # Dynamic variables ($timestamp, $uuid, etc.)
│   │   │
│   │   ├── auth/
│   │   │   ├── basic.go              # Basic auth header computation
│   │   │   ├── bearer.go             # Bearer token header
│   │   │   ├── oauth2.go             # OAuth2 token building (pure parts)
│   │   │   ├── aws.go                # AWS Signature V4 computation
│   │   │   ├── digest.go             # Digest auth computation
│   │   │   └── auth_test.go
│   │   │
│   │   ├── script/
│   │   │   ├── parser.go             # Parse JS test scripts to AST/steps
│   │   │   ├── assertion.go          # Evaluate assertion results
│   │   │   └── assertion_test.go
│   │   │
│   │   ├── collection/
│   │   │   ├── model.go              # Collection, Folder, Request structs
│   │   │   ├── ordering.go           # Request execution ordering logic
│   │   │   └── merge.go              # Merge inherited auth/headers/vars
│   │   │
│   │   ├── importer/
│   │   │   ├── curl.go               # Parse cURL commands to Request
│   │   │   ├── openapi.go            # Parse OpenAPI spec to Collection
│   │   │   ├── postman.go            # Parse Postman collection JSON
│   │   │   └── har.go                # Parse HAR files
│   │   │
│   │   └── exporter/
│   │       ├── curl.go               # Request → cURL command
│   │       ├── openapi.go            # Collection → OpenAPI spec
│   │       ├── markdown.go           # Collection → Markdown docs
│   │       └── codegen.go            # Request → code snippets (Python, JS, Go, etc.)
│   │
│   ├── domain/                        # Domain types and errors
│   │   ├── request.go                # HTTPRequest, HTTPResponse types
│   │   ├── collection.go             # Collection, Folder types
│   │   ├── environment.go            # Environment, Variable types
│   │   ├── script.go                 # Script, Assertion, TestResult types
│   │   ├── config.go                 # App configuration types
│   │   └── errors.go                 # Domain error types
│   │
│   ├── ports/                         # Interface contracts
│   │   │
│   │   ├── driven/                   # Outbound interfaces (core depends on)
│   │   │   ├── httpclient.go         # HTTPClient interface
│   │   │   ├── storage.go            # Storage interface (read/write collections, envs)
│   │   │   ├── scriptengine.go       # ScriptEngine interface (execute JS)
│   │   │   ├── cookiejar.go          # CookieJar interface
│   │   │   └── logger.go             # Logger interface
│   │   │
│   │   └── driving/                  # Inbound interfaces (adapters implement to call app)
│   │       ├── requester.go          # ExecuteRequest, ExecuteCollection
│   │       ├── environments.go       # ManageEnvironments
│   │       ├── collections.go        # ManageCollections
│   │       └── mockserver.go         # StartMockServer, StopMockServer
│   │
│   ├── adapters/                      # IMPERATIVE SHELL — side effects live here
│   │   │
│   │   ├── cli/                      # Driving adapter: CLI
│   │   │   ├── commands/
│   │   │   │   ├── root.go           # Root command, global flags
│   │   │   │   ├── request.go        # `http get/post/put/...` commands
│   │   │   │   ├── run.go            # `http run <collection>` command
│   │   │   │   ├── env.go            # `http env list/set/use` commands
│   │   │   │   ├── collection.go     # `http collection list/import/export`
│   │   │   │   ├── mock.go           # `http mock start/stop`
│   │   │   │   └── config.go         # `http config set/get`
│   │   │   │
│   │   │   ├── output/               # Output formatting strategies
│   │   │   │   ├── formatter.go      # Formatter interface
│   │   │   │   ├── json.go           # --output json
│   │   │   │   ├── table.go          # --output table
│   │   │   │   ├── pretty.go         # Default: colored, formatted
│   │   │   │   ├── minimal.go        # --minimal (status + body only)
│   │   │   │   └── raw.go            # Pipe-friendly raw output
│   │   │   │
│   │   │   └── cli.go                # CLI adapter initialization
│   │   │
│   │   ├── tui/                      # Driving adapter: TUI (future)
│   │   │   ├── app.go                # Bubble Tea application
│   │   │   ├── views/
│   │   │   │   ├── request.go        # Request builder view
│   │   │   │   ├── response.go       # Response viewer
│   │   │   │   ├── collection.go     # Collection browser
│   │   │   │   └── environment.go    # Environment manager
│   │   │   ├── components/
│   │   │   │   ├── editor.go         # Text/JSON editor component
│   │   │   │   ├── list.go           # Selectable list
│   │   │   │   └── tabs.go           # Tab navigation
│   │   │   └── tui.go
│   │   │
│   │   ├── httpclient/               # Driven adapter: HTTP
│   │   │   ├── client.go             # net/http implementation
│   │   │   ├── transport.go          # Custom transport (proxy, certs, timing)
│   │   │   └── mock.go               # Mock for testing
│   │   │
│   │   ├── storage/                  # Driven adapter: File system
│   │   │   ├── filesystem.go         # Real FS implementation
│   │   │   └── memory.go             # In-memory for testing
│   │   │
│   │   ├── engine/                   # Driven adapter: JS engine
│   │   │   ├── goja.go               # Goja implementation
│   │   │   ├── sandbox.go            # pm.* API bindings for scripts
│   │   │   └── mock.go               # Mock for testing
│   │   │
│   │   ├── cookiejar/                # Driven adapter: Cookie storage
│   │   │   ├── jar.go                # Persistent cookie jar
│   │   │   └── mock.go
│   │   │
│   │   └── logger/                   # Driven adapter: Logging
│   │       └── slog.go               # slog-based structured logging
│   │
│   ├── features/                      # VERTICAL SLICES — protocol-specific orchestration
│   │   │
│   │   ├── http/                     # HTTP protocol feature
│   │   │   ├── executor.go           # Orchestrates: core → adapter → assertion
│   │   │   ├── executor_test.go
│   │   │   └── model.go              # HTTP-specific DTOs
│   │   │
│   │   ├── websocket/                # WebSocket protocol feature (future)
│   │   │   ├── client.go
│   │   │   ├── client_test.go
│   │   │   └── model.go
│   │   │
│   │   ├── grpc/                     # gRPC protocol feature (future)
│   │   │   ├── caller.go
│   │   │   ├── caller_test.go
│   │   │   ├── reflection.go
│   │   │   └── model.go
│   │   │
│   │   ├── mockserver/               # Mock server feature
│   │   │   ├── server.go
│   │   │   ├── router.go             # Request matching logic
│   │   │   ├── router_test.go
│   │   │   └── model.go
│   │   │
│   │   ├── runner/                   # Collection runner feature
│   │   │   ├── runner.go             # Sequential/parallel execution
│   │   │   ├── runner_test.go
│   │   │   ├── datafile.go           # CSV/JSON data file iteration
│   │   │   └── report.go             # Run results and reporting
│   │   │
│   │   ├── monitor/                  # Monitoring feature (future)
│   │   │   ├── scheduler.go
│   │   │   └── model.go
│   │   │
│   │   └── loadtest/                 # Performance testing feature (future)
│   │       ├── loadtest.go
│   │       ├── profile.go            # Load profiles (fixed, ramp, spike)
│   │       └── metrics.go            # P50/P90/P95/P99 calculation
│   │
│   ├── platform/                      # Cross-cutting shared utilities
│   │   ├── config/
│   │   │   ├── loader.go             # Load config from file, env, flags
│   │   │   └── defaults.go           # Default values
│   │   │
│   │   └── errors/
│   │       └── errors.go             # Shared error types and wrapping
│   │
│   └── app/                           # Application wiring
│       ├── app.go                    # Constructs all dependencies, wires everything
│       └── app_test.go               # Integration tests
│
├── pkg/                               # Public API (for plugins, external tools)
│   └── plugin/
│       ├── api.go                    # Plugin interfaces
│       └── registry.go               # Plugin registration
│
├── test/                              # E2E and integration test fixtures
│   ├── e2e/
│   │   └── cli_test.go
│   └── fixtures/
│       ├── collections/
│       ├── environments/
│       └── scripts/
│
├── FEATURES.md
├── TECH_STACK.md
├── ARCHITECTURE.md
├── go.mod
└── go.sum
```

---

## Core Design Principles

### 1. Pure Core, Impure Shell

The `internal/core/` package contains **only pure functions**: no I/O, no network calls, no file reads, no logging side effects. Given the same input, they always produce the same output.

```go
// internal/core/request/builder.go — PURE
package request

import "github.com/you/http-cli/internal/domain"

// BuildRequest is a pure function. No side effects.
// Same inputs → same output, every time.
func BuildRequest(
    config domain.RequestConfig,
    vars map[string]string,
) (domain.HTTPRequest, error) {
    url := interpolate(config.URL, vars)
    headers := interpolateHeaders(config.Headers, vars)
    body := interpolateBody(config.Body, vars)

    if err := validate(url, config.Method); err != nil {
        return domain.HTTPRequest{}, err
    }

    return domain.HTTPRequest{
        Method:  config.Method,
        URL:     url,
        Headers: headers,
        Body:    body,
    }, nil
}
```

```go
// internal/core/script/assertion.go — PURE
package script

import "github.com/you/http-cli/internal/domain"

// EvaluateAssertions is pure. Takes data in, returns results out.
func EvaluateAssertions(
    assertions []domain.Assertion,
    resp domain.HTTPResponse,
) []domain.TestResult {
    results := make([]domain.TestResult, len(assertions))
    for i, a := range assertions {
        results[i] = evaluate(a, resp)
    }
    return results
}

func evaluate(a domain.Assertion, resp domain.HTTPResponse) domain.TestResult {
    switch a.Type {
    case domain.AssertStatus:
        return domain.TestResult{
            Name:   a.Name,
            Passed: resp.Status == a.Expected.(int),
        }
    case domain.AssertBodyContains:
        return domain.TestResult{
            Name:   a.Name,
            Passed: strings.Contains(string(resp.Body), a.Expected.(string)),
        }
    // ...
    }
}
```

**Testing pure functions requires zero mocks:**

```go
func TestBuildRequest(t *testing.T) {
    tests := []struct {
        name   string
        config domain.RequestConfig
        vars   map[string]string
        want   domain.HTTPRequest
    }{
        {
            name: "substitutes variables in URL",
            config: domain.RequestConfig{
                Method: "GET",
                URL:    "https://{{host}}/api/{{version}}/users",
            },
            vars: map[string]string{"host": "example.com", "version": "v2"},
            want: domain.HTTPRequest{
                Method: "GET",
                URL:    "https://example.com/api/v2/users",
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := BuildRequest(tt.config, tt.vars)
            assert.NoError(t, err)
            assert.Equal(t, tt.want.URL, got.URL)
        })
    }
}
```

---

### 2. Ports as Contracts

Ports are Go interfaces. Small, focused, following Go's interface segregation idiom.

```go
// internal/ports/driven/httpclient.go
package driven

import (
    "context"
    "github.com/you/http-cli/internal/domain"
)

// HTTPClient is the driven port for making HTTP requests.
// Implemented by adapters/httpclient and adapters/httpclient/mock.
type HTTPClient interface {
    Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error)
}
```

```go
// internal/ports/driven/storage.go
package driven

// Storage is the driven port for persistence.
type Storage interface {
    ReadCollection(path string) (domain.Collection, error)
    WriteCollection(path string, c domain.Collection) error
    ReadEnvironment(path string) (domain.Environment, error)
    WriteEnvironment(path string, e domain.Environment) error
}
```

```go
// internal/ports/driven/scriptengine.go
package driven

import "github.com/you/http-cli/internal/domain"

// ScriptEngine is the driven port for JS script execution.
type ScriptEngine interface {
    Execute(script string, ctx domain.ScriptContext) (domain.ScriptResult, error)
}
```

```go
// internal/ports/driving/requester.go
package driving

import (
    "context"
    "github.com/you/http-cli/internal/domain"
)

// Requester is the driving port that CLI and TUI call into.
type Requester interface {
    Execute(ctx context.Context, req domain.RequestConfig, env domain.Environment) (domain.ExecutionResult, error)
    RunCollection(ctx context.Context, opts domain.RunOptions) (domain.RunReport, error)
}
```

---

### 3. Adapters Are Thin

Adapters translate between external systems and domain types. No business logic.

```go
// internal/adapters/httpclient/client.go
package httpclient

import (
    "context"
    "io"
    "net/http"
    "github.com/you/http-cli/internal/domain"
)

type Client struct {
    http *http.Client
}

func New(opts ...Option) *Client {
    c := &Client{http: &http.Client{}}
    for _, opt := range opts {
        opt(c)
    }
    return c
}

// Do translates domain.HTTPRequest → net/http → domain.HTTPResponse.
// No business logic — just translation.
func (c *Client) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
    httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
    if err != nil {
        return domain.HTTPResponse{}, err
    }
    for k, v := range req.Headers {
        httpReq.Header.Set(k, v)
    }

    resp, err := c.http.Do(httpReq)
    if err != nil {
        return domain.HTTPResponse{}, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return domain.HTTPResponse{}, err
    }

    return domain.HTTPResponse{
        Status:     resp.StatusCode,
        Headers:    resp.Header,
        Body:       body,
        Duration:   /* timing from transport */,
    }, nil
}
```

---

### 4. Features Orchestrate Core + Adapters

Feature slices are the **imperative shell** — they coordinate pure core functions with impure adapter calls.

```go
// internal/features/http/executor.go
package http

import (
    "context"
    "github.com/you/http-cli/internal/core/request"
    "github.com/you/http-cli/internal/core/script"
    "github.com/you/http-cli/internal/core/variable"
    "github.com/you/http-cli/internal/domain"
    "github.com/you/http-cli/internal/ports/driven"
)

type Executor struct {
    httpClient   driven.HTTPClient
    scriptEngine driven.ScriptEngine
    logger       driven.Logger
}

func NewExecutor(hc driven.HTTPClient, se driven.ScriptEngine, l driven.Logger) *Executor {
    return &Executor{httpClient: hc, scriptEngine: se, logger: l}
}

func (e *Executor) Execute(
    ctx context.Context,
    config domain.RequestConfig,
    env domain.Environment,
) (domain.ExecutionResult, error) {

    // PURE: resolve variables
    vars := variable.Resolve(env.Variables, config.CollectionVars)

    // IMPURE: run pre-request script (if any)
    if config.PreRequestScript != "" {
        scriptCtx := domain.ScriptContext{Variables: vars, Request: config}
        result, err := e.scriptEngine.Execute(config.PreRequestScript, scriptCtx)
        if err != nil {
            return domain.ExecutionResult{}, err
        }
        vars = result.UpdatedVariables
        config = result.UpdatedRequest
    }

    // PURE: build the request
    req, err := request.BuildRequest(config, vars)
    if err != nil {
        return domain.ExecutionResult{}, err
    }

    // IMPURE: send the HTTP request
    resp, err := e.httpClient.Do(ctx, req)
    if err != nil {
        return domain.ExecutionResult{}, err
    }

    // PURE: evaluate assertions
    testResults := script.EvaluateAssertions(config.Assertions, resp)

    // IMPURE: run post-response script (if any)
    if config.PostResponseScript != "" {
        scriptCtx := domain.ScriptContext{Variables: vars, Response: resp}
        _, _ = e.scriptEngine.Execute(config.PostResponseScript, scriptCtx)
    }

    return domain.ExecutionResult{
        Request:     req,
        Response:    resp,
        TestResults: testResults,
    }, nil
}
```

Notice the alternating rhythm: **pure → impure → pure → impure → pure**. This is the Functional Core / Imperative Shell pattern in action within the Hexagonal structure.

---

### 5. CLI and TUI Are Interchangeable Driving Adapters

Both CLI and TUI call the same driving ports. They differ only in how they present input/output.

```go
// internal/adapters/cli/commands/request.go
package commands

import (
    "github.com/spf13/cobra"
    "github.com/you/http-cli/internal/ports/driving"
)

func NewRequestCmd(requester driving.Requester) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "get <url>",
        Short: "Send a GET request",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse flags into domain.RequestConfig
            config := parseFlags(cmd, args)
            env := loadActiveEnvironment(cmd)

            // Call the SAME driving port that TUI would call
            result, err := requester.Execute(cmd.Context(), config, env)
            if err != nil {
                return err
            }

            // Format output for terminal
            formatter := selectFormatter(cmd)
            return formatter.Format(os.Stdout, result)
        },
    }
    return cmd
}
```

```go
// internal/adapters/tui/views/request.go (future)
package views

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/you/http-cli/internal/ports/driving"
)

type RequestView struct {
    requester driving.Requester  // SAME interface as CLI uses
    // ... TUI state
}

func (v RequestView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case SendRequestMsg:
        return v, func() tea.Msg {
            // Call the SAME driving port
            result, err := v.requester.Execute(context.Background(), msg.Config, msg.Env)
            return ResponseMsg{Result: result, Err: err}
        }
    }
    // ...
}
```

---

### 6. Application Wiring

One place constructs all dependencies and wires everything together.

```go
// internal/app/app.go
package app

import (
    "github.com/you/http-cli/internal/adapters/httpclient"
    "github.com/you/http-cli/internal/adapters/storage"
    "github.com/you/http-cli/internal/adapters/engine"
    "github.com/you/http-cli/internal/adapters/logger"
    "github.com/you/http-cli/internal/domain"
    httpfeature "github.com/you/http-cli/internal/features/http"
    "github.com/you/http-cli/internal/features/runner"
)

type App struct {
    HTTPExecutor    *httpfeature.Executor
    CollectionRunner *runner.Runner
    Config          *domain.AppConfig
}

func New(cfg *domain.AppConfig) (*App, error) {
    // Construct driven adapters
    httpClient := httpclient.New(
        httpclient.WithTimeout(cfg.Timeout),
        httpclient.WithProxy(cfg.Proxy),
    )
    fs := storage.NewFilesystem(cfg.DataDir)
    jsEngine := engine.NewGoja()
    log := logger.NewSlog(cfg.LogLevel)

    // Construct features (inject driven adapters)
    httpExec := httpfeature.NewExecutor(httpClient, jsEngine, log)
    collRunner := runner.NewRunner(httpExec, fs, log)

    return &App{
        HTTPExecutor:     httpExec,
        CollectionRunner: collRunner,
        Config:           cfg,
    }, nil
}
```

```go
// cmd/http-cli/main.go
package main

import (
    "os"
    "github.com/you/http-cli/internal/app"
    "github.com/you/http-cli/internal/adapters/cli"
    "github.com/you/http-cli/internal/platform/config"
)

func main() {
    cfg := config.Load()
    application, err := app.New(cfg)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    cliAdapter := cli.New(application)
    if err := cliAdapter.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

## Dependency Flow

```
cmd/http-cli/main.go
    │
    ▼
internal/app/app.go          ← Wires everything (composition root)
    │
    ├──► internal/adapters/*  ← Implements port interfaces (impure)
    │
    ├──► internal/features/*  ← Orchestrates core + adapters (imperative shell)
    │       │
    │       ├──► internal/core/*   ← Pure business logic (functional core)
    │       │
    │       └──► internal/ports/*  ← Depends on interfaces only
    │
    └──► internal/domain/*    ← Shared types (no dependencies)
```

**Dependency rule**: Everything points inward. `domain` depends on nothing. `core` depends only on `domain`. `ports` depends only on `domain`. `features` depends on `core` + `ports`. `adapters` implement `ports`. `app` wires it all. `cmd` calls `app`.

---

## Dependency Injection Strategy

Go doesn't need a DI framework. Constructor injection with functional options is idiomatic and sufficient.

### Required Dependencies → Constructor Parameters

```go
func NewExecutor(hc driven.HTTPClient, se driven.ScriptEngine, l driven.Logger) *Executor {
    return &Executor{httpClient: hc, scriptEngine: se, logger: l}
}
```

### Optional Configuration → Functional Options

```go
type Option func(*Client)

func WithTimeout(d time.Duration) Option {
    return func(c *Client) { c.timeout = d }
}

func WithProxy(proxy string) Option {
    return func(c *Client) { c.proxy = proxy }
}

func New(opts ...Option) *Client {
    c := &Client{timeout: 30 * time.Second}  // defaults
    for _, opt := range opts {
        opt(c)
    }
    return c
}
```

---

## Testing Strategy

| Layer | Test Type | Mocks Needed | Example |
|-------|-----------|-------------|---------|
| `core/` | Unit (table-driven) | **None** — pure functions | `TestBuildRequest`, `TestInterpolate`, `TestEvaluateAssertions` |
| `features/` | Integration | Mock ports (driven adapters) | `TestHTTPExecutor` with `MockHTTPClient` |
| `adapters/` | Integration | External systems (httptest, temp dirs) | `TestFilesystemStorage` with temp directory |
| `app/` | Integration | Mock adapters | `TestAppWiring` |
| `cmd/` | E2E | Real binary execution | `TestCLIGetRequest` running the compiled binary |

### Mock Pattern

```go
// Mock directly in test file — no framework needed
type mockHTTPClient struct {
    doFunc func(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error)
}

func (m *mockHTTPClient) Do(ctx context.Context, req domain.HTTPRequest) (domain.HTTPResponse, error) {
    return m.doFunc(ctx, req)
}

func TestExecutor_Execute(t *testing.T) {
    mock := &mockHTTPClient{
        doFunc: func(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
            return domain.HTTPResponse{Status: 200, Body: []byte(`{"ok":true}`)}, nil
        },
    }
    exec := NewExecutor(mock, &mockScriptEngine{}, &mockLogger{})

    result, err := exec.Execute(context.Background(), config, env)
    assert.NoError(t, err)
    assert.Equal(t, 200, result.Response.Status)
}
```

---

## Why Go Naturally Supports This

| Go Feature | Architectural Benefit |
|---|---|
| **Implicit interfaces** | Ports don't need explicit `implements` — any struct with matching methods satisfies the interface |
| **`internal/` package** | Compiler-enforced encapsulation — adapters can't be imported from outside the module |
| **Value types (structs)** | Domain types are naturally immutable when passed by value — supports functional core |
| **First-class functions** | Functional options, middleware, and pure function composition |
| **No inheritance** | Composition over inheritance is enforced — adapters compose, not extend |
| **`go test` built-in** | Table-driven tests with subtests are idiomatic — no test framework needed |
| **Fast compilation** | Rapid feedback loop when developing across layers |

---

## Adding the TUI Later

When the TUI is added, the changes are minimal:

1. **Create `internal/adapters/tui/`** — a new driving adapter using Bubble Tea
2. **Create `cmd/http-cli-tui/main.go`** — new entry point that wires `app.New()` to the TUI adapter
3. **Zero changes to core, features, ports, or driven adapters**

The TUI calls the same `driving.Requester` interface. It just presents input (forms, editors) and output (panels, syntax highlighting) differently.

```
# Before (CLI only)
cmd/http-cli/main.go → app.New() → cli.New(app) → same core

# After (CLI + TUI)
cmd/http-cli/main.go     → app.New() → cli.New(app) → same core
cmd/http-cli-tui/main.go → app.New() → tui.New(app) → same core
```

---

## Adding a New Protocol Later

Example: adding WebSocket support.

1. **`internal/features/websocket/`** — new vertical slice with its own orchestration
2. **`internal/ports/driven/wsclient.go`** — new driven port interface (if needed)
3. **`internal/adapters/wsclient/`** — gorilla/websocket adapter
4. **`internal/adapters/cli/commands/ws.go`** — new `http ws connect <url>` command
5. **Zero changes to HTTP feature, runner, or existing adapters**

---

## Adding Plugins Later

1. **`pkg/plugin/api.go`** — public interfaces that plugins implement
2. Plugins are Go packages imported at compile time (or loaded via RPC for external plugins)
3. Each plugin registers as an adapter for a specific port
4. Core and features remain untouched

```go
// pkg/plugin/api.go
package plugin

type Protocol interface {
    Name() string
    Execute(ctx context.Context, config map[string]any) (map[string]any, error)
}

type AuthProvider interface {
    Name() string
    Sign(req map[string]string) (map[string]string, error)
}
```

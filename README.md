# reqflow

> **Warning**
> This project is in early development and is **not ready for production use**. APIs, commands, and file formats are subject to breaking changes without notice. Expect missing features, incomplete documentation, and untested edge cases. Contributions and feedback are welcome — see the [Roadmap](#roadmap) for current status.

[![Status: Experimental](https://img.shields.io/badge/status-experimental-orange)](https://github.com/ye-kart/reqflow)
[![Go](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A powerful HTTP client for the terminal. Combines curl's simplicity with Postman's organization, adds async workflow orchestration, and is designed from the ground up for both humans and AI agents.

```
reqflow get https://api.example.com/users -o pretty
```

```
HTTP/1.1 200 OK
Content-Type: application/json

[
  {"id": 1, "name": "Alice"},
  {"id": 2, "name": "Bob"}
]

(took 142ms)
```

## Why reqflow?

**curl** is stateless. Every command is isolated — chaining requests, managing auth tokens, and validating responses requires bash gymnastics. Output is unstructured and hostile to automation.

**Postman** solves organization but locks you into a GUI, a cloud account, and an Electron app consuming 500MB+ of RAM.

**Neither handles async workflows.** Modern APIs use background jobs, webhooks, eventual consistency, and multi-step orchestrations. Testing these from the CLI means writing throwaway scripts every time.

reqflow fills the gap:

- **Instant requests** with better defaults than curl
- **Collections and environments** stored as plain files you can version control
- **Async workflow engine** with polling, retries, webhooks, and parallel execution
- **Structured output** that AI agents can parse and reason about

## Install

### From source

```sh
go install github.com/ye-kart/reqflow/cmd/reqflow@latest
```

### Build locally

```sh
git clone https://github.com/ye-kart/reqflow.git
cd reqflow
make build
# Binary is at ./bin/reqflow
```

## Quick start

### Simple requests

```sh
# GET request
reqflow get https://httpbin.org/get

# POST with JSON body
reqflow post https://httpbin.org/post -d '{"name": "reqflow", "type": "cli"}'

# PUT with explicit content type
reqflow put https://api.example.com/users/1 -d '{"name": "Alice"}' --content-type application/json

# DELETE
reqflow delete https://api.example.com/users/1
```

### Headers and query parameters

```sh
# Add headers (repeatable)
reqflow get https://api.example.com/data \
  -H "Accept: application/json" \
  -H "X-Request-ID: abc-123"

# Add query parameters (repeatable)
reqflow get https://api.example.com/users \
  -q "page=2" \
  -q "limit=50"
```

### Authentication

```sh
# Basic auth
reqflow get https://api.example.com/me --auth-basic "alice:secretpass"

# Bearer token
reqflow get https://api.example.com/data --auth-bearer "eyJhbGciOiJIUzI1NiIs..."

# API key in header
reqflow get https://api.example.com/data --auth-apikey-header "X-API-Key:your-key-here"

# API key in query parameter
reqflow get https://api.example.com/data --auth-apikey-query "api_key=your-key-here"
```

### Output formats

```sh
# Pretty (default) — colored, formatted, with timing
reqflow get https://httpbin.org/get -o pretty

# JSON — structured, machine-parseable
reqflow get https://httpbin.org/get -o json

# Raw — response body only, ideal for piping
reqflow get https://httpbin.org/get -o raw | jq '.headers'

# Minimal — status code + body
reqflow get https://httpbin.org/get -o minimal
```

## Use cases

### Working with AI agents

AI coding agents (Claude, Copilot, etc.) need tools that accept declarative input and return structured output. reqflow is designed for this:

```sh
# Structured JSON output for agent consumption
reqflow post https://api.example.com/users \
  -d '{"name": "test"}' \
  -o json

# Raw output for piping into other tools
reqflow get https://api.example.com/config -o raw | jq '.database.host'
```

The `--output json` mode returns a consistent schema with status, headers, body, timing, and size — everything an agent needs to reason about the response without parsing human-formatted text.

### Testing async API workflows (planned)

Submit a job, poll until it completes, verify the result — all in a declarative YAML file:

```yaml
# workflow.yaml
steps:
  - name: submit-job
    method: POST
    url: "{{base_url}}/jobs"
    body:
      type: "export"
      format: "csv"
    extract:
      job_id: "$.id"
    assert:
      - status: 202

  - name: wait-for-completion
    method: GET
    url: "{{base_url}}/jobs/{{job_id}}"
    poll:
      interval: 2s
      timeout: 60s
      until: "$.status == 'completed'"

  - name: download-result
    method: GET
    url: "{{base_url}}/jobs/{{job_id}}/result"
    assert:
      - status: 200
      - body.format: "csv"
```

```sh
reqflow run workflow.yaml -e production
```

### Replacing complex shell scripts

Instead of this:

```sh
# 20 lines of bash with curl, jq, sleep, and temporary files
TOKEN=$(curl -s -X POST https://auth.example.com/token \
  -d "client_id=$ID&client_secret=$SECRET" | jq -r '.access_token')
RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" \
  https://api.example.com/data)
echo "$RESPONSE" | jq '.items[] | select(.active == true)'
```

Write a workflow once, reuse it across environments:

```yaml
steps:
  - name: authenticate
    method: POST
    url: "{{auth_url}}/token"
    body:
      client_id: "{{client_id}}"
      client_secret: "{{client_secret}}"
    extract:
      token: "$.access_token"

  - name: fetch-data
    method: GET
    url: "{{api_url}}/data"
    headers:
      Authorization: "Bearer {{token}}"
    assert:
      - status: 200
```

### API testing in CI/CD

reqflow collections are plain files. Commit them to your repo, run them in CI:

```sh
# Run a collection against staging
reqflow run tests/api-smoke.yaml -e staging

# Exit code tells CI if tests passed
echo $?  # 0 = all passed, 4 = assertion failure
```

## CLI reference

```
Usage:
  reqflow [command]

Commands:
  get         Send a GET request
  post        Send a POST request
  put         Send a PUT request
  patch       Send a PATCH request
  delete      Send a DELETE request

Global Flags:
  -o, --output string       Output format: pretty, json, raw, minimal (default "pretty")
      --no-color            Disable colored output
  -t, --timeout duration    Request timeout (default 30s)
  -v, --verbose             Show request/response details
  -H, --header strings      Add headers (format "Key: Value", repeatable)
  -q, --query strings       Add query params (format "key=value", repeatable)

Body Flags (post, put, patch):
  -d, --data string         Request body
      --content-type string Content-Type header

Auth Flags:
      --auth-basic string           Basic auth (username:password)
      --auth-bearer string          Bearer token
      --auth-apikey-header string   API key in header (HeaderName:Value)
      --auth-apikey-query string    API key as query param (paramName=Value)
```

### Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success (2xx response) |
| 1 | HTTP error (4xx/5xx) or general error |

Full exit code semantics (planned):

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | HTTP error (4xx/5xx response) |
| 2 | Network/connection error |
| 3 | Timeout |
| 4 | Assertion failure |
| 5 | Configuration/validation error |
| 6 | Workflow step failure |

## Architecture

reqflow uses a hybrid architecture designed for testability and extensibility:

- **Hexagonal (Ports & Adapters)** — CLI and TUI (future) are interchangeable driving adapters. HTTP client, file storage, and script engine are driven adapters behind interfaces.
- **Functional Core / Imperative Shell** — Pure functions for request building, variable interpolation, and auth computation. All side effects (network, file I/O) are pushed to the shell layer.
- **Vertical Slices** — Each protocol (HTTP, WebSocket, gRPC) is an independent feature slice.

```
cmd/reqflow/           Entry point
internal/
  app/                 Dependency wiring
  core/                Pure business logic (no I/O)
    request/           Request building and validation
    variable/          Variable interpolation and resolution
    auth/              Authentication computation
  features/            Feature orchestration (imperative shell)
    http/              HTTP request execution
  adapters/            External system adapters
    cli/               Cobra commands + output formatters
    httpclient/        net/http wrapper
  ports/               Interface contracts
  domain/              Shared types
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design document.

## Roadmap

### Phase 1 — MVP (current)
- [x] HTTP methods (GET, POST, PUT, PATCH, DELETE)
- [x] Headers and query parameters
- [x] Basic, Bearer, and API Key authentication
- [x] Output formats (pretty, JSON, raw, minimal)
- [x] Variable substitution in URLs, headers, and query params
- [x] Request timeout configuration
- [ ] Environment file loading
- [ ] cURL import/export
- [ ] Config file support
- [ ] Verbose mode (request details)
- [ ] Shell completions

### Phase 2 — Async Workflows & AI Agent Support
- [ ] Multi-step workflow engine (YAML)
- [ ] Poll-until-ready with interval and timeout
- [ ] Retry with backoff (fixed, exponential)
- [ ] Parallel step execution with dependencies
- [ ] Webhook listener for async callbacks
- [ ] Variable extraction and chaining between steps
- [ ] Structured exit codes
- [ ] `--extract` and `--assert` inline flags

### Phase 3 — Collections & Scripting
- [ ] Collection and folder management
- [ ] Collection runner
- [ ] Pre-request and post-response JavaScript scripts
- [ ] Test assertions with reporting
- [ ] Data-driven testing (CSV/JSON)
- [ ] OAuth 2.0, Digest, AWS Signature auth
- [ ] Cookie management

### Phase 4 — Advanced
- [ ] Mock servers
- [ ] Monitoring and scheduled runs
- [ ] WebSocket, GraphQL, gRPC support
- [ ] Performance/load testing
- [ ] API documentation generation
- [ ] Interactive TUI mode

### Phase 5 — Ecosystem
- [ ] Plugin/extension system
- [ ] Full import/export (Postman, OpenAPI, Swagger, HAR, Insomnia)
- [ ] Response history and comparison
- [ ] Code generation from requests

See [FEATURES.md](FEATURES.md) for the complete feature specification.

## Project documentation

| Document | Description |
|----------|-------------|
| [VISION.md](VISION.md) | Why reqflow exists, who it's for, design principles |
| [FEATURES.md](FEATURES.md) | Complete feature specification across 21 categories |
| [TECH_STACK.md](TECH_STACK.md) | Technology comparison (Rust vs Go vs TypeScript) and library choices |
| [ARCHITECTURE.md](ARCHITECTURE.md) | Hybrid architecture design with code examples |

## Development

### Prerequisites
- Go 1.22+

### Build and test

```sh
make build         # Build binary to ./bin/reqflow
make test          # Run all tests
make test-verbose  # Run tests with verbose output
make test-cover    # Generate coverage report
make lint          # Run linter (requires golangci-lint)
make clean         # Remove build artifacts
```

### Project structure

The project follows Go conventions with the [Hexagonal Architecture](ARCHITECTURE.md) pattern:

- `cmd/` — Binary entry points
- `internal/core/` — Pure business logic (zero dependencies, fully testable)
- `internal/adapters/` — External system integrations
- `internal/features/` — Feature orchestration
- `internal/domain/` — Shared domain types
- `internal/ports/` — Interface definitions

## Contributing

reqflow is in active early development. The architecture is designed to make contributing straightforward:

1. Core logic in `internal/core/` is pure functions — easy to understand and test
2. New features go in `internal/features/` as vertical slices
3. New protocols (WebSocket, gRPC) are independent adapters
4. All development follows Red-Green-Refactor TDD

## License

MIT

# reqflow — Vision

> curl's power + Postman's organization + async workflow orchestration, designed for both humans and AI agents.

---

## The Problem

### curl is not enough

curl is a brilliant tool for sending a single HTTP request. But modern API development demands more:

- **Stateless by design** — every command is isolated. Chaining requests (create user → get token → call API) requires manual bash scripting, temporary files, and `jq` gymnastics.
- **No workflow concept** — there's no way to express "send this, wait for that, then verify this."
- **No assertions** — you can see the response, but validating it programmatically means writing wrapper scripts.
- **No organization** — 50 curl commands in a README don't scale. No collections, no environments, no variable management.
- **Hostile to automation** — inconsistent output formats, no structured errors, no meaningful exit codes for different failure modes.

### Postman is too much (and too closed)

Postman solves the organization problem but creates new ones:

- **GUI-only workflow** — can't be used in CI/CD, can't be driven by AI agents, can't be composed with unix tools.
- **Cloud-dependent** — collections live in Postman's cloud. Data leaves your machine.
- **Heavy** — Electron app consuming 500MB+ RAM to send a GET request.
- **No async workflows** — can't poll for completion, can't listen for webhooks, can't orchestrate multi-step async processes.

### Nothing handles async workflows from the CLI

Modern APIs are increasingly asynchronous:
- Submit a job → poll for completion → fetch results
- Trigger an export → receive a webhook callback → download the file
- Create resources in parallel → combine them in a follow-up request
- Call an eventually-consistent endpoint → retry until the data propagates

Today, testing these workflows means writing custom scripts every time. There's no declarative, reusable way to express and run them.

---

## The Solution

**reqflow** is a CLI tool that combines:

1. **Instant HTTP requests** — as easy as curl, with better defaults and colored output
2. **Organized collections** — Postman-style collections, environments, and variables, stored as plain files
3. **Async workflow engine** — declarative YAML workflows with polling, webhooks, retries, parallelism, and step chaining
4. **AI-agent-first design** — structured JSON output, predictable exit codes, declarative configs that agents can read and write

---

## Who Is This For

### Developers working with AI agents

AI coding agents (Claude, Copilot, etc.) are increasingly involved in API development and debugging. They need tools that:
- Accept declarative input (YAML/JSON files, not long flag-heavy commands)
- Return structured, parseable output (not human-formatted text)
- Have predictable behavior (consistent exit codes, no interactive prompts)
- Store state as files (collections, environments) that agents can inspect and modify

reqflow is designed to be the HTTP tool that AI agents reach for.

### Backend developers testing async systems

If your system involves queues, background jobs, webhooks, eventual consistency, or multi-service orchestration, you need more than "send request, check response." You need workflows.

### Teams who want Postman's power without Postman's weight

Collections and environments as version-controlled files. No cloud account. No Electron app. Works in CI/CD natively.

---

## Design Principles

### 1. Files over cloud
Collections, environments, workflows — everything is a plain YAML/JSON file. Version control them. Review them in PRs. Share them via Git.

### 2. Composable with unix
reqflow respects stdin/stdout, supports piping, uses meaningful exit codes, and auto-detects TTY vs pipe for output formatting.

### 3. Progressive complexity
`reqflow get https://api.example.com` works immediately. Collections, workflows, and scripting are there when you need them.

### 4. AI-native
Structured output, declarative input, predictable behavior. An AI agent should be able to use reqflow as effectively as a human.

### 5. Async-first workflows
Polling, webhooks, retries, parallelism, and step dependencies are first-class concepts, not afterthoughts bolted on with shell scripts.

---

## What reqflow Is NOT

- **Not a GUI** — it's a CLI tool (with a TUI planned for interactive exploration)
- **Not a load testing framework** — it has basic performance testing, but k6/locust are better for serious load testing
- **Not a monitoring platform** — it can run scheduled checks, but Datadog/Grafana are for production monitoring
- **Not trying to replace curl** — for a single, quick request, curl is fine. reqflow shines when you need more.

---

## Competitive Landscape

| Tool | Strengths | What reqflow adds |
|------|-----------|-------------------|
| **curl** | Universal, powerful, scriptable | Collections, variables, assertions, workflows, structured output |
| **httpie** | Beautiful output, intuitive syntax | Collections, workflows, scripting, async support |
| **Postman** | Full-featured, GUI, collaboration | CLI-native, async workflows, AI-agent support, no cloud dependency |
| **Newman** | Runs Postman collections in CLI | Native CLI design (not a Postman addon), async workflows, better output |
| **Insomnia** | Clean GUI, Git sync | CLI-native, async workflows, lighter weight |
| **hurl** | Declarative HTTP testing | Async workflows, variable scoping, richer scripting, collections |
| **Bruno** | Git-friendly, offline API client | CLI-first, async workflows, AI-agent design |

---

## Success Metrics

reqflow succeeds when:

1. An AI agent can test a 5-step async API workflow by generating a YAML file and running `reqflow run`
2. A developer can replace a 50-line bash script (curl + jq + sleep + polling) with a 15-line workflow file
3. A team can share API collections via Git and run them in CI with zero configuration
4. A single `reqflow` command gives you what would take curl + jq + grep + awk + sleep

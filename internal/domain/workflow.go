package domain

import "time"

// Workflow represents a named sequence of HTTP steps to execute.
type Workflow struct {
	Name  string
	Env   string // environment name
	Steps []Step
}

// Step represents a single HTTP request within a workflow.
// A step may contain a Parallel block of sub-steps that execute concurrently.
type Step struct {
	Name        string
	Method      HTTPMethod
	URL         string
	Headers     map[string]string
	Body        interface{}        // can be string or map, marshaled to JSON
	Extract     map[string]string  // varName -> JSONPath expression
	Assert      []Assertion
	Auth        *AuthConfig
	ContentType string
	Parallel    []Step // sub-steps to run in parallel
	MaxParallel int    // concurrency limit (0 = unlimited)
	FailFast    bool   // cancel remaining on first failure (default true)
}

// Assertion defines an expected condition on a response.
type Assertion struct {
	Field    string      // "status", "body.field.path", "header.Name"
	Operator string      // "==", "!=", "<", ">", "contains", "exists"
	Expected interface{} // expected value to compare against
}

// StepResult holds the outcome of executing a single workflow step.
type StepResult struct {
	StepName   string
	Request    HTTPRequest
	Response   HTTPResponse
	Assertions []AssertionResult
	Extracted  map[string]string
	Error      error
	Duration   time.Duration
}

// AssertionResult holds the outcome of evaluating a single assertion.
type AssertionResult struct {
	Assertion Assertion
	Passed    bool
	Actual    interface{}
	Message   string
}

// WorkflowResult holds the aggregate outcome of executing a workflow.
type WorkflowResult struct {
	Name        string
	Steps       []StepResult
	TotalPassed int
	TotalFailed int
	Duration    time.Duration
}

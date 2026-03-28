package domain

import "time"

// Workflow represents a named sequence of HTTP steps to execute.
type Workflow struct {
	Name  string
	Env   string // environment name
	Steps []Step
}

// RetryConfig defines retry behavior for a step.
type RetryConfig struct {
	Max          int           // max retry attempts
	Backoff      string        // "fixed", "linear", "exponential"
	InitialDelay time.Duration // starting delay (default 1s)
	RetryOn      []int         // status codes to retry on (e.g., [502, 503, 504])
	RetryOnError bool          // also retry on network errors
	Jitter       bool          // add random jitter
}

// ListenConfig defines a webhook listener that waits for an async callback.
type ListenConfig struct {
	Port    int           // port to listen on
	Path    string        // path to match (e.g., /webhook)
	Timeout time.Duration // max wait time for callback
	Capture string        // variable name to store received body
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
	Poll        *PollConfig
	Retry       *RetryConfig
	Parallel    []Step // sub-steps to run in parallel
	MaxParallel int    // concurrency limit (0 = unlimited)
	FailFast    bool   // cancel remaining on first failure (default true)
	Listen      *ListenConfig // optional webhook listener for async callbacks
}

// PollConfig configures poll-until-ready behavior for async endpoints.
type PollConfig struct {
	Interval time.Duration // time between polls
	Timeout  time.Duration // max total wait time
	Until    string        // JSONPath condition: "$.status == 'completed'"
	Backoff  string        // backoff strategy: "fixed" (default), "linear", "exponential"
}

// PollAttempt records a single poll attempt for debugging/logging.
type PollAttempt struct {
	Number     int
	StatusCode int
	Body       []byte
	Duration   time.Duration
	ConditionMet bool
}

// Assertion defines an expected condition on a response.
type Assertion struct {
	Field    string      // "status", "body.field.path", "header.Name"
	Operator string      // "==", "!=", "<", ">", "contains", "exists"
	Expected interface{} // expected value to compare against
}

// StepResult holds the outcome of executing a single workflow step.
type StepResult struct {
	StepName     string
	Request      HTTPRequest
	Response     HTTPResponse
	Assertions   []AssertionResult
	Extracted    map[string]string
	Error        error
	Duration     time.Duration
	PollAttempts []PollAttempt
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

package domain

const (
	ExitSuccess         = 0 // 2xx response
	ExitHTTPError       = 1 // 4xx/5xx response
	ExitNetworkError    = 2 // Connection refused, DNS failure, etc.
	ExitTimeout         = 3 // Request timed out
	ExitAssertionFailed = 4 // Test assertion failed
	ExitConfigError     = 5 // Invalid config, bad flags, missing file
	ExitWorkflowFailed  = 6 // Workflow step failure
)

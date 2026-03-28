package domain

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var (
	ErrEmptyURL         = errors.New("URL is required")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrInvalidMethod    = errors.New("invalid HTTP method")
	ErrRequestTimeout   = errors.New("request timed out")
	ErrConnectionFailed = errors.New("connection failed")
	ErrVariableNotFound = errors.New("variable not found")
	ErrInvalidAuth      = errors.New("invalid auth configuration")
)

// ExitError is an error that carries a process exit code.
type ExitError struct {
	Code    int
	Message string
	Err     error
}

func (e *ExitError) Error() string {
	return e.Message
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates an ExitError for HTTP 4xx/5xx responses.
func NewHTTPError(statusCode int, err error) *ExitError {
	return &ExitError{
		Code:    ExitHTTPError,
		Message: fmt.Sprintf("HTTP %d", statusCode),
		Err:     err,
	}
}

// NewNetworkError creates an ExitError for network-level failures.
func NewNetworkError(err error) *ExitError {
	return &ExitError{
		Code:    ExitNetworkError,
		Message: fmt.Sprintf("network error: %v", err),
		Err:     err,
	}
}

// NewTimeoutError creates an ExitError for request timeouts.
func NewTimeoutError(err error) *ExitError {
	return &ExitError{
		Code:    ExitTimeout,
		Message: fmt.Sprintf("request timed out: %v", err),
		Err:     err,
	}
}

// NewAssertionError creates an ExitError for test assertion failures.
func NewAssertionError(msg string) *ExitError {
	return &ExitError{
		Code:    ExitAssertionFailed,
		Message: msg,
	}
}

// NewConfigError creates an ExitError for configuration problems.
func NewConfigError(msg string) *ExitError {
	return &ExitError{
		Code:    ExitConfigError,
		Message: msg,
	}
}

// NewWorkflowError creates an ExitError for workflow step failures.
func NewWorkflowError(msg string, err error) *ExitError {
	return &ExitError{
		Code:    ExitWorkflowFailed,
		Message: msg,
		Err:     err,
	}
}

// ClassifyError inspects err and returns the appropriate exit code.
func ClassifyError(err error) int {
	if err == nil {
		return ExitSuccess
	}

	// Check for timeout errors first (most specific).
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return ExitTimeout
	}

	// Check for our typed ExitError.
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}

	// Default to generic failure.
	return 1
}

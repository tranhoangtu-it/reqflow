package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestExitError_Error(t *testing.T) {
	e := &ExitError{Code: 1, Message: "HTTP 404 Not Found"}
	got := e.Error()
	if got != "HTTP 404 Not Found" {
		t.Errorf("Error() = %q, want %q", got, "HTTP 404 Not Found")
	}
}

func TestExitError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("connection refused")
	e := &ExitError{Code: 2, Message: "network error", Err: inner}

	unwrapped := e.Unwrap()
	if unwrapped != inner {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, inner)
	}
}

func TestExitError_Unwrap_Nil(t *testing.T) {
	e := &ExitError{Code: 1, Message: "error"}
	if e.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", e.Unwrap())
	}
}

func TestNewHTTPError_404(t *testing.T) {
	e := NewHTTPError(404, nil)
	if e.Code != ExitHTTPError {
		t.Errorf("Code = %d, want %d", e.Code, ExitHTTPError)
	}
	if e.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestNewHTTPError_500(t *testing.T) {
	inner := fmt.Errorf("server error")
	e := NewHTTPError(500, inner)
	if e.Code != ExitHTTPError {
		t.Errorf("Code = %d, want %d", e.Code, ExitHTTPError)
	}
	if e.Err != inner {
		t.Errorf("Err = %v, want %v", e.Err, inner)
	}
}

func TestNewNetworkError(t *testing.T) {
	inner := fmt.Errorf("dial tcp: connection refused")
	e := NewNetworkError(inner)
	if e.Code != ExitNetworkError {
		t.Errorf("Code = %d, want %d", e.Code, ExitNetworkError)
	}
	if e.Err != inner {
		t.Errorf("Err = %v, want %v", e.Err, inner)
	}
}

func TestNewTimeoutError(t *testing.T) {
	inner := context.DeadlineExceeded
	e := NewTimeoutError(inner)
	if e.Code != ExitTimeout {
		t.Errorf("Code = %d, want %d", e.Code, ExitTimeout)
	}
	if e.Err != inner {
		t.Errorf("Err = %v, want %v", e.Err, inner)
	}
}

func TestNewAssertionError(t *testing.T) {
	e := NewAssertionError("expected status 200, got 404")
	if e.Code != ExitAssertionFailed {
		t.Errorf("Code = %d, want %d", e.Code, ExitAssertionFailed)
	}
	if e.Message != "expected status 200, got 404" {
		t.Errorf("Message = %q, want %q", e.Message, "expected status 200, got 404")
	}
}

func TestNewConfigError(t *testing.T) {
	e := NewConfigError("missing required flag")
	if e.Code != ExitConfigError {
		t.Errorf("Code = %d, want %d", e.Code, ExitConfigError)
	}
	if e.Message != "missing required flag" {
		t.Errorf("Message = %q, want %q", e.Message, "missing required flag")
	}
}

func TestNewWorkflowError(t *testing.T) {
	inner := fmt.Errorf("step 3 failed")
	e := NewWorkflowError("workflow execution failed", inner)
	if e.Code != ExitWorkflowFailed {
		t.Errorf("Code = %d, want %d", e.Code, ExitWorkflowFailed)
	}
	if e.Err != inner {
		t.Errorf("Err = %v, want %v", e.Err, inner)
	}
}

func TestClassifyError_Nil(t *testing.T) {
	code := ClassifyError(nil)
	if code != ExitSuccess {
		t.Errorf("ClassifyError(nil) = %d, want %d", code, ExitSuccess)
	}
}

func TestClassifyError_DeadlineExceeded(t *testing.T) {
	code := ClassifyError(context.DeadlineExceeded)
	if code != ExitTimeout {
		t.Errorf("ClassifyError(DeadlineExceeded) = %d, want %d", code, ExitTimeout)
	}
}

func TestClassifyError_WrappedDeadlineExceeded(t *testing.T) {
	err := fmt.Errorf("request failed: %w", context.DeadlineExceeded)
	code := ClassifyError(err)
	if code != ExitTimeout {
		t.Errorf("ClassifyError(wrapped DeadlineExceeded) = %d, want %d", code, ExitTimeout)
	}
}

func TestClassifyError_ExitError(t *testing.T) {
	e := &ExitError{Code: ExitNetworkError, Message: "network error"}
	code := ClassifyError(e)
	if code != ExitNetworkError {
		t.Errorf("ClassifyError(ExitError) = %d, want %d", code, ExitNetworkError)
	}
}

func TestClassifyError_WrappedExitError(t *testing.T) {
	inner := &ExitError{Code: ExitConfigError, Message: "bad config"}
	err := fmt.Errorf("something: %w", inner)
	code := ClassifyError(err)
	if code != ExitConfigError {
		t.Errorf("ClassifyError(wrapped ExitError) = %d, want %d", code, ExitConfigError)
	}
}

func TestClassifyError_GenericError(t *testing.T) {
	err := errors.New("something went wrong")
	code := ClassifyError(err)
	if code != 1 {
		t.Errorf("ClassifyError(generic) = %d, want 1", code)
	}
}

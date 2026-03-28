package main

import (
	"errors"
	"testing"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestExitCode_NilError(t *testing.T) {
	code := exitCode(nil)
	if code != 0 {
		t.Errorf("exitCode(nil) = %d, want 0", code)
	}
}

func TestExitCode_ExitError(t *testing.T) {
	err := domain.NewNetworkError(errors.New("connection refused"))
	code := exitCode(err)
	if code != domain.ExitNetworkError {
		t.Errorf("exitCode(NetworkError) = %d, want %d", code, domain.ExitNetworkError)
	}
}

func TestExitCode_GenericError(t *testing.T) {
	err := errors.New("something went wrong")
	code := exitCode(err)
	if code != 1 {
		t.Errorf("exitCode(generic) = %d, want 1", code)
	}
}

func TestExitCode_HTTPError(t *testing.T) {
	err := domain.NewHTTPError(404, nil)
	code := exitCode(err)
	if code != domain.ExitHTTPError {
		t.Errorf("exitCode(HTTPError) = %d, want %d", code, domain.ExitHTTPError)
	}
}

package domain

import "testing"

func TestExitCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"ExitSuccess", ExitSuccess, 0},
		{"ExitHTTPError", ExitHTTPError, 1},
		{"ExitNetworkError", ExitNetworkError, 2},
		{"ExitTimeout", ExitTimeout, 3},
		{"ExitAssertionFailed", ExitAssertionFailed, 4},
		{"ExitConfigError", ExitConfigError, 5},
		{"ExitWorkflowFailed", ExitWorkflowFailed, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

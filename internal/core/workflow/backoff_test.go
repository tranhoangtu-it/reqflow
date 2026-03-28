package workflow_test

import (
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/core/workflow"
)

func TestCalculateDelay_Fixed_ReturnsSameDelay(t *testing.T) {
	initial := 1 * time.Second

	for _, attempt := range []int{1, 2, 3, 5, 10} {
		got := workflow.CalculateDelay("fixed", attempt, initial)
		if got != initial {
			t.Errorf("fixed attempt %d: got %v, want %v", attempt, got, initial)
		}
	}
}

func TestCalculateDelay_Linear_IncreasesLinearly(t *testing.T) {
	initial := 1 * time.Second

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 3 * time.Second},
		{5, 5 * time.Second},
	}

	for _, tc := range tests {
		got := workflow.CalculateDelay("linear", tc.attempt, initial)
		if got != tc.want {
			t.Errorf("linear attempt %d: got %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestCalculateDelay_Exponential_Doubles(t *testing.T) {
	initial := 1 * time.Second

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
	}

	for _, tc := range tests {
		got := workflow.CalculateDelay("exponential", tc.attempt, initial)
		if got != tc.want {
			t.Errorf("exponential attempt %d: got %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestCalculateDelay_Attempt0_ReturnsInitialDelay(t *testing.T) {
	initial := 500 * time.Millisecond

	for _, strategy := range []string{"fixed", "linear", "exponential"} {
		got := workflow.CalculateDelay(strategy, 0, initial)
		if got != initial {
			t.Errorf("%s attempt 0: got %v, want %v", strategy, got, initial)
		}
	}
}

func TestCalculateDelay_UnknownStrategy_ReturnsSameAsFixed(t *testing.T) {
	initial := 1 * time.Second

	got := workflow.CalculateDelay("unknown", 3, initial)
	if got != initial {
		t.Errorf("unknown strategy: got %v, want %v (same as fixed)", got, initial)
	}
}

func TestApplyJitter_ResultBetweenDelayAnd150Percent(t *testing.T) {
	delay := 1 * time.Second

	for i := 0; i < 100; i++ {
		got := workflow.ApplyJitter(delay)
		if got < delay || got > delay+delay/2 {
			t.Errorf("jitter iteration %d: got %v, want between %v and %v", i, got, delay, delay+delay/2)
		}
	}
}

func TestApplyJitter_ZeroDelay_ReturnsZero(t *testing.T) {
	got := workflow.ApplyJitter(0)
	if got != 0 {
		t.Errorf("jitter of zero delay: got %v, want 0", got)
	}
}

package domain

import (
	"testing"
	"time"
)

func TestListenConfig_Defaults(t *testing.T) {
	cfg := ListenConfig{
		Port:    8080,
		Path:    "/webhook",
		Timeout: 30 * time.Second,
		Capture: "callback_body",
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.Path != "/webhook" {
		t.Errorf("Path = %q, want %q", cfg.Path, "/webhook")
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want %v", cfg.Timeout, 30*time.Second)
	}
	if cfg.Capture != "callback_body" {
		t.Errorf("Capture = %q, want %q", cfg.Capture, "callback_body")
	}
}

func TestStep_ListenField(t *testing.T) {
	step := Step{
		Name:   "wait for callback",
		Method: MethodGet,
		URL:    "",
		Listen: &ListenConfig{
			Port:    9090,
			Path:    "/callback",
			Timeout: 10 * time.Second,
			Capture: "result",
		},
	}

	if step.Listen == nil {
		t.Fatal("Listen should not be nil")
	}
	if step.Listen.Port != 9090 {
		t.Errorf("Listen.Port = %d, want 9090", step.Listen.Port)
	}
	if step.Listen.Capture != "result" {
		t.Errorf("Listen.Capture = %q, want %q", step.Listen.Capture, "result")
	}
}

func TestStep_ListenNil(t *testing.T) {
	step := Step{
		Name:   "normal step",
		Method: MethodGet,
		URL:    "https://example.com",
	}

	if step.Listen != nil {
		t.Error("Listen should be nil for normal steps")
	}
}

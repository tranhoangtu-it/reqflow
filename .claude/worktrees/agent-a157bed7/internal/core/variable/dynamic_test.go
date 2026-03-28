package variable

import (
	"regexp"
	"strconv"
	"testing"
)

func TestExpandDynamic(t *testing.T) {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	isoRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)

	t.Run("timestamp produces numeric string", func(t *testing.T) {
		vars := map[string]string{"ts": "$timestamp"}
		got := ExpandDynamic(vars)
		if _, err := strconv.ParseInt(got["ts"], 10, 64); err != nil {
			t.Errorf("$timestamp should produce numeric string, got %q", got["ts"])
		}
	})

	t.Run("isoTimestamp produces ISO format", func(t *testing.T) {
		vars := map[string]string{"iso": "$isoTimestamp"}
		got := ExpandDynamic(vars)
		if !isoRegex.MatchString(got["iso"]) {
			t.Errorf("$isoTimestamp should produce ISO format, got %q", got["iso"])
		}
	})

	t.Run("randomInt produces number 0-1000", func(t *testing.T) {
		vars := map[string]string{"r": "$randomInt"}
		got := ExpandDynamic(vars)
		n, err := strconv.Atoi(got["r"])
		if err != nil {
			t.Errorf("$randomInt should produce number string, got %q", got["r"])
		}
		if n < 0 || n > 1000 {
			t.Errorf("$randomInt should be 0-1000, got %d", n)
		}
	})

	t.Run("randomUUID produces UUID v4 format", func(t *testing.T) {
		vars := map[string]string{"id": "$randomUUID"}
		got := ExpandDynamic(vars)
		if !uuidRegex.MatchString(got["id"]) {
			t.Errorf("$randomUUID should produce UUID v4, got %q", got["id"])
		}
	})

	t.Run("guid is alias for UUID format", func(t *testing.T) {
		vars := map[string]string{"id": "$guid"}
		got := ExpandDynamic(vars)
		if !uuidRegex.MatchString(got["id"]) {
			t.Errorf("$guid should produce UUID v4, got %q", got["id"])
		}
	})

	t.Run("regular variables pass through unchanged", func(t *testing.T) {
		vars := map[string]string{
			"host": "example.com",
			"port": "8080",
		}
		got := ExpandDynamic(vars)
		if got["host"] != "example.com" {
			t.Errorf("regular var should pass through, got %q", got["host"])
		}
		if got["port"] != "8080" {
			t.Errorf("regular var should pass through, got %q", got["port"])
		}
	})

	t.Run("mix of dynamic and regular variables", func(t *testing.T) {
		vars := map[string]string{
			"host":  "example.com",
			"ts":    "$timestamp",
			"id":    "$randomUUID",
			"plain": "hello",
		}
		got := ExpandDynamic(vars)

		if got["host"] != "example.com" {
			t.Errorf("regular var should be unchanged, got %q", got["host"])
		}
		if got["plain"] != "hello" {
			t.Errorf("regular var should be unchanged, got %q", got["plain"])
		}
		if _, err := strconv.ParseInt(got["ts"], 10, 64); err != nil {
			t.Errorf("$timestamp should produce numeric string, got %q", got["ts"])
		}
		if !uuidRegex.MatchString(got["id"]) {
			t.Errorf("$randomUUID should produce UUID v4, got %q", got["id"])
		}
	})

	t.Run("partial match is not expanded", func(t *testing.T) {
		vars := map[string]string{
			"val": "prefix$timestamp",
		}
		got := ExpandDynamic(vars)
		if got["val"] != "prefix$timestamp" {
			t.Errorf("partial match should not expand, got %q", got["val"])
		}
	})

	t.Run("empty map returns empty map", func(t *testing.T) {
		got := ExpandDynamic(map[string]string{})
		if len(got) != 0 {
			t.Errorf("empty input should return empty map, got %v", got)
		}
	})
}

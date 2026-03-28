package domain_test

import (
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/domain"
)

func TestCookie_IsExpired(t *testing.T) {
	t.Run("zero time is not expired", func(t *testing.T) {
		c := domain.Cookie{Name: "test", Value: "val"}
		if c.IsExpired() {
			t.Error("cookie with zero expiry should not be expired")
		}
	})

	t.Run("future time is not expired", func(t *testing.T) {
		c := domain.Cookie{
			Name:    "test",
			Value:   "val",
			Expires: time.Now().Add(1 * time.Hour),
		}
		if c.IsExpired() {
			t.Error("cookie with future expiry should not be expired")
		}
	})

	t.Run("past time is expired", func(t *testing.T) {
		c := domain.Cookie{
			Name:    "test",
			Value:   "val",
			Expires: time.Now().Add(-1 * time.Hour),
		}
		if !c.IsExpired() {
			t.Error("cookie with past expiry should be expired")
		}
	})
}

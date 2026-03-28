package cookiejar_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/adapters/cookiejar"
	"github.com/ye-kart/reqflow/internal/domain"
)

func TestSetCookies_StoresCookies(t *testing.T) {
	jar := cookiejar.New()

	err := jar.SetCookies("http://example.com/path", []domain.Cookie{
		{Name: "session", Value: "abc123", Domain: "example.com", Path: "/"},
	})
	if err != nil {
		t.Fatalf("SetCookies: unexpected error: %v", err)
	}

	cookies, err := jar.GetCookies("http://example.com/path")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}

	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "session" || cookies[0].Value != "abc123" {
		t.Errorf("expected session=abc123, got %s=%s", cookies[0].Name, cookies[0].Value)
	}
}

func TestGetCookies_RespectsDomainMatching(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "site", Value: "example", Domain: "example.com", Path: "/"},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "site", Value: "other", Domain: "other.com", Path: "/"},
	})

	cookies, err := jar.GetCookies("http://example.com/")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}

	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie for example.com, got %d", len(cookies))
	}
	if cookies[0].Value != "example" {
		t.Errorf("expected value 'example', got %q", cookies[0].Value)
	}
}

func TestGetCookies_RespectsPathMatching(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/api", []domain.Cookie{
		{Name: "api_token", Value: "tok", Domain: "example.com", Path: "/api"},
	})

	// Should match /api/users
	cookies, err := jar.GetCookies("http://example.com/api/users")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie for /api/users, got %d", len(cookies))
	}

	// Should NOT match /other
	cookies, err = jar.GetCookies("http://example.com/other")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}
	if len(cookies) != 0 {
		t.Errorf("expected 0 cookies for /other, got %d", len(cookies))
	}
}

func TestGetCookies_ExpiredCookiesNotReturned(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{
			Name:    "expired",
			Value:   "old",
			Domain:  "example.com",
			Path:    "/",
			Expires: time.Now().Add(-1 * time.Hour),
		},
		{
			Name:   "fresh",
			Value:  "new",
			Domain: "example.com",
			Path:   "/",
		},
	})

	cookies, err := jar.GetCookies("http://example.com/")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}

	if len(cookies) != 1 {
		t.Fatalf("expected 1 non-expired cookie, got %d", len(cookies))
	}
	if cookies[0].Name != "fresh" {
		t.Errorf("expected 'fresh' cookie, got %q", cookies[0].Name)
	}
}

func TestClear_RemovesAllCookies(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "a", Value: "1", Domain: "example.com", Path: "/"},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "b", Value: "2", Domain: "other.com", Path: "/"},
	})

	err := jar.Clear()
	if err != nil {
		t.Fatalf("Clear: unexpected error: %v", err)
	}

	all, err := jar.All()
	if err != nil {
		t.Fatalf("All: unexpected error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected 0 cookies after Clear, got %d", len(all))
	}
}

func TestClearDomain_RemovesDomainCookiesOnly(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "a", Value: "1", Domain: "example.com", Path: "/"},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "b", Value: "2", Domain: "other.com", Path: "/"},
	})

	err := jar.ClearDomain("example.com")
	if err != nil {
		t.Fatalf("ClearDomain: unexpected error: %v", err)
	}

	all, err := jar.All()
	if err != nil {
		t.Fatalf("All: unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 cookie after ClearDomain, got %d", len(all))
	}
	if all[0].Name != "b" {
		t.Errorf("expected cookie 'b' to remain, got %q", all[0].Name)
	}
}

func TestAll_ReturnsAllNonExpiredCookies(t *testing.T) {
	jar := cookiejar.New()

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "a", Value: "1", Domain: "example.com", Path: "/"},
		{
			Name:    "expired",
			Value:   "old",
			Domain:  "example.com",
			Path:    "/",
			Expires: time.Now().Add(-1 * time.Hour),
		},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "b", Value: "2", Domain: "other.com", Path: "/"},
	})

	all, err := jar.All()
	if err != nil {
		t.Fatalf("All: unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 non-expired cookies, got %d", len(all))
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "cookies.json")

	jar := cookiejar.New()
	expires := time.Now().Add(24 * time.Hour).Truncate(time.Second)

	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{
			Name:     "session",
			Value:    "abc123",
			Domain:   "example.com",
			Path:     "/",
			Expires:  expires,
			Secure:   true,
			HTTPOnly: true,
		},
	})

	err := jar.Save(path)
	if err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	// Verify file exists.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected cookies.json file to be created")
	}

	// Load into a new jar.
	jar2 := cookiejar.New()
	err = jar2.Load(path)
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}

	cookies, err := jar2.GetCookies("http://example.com/")
	if err != nil {
		t.Fatalf("GetCookies: unexpected error: %v", err)
	}

	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie after load, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "session" {
		t.Errorf("Name: want 'session', got %q", c.Name)
	}
	if c.Value != "abc123" {
		t.Errorf("Value: want 'abc123', got %q", c.Value)
	}
	if c.Domain != "example.com" {
		t.Errorf("Domain: want 'example.com', got %q", c.Domain)
	}
	if c.Secure != true {
		t.Errorf("Secure: want true, got false")
	}
	if c.HTTPOnly != true {
		t.Errorf("HTTPOnly: want true, got false")
	}
	if !c.Expires.Equal(expires) {
		t.Errorf("Expires: want %v, got %v", expires, c.Expires)
	}
}

func TestLoad_NonExistentFile_NoError(t *testing.T) {
	jar := cookiejar.New()
	err := jar.Load("/nonexistent/path/cookies.json")
	if err != nil {
		t.Fatalf("Load non-existent file should not error, got: %v", err)
	}

	all, _ := jar.All()
	if len(all) != 0 {
		t.Errorf("expected 0 cookies, got %d", len(all))
	}
}

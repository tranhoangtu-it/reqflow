package commands_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/adapters/cookiejar"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

func newTestAppWithJar(jar *cookiejar.Jar) *app.App {
	mock := &mockHTTPClient{
		doFunc: func(_ context.Context, _ domain.HTTPRequest) (domain.HTTPResponse, error) {
			return domain.HTTPResponse{StatusCode: 200, Body: []byte("ok")}, nil
		},
	}
	a := newTestApp(mock)
	a.CookieJar = jar
	return a
}

func TestCookieListCommand_ShowsStoredCookies(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "session", Value: "abc", Domain: "example.com", Path: "/"},
		{Name: "token", Value: "xyz", Domain: "example.com", Path: "/api"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "session") {
		t.Errorf("expected output to contain 'session', got:\n%s", output)
	}
	if !strings.Contains(output, "token") {
		t.Errorf("expected output to contain 'token', got:\n%s", output)
	}
}

func TestCookieListCommand_FiltersByDomain(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "a", Value: "1", Domain: "example.com", Path: "/"},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "b", Value: "2", Domain: "other.com", Path: "/"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "list", "--domain", "example.com"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "a") {
		t.Errorf("expected output to contain cookie 'a', got:\n%s", output)
	}
	if strings.Contains(output, "other.com") {
		t.Errorf("expected output NOT to contain 'other.com', got:\n%s", output)
	}
}

func TestCookieListCommand_EmptyJar(t *testing.T) {
	jar := cookiejar.New()

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No cookies") {
		t.Errorf("expected 'No cookies' message, got:\n%s", output)
	}
}

func TestCookieClearCommand_EmptiesJar(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "session", Value: "abc", Domain: "example.com", Path: "/"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "clear"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, _ := jar.All()
	if len(all) != 0 {
		t.Errorf("expected 0 cookies after clear, got %d", len(all))
	}
}

func TestCookieClearCommand_ClearsDomainOnly(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "a", Value: "1", Domain: "example.com", Path: "/"},
	})
	_ = jar.SetCookies("http://other.com/", []domain.Cookie{
		{Name: "b", Value: "2", Domain: "other.com", Path: "/"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "clear", "--domain", "example.com"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, _ := jar.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 cookie after domain clear, got %d", len(all))
	}
	if all[0].Name != "b" {
		t.Errorf("expected cookie 'b' to remain, got %q", all[0].Name)
	}
}

func TestCookieDeleteCommand_DeletesSpecificCookie(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{Name: "keep", Value: "yes", Domain: "example.com", Path: "/"},
		{Name: "remove", Value: "no", Domain: "example.com", Path: "/"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "delete", "remove", "--domain", "example.com"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, _ := jar.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 cookie after delete, got %d", len(all))
	}
	if all[0].Name != "keep" {
		t.Errorf("expected 'keep' cookie to remain, got %q", all[0].Name)
	}
}

func TestCookieListCommand_DoesNotShowExpired(t *testing.T) {
	jar := cookiejar.New()
	_ = jar.SetCookies("http://example.com/", []domain.Cookie{
		{
			Name:    "old",
			Value:   "expired",
			Domain:  "example.com",
			Path:    "/",
			Expires: time.Now().Add(-1 * time.Hour),
		},
		{Name: "fresh", Value: "valid", Domain: "example.com", Path: "/"},
	})

	a := newTestAppWithJar(jar)
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"cookie", "list"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "old") {
		t.Errorf("expected output NOT to contain expired cookie 'old', got:\n%s", output)
	}
	if !strings.Contains(output, "fresh") {
		t.Errorf("expected output to contain 'fresh', got:\n%s", output)
	}
}

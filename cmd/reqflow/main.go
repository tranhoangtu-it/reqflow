package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ye-kart/reqflow/internal/adapters/cli"
	"github.com/ye-kart/reqflow/internal/adapters/cookiejar"
	"github.com/ye-kart/reqflow/internal/adapters/httpclient"
	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

func main() {
	if err := run(); err != nil {
		var exitErr *domain.ExitError
		if errors.As(err, &exitErr) {
			fmt.Fprintln(os.Stderr, exitErr.Message)
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// exitCode returns the process exit code for the given error.
func exitCode(err error) int {
	return domain.ClassifyError(err)
}

// defaultCookiePath returns the default path for the cookie jar file.
func defaultCookiePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".reqflow", "cookies.json")
	}
	return filepath.Join(home, ".reqflow", "cookies.json")
}

func run() error {
	// Create cookie jar and load persisted cookies.
	jar := cookiejar.New()
	cookiePath := defaultCookiePath()
	_ = jar.Load(cookiePath) // best-effort load; missing file is fine

	// Create adapters.
	httpClient := httpclient.New(httpclient.WithCookieJar(jar))
	store := storage.NewFilesystem()

	// Create the application coordinator.
	a := app.New(httpClient, store)
	a.CookieJar = jar

	// Create and execute the CLI.
	c := cli.New(a)
	err := c.Execute()

	// Persist cookies after execution (best-effort).
	_ = jar.Save(cookiePath)

	return err
}

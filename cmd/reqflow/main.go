package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ye-kart/reqflow/internal/adapters/cli"
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

func run() error {
	// Create adapters.
	httpClient := httpclient.New()
	store := storage.NewFilesystem()

	// Create the application coordinator.
	a := app.New(httpClient, store)

	// Create and execute the CLI.
	c := cli.New(a)
	return c.Execute()
}

package main

import (
	"fmt"
	"os"

	"github.com/ye-kart/reqflow/internal/adapters/cli"
	"github.com/ye-kart/reqflow/internal/adapters/httpclient"
	"github.com/ye-kart/reqflow/internal/adapters/storage"
	"github.com/ye-kart/reqflow/internal/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
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

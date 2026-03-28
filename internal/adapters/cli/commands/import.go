package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/adapters/cli/output"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/core/importer"
	"github.com/ye-kart/reqflow/internal/domain"
)

func newImportCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import requests from external formats",
	}

	cmd.AddCommand(newImportCurlCommand(a))
	return cmd
}

func newImportCurlCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "curl <curl-command>",
		Short: "Import and execute a cURL command",
		Long:  "Parse a cURL command string and execute the equivalent HTTP request.",
		Args:  cobra.ExactArgs(1),
		RunE:  makeImportCurlRunE(a),
	}

	cmd.Flags().Bool("dry-run", false, "parse and display the request without executing")
	return cmd
}

func makeImportCurlRunE(a *app.App) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		curlCmd := args[0]

		config, err := importer.ParseCurl(curlCmd)
		if err != nil {
			return fmt.Errorf("failed to parse curl command: %w", err)
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return printDryRun(cmd, config)
		}

		// Execute the request.
		timeout := config.Timeout
		if timeout == 0 {
			t, _ := cmd.Flags().GetDuration("timeout")
			timeout = resolveTimeout(t)
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		result, err := a.HTTPExecutor.Execute(ctx, config, nil)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		outputFmt, _ := cmd.Flags().GetString("output")
		noColor, _ := cmd.Flags().GetBool("no-color")
		formatter := output.New(domain.OutputFormat(outputFmt), noColor)
		return formatter.FormatResponse(os.Stdout, result.Response)
	}
}

func printDryRun(cmd *cobra.Command, config domain.RequestConfig) error {
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Method:  %s\n", config.Method)
	fmt.Fprintf(w, "URL:     %s\n", config.URL)
	if len(config.Headers) > 0 {
		fmt.Fprintln(w, "Headers:")
		for _, h := range config.Headers {
			fmt.Fprintf(w, "  %s: %s\n", h.Key, h.Value)
		}
	}
	if len(config.Body) > 0 {
		fmt.Fprintf(w, "Body:    %s\n", string(config.Body))
	}
	if config.Auth != nil {
		fmt.Fprintf(w, "Auth:    %s\n", config.Auth.Type)
	}
	return nil
}

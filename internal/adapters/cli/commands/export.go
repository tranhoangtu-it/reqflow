package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/core/exporter"
	"github.com/ye-kart/reqflow/internal/domain"
)

func newExportCommand(_ *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export requests to external formats",
	}

	cmd.AddCommand(newExportCurlCommand())
	return cmd
}

func newExportCurlCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "curl",
		Short: "Export the last request as a cURL command",
		Long:  "Print the equivalent curl command for the last executed request.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This would require storing the last request config.
			// For now, this is a placeholder that demonstrates the export subcommand exists.
			return fmt.Errorf("no request to export; use --curl flag on request commands instead")
		},
	}
	return cmd
}

// printCurlExport writes the curl equivalent of the config to the command output.
func printCurlExport(cmd *cobra.Command, config domain.RequestConfig) error {
	w := cmd.OutOrStdout()
	curlStr := exporter.ExportCurl(config)
	fmt.Fprintln(w, curlStr)
	return nil
}

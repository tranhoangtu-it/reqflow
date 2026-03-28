package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/domain"
	"github.com/ye-kart/reqflow/internal/platform/config"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage reqflow configuration",
	}

	cmd.AddCommand(newConfigListCommand())
	cmd.AddCommand(newConfigGetCommand())
	cmd.AddCommand(newConfigSetCommand())

	return cmd
}

func newConfigListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all resolved configuration values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFromContext(cmd.Context())
			if cfg == nil {
				return fmt.Errorf("config not loaded")
			}

			fmt.Fprintf(cmd.OutOrStdout(), "default_output: %s\n", cfg.Output)
			fmt.Fprintf(cmd.OutOrStdout(), "default_timeout: %s\n", cfg.Timeout)
			fmt.Fprintf(cmd.OutOrStdout(), "no_color: %v\n", cfg.NoColor)
			if len(cfg.DefaultHeaders) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "default_headers:")
				for _, h := range cfg.DefaultHeaders {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s: %s\n", h.Key, h.Value)
				}
			}

			return nil
		},
	}
}

func newConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a specific configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := configFromContext(cmd.Context())
			if cfg == nil {
				return fmt.Errorf("config not loaded")
			}

			key := args[0]
			switch key {
			case "default_output":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.Output)
			case "default_timeout":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.Timeout)
			case "no_color":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.NoColor)
			case "default_headers":
				for _, h := range cfg.DefaultHeaders {
					fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", h.Key, h.Value)
				}
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			return nil
		},
	}
}

func newConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]
			local, _ := cmd.Flags().GetBool("local")

			opts := configOptsFromContext(cmd.Context())
			cfgPath, loadOpts := resolveConfigPath(local, opts)

			// Load existing config from the target file, or start fresh.
			existing, err := config.LoadFile(cfgPath)
			if err != nil {
				existing = &domain.AppConfig{}
			}

			// Also load merged config to get defaults for unset fields.
			_ = loadOpts

			switch key {
			case "default_output":
				existing.Output = domain.OutputFormat(value)
			case "default_timeout":
				d, err := time.ParseDuration(value)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				existing.Timeout = d
			case "no_color":
				switch value {
				case "true":
					existing.NoColor = true
				case "false":
					existing.NoColor = false
				default:
					return fmt.Errorf("invalid value for no_color: %s (expected true or false)", value)
				}
			default:
				return fmt.Errorf("unknown config key: %s", key)
			}

			return config.SaveFile(cfgPath, existing)
		},
	}

	cmd.Flags().Bool("local", false, "set in project-local config instead of global")

	return cmd
}

// resolveConfigPath determines the config file path based on the --local flag.
func resolveConfigPath(local bool, opts []config.LoadOption) (string, []config.LoadOption) {
	if local {
		dir := config.ResolveProjectDir(opts)
		return filepath.Join(dir, ".reqflow.yaml"), opts
	}
	dir := config.ResolveGlobalDir(opts)
	return filepath.Join(dir, "config.yaml"), opts
}

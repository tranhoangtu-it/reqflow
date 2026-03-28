package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// OutputFormatCompletions returns valid values for the --output flag.
func OutputFormatCompletions() []string {
	return []string{"pretty", "json", "raw", "minimal"}
}

// HeaderNameCompletions returns common HTTP header names for -H flag completion.
func HeaderNameCompletions() []string {
	return []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"User-Agent",
		"Cache-Control",
		"X-Request-ID",
		"X-API-Key",
	}
}

// AuthBasicCompletions returns completion hints for the --auth-basic flag.
func AuthBasicCompletions() []string {
	return []string{"username:password"}
}

// AuthBearerCompletions returns completion hints for the --auth-bearer flag.
func AuthBearerCompletions() []string {
	return []string{"token"}
}

// newCompletionCommand creates a custom completion command that overrides Cobra's
// default, adding helpful installation instructions.
func newCompletionCommand() *cobra.Command {
	long := `Generate shell completion scripts for reqflow.

Installation:

  # Bash
  echo 'source <(reqflow completion bash)' >> ~/.bashrc

  # Zsh
  echo 'source <(reqflow completion zsh)' >> ~/.zshrc

  # Fish
  reqflow completion fish | source

  # PowerShell
  reqflow completion powershell | Out-String | Invoke-Expression
`

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long:  long,
		Args:  cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(cmd.OutOrStdout(), true)
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}

// registerFlagCompletions registers custom completion functions for flags on the root command.
func registerFlagCompletions(root *cobra.Command) {
	_ = root.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return OutputFormatCompletions(), cobra.ShellCompDirectiveNoFileComp
	})

	_ = root.RegisterFlagCompletionFunc("header", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return HeaderNameCompletions(), cobra.ShellCompDirectiveNoFileComp
	})

	// Register auth flag completions on all subcommands that have them.
	for _, sub := range root.Commands() {
		if sub.Flags().Lookup("auth-basic") != nil {
			_ = sub.RegisterFlagCompletionFunc("auth-basic", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return AuthBasicCompletions(), cobra.ShellCompDirectiveNoFileComp
			})
		}
		if sub.Flags().Lookup("auth-bearer") != nil {
			_ = sub.RegisterFlagCompletionFunc("auth-bearer", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
				return AuthBearerCompletions(), cobra.ShellCompDirectiveNoFileComp
			})
		}
	}
}

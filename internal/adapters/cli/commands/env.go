package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/app"
	"github.com/ye-kart/reqflow/internal/domain"
)

// defaultEnvDir returns the default environment directory (~/.reqflow/environments/).
func defaultEnvDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".reqflow/environments"
	}
	return filepath.Join(home, ".reqflow", "environments")
}

func newEnvCommand(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environments",
		Long:  "Manage environment files for variable substitution in requests.",
	}

	cmd.PersistentFlags().String("env-dir", defaultEnvDir(), "directory containing environment files")

	cmd.AddCommand(newEnvListCommand(a))
	cmd.AddCommand(newEnvShowCommand(a))
	cmd.AddCommand(newEnvCreateCommand(a))
	cmd.AddCommand(newEnvSetCommand(a))

	return cmd
}

func newEnvListCommand(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available environments",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, _ := cmd.Flags().GetString("env-dir")

			names, err := a.Storage.ListEnvironments(dir)
			if err != nil {
				return fmt.Errorf("listing environments: %w", err)
			}

			if len(names) == 0 {
				cmd.Println("No environments found.")
				return nil
			}

			for _, name := range names {
				cmd.Println(name)
			}
			return nil
		},
	}
}

func newEnvShowCommand(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Display variables in an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("env-dir")
			name := args[0]
			path := filepath.Join(dir, name+".yaml")

			env, err := a.Storage.ReadEnvironment(path)
			if err != nil {
				return fmt.Errorf("reading environment %q: %w", name, err)
			}

			cmd.Printf("Environment: %s\n", env.Name)
			if len(env.Variables) == 0 {
				cmd.Println("  (no variables)")
				return nil
			}

			for _, v := range env.Variables {
				cmd.Printf("  %s = %s\n", v.Key, v.Value)
			}
			return nil
		},
	}
}

func newEnvCreateCommand(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new empty environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("env-dir")
			name := args[0]
			path := filepath.Join(dir, name+".yaml")

			env := domain.Environment{
				Name: name,
			}

			if err := a.Storage.WriteEnvironment(path, env); err != nil {
				return fmt.Errorf("creating environment %q: %w", name, err)
			}

			cmd.Printf("Created environment %q\n", name)
			return nil
		},
	}
}

func newEnvSetCommand(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> <key> <value>",
		Short: "Set a variable in an environment",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("env-dir")
			name := args[0]
			key := args[1]
			value := args[2]
			path := filepath.Join(dir, name+".yaml")

			// Read existing environment or start fresh.
			env, err := a.Storage.ReadEnvironment(path)
			if err != nil {
				env = domain.Environment{Name: name}
			}

			// Update or add the variable.
			found := false
			for i, v := range env.Variables {
				if v.Key == key {
					env.Variables[i].Value = value
					found = true
					break
				}
			}
			if !found {
				env.Variables = append(env.Variables, domain.Variable{
					Key:   key,
					Value: value,
					Scope: domain.ScopeEnvironment,
				})
			}

			if err := a.Storage.WriteEnvironment(path, env); err != nil {
				return fmt.Errorf("writing environment %q: %w", name, err)
			}

			cmd.Printf("Set %s=%s in environment %q\n", key, value, name)
			return nil
		},
	}
}

package cli

import (
	"github.com/spf13/cobra"
	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
)

// CLI wraps the root cobra command.
type CLI struct {
	root *cobra.Command
}

// New creates a new CLI wired to the given App.
func New(a *app.App) *CLI {
	return &CLI{
		root: commands.NewRootCommand(a),
	}
}

// Execute runs the CLI, parsing os.Args and dispatching to the appropriate command.
func (c *CLI) Execute() error {
	return c.root.Execute()
}

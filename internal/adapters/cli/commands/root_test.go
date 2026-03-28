package commands_test

import (
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
)

func TestRootCommand_HasExpectedSubcommands(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	expectedCmds := []string{"get", "post", "put", "patch", "delete", "config"}
	subCmds := root.Commands()

	cmdNames := make(map[string]bool)
	for _, cmd := range subCmds {
		cmdNames[cmd.Name()] = true
	}

	for _, expected := range expectedCmds {
		if !cmdNames[expected] {
			t.Errorf("expected subcommand %q not found; have: %v", expected, cmdNames)
		}
	}
}

func TestRootCommand_HasGlobalFlags(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	flags := []struct {
		name      string
		shorthand string
	}{
		{"output", "o"},
		{"no-color", ""},
		{"timeout", "t"},
		{"verbose", "v"},
		{"header", "H"},
		{"query", "q"},
	}

	for _, f := range flags {
		flag := root.PersistentFlags().Lookup(f.name)
		if flag == nil {
			t.Errorf("expected persistent flag %q not found", f.name)
			continue
		}
		if f.shorthand != "" && flag.Shorthand != f.shorthand {
			t.Errorf("flag %q: expected shorthand %q, got %q", f.name, f.shorthand, flag.Shorthand)
		}
	}
}

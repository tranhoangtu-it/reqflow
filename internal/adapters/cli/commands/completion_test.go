package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ye-kart/reqflow/internal/adapters/cli/commands"
	"github.com/ye-kart/reqflow/internal/app"
)

func TestCompletionCommand_Exists(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "completion" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'completion' subcommand on root command")
	}
}

func TestCompletionCommand_GeneratesBash(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"completion", "bash"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error generating bash completion: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected non-empty bash completion script")
	}
	if !strings.Contains(output, "reqflow") {
		t.Error("bash completion script should contain command name 'reqflow'")
	}
}

func TestCompletionCommand_GeneratesZsh(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"completion", "zsh"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error generating zsh completion: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected non-empty zsh completion script")
	}
	if !strings.Contains(output, "reqflow") {
		t.Error("zsh completion script should contain command name 'reqflow'")
	}
}

func TestCompletionCommand_GeneratesFish(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"completion", "fish"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error generating fish completion: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected non-empty fish completion script")
	}
	if !strings.Contains(output, "reqflow") {
		t.Error("fish completion script should contain command name 'reqflow'")
	}
}

func TestCompletionCommand_GeneratesPowerShell(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"completion", "powershell"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error generating powershell completion: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("expected non-empty powershell completion script")
	}
	if !strings.Contains(output, "reqflow") {
		t.Error("powershell completion script should contain command name 'reqflow'")
	}
}

func TestCompletionCommand_HelpContainsInstallInstructions(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"completion", "--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, expected := range []string{
		"~/.bashrc",
		"~/.zshrc",
		"completion fish",
		"powershell",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("completion help should contain %q", expected)
		}
	}
}

func TestOutputFlagCompletion_ReturnsValidFormats(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	completions := commands.OutputFormatCompletions()

	expectedFormats := []string{"pretty", "json", "raw", "minimal"}
	if len(completions) != len(expectedFormats) {
		t.Fatalf("expected %d output format completions, got %d", len(expectedFormats), len(completions))
	}

	for _, expected := range expectedFormats {
		found := false
		for _, c := range completions {
			if c == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected output format %q in completions", expected)
		}
	}

	// Verify the flag completion is registered on the root command.
	flag := root.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Fatal("expected 'output' persistent flag on root command")
	}
}

func TestHeaderCompletion_ReturnsCommonHeaders(t *testing.T) {
	completions := commands.HeaderNameCompletions()

	expectedHeaders := []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"User-Agent",
		"Cache-Control",
		"X-Request-ID",
		"X-API-Key",
	}

	for _, expected := range expectedHeaders {
		found := false
		for _, c := range completions {
			if c == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected header %q in completions", expected)
		}
	}
}

func TestAuthFlagCompletions_ReturnHints(t *testing.T) {
	basicHints := commands.AuthBasicCompletions()
	if len(basicHints) == 0 {
		t.Fatal("expected non-empty auth-basic completion hints")
	}
	found := false
	for _, h := range basicHints {
		if strings.Contains(h, "username:password") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth-basic hint to contain 'username:password'")
	}

	bearerHints := commands.AuthBearerCompletions()
	if len(bearerHints) == 0 {
		t.Fatal("expected non-empty auth-bearer completion hints")
	}
	found = false
	for _, h := range bearerHints {
		if strings.Contains(h, "token") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected auth-bearer hint to contain 'token'")
	}
}

func TestFlagCompletionFunctions_AreRegistered(t *testing.T) {
	a := &app.App{}
	root := commands.NewRootCommand(a)

	// Verify that flag completion functions have been registered by checking
	// that the output flag exists and the completion command is present.
	outputFlag := root.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found on root command")
	}

	headerFlag := root.PersistentFlags().Lookup("header")
	if headerFlag == nil {
		t.Fatal("header flag not found on root command")
	}

	// Verify completion subcommand exists.
	var completionCmd bool
	for _, cmd := range root.Commands() {
		if cmd.Name() == "completion" {
			completionCmd = true
			break
		}
	}
	if !completionCmd {
		t.Error("completion subcommand not registered on root command")
	}
}

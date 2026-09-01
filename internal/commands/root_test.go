package commands

import (
	"bytes"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"version"})

	if err := command.Execute(); err != nil {
		t.Fatalf("executing version command: %v", err)
	}

	if got, want := output.String(), "0.0.0-dev\n"; got != want {
		t.Errorf("version output = %q, want %q", got, want)
	}
}

func TestRootCommandDefaultsToTUIAndRejectsArguments(t *testing.T) {
	command := NewRootCommand()
	if command.RunE == nil {
		t.Error("root command has no TUI handler")
	}
	if err := command.Args(command, []string{"unexpected-argument"}); err == nil {
		t.Error("root command accepted a positional argument")
	}
}

package commands

import (
	"bytes"
	"strings"
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

func TestRootCommandHelpAndInvalidFlags(t *testing.T) {
	t.Run("help", func(t *testing.T) {
		command := NewRootCommand()
		output := new(bytes.Buffer)
		command.SetOut(output)
		command.SetArgs([]string{"--help"})
		if err := command.Execute(); err != nil {
			t.Fatalf("executing root help: %v", err)
		}
		if !strings.Contains(output.String(), "A terminal-native tracker for Beans-format tasks") {
			t.Errorf("help output = %q", output.String())
		}
	})

	t.Run("invalid flag", func(t *testing.T) {
		command := NewRootCommand()
		command.SetArgs([]string{"--unknown"})
		if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "unknown flag") {
			t.Errorf("root flag error = %v", err)
		}
	})

	t.Run("positional argument", func(t *testing.T) {
		command := NewRootCommand()
		command.SetArgs([]string{"unexpected-argument"})
		if err := command.Execute(); err == nil {
			t.Error("root command accepted a positional argument")
		}
	})
}

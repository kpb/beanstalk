package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionCommandGeneratesScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			command := NewRootCommand()
			output := new(bytes.Buffer)
			command.SetOut(output)
			command.SetArgs([]string{"completion", shell})

			if err := command.Execute(); err != nil {
				t.Fatalf("generating %s completion: %v", shell, err)
			}
			if !strings.Contains(output.String(), "beanstalk") {
				t.Errorf("%s completion does not reference the command: %q", shell, output.String())
			}
		})
	}
}

func TestCompletionCommandRejectsUnsupportedShell(t *testing.T) {
	command := NewRootCommand()
	command.SetArgs([]string{"completion", "unsupported"})

	if err := command.Execute(); err == nil {
		t.Error("completion command accepted an unsupported shell")
	}
}

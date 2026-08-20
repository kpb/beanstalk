package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrimeCommandPrintsDefaultInstructionsOutsideProject(t *testing.T) {
	t.Chdir(t.TempDir())

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"prime"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing prime command: %v", err)
	}
	if got, want := output.String(), defaultPrimeInstructions; got != want {
		t.Errorf("prime output = %q, want %q", got, want)
	}
}

func TestPrimeCommandUsesCurrentDirectoryOverride(t *testing.T) {
	workingDirectory := t.TempDir()
	instructions := "Use this project's workflow.\n"
	config := "prime:\n  instructions: |\n    Use this project's workflow.\n"
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beanstalk.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"prime"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing prime command: %v", err)
	}
	if got := output.String(); got != instructions {
		t.Errorf("prime output = %q, want %q", got, instructions)
	}
}

func TestPrimeCommandDoesNotSearchParentDirectories(t *testing.T) {
	parent := t.TempDir()
	if err := os.WriteFile(filepath.Join(parent, ".beanstalk.yaml"), []byte("prime:\n  instructions: parent\n"), 0o644); err != nil {
		t.Fatalf("writing parent config: %v", err)
	}
	workingDirectory := filepath.Join(parent, "child")
	if err := os.Mkdir(workingDirectory, 0o755); err != nil {
		t.Fatalf("creating child directory: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"prime"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing prime command: %v", err)
	}
	if got, want := output.String(), defaultPrimeInstructions; got != want {
		t.Errorf("prime output = %q, want default instructions", got)
	}
}

func TestPrimeCommandReportsMalformedConfig(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beanstalk.yaml"), []byte("prime: ["), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	command.SetArgs([]string{"prime"})
	err := command.Execute()
	if err == nil {
		t.Fatal("prime command succeeded with malformed config")
	}
	if !strings.Contains(err.Error(), "parsing .beanstalk.yaml") {
		t.Errorf("prime error = %v, want parsing error", err)
	}
}

func TestPrimeCommandRejectsArgumentsAndJSON(t *testing.T) {
	for _, args := range [][]string{{"prime", "extra"}, {"prime", "--json"}} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("prime %v succeeded", args[1:])
		}
	}
}

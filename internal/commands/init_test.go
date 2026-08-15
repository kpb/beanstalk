package commands

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	workingDirectory := filepath.Join(t.TempDir(), "my-project")
	if err := os.Mkdir(workingDirectory, 0o755); err != nil {
		t.Fatalf("creating working directory: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"init"})

	if err := command.Execute(); err != nil {
		t.Fatalf("executing init command: %v", err)
	}
	if got, want := output.String(), "Initialized .beans and .beans.yml\n"; got != want {
		t.Errorf("init output = %q, want %q", got, want)
	}

	gitignore, err := os.ReadFile(filepath.Join(workingDirectory, ".beans", ".gitignore"))
	if err != nil {
		t.Fatalf("reading generated gitignore: %v", err)
	}
	if got, want := string(gitignore), beansGitignore; got != want {
		t.Errorf("gitignore = %q, want %q", got, want)
	}

	config, err := os.ReadFile(filepath.Join(workingDirectory, ".beans.yml"))
	if err != nil {
		t.Fatalf("reading generated config: %v", err)
	}
	for _, want := range []string{"name: \"my-project\"", "prefix: \"my-project-\"", "base_ref: \"main\""} {
		if !strings.Contains(string(config), want) {
			t.Errorf("generated config does not contain %q", want)
		}
	}
}

func TestInitCommandRefusesExistingPaths(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating existing .beans directory: %v", err)
	}

	command := NewRootCommand()
	command.SetArgs([]string{"init"})
	err := command.Execute()
	if err == nil {
		t.Fatal("init command succeeded with an existing .beans directory")
	}
	if _, err := os.Stat(filepath.Join(workingDirectory, ".beans.yml")); !os.IsNotExist(err) {
		t.Errorf(".beans.yml was created after failed initialization: %v", err)
	}
}

func TestInitCommandReportsExistingProject(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating existing .beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), nil, 0o644); err != nil {
		t.Fatalf("creating existing config: %v", err)
	}

	command := NewRootCommand()
	command.SetArgs([]string{"init"})
	err := command.Execute()
	if err == nil {
		t.Fatal("init command succeeded with an existing Beans project")
	}
	if !errors.Is(err, errAlreadyInitialized) {
		t.Errorf("init error = %v, want errAlreadyInitialized", err)
	}
}

func TestInitCommandRefusesExistingConfig(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), nil, 0o644); err != nil {
		t.Fatalf("creating existing config: %v", err)
	}

	command := NewRootCommand()
	command.SetArgs([]string{"init"})
	err := command.Execute()
	if err == nil {
		t.Fatal("init command succeeded with an existing config")
	}
	if _, err := os.Stat(filepath.Join(workingDirectory, ".beans")); !os.IsNotExist(err) {
		t.Errorf(".beans was created after failed initialization: %v", err)
	}
}

func TestInitializeReportsFilesystemFailure(t *testing.T) {
	workingDirectory := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(workingDirectory, nil, 0o644); err != nil {
		t.Fatalf("creating file in place of working directory: %v", err)
	}

	err := initialize(workingDirectory)
	if err == nil {
		t.Fatal("initialize succeeded with a file in place of the working directory")
	}
	var pathError *os.PathError
	if !errors.As(err, &pathError) {
		t.Errorf("initialize error = %v, want wrapped *os.PathError", err)
	}
}

func TestBaseRefUsesOriginHead(t *testing.T) {
	workingDirectory := t.TempDir()
	if output, err := exec.Command("git", "init", "--quiet", workingDirectory).CombinedOutput(); err != nil {
		t.Fatalf("initializing git repository: %v\n%s", err, output)
	}
	if output, err := exec.Command("git", "-C", workingDirectory, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/trunk").CombinedOutput(); err != nil {
		t.Fatalf("setting origin HEAD: %v\n%s", err, output)
	}

	if got, want := baseRef(workingDirectory), "origin/trunk"; got != want {
		t.Errorf("baseRef() = %q, want %q", got, want)
	}
}

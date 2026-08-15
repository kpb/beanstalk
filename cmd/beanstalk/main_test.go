package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunReportsAlreadyInitializedProject(t *testing.T) {
	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating existing .beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), nil, 0o644); err != nil {
		t.Fatalf("creating existing config: %v", err)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if got, want := run([]string{"init"}, stdout, stderr), 1; got != want {
		t.Errorf("run exit code = %d, want %d", got, want)
	}
	if got, want := stdout.String(), ""; got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
	if got, want := stderr.String(), "error: this directory is already initialized: .beans and .beans.yml already exist\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}

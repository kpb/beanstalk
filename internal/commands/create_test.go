package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
	"gopkg.in/yaml.v3"
)

func TestCreateCommand(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"create", "Add", "login", "--priority", "high", "--tag", "api", "--tag", "auth", "--body", "Implement OAuth."})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing create command: %v", err)
	}
	if !strings.HasPrefix(output.String(), "Created ") || !strings.HasSuffix(output.String(), "--add-login.md\n") {
		t.Errorf("create output = %q", output.String())
	}

	entries, err := os.ReadDir(filepath.Join(workingDirectory, ".beans"))
	if err != nil {
		t.Fatalf("reading beans directory: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("beans entries = %d, want 2", len(entries))
	}
	var beanName string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".md") {
			beanName = entry.Name()
		}
	}
	contents, err := os.ReadFile(filepath.Join(workingDirectory, ".beans", beanName))
	if err != nil {
		t.Fatalf("reading bean: %v", err)
	}
	for _, want := range []string{"# ", "title: Add login", "status: todo", "type: task", "priority: high", "- api", "- auth", "Implement OAuth."} {
		if !strings.Contains(string(contents), want) {
			t.Errorf("bean does not contain %q:\n%s", want, contents)
		}
	}
}

func TestCreateCommandJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"create", "JSON", "bean", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing create command: %v", err)
	}
	var response struct {
		Success bool       `json:"success"`
		Bean    beans.Bean `json:"bean"`
		Message string     `json:"message"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if !response.Success || response.Message != "Bean created" || response.Bean.Title != "JSON bean" || response.Bean.Path == "" {
		t.Errorf("JSON response = %#v", response)
	}
}

func TestCreateCommandRequiresInitializedProject(t *testing.T) {
	t.Chdir(t.TempDir())
	command := NewRootCommand()
	command.SetArgs([]string{"create", "Missing", "project"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "run beanstalk init first") {
		t.Errorf("create error = %v", err)
	}
}

func TestCreateCommandRejectsInvalidStatus(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)
	command := NewRootCommand()
	command.SetArgs([]string{"create", "Invalid", "--status", "unknown"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("create error = %v", err)
	}
}

func initializedProject(t *testing.T) string {
	t.Helper()
	workingDirectory := t.TempDir()
	if err := initialize(workingDirectory); err != nil {
		t.Fatalf("initializing project: %v", err)
	}
	return workingDirectory
}

func TestRenderBeanIsValidYAML(t *testing.T) {
	contents, err := beans.Render(beans.Bean{ID: "test-a1b2", Title: "Test", Status: "todo", Type: "task"})
	if err != nil {
		t.Fatalf("rendering bean: %v", err)
	}
	parts := strings.Split(string(contents), "---\n")
	var metadata map[string]any
	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		t.Fatalf("parsing rendered YAML: %v", err)
	}
	if metadata["title"] != "Test" {
		t.Errorf("title = %v", metadata["title"])
	}
}

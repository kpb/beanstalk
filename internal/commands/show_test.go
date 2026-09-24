package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestShowCommandDisplaysTaskAndJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.August, 16, 13, 30, 0, 0, time.UTC)
	writeBean(t, workingDirectory, ".beans/archive/project-a1--task.md", beans.Bean{ID: "project-a1", Slug: "task", Title: "Task title", Status: "in-progress", Type: "feature", Priority: "high", Tags: []string{"api", "auth"}, Parent: "project-parent", CreatedAt: createdAt, UpdatedAt: updatedAt, Body: "Implement the endpoint."})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"show", "project-a1"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing show command: %v", err)
	}
	want := "ID: project-a1\nStatus: in-progress\nType: feature\nPriority: high\nTags: api, auth\nParent: project-parent\nCreated: 2026-08-15T12:00:00Z\nUpdated: 2026-08-16T13:30:00Z\n\nTask title\n\nImplement the endpoint.\n"
	if got := output.String(); got != want {
		t.Errorf("show output = %q, want %q", got, want)
	}

	command = NewRootCommand()
	output.Reset()
	command.SetOut(output)
	command.SetArgs([]string{"show", "project-a1", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing JSON show command: %v", err)
	}
	var shown beans.Bean
	if err := json.Unmarshal(output.Bytes(), &shown); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if shown.ID != "project-a1" || shown.Path != "archive/project-a1--task.md" || shown.Parent != "project-parent" || shown.Body != "Implement the endpoint." || !shown.CreatedAt.Equal(createdAt) || !shown.UpdatedAt.Equal(updatedAt) {
		t.Errorf("shown bean = %#v", shown)
	}
}

func TestShowCommandEscapesTerminalControlsButPreservesJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	bean := beans.Bean{ID: "project-\x1b[31m", Slug: "task", Title: "Task\rtitle", Status: "todo", Type: "task", Tags: []string{"tag\a"}, Body: "Body\x1b[2J\nnext\rline"}
	contents := "---\ntitle: \"Task\\x0dtitle\"\nstatus: todo\ntype: task\npriority: normal\ntags:\n  - \"tag\\x07\"\n---\n" + bean.Body
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans", bean.ID+"--task.md"), []byte(contents), 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"show", bean.ID})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing show command: %v", err)
	}
	for _, control := range []string{"\x1b", "\r", "\a"} {
		if strings.Contains(output.String(), control) {
			t.Errorf("human output contains control %q: %q", control, output.String())
		}
	}
	for _, want := range []string{`project-\x1b[31m`, `Task\x0dtitle`, `tag\x07`, `Body\x1b[2J`, `next\x0dline`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("human output does not contain %q: %q", want, output.String())
		}
	}

	command = NewRootCommand()
	output.Reset()
	command.SetOut(output)
	command.SetArgs([]string{"show", bean.ID, "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing JSON show command: %v", err)
	}
	var shown beans.Bean
	if err := json.Unmarshal(output.Bytes(), &shown); err != nil {
		t.Fatalf("decoding JSON output: %v", err)
	}
	if shown.ID != bean.ID || shown.Title != bean.Title || shown.Tags[0] != bean.Tags[0] || shown.Body != bean.Body {
		t.Errorf("JSON bean = %#v, want original controls", shown)
	}
}

func TestShowCommandReportsMissingAndDuplicateIDs(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--first.md", beans.Bean{ID: "project-a1", Title: "First", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/archive/project-a1--second.md", beans.Bean{ID: "project-a1", Title: "Second", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	for _, test := range []struct {
		args        []string
		errorString string
	}{
		{args: []string{"show", "missing"}, errorString: "bean not found"},
		{args: []string{"show", "project-a1"}, errorString: "multiple beans have the same ID"},
		{args: []string{"show"}},
	} {
		command := NewRootCommand()
		command.SetArgs(test.args)
		if err := command.Execute(); err == nil {
			t.Errorf("show %v succeeded", test.args)
		} else if test.errorString != "" && !strings.Contains(err.Error(), test.errorString) {
			t.Errorf("show %v error = %v", test.args, err)
		}
	}
}

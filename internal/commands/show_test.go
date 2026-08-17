package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestShowCommandDisplaysTaskAndJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	createdAt := time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.August, 16, 13, 30, 0, 0, time.UTC)
	writeBean(t, workingDirectory, ".beans/archive/project-a1--task.md", beans.Bean{ID: "project-a1", Slug: "task", Title: "Task title", Status: "in-progress", Type: "feature", Priority: "high", Tags: []string{"api", "auth"}, CreatedAt: createdAt, UpdatedAt: updatedAt, Body: "Implement the endpoint."})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"show", "project-a1"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing show command: %v", err)
	}
	want := "ID: project-a1\nStatus: in-progress\nType: feature\nPriority: high\nTags: api, auth\nCreated: 2026-08-15T12:00:00Z\nUpdated: 2026-08-16T13:30:00Z\n\nTask title\n\nImplement the endpoint.\n"
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
	if shown.ID != "project-a1" || shown.Path != "archive/project-a1--task.md" || shown.Body != "Implement the endpoint." || !shown.CreatedAt.Equal(createdAt) || !shown.UpdatedAt.Equal(updatedAt) {
		t.Errorf("shown bean = %#v", shown)
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

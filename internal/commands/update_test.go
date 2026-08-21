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

func TestUpdateCommandUpdatesStatusAndPreservesTaskContents(t *testing.T) {
	workingDirectory := initializedProject(t)
	path := filepath.Join(workingDirectory, ".beans", "archive", "project-a1--task.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating archive: %v", err)
	}
	contents := "---\n# project-a1\ntitle: Keep task details\nstatus: todo\ntype: task\nparent: project-parent\ntags:\n  - api\ncreated_at: 2026-08-15T12:00:00Z\nupdated_at: 2026-08-15T12:00:00Z\ncustom_field: preserve me\n---\nKeep this body exactly.\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"update", "project-a1", "--status", "in-progress", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing update command: %v", err)
	}
	var response struct {
		Success bool       `json:"success"`
		Bean    beans.Bean `json:"bean"`
		Message string     `json:"message"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if !response.Success || response.Message != "Bean updated" || response.Bean.Status != "in-progress" || response.Bean.Path != "archive/project-a1--task.md" {
		t.Errorf("JSON response = %#v", response)
	}
	if got, want := response.Bean.CreatedAt, time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("created_at = %s, want %s", got, want)
	}
	if !response.Bean.UpdatedAt.After(response.Bean.CreatedAt) {
		t.Errorf("updated_at = %s, want after %s", response.Bean.UpdatedAt, response.Bean.CreatedAt)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading updated bean: %v", err)
	}
	for _, want := range []string{"# project-a1", "status: in-progress", "parent: project-parent", "custom_field: preserve me", "Keep this body exactly.\n"} {
		if !strings.Contains(string(updated), want) {
			t.Errorf("updated bean does not contain %q:\n%s", want, updated)
		}
	}
	loaded, err := beans.Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading updated bean: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded beans = %d, want 1", len(loaded))
	}
	bean := loaded[0]
	if bean.Status != "in-progress" || bean.Parent != "project-parent" || bean.Body != "Keep this body exactly." || !bean.CreatedAt.Equal(response.Bean.CreatedAt) || !bean.UpdatedAt.Equal(response.Bean.UpdatedAt) {
		t.Errorf("loaded bean = %#v", bean)
	}
}

func TestUpdateCommandReportsInvalidRequests(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--task.md", beans.Bean{ID: "project-a1", Title: "Task", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	for _, args := range [][]string{
		{"update", "project-a1"},
		{"update", "project-a1", "--status", "unknown"},
		{"update", "missing", "--status", "completed"},
		{"update", "project-a1", "extra", "--status", "completed"},
	} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("update %v succeeded", args)
		}
	}
}

func TestUpdateCommandRejectsDuplicateIDs(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--first.md", beans.Bean{ID: "project-a1", Title: "First", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/archive/project-a1--second.md", beans.Bean{ID: "project-a1", Title: "Second", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	command.SetArgs([]string{"update", "project-a1", "--status", "completed"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "multiple beans have the same ID") {
		t.Errorf("update error = %v", err)
	}
}

func TestUpdateCommandChangesAndRemovesParent(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-parent--parent.md", beans.Bean{ID: "project-parent", Title: "Parent", Status: "todo", Type: "feature"})
	writeBean(t, workingDirectory, ".beans/project-child--child.md", beans.Bean{ID: "project-child", Title: "Child", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	command.SetArgs([]string{"update", "project-child", "--parent", "project-parent"})
	if err := command.Execute(); err != nil {
		t.Fatalf("setting parent: %v", err)
	}
	child, err := beans.Find(workingDirectory, "project-child")
	if err != nil {
		t.Fatalf("finding child: %v", err)
	}
	if child.Parent != "project-parent" {
		t.Errorf("parent = %q", child.Parent)
	}

	command = NewRootCommand()
	command.SetArgs([]string{"update", "project-child", "--parent", ""})
	if err := command.Execute(); err != nil {
		t.Fatalf("removing parent: %v", err)
	}
	child, err = beans.Find(workingDirectory, "project-child")
	if err != nil {
		t.Fatalf("finding child: %v", err)
	}
	if child.Parent != "" {
		t.Errorf("parent = %q, want empty", child.Parent)
	}
}

func TestUpdateCommandRejectsInvalidParents(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a--a.md", beans.Bean{ID: "project-a", Title: "A", Status: "todo", Type: "task", Parent: "project-b"})
	writeBean(t, workingDirectory, ".beans/project-b--b.md", beans.Bean{ID: "project-b", Title: "B", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	for _, args := range [][]string{
		{"update", "project-a", "--parent", "project-a"},
		{"update", "project-a", "--parent", "missing"},
		{"update", "project-b", "--parent", "project-a"},
	} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("update %v succeeded", args)
		}
	}
}

func TestUpdateCommandRejectsDuplicateParentID(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-child--child.md", beans.Bean{ID: "project-child", Title: "Child", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/project-parent--first.md", beans.Bean{ID: "project-parent", Title: "First", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/archive/project-parent--second.md", beans.Bean{ID: "project-parent", Title: "Second", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	command.SetArgs([]string{"update", "project-child", "--parent", "project-parent"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "multiple beans have the same ID") {
		t.Errorf("update error = %v", err)
	}
}

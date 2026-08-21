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

func TestListCommand(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-b2--fix-parser.md", beans.Bean{ID: "project-b2", Slug: "fix-parser", Title: "Fix parser", Status: "in-progress", Type: "bug", Priority: "high"})
	writeBean(t, workingDirectory, ".beans/project-a1--add-login.md", beans.Bean{ID: "project-a1", Slug: "add-login", Title: "Add login", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/archive/project-c3--old-task.md", beans.Bean{ID: "project-c3", Slug: "old-task", Title: "Old task", Status: "completed", Type: "task"})
	writeBean(t, workingDirectory, ".beans/.conversations/ignored.md", beans.Bean{ID: "ignored", Title: "Ignored", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"list"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing list command: %v", err)
	}
	if got, want := output.String(), "project-b2  in-progress  bug  -  Fix parser\nproject-a1  todo  task  -  Add login\nproject-c3  completed  task  -  Old task\n"; got != want {
		t.Errorf("list output = %q, want %q", got, want)
	}
}

func TestListCommandFiltersAndJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--task.md", beans.Bean{ID: "project-a1", Slug: "task", Title: "Task", Status: "todo", Type: "task", Parent: "project-parent", Body: "Not included"})
	writeBean(t, workingDirectory, ".beans/project-b2--bug.md", beans.Bean{ID: "project-b2", Slug: "bug", Title: "Bug", Status: "todo", Type: "bug"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"list", "--status", "todo", "--type", "task", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing list command: %v", err)
	}
	var listed []beans.Bean
	if err := json.Unmarshal(output.Bytes(), &listed); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if len(listed) != 1 || listed[0].ID != "project-a1" || listed[0].Priority != "normal" || listed[0].Parent != "project-parent" || listed[0].Body != "" {
		t.Errorf("JSON beans = %#v", listed)
	}
}

func TestListCommandAcceptsRepeatedFilters(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--task.md", beans.Bean{ID: "project-a1", Slug: "task", Title: "Task", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/project-b2--bug.md", beans.Bean{ID: "project-b2", Slug: "bug", Title: "Bug", Status: "draft", Type: "bug"})
	writeBean(t, workingDirectory, ".beans/project-c3--feature.md", beans.Bean{ID: "project-c3", Slug: "feature", Title: "Feature", Status: "todo", Type: "feature"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"list", "--status", "todo", "--status", "draft", "--type", "task", "--type", "bug"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing list command: %v", err)
	}
	if got, want := output.String(), "project-a1  todo  task  -  Task\nproject-b2  draft  bug  -  Bug\n"; got != want {
		t.Errorf("filtered list output = %q, want %q", got, want)
	}
}

func TestListCommandRejectsInvalidArgumentsAndFilters(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)
	for _, args := range [][]string{{"list", "extra"}, {"list", "--status", "unknown"}, {"list", "--type", "unknown"}} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("list %v succeeded", args)
		}
	}
}

func TestListCommandReportsEmptyAndUninitializedProjects(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)
	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"list"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing empty list command: %v", err)
	}
	if got, want := output.String(), "No beans found.\n"; got != want {
		t.Errorf("empty list output = %q, want %q", got, want)
	}

	t.Chdir(t.TempDir())
	command = NewRootCommand()
	command.SetArgs([]string{"list"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "run beanstalk init first") {
		t.Errorf("uninitialized list error = %v", err)
	}
}

func writeBean(t *testing.T, workingDirectory, relativePath string, bean beans.Bean) {
	t.Helper()
	path := filepath.Join(workingDirectory, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating bean directory: %v", err)
	}
	if bean.CreatedAt.IsZero() {
		bean.CreatedAt = time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)
		bean.UpdatedAt = bean.CreatedAt
	}
	contents, err := beans.Render(bean)
	if err != nil {
		t.Fatalf("rendering bean: %v", err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}
}

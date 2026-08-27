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
	if got, want := output.String(), "ID                  S  T  TITLE\nproject-b2          >  B  Fix parser\nproject-a1          o  T  Add login\nproject-c3          x  T  Old task\n"; got != want {
		t.Errorf("list output = %q, want %q", got, want)
	}
}

func TestListCommandFiltersAndJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--task.md", beans.Bean{ID: "project-a1", Slug: "task", Title: "Task", Status: "todo", Type: "task", Parent: "project-parent", Body: "Not included"})
	writeBean(t, workingDirectory, ".beans/project-parent--parent.md", beans.Bean{ID: "project-parent", Title: "Parent", Status: "todo", Type: "epic"})
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
	if got, want := output.String(), "ID                  S  T  TITLE\nproject-a1          o  T  Task\nproject-b2          ?  B  Bug\n"; got != want {
		t.Errorf("filtered list output = %q, want %q", got, want)
	}
}

func TestListCommandRendersHierarchyWithFilteredParentContext(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-parent--parent.md", beans.Bean{ID: "project-parent", Title: "Parent", Status: "completed", Type: "epic"})
	writeBean(t, workingDirectory, ".beans/project-child-a--first-child.md", beans.Bean{ID: "project-child-a", Title: "First child", Status: "todo", Type: "task", Parent: "project-parent"})
	writeBean(t, workingDirectory, ".beans/project-grandchild--grandchild.md", beans.Bean{ID: "project-grandchild", Title: "Grandchild", Status: "todo", Type: "task", Parent: "project-child-a"})
	writeBean(t, workingDirectory, ".beans/project-child-b--second-child.md", beans.Bean{ID: "project-child-b", Title: "Second child", Status: "todo", Type: "task", Parent: "project-parent"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"list", "--status", "todo"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing filtered list command: %v", err)
	}
	if got, want := output.String(), "ID                  S  T  TITLE\nproject-parent      x  E  Parent\nproject-child-a     o  T  |- First child\nproject-grandchild  o  T  |  `- Grandchild\nproject-child-b     o  T  `- Second child\n"; got != want {
		t.Errorf("hierarchical list output = %q, want %q", got, want)
	}
}

func TestListCommandReportsMalformedParentHierarchy(t *testing.T) {
	tests := []struct {
		name  string
		beans []beans.Bean
		want  string
	}{
		{
			name:  "missing parent",
			beans: []beans.Bean{{ID: "project-child", Title: "Child", Status: "todo", Type: "task", Parent: "missing"}},
			want:  "parent bean not found: missing",
		},
		{
			name: "cycle",
			beans: []beans.Bean{
				{ID: "project-a", Title: "A", Status: "todo", Type: "task", Parent: "project-b"},
				{ID: "project-b", Title: "B", Status: "todo", Type: "task", Parent: "project-a"},
			},
			want: "parent link would create a cycle",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workingDirectory := initializedProject(t)
			for _, bean := range test.beans {
				writeBean(t, workingDirectory, ".beans/"+bean.ID+"--task.md", bean)
			}
			t.Chdir(workingDirectory)

			command := NewRootCommand()
			command.SetArgs([]string{"list"})
			if err := command.Execute(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("list error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestListCommandReportsImportedMetadataDiagnostics(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-task--task.md", beans.Bean{ID: "project-task", Title: "Task", Status: "unsupported", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	command.SetArgs([]string{"list"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), `invalid bean status "unsupported": project-task`) {
		t.Errorf("list error = %v", err)
	}
}

func TestListMarkers(t *testing.T) {
	for _, test := range []struct {
		name string
		got  string
		want string
	}{
		{"draft status", statusMarker("draft"), "?"},
		{"todo status", statusMarker("todo"), "o"},
		{"in-progress status", statusMarker("in-progress"), ">"},
		{"completed status", statusMarker("completed"), "x"},
		{"scrapped status", statusMarker("scrapped"), "-"},
		{"unknown status", statusMarker("unknown"), "!"},
		{"milestone type", typeMarker("milestone"), "M"},
		{"epic type", typeMarker("epic"), "E"},
		{"bug type", typeMarker("bug"), "B"},
		{"feature type", typeMarker("feature"), "F"},
		{"task type", typeMarker("task"), "T"},
		{"unknown type", typeMarker("unknown"), "!"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Errorf("marker = %q, want %q", test.got, test.want)
			}
		})
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

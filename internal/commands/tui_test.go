package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestLoadTUITasksSortsAndReportsProjectErrors(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-b--second.md", beans.Bean{ID: "project-b", Title: "Second", Status: "todo", Type: "task"})
	writeBean(t, workingDirectory, ".beans/project-a--first.md", beans.Bean{ID: "project-a", Title: "First", Status: "in-progress", Type: "task"})
	writeBean(t, workingDirectory, ".beans/archive/project-c--archived.md", beans.Bean{ID: "project-c", Title: "Archived", Status: "completed", Type: "task"})

	loaded, err := loadTUITasks(workingDirectory, false)
	if err != nil {
		t.Fatalf("loading TUI tasks: %v", err)
	}
	if got, want := []string{loaded[0].ID, loaded[1].ID}, []string{"project-a", "project-b"}; !equalStringSlices(got, want) {
		t.Errorf("loaded IDs = %v, want %v", got, want)
	}
	loaded, err = loadTUITasks(workingDirectory, true)
	if err != nil {
		t.Fatalf("loading archived TUI tasks: %v", err)
	}
	if got, want := []string{loaded[0].ID, loaded[1].ID, loaded[2].ID}, []string{"project-a", "project-b", "project-c"}; !equalStringSlices(got, want) {
		t.Errorf("loaded archived IDs = %v, want %v", got, want)
	}

	_, err = loadTUITasks(t.TempDir(), false)
	if !errors.Is(err, beans.ErrNotInitialized) {
		t.Errorf("load error = %v, want ErrNotInitialized", err)
	}
}

func TestLoadTUITasksReportsImportedTaskDiagnostics(t *testing.T) {
	for _, test := range []struct {
		name string
		bean beans.Bean
		want string
	}{
		{
			name: "unsupported metadata",
			bean: beans.Bean{ID: "project-task", Title: "Task", Status: "unsupported", Type: "task"},
			want: `invalid bean status "unsupported": project-task`,
		},
		{
			name: "missing parent",
			bean: beans.Bean{ID: "project-task", Title: "Task", Status: "todo", Type: "task", Parent: "missing"},
			want: "parent bean not found: missing",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			workingDirectory := initializedProject(t)
			writeBean(t, workingDirectory, ".beans/project-task--task.md", test.bean)

			_, err := loadTUITasks(workingDirectory, false)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("TUI load error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestTUICommandIsRegistered(t *testing.T) {
	command := NewRootCommand()
	found, _, err := command.Find([]string{"tui"})
	if err != nil || found == nil || found.Name() != "tui" || !strings.Contains(found.Short, "interactively") {
		t.Errorf("tui command = %#v, error = %v", found, err)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

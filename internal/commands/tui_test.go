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

	loaded, err := loadTUITasks(workingDirectory)
	if err != nil {
		t.Fatalf("loading TUI tasks: %v", err)
	}
	if got, want := []string{loaded[0].ID, loaded[1].ID}, []string{"project-a", "project-b"}; !equalStringSlices(got, want) {
		t.Errorf("loaded IDs = %v, want %v", got, want)
	}

	_, err = loadTUITasks(t.TempDir())
	if !errors.Is(err, beans.ErrNotInitialized) {
		t.Errorf("load error = %v, want ErrNotInitialized", err)
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

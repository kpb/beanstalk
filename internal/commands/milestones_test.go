package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestMilestonesCommandDisplaysActiveProgressAndJSON(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-release--release.md", beans.Bean{ID: "project-release", Title: "Release", Status: "todo", Type: "milestone"})
	writeBean(t, workingDirectory, ".beans/project-done--done.md", beans.Bean{ID: "project-done", Title: "Done", Status: "completed", Type: "task", Parent: "project-release"})
	writeBean(t, workingDirectory, ".beans/project-scrapped--scrapped.md", beans.Bean{ID: "project-scrapped", Title: "Scrapped", Status: "scrapped", Type: "task", Parent: "project-release"})
	writeBean(t, workingDirectory, ".beans/project-past--past.md", beans.Bean{ID: "project-past", Title: "Past", Status: "completed", Type: "milestone"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"milestones"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing milestones command: %v", err)
	}
	if got, want := output.String(), "project-release  todo  2/2 (100%)  Release\n"; got != want {
		t.Errorf("milestones output = %q, want %q", got, want)
	}

	command = NewRootCommand()
	output.Reset()
	command.SetOut(output)
	command.SetArgs([]string{"milestones", "--all", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing JSON milestones command: %v", err)
	}
	var progresses []beans.MilestoneProgress
	if err := json.Unmarshal(output.Bytes(), &progresses); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if len(progresses) != 2 || progresses[0].ID != "project-release" || progresses[0].Resolved != 2 || progresses[0].Completed != 1 || progresses[0].Scrapped != 1 || progresses[0].Statuses.Scrapped != 1 || progresses[1].ID != "project-past" {
		t.Errorf("progresses = %#v", progresses)
	}
}

func TestMilestonesCommandReportsEmptyAndInvalidRequests(t *testing.T) {
	workingDirectory := initializedProject(t)
	t.Chdir(workingDirectory)
	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"milestones"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing empty milestones command: %v", err)
	}
	if got, want := output.String(), "No milestones found.\n"; got != want {
		t.Errorf("milestones output = %q, want %q", got, want)
	}

	command = NewRootCommand()
	command.SetArgs([]string{"milestones", "extra"})
	if err := command.Execute(); err == nil {
		t.Error("milestones command accepted an argument")
	}
	_, err := loadMilestoneProgresses(t.TempDir(), false)
	if !errors.Is(err, beans.ErrNotInitialized) {
		t.Errorf("load error = %v, want ErrNotInitialized", err)
	}
}

func TestMilestonesCommandRejectsInvalidImportedHierarchy(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-release--release.md", beans.Bean{ID: "project-release", Title: "Release", Status: "todo", Type: "milestone"})
	writeBean(t, workingDirectory, ".beans/project-child--child.md", beans.Bean{ID: "project-child", Title: "Child", Status: "todo", Type: "task", Parent: "missing"})
	t.Chdir(workingDirectory)
	command := NewRootCommand()
	command.SetArgs([]string{"milestones"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "parent bean not found") {
		t.Errorf("milestones error = %v", err)
	}
}

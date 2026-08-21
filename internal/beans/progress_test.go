package beans

import (
	"errors"
	"strings"
	"testing"
)

func TestMilestoneProgressesCountsNestedLeaves(t *testing.T) {
	loaded := []Bean{
		{ID: "milestone", Title: "Release", Status: "todo", Type: "milestone"},
		{ID: "epic", Title: "Epic", Status: "in-progress", Type: "epic", Parent: "milestone"},
		{ID: "todo", Title: "Todo", Status: "todo", Type: "task", Parent: "epic"},
		{ID: "done", Title: "Done", Status: "completed", Type: "task", Parent: "milestone"},
		{ID: "scrapped", Title: "Scrapped", Status: "scrapped", Type: "task", Parent: "milestone"},
		{ID: "nested", Title: "Nested", Status: "draft", Type: "milestone", Parent: "milestone"},
	}

	progresses, err := MilestoneProgresses(loaded)
	if err != nil {
		t.Fatalf("calculating progress: %v", err)
	}
	if len(progresses) != 2 {
		t.Fatalf("progresses = %#v", progresses)
	}
	progress := progresses[0]
	if progress.ID != "milestone" || progress.Total != 4 || progress.Resolved != 2 || progress.Completed != 1 || progress.Scrapped != 1 || progress.Percent != 50 {
		t.Errorf("progress = %#v", progress)
	}
	if got, want := progress.Statuses, (ProgressStatuses{Todo: 1, Draft: 1, Completed: 1, Scrapped: 1}); got != want {
		t.Errorf("statuses = %#v, want %#v", got, want)
	}
	if nested := progresses[1]; nested.Total != 0 || nested.Resolved != 0 || nested.Percent != 0 {
		t.Errorf("nested progress = %#v", nested)
	}
}

func TestMilestoneProgressesRejectsInvalidHierarchiesAndStatuses(t *testing.T) {
	for _, test := range []struct {
		name   string
		loaded []Bean
		err    error
		text   string
	}{
		{
			name:   "duplicate IDs",
			loaded: []Bean{{ID: "milestone", Title: "One", Status: "todo", Type: "milestone"}, {ID: "milestone", Title: "Two", Status: "todo", Type: "task"}},
			err:    ErrDuplicateBeanID,
		},
		{
			name:   "missing parent",
			loaded: []Bean{{ID: "milestone", Title: "Release", Status: "todo", Type: "milestone"}, {ID: "child", Title: "Child", Status: "todo", Type: "task", Parent: "missing"}},
			text:   "parent bean not found",
		},
		{
			name:   "cycle",
			loaded: []Bean{{ID: "milestone", Title: "Release", Status: "todo", Type: "milestone", Parent: "child"}, {ID: "child", Title: "Child", Status: "todo", Type: "task", Parent: "milestone"}},
			err:    ErrParentCycle,
		},
		{
			name:   "unknown leaf status",
			loaded: []Bean{{ID: "milestone", Title: "Release", Status: "todo", Type: "milestone"}, {ID: "child", Title: "Child", Status: "unknown", Type: "task", Parent: "milestone"}},
			text:   "invalid leaf status",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := MilestoneProgresses(test.loaded)
			if err == nil {
				t.Fatal("MilestoneProgresses succeeded")
			}
			if test.err != nil && !errors.Is(err, test.err) {
				t.Errorf("error = %v, want %v", err, test.err)
			}
			if test.text != "" && !strings.Contains(err.Error(), test.text) {
				t.Errorf("error = %v, want %q", err, test.text)
			}
		})
	}
}

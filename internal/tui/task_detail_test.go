package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestRenderTaskDetailIncludesMetadataHierarchyProgressAndBody(t *testing.T) {
	created := time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)
	loaded := []beans.Bean{
		{ID: "release", Title: "Release", Status: "in-progress", Type: "milestone"},
		{ID: "epic", Title: "Interactive workflow", Status: "todo", Type: "epic", Parent: "release"},
		{ID: "selected", Title: "Selected task", Status: "todo", Type: "task", Priority: "high", Tags: []string{"tui", "detail"}, Parent: "epic", CreatedAt: created, UpdatedAt: created, Body: "Implement a detailed view.\n\nKeep its contents readable."},
		{ID: "child", Title: "Child task", Status: "completed", Type: "task", Parent: "selected"},
	}

	got := renderTaskDetail(loaded, loaded[2], 120, 0)
	for _, want := range []string{
		"Selected task", "ID: selected", "Status: todo", "Type: task", "Priority: high", "Tags: tui, detail",
		"Created: 2026-08-20T10:00:00Z", "Parent: epic Interactive workflow", "Children:", "  child Child task",
		"Milestone: release Release (1/1, 100%)", "Body:", "Implement a detailed view.", "Keep its contents readable.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("detail does not contain %q:\n%s", want, got)
		}
	}
}

func TestRenderTaskDetailWrapsAndBoundsOutput(t *testing.T) {
	selected := beans.Bean{ID: "task", Title: "A very long task title", Status: "todo", Type: "task", Priority: "normal", Body: "This body line needs wrapping."}
	got := renderTaskDetail([]beans.Bean{selected}, selected, 12, 4)
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("detail line count = %d, want 4:\n%s", len(lines), got)
	}
	for _, line := range lines {
		if len([]rune(line)) > 12 {
			t.Errorf("line width = %d, want at most 12: %q", len([]rune(line)), line)
		}
	}
}

func TestWrapDetailLinePreservesMarkdownWhitespace(t *testing.T) {
	for _, line := range []string{"    code block", "  - nested list item"} {
		wrapped := wrapDetailLine(line, 8)
		if got := strings.Join(wrapped, ""); got != line {
			t.Errorf("wrapped line = %q, want original whitespace preserved as %q", got, line)
		}
	}
}

func TestRenderTaskDetailUsesPlaceholdersForMissingOptionalData(t *testing.T) {
	selected := beans.Bean{ID: "task", Title: "Task", Status: "todo", Type: "task", Priority: "normal"}
	got := renderTaskDetail([]beans.Bean{selected}, selected, 80, 0)
	for _, want := range []string{"Tags: -", "Parent: -", "Children: -", "Body:\n-"} {
		if !strings.Contains(got, want) {
			t.Errorf("detail does not contain %q:\n%s", want, got)
		}
	}
}

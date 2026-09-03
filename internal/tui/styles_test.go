package tui

import (
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestStatusLabelsUseDistinctSemanticStyles(t *testing.T) {
	for status, color := range map[string]string{
		"draft":       ansiDim,
		"todo":        ansiYellow,
		"in-progress": ansiBlue,
		"completed":   ansiGreen,
		"scrapped":    ansiMagenta,
	} {
		if got, want := statusLabel(status), color+status+ansiReset; got != want {
			t.Errorf("statusLabel(%q) = %q, want %q", status, got, want)
		}
	}
}

func TestTaskRowViewKeepsStyledRowsWithinTerminalWidth(t *testing.T) {
	row := taskRow{bean: beans.Bean{ID: "project-a", Title: "A task title", Status: "in-progress", Priority: "normal", Type: "task"}}
	view := taskRowView(row, nil, nil, true, 50)
	if got := displayWidth(view); got != 50 {
		t.Errorf("display width = %d, want 50: %q", got, view)
	}
	for _, want := range []string{ansiCyan + ">", ansiBlue + "in-progress", ansiReset} {
		if !strings.Contains(view, want) {
			t.Errorf("styled row does not contain %q: %q", want, view)
		}
	}
}

func TestJoinPanesAlignsStyledContentByDisplayWidth(t *testing.T) {
	view := joinPanes([]string{heading("Tasks")}, []string{"Details"}, 10, 10)
	line := strings.TrimSuffix(view, "\n")
	right := strings.Index(line, "Details")
	if got, want := displayWidth(line[:right]), 11; got != want {
		t.Errorf("right pane column = %d, want %d: %q", got, want, line)
	}
}

func TestBorderedPaneHasConsistentDimensions(t *testing.T) {
	pane := borderedPane("Tasks", []string{"first", "second"}, 20, 6)
	if got, want := len(pane), 6; got != want {
		t.Fatalf("pane height = %d, want %d", got, want)
	}
	if !strings.HasPrefix(pane[0], "╭ Tasks ") || !strings.HasSuffix(pane[0], "╮") {
		t.Errorf("top border = %q", pane[0])
	}
	if !strings.HasPrefix(pane[len(pane)-1], "╰") || !strings.HasSuffix(pane[len(pane)-1], "╯") {
		t.Errorf("bottom border = %q", pane[len(pane)-1])
	}
	for _, line := range pane {
		if got, want := displayWidth(line), 20; got != want {
			t.Errorf("pane width = %d, want %d: %q", got, want, line)
		}
	}
}

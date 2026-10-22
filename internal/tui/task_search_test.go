package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/kpb/beanstalk/internal/beans"
)

func TestTaskSearchCancelAndNavigateCollapsedMatch(t *testing.T) {
	loaded := testBeans()
	loaded[1].Body = "Find ME"
	loaded[2].Title = "Find another"
	model := NewTaskList(loaded)
	model = updateTaskList(t, model, key("down"))
	model = updateTaskList(t, model, key("left"))
	model = updateTaskList(t, model, key("left"))
	model = updateTaskList(t, model, key("/"))
	for _, letter := range "find" {
		model = updateTaskList(t, model, key(string(letter)))
	}
	if len(model.searchMatches) != 2 || model.rows[model.cursor].bean.ID != "project-a" {
		t.Errorf("search model = %#v", model)
	}
	model = updateTaskList(t, model, key("down"))
	if model.searchMatches[model.searchCursor].ID != "project-c" {
		t.Errorf("browsed match = %#v", model.searchMatches[model.searchCursor])
	}
	model = updateTaskList(t, model, key("esc"))
	if model.showSearch || model.rows[model.cursor].bean.ID != "project-a" || !model.collapsed["project-a"] {
		t.Errorf("canceled search = %#v", model)
	}
	model = updateTaskList(t, model, key("/"))
	for _, letter := range "find" {
		model = updateTaskList(t, model, key(string(letter)))
	}
	model = updateTaskList(t, model, key("enter"))
	if model.showSearch || model.rows[model.cursor].bean.ID != "project-b" || model.collapsed["project-a"] {
		t.Errorf("navigated search = %#v", model)
	}
}

func TestTaskSearchEditsQueryAndHandlesNoMatches(t *testing.T) {
	model := NewTaskList(testBeans())
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 40, Height: 5})
	model = updateTaskList(t, model, key("/"))
	model = updateTaskList(t, model, key("z"))
	if view := unstyled(model.View().Content); !strings.Contains(view, "No matches") {
		t.Errorf("search view = %q", view)
	}
	model = updateTaskList(t, model, key("enter"))
	if !model.showSearch {
		t.Error("enter closed zero-match search")
	}
	model = updateTaskList(t, model, tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	if model.searchQuery != "" || len(model.searchMatches) != 0 {
		t.Errorf("cleared search = %#v", model)
	}
}

func TestTaskSearchRefreshAndArchiveToggle(t *testing.T) {
	active := []beans.Bean{{ID: "active", Title: "Search", Status: "todo"}}
	archived := beans.Bean{ID: "archived", Path: "archive/archived.md", Title: "Search", Status: "completed"}
	model := NewTaskList(active, WithTaskLoader(func(showArchived bool) ([]beans.Bean, error) {
		if showArchived {
			return []beans.Bean{active[0], archived}, nil
		}
		return active, nil
	}))
	model = updateTaskList(t, model, key("a"))
	model = updateTaskList(t, model, taskListLoadedMessage{beans: []beans.Bean{active[0], archived}, showArchived: true, request: model.loadRequest})
	model = updateTaskList(t, model, key("/"))
	for _, letter := range "search" {
		model = updateTaskList(t, model, key(string(letter)))
	}
	if len(model.searchMatches) != 2 {
		t.Errorf("archived search matches = %#v", model.searchMatches)
	}
	model = updateTaskList(t, model, key("down"))
	model = updateTaskList(t, model, taskListLoadedMessage{beans: []beans.Bean{archived}, showArchived: true, request: model.loadRequest, refresh: true})
	if len(model.searchMatches) != 1 || model.searchMatches[model.searchCursor].ID != "archived" {
		t.Errorf("matches after refresh = %#v", model.searchMatches)
	}
	model = updateTaskList(t, model, taskListLoadedMessage{beans: active, request: model.loadRequest, refresh: true})
	if len(model.searchMatches) != 1 || model.searchMatches[model.searchCursor].ID != "active" {
		t.Errorf("matches after removal = %#v", model.searchMatches)
	}
}

package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/kpb/beanstalk/internal/beans"
)

func TestTaskListNavigationAndBounds(t *testing.T) {
	model := NewTaskList(testBeans())
	model = updateTaskList(t, model, key("down"))
	model = updateTaskList(t, model, key("j"))
	model = updateTaskList(t, model, key("down"))
	if model.cursor != 2 {
		t.Errorf("cursor = %d, want 2", model.cursor)
	}

	model = updateTaskList(t, model, key("g"))
	if model.cursor != 0 {
		t.Errorf("cursor after g = %d, want 0", model.cursor)
	}
	model = updateTaskList(t, model, key("end"))
	if model.cursor != 2 {
		t.Errorf("cursor after end = %d, want 2", model.cursor)
	}
	model = updateTaskList(t, model, key("up"))
	model = updateTaskList(t, model, key("k"))
	model = updateTaskList(t, model, key("up"))
	if model.cursor != 0 {
		t.Errorf("cursor = %d, want 0", model.cursor)
	}
}

func TestTaskListScrollsAndRenders(t *testing.T) {
	model := NewTaskList(testBeans())
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 54, Height: 7})
	model = updateTaskList(t, model, key("down"))
	model = updateTaskList(t, model, key("down"))
	if model.cursor != 2 || model.offset != 1 {
		t.Errorf("model = %#v, want cursor 2 and offset 1", model)
	}

	view := model.View()
	for _, want := range []string{"Beanstalk tasks (3)", "> project-c", "q quit"} {
		if !strings.Contains(view.Content, want) {
			t.Errorf("view does not contain %q:\n%s", want, view.Content)
		}
	}
	if !view.AltScreen || view.WindowTitle != "Beanstalk" {
		t.Errorf("view settings = %#v", view)
	}
}

func TestTaskListBoundsInitialAndShortTerminalViews(t *testing.T) {
	loaded := append(testBeans(),
		beans.Bean{ID: "project-d", Title: "Fourth", Status: "todo", Priority: "normal", Type: "task"},
		beans.Bean{ID: "project-e", Title: "Fifth", Status: "todo", Priority: "normal", Type: "task"},
		beans.Bean{ID: "project-f", Title: "Sixth", Status: "todo", Priority: "normal", Type: "task"},
	)
	model := NewTaskList(loaded)
	if got, want := model.visibleRows(), defaultVisibleRows; got != want {
		t.Errorf("initial visible rows = %d, want %d", got, want)
	}

	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 80, Height: fixedLines})
	if got := model.visibleRows(); got != 1 {
		t.Errorf("short-terminal visible rows = %d, want 1", got)
	}
	view := model.View().Content
	if !strings.Contains(view, "> project-a") || strings.Contains(view, "project-b") {
		t.Errorf("short-terminal view = %q", view)
	}
	if got := strings.Count(strings.TrimSuffix(view, "\n"), "\n") + 1; got > fixedLines {
		t.Errorf("short-terminal lines = %d, height = %d", got, fixedLines)
	}
}

func TestTaskListHandlesEmptyAndQuit(t *testing.T) {
	model := NewTaskList(nil)
	model = updateTaskList(t, model, key("down"))
	if model.cursor != 0 || model.offset != 0 {
		t.Errorf("empty model = %#v", model)
	}
	if got := model.View().Content; !strings.Contains(got, "No beans found.") {
		t.Errorf("empty view = %q", got)
	}

	updated, command := model.Update(key("q"))
	if command == nil {
		t.Fatal("quit command is nil")
	}
	message := command()
	if _, ok := message.(tea.QuitMsg); !ok {
		t.Errorf("quit message = %T, want tea.QuitMsg", message)
	}
	if _, ok := updated.(TaskList); !ok {
		t.Errorf("updated model = %T, want TaskList", updated)
	}
	_, command = model.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	if command == nil {
		t.Fatal("ctrl-c quit command is nil")
	}
}

func updateTaskList(t *testing.T, model TaskList, message tea.Msg) TaskList {
	t.Helper()
	updated, _ := model.Update(message)
	list, ok := updated.(TaskList)
	if !ok {
		t.Fatalf("updated model = %T, want TaskList", updated)
	}
	return list
}

func key(value string) tea.KeyPressMsg {
	if value == "up" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyUp})
	}
	if value == "down" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	}
	if value == "end" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnd})
	}
	return tea.KeyPressMsg(tea.Key{Text: value, Code: rune(value[0])})
}

func testBeans() []beans.Bean {
	return []beans.Bean{
		{ID: "project-a", Title: "First", Status: "todo", Priority: "normal", Type: "task"},
		{ID: "project-b", Title: "Second", Status: "todo", Priority: "normal", Type: "task", Parent: "project-a"},
		{ID: "project-c", Title: "Third", Status: "todo", Priority: "normal", Type: "task"},
	}
}

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

func TestFlattenTaskTreeOrdersHierarchyAndOrphansStably(t *testing.T) {
	loaded := []beans.Bean{
		{ID: "root-a", Title: "Root A"},
		{ID: "child-a", Title: "Child A", Parent: "root-a"},
		{ID: "grandchild", Title: "Grandchild", Parent: "child-a"},
		{ID: "orphan", Title: "Orphan", Parent: "missing"},
		{ID: "root-b", Title: "Root B"},
		{ID: "child-b", Title: "Child B", Parent: "root-b"},
	}

	rows, children := flattenTaskTree(loaded, nil)
	if got, want := rowIDs(rows), []string{"root-a", "child-a", "grandchild", "orphan", "root-b", "child-b"}; !equalStringSlices(got, want) {
		t.Errorf("row IDs = %v, want %v", got, want)
	}
	if got, want := rowDepths(rows), []int{0, 1, 2, 0, 0, 1}; !equalIntSlices(got, want) {
		t.Errorf("row depths = %v, want %v", got, want)
	}
	if !children["root-a"] || !children["child-a"] || !children["root-b"] || children["orphan"] {
		t.Errorf("children = %v", children)
	}
}

func TestFlattenTaskTreeHandlesCyclesWithoutDroppingRows(t *testing.T) {
	loaded := []beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "cycle-a", Title: "Cycle A", Parent: "cycle-b"},
		{ID: "cycle-b", Title: "Cycle B", Parent: "cycle-a"},
	}

	rows, _ := flattenTaskTree(loaded, nil)
	if got, want := rowIDs(rows), []string{"root", "cycle-a", "cycle-b"}; !equalStringSlices(got, want) {
		t.Errorf("row IDs = %v, want %v", got, want)
	}
	if got, want := rowDepths(rows), []int{0, 0, 1}; !equalIntSlices(got, want) {
		t.Errorf("row depths = %v, want %v", got, want)
	}
}

func TestTaskListCollapseUpdatesRowsAndClampsSelection(t *testing.T) {
	model := NewTaskList([]beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "child", Title: "Child", Parent: "root"},
		{ID: "grandchild", Title: "Grandchild", Parent: "child"},
		{ID: "other", Title: "Other"},
	})
	model.cursor = 3
	model.collapsed["root"] = true
	model.rebuildRows()
	if got, want := rowIDs(model.rows), []string{"root", "other"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs = %v, want %v", got, want)
	}
	if model.cursor != 1 || model.offset != 0 {
		t.Errorf("model after collapse = %#v, want cursor 1 and offset 0", model)
	}

	model.cursor = 0
	model.toggleCurrent()
	if got, want := rowIDs(model.rows), []string{"root", "child", "grandchild", "other"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after expand = %v, want %v", got, want)
	}
}

func TestTaskListRendersAndNavigatesHierarchy(t *testing.T) {
	model := NewTaskList(testBeans())
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 120, Height: 10})
	if view := model.View().Content; !strings.Contains(view, "- First") || !strings.Contains(view, "    Second") {
		t.Errorf("hierarchy view = %q", view)
	}

	model = updateTaskList(t, model, key("right"))
	if model.rows[model.cursor].bean.ID != "project-b" {
		t.Errorf("selected bean after right = %q, want project-b", model.rows[model.cursor].bean.ID)
	}
	model = updateTaskList(t, model, key("left"))
	if model.rows[model.cursor].bean.ID != "project-a" {
		t.Errorf("selected bean after left = %q, want project-a", model.rows[model.cursor].bean.ID)
	}

	model = updateTaskList(t, model, key("h"))
	if got, want := rowIDs(model.rows), []string{"project-a", "project-c"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after collapse = %v, want %v", got, want)
	}
	if view := model.View().Content; !strings.Contains(view, "+ First") {
		t.Errorf("collapsed hierarchy view = %q", view)
	}

	model = updateTaskList(t, model, key("l"))
	if got, want := rowIDs(model.rows), []string{"project-a", "project-b", "project-c"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after expand = %v, want %v", got, want)
	}
}

func TestTaskListNavigatesNestedHierarchy(t *testing.T) {
	model := NewTaskList([]beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "child", Title: "Child", Parent: "root"},
		{ID: "grandchild", Title: "Grandchild", Parent: "child"},
		{ID: "sibling", Title: "Sibling", Parent: "root"},
	})

	model = updateTaskList(t, model, key("right"))
	if selected := model.rows[model.cursor].bean.ID; selected != "child" {
		t.Errorf("selected bean after right = %q, want child", selected)
	}
	model = updateTaskList(t, model, key("right"))
	if selected := model.rows[model.cursor].bean.ID; selected != "grandchild" {
		t.Errorf("selected bean after nested right = %q, want grandchild", selected)
	}
	model = updateTaskList(t, model, key("left"))
	if selected := model.rows[model.cursor].bean.ID; selected != "child" {
		t.Errorf("selected bean after left = %q, want child", selected)
	}
	model = updateTaskList(t, model, key("left"))
	if got, want := rowIDs(model.rows), []string{"root", "child", "sibling"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after collapsing child = %v, want %v", got, want)
	}
	model = updateTaskList(t, model, key("right"))
	if got, want := rowIDs(model.rows), []string{"root", "child", "grandchild", "sibling"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after expanding child = %v, want %v", got, want)
	}
}

func TestTaskListReloadPreservesSelection(t *testing.T) {
	model := NewTaskList(testBeans(), WithTaskLoader(func() ([]beans.Bean, error) {
		return []beans.Bean{
			{ID: "project-c", Title: "Third", Status: "todo", Priority: "normal", Type: "task"},
			{ID: "project-a", Title: "First", Status: "todo", Priority: "normal", Type: "task"},
			{ID: "project-b", Title: "Second", Status: "todo", Priority: "normal", Type: "task", Parent: "project-a"},
		}, nil
	}))
	model = updateTaskList(t, model, key("down"))
	updated, command := model.Update(key("r"))
	if command == nil {
		t.Fatal("reload command is nil")
	}
	list, ok := updated.(TaskList)
	if !ok {
		t.Fatalf("updated model = %T, want TaskList", updated)
	}
	model = updateTaskList(t, list, command())
	if selected := model.rows[model.cursor].bean.ID; selected != "project-b" {
		t.Errorf("selected bean after reload = %q, want project-b", selected)
	}
	if view := model.View().Content; !strings.Contains(view, "r reload") {
		t.Errorf("reload help missing from view = %q", view)
	}
}

func TestTaskListReloadExpandsNewAncestorsAndFallsBackForMissingSelection(t *testing.T) {
	model := NewTaskList([]beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "child", Title: "Child", Parent: "root"},
		{ID: "selected", Title: "Selected"},
	})
	model.collapsed["root"] = true
	model.rebuildRows()
	model.cursor = 1
	model = updateTaskList(t, model, taskListLoadedMessage{beans: []beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "child", Title: "Child", Parent: "root"},
		{ID: "selected", Title: "Selected", Parent: "root"},
	}})
	if got, want := rowIDs(model.rows), []string{"root", "child", "selected"}; !equalStringSlices(got, want) {
		t.Errorf("visible row IDs after reload = %v, want %v", got, want)
	}
	if selected := model.rows[model.cursor].bean.ID; selected != "selected" {
		t.Errorf("selected bean after reload = %q, want selected", selected)
	}

	model = updateTaskList(t, model, taskListLoadedMessage{beans: []beans.Bean{
		{ID: "root", Title: "Root"},
		{ID: "child", Title: "Child", Parent: "root"},
	}})
	if model.cursor != 1 || model.rows[model.cursor].bean.ID != "child" {
		t.Errorf("selection after selected bean disappears = %q at %d, want child at 1", model.rows[model.cursor].bean.ID, model.cursor)
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
	if value == "left" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyLeft})
	}
	if value == "right" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyRight})
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

func rowIDs(rows []taskRow) []string {
	ids := make([]string, len(rows))
	for index, row := range rows {
		ids[index] = row.bean.ID
	}
	return ids
}

func rowDepths(rows []taskRow) []int {
	depths := make([]int, len(rows))
	for index, row := range rows {
		depths[index] = row.depth
	}
	return depths
}

func equalIntSlices(left, right []int) bool {
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

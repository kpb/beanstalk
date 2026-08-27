package tui

import (
	"errors"
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

func TestTaskListRendersSplitDetailPaneAndHelp(t *testing.T) {
	loaded := testBeans()
	loaded[0].Body = "Selected task body"
	model := NewTaskList(loaded,
		WithTaskClaimer(func(string) error { return nil }),
		WithStatusUpdater(func(string, string) error { return nil }),
	)
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: splitPaneWidth, Height: 18})
	view := model.View().Content
	for _, want := range []string{"Tasks (3)", "Task details", "Selected task body", " | "} {
		if !strings.Contains(view, want) {
			t.Errorf("split view does not contain %q:\n%s", want, view)
		}
	}

	model = updateTaskList(t, model, key("?"))
	for _, want := range []string{"Keyboard help", "tab or enter", "c  claim selected todo task", "s  change selected task status", "enter save"} {
		if view := model.View().Content; !strings.Contains(view, want) {
			t.Errorf("help view does not contain %q:\n%s", want, view)
		}
	}
}

func TestTaskListTogglesDetailOnNarrowTerminal(t *testing.T) {
	loaded := testBeans()
	loaded[0].Body = "Selected task body"
	model := NewTaskList(loaded)
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: splitPaneWidth - 1, Height: 18})
	if view := model.View().Content; strings.Contains(view, "Task details") {
		t.Errorf("narrow list view unexpectedly shows detail pane:\n%s", view)
	}

	model = updateTaskList(t, model, key("tab"))
	if view := model.View().Content; !strings.Contains(view, "Task details") || !strings.Contains(view, "Selected task body") {
		t.Errorf("narrow detail view = %q", view)
	}
	model = updateTaskList(t, model, key("enter"))
	if view := model.View().Content; strings.Contains(view, "Task details") {
		t.Errorf("narrow list view after return = %q", view)
	}
}

func TestTaskListShowsMutationFeedbackInNarrowDetailView(t *testing.T) {
	reloadError := errors.New("disk unavailable")
	model := NewTaskList(testBeans(),
		WithTaskClaimer(func(string) error { return nil }),
		WithTaskLoader(func() ([]beans.Bean, error) { return nil, reloadError }),
	)
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: splitPaneWidth - 1, Height: 18})
	model = updateTaskList(t, model, key("tab"))
	updated, command := model.Update(key("c"))
	if command == nil {
		t.Fatal("claim command is nil")
	}
	updated, command = updated.(TaskList).Update(command())
	if command == nil {
		t.Fatal("reload command is nil")
	}
	model = updateTaskList(t, updated.(TaskList), command())
	for _, want := range []string{"Claimed project-a", "Reload failed: disk unavailable"} {
		if view := model.View().Content; !strings.Contains(view, want) {
			t.Errorf("detail view does not contain %q:\n%s", want, view)
		}
	}
}

func TestTaskListChangesStatusAndReloadsPreservingSelection(t *testing.T) {
	updatedStatus := ""
	model := NewTaskList(testBeans(),
		WithStatusUpdater(func(id, status string) error {
			if id != "project-b" {
				t.Errorf("updated ID = %q, want project-b", id)
			}
			updatedStatus = status
			return nil
		}),
		WithTaskLoader(func() ([]beans.Bean, error) {
			loaded := testBeans()
			loaded[1].Status = updatedStatus
			return loaded, nil
		}),
	)
	model = updateTaskList(t, model, key("down"))
	model = updateTaskList(t, model, key("s"))
	if view := model.View().Content; !strings.Contains(view, "Change task status") || !strings.Contains(view, "> todo") {
		t.Errorf("status picker view = %q", view)
	}
	model = updateTaskList(t, model, key("down"))
	updated, command := model.Update(key("enter"))
	if command == nil {
		t.Fatal("status update command is nil")
	}
	list := updated.(TaskList)
	updated, command = list.Update(command())
	if command == nil {
		t.Fatal("reload command is nil")
	}
	model = updateTaskList(t, updated.(TaskList), command())
	if updatedStatus != "in-progress" {
		t.Errorf("updated status = %q, want in-progress", updatedStatus)
	}
	if model.showStatus || model.rows[model.cursor].bean.ID != "project-b" || model.rows[model.cursor].bean.Status != "in-progress" {
		t.Errorf("model after status update = %#v", model)
	}
}

func TestTaskListPreservesDraftStatusInPicker(t *testing.T) {
	updatedStatus := ""
	model := NewTaskList([]beans.Bean{{ID: "project-draft", Title: "Draft", Status: "draft", Priority: "normal", Type: "task"}},
		WithStatusUpdater(func(id, status string) error {
			if id != "project-draft" {
				t.Errorf("updated ID = %q, want project-draft", id)
			}
			updatedStatus = status
			return nil
		}),
	)

	model = updateTaskList(t, model, key("s"))
	if view := model.View().Content; !strings.Contains(view, "> draft") {
		t.Errorf("status picker view = %q", view)
	}
	updated, command := model.Update(key("enter"))
	if command == nil {
		t.Fatal("status update command is nil")
	}
	updated.(TaskList).Update(command())
	if updatedStatus != "draft" {
		t.Errorf("updated status = %q, want draft", updatedStatus)
	}
}

func TestTaskListClaimsTodoTaskAndReloadsPreservingSelection(t *testing.T) {
	claimedID := ""
	model := NewTaskList(testBeans(),
		WithTaskClaimer(func(id string) error {
			claimedID = id
			return nil
		}),
		WithTaskLoader(func() ([]beans.Bean, error) {
			loaded := testBeans()
			loaded[1].Status = "in-progress"
			return loaded, nil
		}),
	)
	model = updateTaskList(t, model, key("down"))
	updated, command := model.Update(key("c"))
	if command == nil {
		t.Fatal("claim command is nil")
	}
	updated, command = updated.(TaskList).Update(command())
	if command == nil {
		t.Fatal("reload command is nil")
	}
	model = updateTaskList(t, updated.(TaskList), command())
	if claimedID != "project-b" {
		t.Errorf("claimed ID = %q, want project-b", claimedID)
	}
	if model.rows[model.cursor].bean.ID != "project-b" || model.rows[model.cursor].bean.Status != "in-progress" {
		t.Errorf("model after claim = %#v", model)
	}
	if view := model.View().Content; !strings.Contains(view, "Claimed project-b") {
		t.Errorf("claim success view = %q", view)
	}
}

func TestTaskListHandlesClaimFailuresWithoutQuitting(t *testing.T) {
	model := NewTaskList(testBeans(), WithTaskClaimer(func(string) error {
		return beans.ErrBeanNotClaimable
	}))
	updated, command := model.Update(key("c"))
	if command == nil {
		t.Fatal("claim command is nil")
	}
	model = updateTaskList(t, updated.(TaskList), command())
	if view := model.View().Content; !strings.Contains(view, "Claim failed: bean is not available to claim") {
		t.Errorf("claim failure view = %q", view)
	}
}

func TestTaskListOnlyClaimsTodoTasks(t *testing.T) {
	claims := 0
	loaded := testBeans()
	loaded[0].Status = "in-progress"
	model := NewTaskList(loaded, WithTaskClaimer(func(string) error {
		claims++
		return nil
	}))
	updated, command := model.Update(key("c"))
	if command != nil {
		t.Fatal("claim command is not nil for an in-progress task")
	}
	model = updated.(TaskList)
	if claims != 0 || !strings.Contains(model.View().Content, "Only todo tasks can be claimed") {
		t.Errorf("model after in-progress claim = %#v, claims = %d", model, claims)
	}
}

func TestTaskListKeepsStatusPickerOpenAfterUpdateFailure(t *testing.T) {
	updateError := errors.New("disk unavailable")
	model := NewTaskList(testBeans(), WithStatusUpdater(func(string, string) error {
		return updateError
	}))
	model = updateTaskList(t, model, key("s"))
	updated, command := model.Update(key("enter"))
	if command == nil {
		t.Fatal("status update command is nil")
	}
	model = updateTaskList(t, updated.(TaskList), command())
	if !model.showStatus || !errors.Is(model.statusErr, updateError) {
		t.Errorf("model after status update failure = %#v", model)
	}
	if view := model.View().Content; !strings.Contains(view, "Update failed: disk unavailable") {
		t.Errorf("status failure view = %q", view)
	}
}

func TestTaskListReportsReloadFailuresAfterMutations(t *testing.T) {
	reloadError := errors.New("disk unavailable")

	t.Run("status update", func(t *testing.T) {
		model := NewTaskList(testBeans(),
			WithStatusUpdater(func(string, string) error { return nil }),
			WithTaskLoader(func() ([]beans.Bean, error) { return nil, reloadError }),
		)
		model = updateTaskList(t, model, key("s"))
		updated, command := model.Update(key("enter"))
		if command == nil {
			t.Fatal("status update command is nil")
		}
		updated, command = updated.(TaskList).Update(command())
		if command == nil {
			t.Fatal("reload command is nil")
		}
		model = updateTaskList(t, updated.(TaskList), command())
		if model.showStatus || !errors.Is(model.reloadErr, reloadError) {
			t.Errorf("model after reload failure = %#v", model)
		}
		if view := model.View().Content; !strings.Contains(view, "Reload failed: disk unavailable") {
			t.Errorf("reload failure view = %q", view)
		}
	})

	t.Run("claim", func(t *testing.T) {
		model := NewTaskList(testBeans(),
			WithTaskClaimer(func(string) error { return nil }),
			WithTaskLoader(func() ([]beans.Bean, error) { return nil, reloadError }),
		)
		updated, command := model.Update(key("c"))
		if command == nil {
			t.Fatal("claim command is nil")
		}
		updated, command = updated.(TaskList).Update(command())
		if command == nil {
			t.Fatal("reload command is nil")
		}
		model = updateTaskList(t, updated.(TaskList), command())
		if !errors.Is(model.reloadErr, reloadError) || model.claimMessage != "Claimed project-a" {
			t.Errorf("model after reload failure = %#v", model)
		}
		if view := model.View().Content; !strings.Contains(view, "Reload failed: disk unavailable") {
			t.Errorf("reload failure view = %q", view)
		}
	})
}

func TestTaskListRendersStatusPickerOnShortTerminals(t *testing.T) {
	model := NewTaskList(testBeans(), WithStatusUpdater(func(string, string) error {
		return nil
	}))
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 40, Height: fixedLines})
	model = updateTaskList(t, model, key("s"))
	if view := model.View().Content; !strings.Contains(view, "Change status") || !strings.Contains(view, "> todo") {
		t.Errorf("compact status picker view = %q", view)
	}
	model = updateTaskList(t, model, key("down"))
	if view := model.View().Content; !strings.Contains(view, "> in-progress") {
		t.Errorf("compact status picker after selection = %q", view)
	}
}

func TestTaskListReservesSpaceForClaimFeedback(t *testing.T) {
	loaded := append(testBeans(),
		beans.Bean{ID: "project-d", Title: "Fourth"},
		beans.Bean{ID: "project-e", Title: "Fifth"},
		beans.Bean{ID: "project-f", Title: "Sixth"},
	)
	model := NewTaskList(loaded, WithTaskClaimer(func(string) error { return nil }))
	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 200, Height: 8})
	model.claimMessage = "Claimed project-a"
	if view := model.View().Content; !strings.Contains(view, "Claimed project-a") || !strings.Contains(view, "q quit") {
		t.Errorf("split view with claim feedback = %q", view)
	}

	model = updateTaskList(t, model, tea.WindowSizeMsg{Width: 40, Height: 3})
	if view := model.View().Content; !strings.Contains(view, "Claimed project-a") || !strings.Contains(view, "q quit") {
		t.Errorf("compact view with claim feedback = %q", view)
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
	if value == "tab" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyTab})
	}
	if value == "enter" {
		return tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	}
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

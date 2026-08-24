// Package tui provides Beanstalk's interactive terminal views.
package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/kpb/beanstalk/internal/beans"
)

const (
	fixedLines         = 5
	defaultVisibleRows = 5
)

// TaskList displays a read-only, keyboard-navigable list of beans.
type TaskList struct {
	beans     []beans.Bean
	rows      []taskRow
	collapsed map[string]bool
	children  map[string]bool
	load      func() ([]beans.Bean, error)
	reloadErr error
	cursor    int
	offset    int
	width     int
	height    int
}

type taskRow struct {
	bean  beans.Bean
	depth int
}

type taskListLoadedMessage struct {
	beans []beans.Bean
	err   error
}

// TaskListOption configures a task-list model.
type TaskListOption func(*TaskList)

// WithTaskLoader enables reloading tasks from the TUI.
func WithTaskLoader(load func() ([]beans.Bean, error)) TaskListOption {
	return func(model *TaskList) {
		model.load = load
	}
}

// NewTaskList constructs a task-list model from already-loaded beans.
func NewTaskList(loaded []beans.Bean, options ...TaskListOption) TaskList {
	model := TaskList{beans: loaded, collapsed: make(map[string]bool)}
	for _, option := range options {
		option(&model)
	}
	model.rebuildRows()
	return model
}

func (m TaskList) Init() tea.Cmd {
	return nil
}

func (m TaskList) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		m.clamp()
	case taskListLoadedMessage:
		if message.err != nil {
			m.reloadErr = message.err
			return m, nil
		}
		m.reloadErr = nil
		m.replaceBeans(message.beans)
	case tea.KeyPressMsg:
		switch message.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.cursor--
		case "down", "j":
			m.cursor++
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.rows) - 1
		case "left", "h":
			if !m.collapseCurrent() {
				m.selectParent()
			}
		case "right", "l":
			if !m.expandCurrent() {
				m.selectFirstChild()
			}
		case "r":
			if m.load != nil {
				return m, loadTasks(m.load)
			}
		}
		m.clamp()
	}
	return m, nil
}

func (m TaskList) View() tea.View {
	view := tea.NewView(m.render())
	view.AltScreen = true
	view.WindowTitle = "Beanstalk"
	return view
}

func (m *TaskList) clamp() {
	if len(m.rows) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.rows)-1))
	rows := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	m.offset = max(0, min(m.offset, max(0, len(m.rows)-rows)))
}

func (m TaskList) visibleRows() int {
	if m.height <= 0 {
		return min(defaultVisibleRows, len(m.rows))
	}
	if m.height <= fixedLines {
		return 1
	}
	return max(1, m.height-fixedLines)
}

func (m TaskList) render() string {
	if m.height > 0 && m.height <= fixedLines {
		return m.compactView()
	}

	var output strings.Builder
	fmt.Fprintf(&output, "Beanstalk tasks (%d)\n\n", len(m.beans))
	if len(m.beans) == 0 {
		output.WriteString("No beans found.\n\nq quit\n")
		return output.String()
	}

	output.WriteString("  ID                 STATUS       PRIORITY  TYPE      PARENT        TITLE\n")
	end := min(len(m.rows), m.offset+m.visibleRows())
	for index := m.offset; index < end; index++ {
		row := m.rows[index]
		bean := row.bean
		parent := bean.Parent
		if parent == "" {
			parent = "-"
		}
		marker := " "
		if index == m.cursor {
			marker = ">"
		}
		line := fmt.Sprintf("%s %-18s %-12s %-9s %-9s %-13s %s", marker, bean.ID, bean.Status, bean.Priority, bean.Type, parent, treeTitle(row, m.children, m.collapsed))
		output.WriteString(truncate(line, m.width))
		output.WriteByte('\n')
	}
	if m.reloadErr != nil {
		fmt.Fprintf(&output, "\nReload failed: %v\n", m.reloadErr)
	}
	help := "j/k navigate  h/l or left/right collapse/expand  g/G first/last"
	if m.load != nil {
		help += "  r reload"
	}
	output.WriteString("\n" + help + "  q quit\n")
	return output.String()
}

func (m TaskList) compactView() string {
	lines := []string{fmt.Sprintf("Beanstalk tasks (%d)", len(m.beans))}
	if len(m.beans) == 0 {
		lines = append(lines, "No beans found.")
	} else {
		row := m.rows[m.cursor]
		lines = append(lines, truncate(fmt.Sprintf("> %s  %s  %s", row.bean.ID, row.bean.Status, treeTitle(row, m.children, m.collapsed)), m.width))
	}
	if m.height >= 3 {
		lines = append(lines, "q quit")
	}
	return strings.Join(lines[:min(len(lines), m.height)], "\n") + "\n"
}

func (m *TaskList) rebuildRows() {
	m.rows, m.children = flattenTaskTree(m.beans, m.collapsed)
	m.clamp()
}

func (m *TaskList) replaceBeans(loaded []beans.Bean) {
	selectedID := ""
	if len(m.rows) > 0 {
		selectedID = m.rows[m.cursor].bean.ID
	}
	m.beans = loaded
	m.expandAncestors(selectedID)
	m.rebuildRows()
	for index, row := range m.rows {
		if row.bean.ID == selectedID {
			m.cursor = index
			m.clamp()
			return
		}
	}
}

func (m *TaskList) expandAncestors(id string) {
	byID := make(map[string]beans.Bean, len(m.beans))
	for _, bean := range m.beans {
		byID[bean.ID] = bean
	}
	visited := make(map[string]bool)
	for id != "" && !visited[id] {
		visited[id] = true
		bean, found := byID[id]
		if !found {
			return
		}
		if bean.Parent != "" {
			m.collapsed[bean.Parent] = false
		}
		id = bean.Parent
	}
}

func loadTasks(load func() ([]beans.Bean, error)) tea.Cmd {
	return func() tea.Msg {
		loaded, err := load()
		return taskListLoadedMessage{beans: loaded, err: err}
	}
}

func (m *TaskList) toggleCurrent() {
	if len(m.rows) == 0 {
		return
	}
	bean := m.rows[m.cursor].bean
	if !m.children[bean.ID] {
		return
	}
	m.collapsed[bean.ID] = !m.collapsed[bean.ID]
	m.rebuildRows()
}

func (m *TaskList) collapseCurrent() bool {
	if len(m.rows) == 0 {
		return false
	}
	id := m.rows[m.cursor].bean.ID
	if !m.children[id] || m.collapsed[id] {
		return false
	}
	m.collapsed[id] = true
	m.rebuildRows()
	return true
}

func (m *TaskList) expandCurrent() bool {
	if len(m.rows) == 0 {
		return false
	}
	id := m.rows[m.cursor].bean.ID
	if !m.children[id] || !m.collapsed[id] {
		return false
	}
	m.collapsed[id] = false
	m.rebuildRows()
	return true
}

func (m *TaskList) selectParent() {
	if len(m.rows) == 0 || m.rows[m.cursor].depth == 0 {
		return
	}
	depth := m.rows[m.cursor].depth
	for index := m.cursor - 1; index >= 0; index-- {
		if m.rows[index].depth < depth {
			m.cursor = index
			return
		}
	}
}

func (m *TaskList) selectFirstChild() {
	if m.cursor+1 < len(m.rows) && m.rows[m.cursor+1].depth > m.rows[m.cursor].depth {
		m.cursor++
	}
}

func treeTitle(row taskRow, children, collapsed map[string]bool) string {
	marker := " "
	if children[row.bean.ID] {
		if collapsed[row.bean.ID] {
			marker = "+"
		} else {
			marker = "-"
		}
	}
	return strings.Repeat("  ", row.depth) + marker + " " + row.bean.Title
}

func flattenTaskTree(loaded []beans.Bean, collapsed map[string]bool) ([]taskRow, map[string]bool) {
	byID := make(map[string]int, len(loaded))
	for index, bean := range loaded {
		if _, found := byID[bean.ID]; !found {
			byID[bean.ID] = index
		}
	}

	children := make(map[string][]int, len(loaded))
	roots := make([]int, 0, len(loaded))
	for index, bean := range loaded {
		if bean.Parent == "" {
			roots = append(roots, index)
			continue
		}
		if _, found := byID[bean.Parent]; !found {
			roots = append(roots, index)
			continue
		}
		children[bean.Parent] = append(children[bean.Parent], index)
	}

	hasChildren := make(map[string]bool, len(children))
	for parent := range children {
		hasChildren[parent] = true
	}
	rows := make([]taskRow, 0, len(loaded))
	visited := make([]bool, len(loaded))
	var visit func(int, int, bool)
	visit = func(index, depth int, visible bool) {
		if visited[index] {
			return
		}
		visited[index] = true
		bean := loaded[index]
		if visible {
			rows = append(rows, taskRow{bean: bean, depth: depth})
		}
		for _, child := range children[bean.ID] {
			visit(child, depth+1, visible && !collapsed[bean.ID])
		}
	}
	for _, root := range roots {
		visit(root, 0, true)
	}
	for index := range loaded {
		visit(index, 0, true)
	}
	return rows, hasChildren
}

func truncate(value string, width int) string {
	runes := []rune(value)
	if width <= 0 || len(runes) <= width {
		return value
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

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
	beans  []beans.Bean
	cursor int
	offset int
	width  int
	height int
}

// NewTaskList constructs a task-list model from already-loaded beans.
func NewTaskList(loaded []beans.Bean) TaskList {
	return TaskList{beans: loaded}
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
			m.cursor = len(m.beans) - 1
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
	if len(m.beans) == 0 {
		m.cursor = 0
		m.offset = 0
		return
	}
	m.cursor = max(0, min(m.cursor, len(m.beans)-1))
	rows := m.visibleRows()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+rows {
		m.offset = m.cursor - rows + 1
	}
	m.offset = max(0, min(m.offset, max(0, len(m.beans)-rows)))
}

func (m TaskList) visibleRows() int {
	if m.height <= 0 {
		return min(defaultVisibleRows, len(m.beans))
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
	end := min(len(m.beans), m.offset+m.visibleRows())
	for index := m.offset; index < end; index++ {
		bean := m.beans[index]
		parent := bean.Parent
		if parent == "" {
			parent = "-"
		}
		marker := " "
		if index == m.cursor {
			marker = ">"
		}
		line := fmt.Sprintf("%s %-18s %-12s %-9s %-9s %-13s %s", marker, bean.ID, bean.Status, bean.Priority, bean.Type, parent, bean.Title)
		output.WriteString(truncate(line, m.width))
		output.WriteByte('\n')
	}
	output.WriteString("\nj/k or arrows navigate  g/G first/last  q quit\n")
	return output.String()
}

func (m TaskList) compactView() string {
	lines := []string{fmt.Sprintf("Beanstalk tasks (%d)", len(m.beans))}
	if len(m.beans) == 0 {
		lines = append(lines, "No beans found.")
	} else {
		bean := m.beans[m.cursor]
		lines = append(lines, truncate(fmt.Sprintf("> %s  %s  %s", bean.ID, bean.Status, bean.Title), m.width))
	}
	if m.height >= 3 {
		lines = append(lines, "q quit")
	}
	return strings.Join(lines[:min(len(lines), m.height)], "\n") + "\n"
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

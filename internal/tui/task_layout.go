package tui

import (
	"fmt"
	"strings"
)

func (m TaskList) usesSplitPane() bool {
	return m.width >= splitPaneWidth && m.height > fixedLines && len(m.rows) > 0
}

func (m TaskList) splitView() string {
	available := m.width - 3 // Space on both sides of the pane separator.
	treeWidth := available * 2 / 5
	detailWidth := available - treeWidth
	return joinPanes(m.treePane(treeWidth), detailLines(m), treeWidth, detailWidth)
}

func (m TaskList) treePane(width int) []string {
	lines := []string{fmt.Sprintf("Tasks (%d)", len(m.beans)), ""}
	rows := max(1, m.height-4)
	end := min(len(m.rows), m.offset+rows)
	for index := m.offset; index < end; index++ {
		row := m.rows[index]
		marker := " "
		if index == m.cursor {
			marker = ">"
		}
		lines = append(lines, truncate(fmt.Sprintf("%s %s  %s  %s", marker, row.bean.ID, row.bean.Status, treeTitle(row, m.children, m.collapsed)), width))
	}
	if m.reloadErr != nil {
		lines = append(lines, "", truncate(fmt.Sprintf("Reload failed: %v", m.reloadErr), width))
	}
	lines = append(lines, "", "j/k navigate  h/l tree  ? help  q quit")
	return fitPaneLines(lines, width, m.height)
}

func detailLines(m TaskList) []string {
	available := m.width - 3
	treeWidth := available * 2 / 5
	detailWidth := available - treeWidth
	return fitPaneLines(strings.Split(strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, detailWidth, m.height), "\n"), "\n"), detailWidth, m.height)
}

func joinPanes(left, right []string, leftWidth, rightWidth int) string {
	lines := make([]string, max(len(left), len(right)))
	for index := range lines {
		leftLine := ""
		if index < len(left) {
			leftLine = truncate(left[index], leftWidth)
		}
		rightLine := ""
		if index < len(right) {
			rightLine = truncate(right[index], rightWidth)
		}
		lines[index] = leftLine + strings.Repeat(" ", max(0, leftWidth-len([]rune(leftLine)))) + " | " + rightLine
	}
	return strings.Join(lines, "\n") + "\n"
}

func fitPaneLines(lines []string, width, height int) []string {
	for index := range lines {
		lines[index] = truncate(lines[index], width)
	}
	if height <= 0 {
		return lines
	}
	if len(lines) > height {
		return lines[:height]
	}
	return append(lines, make([]string, height-len(lines))...)
}

func (m TaskList) helpView() string {
	lines := []string{
		"Keyboard help",
		"",
		"j/k or up/down  move selection",
		"h/l or left/right  collapse, expand, parent, child",
		"g/G or home/end  first or last task",
		"tab or enter  show selected task on narrow terminals",
		"r  reload tasks",
		"c  claim task (available in a later workflow step)",
		"s  change status (available in a later workflow step)",
		"?  close help",
		"q or ctrl+c  quit",
	}
	return boundDetail(lines, m.width, m.height)
}

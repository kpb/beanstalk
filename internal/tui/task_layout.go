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
	notices := m.notices()
	rows := max(1, m.height-3-len(notices))
	end := min(len(m.rows), m.offset+rows)
	for index := m.offset; index < end; index++ {
		row := m.rows[index]
		marker := " "
		if index == m.cursor {
			marker = ">"
		}
		lines = append(lines, truncate(fmt.Sprintf("%s %s  %s  %s", marker, row.bean.ID, row.bean.Status, treeTitle(row, m.children, m.collapsed)), width))
	}
	for _, notice := range notices {
		lines = append(lines, truncate(notice, width))
	}
	help := "j/k navigate  h/l tree"
	if m.load != nil {
		help += "  " + m.archiveToggleLabel()
	}
	if m.claim != nil {
		help += "  c claim"
	}
	if m.updateStatus != nil {
		help += "  s status"
	}
	lines = append(lines, help+"  ? help  q quit")
	return fitPaneLines(lines, width, m.height)
}

func (m TaskList) detailView() string {
	notices := m.notices()
	if len(notices) == 0 {
		return renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, m.height)
	}
	details := strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, 0), "\n")
	lines := append(notices, "")
	lines = append(lines, strings.Split(details, "\n")...)
	return boundDetail(lines, m.width, m.height)
}

func (m TaskList) notices() []string {
	notices := make([]string, 0, 2)
	if m.reloadErr != nil {
		notices = append(notices, fmt.Sprintf("Reload failed: %v", m.reloadErr))
	}
	if m.claimMessage != "" {
		notices = append(notices, m.claimMessage)
	}
	return notices
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
	}
	if m.load != nil {
		lines = append(lines, "a  "+m.archiveToggleLabel()[2:])
	}
	if m.claim != nil {
		lines = append(lines, "c  claim selected todo task")
	}
	if m.updateStatus != nil {
		lines = append(lines, "s  change selected task status", "    j/k select  enter save  esc cancel")
	}
	lines = append(lines, "?  close help", "q or ctrl+c  quit")
	return boundDetail(lines, m.width, m.height)
}

func (m TaskList) statusPickerView() string {
	selected := m.rows[m.cursor].bean
	if m.height > 0 && m.height <= fixedLines {
		lines := []string{
			"Change status",
			"> " + statuses[m.statusCursor],
			"j/k select  enter save",
			"esc cancel",
		}
		if m.statusErr != nil {
			lines = append(lines, fmt.Sprintf("Update failed: %v", m.statusErr))
		}
		return boundDetail(lines, m.width, m.height)
	}
	lines := []string{"Change task status", "", selected.ID + " " + selected.Title, ""}
	for index, status := range statuses {
		marker := " "
		if index == m.statusCursor {
			marker = ">"
		}
		lines = append(lines, marker+" "+status)
	}
	if m.statusErr != nil {
		lines = append(lines, "", fmt.Sprintf("Update failed: %v", m.statusErr))
	}
	lines = append(lines, "", "j/k select  enter save  esc cancel")
	return boundDetail(lines, m.width, m.height)
}

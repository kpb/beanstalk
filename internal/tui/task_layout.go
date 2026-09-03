package tui

import (
	"fmt"
	"strings"
)

func (m TaskList) usesSplitPane() bool {
	return m.width >= splitPaneWidth && m.height > fixedLines && len(m.rows) > 0
}

func (m TaskList) splitView() string {
	available := m.width - 1 // Space between bordered panes.
	treeWidth := available * 2 / 5
	detailWidth := available - treeWidth
	paneHeight := max(3, m.height-1)
	selected := m.rows[m.cursor].bean
	panes := joinPanes(
		borderedPane(fmt.Sprintf("Tasks (%d)", len(m.beans)), m.treePane(treeWidth-2, paneHeight-2), treeWidth, paneHeight),
		borderedPane(selected.ID, detailLines(m, detailWidth-2, paneHeight-2), detailWidth, paneHeight),
		treeWidth,
		detailWidth,
	)
	return panes + muted(truncate(m.splitShortcutHelp(), m.width)) + "\n"
}

func (m TaskList) treePane(width, height int) []string {
	lines := make([]string, 0, height)
	notices := m.notices()
	rows := max(1, height-len(notices))
	end := min(len(m.rows), m.offset+rows)
	for index := m.offset; index < end; index++ {
		row := m.rows[index]
		lines = append(lines, splitTaskRowView(row, m.children, m.collapsed, index == m.cursor, width))
	}
	for _, notice := range notices {
		lines = append(lines, truncate(feedbackMessage(notice), width))
	}
	return fitPaneLines(lines, width, height)
}

func (m TaskList) splitShortcutHelp() string {
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
	return help + "  ? help  q quit"
}

func (m TaskList) detailView() string {
	notices := m.notices()
	if len(notices) == 0 {
		return styleDetail(renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, m.height))
	}
	details := strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, 0), "\n")
	lines := append(notices, "")
	lines = append(lines, strings.Split(details, "\n")...)
	return styleDetail(boundDetail(lines, m.width, m.height))
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

func detailLines(m TaskList, width, height int) []string {
	lines := strings.Split(strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, width, 0), "\n"), "\n")
	// The selected task ID is already the pane title, so omit the repeated detail heading.
	lines = lines[2:]
	lines = fitPaneLines(lines, width, height)
	return strings.Split(strings.TrimSuffix(styleDetail(strings.Join(lines, "\n")+"\n"), "\n"), "\n")
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
		lines[index] = leftLine + strings.Repeat(" ", max(0, leftWidth-displayWidth(leftLine))) + " " + rightLine
	}
	return strings.Join(lines, "\n") + "\n"
}

func borderedPane(title string, content []string, width, height int) []string {
	width = max(4, width)
	height = max(3, height)
	innerWidth := width - 2
	innerHeight := height - 2
	title = truncate(" "+title+" ", innerWidth)
	lines := make([]string, 0, height)
	lines = append(lines, "╭"+title+strings.Repeat("─", max(0, innerWidth-displayWidth(title)))+"╮")
	for index := 0; index < innerHeight; index++ {
		line := ""
		if index < len(content) {
			line = truncate(content[index], innerWidth)
		}
		lines = append(lines, "│"+line+strings.Repeat(" ", max(0, innerWidth-displayWidth(line)))+"│")
	}
	return append(lines, "╰"+strings.Repeat("─", innerWidth)+"╯")
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
	return styleDetail(boundDetail(lines, m.width, m.height))
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
		return styleDetail(boundDetail(lines, m.width, m.height))
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
	return styleDetail(boundDetail(lines, m.width, m.height))
}

func splitTaskRowView(row taskRow, children, collapsed map[string]bool, selected bool, width int) string {
	marker := " "
	if selected {
		marker = ">"
	}
	metadata := typeIndicator(row.bean.Type) + " " + row.bean.Status
	titleWidth := max(1, width-displayWidth(metadata)-3)
	title := truncate(treeTitle(row, children, collapsed), titleWidth)
	padding := strings.Repeat(" ", max(1, width-displayWidth(marker)-1-displayWidth(title)-1-displayWidth(metadata)))
	plain := truncate(marker+" "+title+padding+metadata, width)
	plain = strings.Replace(plain, row.bean.Status, statusLabel(row.bean.Status), 1)
	if selected {
		return styled(plain, ansiBold, ansiCyan)
	}
	return plain
}

func typeIndicator(taskType string) string {
	switch taskType {
	case "milestone":
		return "M"
	case "epic":
		return "E"
	case "feature":
		return "F"
	case "task":
		return "T"
	default:
		return "?"
	}
}

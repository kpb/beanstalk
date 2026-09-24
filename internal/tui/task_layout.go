package tui

import (
	"fmt"
	"strings"

	"github.com/kpb/beanstalk/internal/terminal"
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
		borderedPane(terminal.Text(selected.ID), detailLines(m, detailWidth-2, paneHeight-2), detailWidth, paneHeight),
		treeWidth,
		detailWidth,
	)
	return panes + truncate(m.splitShortcutHelp(), m.width) + "\n"
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
	help := shortcut("j/k", "navigate") + "  " + shortcut("h/l", "tree")
	if m.load != nil {
		help += "  " + shortcut("a", m.archiveToggleLabel()[2:])
	}
	if m.claim != nil {
		help += "  " + shortcut("c", "claim")
	}
	if m.updateStatus != nil {
		help += "  " + shortcut("s", "status")
	}
	return help + "  " + shortcut("?", "help") + "  " + shortcut("q", "quit")
}

func (m TaskList) detailView() string {
	lines := m.detailLines()
	styledLines := strings.Split(strings.TrimSuffix(styleDetail(strings.Join(lines, "\n")+"\n"), "\n"), "\n")
	viewportHeight := m.detailViewportHeight()
	start, end := 0, len(styledLines)
	if viewportHeight > 0 {
		start = min(m.detailOffset, max(0, len(styledLines)-viewportHeight))
		end = min(len(styledLines), start+viewportHeight)
	}
	footer := shortcut("j/k", "scroll") + "  " + shortcut("home/end", "top/bottom") + "  " + shortcut("tab/enter/esc", "back")
	if m.updateStatus != nil {
		footer += "  " + shortcut("s", "status")
	}
	if m.claim != nil {
		footer += "  " + shortcut("c", "claim")
	}
	if m.load != nil {
		footer += "  " + shortcut("r", "reload")
	}
	footer += "  " + shortcut("?", "help") + "  " + shortcut("q", "quit")
	if viewportHeight > 0 && len(styledLines) > viewportHeight {
		footer += fmt.Sprintf("  %d-%d/%d", start+1, end, len(styledLines))
	}
	return strings.Join(styledLines[start:end], "\n") + "\n" + truncate(footer, m.width) + "\n"
}

func (m TaskList) detailLines() []string {
	notices := m.notices()
	if len(notices) == 0 {
		return strings.Split(strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, 0), "\n"), "\n")
	}
	details := strings.TrimSuffix(renderTaskDetail(m.beans, m.rows[m.cursor].bean, m.width, 0), "\n")
	lines := append(notices, "")
	lines = append(lines, strings.Split(details, "\n")...)
	return lines
}

func (m TaskList) notices() []string {
	notices := make([]string, 0, 2)
	if m.reloadErr != nil {
		notices = append(notices, terminal.Text(fmt.Sprintf("Reload failed: %v", m.reloadErr)))
	}
	if m.claimMessage != "" {
		notices = append(notices, terminal.Text(m.claimMessage))
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
	if title != "" {
		title = truncate(" "+title+" ", innerWidth)
	}
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
	if m.width <= 0 || m.height <= fixedLines {
		return styleDetail(boundDetail(m.helpLines(), m.width, m.height))
	}

	background := m.helpBackgroundView()
	lines := strings.Split(strings.TrimSuffix(background, "\n"), "\n")
	for len(lines) < m.height {
		lines = append(lines, "")
	}
	lines = lines[:m.height]

	modalWidth := min(72, max(4, m.width-4))
	modalLines := m.paddedHelpLines()
	modalHeight := min(len(modalLines)+2, max(3, m.height-2))
	modalLines = fitHelpLines(modalLines, modalHeight-2)
	modal := styleHelpModal(borderedPane("", modalLines, modalWidth, modalHeight))
	top := max(0, (m.height-len(modal))/2)
	left := max(0, (m.width-modalWidth)/2)
	for index := range modal {
		lines[top+index] = overlayHelpModalLine(lines[top+index], modal[index], left, m.width)
	}
	for index := range lines {
		if index < top || index >= top+len(modal) {
			lines[index] = dimmedBackground(truncate(plainText(lines[index]), m.width))
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func (m TaskList) helpBackgroundView() string {
	if m.usesSplitPane() {
		return m.splitView()
	}
	return m.listView()
}

func (m TaskList) helpLines() []string {
	hints := []helpHint{
		{"j/k or up/down", "move selection (scroll details when open)"},
		{"h/l or left/right", "collapse, expand, parent, child"},
		{"g/G or home/end", "first or last task"},
		{"tab or enter", "show selected task full screen"},
		{"esc", "return from task details"},
		{"home/end", "top or bottom of details when open"},
		{"r", "reload tasks"},
	}
	if m.load != nil {
		hints = append(hints, helpHint{"a", m.archiveToggleLabel()[2:]})
	}
	if m.claim != nil {
		hints = append(hints, helpHint{"c", "claim selected todo task"})
	}
	if m.updateStatus != nil {
		hints = append(hints, helpHint{"s", "change selected task status"})
	}
	keyWidth := helpHintWidth(hints)
	modalWidth := min(72, max(4, m.width-4)) - 2
	if m.width > 0 && modalWidth < keyWidth+25 {
		keyWidth = 0
	}

	lines := []string{styled("Keyboard Shortcuts", ansiBold, ansiMagenta), ""}
	for _, hint := range hints {
		lines = append(lines, formatHelpHint(hint, keyWidth))
	}
	if m.updateStatus != nil {
		lines = append(lines, "    "+shortcut("j/k", "select")+"  "+shortcut("enter", "save")+"  "+shortcut("esc", "cancel"))
	}
	lines = append(lines, formatHelpHint(helpHint{"q or ctrl+c", "quit"}, keyWidth), "", formatHelpHint(helpHint{"? or esc", "close help"}, keyWidth))
	return lines
}

type helpHint struct {
	key         string
	description string
}

func helpHintWidth(hints []helpHint) int {
	width := 0
	for _, hint := range hints {
		width = max(width, displayWidth(hint.key))
	}
	return width
}

func formatHelpHint(hint helpHint, keyWidth int) string {
	return shortcut(hint.key, strings.Repeat(" ", max(0, keyWidth-displayWidth(hint.key)))+hint.description)
}

func (m TaskList) paddedHelpLines() []string {
	lines := append([]string(nil), m.helpLines()...)
	for index := range lines {
		lines[index] = "  " + lines[index]
	}
	return append(lines, "")
}

func fitHelpLines(lines []string, height int) []string {
	if len(lines) <= height {
		return lines
	}
	footer := lines[len(lines)-4 : len(lines)-1]
	if height <= len(footer) {
		return footer[len(footer)-height:]
	}
	lines = append([]string(nil), lines[:height-len(footer)]...)
	return append(lines, footer...)
}

func styleHelpModal(lines []string) []string {
	for index, line := range lines {
		if index == 0 || index == len(lines)-1 {
			lines[index] = styled(line, ansiMagenta)
			continue
		}
		lines[index] = styled("│", ansiMagenta) + strings.TrimSuffix(strings.TrimPrefix(line, "│"), "│") + styled("│", ansiMagenta)
	}
	return lines
}

func overlayHelpModalLine(background, modal string, left, width int) string {
	background = plainText(background)
	if displayWidth(background) < width {
		background += strings.Repeat(" ", width-displayWidth(background))
	}
	modalWidth := displayWidth(modal)
	return dimmedBackground(displaySlice(background, 0, left)) + modal + dimmedBackground(displaySlice(background, left+modalWidth, width))
}

func displaySlice(value string, start, end int) string {
	if end <= start {
		return ""
	}
	var output strings.Builder
	position := 0
	for _, rune := range value {
		if position >= end {
			break
		}
		if position >= start {
			output.WriteRune(rune)
		}
		position++
	}
	return output.String()
}

func (m TaskList) statusPickerView() string {
	selected := m.rows[m.cursor].bean
	if m.height > 0 && m.height <= fixedLines {
		lines := []string{
			"Change status",
			"> " + statuses[m.statusCursor],
			shortcut("j/k", "select") + "  " + shortcut("enter", "save"),
			shortcut("esc", "cancel"),
		}
		if m.statusErr != nil {
			lines = append(lines, terminal.Text(fmt.Sprintf("Update failed: %v", m.statusErr)))
		}
		return styleDetail(boundDetail(lines, m.width, m.height))
	}
	lines := []string{"Change task status", "", terminal.Text(selected.ID) + " " + terminal.Text(selected.Title), ""}
	for index, status := range statuses {
		marker := " "
		if index == m.statusCursor {
			marker = ">"
		}
		lines = append(lines, marker+" "+status)
	}
	if m.statusErr != nil {
		lines = append(lines, "", terminal.Text(fmt.Sprintf("Update failed: %v", m.statusErr)))
	}
	lines = append(lines, "", shortcut("j/k", "select")+"  "+shortcut("enter", "save")+"  "+shortcut("esc", "cancel"))
	return styleDetail(boundDetail(lines, m.width, m.height))
}

func splitTaskRowView(row taskRow, children, collapsed map[string]bool, selected bool, width int) string {
	marker := " "
	if selected {
		marker = ">"
	}
	metadata := typeIndicator(row.bean.Type) + " " + terminal.Text(row.bean.Status)
	titleWidth := max(1, width-displayWidth(metadata)-3)
	title := truncate(terminal.Text(treeTitle(row, children, collapsed)), titleWidth)
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

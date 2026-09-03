package tui

import (
	"fmt"
	"strings"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
)

func styled(value string, attributes ...string) string {
	return strings.Join(attributes, "") + value + ansiReset
}

func heading(value string) string {
	return styled(value, ansiBold, ansiCyan)
}

func muted(value string) string {
	return styled(value, ansiDim)
}

func statusStyle(status string) string {
	switch status {
	case "draft":
		return ansiDim
	case "todo":
		return ansiYellow
	case "in-progress":
		return ansiBlue
	case "completed":
		return ansiGreen
	case "scrapped":
		return ansiMagenta
	default:
		return ""
	}
}

func statusLabel(status string) string {
	if color := statusStyle(status); color != "" {
		return styled(status, color)
	}
	return status
}

func taskRowView(row taskRow, children, collapsed map[string]bool, selected bool, width int) string {
	bean := row.bean
	parent := bean.Parent
	if parent == "" {
		parent = "-"
	}
	marker := " "
	if selected {
		marker = ">"
	}
	plain := truncate(fmt.Sprintf("%s %-18s | %-12s | %-9s | %-9s | %-13s | %s", marker, bean.ID, bean.Status, bean.Priority, bean.Type, parent, treeTitle(row, children, collapsed)), width)
	plain = strings.Replace(plain, bean.Status, statusLabel(bean.Status), 1)
	if selected {
		return styled(plain, ansiBold, ansiCyan)
	}
	return plain
}

func styleDetail(detail string) string {
	lines := strings.Split(strings.TrimSuffix(detail, "\n"), "\n")
	for index, line := range lines {
		switch {
		case index == 0:
			lines[index] = heading(line)
		case index == 2:
			lines[index] = styled(line, ansiBold)
		case strings.HasPrefix(line, "Status: "):
			lines[index] = styled(line, statusStyle(strings.TrimPrefix(line, "Status: ")))
		case strings.HasSuffix(line, ":"):
			lines[index] = styled(line, ansiBold)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

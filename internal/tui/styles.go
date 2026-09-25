package tui

import (
	"fmt"
	"strings"

	"github.com/kpb/beanstalk/internal/terminal"
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
	ansiGray    = "\x1b[90m"
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

func dimmedBackground(value string) string {
	return styled(value, ansiDim, ansiGray)
}

func shortcut(key, description string) string {
	return styled(key, ansiCyan) + muted(" "+description)
}

func plainText(value string) string {
	var output strings.Builder
	for index := 0; index < len(value); {
		if end, found := ansiSequenceEnd(value, index); found {
			index = end
			continue
		}
		rune, size := runeAt(value, index)
		output.WriteRune(rune)
		index += size
	}
	return output.String()
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

func statusCell(status string, width int) string {
	status = terminal.Text(status)
	return statusLabel(status) + strings.Repeat(" ", max(0, width-displayWidth(status)))
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
	plain := truncate(fmt.Sprintf("%s %-18s | %s | %-9s | %-9s | %-13s | %s", marker, terminal.Text(bean.ID), statusCell(bean.Status, 12), terminal.Text(bean.Priority), terminal.Text(bean.Type), terminal.Text(parent), terminal.Text(treeTitle(row, children, collapsed))), width)
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

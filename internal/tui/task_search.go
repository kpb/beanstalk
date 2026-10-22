package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/kpb/beanstalk/internal/beans"
	"github.com/kpb/beanstalk/internal/terminal"
)

func (m *TaskList) refreshSearch() {
	selectedID := ""
	if m.searchCursor < len(m.searchMatches) {
		selectedID = m.searchMatches[m.searchCursor].ID
	}
	m.searchMatches = beans.Search(m.beans, m.searchQuery)
	m.searchCursor = 0
	for index, bean := range m.searchMatches {
		if bean.ID == selectedID {
			m.searchCursor = index
			break
		}
	}
}

func (m TaskList) updateSearch(message tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.showSearch = false
	case "enter":
		if len(m.searchMatches) > 0 {
			id := m.searchMatches[m.searchCursor].ID
			m.expandAncestors(id)
			m.rebuildRows()
			for index, row := range m.rows {
				if row.bean.ID == id {
					m.cursor = index
					m.clamp()
					break
				}
			}
			m.showSearch = false
			m.showDetails = false
		}
	case "up":
		m.searchCursor = max(0, m.searchCursor-1)
	case "down":
		m.searchCursor = min(len(m.searchMatches)-1, m.searchCursor+1)
	case "home":
		m.searchCursor = 0
	case "end":
		m.searchCursor = max(0, len(m.searchMatches)-1)
	case "backspace":
		if len(m.searchQuery) > 0 {
			runes := []rune(m.searchQuery)
			m.searchQuery = string(runes[:len(runes)-1])
			m.refreshSearch()
		}
	default:
		if message.Text != "" && !message.Mod.Contains(tea.ModCtrl) && !message.Mod.Contains(tea.ModAlt) {
			m.searchQuery += message.Text
			m.refreshSearch()
		}
	}
	return m, nil
}

func (m TaskList) searchView() string {
	if m.height > 0 && m.height <= fixedLines {
		result := "Type a query to search."
		if strings.TrimSpace(m.searchQuery) != "" {
			result = "No matches."
			if len(m.searchMatches) > 0 {
				bean := m.searchMatches[m.searchCursor]
				result = fmt.Sprintf("%d matches: %s %s", len(m.searchMatches), terminal.Text(bean.ID), terminal.Text(bean.Title))
			}
		}
		return styleDetail(boundDetail([]string{
			"Search tasks",
			"Query: " + terminal.Text(m.searchQuery) + "_",
			result,
			shortcut("up/down", "select") + "  " + shortcut("enter", "open") + "  " + shortcut("esc", "cancel"),
		}, m.width, m.height))
	}
	lines := []string{"Search tasks", "", "Query: " + terminal.Text(m.searchQuery) + "_", ""}
	if strings.TrimSpace(m.searchQuery) == "" {
		lines = append(lines, "Type a query to search IDs, titles, and bodies.")
	} else if len(m.searchMatches) == 0 {
		lines = append(lines, "No matches.")
	} else {
		lines = append(lines, fmt.Sprintf("%d matches", len(m.searchMatches)))
		available := len(m.searchMatches)
		if m.height > 0 {
			available = max(0, m.height-7)
		}
		start := max(0, m.searchCursor-available+1)
		for index := start; index < min(len(m.searchMatches), start+available); index++ {
			bean := m.searchMatches[index]
			marker := " "
			if index == m.searchCursor {
				marker = ">"
			}
			lines = append(lines, fmt.Sprintf("%s %s  %s", marker, terminal.Text(bean.ID), terminal.Text(bean.Title)))
		}
	}
	lines = append(lines, "", shortcut("up/down", "select")+"  "+shortcut("enter", "open")+"  "+shortcut("esc", "cancel"))
	return styleDetail(boundDetail(lines, m.width, m.height))
}

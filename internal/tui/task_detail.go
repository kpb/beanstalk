package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
)

// renderTaskDetail renders a selected task within the space allocated to a detail pane.
func renderTaskDetail(loaded []beans.Bean, selected beans.Bean, width, height int) string {
	lines := []string{"Task details", "", selected.Title}
	lines = append(lines,
		"ID: "+selected.ID,
		"Status: "+selected.Status,
		"Type: "+selected.Type,
		"Priority: "+selected.Priority,
		"Tags: "+displayTags(selected.Tags),
		"Created: "+selected.CreatedAt.UTC().Format(time.RFC3339),
		"Updated: "+selected.UpdatedAt.UTC().Format(time.RFC3339),
	)

	byID, children := taskHierarchy(loaded)
	lines = append(lines, hierarchyDetails(selected, byID, children)...)
	lines = append(lines, milestoneDetails(loaded, selected, byID)...)
	lines = append(lines, "", "Body:")
	if selected.Body == "" {
		lines = append(lines, "-")
	} else {
		lines = append(lines, strings.Split(selected.Body, "\n")...)
	}
	return boundDetail(lines, width, height)
}

func taskHierarchy(loaded []beans.Bean) (map[string]beans.Bean, map[string][]beans.Bean) {
	byID := make(map[string]beans.Bean, len(loaded))
	children := make(map[string][]beans.Bean)
	for _, bean := range loaded {
		byID[bean.ID] = bean
		if bean.Parent != "" {
			children[bean.Parent] = append(children[bean.Parent], bean)
		}
	}
	return byID, children
}

func hierarchyDetails(selected beans.Bean, byID map[string]beans.Bean, children map[string][]beans.Bean) []string {
	parent := "-"
	if selected.Parent != "" {
		parent = selected.Parent
		if bean, found := byID[selected.Parent]; found {
			parent += " " + bean.Title
		}
	}
	lines := []string{"Parent: " + parent}
	if len(children[selected.ID]) == 0 {
		return append(lines, "Children: -")
	}
	lines = append(lines, "Children:")
	for _, child := range children[selected.ID] {
		lines = append(lines, "  "+child.ID+" "+child.Title)
	}
	return lines
}

func milestoneDetails(loaded []beans.Bean, selected beans.Bean, byID map[string]beans.Bean) []string {
	milestone := nearestMilestone(selected, byID)
	if milestone.ID == "" {
		return nil
	}
	progresses, err := beans.MilestoneProgresses(loaded)
	if err != nil {
		return []string{"Milestone: " + milestone.ID + " " + milestone.Title}
	}
	for _, progress := range progresses {
		if progress.ID == milestone.ID {
			return []string{fmt.Sprintf("Milestone: %s %s (%d/%d, %d%%)", progress.ID, progress.Title, progress.Resolved, progress.Total, progress.Percent)}
		}
	}
	return nil
}

func nearestMilestone(selected beans.Bean, byID map[string]beans.Bean) beans.Bean {
	visited := make(map[string]bool)
	for bean := selected; bean.ID != "" && !visited[bean.ID]; bean = byID[bean.Parent] {
		visited[bean.ID] = true
		if bean.Type == "milestone" {
			return bean
		}
	}
	return beans.Bean{}
}

func displayTags(tags []string) string {
	if len(tags) == 0 {
		return "-"
	}
	return strings.Join(tags, ", ")
}

func boundDetail(lines []string, width, height int) string {
	var rendered []string
	for _, line := range lines {
		rendered = append(rendered, wrapDetailLine(line, width)...)
	}
	if height > 0 && len(rendered) > height {
		rendered = rendered[:height]
	}
	return strings.Join(rendered, "\n") + "\n"
}

func wrapDetailLine(line string, width int) []string {
	runes := []rune(line)
	if width <= 0 || len(runes) <= width {
		return []string{line}
	}

	lines := make([]string, 0, (len(runes)+width-1)/width)
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	return append(lines, string(runes))
}

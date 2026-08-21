package beans

import "fmt"

// MilestoneProgress summarizes the leaf descendants of a milestone.
type MilestoneProgress struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Status    string           `json:"status"`
	Total     int              `json:"total"`
	Resolved  int              `json:"resolved"`
	Completed int              `json:"completed"`
	Scrapped  int              `json:"scrapped"`
	Percent   int              `json:"percent"`
	Statuses  ProgressStatuses `json:"statuses"`
}

// ProgressStatuses counts every supported task status among a milestone's leaves.
type ProgressStatuses struct {
	Todo       int `json:"todo"`
	Draft      int `json:"draft"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Scrapped   int `json:"scrapped"`
}

// MilestoneProgresses calculates leaf-descendant progress for every milestone.
func MilestoneProgresses(loaded []Bean) ([]MilestoneProgress, error) {
	byID, children, err := hierarchy(loaded)
	if err != nil {
		return nil, err
	}
	if err := validateAcyclic(byID, children); err != nil {
		return nil, err
	}

	progresses := make([]MilestoneProgress, 0)
	for _, bean := range loaded {
		if bean.Type != "milestone" {
			continue
		}
		progress, err := milestoneProgress(bean, byID, children)
		if err != nil {
			return nil, err
		}
		progresses = append(progresses, progress)
	}
	return progresses, nil
}

func hierarchy(loaded []Bean) (map[string]Bean, map[string][]string, error) {
	byID := make(map[string]Bean, len(loaded))
	for _, bean := range loaded {
		if _, found := byID[bean.ID]; found {
			return nil, nil, fmt.Errorf("%w: %s", ErrDuplicateBeanID, bean.ID)
		}
		byID[bean.ID] = bean
	}
	children := make(map[string][]string, len(loaded))
	for _, bean := range loaded {
		if bean.Parent == "" {
			continue
		}
		if _, found := byID[bean.Parent]; !found {
			return nil, nil, fmt.Errorf("parent bean not found: %s", bean.Parent)
		}
		children[bean.Parent] = append(children[bean.Parent], bean.ID)
	}
	return byID, children, nil
}

func validateAcyclic(byID map[string]Bean, children map[string][]string) error {
	state := make(map[string]int, len(byID))
	var visit func(string) error
	visit = func(id string) error {
		state[id] = 1
		for _, child := range children[id] {
			if state[child] == 1 {
				return fmt.Errorf("%w: %s", ErrParentCycle, child)
			}
			if state[child] == 0 {
				if err := visit(child); err != nil {
					return err
				}
			}
		}
		state[id] = 2
		return nil
	}
	for id := range byID {
		if state[id] == 0 {
			if err := visit(id); err != nil {
				return err
			}
		}
	}
	return nil
}

func milestoneProgress(milestone Bean, byID map[string]Bean, children map[string][]string) (MilestoneProgress, error) {
	progress := MilestoneProgress{ID: milestone.ID, Title: milestone.Title, Status: milestone.Status}
	for _, child := range children[milestone.ID] {
		if err := addLeaves(&progress, child, byID, children); err != nil {
			return MilestoneProgress{}, err
		}
	}
	if progress.Total > 0 {
		progress.Percent = progress.Resolved * 100 / progress.Total
	}
	return progress, nil
}

func addLeaves(progress *MilestoneProgress, id string, byID map[string]Bean, children map[string][]string) error {
	if descendants := children[id]; len(descendants) > 0 {
		for _, child := range descendants {
			if err := addLeaves(progress, child, byID, children); err != nil {
				return err
			}
		}
		return nil
	}

	progress.Total++
	switch byID[id].Status {
	case "todo":
		progress.Statuses.Todo++
	case "draft":
		progress.Statuses.Draft++
	case "in-progress":
		progress.Statuses.InProgress++
	case "completed":
		progress.Completed++
		progress.Resolved++
		progress.Statuses.Completed++
	case "scrapped":
		progress.Scrapped++
		progress.Resolved++
		progress.Statuses.Scrapped++
	default:
		return fmt.Errorf("invalid leaf status %q", byID[id].Status)
	}
	return nil
}

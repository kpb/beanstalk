package beans

import (
	"slices"
	"strings"
)

// Sort orders beans consistently for terminal presentation.
func Sort(loaded []Bean) {
	statusOrder := map[string]int{"in-progress": 0, "todo": 1, "draft": 2, "completed": 3, "scrapped": 4}
	priorityOrder := map[string]int{"critical": 0, "high": 1, "normal": 2, "low": 3, "deferred": 4}
	typeOrder := map[string]int{"milestone": 0, "epic": 1, "bug": 2, "feature": 3, "task": 4}
	slices.SortFunc(loaded, func(left, right Bean) int {
		for _, comparison := range []int{
			orderValue(statusOrder, left.Status) - orderValue(statusOrder, right.Status),
			orderValue(priorityOrder, left.Priority) - orderValue(priorityOrder, right.Priority),
			orderValue(typeOrder, left.Type) - orderValue(typeOrder, right.Type),
		} {
			if comparison != 0 {
				return comparison
			}
		}
		if comparison := strings.Compare(strings.ToLower(left.Title), strings.ToLower(right.Title)); comparison != 0 {
			return comparison
		}
		return strings.Compare(left.ID, right.ID)
	})
}

func orderValue(order map[string]int, value string) int {
	if index, found := order[value]; found {
		return index
	}
	return len(order)
}

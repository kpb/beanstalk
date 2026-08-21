package beans

import "testing"

func TestSortUsesIDToBreakTitleTies(t *testing.T) {
	loaded := []Bean{
		{ID: "project-b2", Title: "Same", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "project-a1", Title: "Same", Status: "todo", Type: "task", Priority: "normal"},
		{ID: "project-c3", Title: "Unknown", Status: "unknown", Type: "unknown", Priority: "unknown"},
	}
	Sort(loaded)
	if got, want := []string{loaded[0].ID, loaded[1].ID, loaded[2].ID}, []string{"project-a1", "project-b2", "project-c3"}; !equalStrings(got, want) {
		t.Errorf("sorted IDs = %v, want %v", got, want)
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

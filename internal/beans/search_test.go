package beans

import "testing"

func TestSearchMatchesIDTitleAndBodyWithoutCase(t *testing.T) {
	loaded := []Bean{
		{ID: "Project-ALPHA", Title: "First"},
		{ID: "project-b", Title: "Alpha feature"},
		{ID: "project-c", Title: "Third", Body: "## ALPHA\nDetails"},
		{ID: "project-d", Title: "Other"},
	}
	matches := Search(loaded, "  alpha  ")
	if len(matches) != 3 || matches[0].ID != "Project-ALPHA" || matches[1].ID != "project-b" || matches[2].ID != "project-c" {
		t.Errorf("matches = %#v", matches)
	}
	if matches := Search(loaded, " \t "); len(matches) != 0 {
		t.Errorf("empty query matches = %#v", matches)
	}
	if matches := Search(loaded, "absent"); len(matches) != 0 {
		t.Errorf("missing query matches = %#v", matches)
	}
}

package beans

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestUpdateRoundTripsImportedBeans(t *testing.T) {
	for _, test := range []struct {
		name    string
		fixture string
		crlf    bool
	}{
		{name: "full", fixture: "roundtrip-full.md"},
		{name: "legacy", fixture: "roundtrip-legacy.md"},
		{name: "CRLF", fixture: "roundtrip-full.md", crlf: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			workingDirectory, path, originalMetadata, originalBody := roundTripProject(t, test.fixture, test.crlf)
			originalParent, _ := originalMetadata["parent"].(string)
			status := "in-progress"
			updatedAt := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
			if _, err := Update(workingDirectory, "project-a1", UpdateFields{Status: &status}, updatedAt); err != nil {
				t.Fatalf("updating status: %v", err)
			}
			assertRoundTrip(t, path, originalMetadata, originalBody, status, originalParent)

			parent := "project-parent"
			if _, err := Update(workingDirectory, "project-a1", UpdateFields{Parent: &parent}, updatedAt); err != nil {
				t.Fatalf("setting parent: %v", err)
			}
			assertRoundTrip(t, path, originalMetadata, originalBody, status, parent)
			parent = ""
			if _, err := Update(workingDirectory, "project-a1", UpdateFields{Parent: &parent}, updatedAt); err != nil {
				t.Fatalf("removing parent: %v", err)
			}
			assertRoundTrip(t, path, originalMetadata, originalBody, status, "")

			body := "Replacement body.\n\nSecond paragraph."
			if _, err := Update(workingDirectory, "project-a1", UpdateFields{Body: &body}, updatedAt); err != nil {
				t.Fatalf("replacing body: %v", err)
			}
			assertRoundTrip(t, path, originalMetadata, body, status, "")
			loaded, err := Find(workingDirectory, "project-a1")
			if err != nil || loaded.Body != body || loaded.Status != status || loaded.Parent != "" {
				t.Errorf("reloaded bean = %#v, error = %v", loaded, err)
			}
			if test.fixture == "roundtrip-legacy.md" && (loaded.Type != "task" || loaded.Priority != "normal") {
				t.Errorf("legacy defaults after update = %q/%q, want task/normal", loaded.Type, loaded.Priority)
			}
		})
	}
}

func TestClaimRoundTripsImportedBean(t *testing.T) {
	workingDirectory, path, originalMetadata, originalBody := roundTripProject(t, "roundtrip-full.md", false)
	if _, err := Claim(workingDirectory, "project-a1", time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("claiming bean: %v", err)
	}
	assertRoundTrip(t, path, originalMetadata, originalBody, "in-progress", "project-parent")
}

func roundTripProject(t *testing.T, fixture string, crlf bool) (string, string, map[string]any, string) {
	t.Helper()
	workingDirectory := t.TempDir()
	directory := filepath.Join(workingDirectory, ".beans")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	parent, err := Render(Bean{ID: "project-parent", Title: "Parent", Status: "todo", Type: "task"})
	if err != nil {
		t.Fatalf("rendering parent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "project-parent--parent.md"), parent, 0o644); err != nil {
		t.Fatalf("writing parent: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join("testdata", fixture))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	if crlf {
		contents = []byte(strings.ReplaceAll(string(contents), "\n", "\r\n"))
	}
	path := filepath.Join(directory, "project-a1--imported.md")
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	metadata, body, err := splitFrontMatter(string(contents))
	if err != nil {
		t.Fatalf("splitting fixture: %v", err)
	}
	var original map[string]any
	if err := yaml.Unmarshal([]byte(metadata), &original); err != nil {
		t.Fatalf("decoding fixture: %v", err)
	}
	return workingDirectory, path, original, body
}

func assertRoundTrip(t *testing.T, path string, originalMetadata map[string]any, wantBody, wantStatus, wantParent string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading updated bean: %v", err)
	}
	if !strings.HasPrefix(string(contents), "---\n") {
		t.Errorf("updated front matter is not LF-delimited: %q", contents)
	}
	metadata, body, err := splitFrontMatter(string(contents))
	if err != nil {
		t.Fatalf("splitting updated bean: %v", err)
	}
	if body != wantBody {
		t.Errorf("body = %q, want %q", body, wantBody)
	}
	var updated map[string]any
	if err := yaml.Unmarshal([]byte(metadata), &updated); err != nil {
		t.Fatalf("decoding updated metadata: %v", err)
	}
	if updated["status"] != wantStatus || updated["updated_at"] != "2026-10-01T12:00:00Z" {
		t.Errorf("status/updated_at = %v/%v", updated["status"], updated["updated_at"])
	}
	if wantParent == "" {
		if _, found := updated["parent"]; found {
			t.Errorf("removed parent still present: %v", updated["parent"])
		}
	} else if updated["parent"] != wantParent {
		t.Errorf("parent = %v, want %q", updated["parent"], wantParent)
	}
	unchanged := make(map[string]any)
	for key, value := range originalMetadata {
		if key != "status" && key != "parent" && key != "updated_at" {
			unchanged[key] = value
		}
	}
	for key, want := range unchanged {
		if !reflect.DeepEqual(updated[key], want) {
			t.Errorf("metadata %q = %#v, want %#v", key, updated[key], want)
		}
	}
}

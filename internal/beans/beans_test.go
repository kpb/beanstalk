package beans

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadParsesNestedBeansAndDefaultsLegacyFields(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	path := filepath.Join(workingDirectory, ".beans", "archive", "project-a1--legacy.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating archive directory: %v", err)
	}
	contents := "---\ntitle: Legacy bean\nstatus: todo\ncreated_at: 2026-08-15T12:00:00Z\nupdated_at: 2026-08-15T13:00:00Z\n---\nBody\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}

	loaded, err := Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading beans: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded beans = %d, want 1", len(loaded))
	}
	bean := loaded[0]
	if bean.ID != "project-a1" || bean.Slug != "legacy" || bean.Path != "archive/project-a1--legacy.md" || bean.Type != "task" || bean.Priority != "normal" || len(bean.Tags) != 0 || bean.Body != "Body" || !bean.CreatedAt.Equal(time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC)) || !bean.UpdatedAt.Equal(time.Date(2026, time.August, 15, 13, 0, 0, 0, time.UTC)) {
		t.Errorf("loaded bean = %#v", bean)
	}
}

func TestLoadReportsMalformedFrontMatter(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans", "project-a1--broken.md"), []byte("title: missing delimiters\n"), 0o644); err != nil {
		t.Fatalf("writing malformed bean: %v", err)
	}

	_, err := Load(workingDirectory)
	if err == nil || !strings.Contains(err.Error(), "project-a1--broken.md") || !strings.Contains(err.Error(), "missing opening front matter delimiter") {
		t.Errorf("load error = %v", err)
	}
}

func TestLoadSupportsCustomPathAndClosingDelimiterAtEOF(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, "tasks"), 0o755); err != nil {
		t.Fatalf("creating tasks directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: tasks\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	contents := "---\ntitle: End of file\nstatus: todo\n---"
	if err := os.WriteFile(filepath.Join(workingDirectory, "tasks", "project-a1--eof.md"), []byte(contents), 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}

	loaded, err := Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading beans: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Title != "End of file" || loaded[0].Path != "project-a1--eof.md" {
		t.Errorf("loaded beans = %#v", loaded)
	}
}

func TestLoadRequiresInitializedProject(t *testing.T) {
	_, err := Load(t.TempDir())
	if !errors.Is(err, ErrNotInitialized) {
		t.Errorf("load error = %v, want ErrNotInitialized", err)
	}
}

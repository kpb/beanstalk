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

func TestLoadAndRenderPreserveParent(t *testing.T) {
	workingDirectory := t.TempDir()
	if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	path := filepath.Join(workingDirectory, ".beans", "project-child--child.md")
	contents, err := Render(Bean{ID: "project-child", Title: "Child", Status: "todo", Type: "task", Parent: "project-parent"})
	if err != nil {
		t.Fatalf("rendering bean: %v", err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}
	parentContents, err := Render(Bean{ID: "project-parent", Title: "Parent", Status: "todo", Type: "task"})
	if err != nil {
		t.Fatalf("rendering parent bean: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans", "project-parent--parent.md"), parentContents, 0o644); err != nil {
		t.Fatalf("writing parent bean: %v", err)
	}

	loaded, err := Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading beans: %v", err)
	}
	if len(loaded) != 2 || loaded[0].Parent != "project-parent" {
		t.Errorf("loaded beans = %#v", loaded)
	}
}

func TestLoadValidatesSupportedMetadataAndHierarchy(t *testing.T) {
	tests := []struct {
		name  string
		beans []Bean
		err   error
		text  string
	}{
		{
			name:  "unsupported status",
			beans: []Bean{{ID: "project-a1", Title: "Task", Status: "unknown", Type: "task"}},
			err:   ErrInvalidBeanStatus,
			text:  `invalid bean status "unknown": project-a1`,
		},
		{
			name:  "unsupported type",
			beans: []Bean{{ID: "project-a1", Title: "Task", Status: "todo", Type: "unknown"}},
			err:   ErrInvalidBeanType,
			text:  `invalid bean type "unknown": project-a1`,
		},
		{
			name:  "unsupported priority",
			beans: []Bean{{ID: "project-a1", Title: "Task", Status: "todo", Type: "task", Priority: "unknown"}},
			err:   ErrInvalidBeanPriority,
			text:  `invalid bean priority "unknown": project-a1`,
		},
		{
			name:  "missing parent",
			beans: []Bean{{ID: "project-a1", Title: "Task", Status: "todo", Type: "task", Parent: "missing"}},
			text:  "parent bean not found: missing",
		},
		{
			name: "parent cycle",
			beans: []Bean{
				{ID: "project-a1", Title: "One", Status: "todo", Type: "task", Parent: "project-b2"},
				{ID: "project-b2", Title: "Two", Status: "todo", Type: "task", Parent: "project-a1"},
			},
			err:  ErrParentCycle,
			text: "parent link would create a cycle",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workingDirectory := t.TempDir()
			if err := os.Mkdir(filepath.Join(workingDirectory, ".beans"), 0o755); err != nil {
				t.Fatalf("creating beans directory: %v", err)
			}
			if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
				t.Fatalf("writing config: %v", err)
			}
			for _, bean := range test.beans {
				contents, err := Render(bean)
				if err != nil {
					t.Fatalf("rendering bean: %v", err)
				}
				if err := os.WriteFile(filepath.Join(workingDirectory, ".beans", bean.ID+"--task.md"), contents, 0o644); err != nil {
					t.Fatalf("writing bean: %v", err)
				}
			}

			_, err := Load(workingDirectory)
			if err == nil || (test.err != nil && !errors.Is(err, test.err)) || !strings.Contains(err.Error(), test.text) {
				t.Errorf("load error = %v, want %q", err, test.text)
			}
		})
	}
}

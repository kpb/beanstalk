package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestArchiveCommandMovesResolvedBeans(t *testing.T) {
	workingDirectory := initializedProject(t)
	completed := beans.Bean{ID: "project-completed", Slug: "completed", Title: "Completed", Status: "completed", Type: "task", Body: "Keep this body."}
	scrapped := beans.Bean{ID: "project-scrapped", Slug: "scrapped", Title: "Scrapped", Status: "scrapped", Type: "bug"}
	writeBean(t, workingDirectory, ".beans/project-completed--completed.md", completed)
	writeBean(t, workingDirectory, ".beans/project-scrapped--scrapped.md", scrapped)
	writeBean(t, workingDirectory, ".beans/project-todo--todo.md", beans.Bean{ID: "project-todo", Title: "Todo", Status: "todo", Type: "task"})
	source := filepath.Join(workingDirectory, ".beans", "project-completed--completed.md")
	contents, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("reading source: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(source, 0o640); err != nil {
			t.Fatalf("setting source permissions: %v", err)
		}
	}
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"archive", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing archive command: %v", err)
	}
	var response struct {
		Success bool         `json:"success"`
		Beans   []beans.Bean `json:"beans"`
		Message string       `json:"message"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decoding archive response: %v\n%s", err, output.String())
	}
	if !response.Success || len(response.Beans) != 2 || response.Beans[0].Path != "archive/project-completed--completed.md" || response.Beans[1].Path != "archive/project-scrapped--scrapped.md" {
		t.Errorf("archive response = %#v", response)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Errorf("source still exists, stat error = %v", err)
	}
	destination := filepath.Join(workingDirectory, ".beans", "archive", "project-completed--completed.md")
	archivedContents, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("reading archived bean: %v", err)
	}
	if string(archivedContents) != string(contents) {
		t.Errorf("archived contents = %q, want %q", archivedContents, contents)
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatalf("stating archived bean: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o640 {
		t.Errorf("archived permissions = %o, want 640", info.Mode().Perm())
	}
	if bean, err := beans.Find(workingDirectory, "project-completed"); err != nil || bean.Path != "archive/project-completed--completed.md" {
		t.Errorf("finding archived bean = %#v, %v", bean, err)
	}
	if _, err := os.Stat(filepath.Join(workingDirectory, ".beans", "project-todo--todo.md")); err != nil {
		t.Errorf("todo bean was moved: %v", err)
	}
}

func TestArchiveCommandReportsNoResolvedBeans(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-todo--todo.md", beans.Bean{ID: "project-todo", Title: "Todo", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"archive"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing archive command: %v", err)
	}
	if got, want := output.String(), "No resolved beans to archive.\n"; got != want {
		t.Errorf("archive output = %q, want %q", got, want)
	}
}

func TestArchiveCommandIsRegistered(t *testing.T) {
	command := NewRootCommand()
	found, _, err := command.Find([]string{"archive"})
	if err != nil || found == nil || found.Name() != "archive" {
		t.Errorf("archive command = %#v, error = %v", found, err)
	}
}

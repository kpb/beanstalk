package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestSearchCommandTextAndJSON(t *testing.T) {
	directory := initializedProject(t)
	writeBean(t, directory, ".beans/project-a--one.md", beans.Bean{ID: "project-a", Title: "Match title", Status: "todo", Type: "task", Body: "body"})
	writeBean(t, directory, ".beans/project-b--two.md", beans.Bean{ID: "project-b", Title: "Other", Status: "in-progress", Type: "task", Body: "MATCH in body"})
	writeBean(t, directory, ".beans/archive/project-c--old.md", beans.Bean{ID: "project-c", Title: "Archived match", Status: "completed", Type: "task"})
	t.Chdir(directory)

	run := func(args ...string) string {
		t.Helper()
		command := NewRootCommand()
		output := new(bytes.Buffer)
		command.SetOut(output)
		command.SetArgs(append([]string{"search"}, args...))
		if err := command.Execute(); err != nil {
			t.Fatalf("search %v: %v", args, err)
		}
		return output.String()
	}
	text := run("mAtCh")
	if strings.Index(text, "project-b") < 0 || strings.Index(text, "project-a") < strings.Index(text, "project-b") || strings.Contains(text, "project-c") {
		t.Errorf("text search = %q", text)
	}
	var results []beans.Bean
	if err := json.Unmarshal([]byte(run("match", "--all", "--json")), &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 || results[2].ID != "project-c" || results[0].Body != "" || results[1].Body != "" {
		t.Errorf("JSON results = %#v", results)
	}
	if got := run(" ", "--json"); got != "[]\n" {
		t.Errorf("empty query JSON = %q", got)
	}
	if got := run("missing"); got != "No beans found.\n" {
		t.Errorf("no matches text = %q", got)
	}
}

func TestSearchCommandRequiresOneQuery(t *testing.T) {
	for _, args := range [][]string{{"search"}, {"search", "one", "two"}} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Errorf("args %v accepted", args)
		}
	}
}

package commands

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kpb/beanstalk/internal/beans"
)

func TestClaimCommand(t *testing.T) {
	workingDirectory := initializedProject(t)
	writeBean(t, workingDirectory, ".beans/project-a1--task.md", beans.Bean{ID: "project-a1", Title: "Task", Status: "todo", Type: "task"})
	t.Chdir(workingDirectory)

	command := NewRootCommand()
	output := new(bytes.Buffer)
	command.SetOut(output)
	command.SetArgs([]string{"claim", "project-a1", "--json"})
	if err := command.Execute(); err != nil {
		t.Fatalf("executing claim command: %v", err)
	}
	var response struct {
		Success bool       `json:"success"`
		Bean    beans.Bean `json:"bean"`
		Message string     `json:"message"`
	}
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatalf("decoding JSON output: %v\n%s", err, output.String())
	}
	if !response.Success || response.Message != "Bean claimed" || response.Bean.Status != "in-progress" {
		t.Errorf("JSON response = %#v", response)
	}

	command = NewRootCommand()
	command.SetArgs([]string{"claim", "project-a1"})
	if err := command.Execute(); err == nil || !strings.Contains(err.Error(), "not available to claim") {
		t.Errorf("second claim error = %v", err)
	}
}

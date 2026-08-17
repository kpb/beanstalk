package beans

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestClaimAllowsOnlyOneConcurrentClaimant(t *testing.T) {
	workingDirectory := claimTestProject(t)

	start := make(chan struct{})
	errorsByClaimant := make(chan error, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Go(func() {
			<-start
			_, err := Claim(workingDirectory, "project-a1", time.Now())
			errorsByClaimant <- err
		})
	}
	close(start)
	group.Wait()
	close(errorsByClaimant)

	successes := 0
	conflicts := 0
	for err := range errorsByClaimant {
		if err == nil {
			successes++
			continue
		}
		if errors.Is(err, ErrBeanNotClaimable) {
			conflicts++
			continue
		}
		t.Errorf("claim error = %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Errorf("claims: %d successes, %d conflicts; want 1 each", successes, conflicts)
	}

	loaded, err := Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading claimed bean: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Status != "in-progress" {
		t.Errorf("loaded beans = %#v", loaded)
	}
}

func TestClaimDoesNotOverwriteConcurrentUpdate(t *testing.T) {
	workingDirectory := claimTestProject(t)
	start := make(chan struct{})
	var group sync.WaitGroup
	claimErrors := make(chan error, 1)
	updateErrors := make(chan error, 1)
	group.Go(func() {
		<-start
		_, err := Claim(workingDirectory, "project-a1", time.Now())
		claimErrors <- err
	})
	group.Go(func() {
		<-start
		_, err := UpdateStatus(workingDirectory, "project-a1", "completed", time.Now())
		updateErrors <- err
	})
	close(start)
	group.Wait()

	if err := <-updateErrors; err != nil {
		t.Fatalf("updating bean: %v", err)
	}
	if err := <-claimErrors; err != nil && !errors.Is(err, ErrBeanNotClaimable) {
		t.Errorf("claim error = %v", err)
	}
	loaded, err := Load(workingDirectory)
	if err != nil {
		t.Fatalf("loading updated bean: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Status != "completed" {
		t.Errorf("loaded beans = %#v", loaded)
	}
}

func claimTestProject(t *testing.T) string {
	t.Helper()
	workingDirectory := t.TempDir()
	beansDirectory := filepath.Join(workingDirectory, ".beans")
	if err := os.Mkdir(beansDirectory, 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workingDirectory, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}
	contents, err := Render(Bean{
		ID:        "project-a1",
		Title:     "Task",
		Status:    "todo",
		Type:      "task",
		CreatedAt: time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, time.August, 15, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("rendering bean: %v", err)
	}
	if err := os.WriteFile(filepath.Join(beansDirectory, "project-a1--task.md"), contents, 0o644); err != nil {
		t.Fatalf("writing bean: %v", err)
	}
	return workingDirectory
}

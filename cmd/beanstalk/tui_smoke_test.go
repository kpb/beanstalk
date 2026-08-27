//go:build darwin || linux

package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/creack/pty"
)

func TestTUIExecutableSmoke(t *testing.T) {
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".beans"), 0o755); err != nil {
		t.Fatalf("creating beans directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project, ".beans.yml"), []byte("beans:\n  path: .beans\n"), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	binary := filepath.Join(t.TempDir(), "beanstalk")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building beanstalk: %v\n%s", err, output)
	}

	command := exec.Command(binary, "tui")
	command.Dir = project
	command.Env = append(os.Environ(), "TERM=xterm-256color")
	terminal, err := pty.Start(command)
	if err != nil {
		t.Fatalf("starting TUI: %v", err)
	}
	defer terminal.Close()
	if err := pty.Setsize(terminal, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		t.Fatalf("setting terminal size: %v", err)
	}

	exited := make(chan error, 1)
	go func() { exited <- command.Wait() }()
	finished := false
	defer func() {
		if !finished {
			_ = command.Process.Kill()
			<-exited
		}
	}()

	terminal.SetReadDeadline(time.Now().Add(5 * time.Second))
	var output bytes.Buffer
	buffer := make([]byte, 1024)
	for !bytes.Contains(output.Bytes(), []byte("Beanstalk tasks (0)")) {
		count, err := terminal.Read(buffer)
		if count > 0 {
			output.Write(buffer[:count])
		}
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				t.Fatalf("TUI did not render: %s", output.String())
			}
			if errors.Is(err, io.EOF) {
				t.Fatalf("TUI exited before rendering: %s", output.String())
			}
			t.Fatalf("reading TUI output: %v", err)
		}
	}
	if _, err := terminal.Write([]byte("q")); err != nil {
		t.Fatalf("quitting TUI: %v", err)
	}

	select {
	case err := <-exited:
		finished = true
		if err != nil {
			t.Fatalf("TUI exit error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("TUI did not exit after q")
	}
}

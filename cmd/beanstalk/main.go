// Beanstalk is a terminal-native tracker for Beans-format tasks.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kpb/beanstalk/internal/commands"
	"github.com/kpb/beanstalk/internal/terminal"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	command := commands.NewRootCommand()
	command.SetArgs(args)
	command.SetOut(stdout)
	command.SetErr(stderr)

	if err := command.Execute(); err != nil {
		fmt.Fprintln(stderr, "error:", terminal.Text(err.Error()))
		return 1
	}

	return 0
}

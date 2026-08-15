// Beanstalk is a terminal-native tracker for Beans-format tasks.
package main

import (
	"os"

	"github.com/kpb/beanstalk/internal/commands"
)

func main() {
	if err := commands.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

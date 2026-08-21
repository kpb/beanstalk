// Package commands contains the Beanstalk command-line interface.
package commands

import "github.com/spf13/cobra"

var version = "0.0.0-dev"

// NewRootCommand constructs the top-level Beanstalk command.
func NewRootCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "beanstalk",
		Short:         "A terminal-native tracker for Beans-format tasks",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	command.AddCommand(newInitCommand())
	command.AddCommand(newCreateCommand())
	command.AddCommand(newListCommand())
	command.AddCommand(newTUICommand())
	command.AddCommand(newShowCommand())
	command.AddCommand(newUpdateCommand())
	command.AddCommand(newClaimCommand())
	command.AddCommand(newPrimeCommand())
	command.AddCommand(newVersionCommand())
	return command
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the Beanstalk version",
		Run: func(command *cobra.Command, args []string) {
			command.Println(version)
		},
	}
}

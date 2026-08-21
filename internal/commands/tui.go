package commands

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/kpb/beanstalk/internal/beans"
	"github.com/kpb/beanstalk/internal/tui"
	"github.com/spf13/cobra"
)

func newTUICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Browse Beans-format tasks interactively",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			loaded, err := loadTUITasks(workingDirectory)
			if err != nil {
				return err
			}
			_, err = tea.NewProgram(tui.NewTaskList(loaded), tea.WithInput(command.InOrStdin()), tea.WithOutput(command.OutOrStdout())).Run()
			return err
		},
	}
}

func loadTUITasks(workingDirectory string) ([]beans.Bean, error) {
	loaded, err := beans.Load(workingDirectory)
	if err != nil {
		return nil, err
	}
	beans.Sort(loaded)
	return loaded, nil
}

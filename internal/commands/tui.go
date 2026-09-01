package commands

import (
	"fmt"
	"os"
	"time"

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
		RunE:  runTUI,
	}
}

func runTUI(command *cobra.Command, args []string) error {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}
	loaded, err := loadTUITasks(workingDirectory, false)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(tui.NewTaskList(loaded,
		tui.WithTaskLoader(func(showArchived bool) ([]beans.Bean, error) {
			return loadTUITasks(workingDirectory, showArchived)
		}),
		tui.WithTaskClaimer(func(id string) error {
			_, err := beans.Claim(workingDirectory, id, time.Now())
			return err
		}),
		tui.WithStatusUpdater(func(id, status string) error {
			_, err := beans.UpdateStatus(workingDirectory, id, status, time.Now())
			return err
		}),
	), tea.WithInput(command.InOrStdin()), tea.WithOutput(command.OutOrStdout())).Run()
	return err
}

func loadTUITasks(workingDirectory string, showArchived bool) ([]beans.Bean, error) {
	loaded, err := beans.Load(workingDirectory)
	if err != nil {
		return nil, err
	}
	if !showArchived {
		loaded = beans.Active(loaded)
	}
	beans.Sort(loaded)
	return loaded, nil
}

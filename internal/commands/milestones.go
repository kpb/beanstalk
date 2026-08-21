package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type milestonesOptions struct {
	all  bool
	json bool
}

func newMilestonesCommand() *cobra.Command {
	options := milestonesOptions{}
	command := &cobra.Command{
		Use:   "milestones",
		Short: "Show milestone progress",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			progresses, err := loadMilestoneProgresses(workingDirectory, options.all)
			if err != nil {
				return err
			}
			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(progresses)
			}
			if len(progresses) == 0 {
				command.Println("No milestones found.")
				return nil
			}
			for _, progress := range progresses {
				command.Printf("%s  %s  %d/%d (%d%%)  %s\n", progress.ID, progress.Status, progress.Resolved, progress.Total, progress.Percent, progress.Title)
			}
			return nil
		},
	}
	command.Flags().BoolVar(&options.all, "all", false, "Include completed and scrapped milestones")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

func loadMilestoneProgresses(workingDirectory string, all bool) ([]beans.MilestoneProgress, error) {
	loaded, err := beans.Load(workingDirectory)
	if err != nil {
		return nil, err
	}
	beans.Sort(loaded)
	progresses, err := beans.MilestoneProgresses(loaded)
	if err != nil {
		return nil, err
	}
	if all {
		return progresses, nil
	}
	active := make([]beans.MilestoneProgress, 0, len(progresses))
	for _, progress := range progresses {
		if progress.Status == "todo" || progress.Status == "draft" || progress.Status == "in-progress" {
			active = append(active, progress)
		}
	}
	return active, nil
}

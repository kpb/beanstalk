package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	status string
	json   bool
}

func newUpdateCommand() *cobra.Command {
	options := updateOptions{}
	command := &cobra.Command{
		Use:     "update <id>",
		Aliases: []string{"u"},
		Short:   "Update a Beans-format task",
		Args:    cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if options.status == "" {
				return fmt.Errorf("--status is required")
			}
			if !beanStatuses[options.status] {
				return fmt.Errorf("invalid status %q", options.status)
			}
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			bean, err := beans.UpdateStatus(workingDirectory, args[0], options.status, time.Now())
			if err != nil {
				return err
			}
			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(struct {
					Success bool       `json:"success"`
					Bean    beans.Bean `json:"bean"`
					Message string     `json:"message"`
				}{true, bean, "Bean updated"})
			}
			command.Printf("Updated %s status to %s\n", bean.ID, bean.Status)
			return nil
		},
	}
	command.Flags().StringVarP(&options.status, "status", "s", "", "New status")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

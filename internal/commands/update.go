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
	parent string
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
			statusChanged := command.Flags().Changed("status")
			parentChanged := command.Flags().Changed("parent")
			if !statusChanged && !parentChanged {
				return fmt.Errorf("at least one of --status or --parent is required")
			}
			if statusChanged && (options.status == "" || !beanStatuses[options.status]) {
				return fmt.Errorf("invalid status %q", options.status)
			}
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			fields := beans.UpdateFields{}
			if statusChanged {
				fields.Status = &options.status
			}
			if parentChanged {
				fields.Parent = &options.parent
			}
			bean, err := beans.Update(workingDirectory, args[0], fields, time.Now())
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
			if statusChanged && !parentChanged {
				command.Printf("Updated %s status to %s\n", bean.ID, bean.Status)
			} else {
				command.Printf("Updated %s\n", bean.ID)
			}
			return nil
		},
	}
	command.Flags().StringVarP(&options.status, "status", "s", "", "New status")
	command.Flags().StringVar(&options.parent, "parent", "", "Parent bean ID; pass an empty value to remove")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

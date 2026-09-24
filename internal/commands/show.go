package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/kpb/beanstalk/internal/terminal"
	"github.com/spf13/cobra"
)

type showOptions struct {
	json bool
}

func newShowCommand() *cobra.Command {
	options := showOptions{}
	command := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a Beans-format task",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			bean, err := beans.Find(workingDirectory, args[0])
			if err != nil {
				return err
			}
			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(bean)
			}

			parent := bean.Parent
			if parent == "" {
				parent = "-"
			}
			command.Printf("ID: %s\nStatus: %s\nType: %s\nPriority: %s\nTags: %s\nParent: %s\nCreated: %s\nUpdated: %s\n\n%s\n", terminal.Text(bean.ID), terminal.Text(bean.Status), terminal.Text(bean.Type), terminal.Text(bean.Priority), terminal.Text(strings.Join(bean.Tags, ", ")), terminal.Text(parent), bean.CreatedAt.UTC().Format(time.RFC3339), bean.UpdatedAt.UTC().Format(time.RFC3339), terminal.Text(bean.Title))
			if bean.Body != "" {
				command.Printf("\n%s\n", terminal.Lines(bean.Body))
			}
			return nil
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

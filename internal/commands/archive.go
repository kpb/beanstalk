package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type archiveOptions struct {
	json bool
}

func newArchiveCommand() *cobra.Command {
	options := archiveOptions{}
	command := &cobra.Command{
		Use:   "archive",
		Short: "Archive completed and scrapped tasks",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			archived, err := beans.Archive(workingDirectory)
			if err != nil {
				return err
			}
			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(struct {
					Success bool         `json:"success"`
					Beans   []beans.Bean `json:"beans"`
					Message string       `json:"message"`
				}{true, archived, fmt.Sprintf("Archived %d beans", len(archived))})
			}
			if len(archived) == 0 {
				command.Println("No resolved beans to archive.")
				return nil
			}
			command.Printf("Archived %d beans.\n", len(archived))
			return nil
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

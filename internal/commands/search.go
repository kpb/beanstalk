package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/kpb/beanstalk/internal/terminal"
	"github.com/spf13/cobra"
)

func newSearchCommand() *cobra.Command {
	var all, asJSON bool
	command := &cobra.Command{
		Use:   "search <text>",
		Short: "Search task IDs, titles, and bodies",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			loaded, err := beans.Load(workingDirectory)
			if err != nil {
				return err
			}
			if !all {
				loaded = beans.Active(loaded)
			}
			beans.Sort(loaded)
			matches := beans.Search(loaded, args[0])
			if asJSON {
				for index := range matches {
					matches[index].Body = ""
				}
				return json.NewEncoder(command.OutOrStdout()).Encode(matches)
			}
			if len(matches) == 0 {
				command.Println("No beans found.")
				return nil
			}
			command.Printf("%-18s  S  T  TITLE\n", "ID")
			for _, bean := range matches {
				command.Printf("%-18s  %s  %s  %s\n", terminal.Text(bean.ID), statusMarker(bean.Status), typeMarker(bean.Type), terminal.Text(bean.Title))
			}
			return nil
		},
	}
	command.Flags().BoolVar(&all, "all", false, "Include archived tasks")
	command.Flags().BoolVar(&asJSON, "json", false, "Output JSON")
	return command
}

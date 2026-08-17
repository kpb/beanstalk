package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type claimOptions struct {
	json bool
}

func newClaimCommand() *cobra.Command {
	options := claimOptions{}
	command := &cobra.Command{
		Use:   "claim <id>",
		Short: "Claim a todo task for work",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			bean, err := beans.Claim(workingDirectory, args[0], time.Now())
			if err != nil {
				return err
			}
			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(struct {
					Success bool       `json:"success"`
					Bean    beans.Bean `json:"bean"`
					Message string     `json:"message"`
				}{true, bean, "Bean claimed"})
			}
			command.Printf("Claimed %s\n", bean.ID)
			return nil
		},
	}
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

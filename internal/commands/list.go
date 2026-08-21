package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type listOptions struct {
	statuses []string
	types    []string
	json     bool
}

func newListCommand() *cobra.Command {
	options := listOptions{}
	command := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List Beans-format tasks",
		Args:    cobra.NoArgs,
		RunE: func(command *cobra.Command, args []string) error {
			if err := validateListOptions(options); err != nil {
				return err
			}
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			loaded, err := beans.Load(workingDirectory)
			if err != nil {
				return err
			}
			filtered := filterBeans(loaded, options)
			sortBeans(filtered)
			if options.json {
				for index := range filtered {
					filtered[index].Body = ""
				}
				return json.NewEncoder(command.OutOrStdout()).Encode(filtered)
			}
			if len(filtered) == 0 {
				command.Println("No beans found.")
				return nil
			}
			for _, bean := range filtered {
				parent := bean.Parent
				if parent == "" {
					parent = "-"
				}
				command.Printf("%s  %s  %s  %s  %s\n", bean.ID, bean.Status, bean.Type, parent, bean.Title)
			}
			return nil
		},
	}
	command.Flags().StringArrayVarP(&options.statuses, "status", "s", nil, "Status to include (repeatable)")
	command.Flags().StringArrayVarP(&options.types, "type", "t", nil, "Type to include (repeatable)")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

func validateListOptions(options listOptions) error {
	for _, status := range options.statuses {
		if !beanStatuses[status] {
			return fmt.Errorf("invalid status %q", status)
		}
	}
	for _, typeName := range options.types {
		if !beanTypes[typeName] {
			return fmt.Errorf("invalid type %q", typeName)
		}
	}
	return nil
}

func filterBeans(loaded []beans.Bean, options listOptions) []beans.Bean {
	filtered := make([]beans.Bean, 0, len(loaded))
	for _, bean := range loaded {
		if len(options.statuses) > 0 && !slices.Contains(options.statuses, bean.Status) {
			continue
		}
		if len(options.types) > 0 && !slices.Contains(options.types, bean.Type) {
			continue
		}
		filtered = append(filtered, bean)
	}
	return filtered
}

func sortBeans(loaded []beans.Bean) {
	statusOrder := map[string]int{"in-progress": 0, "todo": 1, "draft": 2, "completed": 3, "scrapped": 4}
	priorityOrder := map[string]int{"critical": 0, "high": 1, "normal": 2, "low": 3, "deferred": 4}
	typeOrder := map[string]int{"milestone": 0, "epic": 1, "bug": 2, "feature": 3, "task": 4}
	slices.SortFunc(loaded, func(left, right beans.Bean) int {
		for _, comparison := range []int{
			statusOrderValue(statusOrder, left.Status) - statusOrderValue(statusOrder, right.Status),
			statusOrderValue(priorityOrder, left.Priority) - statusOrderValue(priorityOrder, right.Priority),
			statusOrderValue(typeOrder, left.Type) - statusOrderValue(typeOrder, right.Type),
		} {
			if comparison != 0 {
				return comparison
			}
		}
		if comparison := strings.Compare(strings.ToLower(left.Title), strings.ToLower(right.Title)); comparison != 0 {
			return comparison
		}
		return strings.Compare(left.ID, right.ID)
	})
}

func statusOrderValue(order map[string]int, value string) int {
	if index, found := order[value]; found {
		return index
	}
	return len(order)
}

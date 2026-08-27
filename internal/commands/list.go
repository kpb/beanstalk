package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

type listOptions struct {
	statuses []string
	types    []string
	json     bool
}

type listRow struct {
	bean  beans.Bean
	title string
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
			beans.Sort(loaded)
			filtered := filterBeans(loaded, options)
			if options.json {
				for index := range filtered {
					filtered[index].Body = ""
				}
				return json.NewEncoder(command.OutOrStdout()).Encode(filtered)
			}
			rows, err := listHierarchy(loaded, filtered)
			if err != nil {
				return err
			}
			if len(rows) == 0 {
				command.Println("No beans found.")
				return nil
			}
			for _, row := range rows {
				bean := row.bean
				parent := bean.Parent
				if parent == "" {
					parent = "-"
				}
				command.Printf("%s  %s  %s  %s  %s\n", bean.ID, bean.Status, bean.Type, parent, row.title)
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

func listHierarchy(loaded, matched []beans.Bean) ([]listRow, error) {
	byID := make(map[string]beans.Bean, len(loaded))
	for _, bean := range loaded {
		if _, found := byID[bean.ID]; found {
			return nil, fmt.Errorf("%w: %s", beans.ErrDuplicateBeanID, bean.ID)
		}
		byID[bean.ID] = bean
	}

	included := make(map[string]bool, len(matched))
	for _, bean := range matched {
		for id, visited := bean.ID, map[string]bool{}; id != ""; id = byID[id].Parent {
			if visited[id] {
				return nil, fmt.Errorf("%w: %s", beans.ErrParentCycle, id)
			}
			visited[id] = true
			parent, found := byID[id]
			if !found {
				return nil, fmt.Errorf("parent bean not found: %s", id)
			}
			included[parent.ID] = true
		}
	}

	children := make(map[string][]beans.Bean, len(included))
	roots := make([]beans.Bean, 0, len(included))
	for _, bean := range loaded {
		if !included[bean.ID] {
			continue
		}
		if bean.Parent == "" {
			roots = append(roots, bean)
			continue
		}
		children[bean.Parent] = append(children[bean.Parent], bean)
	}

	rows := make([]listRow, 0, len(included))
	var visit func(beans.Bean, string, bool)
	visit = func(bean beans.Bean, prefix string, last bool) {
		title := bean.Title
		if bean.Parent != "" {
			connector := "|- "
			if last {
				connector = "`- "
			}
			title = prefix + connector + title
		}
		rows = append(rows, listRow{bean: bean, title: title})

		childPrefix := prefix
		if bean.Parent != "" {
			if last {
				childPrefix += "   "
			} else {
				childPrefix += "|  "
			}
		}
		for index, child := range children[bean.ID] {
			visit(child, childPrefix, index == len(children[bean.ID])-1)
		}
	}
	for index, root := range roots {
		visit(root, "", index == len(roots)-1)
	}
	return rows, nil
}

package commands

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/kpb/beanstalk/internal/beans"
	"github.com/spf13/cobra"
)

const defaultBeanStatus = "todo"
const defaultBeanType = "task"

var (
	beanStatuses   = map[string]bool{"todo": true, "draft": true, "in-progress": true, "completed": true, "scrapped": true}
	beanTypes      = map[string]bool{"milestone": true, "epic": true, "bug": true, "feature": true, "task": true}
	beanPriorities = map[string]bool{"critical": true, "high": true, "normal": true, "low": true, "deferred": true}
)

type createOptions struct {
	status   string
	typeName string
	priority string
	parent   string
	body     string
	tags     []string
	json     bool
}

func newCreateCommand() *cobra.Command {
	options := createOptions{}
	command := &cobra.Command{
		Use:     "create [title]",
		Aliases: []string{"c", "new"},
		Short:   "Create a Beans-format task",
		RunE: func(command *cobra.Command, args []string) error {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}

			bean, err := createBean(workingDirectory, strings.Join(args, " "), options)
			if err != nil {
				return err
			}

			if options.json {
				return json.NewEncoder(command.OutOrStdout()).Encode(struct {
					Success bool       `json:"success"`
					Bean    beans.Bean `json:"bean"`
					Message string     `json:"message"`
				}{true, bean, "Bean created"})
			}

			command.Printf("Created %s %s\n", bean.ID, bean.Path)
			return nil
		},
	}
	command.Flags().StringVarP(&options.status, "status", "s", "", "Initial status")
	command.Flags().StringVarP(&options.typeName, "type", "t", "", "Bean type")
	command.Flags().StringVarP(&options.priority, "priority", "p", "", "Priority")
	command.Flags().StringVar(&options.parent, "parent", "", "Parent bean ID")
	command.Flags().StringVarP(&options.body, "body", "d", "", "Markdown body")
	command.Flags().StringArrayVar(&options.tags, "tag", nil, "Tag (repeatable)")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

func createBean(workingDirectory, title string, options createOptions) (beans.Bean, error) {
	config, err := beans.LoadConfig(workingDirectory)
	if err != nil {
		return beans.Bean{}, err
	}
	if title == "" {
		title = "Untitled"
	}

	status := options.status
	if status == "" {
		status = config.Beans.DefaultStatus
	}
	if status == "" {
		status = defaultBeanStatus
	}
	if !beanStatuses[status] {
		return beans.Bean{}, fmt.Errorf("invalid status %q", status)
	}

	typeName := options.typeName
	if typeName == "" {
		typeName = config.Beans.DefaultType
	}
	if typeName == "" {
		typeName = defaultBeanType
	}
	if !beanTypes[typeName] {
		return beans.Bean{}, fmt.Errorf("invalid type %q", typeName)
	}
	if options.priority != "" && !beanPriorities[options.priority] {
		return beans.Bean{}, fmt.Errorf("invalid priority %q", options.priority)
	}

	beansPath, err := beans.Directory(workingDirectory, config)
	if err != nil {
		return beans.Bean{}, err
	}

	idLength := config.Beans.IDLength
	if idLength == 0 {
		idLength = 4
	}
	if idLength < 1 {
		return beans.Bean{}, fmt.Errorf("invalid beans.id_length %d", idLength)
	}
	now := time.Now().UTC().Truncate(time.Second)
	for range 10 {
		id, err := newBeanID(config.Beans.Prefix, idLength)
		if err != nil {
			return beans.Bean{}, err
		}
		slug := beanSlug(title)
		name := id + "--" + slug + ".md"
		if slug == "" {
			name = id + ".md"
		}
		path := filepath.Join(beansPath, name)
		if err := beans.ValidateParent(workingDirectory, id, options.parent); err != nil {
			return beans.Bean{}, err
		}
		bean := beans.Bean{ID: id, Slug: slug, Title: title, Status: status, Type: typeName, Priority: options.priority, Tags: options.tags, Parent: options.parent, CreatedAt: now, UpdatedAt: now, Path: name, Body: options.body}
		contents, err := beans.Render(bean)
		if err != nil {
			return beans.Bean{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			_, writeErr := file.Write(contents)
			closeErr := file.Close()
			if writeErr != nil {
				return beans.Bean{}, fmt.Errorf("writing bean: %w", writeErr)
			}
			if closeErr != nil {
				return beans.Bean{}, fmt.Errorf("closing bean: %w", closeErr)
			}
			return bean, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return beans.Bean{}, fmt.Errorf("creating bean: %w", err)
		}
	}
	return beans.Bean{}, errors.New("could not generate a unique bean ID")
}

func newBeanID(prefix string, length int) (string, error) {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating bean ID: %w", err)
	}
	for index := range bytes {
		bytes[index] = alphabet[int(bytes[index])%len(alphabet)]
	}
	return prefix + string(bytes), nil
}

func beanSlug(title string) string {
	var slug strings.Builder
	lastDash := false
	for _, character := range strings.ToLower(title) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			slug.WriteRune(character)
			lastDash = false
		} else if !lastDash && slug.Len() > 0 {
			slug.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(strings.TrimSpace(slug.String()), "-")
}

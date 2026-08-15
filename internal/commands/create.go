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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultBeanStatus = "todo"
const defaultBeanType = "task"

var (
	beanStatuses   = map[string]bool{"todo": true, "draft": true, "in-progress": true, "completed": true, "scrapped": true}
	beanTypes      = map[string]bool{"milestone": true, "epic": true, "bug": true, "feature": true, "task": true}
	beanPriorities = map[string]bool{"critical": true, "high": true, "normal": true, "low": true, "deferred": true}
)

type beansConfig struct {
	Beans struct {
		Path          string `yaml:"path"`
		Prefix        string `yaml:"prefix"`
		IDLength      int    `yaml:"id_length"`
		DefaultStatus string `yaml:"default_status"`
		DefaultType   string `yaml:"default_type"`
	} `yaml:"beans"`
}

type createdBean struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
	Priority  string    `json:"priority,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Path      string    `json:"path"`
}

type createOptions struct {
	status   string
	typeName string
	priority string
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
					Success bool        `json:"success"`
					Bean    createdBean `json:"bean"`
					Message string      `json:"message"`
				}{true, bean, "Bean created"})
			}

			command.Printf("Created %s %s\n", bean.ID, bean.Path)
			return nil
		},
	}
	command.Flags().StringVarP(&options.status, "status", "s", "", "Initial status")
	command.Flags().StringVarP(&options.typeName, "type", "t", "", "Bean type")
	command.Flags().StringVarP(&options.priority, "priority", "p", "", "Priority")
	command.Flags().StringVarP(&options.body, "body", "d", "", "Markdown body")
	command.Flags().StringArrayVar(&options.tags, "tag", nil, "Tag (repeatable)")
	command.Flags().BoolVar(&options.json, "json", false, "Output JSON")
	return command
}

func createBean(workingDirectory, title string, options createOptions) (createdBean, error) {
	config, err := loadBeansConfig(workingDirectory)
	if err != nil {
		return createdBean{}, err
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
		return createdBean{}, fmt.Errorf("invalid status %q", status)
	}

	typeName := options.typeName
	if typeName == "" {
		typeName = config.Beans.DefaultType
	}
	if typeName == "" {
		typeName = defaultBeanType
	}
	if !beanTypes[typeName] {
		return createdBean{}, fmt.Errorf("invalid type %q", typeName)
	}
	if options.priority != "" && !beanPriorities[options.priority] {
		return createdBean{}, fmt.Errorf("invalid priority %q", options.priority)
	}

	beansPath := config.Beans.Path
	if beansPath == "" {
		beansPath = ".beans"
	}
	if !filepath.IsAbs(beansPath) {
		beansPath = filepath.Join(workingDirectory, beansPath)
	}
	if info, err := os.Stat(beansPath); err != nil {
		if os.IsNotExist(err) {
			return createdBean{}, errors.New("Beans project is not initialized; run beanstalk init first")
		}
		return createdBean{}, fmt.Errorf("checking beans directory: %w", err)
	} else if !info.IsDir() {
		return createdBean{}, fmt.Errorf("beans path is not a directory: %s", beansPath)
	}

	idLength := config.Beans.IDLength
	if idLength == 0 {
		idLength = 4
	}
	if idLength < 1 {
		return createdBean{}, fmt.Errorf("invalid beans.id_length %d", idLength)
	}
	now := time.Now().UTC().Truncate(time.Second)
	for range 10 {
		id, err := newBeanID(config.Beans.Prefix, idLength)
		if err != nil {
			return createdBean{}, err
		}
		slug := beanSlug(title)
		name := id + "--" + slug + ".md"
		if slug == "" {
			name = id + ".md"
		}
		path := filepath.Join(beansPath, name)
		bean := createdBean{ID: id, Title: title, Status: status, Type: typeName, Priority: options.priority, Tags: options.tags, CreatedAt: now, UpdatedAt: now, Path: name}
		contents, err := renderBean(bean, options.body)
		if err != nil {
			return createdBean{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err == nil {
			_, writeErr := file.Write(contents)
			closeErr := file.Close()
			if writeErr != nil {
				return createdBean{}, fmt.Errorf("writing bean: %w", writeErr)
			}
			if closeErr != nil {
				return createdBean{}, fmt.Errorf("closing bean: %w", closeErr)
			}
			return bean, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return createdBean{}, fmt.Errorf("creating bean: %w", err)
		}
	}
	return createdBean{}, errors.New("could not generate a unique bean ID")
}

func loadBeansConfig(workingDirectory string) (beansConfig, error) {
	path := filepath.Join(workingDirectory, ".beans.yml")
	contents, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return beansConfig{}, errors.New("Beans project is not initialized; run beanstalk init first")
		}
		return beansConfig{}, fmt.Errorf("reading .beans.yml: %w", err)
	}
	var config beansConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return beansConfig{}, fmt.Errorf("parsing .beans.yml: %w", err)
	}
	return config, nil
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

func renderBean(bean createdBean, body string) ([]byte, error) {
	frontMatter := struct {
		Title     string    `yaml:"title"`
		Status    string    `yaml:"status"`
		Type      string    `yaml:"type"`
		Priority  string    `yaml:"priority,omitempty"`
		Tags      []string  `yaml:"tags,omitempty"`
		CreatedAt time.Time `yaml:"created_at"`
		UpdatedAt time.Time `yaml:"updated_at"`
	}{bean.Title, bean.Status, bean.Type, bean.Priority, bean.Tags, bean.CreatedAt, bean.UpdatedAt}
	metadata, err := yaml.Marshal(frontMatter)
	if err != nil {
		return nil, fmt.Errorf("encoding bean metadata: %w", err)
	}
	contents := fmt.Sprintf("---\n# %s\n%s---\n%s", bean.ID, metadata, body)
	return []byte(strings.TrimRight(contents, "\n") + "\n"), nil
}

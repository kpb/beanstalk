// Package beans reads and writes Beans-format task files.
package beans

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrNotInitialized      = errors.New("Beans project is not initialized; run beanstalk init first")
	ErrInvalidBeanStatus   = errors.New("invalid bean status")
	ErrInvalidBeanType     = errors.New("invalid bean type")
	ErrInvalidBeanPriority = errors.New("invalid bean priority")
)

var (
	validStatuses   = map[string]bool{"todo": true, "draft": true, "in-progress": true, "completed": true, "scrapped": true}
	validTypes      = map[string]bool{"milestone": true, "epic": true, "bug": true, "feature": true, "task": true}
	validPriorities = map[string]bool{"critical": true, "high": true, "normal": true, "low": true, "deferred": true}
)

type Config struct {
	Beans struct {
		Path          string `yaml:"path"`
		Prefix        string `yaml:"prefix"`
		IDLength      int    `yaml:"id_length"`
		DefaultStatus string `yaml:"default_status"`
		DefaultType   string `yaml:"default_type"`
	} `yaml:"beans"`
}

type Bean struct {
	ID        string    `json:"id" yaml:"-"`
	Slug      string    `json:"slug,omitempty" yaml:"-"`
	Path      string    `json:"path" yaml:"-"`
	Title     string    `json:"title" yaml:"title"`
	Status    string    `json:"status" yaml:"status"`
	Type      string    `json:"type,omitempty" yaml:"type"`
	Priority  string    `json:"priority,omitempty" yaml:"priority"`
	Tags      []string  `json:"tags,omitempty" yaml:"tags"`
	Parent    string    `json:"parent,omitempty" yaml:"parent,omitempty"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	Body      string    `json:"body,omitempty" yaml:"-"`
}

func LoadConfig(workingDirectory string) (Config, error) {
	contents, err := os.ReadFile(filepath.Join(workingDirectory, ".beans.yml"))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, ErrNotInitialized
		}
		return Config{}, fmt.Errorf("reading .beans.yml: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return Config{}, fmt.Errorf("parsing .beans.yml: %w", err)
	}
	return config, nil
}

func Directory(workingDirectory string, config Config) (string, error) {
	path := config.Beans.Path
	if path == "" {
		path = ".beans"
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(workingDirectory, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotInitialized
		}
		return "", fmt.Errorf("checking beans directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("beans path is not a directory: %s", path)
	}
	return path, nil
}

func Load(workingDirectory string) ([]Bean, error) {
	loaded, err := load(workingDirectory)
	if err != nil {
		return nil, err
	}
	if err := Validate(loaded); err != nil {
		return nil, err
	}
	return loaded, nil
}

func load(workingDirectory string) ([]Bean, error) {
	config, err := LoadConfig(workingDirectory)
	if err != nil {
		return nil, err
	}
	directory, err := Directory(workingDirectory, config)
	if err != nil {
		return nil, err
	}
	var loaded []Bean
	err = filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != directory && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}
		bean, err := parse(path, directory)
		if err != nil {
			return err
		}
		loaded = append(loaded, bean)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading beans: %w", err)
	}
	return loaded, nil
}

// Validate checks the metadata Beanstalk supports while retaining unknown front-matter fields.
func Validate(loaded []Bean) error {
	for _, bean := range loaded {
		if !validStatuses[bean.Status] {
			return fmt.Errorf("%w %q: %s", ErrInvalidBeanStatus, bean.Status, bean.ID)
		}
		if !validTypes[bean.Type] {
			return fmt.Errorf("%w %q: %s", ErrInvalidBeanType, bean.Type, bean.ID)
		}
		if !validPriorities[bean.Priority] {
			return fmt.Errorf("%w %q: %s", ErrInvalidBeanPriority, bean.Priority, bean.ID)
		}
	}
	byID, children, err := hierarchy(loaded)
	if err != nil {
		return err
	}
	return validateAcyclic(byID, children)
}

func parse(path, directory string) (Bean, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Bean{}, fmt.Errorf("reading %s: %w", path, err)
	}
	metadata, body, err := splitFrontMatter(string(contents))
	if err != nil {
		return Bean{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	var bean Bean
	if err := yaml.Unmarshal([]byte(metadata), &bean); err != nil {
		return Bean{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if bean.Title == "" || bean.Status == "" {
		return Bean{}, fmt.Errorf("parsing %s: title and status are required", path)
	}
	base := strings.TrimSuffix(filepath.Base(path), ".md")
	bean.ID, bean.Slug = filenameParts(base)
	if bean.ID == "" {
		return Bean{}, fmt.Errorf("parsing %s: invalid bean filename", path)
	}
	relativePath, err := filepath.Rel(directory, path)
	if err != nil {
		return Bean{}, fmt.Errorf("getting relative path for %s: %w", path, err)
	}
	bean.Path = filepath.ToSlash(relativePath)
	bean.Body = strings.TrimSuffix(body, "\n")
	if bean.Type == "" {
		bean.Type = "task"
	}
	if bean.Priority == "" {
		bean.Priority = "normal"
	}
	if bean.Tags == nil {
		bean.Tags = []string{}
	}
	info, err := os.Stat(path)
	if err != nil {
		return Bean{}, fmt.Errorf("checking %s: %w", path, err)
	}
	if bean.CreatedAt.IsZero() {
		bean.CreatedAt = info.ModTime().UTC().Truncate(time.Second)
	}
	if bean.UpdatedAt.IsZero() {
		bean.UpdatedAt = bean.CreatedAt
	}
	return bean, nil
}

func splitFrontMatter(contents string) (string, string, error) {
	if !strings.HasPrefix(contents, "---\n") {
		return "", "", errors.New("missing opening front matter delimiter")
	}
	remaining := contents[len("---\n"):]
	index := strings.Index(remaining, "\n---\n")
	if index >= 0 {
		return remaining[:index], remaining[index+len("\n---\n"):], nil
	}
	if strings.HasSuffix(remaining, "\n---") {
		return strings.TrimSuffix(remaining, "\n---"), "", nil
	}
	return "", "", errors.New("missing closing front matter delimiter")
}

func filenameParts(filename string) (string, string) {
	if id, slug, found := strings.Cut(filename, "--"); found {
		return id, slug
	}
	return filename, ""
}

func Render(bean Bean) ([]byte, error) {
	frontMatter := struct {
		Title     string    `yaml:"title"`
		Status    string    `yaml:"status"`
		Type      string    `yaml:"type"`
		Priority  string    `yaml:"priority,omitempty"`
		Tags      []string  `yaml:"tags,omitempty"`
		Parent    string    `yaml:"parent,omitempty"`
		CreatedAt time.Time `yaml:"created_at"`
		UpdatedAt time.Time `yaml:"updated_at"`
	}{bean.Title, bean.Status, bean.Type, bean.Priority, bean.Tags, bean.Parent, bean.CreatedAt, bean.UpdatedAt}
	metadata, err := yaml.Marshal(frontMatter)
	if err != nil {
		return nil, fmt.Errorf("encoding bean metadata: %w", err)
	}
	contents := fmt.Sprintf("---\n# %s\n%s---\n%s", bean.ID, metadata, bean.Body)
	return []byte(strings.TrimRight(contents, "\n") + "\n"), nil
}

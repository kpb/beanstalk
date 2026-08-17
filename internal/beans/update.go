package beans

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
)

var (
	ErrBeanNotFound     = errors.New("bean not found")
	ErrDuplicateBeanID  = errors.New("multiple beans have the same ID")
	ErrBeanNotClaimable = errors.New("bean is not available to claim")
)

// Find returns the bean with the exact ID from the configured Beans directory.
func Find(workingDirectory, id string) (Bean, error) {
	loaded, err := Load(workingDirectory)
	if err != nil {
		return Bean{}, err
	}

	var found *Bean
	for index := range loaded {
		if loaded[index].ID != id {
			continue
		}
		if found != nil {
			return Bean{}, fmt.Errorf("%w: %s", ErrDuplicateBeanID, id)
		}
		found = &loaded[index]
	}
	if found == nil {
		return Bean{}, fmt.Errorf("%w: %s", ErrBeanNotFound, id)
	}
	return *found, nil
}

// UpdateStatus updates a bean's status while preserving its other front matter and body.
func UpdateStatus(workingDirectory, id, status string, updatedAt time.Time) (Bean, error) {
	_, path, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	lock, err := lockBean(path)
	if err != nil {
		return Bean{}, err
	}
	defer lock.Unlock()
	return updateStatusLocked(workingDirectory, id, status, updatedAt)
}

// Claim transitions a todo bean to in-progress without allowing another claimant to win the same task.
func Claim(workingDirectory, id string, updatedAt time.Time) (Bean, error) {
	_, path, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	lock, err := lockBean(path)
	if err != nil {
		return Bean{}, err
	}
	defer lock.Unlock()

	bean, _, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	if bean.Status != "todo" {
		return Bean{}, fmt.Errorf("%w: %s is %s", ErrBeanNotClaimable, id, bean.Status)
	}
	return updateStatusLocked(workingDirectory, id, "in-progress", updatedAt)
}

func findBeanPath(workingDirectory, id string) (Bean, string, error) {
	bean, err := Find(workingDirectory, id)
	if err != nil {
		return Bean{}, "", err
	}
	config, err := LoadConfig(workingDirectory)
	if err != nil {
		return Bean{}, "", err
	}
	directory, err := Directory(workingDirectory, config)
	if err != nil {
		return Bean{}, "", err
	}
	return bean, filepath.Join(directory, filepath.FromSlash(bean.Path)), nil
}

func lockBean(path string) (*flock.Flock, error) {
	lock := flock.New(path)
	if err := lock.Lock(); err != nil {
		return nil, fmt.Errorf("locking %s: %w", path, err)
	}
	return lock, nil
}

func updateStatusLocked(workingDirectory, id, status string, updatedAt time.Time) (Bean, error) {
	bean, path, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	updatedAt = updatedAt.UTC().Truncate(time.Second)
	if err := updateStatusFile(path, status, updatedAt); err != nil {
		return Bean{}, err
	}
	bean.Status = status
	bean.UpdatedAt = updatedAt
	return bean, nil
}

func updateStatusFile(path, status string, updatedAt time.Time) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}
	metadata, body, err := splitFrontMatter(string(contents))
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	var document yaml.Node
	if err := yaml.Unmarshal([]byte(metadata), &document); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}
	if err := setMetadataValue(&document, "status", status); err != nil {
		return fmt.Errorf("updating %s: %w", path, err)
	}
	if err := setMetadataValue(&document, "updated_at", updatedAt.Format(time.RFC3339)); err != nil {
		return fmt.Errorf("updating %s: %w", path, err)
	}
	encoded, err := yaml.Marshal(&document)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("checking %s: %w", path, err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".bean-*")
	if err != nil {
		return fmt.Errorf("creating temporary bean file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		temporary.Close()
		return fmt.Errorf("setting temporary bean permissions: %w", err)
	}
	if _, err := temporary.Write([]byte("---\n" + string(encoded) + "---\n" + body)); err != nil {
		temporary.Close()
		return fmt.Errorf("writing temporary bean file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("syncing temporary bean file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing temporary bean file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replacing bean file: %w", err)
	}
	if err := syncDirectory(filepath.Dir(path)); err != nil {
		return fmt.Errorf("syncing bean directory: %w", err)
	}
	return nil
}

func setMetadataValue(document *yaml.Node, key, value string) error {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return errors.New("front matter must be a YAML mapping")
	}
	mapping := document.Content[0]
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value != key {
			continue
		}
		mapping.Content[index+1] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
		return nil
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
	)
	return nil
}

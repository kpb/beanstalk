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
	ErrParentCycle      = errors.New("parent link would create a cycle")
)

const parentLockFilename = ".beanstalk-parent.lock"

type UpdateFields struct {
	Status *string
	Parent *string
}

// Find returns the bean with the exact ID from the configured Beans directory.
func Find(workingDirectory, id string) (Bean, error) {
	loaded, err := load(workingDirectory)
	if err != nil {
		return Bean{}, err
	}
	return find(loaded, id)
}

func find(loaded []Bean, id string) (Bean, error) {
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
	return Update(workingDirectory, id, UpdateFields{Status: &status}, updatedAt)
}

// Update changes supported bean metadata while preserving other front matter and body.
func Update(workingDirectory, id string, fields UpdateFields, updatedAt time.Time) (Bean, error) {
	if fields.Parent != nil {
		lock, err := lockParents(workingDirectory)
		if err != nil {
			return Bean{}, err
		}
		defer lock.Unlock()
	}
	_, path, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	lock, err := lockBean(path)
	if err != nil {
		return Bean{}, err
	}
	defer lock.Unlock()
	if fields.Parent != nil {
		if err := ValidateParent(workingDirectory, id, *fields.Parent); err != nil {
			return Bean{}, err
		}
	}
	return updateLocked(workingDirectory, id, fields, updatedAt)
}

func lockParents(workingDirectory string) (*flock.Flock, error) {
	config, err := LoadConfig(workingDirectory)
	if err != nil {
		return nil, err
	}
	directory, err := Directory(workingDirectory, config)
	if err != nil {
		return nil, err
	}
	return lockBean(filepath.Join(directory, parentLockFilename))
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
	return updateLocked(workingDirectory, id, UpdateFields{Status: &status}, updatedAt)
}

func updateLocked(workingDirectory, id string, fields UpdateFields, updatedAt time.Time) (Bean, error) {
	bean, path, err := findBeanPath(workingDirectory, id)
	if err != nil {
		return Bean{}, err
	}
	updatedAt = updatedAt.UTC().Truncate(time.Second)
	if err := updateBeanFile(path, fields, updatedAt); err != nil {
		return Bean{}, err
	}
	if fields.Status != nil {
		bean.Status = *fields.Status
	}
	if fields.Parent != nil {
		bean.Parent = *fields.Parent
	}
	bean.UpdatedAt = updatedAt
	return bean, nil
}

func updateBeanFile(path string, fields UpdateFields, updatedAt time.Time) error {
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
	if fields.Status != nil {
		if err := setMetadataValue(&document, "status", *fields.Status); err != nil {
			return fmt.Errorf("updating %s: %w", path, err)
		}
	}
	if fields.Parent != nil {
		if *fields.Parent == "" {
			if err := deleteMetadataValue(&document, "parent"); err != nil {
				return fmt.Errorf("updating %s: %w", path, err)
			}
		} else if err := setMetadataValue(&document, "parent", *fields.Parent); err != nil {
			return fmt.Errorf("updating %s: %w", path, err)
		}
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

// ValidateParent ensures that parent names one existing bean without creating a cycle.
func ValidateParent(workingDirectory, id, parent string) error {
	if parent == "" {
		return nil
	}
	if parent == id {
		return fmt.Errorf("%w: %s", ErrParentCycle, id)
	}

	loaded, err := Load(workingDirectory)
	if err != nil {
		return err
	}
	byID := make(map[string]Bean, len(loaded))
	for _, bean := range loaded {
		if _, found := byID[bean.ID]; found {
			return fmt.Errorf("%w: %s", ErrDuplicateBeanID, bean.ID)
		}
		byID[bean.ID] = bean
	}

	visited := map[string]bool{id: true}
	for parent != "" {
		if visited[parent] {
			return fmt.Errorf("%w: %s", ErrParentCycle, parent)
		}
		visited[parent] = true
		bean, found := byID[parent]
		if !found {
			return fmt.Errorf("parent bean not found: %s", parent)
		}
		parent = bean.Parent
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

func deleteMetadataValue(document *yaml.Node, key string) error {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return errors.New("front matter must be a YAML mapping")
	}
	mapping := document.Content[0]
	for index := 0; index < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value != key {
			continue
		}
		mapping.Content = append(mapping.Content[:index], mapping.Content[index+2:]...)
		return nil
	}
	return nil
}

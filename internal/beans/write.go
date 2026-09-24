package beans

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomically replaces path with contents after fully writing it to a temporary file.
func WriteFileAtomically(path string, contents []byte, mode os.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".bean-*")
	if err != nil {
		return fmt.Errorf("creating temporary bean file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return fmt.Errorf("setting temporary bean permissions: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
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

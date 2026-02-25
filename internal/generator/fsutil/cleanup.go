package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanOutputDirPreserveGit removes all top-level entries in outputDir except
// .git, so generation can safely target an existing git repository root.
func CleanOutputDirPreserveGit(outputDir string) error {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return fmt.Errorf("output directory is empty")
	}

	info, err := os.Stat(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			if mkErr := os.MkdirAll(outputDir, 0o755); mkErr != nil {
				return fmt.Errorf("create output directory %q: %w", outputDir, mkErr)
			}
			return nil
		}
		return fmt.Errorf("stat output directory %q: %w", outputDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("output path %q is not a directory", outputDir)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("read output directory %q: %w", outputDir, err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if name == ".git" {
			continue
		}
		target := filepath.Join(outputDir, name)
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove output entry %q: %w", target, err)
		}
	}
	return nil
}

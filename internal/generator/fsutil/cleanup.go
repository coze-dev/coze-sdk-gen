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
	return CleanOutputDirPreserveEntries(outputDir, []string{".git"})
}

// CleanOutputDirPreserveEntries removes all top-level entries in outputDir
// except the given preserve entries. Preserve entries can be exact names or
// glob patterns (for example, "*_test.go"), both matched against top-level
// entry names.
func CleanOutputDirPreserveEntries(outputDir string, preserveEntries []string) error {
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
	preserve := map[string]struct{}{}
	preservePatterns := make([]string, 0)
	for _, entry := range preserveEntries {
		name := topLevelEntryName(entry)
		if name == "" {
			continue
		}
		if isGlobPattern(name) {
			preservePatterns = append(preservePatterns, name)
			continue
		}
		preserve[name] = struct{}{}
	}
	for _, entry := range entries {
		name := entry.Name()
		if _, ok := preserve[name]; ok {
			continue
		}
		if matchesGlobPatterns(name, preservePatterns) {
			continue
		}
		target := filepath.Join(outputDir, name)
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("remove output entry %q: %w", target, err)
		}
	}
	return nil
}

func topLevelEntryName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.Clean(trimmed)
	if cleaned == "." || cleaned == ".." {
		return ""
	}
	if strings.HasPrefix(cleaned, "../") {
		return ""
	}
	parts := strings.Split(cleaned, "/")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func isGlobPattern(value string) bool {
	return strings.ContainsAny(value, "*?[")
}

func matchesGlobPatterns(name string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

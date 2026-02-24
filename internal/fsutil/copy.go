package fsutil

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type CopyResult struct {
	CopiedFiles int
}

func CopySelected(srcRoot string, dstRoot string, includes []string, excludes []string) (CopyResult, error) {
	result := CopyResult{}
	if len(includes) == 0 {
		return result, fmt.Errorf("includes should not be empty")
	}

	if err := os.MkdirAll(dstRoot, 0o755); err != nil {
		return result, fmt.Errorf("create destination root %q: %w", dstRoot, err)
	}

	seen := map[string]struct{}{}
	for _, include := range includes {
		include = normalizeRelative(include)
		if include == "" {
			continue
		}
		if _, ok := seen[include]; ok {
			continue
		}
		seen[include] = struct{}{}

		if include == "." {
			entries, err := os.ReadDir(srcRoot)
			if err != nil {
				return result, fmt.Errorf("read source root %q: %w", srcRoot, err)
			}
			for _, entry := range entries {
				rel := entry.Name()
				if isExcluded(rel, excludes) {
					continue
				}
				if err := copyPath(srcRoot, dstRoot, rel, excludes, &result.CopiedFiles); err != nil {
					return result, err
				}
			}
			continue
		}

		if isExcluded(include, excludes) {
			continue
		}

		if err := copyPath(srcRoot, dstRoot, include, excludes, &result.CopiedFiles); err != nil {
			return result, err
		}
	}

	return result, nil
}

func copyPath(srcRoot, dstRoot, rel string, excludes []string, copiedFiles *int) error {
	rel = normalizeRelative(rel)
	srcPath := filepath.Join(srcRoot, rel)
	info, err := os.Lstat(srcPath)
	if err != nil {
		return fmt.Errorf("stat source path %q: %w", srcPath, err)
	}

	if info.IsDir() {
		return filepath.WalkDir(srcPath, func(pathName string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			relPath, err := filepath.Rel(srcRoot, pathName)
			if err != nil {
				return err
			}
			relPath = normalizeRelative(relPath)
			if relPath == "." {
				return nil
			}

			if isExcluded(relPath, excludes) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			sourceInfo, err := d.Info()
			if err != nil {
				return err
			}
			targetPath := filepath.Join(dstRoot, relPath)
			if d.IsDir() {
				if err := os.MkdirAll(targetPath, sourceInfo.Mode().Perm()); err != nil {
					return fmt.Errorf("create directory %q: %w", targetPath, err)
				}
				return nil
			}

			if err := copyFile(pathName, targetPath, sourceInfo.Mode()); err != nil {
				return err
			}
			*copiedFiles = *copiedFiles + 1
			return nil
		})
	}

	targetPath := filepath.Join(dstRoot, rel)
	if err := copyFile(srcPath, targetPath, info.Mode()); err != nil {
		return err
	}
	*copiedFiles = *copiedFiles + 1
	return nil
}

func copyFile(srcPath, dstPath string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create parent directory %q: %w", filepath.Dir(dstPath), err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", srcPath, err)
	}
	defer srcFile.Close()

	tmpPath := dstPath + ".tmp"
	dstFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
	if err != nil {
		return fmt.Errorf("create target file %q: %w", tmpPath, err)
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return fmt.Errorf("copy file %q to %q: %w", srcPath, dstPath, err)
	}
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("close target file %q: %w", tmpPath, err)
	}
	if err := os.Chmod(tmpPath, mode.Perm()); err != nil {
		return fmt.Errorf("chmod target file %q: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		return fmt.Errorf("rename %q to %q: %w", tmpPath, dstPath, err)
	}
	return nil
}

func isExcluded(rel string, excludes []string) bool {
	rel = normalizeRelative(rel)
	if rel == ".git" || strings.HasPrefix(rel, ".git/") {
		return true
	}

	for _, pattern := range excludes {
		pattern = normalizeRelative(pattern)
		if pattern == "" {
			continue
		}
		if rel == pattern || strings.HasPrefix(rel, pattern+"/") {
			return true
		}
		if matched, _ := path.Match(pattern, rel); matched {
			return true
		}
	}
	return false
}

func normalizeRelative(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = filepath.ToSlash(filepath.Clean(value))
	if value == "/" {
		return "."
	}
	value = strings.TrimPrefix(value, "./")
	if value == "" {
		return "."
	}
	return value
}

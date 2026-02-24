package dirdiff

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type DifferenceType string

const (
	MissingInTarget DifferenceType = "missing_in_target"
	ExtraInTarget   DifferenceType = "extra_in_target"
	TypeMismatch    DifferenceType = "type_mismatch"
	ContentMismatch DifferenceType = "content_mismatch"
	ModeMismatch    DifferenceType = "mode_mismatch"
)

type Difference struct {
	Path string
	Type DifferenceType
}

type snapshotEntry struct {
	IsDir bool
	Mode  fs.FileMode
	Hash  [32]byte
}

func CompareDirs(source string, target string, excludes []string) ([]Difference, error) {
	srcSnapshot, err := snapshot(source, excludes)
	if err != nil {
		return nil, err
	}
	tgtSnapshot, err := snapshot(target, excludes)
	if err != nil {
		return nil, err
	}

	allKeys := map[string]struct{}{}
	for k := range srcSnapshot {
		allKeys[k] = struct{}{}
	}
	for k := range tgtSnapshot {
		allKeys[k] = struct{}{}
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	diffs := make([]Difference, 0)
	for _, key := range keys {
		srcEntry, srcOK := srcSnapshot[key]
		tgtEntry, tgtOK := tgtSnapshot[key]
		switch {
		case srcOK && !tgtOK:
			diffs = append(diffs, Difference{Path: key, Type: MissingInTarget})
		case !srcOK && tgtOK:
			diffs = append(diffs, Difference{Path: key, Type: ExtraInTarget})
		default:
			if srcEntry.IsDir != tgtEntry.IsDir {
				diffs = append(diffs, Difference{Path: key, Type: TypeMismatch})
				continue
			}
			if srcEntry.Mode.Perm() != tgtEntry.Mode.Perm() {
				diffs = append(diffs, Difference{Path: key, Type: ModeMismatch})
			}
			if !srcEntry.IsDir && srcEntry.Hash != tgtEntry.Hash {
				diffs = append(diffs, Difference{Path: key, Type: ContentMismatch})
			}
		}
	}

	return diffs, nil
}

func snapshot(root string, excludes []string) (map[string]snapshotEntry, error) {
	entries := map[string]snapshotEntry{}
	err := filepath.WalkDir(root, func(pathName string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, pathName)
		if err != nil {
			return err
		}
		rel = normalizeRelative(rel)
		if rel == "." {
			return nil
		}

		if isExcluded(rel, excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		entry := snapshotEntry{
			IsDir: d.IsDir(),
			Mode:  info.Mode(),
		}
		if !d.IsDir() {
			hashValue, err := fileHash(pathName)
			if err != nil {
				return err
			}
			entry.Hash = hashValue
		}
		entries[rel] = entry
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("build snapshot for %q: %w", root, err)
	}
	return entries, nil
}

func fileHash(pathName string) ([32]byte, error) {
	f, err := os.Open(pathName)
	if err != nil {
		return [32]byte{}, fmt.Errorf("open file %q: %w", pathName, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, fmt.Errorf("hash file %q: %w", pathName, err)
	}

	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum, nil
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

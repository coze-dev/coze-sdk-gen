package dirdiff

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareDirsNoDiff(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "a", "b.txt"), "hello")
	writeFile(t, filepath.Join(dst, "a", "b.txt"), "hello")

	diffs, err := CompareDirs(src, dst, nil)
	if err != nil {
		t.Fatalf("CompareDirs() error = %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no differences, got %v", diffs)
	}
}

func TestCompareDirsWithDiffs(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "same.txt"), "same")
	writeFile(t, filepath.Join(dst, "same.txt"), "same")

	writeFile(t, filepath.Join(src, "only-src.txt"), "src")
	writeFile(t, filepath.Join(dst, "only-dst.txt"), "dst")
	writeFile(t, filepath.Join(src, "changed.txt"), "source")
	writeFile(t, filepath.Join(dst, "changed.txt"), "target")

	diffs, err := CompareDirs(src, dst, nil)
	if err != nil {
		t.Fatalf("CompareDirs() error = %v", err)
	}
	if len(diffs) < 3 {
		t.Fatalf("expected at least 3 differences, got %v", diffs)
	}

	var hasMissing, hasExtra, hasContent bool
	for _, diff := range diffs {
		switch diff.Type {
		case MissingInTarget:
			hasMissing = true
		case ExtraInTarget:
			hasExtra = true
		case ContentMismatch:
			hasContent = true
		}
	}
	if !hasMissing || !hasExtra || !hasContent {
		t.Fatalf("unexpected diff types: %v", diffs)
	}
}

func TestCompareDirsExclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "ignored", "x.txt"), "1")
	writeFile(t, filepath.Join(dst, "ignored", "x.txt"), "2")
	writeFile(t, filepath.Join(src, "kept.txt"), "k")
	writeFile(t, filepath.Join(dst, "kept.txt"), "k")

	diffs, err := CompareDirs(src, dst, []string{"ignored"})
	if err != nil {
		t.Fatalf("CompareDirs() error = %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("expected no differences after exclude, got %v", diffs)
	}
}

func TestCompareDirsModeMismatch(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "run.sh")
	dstFile := filepath.Join(dst, "run.sh")
	writeFile(t, srcFile, "echo hi")
	writeFile(t, dstFile, "echo hi")

	if err := os.Chmod(srcFile, 0o755); err != nil {
		t.Fatalf("chmod src file: %v", err)
	}
	if err := os.Chmod(dstFile, 0o644); err != nil {
		t.Fatalf("chmod dst file: %v", err)
	}

	diffs, err := CompareDirs(src, dst, nil)
	if err != nil {
		t.Fatalf("CompareDirs() error = %v", err)
	}

	found := false
	for _, diff := range diffs {
		if diff.Type == ModeMismatch && diff.Path == "run.sh" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mode mismatch diff, got %v", diffs)
	}
}

func TestCompareDirsMissingRoot(t *testing.T) {
	_, err := CompareDirs(filepath.Join(t.TempDir(), "missing"), t.TempDir(), nil)
	if err == nil {
		t.Fatal("expected error when source root is missing")
	}
}

func writeFile(t *testing.T, pathName string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(pathName), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", pathName, err)
	}
	if err := os.WriteFile(pathName, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", pathName, err)
	}
}

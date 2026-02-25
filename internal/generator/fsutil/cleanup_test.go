package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanOutputDirPreserveGit(t *testing.T) {
	out := t.TempDir()

	gitHead := filepath.Join(out, ".git", "HEAD")
	if err := os.MkdirAll(filepath.Dir(gitHead), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(gitHead, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write git head: %v", err)
	}
	stale := filepath.Join(out, "stale.txt")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	if err := CleanOutputDirPreserveGit(out); err != nil {
		t.Fatalf("CleanOutputDirPreserveGit() error = %v", err)
	}

	if _, err := os.Stat(gitHead); err != nil {
		t.Fatalf("expected %s to be preserved, stat error = %v", gitHead, err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be removed, stat error = %v", stale, err)
	}
}

func TestCleanOutputDirPreserveGitCreatesDir(t *testing.T) {
	out := filepath.Join(t.TempDir(), "new-output")
	if err := CleanOutputDirPreserveGit(out); err != nil {
		t.Fatalf("CleanOutputDirPreserveGit() error = %v", err)
	}
	if info, err := os.Stat(out); err != nil || !info.IsDir() {
		t.Fatalf("expected output dir to be created, stat err = %v", err)
	}
}

func TestCleanOutputDirPreserveEntries(t *testing.T) {
	out := t.TempDir()

	for _, rel := range []string{
		".github/workflows/ci.yml",
		"examples/demo.py",
		"tests/test_demo.py",
		"README.md",
		"cozepy/old.py",
	} {
		path := filepath.Join(out, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := CleanOutputDirPreserveEntries(
		out,
		[]string{".github/workflows", "examples", "tests/test_demo.py", "README.md"},
	); err != nil {
		t.Fatalf("CleanOutputDirPreserveEntries() error = %v", err)
	}

	for _, rel := range []string{
		".github/workflows/ci.yml",
		"examples/demo.py",
		"tests/test_demo.py",
		"README.md",
	} {
		if _, err := os.Stat(filepath.Join(out, rel)); err != nil {
			t.Fatalf("expected preserved path %s, stat error=%v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(out, "cozepy", "old.py")); !os.IsNotExist(err) {
		t.Fatalf("expected generated content to be cleaned, stat error=%v", err)
	}
}

func TestCleanOutputDirPreserveEntriesWithGlob(t *testing.T) {
	out := t.TempDir()

	for _, rel := range []string{
		"client_test.go",
		"request_test.go",
		"client.go",
		"README.md",
	} {
		path := filepath.Join(out, rel)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	if err := CleanOutputDirPreserveEntries(out, []string{"*_test.go"}); err != nil {
		t.Fatalf("CleanOutputDirPreserveEntries() error = %v", err)
	}

	for _, rel := range []string{"client_test.go", "request_test.go"} {
		if _, err := os.Stat(filepath.Join(out, rel)); err != nil {
			t.Fatalf("expected glob-preserved path %s, stat error=%v", rel, err)
		}
	}
	for _, rel := range []string{"client.go", "README.md"} {
		if _, err := os.Stat(filepath.Join(out, rel)); !os.IsNotExist(err) {
			t.Fatalf("expected non-matching file %s to be removed, stat error=%v", rel, err)
		}
	}
}

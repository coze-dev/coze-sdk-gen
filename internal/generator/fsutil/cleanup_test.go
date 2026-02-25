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

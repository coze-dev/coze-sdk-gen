package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopySelectedAll(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "README.md"), "doc")
	writeFile(t, filepath.Join(src, "cozepy", "a.py"), "print('a')")
	writeFile(t, filepath.Join(src, ".git", "HEAD"), "ref: refs/heads/main")

	result, err := CopySelected(src, dst, []string{"."}, nil)
	if err != nil {
		t.Fatalf("CopySelected() error = %v", err)
	}
	if result.CopiedFiles != 2 {
		t.Fatalf("expected 2 copied files, got %d", result.CopiedFiles)
	}

	assertFileContent(t, filepath.Join(dst, "README.md"), "doc")
	assertFileContent(t, filepath.Join(dst, "cozepy", "a.py"), "print('a')")
	if _, err := os.Stat(filepath.Join(dst, ".git", "HEAD")); !os.IsNotExist(err) {
		t.Fatalf("expected .git to be excluded, err=%v", err)
	}
}

func TestCopySelectedExcludePattern(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "cozepy", "a.py"), "A")
	writeFile(t, filepath.Join(src, "tests", "t.py"), "T")

	result, err := CopySelected(src, dst, []string{"."}, []string{"tests"})
	if err != nil {
		t.Fatalf("CopySelected() error = %v", err)
	}
	if result.CopiedFiles != 1 {
		t.Fatalf("expected 1 copied file, got %d", result.CopiedFiles)
	}
	if _, err := os.Stat(filepath.Join(dst, "tests", "t.py")); !os.IsNotExist(err) {
		t.Fatalf("expected tests directory to be excluded, err=%v", err)
	}
}

func TestCopySelectedMissingInclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if _, err := CopySelected(src, dst, []string{"missing"}, nil); err == nil {
		t.Fatal("expected error for missing include path")
	}
}

func TestCopySelectedEmptyIncludes(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if _, err := CopySelected(src, dst, nil, nil); err == nil {
		t.Fatal("expected error for empty includes")
	}
}

func TestCopySelectedSingleFileAndWildcardExclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "a.py"), "A")
	writeFile(t, filepath.Join(src, "test_a.py"), "TA")

	result, err := CopySelected(src, dst, []string{"a.py", "test_a.py"}, []string{"test_*"})
	if err != nil {
		t.Fatalf("CopySelected() error = %v", err)
	}
	if result.CopiedFiles != 1 {
		t.Fatalf("expected 1 copied file, got %d", result.CopiedFiles)
	}
	assertFileContent(t, filepath.Join(dst, "a.py"), "A")
	if _, err := os.Stat(filepath.Join(dst, "test_a.py")); !os.IsNotExist(err) {
		t.Fatalf("expected wildcard excluded file to be absent, err=%v", err)
	}
}

func TestCopySelectedNormalizeAndSkipExcludedInclude(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	writeFile(t, filepath.Join(src, "visible.txt"), "v")
	writeFile(t, filepath.Join(src, "ignored.txt"), "i")

	result, err := CopySelected(src, dst, []string{"./", "ignored.txt", "visible.txt"}, []string{"ignored.txt"})
	if err != nil {
		t.Fatalf("CopySelected() error = %v", err)
	}
	if result.CopiedFiles != 2 {
		t.Fatalf("expected 2 copied files because visible.txt is included twice, got %d", result.CopiedFiles)
	}
	assertFileContent(t, filepath.Join(dst, "visible.txt"), "v")
	if _, err := os.Stat(filepath.Join(dst, "ignored.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected ignored include file to be skipped, err=%v", err)
	}
}

func TestCopySelectedPreserveFileMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	srcFile := filepath.Join(src, "run.sh")
	writeFile(t, srcFile, "echo hi")
	if err := os.Chmod(srcFile, 0o755); err != nil {
		t.Fatalf("chmod source: %v", err)
	}

	if _, err := CopySelected(src, dst, []string{"run.sh"}, nil); err != nil {
		t.Fatalf("CopySelected() error = %v", err)
	}

	info, err := os.Stat(filepath.Join(dst, "run.sh"))
	if err != nil {
		t.Fatalf("stat copied file: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("expected mode 0755, got %o", info.Mode().Perm())
	}
}

func TestNormalizeRelative(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"./":   ".",
		"/":    ".",
		"./a":  "a",
		"a/b":  "a/b",
		"a//b": "a/b",
	}
	for in, want := range cases {
		if got := normalizeRelative(in); got != want {
			t.Fatalf("normalizeRelative(%q) = %q, want %q", in, got, want)
		}
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

func assertFileContent(t *testing.T, pathName string, expected string) {
	t.Helper()
	content, err := os.ReadFile(pathName)
	if err != nil {
		t.Fatalf("read %s: %v", pathName, err)
	}
	if string(content) != expected {
		t.Fatalf("unexpected file content for %s: %q", pathName, string(content))
	}
}

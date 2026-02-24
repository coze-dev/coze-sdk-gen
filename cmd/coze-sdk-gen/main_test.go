package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--version"}, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "dev") {
		t.Fatalf("expected version output, got %q", out.String())
	}
}

func TestRunGenerate(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	outDir := filepath.Join(tmp, "out")
	cfgPath := filepath.Join(tmp, "generator.yaml")
	swaggerPath := filepath.Join(tmp, "swagger.yaml")

	writeFile(t, filepath.Join(src, "cozepy", "a.py"), "A")
	writeFile(t, filepath.Join(src, "README.md"), "R")

	writeFile(t, swaggerPath, `
paths:
  /v3/chat:
    post:
      operationId: OpenApiChat
`)
	writeFile(t, cfgPath, `
language: python
source_sdk: `+src+`
output_sdk: `+outDir+`
api:
  packages:
    - name: chat
      source_dir: cozepy/chat
      path_prefixes:
        - /v3/chat
  operation_mappings:
    - path: /v3/chat
      method: post
      sdk_methods:
        - chat.create
        - chat.stream
`)

	var out bytes.Buffer
	err := run([]string{"--config", cfgPath, "--swagger", swaggerPath}, &out)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "generated_files=") {
		t.Fatalf("unexpected output: %q", out.String())
	}
	assertFileContains(t, filepath.Join(outDir, "cozepy", "chat", "__init__.py"), "def create")
}

func TestRunInvalidArgs(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--invalid-flag"}, &out); err == nil {
		t.Fatal("expected run() to return flag error")
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

func assertFileContains(t *testing.T, pathName string, expected string) {
	t.Helper()
	content, err := os.ReadFile(pathName)
	if err != nil {
		t.Fatalf("read %s: %v", pathName, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Fatalf("expected %q in %s, got %q", expected, pathName, string(content))
	}
}

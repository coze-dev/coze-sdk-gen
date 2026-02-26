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
	outDir := filepath.Join(tmp, "out")
	cfgPath := filepath.Join(tmp, "generator.yaml")
	swaggerPath := filepath.Join(tmp, "swagger.yaml")

	writeFile(t, swaggerPath, `
paths:
  /v3/chat:
    post:
      operationId: OpenApiChat
`)
	writeFile(t, cfgPath, `
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
	err := run([]string{
		"--config", cfgPath,
		"--swagger", swaggerPath,
		"--language", "python",
		"--output-sdk", outDir,
	}, &out)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if !strings.Contains(out.String(), "generated_files=") {
		t.Fatalf("unexpected output: %q", out.String())
	}
	assertFileContains(t, filepath.Join(outDir, "cozepy", "chat", "__init__.py"), "def create")
}

func TestRunGenerateUsesDefaultRequestCallArgOrder(t *testing.T) {
	tmp := t.TempDir()
	outDir := filepath.Join(tmp, "out")
	cfgPath := filepath.Join(tmp, "generator.yaml")
	swaggerPath := filepath.Join(tmp, "swagger.yaml")

	writeFile(t, swaggerPath, `
paths:
  /v1/demo:
    post:
      operationId: OpenApiDemoCreate
`)
	writeFile(t, cfgPath, `
api:
  packages:
    - name: demo
      source_dir: cozepy/demo
      path_prefixes:
        - /v1/demo
  operation_mappings:
    - path: /v1/demo
      method: post
      sdk_methods:
        - demo.create
      body_builder: raw
      body_fields:
        - name
      body_required_fields:
        - name
      arg_types:
        name: str
      response_type: DemoResp
`)

	var out bytes.Buffer
	err := run([]string{
		"--config", cfgPath,
		"--swagger", swaggerPath,
		"--language", "python",
		"--output-sdk", outDir,
	}, &out)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertFileContains(
		t,
		filepath.Join(outDir, "cozepy", "demo", "__init__.py"),
		`"post", url, False, cast=DemoResp, headers=headers, body=body`,
	)
}

func TestRunInvalidArgs(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"--invalid-flag"}, &out); err == nil {
		t.Fatal("expected run() to return flag error")
	}
}

func TestRunMissingRequiredRuntimeArgs(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"--config", "config/generator.yaml"}, &out)
	if err == nil || !strings.Contains(err.Error(), "--language is required") {
		t.Fatalf("expected missing language error, got: %v", err)
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

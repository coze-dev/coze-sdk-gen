package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/dirdiff"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestGeneratePython(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()

	writeFile(t, filepath.Join(src, "README.md"), "readme")
	writeFile(t, filepath.Join(src, "cozepy", "api.py"), "print('api')")

	cfg := &config.Config{
		Language:  "python",
		SourceSDK: src,
		OutputSDK: out,
		Copy: config.Copy{
			Include: []string{"."},
		},
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v3/chat",
					Method:     "post",
					SDKMethods: []string{"chat.create", "chat.stream"},
				},
			},
		},
	}
	doc := mustParseSwagger(t)

	result, err := GeneratePython(cfg, doc)
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}
	if result.CopiedFiles != 2 {
		t.Fatalf("expected 2 copied files, got %d", result.CopiedFiles)
	}
	assertFileContent(t, filepath.Join(out, "README.md"), "readme")
	assertFileContent(t, filepath.Join(out, "cozepy", "api.py"), "print('api')")
}

func TestGeneratePythonValidationFailure(t *testing.T) {
	cfg := &config.Config{
		Language:  "python",
		SourceSDK: t.TempDir(),
		OutputSDK: t.TempDir(),
		Copy: config.Copy{
			Include: []string{"."},
		},
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v1/not-exist",
					Method:     "post",
					SDKMethods: []string{"demo.method"},
				},
			},
		},
	}
	doc := mustParseSwagger(t)

	if _, err := GeneratePython(cfg, doc); err == nil {
		t.Fatal("expected GeneratePython() to fail with swagger mismatch")
	}
}

func TestGeneratePythonCopyFailure(t *testing.T) {
	cfg := &config.Config{
		Language:  "python",
		SourceSDK: filepath.Join(t.TempDir(), "missing"),
		OutputSDK: t.TempDir(),
		Copy: config.Copy{
			Include: []string{"."},
		},
	}
	doc := mustParseSwagger(t)
	if _, err := GeneratePython(cfg, doc); err == nil {
		t.Fatal("expected GeneratePython() to fail when source sdk does not exist")
	}
}

func TestGeneratePythonDiffFailure(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()

	writeFile(t, filepath.Join(src, "target.txt"), "data")

	cfg := &config.Config{
		Language:  "python",
		SourceSDK: src,
		OutputSDK: out,
		Copy: config.Copy{
			Include: []string{"."},
		},
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v3/chat",
					Method:     "post",
					SDKMethods: []string{"chat.create"},
				},
			},
		},
	}
	doc := mustParseSwagger(t)

	originCompare := compareDirs
	compareDirs = func(source string, target string, excludes []string) ([]dirdiff.Difference, error) {
		return []dirdiff.Difference{
			{Path: "target.txt", Type: dirdiff.ContentMismatch},
		}, nil
	}
	defer func() {
		compareDirs = originCompare
	}()

	if _, err := GeneratePython(cfg, doc); err == nil {
		t.Fatal("expected diff validation failure")
	}
}

func TestGeneratePythonCompareError(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()

	writeFile(t, filepath.Join(src, "target.txt"), "data")

	cfg := &config.Config{
		Language:  "python",
		SourceSDK: src,
		OutputSDK: out,
		Copy: config.Copy{
			Include: []string{"."},
		},
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v3/chat",
					Method:     "post",
					SDKMethods: []string{"chat.create"},
				},
			},
		},
	}
	doc := mustParseSwagger(t)

	originCompare := compareDirs
	compareDirs = func(source string, target string, excludes []string) ([]dirdiff.Difference, error) {
		return nil, os.ErrPermission
	}
	defer func() {
		compareDirs = originCompare
	}()

	if _, err := GeneratePython(cfg, doc); err == nil {
		t.Fatal("expected compare error")
	}
}

func TestRunUnsupportedLanguage(t *testing.T) {
	cfg := &config.Config{Language: "go"}
	if _, err := Run(cfg, mustParseSwagger(t)); err == nil {
		t.Fatal("expected Run() to fail for unimplemented go language")
	}
}

func TestRunNilConfig(t *testing.T) {
	if _, err := Run(nil, nil); err == nil {
		t.Fatal("expected Run() to fail for nil config")
	}
}

func TestRunUnknownLanguage(t *testing.T) {
	cfg := &config.Config{Language: "ruby"}
	if _, err := Run(cfg, mustParseSwagger(t)); err == nil {
		t.Fatal("expected Run() to fail for unsupported language")
	}
}

func mustParseSwagger(t *testing.T) *openapi.Document {
	t.Helper()
	doc, err := openapi.Parse([]byte(`
paths:
  /v3/chat:
    post:
      operationId: OpenApiChat
`))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}
	return doc
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

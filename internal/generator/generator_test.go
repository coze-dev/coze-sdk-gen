package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestGeneratePythonFromSwagger(t *testing.T) {
	out := t.TempDir()
	cfg := testConfig(out)
	doc := mustParseSwagger(t)

	result, err := GeneratePython(cfg, doc)
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}
	if result.GeneratedFiles == 0 {
		t.Fatal("expected generated files")
	}
	if result.GeneratedOps < 3 {
		t.Fatalf("expected >=3 generated operations, got %d", result.GeneratedOps)
	}

	assertFileContains(t, filepath.Join(out, "cozepy", "chat", "__init__.py"), "def create")
	assertFileContains(t, filepath.Join(out, "cozepy", "chat", "__init__.py"), "def stream")
	assertFileContains(t, filepath.Join(out, "cozepy", "coze.py"), "class Coze")
	assertFileContains(t, filepath.Join(out, "cozepy", "coze.py"), "def chat")
	assertFileContains(t, filepath.Join(out, "README.md"), "# Coze Python API SDK")
}

func TestGeneratePythonOnlyMapped(t *testing.T) {
	out := t.TempDir()
	cfg := testConfig(out)
	cfg.API.GenerateOnlyMapped = true
	doc := mustParseSwagger(t)

	result, err := GeneratePython(cfg, doc)
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}
	if result.GeneratedOps != 2 {
		t.Fatalf("expected 2 generated ops from mapping, got %d", result.GeneratedOps)
	}
}

func TestGeneratePythonValidationFailure(t *testing.T) {
	cfg := testConfig(t.TempDir())
	cfg.API.OperationMappings = append(cfg.API.OperationMappings, config.OperationMapping{
		Path:       "/v1/not-exist",
		Method:     "post",
		SDKMethods: []string{"chat.not_exist"},
	})

	if _, err := GeneratePython(cfg, mustParseSwagger(t)); err == nil {
		t.Fatal("expected swagger validation failure")
	}
}

func TestGeneratePythonNilDoc(t *testing.T) {
	cfg := testConfig(t.TempDir())
	if _, err := GeneratePython(cfg, nil); err == nil {
		t.Fatal("expected error for nil swagger")
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

func TestNameHelpers(t *testing.T) {
	if got := normalizePythonIdentifier("class"); got != "class_" {
		t.Fatalf("unexpected reserved keyword normalize result: %s", got)
	}
	if got := defaultMethodName("OpenApiChatCancel", "/v3/chat/cancel", "post"); got != "chat_cancel" {
		t.Fatalf("unexpected default method name: %s", got)
	}
	if got := defaultMethodName("", "/v1/workspaces/{workspace_id}", "get"); got != "workspaces" {
		t.Fatalf("unexpected default path-derived method name: %s", got)
	}
	if got := toSnake("OpenApiChatReq"); got != "open_api_chat_req" {
		t.Fatalf("unexpected snake case value: %s", got)
	}
}

func mustParseSwagger(t *testing.T) *openapi.Document {
	t.Helper()
	doc, err := openapi.Parse([]byte(`
components:
  schemas:
    OpenApiChatReq:
      type: object
      properties:
        bot_id:
          type: string
    OpenApiChatResp:
      type: object
      properties:
        id:
          type: string
paths:
  /v3/chat:
    post:
      operationId: OpenApiChat
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/OpenApiChatReq'
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OpenApiChatResp'
  /v3/chat/cancel:
    post:
      operationId: OpenApiChatCancel
      responses:
        "204":
          description: no content
  /v1/workspaces:
    get:
      operationId: OpenApiWorkspacesList
      responses:
        "200":
          description: ok
`))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}
	return doc
}

func testConfig(output string) *config.Config {
	return &config.Config{
		Language:  "python",
		OutputSDK: output,
		API: config.APIConfig{
			Packages: []config.Package{
				{
					Name:         "chat",
					SourceDir:    "cozepy/chat",
					PathPrefixes: []string{"/v3/chat"},
				},
				{
					Name:         "workspaces",
					SourceDir:    "cozepy/workspaces",
					PathPrefixes: []string{"/v1/workspaces"},
				},
			},
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v3/chat",
					Method:     "post",
					SDKMethods: []string{"chat.create", "chat.stream"},
				},
			},
		},
	}
}

func assertFileContains(t *testing.T, pathName string, expected string) {
	t.Helper()
	content := readFile(t, pathName)
	if !strings.Contains(content, expected) {
		t.Fatalf("expected %q in %s, got:\n%s", expected, pathName, content)
	}
}

func readFile(t *testing.T, pathName string) string {
	t.Helper()
	content, err := os.ReadFile(pathName)
	if err != nil {
		t.Fatalf("read %s: %v", pathName, err)
	}
	return string(content)
}

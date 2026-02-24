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
	if _, err := os.Stat(filepath.Join(out, "cozepy", "types.py")); !os.IsNotExist(err) {
		t.Fatalf("expected types.py to be absent, stat err=%v", err)
	}
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

	chatModule := readFile(t, filepath.Join(out, "cozepy", "chat", "__init__.py"))
	if strings.Contains(chatModule, "def open_api_chat_cancel") {
		t.Fatalf("did not expect default-generated cancel method in mapped-only mode:\n%s", chatModule)
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
	if got := normalizeClassName("open_api_chat_req"); got != "OpenApiChatReq" {
		t.Fatalf("unexpected class name: %s", got)
	}
	if got := defaultMethodName("OpenApiChatCancel", "/v3/chat/cancel", "post"); got != "chat_cancel" {
		t.Fatalf("unexpected default method name: %s", got)
	}
	if got := defaultMethodName("", "/v1/workspaces/{workspace_id}", "get"); got != "workspaces" {
		t.Fatalf("unexpected default path-derived method name: %s", got)
	}
	if got := normalizePackageDir("cozepy/chat/message", "chat"); got != "chat/message" {
		t.Fatalf("unexpected package dir normalize: %s", got)
	}
	if got := normalizePackageDir("", "chat"); got != "chat" {
		t.Fatalf("unexpected fallback package dir normalize: %s", got)
	}
}

func TestRenderOperationMethodAndTypeHelpers(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:    "/v1/demo/{demo_id}",
		Method:  "post",
		Summary: "line1\nline2",
		PathParameters: []openapi.ParameterSpec{
			{Name: "demo_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		QueryParameters: []openapi.ParameterSpec{
			{Name: "page_num", In: "query", Required: false, Schema: &openapi.Schema{Type: "integer"}},
		},
		HeaderParameters: []openapi.ParameterSpec{
			{Name: "X-Trace-Id", In: "header", Required: false, Schema: &openapi.Schema{Type: "string"}},
		},
		RequestBody:       &openapi.RequestBody{Required: false},
		RequestBodySchema: &openapi.Schema{Type: "object"},
		ResponseSchema:    &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}},
	}
	binding := operationBinding{
		PackageName: "demo",
		MethodName:  "run",
		Details:     details,
	}

	asyncCode := renderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "await self._requester.arequest") {
		t.Fatalf("unexpected async method code:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "\"\"\"line1 line2\"\"\"") {
		t.Fatalf("expected escaped docstring, got:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "header_values") {
		t.Fatalf("expected header merge code, got:\n%s", asyncCode)
	}

	syncCode := renderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "self._requester.request") {
		t.Fatalf("unexpected sync method code:\n%s", syncCode)
	}

	if got := pythonTypeForSchemaRequired(doc, &openapi.Schema{Type: "number"}); got != "float" {
		t.Fatalf("unexpected number type mapping: %s", got)
	}
	if got := pythonTypeForSchema(doc, &openapi.Schema{Type: "boolean"}, false); got != "Optional[bool]" {
		t.Fatalf("unexpected optional bool type mapping: %s", got)
	}
	if got := escapeDocstring("a\nb\"\"\""); got != "a b\"" {
		t.Fatalf("unexpected escaped docstring: %q", got)
	}
}

func TestGeneratePythonFromRealConfig(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	cfgPath := filepath.Join(root, "config", "generator.yaml")
	swaggerPath := filepath.Join(root, "exist-repo", "coze-openapi-swagger.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", cfgPath, err)
	}
	cfg.OutputSDK = t.TempDir()

	doc, err := openapi.Load(swaggerPath)
	if err != nil {
		t.Fatalf("openapi.Load(%q) error = %v", swaggerPath, err)
	}

	result, err := GeneratePython(cfg, doc)
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}
	if result.GeneratedOps < 50 {
		t.Fatalf("expected >=50 generated ops, got %d", result.GeneratedOps)
	}
	if result.GeneratedFiles < 100 {
		t.Fatalf("expected >=100 generated files, got %d", result.GeneratedFiles)
	}

	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "apps", "__init__.py"), "PublishStatus")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "folders", "__init__.py"), "children_count")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "conversations", "__init__.py"), "def messages")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "datasets", "__init__.py"), "def process")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "bots", "__init__.py"), "def _list_v1")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def cancel")
}

func TestRenderOperationMethodAdvancedOptions(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/{id}",
		Method: "post",
		PathParameters: []openapi.ParameterSpec{
			{Name: "id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		QueryParameters: []openapi.ParameterSpec{
			{Name: "status", In: "query", Required: false, Schema: &openapi.Schema{Type: "string"}},
		},
		HeaderParameters: []openapi.ParameterSpec{
			{Name: "X-Trace-Id", In: "header", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		RequestBody: &openapi.RequestBody{Required: true},
		RequestBodySchema: &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
	}
	binding := operationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			QueryFields: []config.OperationField{
				{Name: "status", Type: "str", Required: false, UseValue: true},
			},
			BodyFields:         []string{"name"},
			BodyRequiredFields: []string{"name"},
			DisableHeadersArg:  true,
			IgnoreHeaderParams: true,
			RequestStream:      true,
			DataField:          "data.items",
		},
	}

	code := renderOperationMethod(doc, binding, false)
	if strings.Contains(code, "headers: Optional[Dict[str, str]]") {
		t.Fatalf("did not expect headers arg when DisableHeadersArg=true:\n%s", code)
	}
	if strings.Contains(code, "X-Trace-Id") {
		t.Fatalf("did not expect header parameter merge when IgnoreHeaderParams=true:\n%s", code)
	}
	if !strings.Contains(code, "data_field=\"data.items\"") {
		t.Fatalf("expected data_field in request call:\n%s", code)
	}
	if !strings.Contains(code, "request(\"post\", url, True") {
		t.Fatalf("expected stream request flag True:\n%s", code)
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

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	pygen "github.com/coze-dev/coze-sdk-gen/internal/generator/python"
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
	if _, err := os.Stat(filepath.Join(out, "cozepy", "types.py")); !os.IsNotExist(err) {
		t.Fatalf("expected types.py to be absent, stat err=%v", err)
	}
}

func TestGeneratePythonAudioTranscriptionsAsyncCreateFromMapping(t *testing.T) {
	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "python"
	cfg.OutputSDK = t.TempDir()

	if _, err := GeneratePython(cfg, doc); err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}

	modulePath := filepath.Join(cfg.OutputSDK, "cozepy", "audio", "transcriptions", "__init__.py")
	module := readFile(t, modulePath)
	if !strings.Contains(module, "class AsyncTranscriptionsClient(object):") {
		t.Fatalf("expected async transcriptions client class, got:\n%s", module)
	}
	if !strings.Contains(module, "async def create(") {
		t.Fatalf("expected async create method, got:\n%s", module)
	}
	if got := strings.Count(module, "async def create("); got != 1 {
		t.Fatalf("expected exactly one async create method, got %d", got)
	}
	if !strings.Contains(module, "files = {\"file\": _try_fix_file(file)}") {
		t.Fatalf("expected generated files payload from mapping/rendering, got:\n%s", module)
	}
	if !strings.Contains(module, "return await self._requester.arequest(") {
		t.Fatalf("expected async request call, got:\n%s", module)
	}
}

func TestGeneratePythonPreservesConfiguredDiffIgnorePaths(t *testing.T) {
	out := t.TempDir()
	gitHead := filepath.Join(out, ".git", "HEAD")
	if err := os.MkdirAll(filepath.Dir(gitHead), 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	if err := os.WriteFile(gitHead, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write git head: %v", err)
	}
	readmePath := filepath.Join(out, "README.md")
	if err := os.WriteFile(readmePath, []byte("README"), 0o644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	examplePath := filepath.Join(out, "examples", "demo.py")
	if err := os.MkdirAll(filepath.Dir(examplePath), 0o755); err != nil {
		t.Fatalf("mkdir examples dir: %v", err)
	}
	if err := os.WriteFile(examplePath, []byte("print('demo')"), 0o644); err != nil {
		t.Fatalf("write example: %v", err)
	}
	testPath := filepath.Join(out, "tests", "test_demo.py")
	if err := os.MkdirAll(filepath.Dir(testPath), 0o755); err != nil {
		t.Fatalf("mkdir tests dir: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("def test_demo(): pass"), 0o644); err != nil {
		t.Fatalf("write test: %v", err)
	}
	staleFile := filepath.Join(out, "stale.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	cfg := testConfig(out)
	cfg.Diff.IgnorePathsByLanguage = map[string][]string{
		"python": {".git", "README.md", "examples", "tests"},
		"go":     {".git"},
	}
	doc := mustParseSwagger(t)
	if _, err := GeneratePython(cfg, doc); err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}

	if _, err := os.Stat(gitHead); err != nil {
		t.Fatalf("expected .git to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected README to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(examplePath); err != nil {
		t.Fatalf("expected examples to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(testPath); err != nil {
		t.Fatalf("expected tests to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat err=%v", err)
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

func TestGenerateGoFromSwagger(t *testing.T) {
	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "go"
	cfg.OutputSDK = t.TempDir()

	result, err := GenerateGo(cfg, doc)
	if err != nil {
		t.Fatalf("GenerateGo() error = %v", err)
	}
	if result.GeneratedFiles == 0 {
		t.Fatal("expected generated files")
	}
	if result.GeneratedOps < 3 {
		t.Fatalf("expected >=3 generated operations, got %d", result.GeneratedOps)
	}
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "client.go"), "type CozeAPI struct")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "apps.go"), "func (r *apps) List(")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "users.go"), "func (r *users) Me(")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "templates.go"), "func (r *templates) Duplicate(")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "go.mod"), "module github.com/coze-dev/coze-go")
}

func TestGenerateGoPreservesGitDirectory(t *testing.T) {
	out := t.TempDir()
	gitHead := filepath.Join(out, ".git", "HEAD")
	if err := os.MkdirAll(filepath.Dir(gitHead), 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	if err := os.WriteFile(gitHead, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write git head: %v", err)
	}
	staleFile := filepath.Join(out, "stale.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}
	customTestFile := filepath.Join(out, "custom_test.go")
	if err := os.WriteFile(customTestFile, []byte("package coze\n"), 0o644); err != nil {
		t.Fatalf("write custom test file: %v", err)
	}

	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "go"
	cfg.OutputSDK = out
	if _, err := GenerateGo(cfg, doc); err != nil {
		t.Fatalf("GenerateGo() error = %v", err)
	}

	if _, err := os.Stat(gitHead); err != nil {
		t.Fatalf("expected .git to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(customTestFile); err != nil {
		t.Fatalf("expected custom test file to be preserved, stat err=%v", err)
	}
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, stat err=%v", err)
	}
}

func TestGenerateGoPreservesIgnoredReadmeAndGithub(t *testing.T) {
	out := t.TempDir()
	readmePath := filepath.Join(out, "README.md")
	githubWorkflowPath := filepath.Join(out, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(githubWorkflowPath), 0o755); err != nil {
		t.Fatalf("mkdir .github workflow dir: %v", err)
	}
	if err := os.WriteFile(readmePath, []byte("keep-readme"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	if err := os.WriteFile(githubWorkflowPath, []byte("keep-github"), 0o644); err != nil {
		t.Fatalf("write github workflow: %v", err)
	}

	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "go"
	cfg.OutputSDK = out

	if _, err := GenerateGo(cfg, doc); err != nil {
		t.Fatalf("GenerateGo() error = %v", err)
	}

	readmeContent := readFile(t, readmePath)
	if readmeContent != "keep-readme" {
		t.Fatalf("expected README.md to be preserved, got %q", readmeContent)
	}
	githubWorkflowContent := readFile(t, githubWorkflowPath)
	if githubWorkflowContent != "keep-github" {
		t.Fatalf("expected .github workflow to be preserved, got %q", githubWorkflowContent)
	}
}

func TestGenerateGoValidationFailure(t *testing.T) {
	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "go"
	cfg.OutputSDK = t.TempDir()
	cfg.API.OperationMappings = append(cfg.API.OperationMappings, config.OperationMapping{
		Path:       "/v1/not-exist",
		Method:     "post",
		SDKMethods: []string{"chat.not_exist"},
	})

	if _, err := GenerateGo(cfg, doc); err == nil {
		t.Fatal("expected swagger validation failure")
	}
}

func TestRunGoLanguage(t *testing.T) {
	cfg, doc := mustLoadRealConfigAndSwagger(t)
	cfg.Language = "go"
	cfg.OutputSDK = t.TempDir()
	if _, err := Run(cfg, doc); err != nil {
		t.Fatalf("expected Run() to support go language, got error: %v", err)
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
	if got := pygen.NormalizePythonIdentifier("class"); got != "class_" {
		t.Fatalf("unexpected reserved keyword normalize result: %s", got)
	}
	if got := pygen.NormalizeClassName("open_api_chat_req"); got != "OpenApiChatReq" {
		t.Fatalf("unexpected class name: %s", got)
	}
	if got := pygen.DefaultMethodName("OpenApiChatCancel", "/v3/chat/cancel", "post"); got != "chat_cancel" {
		t.Fatalf("unexpected default method name: %s", got)
	}
	if got := pygen.DefaultMethodName("", "/v1/workspaces/{workspace_id}", "get"); got != "workspaces" {
		t.Fatalf("unexpected default path-derived method name: %s", got)
	}
	if got := pygen.NormalizePackageDir("cozepy/chat/message", "chat"); got != "chat/message" {
		t.Fatalf("unexpected package dir normalize: %s", got)
	}
	if got := pygen.NormalizePackageDir("", "chat"); got != "chat" {
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
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "run",
		Details:     details,
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "await self._requester.arequest") {
		t.Fatalf("unexpected async method code:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "\"\"\"\n        line1\n        line2\n        \"\"\"") {
		t.Fatalf("expected escaped docstring, got:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "header_values") {
		t.Fatalf("expected header merge code, got:\n%s", asyncCode)
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "self._requester.request") {
		t.Fatalf("unexpected sync method code:\n%s", syncCode)
	}

	if got := pygen.PythonTypeForSchemaRequired(doc, &openapi.Schema{Type: "number"}); got != "float" {
		t.Fatalf("unexpected number type mapping: %s", got)
	}
	if got := pygen.PythonTypeForSchema(doc, &openapi.Schema{Type: "boolean"}, false); got != "Optional[bool]" {
		t.Fatalf("unexpected optional bool type mapping: %s", got)
	}
	if got := pygen.EscapeDocstring("a\nb\"\"\""); got != "a b\"" {
		t.Fatalf("unexpected escaped docstring: %q", got)
	}
}

func TestGeneratePythonFromRealConfig(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	cfgPath := filepath.Join(root, "config", "generator.yaml")
	swaggerPath := filepath.Join(root, "coze-openapi.yaml")

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
	if result.GeneratedOps < 30 {
		t.Fatalf("expected >=30 generated ops, got %d", result.GeneratedOps)
	}
	if result.GeneratedFiles < 20 {
		t.Fatalf("expected >=20 generated files, got %d", result.GeneratedFiles)
	}

	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "apps", "__init__.py"), "PublishStatus")
	assertFileContains(
		t,
		filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"),
		"from .apps.collaborators import (",
	)
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"), "from .workflows.runs import (")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"), "from .workflows.versions import (")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"), "\"AppCollaborator\"")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"), "\"AddAppCollaboratorResp\"")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "__init__.py"), "\"RemoveAppCollaboratorResp\"")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "coze.py"), "class Coze(object):")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "coze.py"), "class AsyncCoze(object):")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "coze.py"), "def bots(self) -> \"BotsClient\":")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "folders", "__init__.py"), "children_count")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "conversations", "__init__.py"), "def messages")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "datasets", "__init__.py"), "def process")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "bots", "__init__.py"), "def _list_v1")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "bots", "__init__.py"), "use_api_version: int = 1")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "bots", "__init__.py"), "class SimpleBotV1(CozeModel)")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "bots", "__init__.py"), "@field_validator(\"publish_time\", mode=\"before\")")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def cancel")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def create_and_poll")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def _chat_stream_handler")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def build_text(text: str)")
	assertFileNotContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "def _messages_list")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "role: MessageRole")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "content: str")
	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "conversation_id: str")
	assertFileNotContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "chat", "__init__.py"), "role: Optional[MessageRole]")
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
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			QueryFields: []config.OperationField{
				{Name: "status", Type: "str", Required: false, UseValue: true},
			},
			BodyFields:         []string{"name"},
			BodyRequiredFields: []string{"name"},
			IgnoreHeaderParams: true,
			RequestStream:      true,
			DataField:          "data.items",
		},
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if strings.Contains(code, "headers: Optional[Dict[str, str]]") {
		t.Fatalf("did not expect explicit headers signature arg:\n%s", code)
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

func TestRenderOperationMethodStreamWrap(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:     true,
			ResponseType:      "Stream[DemoEvent]",
			AsyncResponseType: "AsyncStream[DemoEvent]",
			ResponseCast:      "None",
			StreamWrap:        true,
			StreamWrapHandler: "handle_demo",
			StreamWrapFields:  []string{"event", "data"},
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "response: IteratorHTTPResponse[str] = self._requester.request(\"post\", url, True, cast=None, headers=headers)") {
		t.Fatalf("expected stream bool request call:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "return Stream(") || !strings.Contains(syncCode, "handler=handle_demo") {
		t.Fatalf("expected wrapped sync stream return:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "resp: AsyncIteratorHTTPResponse[str] = await self._requester.arequest(\"post\", url, True, cast=None, headers=headers)") {
		t.Fatalf("expected async stream bool request call:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "return AsyncStream(") || !strings.Contains(asyncCode, "raw_response=resp._raw_response") {
		t.Fatalf("expected wrapped async stream return:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodStreamWrapYieldAndSyncVarDefault(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:        true,
			ResponseType:         "Stream[DemoEvent]",
			AsyncResponseType:    "AsyncIterator[DemoEvent]",
			ResponseCast:         "None",
			StreamWrap:           true,
			StreamWrapHandler:    "handle_demo",
			StreamWrapFields:     []string{"event", "data"},
			StreamWrapAsyncYield: true,
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "response: IteratorHTTPResponse[str] = self._requester.request(") {
		t.Fatalf("expected sync stream response var to use default name:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "response._raw_response") || !strings.Contains(syncCode, "response.data") {
		t.Fatalf("expected default sync stream var to be reused in return:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "async for item in AsyncStream(") || !strings.Contains(asyncCode, "yield item") {
		t.Fatalf("expected async stream-wrap yield mode:\n%s", asyncCode)
	}
	if strings.Contains(asyncCode, "return AsyncStream(") {
		t.Fatalf("did not expect direct AsyncStream return in yield mode:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodStreamWrapCompactSyncReturn(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:               true,
			ResponseType:                "Stream[DemoEvent]",
			AsyncResponseType:           "AsyncIterator[DemoEvent]",
			ResponseCast:                "None",
			StreamWrap:                  true,
			StreamWrapHandler:           "handle_demo",
			StreamWrapFields:            []string{"event", "data"},
			StreamWrapCompactSyncReturn: true,
		},
	}
	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "return Stream(response._raw_response, response.data, fields=[\"event\", \"data\"], handler=handle_demo)") {
		t.Fatalf("expected compact sync stream return line:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if strings.Contains(asyncCode, "return AsyncStream(resp.data, fields=[\"event\", \"data\"], handler=handle_demo, raw_response=resp._raw_response)") {
		t.Fatalf("did not expect compact async stream return line:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "return AsyncStream(\n            resp.data,\n") {
		t.Fatalf("expected multiline async stream return:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodForceMultilineRequestCallAsyncOnly(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo",
		Method: "post",
		RequestBodySchema: &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
		},
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyFields:                     []string{"name"},
			BodyRequiredFields:             []string{"name"},
			ForceMultilineRequestCallAsync: true,
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if strings.Contains(syncCode, "return self._requester.request(\n") {
		t.Fatalf("sync request call should not be forced multiline by async-only option:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "return await self._requester.arequest(\n") {
		t.Fatalf("async request call should be multiline when async-only option is enabled:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodHeadersExpr(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo",
		Method: "post",
		RequestBodySchema: &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
		},
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyBuilder:        "raw",
			BodyFields:         []string{"name"},
			IgnoreHeaderParams: true,
			HeadersExpr:        "{\"Agw-Js-Conv\": \"str\"}",
		},
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(code, "headers = {\"Agw-Js-Conv\": \"str\"}") {
		t.Fatalf("expected fixed headers assignment via headers_expr:\n%s", code)
	}
	if !strings.Contains(code, "headers=headers") {
		t.Fatalf("expected request call to include headers arg:\n%s", code)
	}
	if strings.Contains(code, "headers: Optional[Dict[str, str]]") {
		t.Fatalf("did not expect explicit headers signature arg:\n%s", code)
	}
}

func TestRenderOperationMethodPaginationRequestArg(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/list",
		Method: "post",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "list_items",
		Details:     details,
		Mapping: &config.OperationMapping{
			Pagination:              "number",
			PaginationDataClass:     "DemoListData",
			PaginationItemType:      "DemoItem",
			PaginationItemsField:    "items",
			PaginationTotalField:    "total",
			PaginationPageNumField:  "page",
			PaginationPageSizeField: "size",
			PaginationRequestArg:    "json",
			ParamAliases: map[string]string{
				"page": "page_num",
				"size": "page_size",
			},
			QueryFields: []config.OperationField{
				{Name: "page", Type: "int", Required: true, Default: "1"},
				{Name: "size", Type: "int", Required: true, Default: "10"},
			},
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "json=") {
		t.Fatalf("expected sync pagination request to use json arg:\n%s", syncCode)
	}
	if strings.Contains(syncCode, "params={") {
		t.Fatalf("did not expect sync pagination request to use params arg:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "json=") {
		t.Fatalf("expected async pagination request to use json arg:\n%s", asyncCode)
	}
	if strings.Contains(asyncCode, "params={") {
		t.Fatalf("did not expect async pagination request to use params arg:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodPaginationPreBodyCode(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/list",
		Method: "post",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "list_items",
		Details:     details,
		Mapping: &config.OperationMapping{
			Pagination:              "number",
			PaginationDataClass:     "DemoListData",
			PaginationItemType:      "DemoItem",
			PaginationItemsField:    "items",
			PaginationTotalField:    "total",
			PaginationPageNumField:  "page",
			PaginationPageSizeField: "size",
			QueryFields: []config.OperationField{
				{Name: "page", Type: "int", Required: true, Default: "1"},
				{Name: "size", Type: "int", Required: true, Default: "10"},
			},
			PreBodyCode: []string{
				"warnings.warn(\"deprecated\")",
			},
		},
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(code, "warnings.warn(\"deprecated\")") {
		t.Fatalf("expected pagination method to include pre_body_code:\n%s", code)
	}
}

func TestMethodBlockOrderingHelpers(t *testing.T) {
	if got := pygen.DetectMethodBlockName(`
@overload
def _create(self) -> None:
    ...
`); got != "_create" {
		t.Fatalf("unexpected detected method name: %q", got)
	}

	blocks := []pygen.ClassMethodBlock{
		{Name: "run_histories", Content: "run_histories", IsChild: true},
		{Name: "messages", Content: "messages"},
		{Name: "create", Content: "create"},
		{Name: "list", Content: "list"},
		{Name: "clone", Content: "clone"},
		{Name: "stream", Content: "stream"},
		{Name: "retrieve", Content: "retrieve"},
	}
	ordered := pygen.OrderClassMethodBlocks(blocks)
	got := []string{
		ordered[0].Name,
		ordered[1].Name,
		ordered[2].Name,
		ordered[3].Name,
		ordered[4].Name,
		ordered[5].Name,
		ordered[6].Name,
	}
	expected := []string{"run_histories", "stream", "create", "clone", "retrieve", "list", "messages"}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("unexpected ordered[%d]: got=%q want=%q", i, got[i], expected[i])
		}
	}

	indented := pygen.IndentCodeBlock("def run(self):\n    return 1\n", 1)
	if !strings.HasPrefix(indented, "    def run") {
		t.Fatalf("unexpected indented block:\n%s", indented)
	}
}

func TestDeduplicateBindingsRenamesConflicts(t *testing.T) {
	bindings := []pygen.OperationBinding{
		{
			PackageName: "demo",
			MethodName:  "create",
			Mapping:     &config.OperationMapping{},
		},
		{
			PackageName: "demo",
			MethodName:  "create",
			Mapping:     &config.OperationMapping{},
		},
	}
	got := pygen.DeduplicateBindings(bindings)
	if got[0].MethodName != "create" {
		t.Fatalf("unexpected first method name: %q", got[0].MethodName)
	}
	if got[1].MethodName != "create_2" {
		t.Fatalf("expected duplicate method to be renamed with suffix, got %q", got[1].MethodName)
	}
}

func TestDeduplicateBindingsRenamesConflictsByClientSide(t *testing.T) {
	bindings := []pygen.OperationBinding{
		{
			PackageName: "demo",
			MethodName:  "create",
			Mapping:     nil,
		},
		{
			PackageName: "demo",
			MethodName:  "create",
			Mapping:     nil,
		},
	}
	got := pygen.DeduplicateBindings(bindings)
	if got[0].MethodName != "create" {
		t.Fatalf("unexpected first method name: %q", got[0].MethodName)
	}
	if got[1].MethodName != "create_2" {
		t.Fatalf("expected duplicate sync method to be renamed with suffix, got %q", got[1].MethodName)
	}
}

func TestGeneratePythonMappingGeneratesSyncAndAsyncByDefault(t *testing.T) {
	out := t.TempDir()
	cfg := testConfig(out)
	cfg.API.GenerateOnlyMapped = true
	cfg.API.OperationMappings = []config.OperationMapping{
		{
			Path:       "/v3/chat",
			Method:     "post",
			SDKMethods: []string{"chat.create"},
		},
	}

	_, err := GeneratePython(cfg, mustParseSwagger(t))
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}

	chatModule := readFile(t, filepath.Join(out, "cozepy", "chat", "__init__.py"))
	if !strings.Contains(chatModule, "class ChatClient(object):") || !strings.Contains(chatModule, "\n    def create(") {
		t.Fatalf("expected sync create method for mapping:\n%s", chatModule)
	}
	if !strings.Contains(chatModule, "class AsyncChatClient(object):") || !strings.Contains(chatModule, "async def create(") {
		t.Fatalf("expected async create method for mapping:\n%s", chatModule)
	}
}

func TestRenderOperationMethodReturnAndAsyncKwargsOptions(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo",
		Method: "get",
	}

	withReturn := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "list_items",
		Details:     details,
		Mapping:     &config.OperationMapping{},
	}, false)
	if !strings.Contains(withReturn, "->") {
		t.Fatalf("expected return annotation in method signature:\n%s", withReturn)
	}

	asyncCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "async_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			AsyncIncludeKwargs: true,
		},
	}, true)
	if !strings.Contains(asyncCode, "**kwargs") {
		t.Fatalf("expected async kwargs passthrough in signature:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodKwargsOnlySignatureForEmptyBodySchema(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:              "/v1/users/me",
		Method:            "get",
		RequestBody:       &openapi.RequestBody{},
		RequestBodySchema: &openapi.Schema{Type: "object"},
	}
	code := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "users",
		MethodName:  "me",
		Details:     details,
		Mapping: &config.OperationMapping{
			ResponseType: "User",
			ResponseCast: "User",
		},
	}, false)
	if !strings.Contains(code, "def me(self, **kwargs)") {
		t.Fatalf("expected kwargs-only method signature without bare '*':\n%s", code)
	}
	if strings.Contains(code, "self, *, **kwargs") {
		t.Fatalf("unexpected invalid kwargs-only signature:\n%s", code)
	}
}

func TestRenderOperationMethodSignatureThresholdByNonKwargs(t *testing.T) {
	doc := mustParseSwagger(t)
	baseBinding := pygen.OperationBinding{
		PackageName: "demo",
		Mapping:     &config.OperationMapping{},
	}

	twoArgsCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: baseBinding.PackageName,
		MethodName:  "list_two",
		Mapping:     baseBinding.Mapping,
		Details: openapi.OperationDetails{
			Path:   "/v1/demo",
			Method: "get",
			QueryParameters: []openapi.ParameterSpec{
				{Name: "a", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "b", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
			},
		},
	}, false)
	if !strings.Contains(twoArgsCode, "def list_two(self, *, a: str, b: str, **kwargs)") {
		t.Fatalf("expected compact signature when non-kwargs args <= 2:\n%s", twoArgsCode)
	}
	if strings.Contains(twoArgsCode, "def list_two(\n") {
		t.Fatalf("did not expect multiline signature when non-kwargs args <= 2:\n%s", twoArgsCode)
	}

	threeArgsCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: baseBinding.PackageName,
		MethodName:  "list_three",
		Mapping:     baseBinding.Mapping,
		Details: openapi.OperationDetails{
			Path:   "/v1/demo",
			Method: "get",
			QueryParameters: []openapi.ParameterSpec{
				{Name: "a", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "b", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "c", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
			},
		},
	}, false)
	if !strings.Contains(threeArgsCode, "def list_three(\n") {
		t.Fatalf("expected multiline signature when non-kwargs args > 2:\n%s", threeArgsCode)
	}
	if !strings.Contains(threeArgsCode, "        **kwargs,\n") {
		t.Fatalf("expected kwargs preserved in multiline signature:\n%s", threeArgsCode)
	}
}

func TestRenderOperationMethodAsyncSignatureThresholdByNonKwargs(t *testing.T) {
	doc := mustParseSwagger(t)
	baseBinding := pygen.OperationBinding{
		PackageName: "demo",
		Mapping:     &config.OperationMapping{},
	}

	twoArgsCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: baseBinding.PackageName,
		MethodName:  "list_two",
		Mapping:     baseBinding.Mapping,
		Details: openapi.OperationDetails{
			Path:   "/v1/demo",
			Method: "get",
			QueryParameters: []openapi.ParameterSpec{
				{Name: "a", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "b", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
			},
		},
	}, true)
	if !strings.Contains(twoArgsCode, "async def list_two(self, *, a: str, b: str, **kwargs)") {
		t.Fatalf("expected compact async signature when non-kwargs args <= 2:\n%s", twoArgsCode)
	}
	if strings.Contains(twoArgsCode, "async def list_two(\n") {
		t.Fatalf("did not expect multiline async signature when non-kwargs args <= 2:\n%s", twoArgsCode)
	}

	threeArgsCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: baseBinding.PackageName,
		MethodName:  "list_three",
		Mapping:     baseBinding.Mapping,
		Details: openapi.OperationDetails{
			Path:   "/v1/demo",
			Method: "get",
			QueryParameters: []openapi.ParameterSpec{
				{Name: "a", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "b", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
				{Name: "c", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
			},
		},
	}, true)
	if !strings.Contains(threeArgsCode, "async def list_three(\n") {
		t.Fatalf("expected multiline async signature when non-kwargs args > 2:\n%s", threeArgsCode)
	}
	if !strings.Contains(threeArgsCode, "        **kwargs,\n") {
		t.Fatalf("expected kwargs preserved in multiline async signature:\n%s", threeArgsCode)
	}
}

func TestRenderOperationMethodAsyncStreamMethodDefaultsToYield(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/workflows/chat",
		Method: "post",
	}
	mapping := &config.OperationMapping{
		RequestStream: true,
		StreamWrap:    true,
		StreamWrapFields: []string{
			"event",
			"data",
		},
		ResponseType:      "Stream[ChatEvent]",
		AsyncResponseType: "AsyncIterator[ChatEvent]",
		ResponseCast:      "None",
		BodyBuilder:       "remove_none_values",
		BodyFields:        []string{"workflow_id"},
		BodyRequiredFields: []string{
			"workflow_id",
		},
		ArgTypes: map[string]string{
			"workflow_id": "str",
		},
	}

	streamCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_chat",
		MethodName:  "stream",
		Details:     details,
		Mapping:     mapping,
	}, true)
	if !strings.Contains(streamCode, "async for item in AsyncStream(") || !strings.Contains(streamCode, "yield item") {
		t.Fatalf("expected async stream method to yield async stream items by default:\n%s", streamCode)
	}
	if strings.Contains(streamCode, "return AsyncStream(") {
		t.Fatalf("did not expect async stream method to return AsyncStream directly:\n%s", streamCode)
	}

	createCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_chat",
		MethodName:  "_create",
		Details:     details,
		Mapping:     mapping,
	}, true)
	if !strings.Contains(createCode, "return AsyncStream(") {
		t.Fatalf("expected non-stream async method to keep returning AsyncStream:\n%s", createCode)
	}
	if strings.Contains(createCode, "async for item in AsyncStream(") {
		t.Fatalf("did not expect non-stream async method to auto-yield async stream:\n%s", createCode)
	}
}

func TestRenderOperationMethodPaginationRequestOrder(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/apps",
		Method: "get",
		QueryParameters: []openapi.ParameterSpec{
			{Name: "workspace_id", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
			{Name: "page_size", In: "query", Required: true, Schema: &openapi.Schema{Type: "integer"}},
			{Name: "page_num", In: "query", Required: true, Schema: &openapi.Schema{Type: "integer"}},
		},
	}

	codeHeadersFirst := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "apps",
		MethodName:  "list",
		Details:     details,
		Mapping: &config.OperationMapping{
			Pagination:          "number",
			PaginationDataClass: "_PrivateListAppsData",
			PaginationItemType:  "App",
		},
	}, false)
	idxHeaders := strings.Index(codeHeadersFirst, "headers=headers")
	idxParams := strings.Index(codeHeadersFirst, "params=")
	idxCast := strings.Index(codeHeadersFirst, "cast=_PrivateListAppsData")
	if idxHeaders == -1 || idxParams == -1 || idxCast == -1 || idxHeaders > idxParams || idxParams > idxCast {
		t.Fatalf("expected pagination request to keep fixed order headers -> params -> cast:\n%s", codeHeadersFirst)
	}
}

func TestRenderOperationMethodPreBodyCode(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/bot/publish",
		Method: "post",
	}
	code := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "bots",
		MethodName:  "publish",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyFields: []string{"bot_id", "connector_ids"},
			ArgTypes: map[string]string{
				"bot_id":        "str",
				"connector_ids": "List[str]",
			},
			PreBodyCode: []string{
				"if not connector_ids:\n    connector_ids = [\"1024\"]",
			},
		},
	}, false)
	if !strings.Contains(code, "if not connector_ids:") || !strings.Contains(code, "connector_ids = [\"1024\"]") {
		t.Fatalf("expected pre body code block in method:\n%s", code)
	}
	if strings.Index(code, "if not connector_ids:") > strings.Index(code, "body =") {
		t.Fatalf("expected pre body code to appear before body assignment:\n%s", code)
	}
	if strings.Contains(code, "body:") {
		t.Fatalf("expected body assignment without explicit annotation:\n%s", code)
	}
}

func TestRenderOperationMethodBodyBuilderRawUsesBodyDirectly(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/workflow/run",
		Method: "post",
	}
	code := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_runs",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyBuilder:        "raw",
			BodyFields:         []string{"workflow_id", "parameters"},
			BodyRequiredFields: []string{"workflow_id"},
			ArgTypes: map[string]string{
				"workflow_id": "str",
				"parameters":  "Dict[str, Any]",
			},
		},
	}, false)
	if !strings.Contains(code, "body = {") {
		t.Fatalf("expected raw body map assignment:\n%s", code)
	}
	if !strings.Contains(code, "body=body") {
		t.Fatalf("expected request call to pass raw body directly:\n%s", code)
	}
}

func TestRenderOperationMethodSingleItemMapsUseMultilineFormat(t *testing.T) {
	doc := mustParseSwagger(t)
	queryDetails := openapi.OperationDetails{
		Path:   "/v1/bot/get_online_info",
		Method: "get",
		QueryParameters: []openapi.ParameterSpec{
			{Name: "bot_id", In: "query", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
	}
	queryCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "bots",
		MethodName:  "_retrieve_v1",
		Details:     queryDetails,
		Mapping: &config.OperationMapping{
			QueryBuilder: "raw",
		},
	}, false)
	if strings.Contains(queryCode, "params = {\"bot_id\": bot_id}") {
		t.Fatalf("expected single-item query map to be multiline:\n%s", queryCode)
	}
	if !strings.Contains(queryCode, "params = {\n            \"bot_id\": bot_id,\n        }") {
		t.Fatalf("expected multiline single-item query map rendering:\n%s", queryCode)
	}

	bodyDetails := openapi.OperationDetails{
		Path:   "/v1/conversations/{conversation_id}",
		Method: "put",
		PathParameters: []openapi.ParameterSpec{
			{Name: "conversation_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
	}
	bodyCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "conversations",
		MethodName:  "update",
		Details:     bodyDetails,
		Mapping: &config.OperationMapping{
			BodyBuilder: "raw",
			BodyFields:  []string{"name"},
			ArgTypes: map[string]string{
				"name": "str",
			},
			BodyRequiredFields: []string{"__none__"},
		},
	}, false)
	if strings.Contains(bodyCode, "body = {\"name\": name}") {
		t.Fatalf("expected single-item body map to be multiline:\n%s", bodyCode)
	}
	if !strings.Contains(bodyCode, "body = {\n            \"name\": name,\n        }") {
		t.Fatalf("expected multiline single-item body map rendering:\n%s", bodyCode)
	}
}

func TestRenderOperationMethodPaginationQueryBuilderUsesDefaultTokenInit(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/workflows/{workflow_id}/versions",
		Method: "get",
		PathParameters: []openapi.ParameterSpec{
			{Name: "workflow_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		QueryParameters: []openapi.ParameterSpec{
			{Name: "publish_status", In: "query", Required: false, Schema: &openapi.Schema{Type: "string"}},
			{Name: "page_size", In: "query", Required: false, Schema: &openapi.Schema{Type: "integer"}},
			{Name: "page_token", In: "query", Required: false, Schema: &openapi.Schema{Type: "string"}},
		},
	}
	mapping := &config.OperationMapping{
		QueryBuilder:             "dump_exclude_none",
		Pagination:               "token",
		PaginationDataClass:      "_PrivateListWorkflowVersionData",
		PaginationItemType:       "WorkflowVersionInfo",
		PaginationPageTokenField: "page_token",
		PaginationPageSizeField:  "page_size",
	}

	syncCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_versions",
		MethodName:  "list",
		Details:     details,
		Mapping:     mapping,
	}, false)
	if !strings.Contains(syncCode, "make_request(\n                \"GET\",") {
		t.Fatalf("expected pagination request to use operation http method in sync request:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "params=dump_exclude_none(") {
		t.Fatalf("expected sync query builder override:\n%s", syncCode)
	}
	if strings.Contains(syncCode, "params = dump_exclude_none(") {
		t.Fatalf("expected sync token pagination params to be inlined:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "page_token=page_token or \"\",") {
		t.Fatalf("expected default pagination token init expr in sync code:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_versions",
		MethodName:  "list",
		Details:     details,
		Mapping:     mapping,
	}, true)
	if !strings.Contains(asyncCode, "amake_request(\n                \"GET\",") {
		t.Fatalf("expected pagination request to use operation http method in async request:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "params=dump_exclude_none(") {
		t.Fatalf("expected async query builder to follow query_builder:\n%s", asyncCode)
	}
	if strings.Contains(asyncCode, "params = dump_exclude_none(") {
		t.Fatalf("expected async token pagination params to be inlined:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "page_token=page_token or \"\",") {
		t.Fatalf("expected default pagination token init expr in async code:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodResponseUnwrapListFirst(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/workflows/{workflow_id}/run_histories/{execute_id}",
		Method: "get",
		PathParameters: []openapi.ParameterSpec{
			{Name: "workflow_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
			{Name: "execute_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_runs_run_histories",
		MethodName:  "retrieve",
		Details:     details,
		Mapping: &config.OperationMapping{
			ResponseType:            "WorkflowRunHistory",
			ResponseCast:            "ListResponse[WorkflowRunHistory]",
			ResponseUnwrapListFirst: true,
		},
	}, false)
	if !strings.Contains(syncCode, "res = self._requester.request(") {
		t.Fatalf("expected response unwrap request assign in sync code:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "data = res.data[0]") || !strings.Contains(syncCode, "data._raw_response = res._raw_response") {
		t.Fatalf("expected unwrap-first-list response handling in sync code:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, pygen.OperationBinding{
		PackageName: "workflows_runs_run_histories",
		MethodName:  "retrieve",
		Details:     details,
		Mapping: &config.OperationMapping{
			ResponseType:            "WorkflowRunHistory",
			ResponseCast:            "ListResponse[WorkflowRunHistory]",
			ResponseUnwrapListFirst: true,
		},
	}, true)
	if !strings.Contains(asyncCode, "res = await self._requester.arequest(") {
		t.Fatalf("expected response unwrap request assign in async code:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "data = res.data[0]") || !strings.Contains(asyncCode, "data._raw_response = res._raw_response") {
		t.Fatalf("expected unwrap-first-list response handling in async code:\n%s", asyncCode)
	}
}

func TestLinesFromCommentOverrideKeepsLeadingSpaces(t *testing.T) {
	lines := pygen.LinesFromCommentOverride([]string{
		"  success: Execution succeeded.",
		"  running: Execution in progress.",
	})
	if len(lines) != 2 {
		t.Fatalf("unexpected comment lines length: %d", len(lines))
	}
	if lines[0] != "  success: Execution succeeded." {
		t.Fatalf("expected leading spaces to be preserved, got %q", lines[0])
	}
	if lines[1] != "  running: Execution in progress." {
		t.Fatalf("expected leading spaces to be preserved, got %q", lines[1])
	}
}

func TestRenderOperationMethodArgDefaultsAsyncOverride(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo",
		Method: "get",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "list_items",
		Details:     details,
		Mapping: &config.OperationMapping{
			QueryFields: []config.OperationField{
				{Name: "page_size", Type: "int", Required: true, Default: "10"},
			},
			ArgDefaultsAsync: map[string]string{
				"page_size": "100",
			},
		},
	}

	syncCode := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "page_size: int = 10") {
		t.Fatalf("expected sync default page_size=10:\n%s", syncCode)
	}

	asyncCode := pygen.RenderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "page_size: int = 100") {
		t.Fatalf("expected async override page_size=100:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodFilesBeforeBody(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/{group_id}/features",
		Method: "post",
		PathParameters: []openapi.ParameterSpec{
			{Name: "group_id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		RequestBodySchema: &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
				"file": {Type: "string"},
			},
			Required: []string{"name", "file"},
		},
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyBuilder:      "remove_none_values",
			BodyFields:       []string{"name"},
			FilesFields:      []string{"file"},
			FilesFieldValues: map[string]string{"file": "_try_fix_file(file)"},
			FilesBeforeBody:  true,
			ArgTypes: map[string]string{
				"group_id": "str",
				"name":     "str",
				"file":     "FileTypes",
			},
		},
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	headersIdx := strings.Index(code, "headers: Optional[dict] = kwargs.get(\"headers\")")
	filesIdx := strings.Index(code, "files = {\"file\": _try_fix_file(file)}")
	bodyIdx := strings.Index(code, "body = remove_none_values(")
	if headersIdx < 0 || filesIdx < 0 || bodyIdx < 0 || !(headersIdx < filesIdx && filesIdx < bodyIdx) {
		t.Fatalf("expected files assignment before body assignment:\n%s", code)
	}
}

func TestRenderOperationMethodPreDocstringCode(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:    "/v1/demo",
		Method:  "post",
		Summary: "Upload demo item.",
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			PreDocstringCode: []string{
				`warnings.warn("deprecated", DeprecationWarning, stacklevel=2)`,
			},
		},
	}
	code := pygen.RenderOperationMethodWithComments(doc, binding, false)
	warnIdx := strings.Index(code, `warnings.warn("deprecated", DeprecationWarning, stacklevel=2)`)
	docIdx := strings.Index(code, "Upload demo item.")
	urlIdx := strings.Index(code, `url = f"{self._base_url}/v1/demo"`)
	if warnIdx < 0 || docIdx < 0 || urlIdx < 0 || !(warnIdx < docIdx && docIdx < urlIdx) {
		t.Fatalf("expected pre_docstring_code to be emitted before docstring and url:\n%s", code)
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

func mustLoadRealConfigAndSwagger(t *testing.T) (*config.Config, *openapi.Document) {
	t.Helper()
	root := filepath.Clean(filepath.Join("..", ".."))
	cfgPath := filepath.Join(root, "config", "generator.yaml")
	swaggerPath := filepath.Join(root, "coze-openapi.yaml")

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load(%q) error = %v", cfgPath, err)
	}
	doc, err := openapi.Load(swaggerPath)
	if err != nil {
		t.Fatalf("openapi.Load(%q) error = %v", swaggerPath, err)
	}
	return cfg, doc
}

func assertFileContains(t *testing.T, pathName string, expected string) {
	t.Helper()
	content := readFile(t, pathName)
	if !strings.Contains(content, expected) {
		t.Fatalf("expected %q in %s, got:\n%s", expected, pathName, content)
	}
}

func assertFileNotContains(t *testing.T, pathName string, unexpected string) {
	t.Helper()
	content := readFile(t, pathName)
	if strings.Contains(content, unexpected) {
		t.Fatalf("did not expect %q in %s, got:\n%s", unexpected, pathName, content)
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

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
	if result.GeneratedOps < 30 {
		t.Fatalf("expected >=30 generated ops, got %d", result.GeneratedOps)
	}
	if result.GeneratedFiles < 20 {
		t.Fatalf("expected >=20 generated files, got %d", result.GeneratedFiles)
	}

	assertFileContains(t, filepath.Join(cfg.OutputSDK, "cozepy", "apps", "__init__.py"), "PublishStatus")
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

func TestRenderOperationMethodStreamWrapAndKeywords(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := operationBinding{
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
			CastKeyword:       true,
			StreamKeyword:     true,
		},
	}

	syncCode := renderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "response: IteratorHTTPResponse[str] = self._requester.request(\"post\", url, stream=True, cast=None, headers=headers)") {
		t.Fatalf("expected keyword stream/cast request call:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "return Stream(") || !strings.Contains(syncCode, "handler=handle_demo") {
		t.Fatalf("expected wrapped sync stream return:\n%s", syncCode)
	}

	asyncCode := renderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "resp: AsyncIteratorHTTPResponse[str] = await self._requester.arequest(\"post\", url, stream=True, cast=None, headers=headers)") {
		t.Fatalf("expected keyword async request call:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "return AsyncStream(") || !strings.Contains(asyncCode, "raw_response=resp._raw_response") {
		t.Fatalf("expected wrapped async stream return:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodStreamWrapYieldAndSyncVarOverride(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := operationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:                 true,
			ResponseType:                  "Stream[DemoEvent]",
			AsyncResponseType:             "AsyncIterator[DemoEvent]",
			ResponseCast:                  "None",
			StreamWrap:                    true,
			StreamWrapHandler:             "handle_demo",
			StreamWrapFields:              []string{"event", "data"},
			StreamWrapAsyncYield:          true,
			StreamWrapSyncResponseVar:     "resp",
			DisableHeadersArg:             true,
			ForceMultilineRequestCall:     false,
			ForceMultilineRequestCallSync: false,
		},
	}

	syncCode := renderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "resp: IteratorHTTPResponse[str] = self._requester.request(") {
		t.Fatalf("expected sync stream response var override:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "resp._raw_response") || !strings.Contains(syncCode, "resp.data") {
		t.Fatalf("expected sync stream var to be reused in return:\n%s", syncCode)
	}

	asyncCode := renderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "async for item in AsyncStream(") || !strings.Contains(asyncCode, "yield item") {
		t.Fatalf("expected async stream-wrap yield mode:\n%s", asyncCode)
	}
	if strings.Contains(asyncCode, "return AsyncStream(") {
		t.Fatalf("did not expect direct AsyncStream return in yield mode:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodStreamWrapCompactAsyncReturn(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := operationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:                true,
			ResponseType:                 "Stream[DemoEvent]",
			AsyncResponseType:            "AsyncIterator[DemoEvent]",
			ResponseCast:                 "None",
			StreamWrap:                   true,
			StreamWrapHandler:            "handle_demo",
			StreamWrapFields:             []string{"event", "data"},
			StreamWrapCompactSyncReturn:  true,
			StreamWrapCompactAsyncReturn: true,
			DisableHeadersArg:            true,
		},
	}
	syncCode := renderOperationMethod(doc, binding, false)
	if !strings.Contains(syncCode, "return Stream(response._raw_response, response.data, fields=[\"event\", \"data\"], handler=handle_demo)") {
		t.Fatalf("expected compact sync stream return line:\n%s", syncCode)
	}

	asyncCode := renderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "return AsyncStream(resp.data, fields=[\"event\", \"data\"], handler=handle_demo, raw_response=resp._raw_response)") {
		t.Fatalf("expected compact async stream return line:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodStreamWrapBlankLineBeforeAsyncReturn(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/stream",
		Method: "post",
	}
	binding := operationBinding{
		PackageName: "demo",
		MethodName:  "stream_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			RequestStream:                  true,
			ResponseType:                   "Stream[DemoEvent]",
			AsyncResponseType:              "AsyncIterator[DemoEvent]",
			ResponseCast:                   "None",
			StreamWrap:                     true,
			StreamWrapCompactAsyncReturn:   true,
			StreamWrapBlankLineBeforeAsync: true,
			DisableHeadersArg:              true,
		},
	}

	asyncCode := renderOperationMethod(doc, binding, true)
	if !strings.Contains(asyncCode, "resp: AsyncIteratorHTTPResponse[str] = await self._requester.arequest(\"post\", url, True, None)\n\n        return AsyncStream(") {
		t.Fatalf("expected a blank line before async stream return:\n%s", asyncCode)
	}
}

func TestMethodBlockOrderingHelpers(t *testing.T) {
	if got := detectMethodBlockName(`
@overload
def _create(self) -> None:
    ...
`); got != "_create" {
		t.Fatalf("unexpected detected method name: %q", got)
	}

	blocks := []classMethodBlock{
		{Name: "messages", Content: "messages"},
		{Name: "create", Content: "create"},
		{Name: "list", Content: "list"},
		{Name: "retrieve", Content: "retrieve"},
	}
	ordered := orderClassMethodBlocks(blocks, []string{"create", "retrieve", "messages"})
	got := []string{ordered[0].Name, ordered[1].Name, ordered[2].Name, ordered[3].Name}
	expected := []string{"create", "retrieve", "messages", "list"}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("unexpected ordered[%d]: got=%q want=%q", i, got[i], expected[i])
		}
	}

	indented := indentCodeBlock("def run(self):\n    return 1\n", 1)
	if !strings.HasPrefix(indented, "    def run") {
		t.Fatalf("unexpected indented block:\n%s", indented)
	}

	orderedChildren := orderChildClients(
		[]config.ChildClient{
			{Attribute: "rooms"},
			{Attribute: "speech"},
			{Attribute: "voices"},
		},
		[]string{"voices", "rooms"},
	)
	if len(orderedChildren) != 3 || orderedChildren[0].Attribute != "voices" || orderedChildren[1].Attribute != "rooms" {
		t.Fatalf("unexpected child order: %+v", orderedChildren)
	}
}

func TestGeneratePythonAsyncOnlyMapping(t *testing.T) {
	out := t.TempDir()
	cfg := testConfig(out)
	cfg.API.GenerateOnlyMapped = true
	cfg.API.OperationMappings = []config.OperationMapping{
		{
			Path:       "/v3/chat",
			Method:     "post",
			SDKMethods: []string{"chat.create"},
			AsyncOnly:  true,
		},
	}

	_, err := GeneratePython(cfg, mustParseSwagger(t))
	if err != nil {
		t.Fatalf("GeneratePython() error = %v", err)
	}

	chatModule := readFile(t, filepath.Join(out, "cozepy", "chat", "__init__.py"))
	if strings.Contains(chatModule, "class ChatClient(object):\n    def __init__") && strings.Contains(chatModule, "\n    def create(") {
		t.Fatalf("did not expect sync create method for async_only mapping:\n%s", chatModule)
	}
	if !strings.Contains(chatModule, "class AsyncChatClient(object):") || !strings.Contains(chatModule, "async def create(") {
		t.Fatalf("expected async create method for async_only mapping:\n%s", chatModule)
	}
}

func TestRenderOperationMethodReturnAndAsyncKwargsOptions(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo",
		Method: "get",
	}

	noReturn := renderOperationMethod(doc, operationBinding{
		PackageName: "demo",
		MethodName:  "list_items",
		Details:     details,
		Mapping: &config.OperationMapping{
			OmitReturnType: true,
		},
	}, false)
	if strings.Contains(noReturn, "->") {
		t.Fatalf("did not expect return annotation when omit_return_type=true:\n%s", noReturn)
	}

	asyncCode := renderOperationMethod(doc, operationBinding{
		PackageName: "demo",
		MethodName:  "async_call",
		Details:     details,
		Mapping: &config.OperationMapping{
			DisableHeadersArg:  true,
			AsyncIncludeKwargs: true,
		},
	}, true)
	if !strings.Contains(asyncCode, "**kwargs") {
		t.Fatalf("expected async kwargs passthrough in signature:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodKwargsOnlySignature(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/users/me",
		Method: "get",
	}
	code := renderOperationMethod(doc, operationBinding{
		PackageName: "users",
		MethodName:  "me",
		Details:     details,
		Mapping: &config.OperationMapping{
			UseKwargsHeaders:   true,
			DisableRequestBody: true,
			ResponseType:       "User",
			ResponseCast:       "User",
		},
	}, false)
	if !strings.Contains(code, "def me(self, **kwargs)") {
		t.Fatalf("expected kwargs-only method signature without bare '*':\n%s", code)
	}
	if strings.Contains(code, "self, *, **kwargs") {
		t.Fatalf("unexpected invalid kwargs-only signature:\n%s", code)
	}
}

func TestRenderOperationMethodDelegateTo(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v3/chat",
		Method: "post",
	}

	syncCode := renderOperationMethod(doc, operationBinding{
		PackageName: "chat",
		MethodName:  "create",
		Details:     details,
		Mapping: &config.OperationMapping{
			DelegateTo:       "_create",
			DelegateCallArgs: []string{"bot_id=bot_id", "stream=False"},
			ArgTypes: map[string]string{
				"bot_id": "str",
			},
			BodyFields: []string{"bot_id"},
		},
	}, false)
	if !strings.Contains(syncCode, "return self._create(") {
		t.Fatalf("expected sync delegate return call:\n%s", syncCode)
	}
	if strings.Contains(syncCode, "self._requester.request") {
		t.Fatalf("did not expect direct requester call for delegated method:\n%s", syncCode)
	}

	asyncCode := renderOperationMethod(doc, operationBinding{
		PackageName: "chat",
		MethodName:  "stream",
		Details:     details,
		Mapping: &config.OperationMapping{
			DelegateTo:            "_create",
			DelegateCallArgs:      []string{"bot_id=bot_id", "stream=True", "**kwargs"},
			AsyncDelegateCallArgs: []string{"bot_id=bot_id", "additional_messages=additional_messages", "stream=True", "**kwargs"},
			DelegateAsyncYield:    true,
			UseKwargsHeaders:      true,
			BodyFields:            []string{"bot_id"},
			ArgTypes: map[string]string{
				"bot_id": "str",
			},
			ResponseType: "AsyncIterator[str]",
		},
	}, true)
	if !strings.Contains(asyncCode, "async for item in await self._create(") || !strings.Contains(asyncCode, "yield item") {
		t.Fatalf("expected async delegate yield wrapper:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "additional_messages=additional_messages,\n            stream=True") {
		t.Fatalf("expected async-specific delegate call arg order:\n%s", asyncCode)
	}
}

func TestRenderOperationMethodPaginationOrderOptions(t *testing.T) {
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

	codeHeadersFirst := renderOperationMethod(doc, operationBinding{
		PackageName: "apps",
		MethodName:  "list",
		Details:     details,
		Mapping: &config.OperationMapping{
			Pagination:                    "number",
			PaginationDataClass:           "_PrivateListAppsData",
			PaginationItemType:            "App",
			UseKwargsHeaders:              true,
			PaginationHeadersBeforeParams: true,
		},
	}, false)
	idxHeaders := strings.Index(codeHeadersFirst, "headers=headers")
	idxParams := strings.Index(codeHeadersFirst, "params=")
	if idxHeaders == -1 || idxParams == -1 || idxHeaders > idxParams {
		t.Fatalf("expected headers before params in pagination request:\n%s", codeHeadersFirst)
	}

	codeCastBeforeHeaders := renderOperationMethod(doc, operationBinding{
		PackageName: "apps",
		MethodName:  "list",
		Details:     details,
		Mapping: &config.OperationMapping{
			Pagination:                  "number",
			PaginationDataClass:         "_PrivateListAppsData",
			PaginationItemType:          "App",
			UseKwargsHeaders:            true,
			PaginationCastBeforeHeaders: true,
		},
	}, false)
	idxCast := strings.Index(codeCastBeforeHeaders, "cast=_PrivateListAppsData")
	idxHeaders = strings.Index(codeCastBeforeHeaders, "headers=headers")
	if idxCast == -1 || idxHeaders == -1 || idxCast > idxHeaders {
		t.Fatalf("expected cast before headers in pagination request:\n%s", codeCastBeforeHeaders)
	}
}

func TestRenderOperationMethodPreBodyCode(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/bot/publish",
		Method: "post",
	}
	code := renderOperationMethod(doc, operationBinding{
		PackageName: "bots",
		MethodName:  "publish",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyFields:     []string{"bot_id", "connector_ids"},
			BodyAnnotation: "Dict[str, Any]",
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
	if strings.Index(code, "if not connector_ids:") > strings.Index(code, "body: Dict[str, Any] =") {
		t.Fatalf("expected pre body code to appear before body assignment:\n%s", code)
	}
	if !strings.Contains(code, "body: Dict[str, Any] =") {
		t.Fatalf("expected body annotation in assignment:\n%s", code)
	}
}

func TestRenderOperationMethodBodyCallExprOverride(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/workflow/run",
		Method: "post",
	}
	code := renderOperationMethod(doc, operationBinding{
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
			BodyCallExpr: "remove_none_values(body)",
		},
	}, false)
	if !strings.Contains(code, "body = {") {
		t.Fatalf("expected raw body map assignment:\n%s", code)
	}
	if !strings.Contains(code, "body=remove_none_values(body)") {
		t.Fatalf("expected body call expression override in request call:\n%s", code)
	}
}

func TestRenderOperationMethodPaginationQueryBuilderAndTokenOverrides(t *testing.T) {
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
		QueryBuilder:                "dump_exclude_none",
		QueryBuilderSync:            "dump_exclude_none",
		QueryBuilderAsync:           "remove_none_values",
		Pagination:                  "token",
		PaginationDataClass:         "_PrivateListWorkflowVersionData",
		PaginationItemType:          "WorkflowVersionInfo",
		PaginationPageTokenField:    "page_token",
		PaginationPageSizeField:     "page_size",
		PaginationInitPageTokenExpr: "page_token or \"\"",
		PaginationHTTPMethod:        "get",
	}

	syncCode := renderOperationMethod(doc, operationBinding{
		PackageName: "workflows_versions",
		MethodName:  "list",
		Details:     details,
		Mapping:     mapping,
	}, false)
	if !strings.Contains(syncCode, "make_request(\n                \"get\",") {
		t.Fatalf("expected custom pagination http method in sync request:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "params=dump_exclude_none(") {
		t.Fatalf("expected sync query builder override:\n%s", syncCode)
	}
	if !strings.Contains(syncCode, "page_token=page_token or \"\"") {
		t.Fatalf("expected custom pagination token init expr in sync code:\n%s", syncCode)
	}

	asyncCode := renderOperationMethod(doc, operationBinding{
		PackageName: "workflows_versions",
		MethodName:  "list",
		Details:     details,
		Mapping:     mapping,
	}, true)
	if !strings.Contains(asyncCode, "amake_request(\n                \"get\",") {
		t.Fatalf("expected custom pagination http method in async request:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "params=remove_none_values(") {
		t.Fatalf("expected async query builder override:\n%s", asyncCode)
	}
	if !strings.Contains(asyncCode, "page_token=page_token or \"\"") {
		t.Fatalf("expected custom pagination token init expr in async code:\n%s", asyncCode)
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

	syncCode := renderOperationMethod(doc, operationBinding{
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

	asyncCode := renderOperationMethod(doc, operationBinding{
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
	lines := linesFromCommentOverride([]string{
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

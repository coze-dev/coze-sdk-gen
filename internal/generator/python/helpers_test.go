package python

import (
	"bytes"
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestNamingHelpers(t *testing.T) {
	if got := NormalizePythonIdentifier("class"); got != "class_" {
		t.Fatalf("NormalizePythonIdentifier(class)=%q", got)
	}
	if got := NormalizePackageName(""); got != "default" {
		t.Fatalf("NormalizePackageName(empty)=%q", got)
	}
	if got := NormalizePackageDir("cozepy/chat/message", "chat"); got != "chat/message" {
		t.Fatalf("NormalizePackageDir=%q", got)
	}
	if got := PackageClassName("chat_message"); got != "ChatMessage" {
		t.Fatalf("PackageClassName=%q", got)
	}
	if got := DefaultMethodName("OpenApiChatCancel", "/v3/chat/cancel", "post"); got != "chat_cancel" {
		t.Fatalf("DefaultMethodName=%q", got)
	}
	if got := NormalizeMethodName("_create"); got != "_create" {
		t.Fatalf("NormalizeMethodName=%q", got)
	}
	if got := NormalizeClassName("open_api_chat_req"); got != "OpenApiChatReq" {
		t.Fatalf("NormalizeClassName=%q", got)
	}
	if got := ToSnake("ChatCancel"); got != "chat_cancel" {
		t.Fatalf("ToSnake=%q", got)
	}
	parts := SplitIdentifier("a-b_c")
	if len(parts) != 3 || parts[0] != "a" || parts[1] != "b" || parts[2] != "c" {
		t.Fatalf("SplitIdentifier=%v", parts)
	}
	if got := CollapseUnderscore("a__b___c"); got != "a_b_c" {
		t.Fatalf("CollapseUnderscore=%q", got)
	}
	if got := EscapeDocstring("a\n\"\"\"b"); got != "a \"b" {
		t.Fatalf("EscapeDocstring=%q", got)
	}
}

func TestSchemaAndTypeHelpers(t *testing.T) {
	userSchema := &openapi.Schema{Type: "object"}
	bodySchema := &openapi.Schema{
		Type: "object",
		Properties: map[string]*openapi.Schema{
			"name": {Type: "string"},
		},
	}
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"User": userSchema,
			},
		},
	}

	schemaRef := &openapi.Schema{Ref: "#/components/schemas/User"}
	if got, ok := SchemaTypeName(doc, schemaRef); !ok || got != "User" {
		t.Fatalf("SchemaTypeName got=%q ok=%v", got, ok)
	}
	if got, ok := SchemaTypeNameWithAliases(doc, schemaRef, map[string]string{"User": "APIUser"}); !ok || got != "APIUser" {
		t.Fatalf("SchemaTypeNameWithAliases got=%q ok=%v", got, ok)
	}
	if got := PythonTypeForSchemaRequiredWithAliases(doc, schemaRef, nil); got != "User" {
		t.Fatalf("PythonTypeForSchemaRequiredWithAliases(ref)=%q", got)
	}
	if got := PythonTypeForSchema(doc, &openapi.Schema{Type: "string"}, false); got != "Optional[str]" {
		t.Fatalf("PythonTypeForSchema(optional str)=%q", got)
	}
	if got := PythonTypeForSchemaRequiredWithAliases(doc, &openapi.Schema{
		Type:  "array",
		Items: &openapi.Schema{Type: "integer"},
	}, nil); got != "List[int]" {
		t.Fatalf("PythonTypeForSchemaRequiredWithAliases(array)=%q", got)
	}
	if got := PythonTypeForSchemaRequired(doc, &openapi.Schema{Type: "boolean"}); got != "bool" {
		t.Fatalf("PythonTypeForSchemaRequired(bool)=%q", got)
	}
	if got := TypeOverride("name", false, "str", map[string]string{"name": "int"}); got != "Optional[int]" {
		t.Fatalf("TypeOverride=%q", got)
	}
	if got := TypeOverride("name", true, "str", map[string]string{"name": "int"}); got != "int" {
		t.Fatalf("TypeOverride(required)=%q", got)
	}
	if got := OperationArgName("app_id", map[string]string{"app_id": "app_id_alias"}); got != "app_id_alias" {
		t.Fatalf("OperationArgName=%q", got)
	}
	if got, _ := ReturnTypeInfo(doc, schemaRef); got != "Dict[str, Any]" {
		t.Fatalf("ReturnTypeInfo=%q", got)
	}
	if got, required := RequestBodyTypeInfo(doc, bodySchema, &openapi.RequestBody{Required: true}); got != "Dict[str, Any]" || !required {
		t.Fatalf("RequestBodyTypeInfo got=%q required=%v", got, required)
	}
	if got := BodyFieldSchema(doc, bodySchema, "name"); got == nil || got.Type != "string" {
		t.Fatalf("BodyFieldSchema(name)=%#v", got)
	}
	if got := BodyFieldSchema(doc, bodySchema, "missing"); got != nil {
		t.Fatalf("BodyFieldSchema(missing)=%#v", got)
	}
}

func TestInferResponseCast(t *testing.T) {
	tests := []struct {
		name        string
		mapping     *config.OperationMapping
		returnType  string
		defaultCast string
		expected    string
	}{
		{
			name: "stream return type uses none cast",
			mapping: &config.OperationMapping{
				ResponseType: "Stream[ChatEvent]",
			},
			returnType: "Stream[ChatEvent]",
			expected:   "None",
		},
		{
			name: "list return type uses list cast",
			mapping: &config.OperationMapping{
				ResponseType: "List[Document]",
			},
			returnType: "List[Document]",
			expected:   "[Document]",
		},
		{
			name: "unwrap list first uses list response cast",
			mapping: &config.OperationMapping{
				ResponseType:            "WorkflowRunHistory",
				ResponseUnwrapListFirst: true,
			},
			returnType: "WorkflowRunHistory",
			expected:   "ListResponse[WorkflowRunHistory]",
		},
		{
			name: "fallback to response type",
			mapping: &config.OperationMapping{
				ResponseType: "User",
			},
			returnType: "User",
			expected:   "User",
		},
		{
			name:        "nil mapping keeps default cast",
			returnType:  "Dict[str, Any]",
			defaultCast: "DemoResp",
			expected:    "DemoResp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferResponseCast(tt.mapping, tt.returnType, tt.defaultCast); got != tt.expected {
				t.Fatalf("inferResponseCast() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestInferPaginationRequestArg(t *testing.T) {
	tests := []struct {
		method string
		want   string
	}{
		{method: "get", want: "params"},
		{method: "head", want: "params"},
		{method: "options", want: "params"},
		{method: "post", want: "json"},
		{method: "put", want: "json"},
		{method: "delete", want: "json"},
		{method: "", want: "json"},
	}

	for _, tt := range tests {
		if got := inferPaginationRequestArg(tt.method); got != tt.want {
			t.Fatalf("inferPaginationRequestArg(%q) = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestCommentAndDocstringHelpers(t *testing.T) {
	buf := &bytes.Buffer{}
	AppendIndentedCode(buf, "    x = 1\n    y = 2\n", 1)
	if got := buf.String(); !strings.Contains(got, "    x = 1\n") || !strings.Contains(got, "    y = 2\n") {
		t.Fatalf("AppendIndentedCode=%q", got)
	}

	EnsureTrailingNewlines(buf, 2)
	if !strings.HasSuffix(buf.String(), "\n\n") {
		t.Fatalf("EnsureTrailingNewlines=%q", buf.String())
	}

	lines := LinesFromCommentOverride([]string{"# hello", " ", "#world "})
	if len(lines) != 2 || lines[0] != " hello" || lines[1] != "world" {
		t.Fatalf("LinesFromCommentOverride=%v", lines)
	}

	if !CodeBlocksHaveLeadingDocstring([]string{"\n\"\"\"doc\"\"\"\n"}) {
		t.Fatal("CodeBlocksHaveLeadingDocstring expected true")
	}
	if CodeBlocksHaveLeadingDocstring([]string{"x = 1"}) {
		t.Fatal("CodeBlocksHaveLeadingDocstring expected false")
	}

	buf.Reset()
	WriteLineComments(buf, 1, []string{"a", "b"})
	if got := buf.String(); got != "    # a\n    # b\n" {
		t.Fatalf("WriteLineComments=%q", got)
	}

	docLines := NormalizedDocstringLines("\n hello \n\nworld\n")
	if len(docLines) != 3 || docLines[0] != " hello " || docLines[2] != "world" {
		t.Fatalf("NormalizedDocstringLines=%v", docLines)
	}

	buf.Reset()
	WriteClassDocstring(buf, 1, "hello", "inline")
	if got := buf.String(); !strings.Contains(got, "\"\"\"hello\"\"\"") {
		t.Fatalf("WriteClassDocstring=%q", got)
	}

	buf.Reset()
	WriteMethodDocstring(buf, 2, "hello\nworld", "block")
	if got := buf.String(); !strings.Contains(got, "        \"\"\"") || !strings.Contains(got, "        world") {
		t.Fatalf("WriteMethodDocstring=%q", got)
	}

	if got := RenderEnumValueLiteral("x"); got != "\"x\"" {
		t.Fatalf("RenderEnumValueLiteral(string)=%q", got)
	}
	if got := RenderEnumValueLiteral(1); got != "1" {
		t.Fatalf("RenderEnumValueLiteral(int)=%q", got)
	}
	if got := RenderEnumValueLiteral(1.5); got != "1.5" {
		t.Fatalf("RenderEnumValueLiteral(float)=%q", got)
	}
	if got := RenderEnumValueLiteral(true); got != "True" {
		t.Fatalf("RenderEnumValueLiteral(bool)=%q", got)
	}

	if got := EnumMemberName("1-a"); got != "VALUE_1_A" {
		t.Fatalf("EnumMemberName=%q", got)
	}

	block := "    @overload\n    def create(self):\n        pass\n"
	if got := DetectMethodBlockName(block); got != "create" {
		t.Fatalf("DetectMethodBlockName=%q", got)
	}
	if got, ok := ParseDefName("async def stream(self):"); !ok || got != "stream" {
		t.Fatalf("ParseDefName got=%q ok=%v", got, ok)
	}
	if !IsDocstringLine("\"\"\"a\"\"\"") || IsDocstringLine("x") {
		t.Fatal("IsDocstringLine mismatch")
	}
	rendered := RenderMethodDocstringLines("hello\nworld", "inline", "    ")
	if len(rendered) < 2 || !strings.HasPrefix(rendered[0], "    \"\"\"hello") {
		t.Fatalf("RenderMethodDocstringLines=%v", rendered)
	}
}

func TestOperationHelpersOrdering(t *testing.T) {
	if got := IndentCodeBlock("x = 1\n", 2); got != "        x = 1\n" {
		t.Fatalf("IndentCodeBlock=%q", got)
	}

	fields := []RenderQueryField{
		{RawName: "cursor"},
		{RawName: "other"},
		{RawName: "limit"},
	}
	paginated := PaginationOrderedFields(fields, "limit", "cursor")
	if len(paginated) != 3 || paginated[0].RawName != "other" || paginated[1].RawName != "limit" || paginated[2].RawName != "cursor" {
		t.Fatalf("PaginationOrderedFields=%v", paginated)
	}

	orderedFields := OrderSignatureQueryFields(
		[]RenderQueryField{
			{RawName: "required_a", Required: true},
			{RawName: "page_size", Required: true, DefaultValue: "20"},
			{RawName: "opt_with_default_inline", Required: false, DefaultValue: "10"},
			{RawName: "opt_without_default", Required: false},
			{RawName: "opt_with_default_override", ArgName: "opt_with_default_override", Required: false},
			{RawName: "cursor", ArgName: "page_num", Required: false},
			{RawName: "required_b", Required: true},
		},
		&config.OperationMapping{
			ArgDefaults: map[string]string{"opt_with_default_override": "1"},
		},
		false,
	)
	if len(orderedFields) != 7 ||
		orderedFields[0].RawName != "required_a" ||
		orderedFields[1].RawName != "required_b" ||
		orderedFields[2].RawName != "opt_without_default" ||
		orderedFields[3].RawName != "opt_with_default_inline" ||
		orderedFields[4].RawName != "opt_with_default_override" ||
		orderedFields[5].RawName != "page_size" ||
		orderedFields[6].RawName != "cursor" {
		t.Fatalf("OrderSignatureQueryFields=%v", orderedFields)
	}
}

func TestOperationHelpersSignatureAndDefaults(t *testing.T) {
	if got := SignatureArgName("**kwargs: Dict[str, Any]"); got != "kwargs" {
		t.Fatalf("SignatureArgName(kwargs)=%q", got)
	}
	if got := SignatureArgName("name: str = \"x\""); got != "name" {
		t.Fatalf("SignatureArgName=%q", got)
	}
	if !IsKwargsSignatureArg(" **kwargs") {
		t.Fatal("IsKwargsSignatureArg expected true")
	}
	normalized := NormalizeSignatureArgs([]string{"a: str", "**kwargs", "b: int"})
	if len(normalized) != 3 || normalized[2] != "**kwargs" {
		t.Fatalf("NormalizeSignatureArgs=%v", normalized)
	}
	unique := OrderedUniqueByPriority([]string{" b ", "a", "b", "c"}, []string{"a", "b"})
	if len(unique) != 3 || unique[0] != "a" || unique[1] != "b" || unique[2] != "c" {
		t.Fatalf("OrderedUniqueByPriority=%v", unique)
	}

	mapping := &config.OperationMapping{
		ArgDefaults:      map[string]string{"x": "1"},
		ArgDefaultsSync:  map[string]string{"x": "2"},
		ArgDefaultsAsync: map[string]string{"x": "3"},
	}
	if got, ok := OperationArgDefault(mapping, "x", "x", false); !ok || got != "2" {
		t.Fatalf("OperationArgDefault(sync) got=%q ok=%v", got, ok)
	}
	if got, ok := OperationArgDefault(mapping, "x", "x", true); !ok || got != "3" {
		t.Fatalf("OperationArgDefault(async) got=%q ok=%v", got, ok)
	}
	if _, ok := OperationArgDefault(nil, "x", "x", false); ok {
		t.Fatal("OperationArgDefault(nil) expected false")
	}

	callArgs := BuildAutoDelegateCallArgs(
		[]string{"a: str", "b: int", "**kwargs"},
		[]string{"stream=True"},
	)
	expected := []string{"a=a", "b=b", "stream=True", "**kwargs"}
	if len(callArgs) != len(expected) {
		t.Fatalf("BuildAutoDelegateCallArgs len=%d args=%v", len(callArgs), callArgs)
	}
	for i := range expected {
		if callArgs[i] != expected[i] {
			t.Fatalf("BuildAutoDelegateCallArgs[%d]=%q want=%q", i, callArgs[i], expected[i])
		}
	}

	buf := &bytes.Buffer{}
	RenderDelegatedCall(buf, "_create", callArgs, true, true)
	if got := buf.String(); !strings.Contains(got, "async for item in await self._create(") || !strings.Contains(got, "stream=True") {
		t.Fatalf("RenderDelegatedCall(async yield)=%q", got)
	}
}

func TestRenderOperationMethodAutoDelegatesToPrivateCreate(t *testing.T) {
	doc := &openapi.Document{}
	details := openapi.OperationDetails{
		Path:   "/v3/chat",
		Method: "post",
	}
	classMethods := map[string]struct{}{
		"_create": {},
	}

	createCode := renderOperationMethodWithContext(
		doc,
		OperationBinding{
			PackageName: "chat",
			MethodName:  "create",
			Details:     details,
			Mapping: &config.OperationMapping{
				BodyFields: []string{"bot_id", "user_id"},
				BodyRequiredFields: []string{
					"bot_id",
					"user_id",
				},
				BodyFixedValues: map[string]string{"stream": "False"},
				ArgTypes: map[string]string{
					"bot_id":  "str",
					"user_id": "str",
				},
				ResponseType: "Chat",
			},
		},
		false,
		"",
		"",
		config.CommentOverrides{},
		classMethods,
	)
	if !strings.Contains(createCode, "return self._create(") || !strings.Contains(createCode, "stream=False") {
		t.Fatalf("expected sync create to auto delegate to _create:\n%s", createCode)
	}
	if strings.Contains(createCode, "self._requester.request(") {
		t.Fatalf("did not expect direct request code for delegated create:\n%s", createCode)
	}

	streamCode := renderOperationMethodWithContext(
		doc,
		OperationBinding{
			PackageName: "chat",
			MethodName:  "stream",
			Details:     details,
			Mapping: &config.OperationMapping{
				BodyFields: []string{"bot_id", "user_id"},
				BodyRequiredFields: []string{
					"bot_id",
					"user_id",
				},
				BodyFixedValues: map[string]string{"stream": "True"},
				ArgTypes: map[string]string{
					"bot_id":  "str",
					"user_id": "str",
				},
				ResponseType: "AsyncIterator[ChatEvent]",
			},
		},
		true,
		"",
		"",
		config.CommentOverrides{},
		classMethods,
	)
	if !strings.Contains(streamCode, "async for item in await self._create(") || !strings.Contains(streamCode, "stream=True") {
		t.Fatalf("expected async stream to auto delegate to _create with yield:\n%s", streamCode)
	}
	if strings.Contains(streamCode, "await self._requester.arequest(") {
		t.Fatalf("did not expect direct request code for delegated stream:\n%s", streamCode)
	}
}

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
	children := []config.ChildClient{
		{Attribute: "c"},
		{Attribute: "a"},
		{Attribute: "b"},
	}
	ordered := OrderChildClients(children, []string{"a", "b"})
	if len(ordered) != 3 || ordered[0].Attribute != "a" || ordered[1].Attribute != "b" || ordered[2].Attribute != "c" {
		t.Fatalf("OrderChildClients=%v", ordered)
	}

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
		[]RenderQueryField{{RawName: "b"}, {RawName: "a"}, {RawName: "c"}},
		[]string{"a", "c"},
	)
	if len(orderedFields) != 3 || orderedFields[0].RawName != "a" || orderedFields[1].RawName != "c" || orderedFields[2].RawName != "b" {
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

	mapping.DelegateCallArgs = []string{"a=a", "  "}
	callArgs := BuildDelegateCallArgs([]string{"self", "a: str"}, mapping, false)
	if len(callArgs) != 1 || callArgs[0] != "a=a" {
		t.Fatalf("BuildDelegateCallArgs(explicit)=%v", callArgs)
	}
	callArgs = BuildDelegateCallArgs([]string{"self", "a: str", "**kwargs"}, nil, false)
	if len(callArgs) != 2 || callArgs[0] != "a=a" || callArgs[1] != "**kwargs" {
		t.Fatalf("BuildDelegateCallArgs(derived)=%v", callArgs)
	}
}

func TestRenderDelegatedCall(t *testing.T) {
	buf := &bytes.Buffer{}
	RenderDelegatedCall(buf, "apps.list", nil, false, false)
	if got := buf.String(); got != "        return self.apps.list()\n" {
		t.Fatalf("RenderDelegatedCall(sync no args)=%q", got)
	}

	buf.Reset()
	RenderDelegatedCall(buf, "apps.list", []string{"a=a"}, true, false)
	if got := buf.String(); !strings.Contains(got, "return await self.apps.list(") || !strings.Contains(got, "a=a") {
		t.Fatalf("RenderDelegatedCall(async)=%q", got)
	}

	buf.Reset()
	RenderDelegatedCall(buf, "apps.stream", []string{"a=a"}, true, true)
	if got := buf.String(); !strings.Contains(got, "async for item in await self.apps.stream(") || !strings.Contains(got, "yield item") {
		t.Fatalf("RenderDelegatedCall(async yield)=%q", got)
	}
}

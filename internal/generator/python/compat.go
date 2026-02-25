package python

import (
	"bytes"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type renderQueryField = RenderQueryField

func operationArgName(raw string, aliases map[string]string) string {
	return OperationArgName(raw, aliases)
}

func bodyFieldSchema(doc *openapi.Document, bodySchema *openapi.Schema, fieldName string) *openapi.Schema {
	return BodyFieldSchema(doc, bodySchema, fieldName)
}

func typeOverride(rawName string, required bool, defaultType string, overrides map[string]string) string {
	return TypeOverride(rawName, required, defaultType, overrides)
}

func returnTypeInfo(doc *openapi.Document, schema *openapi.Schema) (string, string) {
	return ReturnTypeInfo(doc, schema)
}

func requestBodyTypeInfo(doc *openapi.Document, schema *openapi.Schema, body *openapi.RequestBody) (string, bool) {
	return RequestBodyTypeInfo(doc, schema, body)
}

func schemaTypeName(doc *openapi.Document, schema *openapi.Schema) (string, bool) {
	return SchemaTypeName(doc, schema)
}

func pythonTypeForSchema(doc *openapi.Document, schema *openapi.Schema, required bool) string {
	return PythonTypeForSchema(doc, schema, required)
}

func pythonTypeForSchemaWithAliases(
	doc *openapi.Document,
	schema *openapi.Schema,
	required bool,
	aliases map[string]string,
) string {
	return PythonTypeForSchemaWithAliases(doc, schema, required, aliases)
}

func pythonTypeForSchemaRequiredWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) string {
	return PythonTypeForSchemaRequiredWithAliases(doc, schema, aliases)
}

func schemaTypeNameWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) (string, bool) {
	return SchemaTypeNameWithAliases(doc, schema, aliases)
}

func pythonTypeForSchemaRequired(doc *openapi.Document, schema *openapi.Schema) string {
	return PythonTypeForSchemaRequired(doc, schema)
}

func normalizePackageName(name string) string {
	return NormalizePackageName(name)
}

func normalizePackageDir(sourceDir string, fallback string) string {
	return NormalizePackageDir(sourceDir, fallback)
}

func packageClassName(pkgName string) string {
	return PackageClassName(pkgName)
}

func defaultMethodName(operationID string, path string, method string) string {
	return DefaultMethodName(operationID, path, method)
}

func normalizeMethodName(value string) string {
	return NormalizeMethodName(value)
}

func normalizeClassName(value string) string {
	return NormalizeClassName(value)
}

func normalizePythonIdentifier(value string) string {
	return NormalizePythonIdentifier(value)
}

func toSnake(value string) string {
	return ToSnake(value)
}

func splitIdentifier(value string) []string {
	return SplitIdentifier(value)
}

func collapseUnderscore(value string) string {
	return CollapseUnderscore(value)
}

func escapeDocstring(value string) string {
	return EscapeDocstring(value)
}

func appendIndentedCode(buf *bytes.Buffer, code string, indentLevel int) {
	AppendIndentedCode(buf, code, indentLevel)
}

func ensureTrailingNewlines(buf *bytes.Buffer, newlineCount int) {
	EnsureTrailingNewlines(buf, newlineCount)
}

func linesFromCommentOverride(lines []string) []string {
	return LinesFromCommentOverride(lines)
}

func codeBlocksHaveLeadingDocstring(blocks []string) bool {
	return CodeBlocksHaveLeadingDocstring(blocks)
}

func writeLineComments(buf *bytes.Buffer, indentLevel int, lines []string) {
	WriteLineComments(buf, indentLevel, lines)
}

func normalizedDocstringLines(docstring string) []string {
	return NormalizedDocstringLines(docstring)
}

func writeClassDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	WriteClassDocstring(buf, indentLevel, docstring, style)
}

func writeMethodDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	WriteMethodDocstring(buf, indentLevel, docstring, style)
}

func renderEnumValueLiteral(value interface{}) string {
	return RenderEnumValueLiteral(value)
}

func enumMemberName(value string) string {
	return EnumMemberName(value)
}

func detectMethodBlockName(block string) string {
	return DetectMethodBlockName(block)
}

func parseDefName(trimmedLine string) (string, bool) {
	return ParseDefName(trimmedLine)
}

func isDocstringLine(trimmedLine string) bool {
	return IsDocstringLine(trimmedLine)
}

func renderMethodDocstringLines(docstring string, style string, indent string) []string {
	return RenderMethodDocstringLines(docstring, style, indent)
}

func orderChildClients(children []config.ChildClient, orderedAttrs []string) []config.ChildClient {
	return OrderChildClients(children, orderedAttrs)
}

func indentCodeBlock(block string, level int) string {
	return IndentCodeBlock(block, level)
}

func paginationOrderedFields(fields []renderQueryField, pageSizeField string, pageTokenOrNumField string) []renderQueryField {
	return PaginationOrderedFields(fields, pageSizeField, pageTokenOrNumField)
}

func orderSignatureQueryFields(fields []renderQueryField, orderedRawNames []string) []renderQueryField {
	return OrderSignatureQueryFields(fields, orderedRawNames)
}

func signatureArgName(argDecl string) string {
	return SignatureArgName(argDecl)
}

func isKwargsSignatureArg(argDecl string) bool {
	return IsKwargsSignatureArg(argDecl)
}

func orderSignatureArgs(signatureArgs []string, orderedNames []string) []string {
	return OrderSignatureArgs(signatureArgs, orderedNames)
}

func normalizeSignatureArgs(signatureArgs []string) []string {
	return NormalizeSignatureArgs(signatureArgs)
}

func orderedUniqueByPriority(values []string, priority []string) []string {
	return OrderedUniqueByPriority(values, priority)
}

func operationArgDefault(mapping *config.OperationMapping, rawName string, argName string, async bool) (string, bool) {
	return OperationArgDefault(mapping, rawName, argName, async)
}

func buildDelegateCallArgs(signatureArgs []string, mapping *config.OperationMapping, async bool) []string {
	return BuildDelegateCallArgs(signatureArgs, mapping, async)
}

func renderDelegatedCall(buf *bytes.Buffer, target string, args []string, async bool, asyncYield bool) {
	RenderDelegatedCall(buf, target, args, async, asyncYield)
}

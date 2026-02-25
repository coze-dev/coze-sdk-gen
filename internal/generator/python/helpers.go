package python

import (
	"strings"
	"unicode"

	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func OperationArgName(raw string, aliases map[string]string) string {
	name := raw
	if len(aliases) > 0 {
		if alias, ok := aliases[raw]; ok && strings.TrimSpace(alias) != "" {
			name = alias
		}
	}
	return NormalizePythonIdentifier(name)
}

func BodyFieldSchema(doc *openapi.Document, bodySchema *openapi.Schema, fieldName string) *openapi.Schema {
	if bodySchema == nil {
		return nil
	}
	schema := bodySchema
	if doc != nil {
		schema = doc.ResolveSchema(bodySchema)
	}
	if schema == nil || schema.Properties == nil {
		return nil
	}
	field := schema.Properties[fieldName]
	if field == nil {
		return nil
	}
	if doc != nil {
		return doc.ResolveSchema(field)
	}
	return field
}

func TypeOverride(rawName string, required bool, defaultType string, overrides map[string]string) string {
	if len(overrides) == 0 {
		return defaultType
	}
	override := strings.TrimSpace(overrides[rawName])
	if override == "" {
		return defaultType
	}
	if required {
		return override
	}
	if strings.HasPrefix(override, "Optional[") {
		return override
	}
	return "Optional[" + override + "]"
}

func ReturnTypeInfo(doc *openapi.Document, schema *openapi.Schema) (string, string) {
	_ = doc
	_ = schema
	return "Dict[str, Any]", ""
}

func RequestBodyTypeInfo(doc *openapi.Document, schema *openapi.Schema, body *openapi.RequestBody) (string, bool) {
	_ = doc
	_ = schema
	if schema == nil {
		return "", false
	}
	return "Dict[str, Any]", body != nil && body.Required
}

func SchemaTypeName(doc *openapi.Document, schema *openapi.Schema) (string, bool) {
	if schema == nil {
		return "", false
	}
	if name, ok := doc.SchemaName(schema); ok {
		return NormalizeClassName(name), true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved != nil && resolved != schema {
		if name, ok := doc.SchemaName(resolved); ok {
			return NormalizeClassName(name), true
		}
	}
	return "", false
}

func PythonTypeForSchema(doc *openapi.Document, schema *openapi.Schema, required bool) string {
	return PythonTypeForSchemaWithAliases(doc, schema, required, nil)
}

func PythonTypeForSchemaWithAliases(
	doc *openapi.Document,
	schema *openapi.Schema,
	required bool,
	aliases map[string]string,
) string {
	base := PythonTypeForSchemaRequiredWithAliases(doc, schema, aliases)
	if required {
		return base
	}
	if strings.HasPrefix(base, "Optional[") {
		return base
	}
	return "Optional[" + base + "]"
}

func PythonTypeForSchemaRequiredWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) string {
	if schema == nil {
		return "Any"
	}
	if typeName, ok := SchemaTypeNameWithAliases(doc, schema, aliases); ok {
		return typeName
	}

	resolved := doc.ResolveSchema(schema)
	if resolved == nil {
		return "Any"
	}
	if typeName, ok := SchemaTypeNameWithAliases(doc, resolved, aliases); ok {
		return typeName
	}

	switch resolved.Type {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "List[" + PythonTypeForSchemaRequiredWithAliases(doc, resolved.Items, aliases) + "]"
	case "object":
		return "Dict[str, Any]"
	default:
		if len(resolved.Enum) > 0 {
			return "str"
		}
		return "Any"
	}
}

func SchemaTypeNameWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) (string, bool) {
	if schema == nil {
		return "", false
	}
	if name, ok := doc.SchemaName(schema); ok {
		if alias, exists := aliases[name]; exists && strings.TrimSpace(alias) != "" {
			return alias, true
		}
		if len(aliases) > 0 {
			return "", false
		}
		return NormalizeClassName(name), true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved != nil && resolved != schema {
		if name, ok := doc.SchemaName(resolved); ok {
			if alias, exists := aliases[name]; exists && strings.TrimSpace(alias) != "" {
				return alias, true
			}
			if len(aliases) > 0 {
				return "", false
			}
			return NormalizeClassName(name), true
		}
	}
	return "", false
}

func PythonTypeForSchemaRequired(doc *openapi.Document, schema *openapi.Schema) string {
	return PythonTypeForSchemaRequiredWithAliases(doc, schema, nil)
}

func NormalizePackageName(name string) string {
	name = NormalizePythonIdentifier(name)
	if name == "" {
		return "default"
	}
	return name
}

func NormalizePackageDir(sourceDir string, fallback string) string {
	trimmed := strings.TrimSpace(sourceDir)
	if trimmed == "" {
		return fallback
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "cozepy/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" || trimmed == "." {
		return fallback
	}
	return trimmed
}

func PackageClassName(pkgName string) string {
	return NormalizeClassName(pkgName)
}

func DefaultMethodName(operationID string, path string, method string) string {
	if operationID != "" {
		op := strings.TrimSpace(operationID)
		prefixes := []string{"OpenAPI", "OpenApi", "Openapi", "API", "Api"}
		for _, prefix := range prefixes {
			op = strings.TrimPrefix(op, prefix)
		}
		opSnake := ToSnake(op)
		opSnake = strings.TrimPrefix(opSnake, "open_api_")
		if opSnake != "" {
			return NormalizeMethodName(opSnake)
		}
	}

	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		return NormalizeMethodName(ToSnake(part))
	}
	return NormalizeMethodName(method)
}

func NormalizeMethodName(value string) string {
	raw := strings.TrimSpace(value)
	privateMethod := strings.HasPrefix(raw, "_")
	name := NormalizePythonIdentifier(ToSnake(raw))
	if name == "" {
		if privateMethod {
			return "_call"
		}
		return "call"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "method_" + name
	}
	if privateMethod && !strings.HasPrefix(name, "_") {
		name = "_" + name
	}
	return name
}

func NormalizeClassName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "GeneratedModel"
	}

	snake := ToSnake(trimmed)
	if snake == "" {
		return "GeneratedModel"
	}
	parts := strings.Split(snake, "_")
	if len(parts) == 0 {
		return "GeneratedModel"
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	name := strings.Join(parts, "")
	if name == "" {
		name = "GeneratedModel"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "Model" + name
	}
	return name
}

func NormalizePythonIdentifier(value string) string {
	parts := SplitIdentifier(value)
	if len(parts) == 0 {
		return ""
	}
	name := strings.ToLower(strings.Join(parts, "_"))
	name = CollapseUnderscore(name)
	if name == "" {
		return ""
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "_" + name
	}
	if pythonReservedWords[name] {
		name = name + "_"
	}
	return name
}

func ToSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var out []rune
	prevLowerOrDigit := false
	for _, r := range value {
		if unicode.IsUpper(r) {
			if prevLowerOrDigit && len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
			prevLowerOrDigit = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out = append(out, unicode.ToLower(r))
			prevLowerOrDigit = unicode.IsLower(r) || unicode.IsDigit(r)
			continue
		}
		if len(out) > 0 && out[len(out)-1] != '_' {
			out = append(out, '_')
		}
		prevLowerOrDigit = false
	}
	return strings.Trim(CollapseUnderscore(string(out)), "_")
}

func SplitIdentifier(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parts := make([]string, 0)
	var current []rune
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
			continue
		}
		if len(current) > 0 {
			parts = append(parts, string(current))
			current = current[:0]
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func CollapseUnderscore(value string) string {
	var out []rune
	lastUnderscore := false
	for _, r := range value {
		if r == '_' {
			if lastUnderscore {
				continue
			}
			lastUnderscore = true
			out = append(out, r)
			continue
		}
		lastUnderscore = false
		out = append(out, r)
	}
	return string(out)
}

func EscapeDocstring(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\"\"\"", "\"")
	return strings.TrimSpace(value)
}

var pythonReservedWords = map[string]bool{
	"false":    true,
	"none":     true,
	"true":     true,
	"and":      true,
	"as":       true,
	"assert":   true,
	"async":    true,
	"await":    true,
	"break":    true,
	"class":    true,
	"continue": true,
	"def":      true,
	"del":      true,
	"elif":     true,
	"else":     true,
	"except":   true,
	"finally":  true,
	"for":      true,
	"from":     true,
	"global":   true,
	"if":       true,
	"import":   true,
	"in":       true,
	"is":       true,
	"lambda":   true,
	"nonlocal": true,
	"not":      true,
	"or":       true,
	"pass":     true,
	"raise":    true,
	"return":   true,
	"try":      true,
	"while":    true,
	"with":     true,
	"yield":    true,
}

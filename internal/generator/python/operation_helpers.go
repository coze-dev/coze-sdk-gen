package python

import (
	"bytes"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
)

func IndentCodeBlock(block string, level int) string {
	var buf bytes.Buffer
	AppendIndentedCode(&buf, block, level)
	return buf.String()
}

type RenderQueryField struct {
	RawName      string
	ArgName      string
	ValueExpr    string
	TypeName     string
	Required     bool
	DefaultValue string
}

func PaginationOrderedFields(fields []RenderQueryField, pageSizeField string, pageTokenOrNumField string) []RenderQueryField {
	pageSizeField = strings.TrimSpace(pageSizeField)
	pageTokenOrNumField = strings.TrimSpace(pageTokenOrNumField)
	if pageSizeField == "" {
		pageSizeField = "page_size"
	}
	out := make([]RenderQueryField, 0, len(fields))
	var pageSize *RenderQueryField
	var pageTokenOrNum *RenderQueryField
	for i := range fields {
		field := fields[i]
		switch field.RawName {
		case pageSizeField:
			pageSize = &field
		case pageTokenOrNumField:
			pageTokenOrNum = &field
		default:
			out = append(out, field)
		}
	}
	if pageSize != nil {
		out = append(out, *pageSize)
	}
	if pageTokenOrNum != nil {
		out = append(out, *pageTokenOrNum)
	}
	return out
}

func OrderSignatureQueryFields(fields []RenderQueryField, mapping *config.OperationMapping, async bool) []RenderQueryField {
	if len(fields) <= 1 {
		return fields
	}
	tailFields := make([]RenderQueryField, 0, len(fields))
	requiredFields := make([]RenderQueryField, 0, len(fields))
	optionalWithoutDefault := make([]RenderQueryField, 0, len(fields))
	optionalWithDefault := make([]RenderQueryField, 0, len(fields))
	for _, field := range fields {
		if isSignatureTailField(field) {
			tailFields = append(tailFields, field)
			continue
		}
		defaultValue := strings.TrimSpace(field.DefaultValue)
		if override, ok := OperationArgDefault(mapping, field.RawName, field.ArgName, async); ok {
			defaultValue = override
		}
		if field.Required {
			requiredFields = append(requiredFields, field)
			continue
		}
		if defaultValue == "" {
			optionalWithoutDefault = append(optionalWithoutDefault, field)
			continue
		}
		optionalWithDefault = append(optionalWithDefault, field)
	}
	ordered := make([]RenderQueryField, 0, len(fields))
	ordered = append(ordered, requiredFields...)
	ordered = append(ordered, optionalWithoutDefault...)
	ordered = append(ordered, optionalWithDefault...)
	ordered = append(ordered, tailFields...)
	return ordered
}

func isSignatureTailField(field RenderQueryField) bool {
	names := []string{field.RawName, field.ArgName}
	for _, name := range names {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "page_size", "page_number", "page_num":
			return true
		}
	}
	return false
}

func SignatureArgName(argDecl string) string {
	trimmed := strings.TrimSpace(argDecl)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "**") {
		name := strings.TrimSpace(strings.TrimPrefix(trimmed, "**"))
		name = strings.TrimSpace(strings.SplitN(name, ":", 2)[0])
		name = strings.TrimSpace(strings.SplitN(name, "=", 2)[0])
		return name
	}
	name := strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[0])
	name = strings.TrimSpace(strings.SplitN(name, "=", 2)[0])
	return name
}

func IsKwargsSignatureArg(argDecl string) bool {
	return strings.HasPrefix(strings.TrimSpace(argDecl), "**")
}

func NormalizeSignatureArgs(signatureArgs []string) []string {
	if len(signatureArgs) <= 1 {
		return signatureArgs
	}
	normal := make([]string, 0, len(signatureArgs))
	kwargs := make([]string, 0, 1)
	for _, argDecl := range signatureArgs {
		if IsKwargsSignatureArg(argDecl) {
			kwargs = append(kwargs, argDecl)
			continue
		}
		normal = append(normal, argDecl)
	}
	return append(normal, kwargs...)
}

func OrderedUniqueByPriority(values []string, priority []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	ordered := make([]string, 0, len(values))
	for _, name := range priority {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		for _, value := range values {
			if strings.TrimSpace(value) != trimmed {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				break
			}
			ordered = append(ordered, trimmed)
			seen[trimmed] = struct{}{}
			break
		}
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		ordered = append(ordered, trimmed)
		seen[trimmed] = struct{}{}
	}
	return ordered
}

func OperationArgDefault(mapping *config.OperationMapping, rawName string, argName string, async bool) (string, bool) {
	if mapping == nil {
		return "", false
	}
	defaultMaps := make([]map[string]string, 0, 3)
	if async && len(mapping.ArgDefaultsAsync) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaultsAsync)
	}
	if !async && len(mapping.ArgDefaultsSync) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaultsSync)
	}
	if len(mapping.ArgDefaults) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaults)
	}
	if len(defaultMaps) == 0 {
		return "", false
	}
	if argName = strings.TrimSpace(argName); argName != "" {
		for _, defaults := range defaultMaps {
			if value, ok := defaults[argName]; ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value), true
			}
		}
	}
	if rawName = strings.TrimSpace(rawName); rawName != "" {
		for _, defaults := range defaultMaps {
			if value, ok := defaults[rawName]; ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value), true
			}
		}
	}
	return "", false
}

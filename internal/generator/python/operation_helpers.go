package python

import (
	"bytes"
	"fmt"
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
	if async {
		pageSizeDefault := strings.TrimSpace(mapping.PageSizeDefault)
		rawNameTrimmed := strings.TrimSpace(rawName)
		argNameTrimmed := strings.TrimSpace(argName)
		if pageSizeDefault != "" && (argNameTrimmed == "page_size" || rawNameTrimmed == "page_size") {
			return pageSizeDefault, true
		}
	}
	defaultMaps := make([]map[string]string, 0, 2)
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

func BuildAutoDelegateCallArgs(signatureArgs []string, extraNamedArgs []string) []string {
	args := make([]string, 0, len(signatureArgs)+len(extraNamedArgs))
	for _, argDecl := range signatureArgs {
		trimmed := strings.TrimSpace(argDecl)
		if trimmed == "" || trimmed == "*" {
			continue
		}
		name := SignatureArgName(trimmed)
		if name == "" || name == "self" {
			continue
		}
		if IsKwargsSignatureArg(trimmed) {
			args = append(args, fmt.Sprintf("**%s", name))
			continue
		}
		args = append(args, fmt.Sprintf("%s=%s", name, name))
	}
	for _, extra := range extraNamedArgs {
		args = upsertNamedCallArg(args, extra)
	}
	return args
}

func upsertNamedCallArg(args []string, named string) []string {
	trimmed := strings.TrimSpace(named)
	if trimmed == "" {
		return args
	}
	name, ok := parseNamedCallArgName(trimmed)
	if !ok {
		return insertCallArgBeforeKwargs(args, trimmed)
	}
	for i, arg := range args {
		argName, ok := parseNamedCallArgName(arg)
		if !ok {
			continue
		}
		if argName == name {
			args[i] = trimmed
			return args
		}
	}
	return insertCallArgBeforeKwargs(args, trimmed)
}

func parseNamedCallArgName(arg string) (string, bool) {
	trimmed := strings.TrimSpace(arg)
	if trimmed == "" || strings.HasPrefix(trimmed, "**") {
		return "", false
	}
	left, _, hasAssign := strings.Cut(trimmed, "=")
	if !hasAssign {
		return "", false
	}
	name := strings.TrimSpace(left)
	return name, name != ""
}

func insertCallArgBeforeKwargs(args []string, arg string) []string {
	for i, existing := range args {
		if strings.HasPrefix(strings.TrimSpace(existing), "**") {
			args = append(args, "")
			copy(args[i+1:], args[i:])
			args[i] = arg
			return args
		}
	}
	return append(args, arg)
}

func RenderDelegatedCall(buf *bytes.Buffer, target string, args []string, async bool, asyncYield bool) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}
	qualifier := fmt.Sprintf("self.%s", target)
	if len(args) == 0 {
		if async && asyncYield {
			buf.WriteString(fmt.Sprintf("        async for item in await %s():\n", qualifier))
			buf.WriteString("            yield item\n")
			return
		}
		if async {
			buf.WriteString(fmt.Sprintf("        return await %s()\n", qualifier))
			return
		}
		buf.WriteString(fmt.Sprintf("        return %s()\n", qualifier))
		return
	}

	if async && asyncYield {
		buf.WriteString(fmt.Sprintf("        async for item in await %s(\n", qualifier))
		for _, arg := range args {
			buf.WriteString(fmt.Sprintf("            %s,\n", strings.TrimSpace(arg)))
		}
		buf.WriteString("        ):\n")
		buf.WriteString("            yield item\n")
		return
	}
	if async {
		buf.WriteString(fmt.Sprintf("        return await %s(\n", qualifier))
		for _, arg := range args {
			buf.WriteString(fmt.Sprintf("            %s,\n", strings.TrimSpace(arg)))
		}
		buf.WriteString("        )\n")
		return
	}
	buf.WriteString(fmt.Sprintf("        return %s(\n", qualifier))
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("            %s,\n", strings.TrimSpace(arg)))
	}
	buf.WriteString("        )\n")
}

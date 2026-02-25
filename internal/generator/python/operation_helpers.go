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

func OrderChildClients(children []config.ChildClient, orderedAttrs []string) []config.ChildClient {
	if len(children) == 0 || len(orderedAttrs) == 0 {
		return children
	}
	ordered := make([]config.ChildClient, 0, len(children))
	used := make([]bool, len(children))
	for _, rawAttr := range orderedAttrs {
		attr := strings.TrimSpace(rawAttr)
		if attr == "" {
			continue
		}
		for i, child := range children {
			if used[i] {
				continue
			}
			if strings.TrimSpace(child.Attribute) != attr {
				continue
			}
			ordered = append(ordered, child)
			used[i] = true
			break
		}
	}
	for i, child := range children {
		if used[i] {
			continue
		}
		ordered = append(ordered, child)
	}
	return ordered
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

func OrderSignatureQueryFields(fields []RenderQueryField, orderedRawNames []string) []RenderQueryField {
	if len(fields) == 0 || len(orderedRawNames) == 0 {
		return fields
	}
	fieldByName := make(map[string]RenderQueryField, len(fields))
	for _, field := range fields {
		fieldByName[field.RawName] = field
	}
	result := make([]RenderQueryField, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, rawName := range orderedRawNames {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		field, ok := fieldByName[name]
		if !ok {
			continue
		}
		result = append(result, field)
		seen[name] = struct{}{}
	}
	for _, field := range fields {
		if _, ok := seen[field.RawName]; ok {
			continue
		}
		result = append(result, field)
	}
	return result
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

func BuildDelegateCallArgs(signatureArgs []string, mapping *config.OperationMapping, async bool) []string {
	if mapping != nil {
		explicitArgs := mapping.DelegateCallArgs
		if async && len(mapping.AsyncDelegateCallArgs) > 0 {
			explicitArgs = mapping.AsyncDelegateCallArgs
		}
		if len(explicitArgs) > 0 {
			args := make([]string, 0, len(explicitArgs))
			for _, arg := range explicitArgs {
				trimmed := strings.TrimSpace(arg)
				if trimmed == "" {
					continue
				}
				args = append(args, trimmed)
			}
			return args
		}
	}
	args := make([]string, 0, len(signatureArgs))
	for _, argDecl := range signatureArgs {
		trimmed := strings.TrimSpace(argDecl)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			continue
		}
		name := SignatureArgName(trimmed)
		if name == "" {
			continue
		}
		if name == "self" {
			continue
		}
		if IsKwargsSignatureArg(trimmed) {
			args = append(args, fmt.Sprintf("**%s", name))
			continue
		}
		args = append(args, fmt.Sprintf("%s=%s", name, name))
	}
	return args
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
			trimmed := strings.TrimSpace(arg)
			if trimmed == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("            %s,\n", trimmed))
		}
		buf.WriteString("        ):\n")
		buf.WriteString("            yield item\n")
		return
	}
	if async {
		buf.WriteString(fmt.Sprintf("        return await %s(\n", qualifier))
		for _, arg := range args {
			buf.WriteString(fmt.Sprintf("            %s,\n", arg))
		}
		buf.WriteString("        )\n")
		return
	}
	buf.WriteString(fmt.Sprintf("        return %s(\n", qualifier))
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("            %s,\n", arg))
	}
	buf.WriteString("        )\n")
}

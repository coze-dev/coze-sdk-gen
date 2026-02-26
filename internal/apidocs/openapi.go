package apidocs

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	headingPattern           = regexp.MustCompile(`^(#{1,6})\s*(.+?)\s*$`)
	methodPattern            = regexp.MustCompile(`(?i)\b(get|post|put|delete|patch|head|options)\b`)
	httpURLPattern           = regexp.MustCompile(`https?://[A-Za-z0-9._~:/?#\[\]@!$&'()*+,;=%-]+`)
	markdownLinkInlineRegexp = regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	objectOfPattern          = regexp.MustCompile(`(?i)^object\s+of\s+(.+)$`)
	arrayOfPattern           = regexp.MustCompile(`(?i)^array\s+of\s+(.+)$`)
	inBracketPattern         = regexp.MustCompile(`\[([^\]]+)\]`)
	colonPathParamPattern    = regexp.MustCompile(`/:([A-Za-z0-9_]+)`) // /:id -> /{id}
)

type markdownTable struct {
	Headers []string
	Rows    [][]string
}

type tableBlock struct {
	H2    string
	H3    string
	Table markdownTable
}

type tableField struct {
	Name        string
	Type        string
	RequiredRaw string
	Description string
}

func BuildOpenAPI(link Link, markdown string) (map[string]any, bool, error) {
	title, summary, blocks := collectTableBlocks(markdown)
	if title == "" {
		title = strings.TrimSpace(link.Title)
	}

	method, endpoint := extractMethodAndEndpoint(blocks)
	if method == "" || endpoint == "" {
		return nil, false, nil
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, false, fmt.Errorf("parse endpoint %q: %w", endpoint, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, false, nil
	}

	headerTable, pathTable, queryTable, bodyTable, responseTable, componentTables := splitSections(blocks)
	components := buildComponentSchemas(componentTables)
	parameters := append(tableToParameters(pathTable, "path"), tableToParameters(queryTable, "query")...)
	parameters = append(parameters, tableToParameters(headerTable, "header")...)

	operation := map[string]any{
		"operationId":  buildOperationID(link.URL, method),
		"x-source-url": link.URL,
	}
	if title != "" {
		operation["summary"] = title
	}
	if summary != "" {
		operation["description"] = summary
	}
	if len(parameters) > 0 {
		operation["parameters"] = parameters
	}

	if bodyTable != nil {
		if schema := tableToObjectSchema(*bodyTable, ""); len(schema) > 0 {
			operation["requestBody"] = map[string]any{
				"required": true,
				"content": map[string]any{
					inferRequestContentType(headerTable): map[string]any{
						"schema": schema,
					},
				},
			}
		}
	}

	responses := map[string]any{"200": map[string]any{"description": "OK"}}
	if responseTable != nil {
		if schema := tableToObjectSchema(*responseTable, ""); len(schema) > 0 {
			responses = map[string]any{
				"200": map[string]any{
					"description": "OK",
					"content": map[string]any{
						"application/json": map[string]any{"schema": schema},
					},
				},
			}
		}
	}
	operation["responses"] = responses

	server := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	pathKey := normalizeOpenAPIPath(u.Path)
	if pathKey == "" {
		pathKey = "/"
	}

	doc := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   chooseNonEmpty(title, link.Title, "Coze API"),
			"version": "1.0.0",
		},
		"servers": []map[string]any{{"url": server}},
		"paths": map[string]any{
			pathKey: map[string]any{
				strings.ToLower(method): operation,
			},
		},
	}
	if len(components) > 0 {
		doc["components"] = map[string]any{"schemas": components}
	}

	return doc, true, nil
}

func extractMethodAndEndpoint(blocks []tableBlock) (string, string) {
	for _, block := range blocks {
		if !headingEquals(block.H2, "基础信息") {
			continue
		}
		method := ""
		endpoint := ""

		collect := func(key, value string) {
			n := normalizeHeading(key)
			switch {
			case strings.Contains(n, "请求方式"):
				if method == "" {
					method = extractMethod(value)
				}
			case strings.Contains(n, "请求地址"):
				if endpoint == "" {
					endpoint = extractEndpointURL(value)
				}
			}
		}

		if len(block.Table.Headers) >= 2 {
			collect(block.Table.Headers[0], strings.Join(block.Table.Headers[1:], " "))
		}
		for _, row := range block.Table.Rows {
			if len(row) < 2 {
				continue
			}
			collect(row[0], strings.Join(row[1:], " "))
		}
		if method != "" && endpoint != "" {
			return method, endpoint
		}
	}
	return "", ""
}

func splitSections(blocks []tableBlock) (
	headerTable *markdownTable,
	pathTable *markdownTable,
	queryTable *markdownTable,
	bodyTable *markdownTable,
	responseTable *markdownTable,
	componentTables map[string]markdownTable,
) {
	componentTables = make(map[string]markdownTable)
	for _, block := range blocks {
		switch {
		case headingEquals(block.H2, "请求参数"):
			sub := normalizeHeading(block.H3)
			switch sub {
			case "header":
				t := block.Table
				headerTable = &t
			case "path":
				t := block.Table
				pathTable = &t
			case "query":
				t := block.Table
				queryTable = &t
			case "body":
				t := block.Table
				bodyTable = &t
			default:
				name := sanitizeComponentName(block.H3)
				if name != "" {
					if _, exists := componentTables[name]; !exists {
						componentTables[name] = block.Table
					}
				}
			}
		case headingEquals(block.H2, "返回参数"):
			if strings.TrimSpace(block.H3) == "" {
				if responseTable == nil {
					t := block.Table
					responseTable = &t
				}
				continue
			}
			name := sanitizeComponentName(block.H3)
			if name != "" {
				if _, exists := componentTables[name]; !exists {
					componentTables[name] = block.Table
				}
			}
		}
	}
	return
}

func buildComponentSchemas(componentTables map[string]markdownTable) map[string]any {
	if len(componentTables) == 0 {
		return nil
	}
	keys := make([]string, 0, len(componentTables))
	for k := range componentTables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	schemas := make(map[string]any, len(componentTables))
	for _, key := range keys {
		schema := tableToObjectSchema(componentTables[key], "")
		if len(schema) == 0 {
			continue
		}
		schemas[key] = schema
	}
	if len(schemas) == 0 {
		return nil
	}
	return schemas
}

func tableToObjectSchema(table markdownTable, fallbackType string) map[string]any {
	fields := extractFields(table)
	if len(fields) == 0 {
		return nil
	}
	properties := make(map[string]any, len(fields))
	required := make([]string, 0)
	for _, field := range fields {
		schema := schemaForType(field.Type)
		if len(schema) == 0 && fallbackType != "" {
			schema = schemaForType(fallbackType)
		}
		if len(schema) == 0 {
			schema = map[string]any{"type": "string"}
		}
		if desc := cleanupInline(field.Description); desc != "" {
			schema["description"] = desc
		}
		properties[field.Name] = schema
		if isRequired(field.RequiredRaw) {
			required = append(required, field.Name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func tableToParameters(table *markdownTable, in string) []map[string]any {
	if table == nil {
		return nil
	}
	fields := extractFields(*table)
	if len(fields) == 0 {
		return nil
	}

	params := make([]map[string]any, 0, len(fields))
	for _, field := range fields {
		schema := schemaForType(field.Type)
		if len(schema) == 0 {
			schema = map[string]any{"type": "string"}
		}

		required := isRequired(field.RequiredRaw)
		if in == "path" {
			required = true
		}

		param := map[string]any{
			"name":     field.Name,
			"in":       in,
			"required": required,
			"schema":   schema,
		}
		if desc := cleanupInline(field.Description); desc != "" {
			param["description"] = desc
		}
		params = append(params, param)
	}
	return params
}

func extractFields(table markdownTable) []tableField {
	if len(table.Headers) == 0 || len(table.Rows) == 0 {
		return nil
	}
	nameIdx := findHeaderIndex(table.Headers, "参数", "字段", "parameter", "name")
	typeIdx := findHeaderIndex(table.Headers, "类型", "type")
	requiredIdx := findHeaderIndex(table.Headers, "是否必选", "必选", "required")
	descIdx := findHeaderIndex(table.Headers, "说明", "描述", "description")
	valueIdx := findHeaderIndex(table.Headers, "取值", "value", "示例", "example")
	if nameIdx < 0 {
		nameIdx = 0
	}

	fields := make([]tableField, 0, len(table.Rows))
	for _, row := range table.Rows {
		if nameIdx >= len(row) {
			continue
		}
		name := sanitizeFieldName(row[nameIdx])
		if name == "" {
			continue
		}
		field := tableField{Name: name}
		if typeIdx >= 0 && typeIdx < len(row) {
			field.Type = strings.TrimSpace(row[typeIdx])
		}
		if requiredIdx >= 0 && requiredIdx < len(row) {
			field.RequiredRaw = strings.TrimSpace(row[requiredIdx])
		}
		if descIdx >= 0 && descIdx < len(row) {
			field.Description = strings.TrimSpace(row[descIdx])
		} else if valueIdx >= 0 && valueIdx < len(row) {
			field.Description = strings.TrimSpace(row[valueIdx])
		}
		fields = append(fields, field)
	}
	return fields
}

func schemaForType(rawType string) map[string]any {
	typeText := cleanupInline(rawType)
	typeText = strings.TrimSpace(typeText)
	if typeText == "" {
		return nil
	}

	if m := arrayOfPattern.FindStringSubmatch(typeText); len(m) == 2 {
		inner := strings.TrimSpace(m[1])
		item := schemaForType(inner)
		if len(item) == 0 {
			item = map[string]any{"type": "string"}
		}
		return map[string]any{"type": "array", "items": item}
	}
	if m := objectOfPattern.FindStringSubmatch(typeText); len(m) == 2 {
		inner := strings.TrimSpace(m[1])
		if ref := extractRefName(inner); ref != "" {
			return map[string]any{"$ref": "#/components/schemas/" + ref}
		}
		return map[string]any{"type": "object"}
	}

	lower := strings.ToLower(typeText)
	switch {
	case strings.Contains(lower, "json map"):
		return map[string]any{"type": "object", "additionalProperties": true}
	case strings.Contains(lower, " map") || strings.HasSuffix(lower, "map"):
		return map[string]any{"type": "object", "additionalProperties": true}
	case strings.Contains(lower, "string"):
		return map[string]any{"type": "string"}
	case strings.Contains(lower, "long"):
		return map[string]any{"type": "integer", "format": "int64"}
	case strings.Contains(lower, "integer") || strings.Contains(lower, "int"):
		return map[string]any{"type": "integer"}
	case strings.Contains(lower, "double") || strings.Contains(lower, "float") || strings.Contains(lower, "number"):
		return map[string]any{"type": "number", "format": "double"}
	case strings.Contains(lower, "boolean") || strings.Contains(lower, "bool"):
		return map[string]any{"type": "boolean"}
	case strings.Contains(lower, "binary") || strings.Contains(lower, "file"):
		return map[string]any{"type": "string", "format": "binary"}
	case strings.Contains(lower, "object"):
		if ref := extractRefName(typeText); ref != "" {
			return map[string]any{"$ref": "#/components/schemas/" + ref}
		}
		return map[string]any{"type": "object"}
	default:
		if ref := extractRefName(typeText); ref != "" {
			return map[string]any{"$ref": "#/components/schemas/" + ref}
		}
		return map[string]any{"type": "string"}
	}
}

func extractRefName(raw string) string {
	if raw == "" {
		return ""
	}
	if m := inBracketPattern.FindStringSubmatch(raw); len(m) == 2 {
		return sanitizeComponentName(m[1])
	}
	clean := cleanupInline(raw)
	if clean == "" {
		return ""
	}
	if strings.Contains(clean, " ") {
		clean = strings.Fields(clean)[0]
	}
	clean = sanitizeComponentName(clean)
	lower := strings.ToLower(clean)
	switch lower {
	case "", "string", "integer", "int", "long", "double", "float", "number", "boolean", "bool", "object", "array", "json", "map":
		return ""
	default:
		return clean
	}
}

func inferRequestContentType(headerTable *markdownTable) string {
	if headerTable == nil {
		return "application/json"
	}
	for _, row := range headerTable.Rows {
		if len(row) < 2 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(cleanupInline(row[0])))
		name = strings.ReplaceAll(name, " ", "")
		if name != "content-type" {
			continue
		}
		value := strings.ToLower(cleanupInline(strings.Join(row[1:], " ")))
		switch {
		case strings.Contains(value, "multipart/form-data"):
			return "multipart/form-data"
		case strings.Contains(value, "application/x-www-form-urlencoded"):
			return "application/x-www-form-urlencoded"
		case strings.Contains(value, "text/plain"):
			return "text/plain"
		case strings.Contains(value, "application/json"):
			return "application/json"
		}
	}
	return "application/json"
}

func collectTableBlocks(markdown string) (string, string, []tableBlock) {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	title := ""
	summary := firstParagraphAfterTitle(lines)

	currentH2 := ""
	currentH3 := ""
	blocks := make([]tableBlock, 0)
	for i := 0; i < len(lines); {
		line := strings.TrimSpace(lines[i])
		if level, heading, ok := parseHeading(line); ok {
			heading = cleanupInline(heading)
			if level == 1 && title == "" {
				title = heading
			}
			if level == 2 {
				currentH2 = heading
				currentH3 = ""
			} else if level == 3 {
				currentH3 = heading
			}
			i++
			continue
		}
		if table, next, ok := parseTable(lines, i); ok {
			blocks = append(blocks, tableBlock{H2: currentH2, H3: currentH3, Table: table})
			i = next
			continue
		}
		i++
	}
	return title, summary, blocks
}

func parseHeading(line string) (int, string, bool) {
	m := headingPattern.FindStringSubmatch(strings.TrimSpace(line))
	if len(m) != 3 {
		return 0, "", false
	}
	return len(m[1]), strings.TrimSpace(m[2]), true
}

func parseTable(lines []string, start int) (markdownTable, int, bool) {
	if start+1 >= len(lines) {
		return markdownTable{}, start, false
	}
	if !isTableLine(lines[start]) || !isTableSeparatorLine(lines[start+1]) {
		return markdownTable{}, start, false
	}

	headers := splitTableLine(lines[start])
	if len(headers) == 0 {
		return markdownTable{}, start, false
	}
	rows := make([][]string, 0)
	i := start + 2
	for i < len(lines) && isTableLine(lines[i]) {
		row := splitTableLine(lines[i])
		if len(row) < len(headers) {
			filled := make([]string, len(headers))
			copy(filled, row)
			row = filled
		}
		rows = append(rows, row)
		i++
	}
	return markdownTable{Headers: headers, Rows: rows}, i, true
}

func isTableLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "|") && strings.Contains(trimmed, "|")
}

func isTableSeparatorLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	trimmed = strings.Trim(trimmed, "| ")
	if trimmed == "" {
		return false
	}
	parts := strings.Split(trimmed, "|")
	if len(parts) == 0 {
		return false
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Trim(p, "-:") != "" {
			return false
		}
	}
	return true
}

func splitTableLine(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	if trimmed == "" {
		return nil
	}

	cells := make([]string, 0)
	var b strings.Builder
	escaped := false
	for _, r := range trimmed {
		if escaped {
			if r != '|' {
				b.WriteRune('\\')
			}
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == '|' {
			cells = append(cells, strings.TrimSpace(b.String()))
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	cells = append(cells, strings.TrimSpace(b.String()))

	for i := range cells {
		cells[i] = cleanupInline(cells[i])
	}
	return cells
}

func findHeaderIndex(headers []string, keywords ...string) int {
	normalized := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		normalized = append(normalized, normalizeHeading(kw))
	}
	for i, header := range headers {
		h := normalizeHeading(header)
		for _, kw := range normalized {
			if kw != "" && strings.Contains(h, kw) {
				return i
			}
		}
	}
	return -1
}

func headingEquals(heading string, expected string) bool {
	return normalizeHeading(heading) == normalizeHeading(expected)
}

func normalizeHeading(in string) string {
	in = strings.ToLower(cleanupInline(in))
	replacer := strings.NewReplacer(" ", "", "_", "", "*", "", "`", "", "\t", "", "\n", "")
	return replacer.Replace(strings.TrimSpace(in))
}

func sanitizeFieldName(name string) string {
	name = cleanupInline(name)
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if strings.Contains(name, " ") {
		name = strings.Fields(name)[0]
	}
	name = strings.Trim(name, "`* ")
	return name
}

func sanitizeComponentName(name string) string {
	name = cleanupInline(name)
	if name == "" {
		return ""
	}
	if strings.Contains(name, " ") {
		name = strings.Fields(name)[0]
	}
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return ""
	}
	first := rune(out[0])
	if unicode.IsDigit(first) {
		return "Model" + out
	}
	return out
}

func cleanupInline(in string) string {
	if in == "" {
		return ""
	}
	out := strings.TrimSpace(in)
	out = strings.ReplaceAll(out, "<br>", "\n")
	out = strings.ReplaceAll(out, "<br/>", "\n")
	out = strings.ReplaceAll(out, "<br />", "\n")
	out = strings.ReplaceAll(out, "```", "")
	out = strings.ReplaceAll(out, "`", "")
	out = strings.ReplaceAll(out, "**", "")
	out = strings.ReplaceAll(out, "__", "")
	out = markdownLinkInlineRegexp.ReplaceAllString(out, "$1")
	return strings.TrimSpace(out)
}

func extractMethod(value string) string {
	m := methodPattern.FindStringSubmatch(value)
	if len(m) != 2 {
		return ""
	}
	return strings.ToUpper(m[1])
}

func extractEndpointURL(value string) string {
	m := httpURLPattern.FindString(value)
	if m == "" {
		return ""
	}
	return strings.TrimSpace(m)
}

func normalizeOpenAPIPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	path = colonPathParamPattern.ReplaceAllString(path, "/{$1}")
	return path
}

func buildOperationID(rawURL string, method string) string {
	name := baseNameFromURL(rawURL)
	if name == "" {
		name = "operation"
	}
	return fmt.Sprintf("%s_%s", strings.ToLower(method), name)
}

func firstParagraphAfterTitle(lines []string) string {
	start := -1
	for i, line := range lines {
		if level, _, ok := parseHeading(strings.TrimSpace(line)); ok && level == 1 {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	parts := make([]string, 0)
	for i := start; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if len(parts) > 0 {
				break
			}
			continue
		}
		if _, _, ok := parseHeading(line); ok {
			break
		}
		parts = append(parts, cleanupInline(line))
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func isRequired(value string) bool {
	v := strings.ToLower(cleanupInline(value))
	return strings.Contains(v, "必选") || strings.Contains(v, "required") || strings.Contains(v, "yes") || strings.Contains(v, "true")
}

func chooseNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

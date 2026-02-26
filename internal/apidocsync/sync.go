package apidocsync

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultLLMSURL     = "https://docs.coze.cn/llms.txt"
	defaultSection     = "developer_guides"
	defaultOutputRoot  = "docs"
	defaultMarkdownDir = "api-markdown"
	defaultSwaggerDir  = "api-swagger"
	defaultHTTPTimeout = 30 * time.Second
	defaultSpecVersion = "3.0.3"
	defaultDocVersion  = "1.0.0"
	defaultSuccessCode = "200"
	defaultContentType = "application/json"
)

var (
	listItemPattern   = regexp.MustCompile(`^- \[(.+)]\((https?://[^\s)]+/api/open/docs/developer_guides/[^\s)]+)\)\s*$`)
	htmlTagPattern    = regexp.MustCompile(`<[^>]+>`)
	markdownLinkRegex = regexp.MustCompile(`\[([^\]]+)]\([^)]*\)`)
	urlPattern        = regexp.MustCompile(`(?i)(https?|wss?)://[^\s<>()]+`)
	firstWordPattern  = regexp.MustCompile(`[A-Za-z]+`)
	objectOfPattern   = regexp.MustCompile(`(?i)^object\s+of\s+(.+)$`)
	arrayOfPattern    = regexp.MustCompile(`(?i)^array\s+of\s+(.+)$`)
	pathParamPattern  = regexp.MustCompile(`\{([^{}]+)\}`)
	permissionPattern = regexp.MustCompile("`([^`]+)`")
)

var ignoredSchemaSectionNames = map[string]struct{}{
	"header":  {},
	"query":   {},
	"path":    {},
	"body":    {},
	"请求示例":    {},
	"返回示例":    {},
	"非流式响应":   {},
	"流式响应":    {},
	"错误码":     {},
	"接口限制":    {},
	"接口说明":    {},
	"接口信息":    {},
	"基础信息":    {},
	"请求参数":    {},
	"返回参数":    {},
	"示例":      {},
	"api 时序图": {},
}

// Options defines configurable behavior for syncing Coze API docs.
type Options struct {
	LLMSURL        string
	Section        string
	OutputRoot     string
	MarkdownSubdir string
	SwaggerSubdir  string
	HTTPTimeout    time.Duration
}

// Result captures aggregate sync statistics.
type Result struct {
	Section         string
	TotalCandidates int
	Generated       int
	Skipped         int
	MarkdownDir     string
	SwaggerDir      string
}

// Run downloads docs in the configured section and writes markdown and Swagger files.
func Run(ctx context.Context, stdout io.Writer, opts Options) (Result, error) {
	opts = opts.withDefaults()

	client := &http.Client{Timeout: opts.HTTPTimeout}

	llmsContent, err := fetchText(ctx, client, opts.LLMSURL)
	if err != nil {
		return Result{}, fmt.Errorf("fetch llms index: %w", err)
	}

	links, err := parseSectionLinks(llmsContent, opts.Section)
	if err != nil {
		return Result{}, err
	}
	if len(links) == 0 {
		return Result{}, fmt.Errorf("section %q has no entries", opts.Section)
	}

	markdownDir := filepath.Join(opts.OutputRoot, opts.MarkdownSubdir)
	swaggerDir := filepath.Join(opts.OutputRoot, opts.SwaggerSubdir)
	if err := recreateDir(markdownDir); err != nil {
		return Result{}, fmt.Errorf("prepare markdown dir: %w", err)
	}
	if err := recreateDir(swaggerDir); err != nil {
		return Result{}, fmt.Errorf("prepare swagger dir: %w", err)
	}

	result := Result{
		Section:         opts.Section,
		TotalCandidates: len(links),
		MarkdownDir:     markdownDir,
		SwaggerDir:      swaggerDir,
	}

	for _, link := range links {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}

		markdown, err := fetchText(ctx, client, link.URL)
		if err != nil {
			return Result{}, fmt.Errorf("fetch %s: %w", link.URL, err)
		}

		apiDoc, ok := parseAPIDoc(link, markdown)
		if !ok {
			result.Skipped++
			continue
		}

		markdownPath := filepath.Join(markdownDir, link.Slug+".md")
		if err := os.WriteFile(markdownPath, []byte(markdown), 0o644); err != nil {
			return Result{}, fmt.Errorf("write markdown %s: %w", markdownPath, err)
		}

		swaggerYAML, err := buildSwaggerYAML(apiDoc)
		if err != nil {
			return Result{}, fmt.Errorf("build swagger for %s: %w", link.URL, err)
		}
		swaggerPath := filepath.Join(swaggerDir, link.Slug+".yaml")
		if err := os.WriteFile(swaggerPath, swaggerYAML, 0o644); err != nil {
			return Result{}, fmt.Errorf("write swagger %s: %w", swaggerPath, err)
		}

		result.Generated++
	}

	if stdout != nil {
		_, _ = fmt.Fprintf(
			stdout,
			"section=%s total=%d generated=%d skipped=%d markdown_dir=%s swagger_dir=%s\n",
			result.Section,
			result.TotalCandidates,
			result.Generated,
			result.Skipped,
			result.MarkdownDir,
			result.SwaggerDir,
		)
	}

	return result, nil
}

func (o Options) withDefaults() Options {
	if strings.TrimSpace(o.LLMSURL) == "" {
		o.LLMSURL = defaultLLMSURL
	}
	if strings.TrimSpace(o.Section) == "" {
		o.Section = defaultSection
	}
	if strings.TrimSpace(o.OutputRoot) == "" {
		o.OutputRoot = defaultOutputRoot
	}
	if strings.TrimSpace(o.MarkdownSubdir) == "" {
		o.MarkdownSubdir = defaultMarkdownDir
	}
	if strings.TrimSpace(o.SwaggerSubdir) == "" {
		o.SwaggerSubdir = defaultSwaggerDir
	}
	if o.HTTPTimeout <= 0 {
		o.HTTPTimeout = defaultHTTPTimeout
	}
	return o
}

func fetchText(ctx context.Context, client *http.Client, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(body), "\r\n", "\n"), nil
}

func recreateDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

type docLink struct {
	Title string
	URL   string
	Slug  string
}

func parseSectionLinks(content string, section string) ([]docLink, error) {
	lines := strings.Split(content, "\n")
	inSection := false
	seenURL := map[string]struct{}{}
	links := make([]docLink, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### ") {
			headingName := strings.TrimSpace(strings.TrimPrefix(trimmed, "### "))
			if inSection {
				break
			}
			if headingName == section {
				inSection = true
			}
			continue
		}
		if !inSection {
			continue
		}

		match := listItemPattern.FindStringSubmatch(trimmed)
		if len(match) != 3 {
			continue
		}

		title := strings.TrimSpace(match[1])
		rawURL := strings.TrimSpace(match[2])
		if _, exists := seenURL[rawURL]; exists {
			continue
		}
		seenURL[rawURL] = struct{}{}

		links = append(links, docLink{
			Title: title,
			URL:   rawURL,
			Slug:  makeSlug(rawURL, title),
		})
	}

	if !inSection {
		return nil, fmt.Errorf("section %q not found in llms index", section)
	}

	sort.Slice(links, func(i, j int) bool {
		if links[i].Slug == links[j].Slug {
			return links[i].URL < links[j].URL
		}
		return links[i].Slug < links[j].Slug
	})

	return ensureUniqueSlugs(links), nil
}

func ensureUniqueSlugs(links []docLink) []docLink {
	baseUsed := map[string]int{}
	for i := range links {
		base := links[i].Slug
		baseUsed[base]++
		if baseUsed[base] > 1 {
			links[i].Slug = fmt.Sprintf("%s_%d", base, baseUsed[base])
		}
	}
	return links
}

func makeSlug(rawURL string, title string) string {
	u, err := url.Parse(rawURL)
	if err == nil {
		base := path.Base(strings.Trim(u.Path, "/"))
		if slug := sanitizeSlug(base); slug != "" {
			return slug
		}
	}
	if slug := sanitizeSlug(title); slug != "" {
		return slug
	}
	return "api_doc"
}

func sanitizeSlug(value string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9':
			b.WriteRune(ch)
			lastUnderscore = false
		case ch == '_' || ch == '-':
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		default:
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}
	result := strings.Trim(b.String(), "_")
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}
	return strings.ToLower(result)
}

type heading struct {
	Level int
	Title string
	Line  int
}

type markdownTable struct {
	Headers []string
	Rows    [][]string
}

type docField struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

type apiDoc struct {
	Link             docLink
	Title            string
	Description      string
	InterfaceDesc    string
	Permission       string
	Path             string
	ServerURL        string
	HTTPMethod       string
	OriginalMethod   string
	IsWebsocket      bool
	QueryParams      []docField
	PathParams       []docField
	HeaderParams     []docField
	RequestBody      []docField
	ResponseBody     []docField
	ComponentSchemas map[string][]docField
}

func parseAPIDoc(link docLink, markdown string) (apiDoc, bool) {
	lines := strings.Split(markdown, "\n")
	headings := collectHeadings(lines)
	if len(headings) == 0 {
		return apiDoc{}, false
	}

	title := parseTitle(lines, headings, link.Title)
	description := parseLeadDescription(lines, headings)

	methodValue, addressValue, permission, interfaceDesc := parseInterfaceInfo(lines, headings)
	if strings.TrimSpace(addressValue) == "" {
		return apiDoc{}, false
	}

	rawURL, serverURL, pathValue := extractEndpoint(addressValue)
	if pathValue == "" {
		return apiDoc{}, false
	}
	method, originalMethod, isWebsocket := normalizeMethod(methodValue, rawURL)

	queryParams := []docField{}
	pathParams := []docField{}
	headerParams := []docField{}
	requestBody := []docField{}
	responseBody := []docField{}
	componentSchemas := map[string][]docField{}

	requestStart, requestEnd, hasRequest := findSectionRange(lines, headings, "请求参数")
	if hasRequest {
		queryParams, pathParams, headerParams, requestBody = parseRequestParameters(lines, headings, requestStart, requestEnd)
	}

	responseStart, responseEnd, hasResponse := findSectionRange(lines, headings, "返回参数")
	if hasResponse {
		if table := findFirstTable(lines, responseStart, responseEnd); table != nil {
			responseBody = parseFields(table, "response")
		}
	}

	componentSchemas = parseSchemaSections(lines, headings)

	pathParams = mergePathParamsFromPath(pathValue, pathParams)

	return apiDoc{
		Link:             link,
		Title:            title,
		Description:      description,
		InterfaceDesc:    interfaceDesc,
		Permission:       permission,
		Path:             pathValue,
		ServerURL:        serverURL,
		HTTPMethod:       method,
		OriginalMethod:   originalMethod,
		IsWebsocket:      isWebsocket,
		QueryParams:      queryParams,
		PathParams:       pathParams,
		HeaderParams:     headerParams,
		RequestBody:      requestBody,
		ResponseBody:     responseBody,
		ComponentSchemas: componentSchemas,
	}, true
}

func collectHeadings(lines []string) []heading {
	result := make([]heading, 0)
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "#") {
			continue
		}

		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level == 0 || level > 6 {
			continue
		}
		if len(trimmed) <= level || trimmed[level] != ' ' {
			continue
		}

		title := strings.TrimSpace(trimmed[level+1:])
		result = append(result, heading{Level: level, Title: title, Line: i})
	}
	return result
}

func parseTitle(lines []string, headings []heading, fallback string) string {
	for _, h := range headings {
		if h.Level == 1 {
			if title := cleanText(h.Title); title != "" {
				return title
			}
		}
	}
	if fallback != "" {
		return fallback
	}
	return "Coze API"
}

func parseLeadDescription(lines []string, headings []heading) string {
	if len(headings) == 0 {
		return ""
	}

	start := headings[0].Line + 1
	end := len(lines)
	for _, h := range headings {
		if h.Level == 2 {
			end = h.Line
			break
		}
	}

	parts := make([]string, 0)
	for i := start; i < end; i++ {
		line := cleanText(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "|") || strings.HasPrefix(line, "```") {
			continue
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, "\n")
}

func parseInterfaceInfo(lines []string, headings []heading) (method string, address string, permission string, interfaceDesc string) {
	for _, sectionName := range []string{"基础信息", "接口信息"} {
		start, end, ok := findSectionRange(lines, headings, sectionName)
		if !ok {
			continue
		}
		table := findFirstTable(lines, start, end)
		if table == nil {
			continue
		}
		kv := parseKeyValueTable(table)

		if method == "" {
			method = cleanText(getKV(kv, "请求方式"))
		}
		if address == "" {
			address = firstNonEmpty(getKV(kv, "请求地址"), getKV(kv, "URL"))
		}
		if permission == "" {
			permission = parsePermission(getKV(kv, "权限"))
		}
		if interfaceDesc == "" {
			interfaceDesc = cleanText(getKV(kv, "接口说明"))
		}
	}
	return method, address, permission, interfaceDesc
}

func parsePermission(raw string) string {
	if match := permissionPattern.FindStringSubmatch(raw); len(match) == 2 {
		return cleanText(match[1])
	}
	cleaned := cleanText(raw)
	if cleaned == "" {
		return ""
	}
	parts := strings.Fields(cleaned)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func findSectionRange(lines []string, headings []heading, sectionTitle string) (start int, end int, ok bool) {
	for i, h := range headings {
		if cleanText(h.Title) != sectionTitle {
			continue
		}
		start = h.Line + 1
		end = len(lines)
		for j := i + 1; j < len(headings); j++ {
			if headings[j].Level <= h.Level {
				end = headings[j].Line
				break
			}
		}
		return start, end, true
	}
	return 0, 0, false
}

func findFirstTable(lines []string, start int, end int) *markdownTable {
	if start < 0 {
		start = 0
	}
	if end > len(lines) {
		end = len(lines)
	}
	if start >= end {
		return nil
	}

	for i := start; i+1 < end; i++ {
		if !isTableRow(lines[i]) || !isSeparatorRow(lines[i+1]) {
			continue
		}
		j := i + 2
		for j < end && isTableRow(lines[j]) {
			j++
		}
		return parseTable(lines[i:j])
	}
	return nil
}

func parseRequestParameters(lines []string, headings []heading, start int, end int) (
	queryParams []docField,
	pathParams []docField,
	headerParams []docField,
	bodyParams []docField,
) {
	type section struct {
		Title string
		Start int
		End   int
	}

	sections := make([]section, 0)
	for i, h := range headings {
		if h.Line < start || h.Line >= end || h.Level != 3 {
			continue
		}
		sectionStart := h.Line + 1
		sectionEnd := end
		for j := i + 1; j < len(headings); j++ {
			if headings[j].Line >= end {
				break
			}
			if headings[j].Level <= h.Level {
				sectionEnd = headings[j].Line
				break
			}
		}
		sections = append(sections, section{Title: cleanText(h.Title), Start: sectionStart, End: sectionEnd})
	}

	hadStructuredSection := false
	for _, sec := range sections {
		table := findFirstTable(lines, sec.Start, sec.End)
		if table == nil {
			continue
		}
		fields := parseFields(table, sec.Title)
		if len(fields) == 0 {
			continue
		}
		hadStructuredSection = true

		switch strings.ToLower(sec.Title) {
		case "header":
			headerParams = append(headerParams, fields...)
		case "query":
			queryParams = append(queryParams, fields...)
		case "path":
			pathParams = append(pathParams, fields...)
		case "body":
			bodyParams = append(bodyParams, fields...)
		}
	}

	if hadStructuredSection {
		return queryParams, pathParams, headerParams, bodyParams
	}

	if table := findFirstTable(lines, start, end); table != nil {
		bodyParams = parseFields(table, "body")
	}
	return queryParams, pathParams, headerParams, bodyParams
}

func parseSchemaSections(lines []string, headings []heading) map[string][]docField {
	result := map[string][]docField{}
	for i, h := range headings {
		if h.Level != 3 {
			continue
		}
		title := cleanText(h.Title)
		if title == "" {
			continue
		}
		if _, ignored := ignoredSchemaSectionNames[strings.ToLower(title)]; ignored {
			continue
		}
		if _, ignored := ignoredSchemaSectionNames[title]; ignored {
			continue
		}

		start := h.Line + 1
		end := len(lines)
		for j := i + 1; j < len(headings); j++ {
			if headings[j].Level <= h.Level {
				end = headings[j].Line
				break
			}
		}

		table := findFirstTable(lines, start, end)
		if table == nil || !looksLikeSchemaTable(table.Headers) {
			continue
		}

		schemaName := sanitizeSchemaName(title)
		if schemaName == "" {
			continue
		}
		fields := parseFields(table, "schema")
		if len(fields) == 0 {
			continue
		}
		result[schemaName] = fields
	}
	return result
}

func looksLikeSchemaTable(headers []string) bool {
	hasParam := false
	hasType := false
	for _, h := range headers {
		cell := cleanText(h)
		if strings.Contains(cell, "参数") {
			hasParam = true
		}
		if strings.Contains(cell, "类型") {
			hasType = true
		}
	}
	return hasParam && hasType
}

func parseFields(table *markdownTable, section string) []docField {
	if table == nil || len(table.Rows) == 0 {
		return nil
	}

	idxName := 0
	idxType := -1
	idxValue := -1
	idxRequired := -1
	idxDescription := -1

	for i, h := range table.Headers {
		name := cleanText(h)
		switch {
		case strings.Contains(name, "参数") || strings.Contains(strings.ToLower(name), "parameter"):
			idxName = i
		case strings.Contains(name, "类型") || strings.Contains(strings.ToLower(name), "type"):
			idxType = i
		case strings.Contains(name, "取值"):
			idxValue = i
		case strings.Contains(name, "必选") || strings.Contains(strings.ToLower(name), "required"):
			idxRequired = i
		case strings.Contains(name, "说明") || strings.Contains(strings.ToLower(name), "description"):
			idxDescription = i
		}
	}

	fields := make([]docField, 0, len(table.Rows))
	for _, row := range table.Rows {
		name := cleanCell(row, idxName)
		if name == "" || name == "-" {
			continue
		}
		field := docField{Name: name}

		typeText := ""
		if idxType >= 0 {
			typeText = cleanCell(row, idxType)
		} else if idxValue >= 0 {
			typeText = cleanCell(row, idxValue)
		}
		if strings.EqualFold(section, "header") {
			typeText = "string"
		}
		field.Type = typeText

		required := false
		if strings.EqualFold(section, "path") {
			required = true
		} else if idxRequired >= 0 {
			required = parseRequired(cleanCell(row, idxRequired))
		} else if strings.EqualFold(section, "header") && strings.EqualFold(name, "authorization") {
			required = true
		}
		field.Required = required

		if idxDescription >= 0 {
			field.Description = cleanCell(row, idxDescription)
		}

		fields = append(fields, field)
	}

	return fields
}

func mergePathParamsFromPath(pathValue string, existing []docField) []docField {
	present := map[string]struct{}{}
	for _, p := range existing {
		present[p.Name] = struct{}{}
	}
	matches := pathParamPattern.FindAllStringSubmatch(pathValue, -1)
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		name := strings.TrimSpace(match[1])
		if name == "" {
			continue
		}
		if _, ok := present[name]; ok {
			continue
		}
		existing = append(existing, docField{
			Name:     name,
			Type:     "String",
			Required: true,
		})
		present[name] = struct{}{}
	}
	return existing
}

func parseKeyValueTable(table *markdownTable) map[string]string {
	kv := map[string]string{}
	if len(table.Headers) >= 2 {
		key := cleanText(table.Headers[0])
		value := strings.TrimSpace(table.Headers[1])
		if key != "" {
			kv[key] = value
		}
	}
	for _, row := range table.Rows {
		if len(row) < 2 {
			continue
		}
		key := cleanText(row[0])
		value := strings.TrimSpace(row[1])
		if key == "" {
			continue
		}
		kv[key] = value
	}
	return kv
}

func getKV(kv map[string]string, key string) string {
	for k, v := range kv {
		if k == key {
			return v
		}
	}
	return ""
}

func normalizeMethod(rawMethod string, rawURL string) (method string, original string, isWebsocket bool) {
	original = cleanText(rawMethod)
	scheme := ""
	if u, err := url.Parse(rawURL); err == nil {
		scheme = strings.ToLower(u.Scheme)
	}

	if original == "" {
		if scheme == "ws" || scheme == "wss" {
			return http.MethodGet, "WebSocket", true
		}
		return http.MethodPost, http.MethodPost, false
	}

	token := firstWordPattern.FindString(original)
	if token == "" {
		if scheme == "ws" || scheme == "wss" {
			return http.MethodGet, original, true
		}
		return http.MethodPost, original, false
	}

	switch strings.ToUpper(token) {
	case http.MethodGet:
		return http.MethodGet, original, scheme == "ws" || scheme == "wss"
	case http.MethodPost:
		return http.MethodPost, original, scheme == "ws" || scheme == "wss"
	case http.MethodPut:
		return http.MethodPut, original, scheme == "ws" || scheme == "wss"
	case http.MethodDelete:
		return http.MethodDelete, original, scheme == "ws" || scheme == "wss"
	case http.MethodPatch:
		return http.MethodPatch, original, scheme == "ws" || scheme == "wss"
	case http.MethodHead:
		return http.MethodHead, original, scheme == "ws" || scheme == "wss"
	case http.MethodOptions:
		return http.MethodOptions, original, scheme == "ws" || scheme == "wss"
	default:
		if strings.Contains(strings.ToLower(original), "websocket") || scheme == "ws" || scheme == "wss" {
			return http.MethodGet, original, true
		}
		return http.MethodPost, original, false
	}
}

func extractEndpoint(raw string) (fullURL string, serverURL string, pathValue string) {
	match := urlPattern.FindString(raw)
	if match == "" {
		return "", "", ""
	}
	match = strings.Trim(match, "`'\"")
	match = strings.TrimRight(match, ".,);")
	parsed, err := url.Parse(match)
	if err != nil {
		return "", "", ""
	}
	pathValue = parsed.EscapedPath()
	if pathValue == "" {
		pathValue = "/"
	}
	return match, parsed.Scheme + "://" + parsed.Host, pathValue
}

func parseRequired(raw string) bool {
	normalized := strings.ToLower(cleanText(raw))
	switch {
	case strings.Contains(normalized, "必选"):
		return true
	case strings.Contains(normalized, "required"):
		return true
	case normalized == "是":
		return true
	default:
		return false
	}
}

func isTableRow(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "|")
}

func isSeparatorRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	cells := splitTableRow(trimmed)
	if len(cells) == 0 {
		return false
	}
	for _, cell := range cells {
		part := strings.TrimSpace(cell)
		part = strings.Trim(part, "-")
		part = strings.Trim(part, ":")
		if strings.TrimSpace(part) != "" {
			return false
		}
	}
	return true
}

func parseTable(lines []string) *markdownTable {
	if len(lines) < 2 {
		return nil
	}
	headers := splitTableRow(lines[0])
	if len(headers) == 0 {
		return nil
	}

	rows := make([][]string, 0)
	for i := 2; i < len(lines); i++ {
		if isSeparatorRow(lines[i]) {
			continue
		}
		cells := splitTableRow(lines[i])
		if len(cells) == 0 {
			continue
		}
		if len(cells) < len(headers) {
			padding := make([]string, len(headers)-len(cells))
			cells = append(cells, padding...)
		}
		if len(cells) > len(headers) {
			cells = cells[:len(headers)]
		}
		rows = append(rows, cells)
	}
	return &markdownTable{Headers: headers, Rows: rows}
}

func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func cleanCell(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return cleanText(row[index])
}

func cleanText(raw string) string {
	if raw == "" {
		return ""
	}
	s := strings.ReplaceAll(raw, "\r", "")
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = markdownLinkRegex.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "**", "")
	s = strings.ReplaceAll(s, "*", "")
	s = htmlTagPattern.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	fields := strings.Fields(s)
	return strings.TrimSpace(strings.Join(fields, " "))
}

func sanitizeSchemaName(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	for _, ch := range name {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' {
			b.WriteRune(ch)
		}
	}
	result := b.String()
	if result == "" {
		return ""
	}
	if result[0] >= '0' && result[0] <= '9' {
		result = "Schema" + result
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

type openapiDocument struct {
	OpenAPI    string             `yaml:"openapi"`
	Info       openapiInfo        `yaml:"info"`
	Servers    []openapiServer    `yaml:"servers,omitempty"`
	Paths      map[string]pathOps `yaml:"paths"`
	Components *openapiComponents `yaml:"components,omitempty"`
}

type openapiInfo struct {
	Title       string `yaml:"title"`
	Version     string `yaml:"version"`
	Description string `yaml:"description,omitempty"`
}

type openapiServer struct {
	URL string `yaml:"url"`
}

type pathOps map[string]*openapiOperation

type openapiOperation struct {
	OperationID         string                     `yaml:"operationId"`
	Summary             string                     `yaml:"summary,omitempty"`
	Description         string                     `yaml:"description,omitempty"`
	Tags                []string                   `yaml:"tags,omitempty"`
	Parameters          []openapiParameter         `yaml:"parameters,omitempty"`
	RequestBody         *openapiRequestBody        `yaml:"requestBody,omitempty"`
	Responses           map[string]openapiResponse `yaml:"responses"`
	XCozeOriginalMethod string                     `yaml:"x-coze-original-method,omitempty"`
	XCozeTransport      string                     `yaml:"x-coze-transport,omitempty"`
	XCozePermission     string                     `yaml:"x-coze-permission,omitempty"`
	XCozeSource         string                     `yaml:"x-coze-source"`
}

type openapiParameter struct {
	Name        string        `yaml:"name"`
	In          string        `yaml:"in"`
	Description string        `yaml:"description,omitempty"`
	Required    bool          `yaml:"required,omitempty"`
	Schema      openapiSchema `yaml:"schema"`
}

type openapiRequestBody struct {
	Required bool                        `yaml:"required,omitempty"`
	Content  map[string]openapiMediaType `yaml:"content"`
}

type openapiResponse struct {
	Description string                      `yaml:"description"`
	Content     map[string]openapiMediaType `yaml:"content,omitempty"`
}

type openapiMediaType struct {
	Schema openapiSchema `yaml:"schema"`
}

type openapiComponents struct {
	Schemas map[string]openapiSchema `yaml:"schemas,omitempty"`
}

type openapiSchema struct {
	Ref                  string                   `yaml:"$ref,omitempty"`
	Type                 string                   `yaml:"type,omitempty"`
	Format               string                   `yaml:"format,omitempty"`
	Description          string                   `yaml:"description,omitempty"`
	Properties           map[string]openapiSchema `yaml:"properties,omitempty"`
	Required             []string                 `yaml:"required,omitempty"`
	Items                *openapiSchema           `yaml:"items,omitempty"`
	AdditionalProperties interface{}              `yaml:"additionalProperties,omitempty"`
}

func buildSwaggerYAML(doc apiDoc) ([]byte, error) {
	knownSchemas := map[string]string{}
	for name := range doc.ComponentSchemas {
		knownSchemas[strings.ToLower(name)] = name
	}
	unknownSchemas := map[string]struct{}{}

	components := map[string]openapiSchema{}
	for schemaName, fields := range doc.ComponentSchemas {
		properties := map[string]openapiSchema{}
		required := make([]string, 0)
		for _, field := range fields {
			if field.Name == "" {
				continue
			}
			schema := schemaFromType(field.Type, knownSchemas, unknownSchemas)
			schema.Description = field.Description
			properties[field.Name] = schema
			if field.Required {
				required = append(required, field.Name)
			}
		}
		sort.Strings(required)
		components[schemaName] = openapiSchema{
			Type:       "object",
			Properties: properties,
			Required:   required,
		}
	}
	for unknown := range unknownSchemas {
		if _, ok := components[unknown]; ok {
			continue
		}
		components[unknown] = openapiSchema{Type: "object"}
	}

	parameters := make([]openapiParameter, 0)
	for _, field := range doc.PathParams {
		parameters = append(parameters, buildParameter("path", field, knownSchemas, unknownSchemas, true))
	}
	for _, field := range doc.QueryParams {
		parameters = append(parameters, buildParameter("query", field, knownSchemas, unknownSchemas, false))
	}
	for _, field := range doc.HeaderParams {
		parameters = append(parameters, buildParameter("header", field, knownSchemas, unknownSchemas, false))
	}

	requestBody := buildRequestBody(doc.RequestBody, knownSchemas, unknownSchemas)
	responseBody := buildResponseBody(doc.ResponseBody, knownSchemas, unknownSchemas)

	descriptionParts := make([]string, 0)
	if doc.Description != "" {
		descriptionParts = append(descriptionParts, doc.Description)
	}
	if doc.InterfaceDesc != "" && doc.InterfaceDesc != doc.Description {
		descriptionParts = append(descriptionParts, doc.InterfaceDesc)
	}
	operationDescription := strings.Join(descriptionParts, "\n\n")

	operation := &openapiOperation{
		OperationID:         buildOperationID(doc.Link.Slug, doc.HTTPMethod),
		Summary:             doc.Title,
		Description:         operationDescription,
		Tags:                []string{"developer_guides"},
		Parameters:          parameters,
		RequestBody:         requestBody,
		Responses:           map[string]openapiResponse{defaultSuccessCode: responseBody},
		XCozeOriginalMethod: doc.OriginalMethod,
		XCozePermission:     doc.Permission,
		XCozeSource:         doc.Link.URL,
	}
	if doc.IsWebsocket {
		operation.XCozeTransport = "websocket"
	}

	docYAML := openapiDocument{
		OpenAPI: defaultSpecVersion,
		Info: openapiInfo{
			Title:       doc.Title,
			Version:     defaultDocVersion,
			Description: fmt.Sprintf("Generated from %s", doc.Link.URL),
		},
		Paths: map[string]pathOps{
			doc.Path: {
				strings.ToLower(doc.HTTPMethod): operation,
			},
		},
	}
	if doc.ServerURL != "" {
		docYAML.Servers = []openapiServer{{URL: doc.ServerURL}}
	}
	if len(components) > 0 {
		docYAML.Components = &openapiComponents{Schemas: components}
	}

	encoded, err := yaml.Marshal(docYAML)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func buildParameter(in string, field docField, knownSchemas map[string]string, unknownSchemas map[string]struct{}, forceRequired bool) openapiParameter {
	schema := schemaFromType(field.Type, knownSchemas, unknownSchemas)
	required := field.Required
	if forceRequired {
		required = true
	}
	return openapiParameter{
		Name:        field.Name,
		In:          in,
		Description: field.Description,
		Required:    required,
		Schema:      schema,
	}
}

func buildRequestBody(fields []docField, knownSchemas map[string]string, unknownSchemas map[string]struct{}) *openapiRequestBody {
	if len(fields) == 0 {
		return nil
	}
	properties := map[string]openapiSchema{}
	required := make([]string, 0)
	for _, field := range fields {
		if field.Name == "" {
			continue
		}
		schema := schemaFromType(field.Type, knownSchemas, unknownSchemas)
		schema.Description = field.Description
		properties[field.Name] = schema
		if field.Required {
			required = append(required, field.Name)
		}
	}
	sort.Strings(required)

	return &openapiRequestBody{
		Required: len(required) > 0,
		Content: map[string]openapiMediaType{
			defaultContentType: {
				Schema: openapiSchema{
					Type:       "object",
					Properties: properties,
					Required:   required,
				},
			},
		},
	}
}

func buildResponseBody(fields []docField, knownSchemas map[string]string, unknownSchemas map[string]struct{}) openapiResponse {
	if len(fields) == 0 {
		return openapiResponse{Description: "Success"}
	}
	properties := map[string]openapiSchema{}
	required := make([]string, 0)
	for _, field := range fields {
		if field.Name == "" {
			continue
		}
		schema := schemaFromType(field.Type, knownSchemas, unknownSchemas)
		schema.Description = field.Description
		properties[field.Name] = schema
		if field.Required {
			required = append(required, field.Name)
		}
	}
	sort.Strings(required)

	return openapiResponse{
		Description: "Success",
		Content: map[string]openapiMediaType{
			defaultContentType: {
				Schema: openapiSchema{
					Type:       "object",
					Properties: properties,
					Required:   required,
				},
			},
		},
	}
}

func schemaFromType(typeText string, knownSchemas map[string]string, unknownSchemas map[string]struct{}) openapiSchema {
	typeText = cleanText(typeText)
	if typeText == "" {
		return openapiSchema{Type: "string"}
	}

	if match := objectOfPattern.FindStringSubmatch(typeText); len(match) == 2 {
		if refName, ok := resolveSchemaRef(match[1], knownSchemas, unknownSchemas); ok {
			return openapiSchema{Ref: "#/components/schemas/" + refName}
		}
		return openapiSchema{Type: "object"}
	}
	if match := arrayOfPattern.FindStringSubmatch(typeText); len(match) == 2 {
		item := parseArrayItemSchema(match[1], knownSchemas, unknownSchemas)
		return openapiSchema{Type: "array", Items: &item}
	}

	lower := strings.ToLower(typeText)
	switch lower {
	case "string", "str":
		return openapiSchema{Type: "string"}
	case "integer", "int":
		return openapiSchema{Type: "integer", Format: "int32"}
	case "long":
		return openapiSchema{Type: "integer", Format: "int64"}
	case "double":
		return openapiSchema{Type: "number", Format: "double"}
	case "float", "number":
		return openapiSchema{Type: "number"}
	case "boolean", "bool":
		return openapiSchema{Type: "boolean"}
	case "json map", "map", "object", "json":
		return openapiSchema{Type: "object", AdditionalProperties: true}
	}

	if refName, ok := resolveSchemaRef(typeText, knownSchemas, unknownSchemas); ok {
		return openapiSchema{Ref: "#/components/schemas/" + refName}
	}
	return openapiSchema{Type: "string"}
}

func parseArrayItemSchema(typeText string, knownSchemas map[string]string, unknownSchemas map[string]struct{}) openapiSchema {
	typeText = cleanText(typeText)
	lower := strings.ToLower(typeText)
	switch lower {
	case "string", "str":
		return openapiSchema{Type: "string"}
	case "integer", "int":
		return openapiSchema{Type: "integer", Format: "int32"}
	case "long":
		return openapiSchema{Type: "integer", Format: "int64"}
	case "double":
		return openapiSchema{Type: "number", Format: "double"}
	case "float", "number":
		return openapiSchema{Type: "number"}
	case "boolean", "bool":
		return openapiSchema{Type: "boolean"}
	case "json map", "map", "object", "json":
		return openapiSchema{Type: "object", AdditionalProperties: true}
	}

	if refName, ok := resolveSchemaRef(typeText, knownSchemas, unknownSchemas); ok {
		return openapiSchema{Ref: "#/components/schemas/" + refName}
	}
	return openapiSchema{Type: "string"}
}

func resolveSchemaRef(raw string, knownSchemas map[string]string, unknownSchemas map[string]struct{}) (string, bool) {
	cleaned := cleanText(raw)
	cleaned = strings.TrimPrefix(cleaned, "[")
	cleaned = strings.TrimSuffix(cleaned, "]")
	cleaned = sanitizeSchemaName(cleaned)
	if cleaned == "" {
		return "", false
	}
	if known, ok := knownSchemas[strings.ToLower(cleaned)]; ok {
		return known, true
	}
	unknownSchemas[cleaned] = struct{}{}
	return cleaned, true
}

func buildOperationID(slug string, method string) string {
	methodLower := strings.ToLower(method)
	methodPart := methodLower
	if methodLower != "" {
		methodPart = strings.ToUpper(methodLower[:1]) + methodLower[1:]
	}
	name := sanitizeSchemaName(slug)
	if name == "" {
		name = "Operation"
	}
	return "Coze" + methodPart + name
}

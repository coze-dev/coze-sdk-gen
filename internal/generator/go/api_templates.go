package gogen

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type goAPIModuleRenderer struct {
	FileName string
	Render   func(cfg *config.Config, doc *openapi.Document) (string, error)
}

type goSwaggerModuleSpec struct {
	FileName        string
	PackageName     string
	TypeName        string
	ConstructorName string
	CoreFieldName   string
	Children        []goSwaggerModuleChild
}

type goSwaggerModuleChild struct {
	FieldName       string
	TypeName        string
	ConstructorName string
}

type goSwaggerOperationBinding struct {
	MethodName string
	HTTPMethod string
	Path       string
	Summary    string
	IsFile     bool
	Order      int
}

var goInlineAPIModuleRenderers = []goAPIModuleRenderer{
	{FileName: "apps.go", Render: renderGoAppsModule},
	{FileName: "audio_live.go", Render: renderGoAudioLiveModule},
	{FileName: "audio_speech.go", Render: renderGoAudioSpeechModule},
	{FileName: "audio_transcription.go", Render: renderGoAudioTranscriptionsModule},
	{FileName: "chats_messages.go", Render: renderGoChatsMessagesModule},
	{FileName: "files.go", Render: renderGoFilesModule},
	{FileName: "templates.go", Render: renderGoTemplatesModule},
	{FileName: "users.go", Render: renderGoUsersModule},
	{FileName: "workflows_chat.go", Render: renderGoWorkflowsChatModule},
}

var goSwaggerAPIModuleSpecs = []goSwaggerModuleSpec{}

var goCustomAPIModuleRenderers = []goAPIModuleRenderer{
	{
		FileName: "audio.go",
		Render:   renderGoAudioModule,
	},
}

var goGeneratedAPIModuleFiles = buildGoGeneratedAPIModuleFiles()

func buildGoGeneratedAPIModuleFiles() map[string]struct{} {
	files := make(map[string]struct{}, len(goCustomAPIModuleRenderers)+len(goInlineAPIModuleRenderers)+len(goSwaggerAPIModuleSpecs))
	for _, renderer := range goCustomAPIModuleRenderers {
		files[renderer.FileName] = struct{}{}
	}
	for _, renderer := range goInlineAPIModuleRenderers {
		files[renderer.FileName] = struct{}{}
	}
	for _, spec := range goSwaggerAPIModuleSpecs {
		files[spec.FileName] = struct{}{}
	}
	return files
}

func listGoAPIModuleRenderers() []goAPIModuleRenderer {
	renderers := make([]goAPIModuleRenderer, 0, len(goCustomAPIModuleRenderers)+len(goInlineAPIModuleRenderers)+len(goSwaggerAPIModuleSpecs))
	renderers = append(renderers, goCustomAPIModuleRenderers...)
	renderers = append(renderers, goInlineAPIModuleRenderers...)

	specs := append([]goSwaggerModuleSpec(nil), goSwaggerAPIModuleSpecs...)
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].FileName < specs[j].FileName
	})
	for _, spec := range specs {
		specCopy := spec
		renderers = append(renderers, goAPIModuleRenderer{
			FileName: specCopy.FileName,
			Render: func(cfg *config.Config, doc *openapi.Document) (string, error) {
				bindings := buildGoSwaggerOperationBindings(cfg, doc, specCopy.PackageName)
				return renderGoSwaggerModule(specCopy, bindings), nil
			},
		})
	}
	return renderers
}

type goAudioEnumItem struct {
	Name  string
	Value string
}

type goAudioChildDescriptor struct {
	PackageName     string
	FieldName       string
	TypeName        string
	ConstructorName string
}

var goAudioChildDescriptors = []goAudioChildDescriptor{
	{PackageName: "audio_rooms", FieldName: "Rooms", TypeName: "audioRooms", ConstructorName: "newAudioRooms"},
	{PackageName: "audio_speech", FieldName: "Speech", TypeName: "audioSpeech", ConstructorName: "newAudioSpeech"},
	{PackageName: "audio_voices", FieldName: "Voices", TypeName: "audioVoices", ConstructorName: "newAudioVoices"},
	{PackageName: "audio_transcriptions", FieldName: "Transcriptions", TypeName: "audioTranscriptions", ConstructorName: "newAudioTranscriptions"},
	{PackageName: "audio_voiceprint_groups", FieldName: "VoiceprintGroups", TypeName: "audioVoiceprintGroups", ConstructorName: "newAudioVoiceprintGroups"},
	{PackageName: "audio_live", FieldName: "Live", TypeName: "audioLive", ConstructorName: "newAudioLive"},
}

var goAudioFormatFallbackValues = []string{"wav", "pcm", "ogg_opus", "m4a", "aac", "mp3"}

var goLanguageCodeFallbackValues = []string{"zh", "en", "ja", "es", "id", "pt"}

func renderGoAudioModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	audioFormats := buildGoAudioEnumItems(
		cfg,
		doc,
		"audio_speech",
		"AudioFormat",
		[]string{"response_format", "audio_format"},
		goAudioFormatFallbackValues,
	)
	audioFormats = orderGoAudioEnumItemsByValue(audioFormats, goAudioFormatFallbackValues)
	languageCodes := buildGoAudioEnumItems(
		cfg,
		doc,
		"audio_speech",
		"LanguageCode",
		[]string{"language_code", "language"},
		goLanguageCodeFallbackValues,
	)
	languageCodes = orderGoAudioEnumItemsByValue(languageCodes, goLanguageCodeFallbackValues)
	children := buildGoAudioChildren(cfg)

	var buf bytes.Buffer
	buf.WriteString("package coze\n\n")
	writeGoAudioEnum(&buf, "AudioFormat", "audio format type", "f", audioFormats, true)
	buf.WriteString("\n")
	writeGoAudioEnum(&buf, "LanguageCode", "language code", "l", languageCodes, false)
	buf.WriteString("\n")
	buf.WriteString("type audio struct {\n")
	for _, child := range children {
		buf.WriteString(fmt.Sprintf("\t%s *%s\n", child.FieldName, child.TypeName))
	}
	buf.WriteString("}\n\n")
	buf.WriteString("func newAudio(core *core) *audio {\n")
	if len(children) == 0 {
		buf.WriteString("\treturn &audio{}\n")
		buf.WriteString("}\n")
		return buf.String(), nil
	}
	buf.WriteString("\treturn &audio{\n")
	for _, child := range children {
		buf.WriteString(fmt.Sprintf("\t\t%s: %s(core),\n", child.FieldName, child.ConstructorName))
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String(), nil
}

func writeGoAudioEnum(
	buf *bytes.Buffer,
	typeName string,
	comment string,
	receiver string,
	items []goAudioEnumItem,
	withPtr bool,
) {
	if buf == nil || strings.TrimSpace(typeName) == "" {
		return
	}
	items = normalizeGoAudioEnumItems(items)
	buf.WriteString(fmt.Sprintf("// %s represents the %s\n", typeName, comment))
	buf.WriteString(fmt.Sprintf("type %s string\n\n", typeName))
	buf.WriteString("const (\n")
	for _, item := range items {
		buf.WriteString(fmt.Sprintf("\t%s%s %s = %q\n", typeName, item.Name, typeName, item.Value))
	}
	buf.WriteString(")\n\n")
	receiver = strings.TrimSpace(receiver)
	if receiver == "" {
		receiver = "v"
	}
	buf.WriteString(fmt.Sprintf("func (%s %s) String() string {\n", receiver, typeName))
	buf.WriteString(fmt.Sprintf("\treturn string(%s)\n", receiver))
	buf.WriteString("}\n")
	if withPtr {
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("func (%s %s) Ptr() *%s {\n", receiver, typeName, typeName))
		buf.WriteString(fmt.Sprintf("\treturn &%s\n", receiver))
		buf.WriteString("}\n")
	}
}

func buildGoAudioEnumItems(
	cfg *config.Config,
	doc *openapi.Document,
	packageName string,
	modelName string,
	swaggerPropertyNames []string,
	fallbackValues []string,
) []goAudioEnumItem {
	fromConfig := collectGoEnumItemsFromConfig(cfg, packageName, modelName)
	fromSwagger := collectGoEnumItemsFromSwagger(doc, swaggerPropertyNames)
	merged := mergeGoAudioEnumItems(fromConfig, fromSwagger)
	merged = mergeGoAudioEnumItems(merged, enumItemsFromValues(fallbackValues))
	if len(merged) == 0 {
		return enumItemsFromValues(fallbackValues)
	}
	return merged
}

func buildGoAudioChildren(cfg *config.Config) []goAudioChildDescriptor {
	available := map[string]struct{}{}
	if cfg != nil {
		for _, pkg := range cfg.API.Packages {
			name := strings.TrimSpace(pkg.Name)
			if name == "" {
				continue
			}
			available[name] = struct{}{}
		}
	}
	children := make([]goAudioChildDescriptor, 0, len(goAudioChildDescriptors))
	for _, descriptor := range goAudioChildDescriptors {
		if len(available) > 0 {
			if _, ok := available[descriptor.PackageName]; !ok {
				continue
			}
		}
		children = append(children, descriptor)
	}
	return children
}

func collectGoEnumItemsFromConfig(cfg *config.Config, packageName string, modelName string) []goAudioEnumItem {
	if cfg == nil {
		return nil
	}
	pkgName := strings.TrimSpace(packageName)
	model := strings.TrimSpace(modelName)
	if pkgName == "" || model == "" {
		return nil
	}
	items := make([]goAudioEnumItem, 0)
	for _, pkg := range cfg.API.Packages {
		if strings.TrimSpace(pkg.Name) != pkgName {
			continue
		}
		for _, modelSchema := range pkg.ModelSchemas {
			if strings.TrimSpace(modelSchema.Name) != model {
				continue
			}
			for _, enumValue := range modelSchema.EnumValues {
				value := strings.TrimSpace(fmt.Sprint(enumValue.Value))
				value = strings.Trim(value, "\"")
				if value == "" {
					continue
				}
				name := strings.TrimSpace(enumValue.Name)
				if name == "" {
					name = value
				}
				items = append(items, goAudioEnumItem{
					Name:  strings.ToUpper(normalizeGoExportedIdentifier(name)),
					Value: value,
				})
			}
			return normalizeGoAudioEnumItems(items)
		}
	}
	return normalizeGoAudioEnumItems(items)
}

func collectGoEnumItemsFromSwagger(doc *openapi.Document, propertyNames []string) []goAudioEnumItem {
	if doc == nil || len(propertyNames) == 0 {
		return nil
	}
	targets := map[string]struct{}{}
	for _, propertyName := range propertyNames {
		name := strings.ToLower(strings.TrimSpace(propertyName))
		if name == "" {
			continue
		}
		targets[name] = struct{}{}
	}
	if len(targets) == 0 {
		return nil
	}
	values := make([]string, 0, 16)
	for _, details := range doc.ListOperationDetails() {
		if !strings.HasPrefix(strings.TrimSpace(details.Path), "/v1/audio") {
			continue
		}
		for _, param := range details.Parameters {
			if _, ok := targets[strings.ToLower(strings.TrimSpace(param.Name))]; !ok {
				continue
			}
			appendGoSchemaEnumValues(param.Schema, &values)
		}
		collectGoSchemaPropertyEnums(doc, details.RequestBodySchema, targets, &values, map[*openapi.Schema]struct{}{})
		collectGoSchemaPropertyEnums(doc, details.ResponseSchema, targets, &values, map[*openapi.Schema]struct{}{})
	}
	return enumItemsFromValues(values)
}

func collectGoSchemaPropertyEnums(
	doc *openapi.Document,
	schema *openapi.Schema,
	targets map[string]struct{},
	values *[]string,
	visited map[*openapi.Schema]struct{},
) {
	if doc == nil || schema == nil || values == nil {
		return
	}
	resolved := doc.ResolveSchema(schema)
	if resolved == nil {
		return
	}
	if visited == nil {
		visited = map[*openapi.Schema]struct{}{}
	}
	if _, seen := visited[resolved]; seen {
		return
	}
	visited[resolved] = struct{}{}

	propertyNames := make([]string, 0, len(resolved.Properties))
	for propertyName := range resolved.Properties {
		propertyNames = append(propertyNames, propertyName)
	}
	sort.Strings(propertyNames)
	for _, propertyName := range propertyNames {
		propertySchema := resolved.Properties[propertyName]
		if propertySchema == nil {
			continue
		}
		if _, ok := targets[strings.ToLower(strings.TrimSpace(propertyName))]; ok {
			appendGoSchemaEnumValues(propertySchema, values)
		}
		collectGoSchemaPropertyEnums(doc, propertySchema, targets, values, visited)
	}
	collectGoSchemaPropertyEnums(doc, resolved.Items, targets, values, visited)
	for _, item := range resolved.AllOf {
		collectGoSchemaPropertyEnums(doc, item, targets, values, visited)
	}
	for _, item := range resolved.AnyOf {
		collectGoSchemaPropertyEnums(doc, item, targets, values, visited)
	}
	for _, item := range resolved.OneOf {
		collectGoSchemaPropertyEnums(doc, item, targets, values, visited)
	}
	if additional, ok := resolved.AdditionalProperties.(*openapi.Schema); ok {
		collectGoSchemaPropertyEnums(doc, additional, targets, values, visited)
	}
}

func appendGoSchemaEnumValues(schema *openapi.Schema, values *[]string) {
	if schema == nil || values == nil {
		return
	}
	for _, raw := range schema.Enum {
		value := strings.TrimSpace(fmt.Sprint(raw))
		value = strings.Trim(value, "\"")
		if value == "" {
			continue
		}
		*values = append(*values, value)
	}
}

func enumItemsFromValues(values []string) []goAudioEnumItem {
	items := make([]goAudioEnumItem, 0, len(values))
	for _, value := range values {
		cleanValue := strings.TrimSpace(value)
		if cleanValue == "" {
			continue
		}
		items = append(items, goAudioEnumItem{
			Name:  strings.ToUpper(normalizeGoExportedIdentifier(cleanValue)),
			Value: cleanValue,
		})
	}
	return normalizeGoAudioEnumItems(items)
}

func mergeGoAudioEnumItems(base []goAudioEnumItem, extra []goAudioEnumItem) []goAudioEnumItem {
	merged := make([]goAudioEnumItem, 0, len(base)+len(extra))
	seen := map[string]struct{}{}
	appendItems := func(items []goAudioEnumItem) {
		for _, item := range normalizeGoAudioEnumItems(items) {
			if _, ok := seen[item.Value]; ok {
				continue
			}
			seen[item.Value] = struct{}{}
			merged = append(merged, item)
		}
	}
	appendItems(base)
	appendItems(extra)
	return merged
}

func normalizeGoAudioEnumItems(items []goAudioEnumItem) []goAudioEnumItem {
	normalized := make([]goAudioEnumItem, 0, len(items))
	seenValues := map[string]struct{}{}
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		value := strings.TrimSpace(item.Value)
		if value == "" {
			continue
		}
		if name == "" {
			name = strings.ToUpper(normalizeGoExportedIdentifier(value))
		}
		if name == "" {
			continue
		}
		if _, exists := seenValues[value]; exists {
			continue
		}
		seenValues[value] = struct{}{}
		normalized = append(normalized, goAudioEnumItem{
			Name:  name,
			Value: value,
		})
	}
	return normalized
}

func orderGoAudioEnumItemsByValue(items []goAudioEnumItem, orderValues []string) []goAudioEnumItem {
	if len(items) == 0 || len(orderValues) == 0 {
		return items
	}
	valueOrder := make(map[string]int, len(orderValues))
	for idx, value := range orderValues {
		valueOrder[strings.TrimSpace(value)] = idx
	}
	ordered := make([]goAudioEnumItem, 0, len(items))
	seen := map[string]struct{}{}
	appendValue := func(value string) {
		if _, ok := seen[value]; ok {
			return
		}
		for _, item := range items {
			if item.Value != value {
				continue
			}
			ordered = append(ordered, item)
			seen[value] = struct{}{}
			return
		}
	}
	for _, value := range orderValues {
		appendValue(strings.TrimSpace(value))
	}
	for _, item := range items {
		value := item.Value
		if _, preferred := valueOrder[value]; preferred {
			appendValue(value)
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		ordered = append(ordered, item)
		seen[value] = struct{}{}
	}
	return ordered
}

func shouldSkipGoExtraAsset(rel string) bool {
	target := strings.TrimSuffix(rel, ".tpl")
	if target == "README.md" || target == "main" || strings.HasPrefix(target, ".github/") {
		return true
	}
	_, ok := goGeneratedAPIModuleFiles[target]
	return ok
}

func buildGoSwaggerOperationBindings(cfg *config.Config, doc *openapi.Document, packageName string) []goSwaggerOperationBinding {
	if cfg == nil {
		return nil
	}
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return nil
	}

	bindings := make([]goSwaggerOperationBinding, 0)
	for _, mapping := range cfg.API.OperationMappings {
		if cfg.IsIgnored(mapping.Path, mapping.Method) {
			continue
		}
		details, hasDetails := resolveGoOperationDetails(doc, mapping)
		for methodIndex, sdkMethod := range mapping.SDKMethods {
			pkg, method, ok := parseGoSDKMethod(cfg, mapping.Path, sdkMethod)
			if !ok || pkg != packageName {
				continue
			}
			goMethod := normalizeGoExportedIdentifier(method)
			if goMethod == "" {
				continue
			}
			httpMethod := strings.TrimSpace(mapping.Method)
			if httpMethod == "" {
				httpMethod = http.MethodGet
			}

			isFile := false
			summary := ""
			if hasDetails {
				summary = goOperationSummary(details)
				contentType := strings.ToLower(strings.TrimSpace(details.RequestBodyContentType))
				isFile = strings.Contains(contentType, "multipart/form-data")
			}

			order := len(bindings)
			if mapping.Order > 0 {
				order = mapping.Order + methodIndex
			}
			bindings = append(bindings, goSwaggerOperationBinding{
				MethodName: goMethod,
				HTTPMethod: strings.ToUpper(httpMethod),
				Path:       strings.TrimSpace(mapping.Path),
				Summary:    summary,
				IsFile:     isFile,
				Order:      order,
			})
		}
	}

	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Order == bindings[j].Order {
			if bindings[i].MethodName == bindings[j].MethodName {
				if bindings[i].HTTPMethod == bindings[j].HTTPMethod {
					return bindings[i].Path < bindings[j].Path
				}
				return bindings[i].HTTPMethod < bindings[j].HTTPMethod
			}
			return bindings[i].MethodName < bindings[j].MethodName
		}
		return bindings[i].Order < bindings[j].Order
	})

	seen := map[string]int{}
	for i := range bindings {
		name := bindings[i].MethodName
		count := seen[name]
		if count > 0 {
			bindings[i].MethodName = fmt.Sprintf("%s%d", name, count+1)
		}
		seen[name] = count + 1
	}
	return bindings
}

func resolveGoOperationDetails(doc *openapi.Document, mapping config.OperationMapping) (openapi.OperationDetails, bool) {
	if doc == nil {
		return openapi.OperationDetails{}, false
	}
	details, ok := doc.OperationDetails(strings.TrimSpace(mapping.Path), strings.TrimSpace(mapping.Method))
	if !ok || details == nil {
		return openapi.OperationDetails{}, false
	}
	return *details, true
}

func parseGoSDKMethod(cfg *config.Config, path string, sdkMethod string) (string, string, bool) {
	sdkMethod = strings.TrimSpace(sdkMethod)
	if sdkMethod == "" {
		return "", "", false
	}
	parts := strings.Split(sdkMethod, ".")
	switch len(parts) {
	case 1:
		method := strings.TrimSpace(parts[0])
		if method == "" || cfg == nil {
			return "", "", false
		}
		pkg, ok := cfg.ResolvePackage(path, "")
		if !ok {
			return "", "", false
		}
		return strings.TrimSpace(pkg.Name), method, true
	case 2:
		pkg := strings.TrimSpace(parts[0])
		method := strings.TrimSpace(parts[1])
		if pkg == "" || method == "" {
			return "", "", false
		}
		return pkg, method, true
	case 3:
		if strings.TrimSpace(parts[0]) != "go" {
			return "", "", false
		}
		pkg := strings.TrimSpace(parts[1])
		method := strings.TrimSpace(parts[2])
		if pkg == "" || method == "" {
			return "", "", false
		}
		return pkg, method, true
	default:
		return "", "", false
	}
}

func renderGoSwaggerModule(spec goSwaggerModuleSpec, bindings []goSwaggerOperationBinding) string {
	coreField := strings.TrimSpace(spec.CoreFieldName)
	if coreField == "" {
		coreField = "core"
	}

	var buf bytes.Buffer
	buf.WriteString("package coze\n\n")
	if len(bindings) > 0 {
		buf.WriteString("import (\n")
		buf.WriteString("\t\"context\"\n")
		buf.WriteString("\t\"net/http\"\n")
		buf.WriteString(")\n\n")
	}

	for _, binding := range bindings {
		if summary := strings.TrimSpace(binding.Summary); summary != "" {
			buf.WriteString(fmt.Sprintf("// %s %s\n", binding.MethodName, summary))
		}
		buf.WriteString(fmt.Sprintf("func (r *%s) %s(ctx context.Context, req *SwaggerOperationRequest) (*SwaggerOperationResponse, error) {\n", spec.TypeName, binding.MethodName))
		buf.WriteString("\tif req == nil {\n")
		buf.WriteString("\t\treq = &SwaggerOperationRequest{}\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\trequest := &RawRequestReq{\n")
		buf.WriteString(fmt.Sprintf("\t\tMethod: %s,\n", goHTTPMethodConstant(binding.HTTPMethod)))
		buf.WriteString(fmt.Sprintf("\t\tURL:    buildSwaggerOperationURL(%q, req.PathParams, req.QueryParams),\n", binding.Path))
		buf.WriteString("\t\tBody:   req.Body,\n")
		if binding.IsFile {
			buf.WriteString("\t\tIsFile: true,\n")
		}
		buf.WriteString("\t}\n")
		buf.WriteString("\tresponse := new(SwaggerOperationResponse)\n")
		buf.WriteString(fmt.Sprintf("\terr := r.%s.rawRequest(ctx, request, response)\n", coreField))
		buf.WriteString("\treturn response, err\n")
		buf.WriteString("}\n\n")
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", spec.TypeName))
	buf.WriteString(fmt.Sprintf("\t%s *core\n", coreField))
	for _, child := range spec.Children {
		buf.WriteString(fmt.Sprintf("\t%s *%s\n", child.FieldName, child.TypeName))
	}
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func %s(core *core) *%s {\n", spec.ConstructorName, spec.TypeName))
	if len(spec.Children) == 0 {
		buf.WriteString(fmt.Sprintf("\treturn &%s{%s: core}\n", spec.TypeName, coreField))
		buf.WriteString("}\n")
		return buf.String()
	}

	buf.WriteString(fmt.Sprintf("\treturn &%s{\n", spec.TypeName))
	buf.WriteString(fmt.Sprintf("\t\t%s: core,\n", coreField))
	for _, child := range spec.Children {
		buf.WriteString(fmt.Sprintf("\t\t%s: %s(core),\n", child.FieldName, child.ConstructorName))
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String()
}

func goHTTPMethodConstant(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "http.MethodGet"
	case http.MethodPost:
		return "http.MethodPost"
	case http.MethodPut:
		return "http.MethodPut"
	case http.MethodDelete:
		return "http.MethodDelete"
	case http.MethodPatch:
		return "http.MethodPatch"
	case http.MethodHead:
		return "http.MethodHead"
	case http.MethodOptions:
		return "http.MethodOptions"
	case http.MethodTrace:
		return "http.MethodTrace"
	default:
		return fmt.Sprintf("%q", strings.ToUpper(strings.TrimSpace(method)))
	}
}

package python

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"

	"github.com/coze-dev/coze-sdk-gen/parser"
	"gopkg.in/yaml.v3"
)

//go:embed templates/sdk.tmpl
var templateFS embed.FS

//go:embed config.yaml
var configFS embed.FS

type PagedOperationConfig struct {
	Enabled      bool              `yaml:"enabled"`
	ParamMapping map[string]string `yaml:"param_mapping"`
	// esponseClass string            `yaml:"response_class"`
	ItemType string `yaml:"item_type"`
}

type ModuleConfig struct {
	EnumNameMapping           map[string]string               `yaml:"enum_name_mapping"`
	OperationNameMapping      map[string]string               `yaml:"operation_name_mapping"`
	ResponseTypeModify        map[string]string               `yaml:"response_type_modify"`
	TypeMapping               map[string]string               `yaml:"type_mapping"`
	SkipOptionalFieldsClasses []string                        `yaml:"skip_optional_fields_classes"`
	PagedOperations           map[string]PagedOperationConfig `yaml:"paged_operations"`
}

type Config struct {
	Modules map[string]ModuleConfig `yaml:"modules"`
}

// Generator handles Python SDK generation using parser2
type Generator struct {
	classes    []PythonClass
	config     Config
	moduleName string
}

// pythonTypeMapping maps our types to Python types
var pythonTypeMapping = map[parser.PrimitiveKind]string{
	parser.PrimitiveString:  "str",
	parser.PrimitiveInt:     "int",
	parser.PrimitiveFloat:   "float",
	parser.PrimitiveBool:    "bool",
	parser.PrimitiveBinary:  "bytes",
	parser.PrimitiveUnknown: "Any",
}

// PythonClass represents a Python class
type PythonClass struct {
	Name        string
	Description string
	Fields      []PythonField
	Methods     []string
	BaseClass   string
	IsEnum      bool
	EnumValues  []PythonEnumValue
	ShouldSkip  bool
	IsPass      bool
}

// PythonEnumValue represents a Python enum value
type PythonEnumValue struct {
	Name        string
	Value       string
	Description string
}

// PythonField represents a Python class field
type PythonField struct {
	Name        string
	Type        string
	Description string
	IsMethod    bool
	Default     string
}

// PythonOperation represents a Python API operation
type PythonOperation struct {
	Name                string
	Description         string
	Path                string
	Method              string
	Params              []PythonParam
	BodyParams          []PythonParam
	QueryParams         []PythonParam
	ResponseType        string
	ResponseDescription string
	HasBody             bool
	HasQueryParams      bool
	ModuleName          string
	HeaderParams        []PythonParam
	HasHeaders          bool
	StaticHeaders       map[string]string
	// page
	IsPaged           bool
	ResponseCast      string
	AsyncResponseType string
	PageIndexName     string
	PageSizeName      string
	HasFileUpload     bool
}

// PythonParam represents a Python parameter
type PythonParam struct {
	Name         string
	JsonName     string
	Type         string
	Description  string
	DefaultValue string
	HasDefault   bool
	IsModel      bool
}

// PythonModule represents a converted Python module
type PythonModule struct {
	Operations    []PythonOperation
	Classes       []PythonClass
	HasFileUpload bool
}

func (g *Generator) loadConfig() error {
	configData, err := configFS.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to read config.yaml: %w", err)
	}

	if err := yaml.Unmarshal(configData, &g.config); err != nil {
		return fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	return nil
}

// Generate generates Python SDK code from parsed OpenAPI data
func (g *Generator) Generate(ctx context.Context, yamlContent []byte) (map[string]string, error) {
	// Load config first
	if err := g.loadConfig(); err != nil {
		return nil, err
	}

	// Create new parser2 instance
	p, err := parser.NewParser(&parser.ModuleConfig{
		GenerateUnnamedResponseType: func(h *parser.HttpHandler) (string, bool) {
			if h.GetActualResponseBody() == nil {
				return fmt.Sprintf("%sResp", h.Name), true
			}

			return "", false
		},
		ChangeHttpHandlerResponseType: map[string]string{
			"CreateDraftBot":  "Bot",
			"UpdateDraftBot":  "Bot",
			"PublishDraftBot": "Bot",
		},
		RenameTypes: map[string]string{
			"SpacePublishedBotsInfo": "_PrivateListBotsData",
		},
		RenameHandlers: map[string]string{
			"RetrieveFileOpen": "retrieve",
			"UploadFileOpen":   "upload",
		},
		ChangeFields: map[string]map[string]*parser.FieldModification{
			"File": {
				"id": {
					Requirement: parser.FieldRequirementRequired,
				},
			},
		},
		HandlerOrdering: map[string][]string{
			"files": {"UploadFileOpen", "RetrieveFileOpen"},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create parser2 failed: %w", err)
	}

	// Parse OpenAPI spec
	modules, err := p.ParseOpenAPI(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI failed: %w", err)
	}

	// Generate code for each module
	files := make(map[string]string)

	// Read template
	tmpl, err := template.New("python").Funcs(template.FuncMap{
		"title": func(x string) string {
			return strings.ReplaceAll(strings.Title(x), ".", "")
		},
	}).Parse(g.getTemplate())
	if err != nil {
		return nil, fmt.Errorf("parse template failed: %w", err)
	}

	// Convert modules to Python-specific format
	for moduleName, module := range modules {
		pythonModule := g.convertModule(module)
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]interface{}{
			"ModuleName":    moduleName,
			"Operations":    pythonModule.Operations,
			"Classes":       pythonModule.Classes,
			"HasFileUpload": pythonModule.HasFileUpload,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template failed: %w", err)
		}
		files[fmt.Sprintf("%s", moduleName)] = buf.String()
	}

	return files, nil
}

func (g *Generator) convertModule(module *parser.Module) PythonModule {
	// Store current module name
	g.moduleName = module.Name

	// Convert types to classes
	classes := make([]PythonClass, 0)
	for _, ty := range module.Types {
		if pythonClass := g.convertType(ty); pythonClass != nil {
			classes = append(classes, *pythonClass)
		}
	}
	g.classes = classes

	// Convert operations
	operations := make([]PythonOperation, 0)
	hasFileUpload := false
	for _, handler := range module.HttpHandlers {
		if op := g.convertHandler(&handler); op != nil {
			operations = append(operations, *op)
			if op.HasFileUpload {
				hasFileUpload = true
			}
		}
	}

	return PythonModule{
		Operations:    operations,
		Classes:       classes,
		HasFileUpload: hasFileUpload,
	}
}

func (g *Generator) convertType(ty *parser.Ty) *PythonClass {
	if !ty.IsNamed {
		return nil
	}

	// Apply type mapping if exists
	if g.config.Modules[g.moduleName].TypeMapping[ty.Name] != "" {
		ty.Name = g.config.Modules[g.moduleName].TypeMapping[ty.Name]
	}

	pythonClass := &PythonClass{
		Name:        ty.Name,
		Description: g.formatDescription(ty.Description),
		BaseClass:   "CozeModel",
	}

	// Handle enums
	if len(ty.EnumValues) > 0 {
		pythonClass.IsEnum = true
		pythonClass.BaseClass = "IntEnum"
		for _, value := range ty.EnumValues {
			pythonClass.EnumValues = append(pythonClass.EnumValues, PythonEnumValue{
				Name:  g.toEnumName(value.Name),
				Value: fmt.Sprintf("%v", value.Val),
			})
		}
		return pythonClass
	}

	// Skip optional fields for configured classes
	skipOptionalFields := false
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		for _, skipClass := range moduleConfig.SkipOptionalFieldsClasses {
			if skipClass == ty.Name {
				skipOptionalFields = true
				break
			}
		}
	}

	// Convert fields
	for _, field := range ty.Fields {
		fieldType := g.getFieldType(field.Type)
		if !field.Required && !skipOptionalFields {
			fieldType = fmt.Sprintf("Optional[%s]", fieldType)
		}

		pythonField := PythonField{
			Name:        g.toPythonVarName(field.Name),
			Type:        fieldType,
			Description: g.formatDescription(field.Description),
			Default:     field.Default,
		}
		if pythonField.Default == "" && !field.Required {
			pythonField.Default = "None"
		}
		pythonClass.Fields = append(pythonClass.Fields, pythonField)
	}

	if ty.HasOnlyStatusFields() {
		pythonClass.IsPass = true
	}

	return pythonClass
}

func removeOptional(t string) string {
	if strings.HasPrefix(t, "Optional[") && strings.HasSuffix(t, "]") {
		return t[9 : len(t)-1]
	}
	return t
}

func marshal(v any) string {
	res, _ := json.Marshal(v)
	return string(res)
}

func (g *Generator) convertHandler(handler *parser.HttpHandler) *PythonOperation {
	operation := &PythonOperation{
		Name:        g.toPythonMethodName(handler.Name),
		Description: handler.Description,
		Path:        handler.Path,
		Method:      strings.ToUpper(handler.Method),
	}

	// Convert parameters
	var headerParams []PythonParam
	staticHeaders := make(map[string]string)

	// Handle path parameters
	for _, param := range handler.PathParams {
		pythonParam := g.convertParam(&param)
		operation.Params = append(operation.Params, pythonParam)
	}

	// Handle query parameters
	for _, param := range handler.QueryParams {
		pythonParam := g.convertParam(&param)
		operation.QueryParams = append(operation.QueryParams, pythonParam)
		operation.Params = append(operation.Params, pythonParam)
		operation.HasQueryParams = true
	}

	// Handle header parameters
	for _, param := range handler.HeaderParams {
		pythonParam := g.convertParam(&param)
		headerParams = append(headerParams, pythonParam)
		operation.Params = append(operation.Params, pythonParam)
	}

	// Handle request body
	if handler.RequestBody != nil {
		operation.HasBody = true
		switch handler.ContentType {
		case parser.ContentTypeFile:
			operation.HasFileUpload = true
			for _, field := range handler.RequestBody.Fields {
				pythonParam := g.convertParam(&field)
				if field.Type.PrimitiveKind == parser.PrimitiveBinary {
					pythonParam.Type = "FileTypes"
				}
				operation.BodyParams = append(operation.BodyParams, pythonParam)
				operation.Params = append(operation.Params, pythonParam)
			}
		case parser.ContentTypeJson:
			for _, field := range handler.RequestBody.Fields {
				pythonParam := g.convertParam(&field)
				operation.BodyParams = append(operation.BodyParams, pythonParam)
				operation.Params = append(operation.Params, pythonParam)
			}
		default:
			panic(fmt.Sprintf("unsupported content type %q", handler.ContentType))
		}
	}

	// Handle response body using GetActualResponseBody
	if actualBody := handler.GetActualResponseBody(); actualBody != nil {
		operation.ResponseType = g.getFieldType(actualBody)
	} else if handler.ResponseBody != nil {
		operation.ResponseType = g.getFieldType(handler.ResponseBody)
	}

	// Check if this is a paged operation using GetPageInfo
	if pageInfo := handler.GetPageInfo(nil, nil); pageInfo != nil {
		operation.IsPaged = true
		if b := handler.GetActualResponseBody(); b != nil {
			operation.ResponseCast = g.getFieldType(handler.GetActualResponseBody())
		} else {
			operation.ResponseCast = "unknown_paged_response"
		}
		operation.ResponseType = fmt.Sprintf("NumberPaged[%s]", pageInfo.ItemType.Name)
		operation.AsyncResponseType = fmt.Sprintf("AsyncNumberPaged[%s]", pageInfo.ItemType.Name)
		operation.PageIndexName = g.toPythonVarName(pageInfo.PageIndexName)
		operation.PageSizeName = g.toPythonVarName(pageInfo.PageSizeName)

		for i, param := range operation.Params {
			if param.Name == operation.PageIndexName {
				operation.Params[i].DefaultValue = "1"
				operation.Params[i].Type = removeOptional(operation.Params[i].Type)
			}
			if param.Name == operation.PageSizeName {
				operation.Params[i].DefaultValue = "20"
				operation.Params[i].Type = removeOptional(operation.Params[i].Type)
			}
		}
	}

	// Update headers
	if len(headerParams) > 0 || len(staticHeaders) > 0 {
		operation.HeaderParams = headerParams
		operation.StaticHeaders = staticHeaders
		operation.HasHeaders = true
	}

	return operation
}

func (g *Generator) convertParam(field *parser.TyField) PythonParam {
	fieldType := g.getFieldType(field.Type)
	if !field.Required {
		fieldType = fmt.Sprintf("Optional[%s]", fieldType)
	}

	param := PythonParam{
		Name:        g.toPythonVarName(field.Name),
		JsonName:    field.Name,
		Type:        fieldType,
		Description: field.Description,
		IsModel:     field.Type.IsNamed,
	}

	if !field.Required {
		param.DefaultValue = "None"
		param.HasDefault = true
	}

	return param
}

func (g *Generator) getFieldType(ty *parser.Ty) string {
	if ty == nil {
		return "Any"
	}

	switch ty.Kind {
	case parser.TyKindPrimitive:
		if pyType, ok := pythonTypeMapping[ty.PrimitiveKind]; ok {
			return pyType
		}
		return "Any"

	case parser.TyKindArray:
		if ty.ElementType != nil {
			elemType := g.getFieldType(ty.ElementType)
			return fmt.Sprintf("List[%s]", elemType)
		}
		return "List[Any]"

	case parser.TyKindMap:
		if ty.ValueType != nil {
			valueType := g.getFieldType(ty.ValueType)
			return fmt.Sprintf("Dict[str, %s]", valueType)
		}
		return "Dict[str, Any]"

	case parser.TyKindObject:
		if ty.IsNamed {
			return ty.Name
		}
		return "Dict[str, Any]"

	default:
		return "Any"
	}
}

func (g *Generator) formatDescription(desc string) string {
	if desc == "" {
		return desc
	}
	// Remove escape characters
	desc = strings.ReplaceAll(desc, "\\", "")
	// Convert consecutive newlines to single newline
	desc = regexp.MustCompile(`\n\s*\n+`).ReplaceAllString(desc, "\n")
	// Add indentation after each newline
	desc = regexp.MustCompile(`\n`).ReplaceAllString(desc, "\n    ")
	// Trim leading/trailing whitespace
	desc = strings.TrimSpace(desc)
	return desc
}

func (g *Generator) toPythonMethodName(name string) string {
	// First check if there's a mapping in the module-specific config
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		if mappedName, ok := moduleConfig.OperationNameMapping[name]; ok {
			return mappedName
		}
	}

	// If no mapping found, use the default conversion logic
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func (g *Generator) toPythonVarName(name string) string {
	// Replace any non-alphanumeric characters with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	name = reg.ReplaceAllString(name, "_")

	// Add underscore before capital letters (camelCase to snake_case)
	reg = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	name = reg.ReplaceAllString(name, "${1}_${2}")

	// Convert to lowercase
	name = strings.ToLower(name)

	// Remove consecutive underscores
	reg = regexp.MustCompile(`_+`)
	name = reg.ReplaceAllString(name, "_")

	// Trim leading and trailing underscores
	name = strings.Trim(name, "_")

	// If empty or starts with a number, prefix with underscore
	if name == "" || regexp.MustCompile(`^[0-9]`).MatchString(name) {
		name = "_" + name
	}

	return name
}

func (g *Generator) toEnumName(name string) string {
	// First check if there's a mapping in the module-specific config
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		if mappedName, ok := moduleConfig.EnumNameMapping[name]; ok {
			return mappedName
		}
	}

	// Check if the name is already in uppercase with underscores format
	isUpperWithUnderscores := true
	for i, r := range name {
		if i > 0 && r == '_' {
			continue
		}
		if r < 'A' || r > 'Z' {
			isUpperWithUnderscores = false
			break
		}
	}
	if isUpperWithUnderscores {
		return name
	}

	// If no mapping found and not already in correct format, use the default conversion logic
	// First convert camelCase to snake_case
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	snakeCase := strings.ToLower(result.String())

	// Replace any non-alphanumeric characters with underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	name = reg.ReplaceAllString(snakeCase, "_")

	// Remove consecutive underscores
	reg = regexp.MustCompile(`_+`)
	name = reg.ReplaceAllString(name, "_")

	// Trim leading and trailing underscores and convert to uppercase
	return strings.ToUpper(strings.Trim(name, "_"))
}

func (g *Generator) getTemplate() string {
	// Read template from embedded file
	templateContent, err := fs.ReadFile(templateFS, "templates/sdk.tmpl")
	if err != nil {
		return ""
	}
	return string(templateContent)
}

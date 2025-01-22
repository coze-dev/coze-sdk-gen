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
)

//go:embed templates/sdk.tmpl
var templateFS embed.FS

//go:embed config.json
var configFS embed.FS

type PagedOperationConfig struct {
	Enabled         bool              `json:"enabled"`
	ParamMapping    map[string]string `json:"param_mapping"`
	ResponseClass   string            `json:"response_class"`
	ReturnType      string            `json:"return_type"`
	AsyncReturnType string            `json:"async_return_type"`
}

type ModuleConfig struct {
	EnumNameMapping           map[string]string               `json:"enum_name_mapping"`
	OperationNameMapping      map[string]string               `json:"operation_name_mapping"`
	ResponseTypeMapping       map[string]string               `json:"response_type_mapping"`
	TypeMapping               map[string]string               `json:"type_mapping"`
	SkipOptionalFieldsClasses []string                        `json:"skip_optional_fields_classes"`
	PagedOperations           map[string]PagedOperationConfig `json:"paged_operations"`
}

type Config struct {
	Modules map[string]ModuleConfig `json:"modules"`
}

// Generator handles Python SDK generation
type Generator struct {
	classes    []PythonClass
	config     Config
	moduleName string
}

// pythonTypeMapping maps OpenAPI types to Python types
var pythonTypeMapping = map[string]string{
	"string":  "str",
	"integer": "int",
	"number":  "float",
	"boolean": "bool",
	"array":   "List",
	"object":  "dict",
}

// PythonClass represents a Python class
type PythonClass struct {
	Name        string
	Description string
	Fields      []PythonField
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
	IsPaged             bool
	ResponseClass       string
	AsyncResponseType   string
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

func (g *Generator) loadConfig() error {
	configData, err := configFS.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("failed to read config.json: %w", err)
	}

	if err := json.Unmarshal(configData, &g.config); err != nil {
		return fmt.Errorf("failed to parse config.json: %w", err)
	}

	return nil
}

// Generate generates Python SDK code from parsed OpenAPI data
func (g *Generator) Generate(ctx context.Context, yamlContent []byte) (map[string]string, error) {
	// Load config first
	if err := g.loadConfig(); err != nil {
		return nil, err
	}

	p := parser.Parser{}
	modules, _, err := p.ParseOpenAPI(ctx, yamlContent)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI failed: %w", err)
	}

	// Generate code for each module
	files := make(map[string]string)

	// Read template
	tmpl, err := template.New("python").Parse(g.getTemplate())
	if err != nil {
		return nil, fmt.Errorf("parse template failed: %w", err)
	}

	// Convert modules to Python-specific format
	for moduleName, module := range modules {
		pythonModule := g.convertModule(module)
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]interface{}{
			"ModuleName": moduleName,
			"Operations": pythonModule.Operations,
			"Classes":    pythonModule.Classes,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template failed: %w", err)
		}
		files[fmt.Sprintf("%s", moduleName)] = buf.String()
	}

	return files, nil
}

func (g Generator) convertModule(module parser.Module) struct {
	Operations []PythonOperation
	Classes    []PythonClass
} {
	// Store current module name
	g.moduleName = module.Name

	classes := make([]PythonClass, 0, len(module.Classes))
	for _, class := range module.Classes {
		pythonClass := g.convertClass(class)
		classes = append(classes, pythonClass)
	}
	g.classes = classes

	operations := make([]PythonOperation, 0, len(module.Operations))
	for _, op := range module.Operations {
		operations = append(operations, g.convertOperation(op))
	}

	return struct {
		Operations []PythonOperation
		Classes    []PythonClass
	}{
		Operations: operations,
		Classes:    classes,
	}
}

func (g Generator) convertClass(class parser.Class) PythonClass {
	if g.config.Modules[g.moduleName].TypeMapping[class.Name] != "" {
		class.Name = g.config.Modules[g.moduleName].TypeMapping[class.Name]
	}

	pythonClass := PythonClass{
		Name:        class.Name,
		Description: g.formatDescription(class.Description),
		BaseClass:   "CozeModel",
	}

	if class.IsEnum {
		pythonClass.IsEnum = true
		pythonClass.BaseClass = "IntEnum"
		for _, value := range class.EnumValues {
			pythonClass.EnumValues = append(pythonClass.EnumValues, PythonEnumValue{
				Name:        g.toEnumName(value.Name),
				Value:       fmt.Sprintf("%v", value.Value),
				Description: value.Description,
			})
		}
		return pythonClass
	}

	if class.IsResponse {
		// Check if there's a mapping for this response type
		if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
			if _, ok := moduleConfig.ResponseTypeMapping[class.Name]; ok {
				// If mapped, skip this class
				pythonClass.ShouldSkip = true
			}
		}

		// Filter out code and msg fields for response classes
		var filteredFields []parser.Field
		for _, field := range class.Fields {
			if field.Name != "code" && field.Name != "msg" && field.Name != "detail" {
				filteredFields = append(filteredFields, field)
			}
		}

		// Skip class generation if only one field named "data" exists
		if len(filteredFields) == 1 && filteredFields[0].Name == "data" {
			pythonClass.ShouldSkip = true
		}

		// Add pass if no fields remain
		if len(filteredFields) == 0 {
			pythonClass.IsPass = true
			return pythonClass
		}

		class.Fields = filteredFields
	}

	// Track required fields
	requiredFields := make(map[string]bool)
	for _, fieldName := range class.Fields {
		if fieldName.Required {
			requiredFields[fieldName.Name] = true
		}
	}

	// Check if this class should skip optional field annotations
	skipOptionalFields := false
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		for _, skipClass := range moduleConfig.SkipOptionalFieldsClasses {
			if skipClass == class.Name {
				skipOptionalFields = true
				break
			}
		}
	}

	for _, field := range class.Fields {
		fieldType := g.getFieldType(field.Schema)
		if !field.Required && !skipOptionalFields {
			fieldType = fmt.Sprintf("Optional[%s]", fieldType)
		}
		pythonField := PythonField{
			Name:        g.toPythonVarName(field.Name),
			Type:        fieldType,
			Description: g.formatDescription(field.Description),
		}
		pythonClass.Fields = append(pythonClass.Fields, pythonField)
	}

	return pythonClass
}

func (g Generator) convertOperation(op parser.Operation) PythonOperation {
	operation := PythonOperation{
		Name:                g.toPythonMethodName(op.Name),
		Description:         op.Description,
		Path:                op.Path,
		Method:              op.Method,
		ResponseDescription: op.ResponseDescription,
	}

	// Check if this is a paged operation
	var pagedConfig *PagedOperationConfig
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		if config, ok := moduleConfig.PagedOperations[op.Name]; ok && config.Enabled {
			pagedConfig = &config
			operation.IsPaged = true
			operation.ResponseClass = config.ResponseClass
			operation.AsyncResponseType = config.AsyncReturnType
		}
	}

	// Handle parameters
	var headerParams []PythonParam
	var staticHeaders = make(map[string]string)
	for _, param := range op.Parameters {
		// Check if it's a header parameter with single enum value
		if param.In == "header" && param.Schema.Enum != nil && len(param.Schema.Enum) == 1 {
			// Use the single enum value directly
			staticHeaders[param.JsonName] = fmt.Sprintf("%v", param.Schema.Enum[0])
			continue
		}

		pythonType := g.getFieldType(param.Schema)
		paramName := param.Name

		// Apply parameter mapping for paged operations
		var isRequiredPagedParam bool
		if pagedConfig != nil {
			if mappedName, ok := pagedConfig.ParamMapping[param.JsonName]; ok {
				paramName = mappedName
				isRequiredPagedParam = true
			}
		}

		// Only add Optional if not required and not a required paged parameter
		if !param.Required && !isRequiredPagedParam {
			pythonType = fmt.Sprintf("Optional[%s]", pythonType)
		}

		pythonParam := PythonParam{
			Name:        g.toPythonVarName(paramName),
			JsonName:    param.JsonName,
			Type:        pythonType,
			Description: param.Description,
		}

		if !param.Required && !isRequiredPagedParam {
			pythonParam.DefaultValue = "None"
			pythonParam.HasDefault = true
		}

		// Check if parameter is a model type
		if param.Schema.Ref != "" {
			pythonParam.IsModel = true
		}

		operation.Params = append(operation.Params, pythonParam)
		if param.In == "query" {
			operation.QueryParams = append(operation.QueryParams, pythonParam)
			operation.HasQueryParams = true
		} else if param.In == "header" {
			headerParams = append(headerParams, pythonParam)
		}
	}

	// Handle request body
	if op.RequestBody != nil {
		operation.HasBody = true
		for name, prop := range op.RequestBody.Schema.Properties {
			pythonType := g.getFieldType(prop)
			// Request body fields are optional unless explicitly required
			isRequired := false
			for _, req := range op.RequestBody.Schema.Required {
				if req == name {
					isRequired = true
					break
				}
			}
			if !isRequired {
				pythonType = fmt.Sprintf("Optional[%s]", pythonType)
			}

			pythonParam := PythonParam{
				Name:        g.toPythonVarName(name),
				JsonName:    name,
				Type:        pythonType,
				Description: prop.Description,
				IsModel:     prop.Ref != "" || (prop.Type == parser.SchemaTypeArray && prop.Items != nil && prop.Items.Ref != ""),
			}
			if !isRequired {
				pythonParam.DefaultValue = "None"
				pythonParam.HasDefault = true
			}
			operation.Params = append(operation.Params, pythonParam)
			operation.BodyParams = append(operation.BodyParams, pythonParam)
		}
	}

	// Handle response
	operation.ResponseType = g.getFieldType(op.ResponseSchema)
	// Check if there's a mapping for this response type
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		if mappedType, ok := moduleConfig.ResponseTypeMapping[operation.ResponseType]; ok {
			operation.ResponseType = mappedType
		}
		// Override response type for paged operations
		if pagedConfig != nil {
			operation.ResponseType = pagedConfig.ReturnType
		}
	}

	// Update template to include headers
	if len(headerParams) > 0 || len(staticHeaders) > 0 {
		operation.HeaderParams = headerParams
		operation.StaticHeaders = staticHeaders
		operation.HasHeaders = true
	}

	return operation
}

func (g Generator) getFieldType(schema parser.Schema) string {
	// Handle response schema with data field
	if schema.IsResponse {
		if schema.Ref != "" {

			// Extract class name from ref (already processed in parser)
			refClassName := schema.Ref

			// Find the referenced class
			for _, class := range g.classes {
				if class.Name == refClassName {
					// Check if it has a data field
					for _, field := range class.Fields {
						if field.Name == "data" {
							// Remove Optional[] wrapper and return the type of data field directly
							if strings.HasPrefix(field.Type, "Optional[") && strings.HasSuffix(field.Type, "]") {
								return field.Type[9 : len(field.Type)-1]
							}
							return field.Type
						}
					}
					// No data field found, return the class name itself
					return refClassName
				}
			}
			// Class not found in g.classes, return the ref name
			return refClassName
		}
	}

	// If it's a reference, use the referenced type name (already processed in parser)
	if schema.Ref != "" {
		return schema.Ref
	}

	// Handle arrays
	if schema.Type == parser.SchemaTypeArray && schema.Items != nil {
		itemType := g.getFieldType(*schema.Items)
		return fmt.Sprintf("List[%s]", itemType)
	}

	// Get base type
	var baseType string
	switch schema.Type {
	case parser.SchemaTypeString:
		baseType = pythonTypeMapping["string"]
	case parser.SchemaTypeInteger:
		baseType = pythonTypeMapping["integer"]
	case parser.SchemaTypeNumber:
		baseType = pythonTypeMapping["number"]
	case parser.SchemaTypeBoolean:
		baseType = pythonTypeMapping["boolean"]
	case parser.SchemaTypeObject:
		if len(schema.Properties) > 0 {
			baseType = pythonTypeMapping["object"]
		} else {
			baseType = "Any"
		}
	default:
		baseType = "Any"
	}

	return baseType
}

func (g Generator) formatDescription(desc string) string {
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

func (g Generator) toPythonMethodName(name string) string {
	// First check if there's a mapping in the module-specific config
	if moduleConfig, ok := g.config.Modules[g.moduleName]; ok {
		if mappedName, ok := moduleConfig.OperationNameMapping[name]; ok {
			return mappedName
		}
	}

	// If no mapping found, use the default conversion logic
	// Convert method names like GetBot to get_bot
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func (g Generator) toPythonVarName(name string) string {
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

// Helper function to convert enum names to uppercase with underscores
func (g Generator) toEnumName(name string) string {
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

func (g Generator) getTemplate() string {
	// Read template from embedded file
	templateContent, err := fs.ReadFile(templateFS, "templates/sdk.tmpl")
	if err != nil {
		return ""
	}
	return string(templateContent)
}

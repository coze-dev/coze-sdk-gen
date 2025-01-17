package python

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"

	"github.com/coze-dev/coze-sdk-gen/parser"
)

//go:embed templates/sdk.tmpl
var templateFS embed.FS

// Generator handles Python SDK generation
type Generator struct{}

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
}

// PythonParam represents a Python parameter
type PythonParam struct {
	Name         string
	JsonName     string
	Type         string
	Description  string
	DefaultValue string
	IsModel      bool
}

// Generate generates Python SDK code from parsed OpenAPI data
func (g Generator) Generate(ctx context.Context, yamlContent []byte) (map[string]string, error) {
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
	operations := make([]PythonOperation, 0, len(module.Operations))
	for _, op := range module.Operations {
		operations = append(operations, g.convertOperation(op))
	}

	classes := make([]PythonClass, 0, len(module.Classes))
	for _, class := range module.Classes {
		classes = append(classes, g.convertClass(class))
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
				Name:        value.Name,
				Value:       fmt.Sprintf("%v", value.Value),
				Description: value.Description,
			})
		}
		return pythonClass
	}

	for _, field := range class.Fields {
		pythonField := PythonField{
			Name:        g.toPythonVarName(field.Name),
			Type:        g.getFieldType(field.Schema),
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

		pythonParam := PythonParam{
			Name:        g.toPythonVarName(param.Name),
			JsonName:    param.JsonName,
			Type:        g.getFieldType(param.Schema),
			Description: param.Description,
		}

		if !param.Required {
			pythonParam.Type = fmt.Sprintf("Optional[%s]", pythonParam.Type)
			pythonParam.DefaultValue = "None"
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
			pythonParam := PythonParam{
				Name:        g.toPythonVarName(name),
				JsonName:    name,
				Type:        g.getFieldType(prop),
				Description: prop.Description,
				IsModel:     prop.Ref != "" || (prop.Type == parser.SchemaTypeArray && prop.Items != nil && prop.Items.Ref != ""),
			}
			operation.Params = append(operation.Params, pythonParam)
			operation.BodyParams = append(operation.BodyParams, pythonParam)
		}
	}

	// Handle response
	if op.ResponseType != nil {
		operation.ResponseType = g.convertType(op.ResponseType)
	} else {
		operation.ResponseType = "Any"
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
	// If it's a reference, use the referenced type name
	if schema.Ref != "" {
		parts := strings.Split(schema.Ref, "/")
		return parts[len(parts)-1]
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

func (g Generator) getTemplate() string {
	// Read template from embedded file
	templateContent, err := fs.ReadFile(templateFS, "templates/sdk.tmpl")
	if err != nil {
		return ""
	}
	return string(templateContent)
}

func (g Generator) convertType(t *parser.Type) string {
	if t == nil {
		return "Any"
	}

	switch t.Kind {
	case parser.TypeKindPrimitive:
		if mappedType, ok := pythonTypeMapping[t.Name]; ok {
			return mappedType
		}
		return "Any"
	case parser.TypeKindArray:
		if t.ItemType != nil {
			itemType := g.convertType(t.ItemType)
			return fmt.Sprintf("List[%s]", itemType)
		}
		return "List[Any]"
	case parser.TypeKindRef:
		return t.Name
	case parser.TypeKindObject:
		if t.ObjectRef != "" {
			return t.ObjectRef
		}
		return "dict"
	default:
		return "Any"
	}
}

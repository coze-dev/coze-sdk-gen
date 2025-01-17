package parser

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Class represents a model/class in the SDK
type Class struct {
	Name        string
	Description string
	Fields      []Field
	BaseClass   string
	IsEnum      bool
	EnumValues  []EnumValue
}

// EnumValue represents a value in an enum
type EnumValue struct {
	Name        string
	Value       string
	Description string
}

// Field represents a field in a class
type Field struct {
	Name        string
	Type        string
	Description string
}

// Operation represents an API operation
type Operation struct {
	Name                string
	Description         string
	Path                string
	Method              string
	Params              []Param
	BodyParams          []Param
	QueryParams         []Param
	ResponseType        string
	ResponseDescription string
	HasBody             bool
	HasQueryParams      bool
	ModuleName          string
	HeaderParams        []Param
	HasHeaders          bool
	StaticHeaders       map[string]string
}

// Param represents a parameter in an operation
type Param struct {
	Name         string
	JsonName     string
	Type         string
	Description  string
	DefaultValue string
	IsModel      bool
}

// Module represents a group of operations
type Module struct {
	Name       string
	Operations []Operation
	Classes    []Class
}

// classDependency represents a class and its dependencies
type classDependency struct {
	name         string
	dependencies map[string]bool
}

// Parser handles OpenAPI parsing
type Parser struct{}

// ParseOpenAPI parses an OpenAPI document and returns modules and classes
func (p Parser) ParseOpenAPI(ctx context.Context, yamlContent []byte) (map[string]Module, []Class, error) {
	// Parse OpenAPI document
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(yamlContent)
	if err != nil {
		log.Printf("[ParseOpenAPI] parse openapi v3 failed. err=%v", err)
		return nil, nil, fmt.Errorf("parse openapi v3 failed: %w", err)
	}

	// Initialize Components if needed
	if doc.Components == nil {
		doc.Components = &openapi3.Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	}

	// Generate response data models from operations
	for _, pathItem := range doc.Paths.Map() {
		for _, op := range pathItem.Operations() {
			if response, ok := op.Responses.Map()["200"]; ok && response.Value.Content != nil {
				for _, content := range response.Value.Content {
					if content.Schema != nil && content.Schema.Value != nil && content.Schema.Value.Properties != nil {
						// Create response data model
						responseClass := p.convertSchemaToClass(op.OperationID+"Resp", &openapi3.SchemaRef{
							Value: content.Schema.Value,
						})
						// Add to components schemas to be generated
						doc.Components.Schemas[responseClass.Name] = &openapi3.SchemaRef{
							Value: content.Schema.Value,
						}
					}
				}
			}
		}
	}

	// Generate classes from schemas
	classes := []Class{}
	if doc.Components != nil && doc.Components.Schemas != nil {
		// Get sorted schema names based on dependencies
		sortedSchemaNames := p.analyzeDependencies(doc.Components.Schemas)

		// Generate classes in dependency order
		for _, name := range sortedSchemaNames {
			schema := doc.Components.Schemas[name]
			class := p.convertSchemaToClass(name, schema)
			classes = append(classes, class)
		}
	}

	// Group operations by module
	moduleOperations := make(map[string][]Operation)
	for path, pathItem := range doc.Paths.Map() {
		// Extract module name from path (second segment)
		segments := strings.Split(strings.Trim(path, "/"), "/")
		if len(segments) < 2 {
			continue
		}
		moduleName := segments[1]

		for method, op := range pathItem.Operations() {
			operation := p.convertOperation(path, method, op)
			operation.ModuleName = moduleName
			moduleOperations[moduleName] = append(moduleOperations[moduleName], operation)
		}
	}

	// Create modules
	modules := make(map[string]Module)
	for moduleName, operations := range moduleOperations {
		// Find all model dependencies for each module
		modelSet := make(map[string]bool)
		var findDependentClasses func(className string)
		findDependentClasses = func(className string) {
			if modelSet[className] {
				return
			}
			modelSet[className] = true
			// Find the class
			for _, class := range classes {
				if class.Name == className {
					// Add dependencies from fields
					for _, field := range class.Fields {
						if strings.Contains(field.Type, "List[") {
							depName := strings.TrimSuffix(strings.TrimPrefix(field.Type, "List["), "]")
							findDependentClasses(depName)
						} else if !strings.Contains(field.Type, "Optional[") && !strings.Contains(field.Type, "Dict[") &&
							field.Type != "str" && field.Type != "int" && field.Type != "float" && field.Type != "bool" && field.Type != "Any" {
							findDependentClasses(field.Type)
						}
					}
					break
				}
			}
		}

		// Start with direct dependencies
		for _, op := range operations {
			// Add response type if it's a model
			if strings.Contains(op.ResponseType, "List[") {
				modelName := strings.TrimSuffix(strings.TrimPrefix(op.ResponseType, "List["), "]")
				findDependentClasses(modelName)
			} else if op.ResponseType != "Any" {
				findDependentClasses(op.ResponseType)
			}

			// Add parameter types if they're models
			for _, param := range op.Params {
				if param.IsModel {
					if strings.Contains(param.Type, "List[") {
						modelName := strings.TrimSuffix(strings.TrimPrefix(param.Type, "List["), "]")
						findDependentClasses(modelName)
					} else {
						findDependentClasses(param.Type)
					}
				}
			}
		}

		// Find all classes needed for this module
		moduleClassList := []Class{}
		for _, class := range classes {
			if modelSet[class.Name] {
				moduleClassList = append(moduleClassList, class)
			}
		}

		modules[moduleName] = Module{
			Name:       moduleName,
			Operations: operations,
			Classes:    moduleClassList,
		}
	}

	return modules, classes, nil
}

func (p Parser) analyzeDependencies(schemas map[string]*openapi3.SchemaRef) []string {
	// Build dependency graph
	deps := make(map[string]*classDependency)
	for name := range schemas {
		deps[name] = &classDependency{
			name:         name,
			dependencies: make(map[string]bool),
		}
	}

	// Analyze dependencies
	for name, schema := range schemas {
		if schema.Value.Properties != nil {
			for _, prop := range schema.Value.Properties {
				if prop.Ref != "" {
					// Extract referenced type name
					parts := strings.Split(prop.Ref, "/")
					depName := parts[len(parts)-1]
					deps[name].dependencies[depName] = true
				} else if prop.Value != nil && prop.Value.Items != nil && prop.Value.Items.Ref != "" {
					// Handle array item dependencies
					parts := strings.Split(prop.Value.Items.Ref, "/")
					depName := parts[len(parts)-1]
					deps[name].dependencies[depName] = true
				}
			}
		}
	}

	// Topological sort
	var sorted []string
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(name string)
	visit = func(name string) {
		if visiting[name] {
			// Handle circular dependencies
			return
		}
		if visited[name] {
			return
		}
		visiting[name] = true
		for dep := range deps[name].dependencies {
			visit(dep)
		}
		visiting[name] = false
		visited[name] = true
		sorted = append(sorted, name)
	}

	// Visit all nodes
	for name := range deps {
		if !visited[name] {
			visit(name)
		}
	}

	return sorted
}

func (p Parser) convertOperation(path string, method string, op *openapi3.Operation) Operation {
	operation := Operation{
		Name:        p.toMethodName(op.OperationID),
		Description: op.Description,
		Path:        path,
		Method:      strings.ToUpper(method),
	}

	// Handle parameters
	var headerParams []Param
	var staticHeaders = make(map[string]string)
	for _, param := range op.Parameters {
		// Check if it's a header parameter with single enum value
		if param.Value.In == "header" && param.Value.Schema != nil && param.Value.Schema.Value != nil &&
			param.Value.Schema.Value.Enum != nil && len(param.Value.Schema.Value.Enum) == 1 {
			// Use the single enum value directly
			staticHeaders[param.Value.Name] = fmt.Sprintf("%v", param.Value.Schema.Value.Enum[0])
			continue
		}

		pythonParam := Param{
			Name:        p.toVarName(param.Value.Name),
			JsonName:    param.Value.Name,
			Type:        p.getFieldType(param.Value.Schema),
			Description: param.Value.Description,
		}

		if param.Value.Required {
			pythonParam.Type = p.getFieldType(param.Value.Schema)
		} else {
			pythonParam.Type = fmt.Sprintf("Optional[%s]", p.getFieldType(param.Value.Schema))
			pythonParam.DefaultValue = "None"
		}

		// Check if parameter is a model type
		if param.Value.Schema != nil && param.Value.Schema.Ref != "" {
			pythonParam.IsModel = true
		}

		operation.Params = append(operation.Params, pythonParam)
		if param.Value.In == "query" {
			operation.QueryParams = append(operation.QueryParams, pythonParam)
			operation.HasQueryParams = true
		} else if param.Value.In == "header" {
			headerParams = append(headerParams, pythonParam)
		}
	}

	// Handle request body
	if op.RequestBody != nil && op.RequestBody.Value.Content != nil {
		for _, content := range op.RequestBody.Value.Content {
			if content.Schema != nil {
				operation.HasBody = true
				if content.Schema.Value.Properties != nil {
					for name, prop := range content.Schema.Value.Properties {
						pythonParam := Param{
							Name:        p.toVarName(name),
							JsonName:    name,
							Type:        p.getFieldType(prop),
							Description: prop.Value.Description,
							IsModel:     prop.Ref != "" || (prop.Value != nil && prop.Value.Items != nil && prop.Value.Items.Ref != ""),
						}
						operation.Params = append(operation.Params, pythonParam)
						operation.BodyParams = append(operation.BodyParams, pythonParam)
					}
				}
			}
		}
	}

	// Handle response
	if response, ok := op.Responses.Map()["200"]; ok && response.Value.Content != nil {
		for mediaType, content := range response.Value.Content {
			if strings.Contains(mediaType, "application/json") && content.Schema != nil {
				// Set response type to OperationID + "Resp"
				operation.ResponseType = op.OperationID + "Resp"
				if response.Value.Description != nil {
					operation.ResponseDescription = *response.Value.Description
				}
				break
			}
		}
	}
	if operation.ResponseType == "" {
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

func (p Parser) toMethodName(name string) string {
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

func (p Parser) convertSchemaToClass(name string, schema *openapi3.SchemaRef) Class {
	class := Class{
		Name:        name,
		Description: p.formatDescription(schema.Value.Description),
		Fields:      []Field{},
		BaseClass:   "CozeModel",
	}

	if schema.Value.Title != "" {
		class.Description = p.formatDescription(schema.Value.Title)
	}

	// Handle enum types
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 && (*schema.Value.Type)[0] == "integer" && schema.Value.Enum != nil {
		class.IsEnum = true
		class.BaseClass = "IntEnum"
		for _, value := range schema.Value.Enum {
			// Convert value to int
			intValue := int(value.(float64))
			enumName := fmt.Sprintf("VALUE_%d", intValue)
			enumDesc := fmt.Sprintf("Value %d", intValue)

			if enumName != "" {
				class.EnumValues = append(class.EnumValues, EnumValue{
					Name:        enumName,
					Value:       fmt.Sprintf("%d", intValue),
					Description: enumDesc,
				})
			}
		}
		return class
	}

	if schema.Value.Properties != nil {
		for propName, prop := range schema.Value.Properties {
			field := Field{
				Name:        p.toVarName(propName),
				Type:        p.getFieldType(prop),
				Description: p.formatDescription(prop.Value.Description),
			}
			if prop.Value.Title != "" {
				field.Description = p.formatDescription(prop.Value.Title)
			}
			class.Fields = append(class.Fields, field)
		}
	}

	return class
}

func (p Parser) formatDescription(desc string) string {
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

func (p Parser) getFieldType(schema *openapi3.SchemaRef) string {
	if schema.Value == nil {
		return "Any"
	}

	// If it's a reference, use the referenced type name
	if schema.Ref != "" {
		parts := strings.Split(schema.Ref, "/")
		return parts[len(parts)-1]
	}

	// Handle arrays
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 && (*schema.Value.Type)[0] == "array" && schema.Value.Items != nil {
		itemType := p.getFieldType(schema.Value.Items)
		return fmt.Sprintf("List[%s]", itemType)
	}

	// Handle optional fields
	isOptional := len(schema.Value.Required) == 0
	var baseType string
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 {
		baseType = p.getBaseType((*schema.Value.Type)[0])
	}
	if baseType == "" {
		if schema.Value.Properties != nil {
			// If it has properties but no type, it's an object
			baseType = "dict"
		} else {
			baseType = "Any"
		}
	}

	if isOptional {
		return fmt.Sprintf("Optional[%s]", baseType)
	}
	return baseType
}

func (p Parser) getBaseType(openAPIType string) string {
	typeMapping := map[string]string{
		"string":  "str",
		"integer": "int",
		"number":  "float",
		"boolean": "bool",
		"array":   "List",
		"object":  "dict",
	}
	return typeMapping[openAPIType]
}

func (p Parser) toVarName(name string) string {
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

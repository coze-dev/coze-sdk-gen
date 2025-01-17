package parser

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Class represents a model/class in the SDK
type Class struct {
	Name        string
	Description string
	Fields      []Field
	IsEnum      bool
	EnumValues  []EnumValue
}

// EnumValue represents a value in an enum
type EnumValue struct {
	Name        string
	Value       interface{} // Keep the raw value for language-specific processing
	Description string
}

// Field represents a field in a class
type Field struct {
	Name        string
	Description string
	Required    bool
	Schema      Schema // Changed from *SchemaRef
}

// Operation represents an API operation
type Operation struct {
	Name                string
	OperationID         string // Original operation ID from API
	Description         string
	Path                string
	Method              string
	Parameters          []Parameter
	RequestBody         *RequestBody
	ResponseSchema      Schema
	ResponseDescription string
}

// Parameter represents a parameter in an operation
type Parameter struct {
	Name        string
	JsonName    string // Original name from API
	Description string
	Required    bool
	Schema      Schema // Changed from *SchemaRef
	In          string // query, header, path, etc.
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
					if content.Schema != nil && content.Schema.Value != nil {
						// Create response type if it's an object or array of objects
						if content.Schema.Value.Properties != nil ||
							(content.Schema.Value.Type != nil && len(*content.Schema.Value.Type) > 0 &&
								(*content.Schema.Value.Type)[0] == "array" && content.Schema.Value.Items != nil &&
								content.Schema.Value.Items.Value != nil && content.Schema.Value.Items.Value.Properties != nil) {
							// Add to components schemas to be generated
							doc.Components.Schemas[op.OperationID+"Resp"] = content.Schema
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
			moduleOperations[moduleName] = append(moduleOperations[moduleName], operation)
		}
	}

	// Create modules
	modules := make(map[string]Module)
	for moduleName, operations := range moduleOperations {
		// Find all model dependencies for each module
		modelSet := make(map[string]bool)
		var findDependentClasses func(schema Schema)
		findDependentClasses = func(schema Schema) {
			if schema.Ref != "" {
				parts := strings.Split(schema.Ref, "/")
				className := parts[len(parts)-1]
				if modelSet[className] {
					return
				}
				modelSet[className] = true
				// Find the schema
				if refSchema, ok := doc.Components.Schemas[className]; ok {
					findDependentClasses(p.convertOpenAPISchema(refSchema))
				}
			}
			if schema.Type == SchemaTypeObject {
				for _, prop := range schema.Properties {
					findDependentClasses(prop)
				}
			}
			if schema.Type == SchemaTypeArray && schema.Items != nil {
				findDependentClasses(*schema.Items)
			}
		}

		// Start with direct dependencies
		for _, op := range operations {
			// Add response schema dependencies
			findDependentClasses(op.ResponseSchema)

			// Add parameter dependencies
			for _, param := range op.Parameters {
				findDependentClasses(param.Schema)
			}

			// Add request body dependencies
			if op.RequestBody != nil {
				findDependentClasses(op.RequestBody.Schema)
			}
		}

		// Find all classes needed for this module
		moduleClassList := []Class{}
		for _, class := range classes {
			if modelSet[class.Name] {
				moduleClassList = append(moduleClassList, class)
				continue
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
		converted := p.convertOpenAPISchema(schema)
		p.findDependencies(name, converted, deps)
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

func (p Parser) findDependencies(name string, schema Schema, deps map[string]*classDependency) {
	if schema.Ref != "" {
		parts := strings.Split(schema.Ref, "/")
		depName := parts[len(parts)-1]
		deps[name].dependencies[depName] = true
		return
	}

	if schema.Type == SchemaTypeArray && schema.Items != nil {
		p.findDependencies(name, *schema.Items, deps)
		return
	}

	if schema.Type == SchemaTypeObject {
		for _, prop := range schema.Properties {
			p.findDependencies(name, prop, deps)
		}
	}
}

func (p Parser) convertOperation(path string, method string, op *openapi3.Operation) Operation {
	operation := Operation{
		Name:        op.OperationID,
		OperationID: op.OperationID,
		Description: op.Description,
		Path:        path,
		Method:      strings.ToUpper(method),
		RequestBody: nil,
	}

	// Set request body if exists
	if op.RequestBody != nil {
		operation.RequestBody = p.convertOpenAPIRequestBody(op.RequestBody.Value)
	}

	// Handle parameters
	for _, param := range op.Parameters {
		parameter := Parameter{
			Name:        param.Value.Name,
			JsonName:    param.Value.Name,
			Description: param.Value.Description,
			Required:    param.Value.Required,
			Schema:      p.convertOpenAPISchema(param.Value.Schema),
			In:          param.Value.In,
		}
		operation.Parameters = append(operation.Parameters, parameter)
	}

	// Handle response
	if response, ok := op.Responses.Map()["200"]; ok && response.Value.Content != nil {
		for mediaType, content := range response.Value.Content {
			if strings.Contains(mediaType, "application/json") && content.Schema != nil {
				// Set response schema - it's always a reference type
				operation.ResponseSchema = Schema{
					Type: SchemaTypeObject,
					Ref:  fmt.Sprintf("#/components/schemas/%sResp", op.OperationID),
				}

				if response.Value.Description != nil {
					operation.ResponseDescription = *response.Value.Description
				}
				break
			}
		}
	}

	return operation
}

func (p Parser) convertSchemaToClass(name string, schema *openapi3.SchemaRef) Class {
	class := Class{
		Name:        name,
		Description: schema.Value.Description,
	}

	if schema.Value.Title != "" {
		class.Description = schema.Value.Title
	}

	// Handle enum types
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 && schema.Value.Enum != nil {
		class.IsEnum = true
		for _, value := range schema.Value.Enum {
			enumName := fmt.Sprintf("VALUE_%v", value)
			enumDesc := fmt.Sprintf("Value %v", value)

			class.EnumValues = append(class.EnumValues, EnumValue{
				Name:        enumName,
				Value:       value,
				Description: enumDesc,
			})
		}
		return class
	}

	if schema.Value.Properties != nil {
		for propName, prop := range schema.Value.Properties {
			field := Field{
				Name:        propName,
				Description: prop.Value.Description,
				Required:    p.isFieldRequired(propName, schema),
				Schema:      p.convertOpenAPISchema(prop),
			}
			if prop.Value.Title != "" {
				field.Description = prop.Value.Title
			}
			class.Fields = append(class.Fields, field)
		}
	}

	return class
}

func (p Parser) isFieldRequired(fieldName string, schema *openapi3.SchemaRef) bool {
	if schema.Value.Required == nil {
		return false
	}
	for _, required := range schema.Value.Required {
		if required == fieldName {
			return true
		}
	}
	return false
}

// convertOpenAPISchema converts an OpenAPI schema to our internal schema type
func (p Parser) convertOpenAPISchema(schema *openapi3.SchemaRef) Schema {
	if schema == nil {
		return Schema{Type: SchemaTypeObject}
	}

	if schema.Ref != "" {
		return Schema{
			Type: SchemaTypeObject,
			Ref:  schema.Ref,
		}
	}

	if schema.Value == nil {
		return Schema{Type: SchemaTypeObject}
	}

	schemaType := SchemaTypeObject
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 {
		switch (*schema.Value.Type)[0] {
		case "string":
			schemaType = SchemaTypeString
		case "integer":
			schemaType = SchemaTypeInteger
		case "number":
			schemaType = SchemaTypeNumber
		case "boolean":
			schemaType = SchemaTypeBoolean
		case "array":
			schemaType = SchemaTypeArray
		}
	}

	properties := make(map[string]Schema)
	if schema.Value.Properties != nil {
		for name, prop := range schema.Value.Properties {
			properties[name] = p.convertOpenAPISchema(prop)
		}
	}

	var items *Schema
	if schema.Value.Items != nil {
		converted := p.convertOpenAPISchema(schema.Value.Items)
		items = &converted
	}

	return Schema{
		Type:        schemaType,
		Description: schema.Value.Description,
		Required:    schema.Value.Required,
		Properties:  properties,
		Items:       items,
		Enum:        schema.Value.Enum,
		Format:      schema.Value.Format,
	}
}

// convertOpenAPIRequestBody converts an OpenAPI request body to our internal type
func (p Parser) convertOpenAPIRequestBody(body *openapi3.RequestBody) *RequestBody {
	if body == nil {
		return nil
	}

	for _, mediaContent := range body.Content {
		if mediaContent.Schema != nil {
			return &RequestBody{
				Required: body.Required,
				Schema:   p.convertOpenAPISchema(mediaContent.Schema),
			}
		}
	}

	return nil
}

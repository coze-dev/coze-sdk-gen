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
	IsResponse  bool // Whether this class is a response model
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
	// Parse and initialize OpenAPI document
	doc, err := p.initializeOpenAPIDoc(yamlContent)
	if err != nil {
		return nil, nil, err
	}

	// Generate response models from operations
	p.generateResponseModels(doc)

	// Generate classes from schemas
	classes := p.generateClasses(doc)

	// Group operations by module and create final modules
	modules := p.generateModules(doc, classes)

	return modules, classes, nil
}

// initializeOpenAPIDoc parses and initializes the OpenAPI document
func (p Parser) initializeOpenAPIDoc(yamlContent []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(yamlContent)
	if err != nil {
		log.Printf("[ParseOpenAPI] parse openapi v3 failed. err=%v", err)
		return nil, fmt.Errorf("parse openapi v3 failed: %w", err)
	}

	// Initialize Components if needed
	if doc.Components == nil {
		doc.Components = &openapi3.Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	}

	return doc, nil
}

// getResponseSchemaName returns a consistent name for response schemas
func (p Parser) getResponseSchemaName(operationID string) string {
	return fmt.Sprintf("%sResp", operationID)
}

// processResponseContent extracts and processes the response schema from operation content
func (p Parser) processResponseContent(content *openapi3.MediaType) *openapi3.SchemaRef {
	if content == nil || content.Schema == nil || content.Schema.Value == nil {
		return nil
	}

	schema := content.Schema
	schemaValue := schema.Value

	// Only process if it's an object or array of objects
	if schemaValue.Properties != nil || p.isArrayOfObjects(schemaValue) {
		// Create a copy of the schema
		newSchema := *schema
		// Merge existing extensions with is_response_schema
		if newSchema.Extensions == nil {
			newSchema.Extensions = map[string]any{"is_response_schema": true}
		} else {
			// Create a new map to avoid modifying the original
			extensions := make(map[string]any, len(newSchema.Extensions)+1)
			for k, v := range newSchema.Extensions {
				extensions[k] = v
			}
			extensions["is_response_schema"] = true
			newSchema.Extensions = extensions
		}
		return &newSchema
	}

	return nil
}

// isArrayOfObjects checks if the schema represents an array of objects
func (p Parser) isArrayOfObjects(schema *openapi3.Schema) bool {
	return schema.Type != nil &&
		len(*schema.Type) > 0 &&
		(*schema.Type)[0] == "array" &&
		schema.Items != nil &&
		schema.Items.Value != nil &&
		schema.Items.Value.Properties != nil
}

// generateResponseModels generates response data models from operations and adds them to the components schemas
func (p Parser) generateResponseModels(doc *openapi3.T) {
	// Ensure components and schemas are initialized
	if doc.Components == nil {
		doc.Components = &openapi3.Components{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = make(map[string]*openapi3.SchemaRef)
	}

	// Process each operation's response
	for _, pathItem := range doc.Paths.Map() {
		for _, op := range pathItem.Operations() {
			// Get 200 response if exists
			response, ok := op.Responses.Map()["200"]
			if !ok || response.Value.Content == nil {
				continue
			}

			// Process JSON content type
			for contentType, content := range response.Value.Content {
				if !strings.Contains(contentType, "application/json") {
					continue
				}

				if schema := p.processResponseContent(content); schema != nil {
					// Add schema to components with consistent naming
					schemaName := p.getResponseSchemaName(op.OperationID)
					doc.Components.Schemas[schemaName] = schema

					// Update the response content to reference the schema
					content.Schema = &openapi3.SchemaRef{
						Ref: fmt.Sprintf("#/components/schemas/%s", schemaName),
					}
				}
			}
		}
	}
}

// generateClasses generates classes from schemas
func (p Parser) generateClasses(doc *openapi3.T) []Class {
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
	return classes
}

// generateModules groups operations by module and creates final modules
func (p Parser) generateModules(doc *openapi3.T, classes []Class) map[string]Module {
	// Group operations by module
	moduleOperations := p.groupOperationsByModule(doc)

	// Create final modules with their dependencies
	return p.createModulesWithDependencies(moduleOperations, classes, doc)
}

// groupOperationsByModule groups operations by their module name
func (p Parser) groupOperationsByModule(doc *openapi3.T) map[string][]Operation {
	moduleOperations := make(map[string][]Operation)
	for path, pathItem := range doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			// Use tags as module name, joined by dots
			moduleName := strings.Join(op.Tags, ".")
			if moduleName == "" {
				// Fallback to default module if no tags
				moduleName = "default"
			}

			operation := p.convertOperation(path, method, op)
			moduleOperations[moduleName] = append(moduleOperations[moduleName], operation)
		}
	}
	return moduleOperations
}

// createModulesWithDependencies creates modules with their class dependencies
func (p Parser) createModulesWithDependencies(moduleOperations map[string][]Operation, classes []Class, doc *openapi3.T) map[string]Module {
	modules := make(map[string]Module)
	for moduleName, operations := range moduleOperations {
		// Find all model dependencies for this module
		modelSet := p.findModuleDependencies(operations, doc)

		// Create module class list
		moduleClassList := p.createModuleClassList(classes, modelSet)

		modules[moduleName] = Module{
			Name:       moduleName,
			Operations: operations,
			Classes:    moduleClassList,
		}
	}
	return modules
}

// findModuleDependencies finds all class dependencies for a module
func (p Parser) findModuleDependencies(operations []Operation, doc *openapi3.T) map[string]bool {
	modelSet := make(map[string]bool)
	var findDependentClasses func(schema Schema)
	findDependentClasses = func(schema Schema) {
		if schema.Ref != "" {
			if modelSet[schema.Ref] {
				return
			}
			modelSet[schema.Ref] = true
			// Find the schema
			if refSchema, ok := doc.Components.Schemas[schema.Ref]; ok {
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

	// Process all operations
	for _, op := range operations {
		findDependentClasses(op.ResponseSchema)
		for _, param := range op.Parameters {
			findDependentClasses(param.Schema)
		}
		if op.RequestBody != nil {
			findDependentClasses(op.RequestBody.Schema)
		}
	}

	return modelSet
}

// createModuleClassList creates a list of classes needed for a module
func (p Parser) createModuleClassList(classes []Class, modelSet map[string]bool) []Class {
	var moduleClassList []Class
	for _, class := range classes {
		if modelSet[class.Name] {
			moduleClassList = append(moduleClassList, class)
		}
	}
	return moduleClassList
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
		deps[name].dependencies[schema.Ref] = true
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
				// Process ref to get the last component
				ref := content.Schema.Ref
				if ref != "" {
					parts := strings.Split(ref, "/")
					ref = parts[len(parts)-1]
				}
				operation.ResponseSchema = Schema{
					Type:       SchemaTypeObject,
					Ref:        ref,
					IsResponse: true,
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
	// Check if this is a response schema
	isResponseSchema := false
	if val, ok := schema.Extensions["is_response_schema"].(bool); ok {
		isResponseSchema = val
	}

	class := Class{
		Name:        name,
		Description: schema.Value.Description,
		IsResponse:  isResponseSchema,
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

	// Check if this is a response schema
	isResponseSchema := false
	if val, ok := schema.Extensions["is_response_schema"].(bool); ok {
		isResponseSchema = val
	}

	if schema.Ref != "" {
		if isResponseSchema {
			fmt.Printf("is_response_schema: ref\n")
		}
		// Process ref to get the last component
		parts := strings.Split(schema.Ref, "/")
		ref := parts[len(parts)-1]
		return Schema{
			Type:       SchemaTypeObject,
			Ref:        ref,
			IsResponse: isResponseSchema,
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
		IsResponse:  isResponseSchema,
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

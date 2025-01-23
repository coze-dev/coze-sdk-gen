package parser

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/util"
	"github.com/getkin/kin-openapi/openapi3"
)

// TyKind represents the kind of a type
type TyKind int

const (
	TyKindPrimitive TyKind = iota
	TyKindObject
	TyKindArray
)

// PrimitiveKind represents primitive types
type PrimitiveKind string

const (
	PrimitiveInt     PrimitiveKind = "int"
	PrimitiveFloat   PrimitiveKind = "float"
	PrimitiveString  PrimitiveKind = "string"
	PrimitiveBool    PrimitiveKind = "bool"
	PrimitiveBinary  PrimitiveKind = "binary"
	PrimitiveUnknown PrimitiveKind = "unknown"
)

// Ty represents a type in the schema
type Ty struct {
	Name        string
	Description string
	Kind        TyKind
	Module      string // The module this type belongs to

	// For primitive types
	PrimitiveKind PrimitiveKind
	EnumValues    []TyEnumValue // Optional enum values for primitive types

	// For object types
	Fields []TyField

	// For array types
	ElementType *Ty

	// Metadata
	IsNamed bool // Whether this is a named type (from components)
}

// TyField represents a field in an object type
type TyField struct {
	Name        string
	Description string
	Type        *Ty
	Required    bool
}

type TyEnumValue struct {
	Name string
	Val  interface{}
}

// TyParameter represents an operation parameter
type TyParameter struct {
	Name        string
	Description string
	Required    bool
	Type        *Ty
}

// HttpHandler represents an API operation
type HttpHandler struct {
	Name        string
	Description string
	Path        string
	Method      string

	// Parameters split by location
	HeaderParams []TyParameter
	PathParams   []TyParameter
	QueryParams  []TyParameter

	// Request and Response
	RequestBody  *Ty
	ResponseBody *Ty
}

// TyModule represents a group of operations and types
type TyModule struct {
	Name         string
	HttpHandlers []HttpHandler
	Types        []*Ty
}

// ModuleConfig represents the configuration for type-to-module mapping
type ModuleConfig struct {
	TypeModuleMap map[string]string `json:"type_module_map"` // Maps type names to module names
}

// Parser2 handles OpenAPI parsing with the new schema design
type Parser2 struct {
	types   map[string]*Ty       // All types indexed by name
	modules map[string]*TyModule // All modules
	config  *ModuleConfig        // Module configuration
	doc     *openapi3.T          // The OpenAPI document
}

// NewParser2 creates a new Parser2 instance
func NewParser2(config *ModuleConfig) (*Parser2, error) {
	return &Parser2{
		types:   make(map[string]*Ty),
		modules: make(map[string]*TyModule),
		config:  config,
	}, nil
}

// ParseOpenAPI parses an OpenAPI document and returns modules
func (p *Parser2) ParseOpenAPI(yamlContent []byte) (map[string]*TyModule, error) {
	// Parse OpenAPI document
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}
	p.doc = doc

	// Process all named types from components
	if err := p.processNamedTypes(); err != nil {
		return nil, err
	}

	// Process all operations and their types
	if err := p.processOperations(); err != nil {
		return nil, err
	}

	// Assign types to modules
	if err := p.assignTypesToModules(); err != nil {
		return nil, err
	}

	return p.modules, nil
}

// GetType returns a type by name
func (p *Parser2) GetType(name string) *Ty {
	return p.types[name]
}

// processNamedTypes processes all named types from components
func (p *Parser2) processNamedTypes() error {
	if p.doc.Components == nil || p.doc.Components.Schemas == nil {
		return nil
	}

	for name, schema := range p.doc.Components.Schemas {
		ty, err := p.convertSchema(schema, name, true)
		if err != nil {
			return fmt.Errorf("failed to convert schema %s: %w", name, err)
		}
		p.types[name] = ty
	}
	return nil
}

// processOperations processes all operations and their types
func (p *Parser2) processOperations() error {
	for path, pathItem := range p.doc.Paths.Map() {
		for method, op := range pathItem.Operations() {
			handler, err := p.convertOperation(path, method, op)
			if err != nil {
				return fmt.Errorf("failed to convert operation %s: %w", op.OperationID, err)
			}

			// Get or create module
			moduleName := "default"
			if len(op.Tags) > 0 {
				moduleName = op.Tags[0]
			}

			module, ok := p.modules[moduleName]
			if !ok {
				module = &TyModule{Name: moduleName}
				p.modules[moduleName] = module
			}

			module.HttpHandlers = append(module.HttpHandlers, *handler)
		}
	}
	return nil
}

// convertSchema converts an OpenAPI schema to our type system
func (p *Parser2) convertSchema(schema *openapi3.SchemaRef, name string, isNamed bool) (*Ty, error) {
	if schema == nil || schema.Value == nil {
		return nil, fmt.Errorf("nil schema")
	}

	// If it's a reference and we've already processed it, return the existing type
	if schema.Ref != "" {
		refName := getRefName(schema.Ref)
		if existing := p.types[refName]; existing != nil {
			return existing, nil
		}
	}

	ty := &Ty{
		Name:        name,
		IsNamed:     isNamed,
		Description: util.Choose(schema.Value.Title != "", schema.Value.Title, schema.Value.Description),
	}

	// Determine the kind of type
	if schema.Value.Type != nil && len(*schema.Value.Type) > 0 {
		switch (*schema.Value.Type)[0] {
		case "array":
			ty.Kind = TyKindArray
			if schema.Value.Items != nil {
				elementType, err := p.convertSchema(schema.Value.Items, "", false)
				if err != nil {
					return nil, fmt.Errorf("failed to convert array element type: %w", err)
				}
				ty.ElementType = elementType
			}

		case "object":
			ty.Kind = TyKindObject
			for _, pname := range schema.Value.Extensions["x-coze-order"].([]interface{}) {
				propName := pname.(string)
				prop := schema.Value.Properties[propName]

				field, err := p.convertField(propName, prop, schema.Value.Required)
				if err != nil {
					return nil, fmt.Errorf("failed to convert field %s: %w", propName, err)
				}
				ty.Fields = append(ty.Fields, *field)
			}

		default:
			ty.Kind = TyKindPrimitive
			ty.PrimitiveKind = p.convertPrimitiveType(*schema.Value.Type, schema.Value.Format)
			if schema.Value.Enum != nil {
				for _, val := range schema.Value.Enum {
					ty.EnumValues = append(ty.EnumValues, TyEnumValue{Name: "", Val: val})
				}
				if enumNames, ok := schema.Value.Extensions["x-coze-enum-names"].([]interface{}); ok && len(enumNames) == len(schema.Value.Enum) {
					for i := range ty.EnumValues {
						ty.EnumValues[i].Name = enumNames[i].(string)
					}
				}
			}
		}
	}

	// Store named types in the type map
	if isNamed {
		p.types[name] = ty
	}

	return ty, nil
}

// convertField converts a schema property to a field
func (p *Parser2) convertField(name string, schema *openapi3.SchemaRef, required []string) (*TyField, error) {
	fieldType, err := p.convertSchema(schema, "", false)
	if err != nil {
		return nil, err
	}

	isRequired := false
	for _, req := range required {
		if req == name {
			isRequired = true
			break
		}
	}

	return &TyField{
		Name:        name,
		Description: schema.Value.Description,
		Type:        fieldType,
		Required:    isRequired,
	}, nil
}

// convertOperation converts an OpenAPI operation to our HttpHandler
func (p *Parser2) convertOperation(path, method string, op *openapi3.Operation) (*HttpHandler, error) {
	handler := &HttpHandler{
		Name:        op.OperationID,
		Description: op.Description,
		Path:        path,
		Method:      method,
	}

	// Convert parameters
	for _, param := range op.Parameters {
		if param.Value == nil {
			continue
		}

		paramType, err := p.convertSchema(param.Value.Schema, "", false)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter schema: %w", err)
		}

		parameter := TyParameter{
			Name:        param.Value.Name,
			Description: param.Value.Description,
			Required:    param.Value.Required,
			Type:        paramType,
		}

		switch param.Value.In {
		case "header":
			handler.HeaderParams = append(handler.HeaderParams, parameter)
		case "path":
			handler.PathParams = append(handler.PathParams, parameter)
		case "query":
			handler.QueryParams = append(handler.QueryParams, parameter)
		}
	}

	// Convert request body
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for _, content := range op.RequestBody.Value.Content {
			if content.Schema != nil {
				requestType, err := p.convertSchema(content.Schema, "", false)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request body schema: %w", err)
				}
				handler.RequestBody = requestType
				break
			}
		}
	}

	// Convert response body
	if response, ok := op.Responses.Map()["200"]; ok && response.Value.Content != nil {
		for _, content := range response.Value.Content {
			if content.Schema != nil {
				responseType, err := p.convertSchema(content.Schema, "", false)
				if err != nil {
					return nil, fmt.Errorf("failed to convert response schema: %w", err)
				}
				handler.ResponseBody = responseType
				break
			}
		}
	}

	return handler, nil
}

// assignTypesToModules assigns types to modules based on configuration or usage
func (p *Parser2) assignTypesToModules() error {
	// First, try to assign types based on configuration
	if p.config != nil {
		for typeName, moduleName := range p.config.TypeModuleMap {
			if ty := p.types[typeName]; ty != nil {
				ty.Module = moduleName
				if module := p.modules[moduleName]; module != nil {
					module.Types = append(module.Types, ty)
				}
			}
		}
	}

	// For remaining types, assign based on usage
	for _, ty := range p.types {
		// Find the first module that uses this type
		for _, module := range p.modules {
			if p.isTypeUsedInModule(ty, module) {
				if ty.Module == "" {
					ty.Module = module.Name
				}
				module.Types = append(module.Types, ty)
				break
			}
		}
	}

	return nil
}

// isTypeUsedInModule checks if a type is used in a module
func (p *Parser2) isTypeUsedInModule(ty *Ty, module *TyModule) bool {
	for _, handler := range module.HttpHandlers {
		if p.isTypeUsedInHandler(ty, &handler) {
			return true
		}
	}
	return false
}

// isTypeUsedInHandler checks if a type is used in a handler
func (p *Parser2) isTypeUsedInHandler(ty *Ty, handler *HttpHandler) bool {
	// Check request body
	if handler.RequestBody == ty {
		return true
	}

	// Check response body
	if handler.ResponseBody == ty {
		return true
	}

	// Check parameters
	for _, param := range handler.HeaderParams {
		if param.Type == ty {
			return true
		}
	}
	for _, param := range handler.PathParams {
		if param.Type == ty {
			return true
		}
	}
	for _, param := range handler.QueryParams {
		if param.Type == ty {
			return true
		}
	}

	return false
}

// convertPrimitiveType converts OpenAPI type to our primitive type
func (p *Parser2) convertPrimitiveType(typ []string, format string) PrimitiveKind {
	if len(typ) == 0 {
		return PrimitiveUnknown
	}

	switch typ[0] {
	case "integer":
		return PrimitiveInt
	case "number":
		return PrimitiveFloat
	case "string":
		if format == "binary" {
			return PrimitiveBinary
		}
		return PrimitiveString
	case "boolean":
		return PrimitiveBool
	default:
		return PrimitiveUnknown
	}
}

// getRefName extracts the name from a reference
func getRefName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/util"
	"github.com/getkin/kin-openapi/openapi3"
)

// TyKind represents the kind of a type
type TyKind string

const (
	TyKindPrimitive TyKind = "primimtive"
	TyKindObject    TyKind = "object"
	TyKindArray     TyKind = "array"
	TyKindMap       TyKind = "map" // New type for map
)

// PrimitiveKind represents primitive types
type PrimitiveKind string

const (
	PrimitiveInt     PrimitiveKind = "int"
	PrimitiveFloat   PrimitiveKind = "float"
	PrimitiveString  PrimitiveKind = "string"
	PrimitiveBool    PrimitiveKind = "bool"
	PrimitiveBinary  PrimitiveKind = "binary"
	PrimitiveUnknown PrimitiveKind = ""
)

// Ty represents a type in the schema
type Ty struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Kind        TyKind `json:"kind"`
	Module      string `json:"module,omitempty"` // The module this type belongs to

	// For primitive types
	PrimitiveKind PrimitiveKind `json:"primitive_kind,omitempty"`
	EnumValues    []TyEnumValue `json:"enum_values,omitempty"` // Optional enum values for primitive types

	// For object types
	Fields []TyField `json:"fields,omitempty"`

	// For array types
	ElementType *Ty `json:"element_type,omitempty"`

	// For map types
	ValueType *Ty `json:"value_type,omitempty"` // Type of the map values

	// Metadata
	IsNamed bool `json:"is_named,omitempty"` // Whether this is a named type (from components)
}

// TyField represents a field in an object type
type TyField struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        *Ty    `json:"type"`
	Required    bool   `json:"required,omitempty"`
}

type TyEnumValue struct {
	Name string      `json:"name,omitempty"`
	Val  interface{} `json:"val"`
}

// ContentType represents the request content type
type ContentType string

const (
	ContentTypeJson ContentType = "json"
	ContentTypeFile ContentType = "file"
)

// HttpHandler represents an API operation
type HttpHandler struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	Method      string `json:"method"`

	// Content Type
	ContentType ContentType `json:"content_type"`

	// Parameters split by location
	HeaderParams []TyField `json:"header_params,omitempty"`
	PathParams   []TyField `json:"path_params,omitempty"`
	QueryParams  []TyField `json:"query_params,omitempty"`

	// Request and Response
	RequestBody  *Ty `json:"request_body"`
	ResponseBody *Ty `json:"response_body"`
}

// TyModule represents a group of operations and types
type TyModule struct {
	Name         string        `json:"name"`
	HttpHandlers []HttpHandler `json:"http_handlers"`
	Types        []*Ty         `json:"types"`
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

// TODO: delete this
func marshal(v any) string {
	res, _ := json.Marshal(v)
	return string(res)
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
			return fmt.Errorf("failed to convert schema %s: %+v err: %w", name, schema, err)
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
				moduleName = strings.Join(op.Tags, ".")
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
	if schema == nil {
		return nil, fmt.Errorf("nil schema. s=%v", marshal(schema))
	}
	if schema.Value == nil {
		return nil, fmt.Errorf("nil schema value. v=%v", marshal(schema.Value))
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

	// Check if it's a map type first
	if schema.Value.AdditionalProperties.Schema != nil {
		ty.Kind = TyKindMap
		valueType, err := p.convertSchema(schema.Value.AdditionalProperties.Schema, "", false)
		if err != nil {
			return nil, fmt.Errorf("failed to convert map value type: %w", err)
		}
		ty.ValueType = valueType
		return ty, nil
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
			// TODO: map: ref: additionalProperties
			// Check if x-coze-order exists and is not nil
			if order, ok := schema.Value.Extensions["x-coze-order"]; ok && order != nil {
				// Process properties in order
				for _, pname := range order.([]interface{}) {
					propName := pname.(string)
					prop := schema.Value.Properties[propName]
					if prop == nil {
						continue
					}

					field, err := p.convertField(propName, prop, schema.Value.Required)
					if err != nil {
						return nil, fmt.Errorf("failed to convert field %s: %w", propName, err)
					}
					ty.Fields = append(ty.Fields, *field)
				}
			} else {
				// If no order specified, process properties in map order
				for propName, prop := range schema.Value.Properties {
					field, err := p.convertField(propName, prop, schema.Value.Required)
					if err != nil {
						return nil, fmt.Errorf("failed to convert field %s: %w", propName, err)
					}
					ty.Fields = append(ty.Fields, *field)
				}
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
	} else {
		ty.Description = ""
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
		Description: util.Choose(schema.Value.Title != "", schema.Value.Title, schema.Value.Description),
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
		ContentType: ContentTypeJson, // Default to JSON
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

		parameter := TyField{
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
		for contentType, content := range op.RequestBody.Value.Content {
			if content.Schema != nil {
				requestType, err := p.convertSchema(content.Schema, "", false)
				if err != nil {
					return nil, fmt.Errorf("failed to convert request body schema: %w", err)
				}
				handler.RequestBody = requestType

				// Set content type based on request body content type
				switch contentType {
				case "multipart/form-data":
					handler.ContentType = ContentTypeFile
				case "application/json":
					handler.ContentType = ContentTypeJson
				}
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

	// Pre-calculate handler dependencies
	handlerDeps := make(map[*HttpHandler]map[*Ty]bool)
	for _, module := range p.modules {
		for i := range module.HttpHandlers {
			deps := make(map[*Ty]bool)
			p.collectHandlerTypes(&module.HttpHandlers[i], deps)
			handlerDeps[&module.HttpHandlers[i]] = deps
		}
	}

	// For remaining types, assign based on usage
	for _, ty := range p.types {
		if ty.Module != "" {
			continue // Skip if already assigned
		}
		// Find the first module that uses this type
		for _, module := range p.modules {
			if p.isTypeUsedInModule(ty, module, handlerDeps) {
				ty.Module = module.Name
				module.Types = append(module.Types, ty)
				break
			}
		}
	}

	return nil
}

// isTypeUsedInModule checks if a type is used in a module
func (p *Parser2) isTypeUsedInModule(ty *Ty, module *TyModule, handlerDeps map[*HttpHandler]map[*Ty]bool) bool {
	for i := range module.HttpHandlers {
		if deps := handlerDeps[&module.HttpHandlers[i]]; deps != nil && deps[ty] {
			return true
		}
	}
	return false
}

// collectHandlerTypes recursively collects all types used in a handler
func (p *Parser2) collectHandlerTypes(handler *HttpHandler, deps map[*Ty]bool) {
	// Helper function to collect types from a single type
	var collectFromType func(*Ty)
	collectFromType = func(t *Ty) {
		if t == nil || deps[t] {
			return
		}
		deps[t] = true

		switch t.Kind {
		case TyKindObject:
			for _, field := range t.Fields {
				collectFromType(field.Type)
			}
		case TyKindArray:
			collectFromType(t.ElementType)
		}
	}

	// Collect from request body
	collectFromType(handler.RequestBody)

	// Collect from response body
	collectFromType(handler.ResponseBody)

	// Collect from parameters
	for _, param := range handler.HeaderParams {
		collectFromType(param.Type)
	}
	for _, param := range handler.PathParams {
		collectFromType(param.Type)
	}
	for _, param := range handler.QueryParams {
		collectFromType(param.Type)
	}
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

package parser

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/util"
	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

// TyKind represents the kind of a type
type TyKind string

const (
	TyKindPrimitive TyKind = "primimtive"
	TyKindObject    TyKind = "object"
	TyKindArray     TyKind = "array"
	TyKindMap       TyKind = "map" // New type for map
)

// FieldRequirementChange represents the type of change to make to a field's requirement
type FieldRequirementChange int

const (
	FieldRequirementNoChange FieldRequirementChange = iota
	FieldRequirementRequired
	FieldRequirementOptional
)

// FieldModification represents modifications to be made to a field
type FieldModification struct {
	Requirement FieldRequirementChange `json:"requirement,omitempty"` // Change to make to field's requirement
	Default     string                 `json:"default,omitempty"`     // Default value to set for the field
}

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
	Default     string `json:"default,omitempty"`
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

// Default pagination parameter candidates
var (
	DefaultPageIndexCandidates = []string{"page_index", "page_num"}
	DefaultPageSizeCandidates  = []string{"page_size", "page_num"}
)

// PageInfo represents pagination information
type PageInfo struct {
	ItemType      *Ty    // The type of items in the paginated array
	PageIndexName string // The parameter name for page index/number
	PageSizeName  string // The parameter name for page size
}

// GetPageInfo checks if this handler represents a paginated request and returns pagination details.
// A request is considered paginated if:
// 1. It's a GET request
// 2. Has two query parameters matching the candidates for page index and size
// 3. The actual response body contains an array field
// Returns nil if the handler is not a paginated request.
func (h *HttpHandler) GetPageInfo(pageIndexCandidates, pageSizeCandidates []string) *PageInfo {
	// Use default candidates if none provided
	if len(pageIndexCandidates) == 0 {
		pageIndexCandidates = DefaultPageIndexCandidates
	}
	if len(pageSizeCandidates) == 0 {
		pageSizeCandidates = DefaultPageSizeCandidates
	}

	// Check if it's a GET request
	if h.Method != "GET" {
		return nil
	}

	// Check query parameters
	var pageIndex, pageSize string
	paramNames := make(map[string]bool)
	for _, param := range h.QueryParams {
		paramNames[param.Name] = true
	}

	// Check for page index parameter
	for _, candidate := range pageIndexCandidates {
		if paramNames[candidate] {
			pageIndex = candidate
			break
		}
	}

	// Check for page size parameter
	for _, candidate := range pageSizeCandidates {
		if paramNames[candidate] && candidate != pageIndex {
			pageSize = candidate
			break
		}
	}

	if pageIndex == "" || pageSize == "" {
		return nil
	}

	// Check response body
	actualBody := h.GetActualResponseBody()
	if actualBody == nil || actualBody.Kind != TyKindObject {
		return nil
	}

	// Look for an array field in the actual response body
	for _, field := range actualBody.Fields {
		if field.Type.Kind == TyKindArray {
			return &PageInfo{
				ItemType:      field.Type.ElementType,
				PageIndexName: pageIndex,
				PageSizeName:  pageSize,
			}
		}
	}
	return nil
}

// GetActualResponseBody returns the "actual" response body type.
// If ResponseBody has a "data" field, returns its type, otherwise returns nil.
func (h *HttpHandler) GetActualResponseBody() *Ty {
	if h.ResponseBody == nil || h.ResponseBody.Kind != TyKindObject {
		return nil
	}

	for _, field := range h.ResponseBody.Fields {
		if field.Name == "data" {
			return field.Type
		}
	}
	return nil
}

// HasOnlyStatusFields checks if the response body only contains status-related fields
// like code, msg, detail, etc., indicating it's just a status response without meaningful data.
func (h *Ty) HasOnlyStatusFields() bool {
	if h == nil || h.Kind != TyKindObject {
		return false
	}

	statusFields := map[string]bool{
		"code":   true,
		"msg":    true,
		"detail": true,
		"logid":  true,
	}

	for _, field := range h.Fields {
		if !statusFields[field.Name] {
			return false
		}
	}

	return len(h.Fields) > 0
}

// Module represents a group of operations and types
type Module struct {
	Name         string        `json:"name"`
	HttpHandlers []HttpHandler `json:"http_handlers"`
	Types        []*Ty         `json:"types"`
}

// ModuleConfig represents the configuration for type-to-module mapping
type ModuleConfig struct {
	TypeModuleMap                 map[string]string                        `json:"type_module_map"`                   // Maps type names to module names
	GenerateUnnamedResponseType   func(*HttpHandler) (string, bool)        `json:"generate_unnamed_response_type"`    // if response type is unnamed, will auto gen named
	ChangeHttpHandlerResponseType map[string]string                        `json:"change_http_handler_response_type"` // change response type for http handler
	RenameTypes                   map[string]string                        `json:"rename_types"`                      // rename types, key is old name, value is new name
	RenameHandlers                map[string]string                        `json:"rename_handlers"`                   // rename http handlers, key is old name, value is new name
	ChangeFields                  map[string]map[string]*FieldModification `json:"change_fields"`                     // change field properties, first key is type name, second key is field name
	HandlerOrdering               map[string][]string                      `json:"handler_ordering"`                  // order handlers in modules, key is module name, value is ordered handler names
}

// Parser handles OpenAPI parsing with the new schema design
type Parser struct {
	namedTypes map[string]*Ty     // All types indexed by name
	modules    map[string]*Module // All modules
	config     *ModuleConfig      // Module configuration
	doc        *openapi3.T        // The OpenAPI document
}

// NewParser creates a new Parser2 instance
func NewParser(config *ModuleConfig) (*Parser, error) {
	if config == nil {
		config = &ModuleConfig{}
	}

	return &Parser{
		namedTypes: make(map[string]*Ty),
		modules:    make(map[string]*Module),
		config:     config,
	}, nil
}

// TODO: delete this
func marshal(v any) string {
	res, _ := json.Marshal(v)
	return string(res)
}

// generateUnnamedResponseTypes generates names for unnamed response types and adds them to modules
func (p *Parser) generateUnnamedResponseTypes() error {
	if p.config.GenerateUnnamedResponseType == nil {
		return nil
	}

	for _, module := range p.modules {
		for i := range module.HttpHandlers {
			if module.HttpHandlers[i].ResponseBody == nil || module.HttpHandlers[i].ResponseBody.IsNamed {
				continue
			}

			name, gen := p.config.GenerateUnnamedResponseType(&module.HttpHandlers[i])
			if !gen {
				continue
			}

			module.HttpHandlers[i].ResponseBody.Name = name
			module.HttpHandlers[i].ResponseBody.IsNamed = true
			module.Types = append(module.Types, module.HttpHandlers[i].ResponseBody)
		}
	}
	return nil
}

// changeFieldRequirements changes field requirements based on configuration
func (p *Parser) changeFieldRequirements() error {
	if len(p.config.ChangeFields) == 0 {
		return nil
	}

	for typeName, fieldModifications := range p.config.ChangeFields {
		ty, ok := p.namedTypes[typeName]
		if !ok {
			return fmt.Errorf("type %s not found", typeName)
		}

		if ty.Kind != TyKindObject {
			return fmt.Errorf("type %s is not an object type", typeName)
		}

		for fieldName, modification := range fieldModifications {
			if modification == nil {
				continue
			}

			found := false
			for i := range ty.Fields {
				if ty.Fields[i].Name == fieldName {
					// Apply requirement changes
					switch modification.Requirement {
					case FieldRequirementRequired:
						ty.Fields[i].Required = true
					case FieldRequirementOptional:
						ty.Fields[i].Required = false
					}

					// Apply default value if provided
					if modification.Default != "" {
						ty.Fields[i].Default = modification.Default
					}

					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("field %s not found in type %s", fieldName, typeName)
			}
		}
	}
	return nil
}

// orderHandlers orders handlers in modules based on configuration
func (p *Parser) orderHandlers() error {
	if len(p.config.HandlerOrdering) == 0 {
		return nil
	}

	for moduleName, orderedNames := range p.config.HandlerOrdering {
		module, ok := p.modules[moduleName]
		if !ok {
			return fmt.Errorf("module %s not found", moduleName)
		}

		// Create a map for quick lookup of handler positions
		orderMap := make(map[string]int)
		for i, name := range orderedNames {
			orderMap[name] = i
		}

		// Sort handlers based on the order map
		slices.SortStableFunc(module.HttpHandlers, func(a, b HttpHandler) int {
			aPos, aExists := orderMap[a.Name]
			bPos, bExists := orderMap[b.Name]

			// If both handlers are in the order map, use their positions
			if aExists && bExists {
				return aPos - bPos
			}
			// If only one handler is in the order map, it should come first
			if aExists {
				return -1
			}
			if bExists {
				return 1
			}
			// If neither handler is in the order map, sort by name
			return strings.Compare(a.Name, b.Name)
		})
	}
	return nil
}

// changeResponseTypes changes response types based on configuration
func (p *Parser) changeResponseTypes() error {
	for handlerName, responseType := range p.config.ChangeHttpHandlerResponseType {
		newType, ok := p.namedTypes[responseType]
		if !ok {
			return fmt.Errorf("type %s not found for handler %s", responseType, handlerName)
		}

		for _, module := range p.modules {
			for i := range module.HttpHandlers {
				if module.HttpHandlers[i].Name == handlerName {
					module.HttpHandlers[i].ResponseBody = newType
				}
			}
		}
	}
	return nil
}

// renameTypes renames types based on configuration
func (p *Parser) renameTypes() error {
	if len(p.config.RenameTypes) == 0 {
		return nil
	}

	// First check if all target names are available
	for _, newName := range p.config.RenameTypes {
		if _, exists := p.namedTypes[newName]; exists {
			return fmt.Errorf("cannot rename to %s: type already exists", newName)
		}
	}

	// Perform the renaming
	for oldName, newName := range p.config.RenameTypes {
		if ty, exists := p.namedTypes[oldName]; exists {
			// Update the type name
			ty.Name = newName
			// Update the types map
			delete(p.namedTypes, oldName)
			p.namedTypes[newName] = ty
		}
	}
	return nil
}

// renameHandlers renames HTTP handlers based on configuration
func (p *Parser) renameHandlers() error {
	if len(p.config.RenameHandlers) == 0 {
		return nil
	}

	for _, module := range p.modules {
		for i := range module.HttpHandlers {
			if newName, ok := p.config.RenameHandlers[module.HttpHandlers[i].Name]; ok {
				module.HttpHandlers[i].Name = newName
			}
		}
	}
	return nil
}

// ParseOpenAPI parses an OpenAPI document and returns modules
func (p *Parser) ParseOpenAPI(yamlContent []byte) (map[string]*Module, error) {
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

	// Generate names for unnamed response types
	if err := p.generateUnnamedResponseTypes(); err != nil {
		return nil, err
	}

	// Change field requirements based on configuration
	if err := p.changeFieldRequirements(); err != nil {
		return nil, err
	}

	// Order handlers in modules based on configuration
	if err := p.orderHandlers(); err != nil {
		return nil, err
	}

	// Change response types based on configuration
	if err := p.changeResponseTypes(); err != nil {
		return nil, err
	}

	// Rename types based on configuration
	if err := p.renameTypes(); err != nil {
		return nil, err
	}

	// Rename HTTP handlers based on configuration
	if err := p.renameHandlers(); err != nil {
		return nil, err
	}

	// Assign types to modules
	if err := p.assignTypesToModules(); err != nil {
		return nil, err
	}

	return p.modules, nil
}

// GetType returns a type by name
func (p *Parser) GetType(name string) *Ty {
	return p.namedTypes[name]
}

// processNamedTypes processes all named types from components
func (p *Parser) processNamedTypes() error {
	if p.doc.Components == nil || p.doc.Components.Schemas == nil {
		return nil
	}

	for name, schema := range p.doc.Components.Schemas {
		if p.namedTypes[name] != nil {
			continue
		}

		ty, err := p.convertSchema(schema, name, true)
		if err != nil {
			return fmt.Errorf("failed to convert schema %s: %+v err: %w", name, schema, err)
		}
		p.namedTypes[name] = ty
	}
	return nil
}

// processOperations processes all operations and their types
func (p *Parser) processOperations() error {
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
				module = &Module{Name: moduleName}
				p.modules[moduleName] = module
			}

			module.HttpHandlers = append(module.HttpHandlers, *handler)
		}
	}
	return nil
}

// convertSchema converts an OpenAPI schema to our type system
func (p *Parser) convertSchema(schema *openapi3.SchemaRef, name string, isNamed bool) (*Ty, error) {
	if schema == nil {
		return nil, fmt.Errorf("nil schema. s=%v", marshal(schema))
	}
	if schema.Value == nil {
		return nil, fmt.Errorf("nil schema value. v=%v", marshal(schema.Value))
	}

	// If it's a reference and we've already processed it, return the existing type
	if schema.Ref != "" {
		refName := getRefName(schema.Ref)
		if existing := p.namedTypes[refName]; existing != nil {
			return existing, nil
		} else {
			return p.convertSchema(p.doc.Components.Schemas[refName], refName, true)
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
		p.namedTypes[name] = ty
	} else {
		ty.Description = ""
	}

	return ty, nil
}

// convertField converts a schema property to a field
func (p *Parser) convertField(name string, schema *openapi3.SchemaRef, required []string) (*TyField, error) {
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
func (p *Parser) convertOperation(path, method string, op *openapi3.Operation) (*HttpHandler, error) {
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

// topologicalSortTypes performs a deterministic topological sort of types based on their dependencies
// starting from the given entry points (RequestBody/ResponseBody)
func topologicalSortTypes(entryTypes []*Ty) ([]*Ty, error) {
	if len(entryTypes) == 0 {
		return nil, nil
	}

	// Create a directed graph
	g := simple.NewDirectedGraph()

	// Create maps for tracking
	typeToID := make(map[*Ty]int64)
	idToType := make(map[int64]*Ty)
	var nextID int64 = 1

	// Helper function to recursively add types and their dependencies to the graph
	var addTypeToGraph func(*Ty)
	addTypeToGraph = func(ty *Ty) {
		if ty == nil {
			return
		}

		// If we've already processed this type, skip it
		if _, exists := typeToID[ty]; exists {
			return
		}

		// Add the type to the graph
		id := nextID
		nextID++
		typeToID[ty] = id
		idToType[id] = ty
		g.AddNode(simple.Node(id))

		// Recursively process dependencies
		switch ty.Kind {
		case TyKindObject:
			for _, field := range ty.Fields {
				addTypeToGraph(field.Type)
				if field.Type != nil {
					if fieldID, ok := typeToID[field.Type]; ok {
						g.SetEdge(simple.Edge{F: simple.Node(fieldID), T: simple.Node(id)})
					}
				}
				// Handle array element type in object fields
				if field.Type != nil && field.Type.ElementType != nil {
					addTypeToGraph(field.Type.ElementType)
					if elemID, ok := typeToID[field.Type.ElementType]; ok {
						g.SetEdge(simple.Edge{F: simple.Node(elemID), T: simple.Node(id)})
					}
				}
			}
		case TyKindArray:
			if ty.ElementType != nil {
				addTypeToGraph(ty.ElementType)
				if elemID, ok := typeToID[ty.ElementType]; ok {
					g.SetEdge(simple.Edge{F: simple.Node(elemID), T: simple.Node(id)})
				}
			}
		}
	}

	// Add all entry types and their dependencies to the graph
	for _, ty := range entryTypes {
		addTypeToGraph(ty)
	}

	// Perform topological sort
	sorted, err := topo.SortStabilized(g, func(nodes []graph.Node) {
		sort.Slice(nodes, func(i, j int) bool {
			return idToType[nodes[i].ID()].Name < idToType[nodes[j].ID()].Name
		})
	})
	if err != nil {
		return nil, fmt.Errorf("cycle detected in type dependencies: %w", err)
	}

	result := make([]*Ty, 0, len(sorted))
	for _, node := range sorted {
		result = append(result, idToType[node.ID()])
	}

	return result, nil
}

// Replace the topological sort section in assignTypesToModules with a call to topologicalSortTypes
func (p *Parser) assignTypesToModules() error {
	// First, try to assign types based on configuration
	if p.config != nil {
		for typeName, moduleName := range p.config.TypeModuleMap {
			if ty := p.namedTypes[typeName]; ty != nil {
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
	for _, ty := range p.namedTypes {
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

	// Perform topological sort for each module's types
	for _, module := range p.modules {
		if len(module.Types) == 0 {
			continue
		}

		// Collect entry types from handlers
		var entryTypes []*Ty
		for _, h := range module.HttpHandlers {
			if h.RequestBody != nil {
				entryTypes = append(entryTypes, h.RequestBody)
			}
			if h.ResponseBody != nil {
				entryTypes = append(entryTypes, h.ResponseBody)
			}
		}

		// Deduplicate entry types
		slices.SortStableFunc(entryTypes, func(a, b *Ty) int {
			return strings.Compare(fmt.Sprintf("%p", a), fmt.Sprintf("%p", b))
		})
		entryTypes = slices.Compact(entryTypes)

		sortedTypes, err := topologicalSortTypes(entryTypes)
		if err != nil {
			return fmt.Errorf("cycle detected in type dependencies for module %s: %w", module.Name, err)
		}

		// Filter out unnamed types
		namedTypes := slices.DeleteFunc(sortedTypes, func(ty *Ty) bool {
			return !ty.IsNamed
		})

		module.Types = namedTypes
	}

	return nil
}

// isTypeUsedInModule checks if a type is used in a module
func (p *Parser) isTypeUsedInModule(ty *Ty, module *Module, handlerDeps map[*HttpHandler]map[*Ty]bool) bool {
	for i := range module.HttpHandlers {
		if deps := handlerDeps[&module.HttpHandlers[i]]; deps != nil && deps[ty] {
			return true
		}
	}
	return false
}

// collectHandlerTypes recursively collects all types used in a handler
func (p *Parser) collectHandlerTypes(handler *HttpHandler, deps map[*Ty]bool) {
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
func (p *Parser) convertPrimitiveType(typ []string, format string) PrimitiveKind {
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

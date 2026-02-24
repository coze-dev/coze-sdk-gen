package openapi

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Document struct {
	OpenAPI    string              `yaml:"openapi"`
	Swagger    string              `yaml:"swagger"`
	Paths      map[string]PathItem `yaml:"paths"`
	Components Components          `yaml:"components"`
}

type Components struct {
	Schemas       map[string]*Schema      `yaml:"schemas"`
	Parameters    map[string]*Parameter   `yaml:"parameters"`
	Responses     map[string]*Response    `yaml:"responses"`
	RequestBodies map[string]*RequestBody `yaml:"requestBodies"`
}

type PathItem struct {
	Parameters []*Parameter `yaml:"parameters"`
	Get        *Operation   `yaml:"get"`
	Put        *Operation   `yaml:"put"`
	Post       *Operation   `yaml:"post"`
	Delete     *Operation   `yaml:"delete"`
	Patch      *Operation   `yaml:"patch"`
	Options    *Operation   `yaml:"options"`
	Head       *Operation   `yaml:"head"`
	Trace      *Operation   `yaml:"trace"`
}

type Operation struct {
	OperationID string               `yaml:"operationId"`
	Summary     string               `yaml:"summary"`
	Description string               `yaml:"description"`
	Tags        []string             `yaml:"tags"`
	Deprecated  bool                 `yaml:"deprecated"`
	Parameters  []*Parameter         `yaml:"parameters"`
	RequestBody *RequestBody         `yaml:"requestBody"`
	Responses   map[string]*Response `yaml:"responses"`
}

type Parameter struct {
	Ref         string  `yaml:"$ref"`
	Name        string  `yaml:"name"`
	In          string  `yaml:"in"`
	Description string  `yaml:"description"`
	Required    bool    `yaml:"required"`
	Schema      *Schema `yaml:"schema"`
}

type RequestBody struct {
	Ref         string                `yaml:"$ref"`
	Description string                `yaml:"description"`
	Required    bool                  `yaml:"required"`
	Content     map[string]*MediaType `yaml:"content"`
}

type Response struct {
	Ref         string                `yaml:"$ref"`
	Description string                `yaml:"description"`
	Content     map[string]*MediaType `yaml:"content"`
}

type MediaType struct {
	Schema *Schema `yaml:"schema"`
}

type Schema struct {
	Ref                  string             `yaml:"$ref"`
	Type                 string             `yaml:"type"`
	Format               string             `yaml:"format"`
	Title                string             `yaml:"title"`
	Description          string             `yaml:"description"`
	Nullable             bool               `yaml:"nullable"`
	Enum                 []interface{}      `yaml:"enum"`
	Required             []string           `yaml:"required"`
	Properties           map[string]*Schema `yaml:"properties"`
	Items                *Schema            `yaml:"items"`
	AdditionalProperties interface{}        `yaml:"additionalProperties"`
	AllOf                []*Schema          `yaml:"allOf"`
	OneOf                []*Schema          `yaml:"oneOf"`
	AnyOf                []*Schema          `yaml:"anyOf"`
}

type OperationRef struct {
	Path   string
	Method string
}

type ParameterSpec struct {
	Name     string
	In       string
	Required bool
	Schema   *Schema
}

type OperationDetails struct {
	Path                   string
	Method                 string
	Operation              *Operation
	OperationID            string
	Summary                string
	Tags                   []string
	Parameters             []ParameterSpec
	PathParameters         []ParameterSpec
	QueryParameters        []ParameterSpec
	HeaderParameters       []ParameterSpec
	RequestBody            *RequestBody
	RequestBodySchema      *Schema
	RequestBodyContentType string
	Response               *Response
	ResponseSchema         *Schema
	ResponseContentType    string
}

func Load(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi file %q: %w", path, err)
	}
	return Parse(content)
}

func Parse(content []byte) (*Document, error) {
	var doc Document
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse openapi yaml: %w", err)
	}
	if doc.Paths == nil {
		doc.Paths = map[string]PathItem{}
	}
	if doc.Components.Schemas == nil {
		doc.Components.Schemas = map[string]*Schema{}
	}
	if doc.Components.Parameters == nil {
		doc.Components.Parameters = map[string]*Parameter{}
	}
	if doc.Components.Responses == nil {
		doc.Components.Responses = map[string]*Response{}
	}
	if doc.Components.RequestBodies == nil {
		doc.Components.RequestBodies = map[string]*RequestBody{}
	}
	return &doc, nil
}

func (d *Document) HasOperation(method string, path string) bool {
	_, ok := d.Operation(method, path)
	return ok
}

func (d *Document) Operation(method string, path string) (*Operation, bool) {
	if d == nil {
		return nil, false
	}

	item, ok := d.Paths[path]
	if !ok {
		return nil, false
	}

	switch normalizeMethod(method) {
	case "get":
		if item.Get == nil {
			return nil, false
		}
		return item.Get, true
	case "put":
		if item.Put == nil {
			return nil, false
		}
		return item.Put, true
	case "post":
		if item.Post == nil {
			return nil, false
		}
		return item.Post, true
	case "delete":
		if item.Delete == nil {
			return nil, false
		}
		return item.Delete, true
	case "patch":
		if item.Patch == nil {
			return nil, false
		}
		return item.Patch, true
	case "options":
		if item.Options == nil {
			return nil, false
		}
		return item.Options, true
	case "head":
		if item.Head == nil {
			return nil, false
		}
		return item.Head, true
	case "trace":
		if item.Trace == nil {
			return nil, false
		}
		return item.Trace, true
	default:
		return nil, false
	}
}

func (d *Document) ListOperations() []OperationRef {
	if d == nil {
		return nil
	}

	ops := make([]OperationRef, 0)
	for path, item := range d.Paths {
		if item.Get != nil {
			ops = append(ops, OperationRef{Path: path, Method: "get"})
		}
		if item.Put != nil {
			ops = append(ops, OperationRef{Path: path, Method: "put"})
		}
		if item.Post != nil {
			ops = append(ops, OperationRef{Path: path, Method: "post"})
		}
		if item.Delete != nil {
			ops = append(ops, OperationRef{Path: path, Method: "delete"})
		}
		if item.Patch != nil {
			ops = append(ops, OperationRef{Path: path, Method: "patch"})
		}
		if item.Options != nil {
			ops = append(ops, OperationRef{Path: path, Method: "options"})
		}
		if item.Head != nil {
			ops = append(ops, OperationRef{Path: path, Method: "head"})
		}
		if item.Trace != nil {
			ops = append(ops, OperationRef{Path: path, Method: "trace"})
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path == ops[j].Path {
			return ops[i].Method < ops[j].Method
		}
		return ops[i].Path < ops[j].Path
	})
	return ops
}

func (d *Document) PathsWithPrefix(prefix string) []string {
	if d == nil {
		return nil
	}
	paths := make([]string, 0)
	for path := range d.Paths {
		if strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func (d *Document) OperationDetails(path string, method string) (*OperationDetails, bool) {
	if d == nil {
		return nil, false
	}

	operation, ok := d.Operation(method, path)
	if !ok {
		return nil, false
	}

	pathItem, ok := d.Paths[path]
	if !ok {
		return nil, false
	}

	details := &OperationDetails{
		Path:        path,
		Method:      normalizeMethod(method),
		Operation:   operation,
		OperationID: operation.OperationID,
		Summary:     operation.Summary,
		Tags:        append([]string(nil), operation.Tags...),
	}

	paramMap := map[string]ParameterSpec{}
	for _, parameter := range pathItem.Parameters {
		resolved := d.resolveParameter(parameter)
		if resolved == nil {
			continue
		}
		key := resolved.In + ":" + resolved.Name
		paramMap[key] = ParameterSpec{
			Name:     resolved.Name,
			In:       resolved.In,
			Required: resolved.Required,
			Schema:   d.ResolveSchema(resolved.Schema),
		}
	}
	for _, parameter := range operation.Parameters {
		resolved := d.resolveParameter(parameter)
		if resolved == nil {
			continue
		}
		key := resolved.In + ":" + resolved.Name
		paramMap[key] = ParameterSpec{
			Name:     resolved.Name,
			In:       resolved.In,
			Required: resolved.Required,
			Schema:   d.ResolveSchema(resolved.Schema),
		}
	}

	keys := make([]string, 0, len(paramMap))
	for key := range paramMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		param := paramMap[key]
		details.Parameters = append(details.Parameters, param)
		switch param.In {
		case "path":
			details.PathParameters = append(details.PathParameters, param)
		case "query":
			details.QueryParameters = append(details.QueryParameters, param)
		case "header":
			details.HeaderParameters = append(details.HeaderParameters, param)
		}
	}

	requestBody := d.resolveRequestBody(operation.RequestBody)
	if requestBody != nil {
		details.RequestBody = requestBody
		contentType, mediaType := selectContentType(requestBody.Content)
		details.RequestBodyContentType = contentType
		if mediaType != nil {
			details.RequestBodySchema = d.ResolveSchema(mediaType.Schema)
		}
	}

	response := d.resolveResponse(selectResponse(operation.Responses))
	if response != nil {
		details.Response = response
		contentType, mediaType := selectContentType(response.Content)
		details.ResponseContentType = contentType
		if mediaType != nil {
			details.ResponseSchema = d.ResolveSchema(mediaType.Schema)
		}
	}

	return details, true
}

func (d *Document) ListOperationDetails() []OperationDetails {
	ops := d.ListOperations()
	result := make([]OperationDetails, 0, len(ops))
	for _, op := range ops {
		details, ok := d.OperationDetails(op.Path, op.Method)
		if !ok {
			continue
		}
		result = append(result, *details)
	}
	return result
}

func (d *Document) ResolveSchema(schema *Schema) *Schema {
	if d == nil || schema == nil {
		return nil
	}
	if schema.Ref == "" {
		return schema
	}
	resolved, ok := d.ResolveSchemaRef(schema.Ref)
	if !ok {
		return schema
	}
	return resolved
}

func (d *Document) ResolveSchemaRef(ref string) (*Schema, bool) {
	if d == nil {
		return nil, false
	}
	name, ok := refName(ref, "#/components/schemas/")
	if !ok {
		return nil, false
	}
	schema, ok := d.Components.Schemas[name]
	if !ok || schema == nil {
		return nil, false
	}
	return schema, true
}

func (d *Document) SchemaName(schema *Schema) (string, bool) {
	if schema == nil {
		return "", false
	}
	if schema.Ref != "" {
		return refName(schema.Ref, "#/components/schemas/")
	}
	if d != nil {
		for name, candidate := range d.Components.Schemas {
			if candidate == schema {
				return name, true
			}
		}
	}
	return "", false
}

func (d *Document) CollectSchemaRefsFromOperation(operation OperationDetails) []string {
	refs := map[string]struct{}{}
	visit := func(schema *Schema) {}

	var walk func(schema *Schema)
	walk = func(schema *Schema) {
		if schema == nil {
			return
		}
		if name, ok := d.SchemaName(schema); ok {
			if _, exists := refs[name]; exists {
				return
			}
			refs[name] = struct{}{}
			resolved, ok := d.ResolveSchemaRef("#/components/schemas/" + name)
			if ok {
				walk(resolved)
			}
			return
		}

		for _, property := range schema.Properties {
			walk(d.ResolveSchema(property))
		}
		walk(d.ResolveSchema(schema.Items))
		for _, item := range schema.AllOf {
			walk(d.ResolveSchema(item))
		}
		for _, item := range schema.AnyOf {
			walk(d.ResolveSchema(item))
		}
		for _, item := range schema.OneOf {
			walk(d.ResolveSchema(item))
		}
		switch ap := schema.AdditionalProperties.(type) {
		case map[string]interface{}:
			raw, err := yaml.Marshal(ap)
			if err == nil {
				var extra Schema
				if yaml.Unmarshal(raw, &extra) == nil {
					walk(d.ResolveSchema(&extra))
				}
			}
		}
	}

	visit = walk
	for _, param := range operation.Parameters {
		visit(d.ResolveSchema(param.Schema))
	}
	visit(d.ResolveSchema(operation.RequestBodySchema))
	visit(d.ResolveSchema(operation.ResponseSchema))

	names := make([]string, 0, len(refs))
	for name := range refs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (d *Document) resolveParameter(parameter *Parameter) *Parameter {
	if parameter == nil {
		return nil
	}
	if parameter.Ref == "" {
		return parameter
	}
	name, ok := refName(parameter.Ref, "#/components/parameters/")
	if !ok {
		return parameter
	}
	resolved, ok := d.Components.Parameters[name]
	if !ok {
		return parameter
	}
	return resolved
}

func (d *Document) resolveRequestBody(body *RequestBody) *RequestBody {
	if body == nil {
		return nil
	}
	if body.Ref == "" {
		return body
	}
	name, ok := refName(body.Ref, "#/components/requestBodies/")
	if !ok {
		return body
	}
	resolved, ok := d.Components.RequestBodies[name]
	if !ok {
		return body
	}
	return resolved
}

func (d *Document) resolveResponse(response *Response) *Response {
	if response == nil {
		return nil
	}
	if response.Ref == "" {
		return response
	}
	name, ok := refName(response.Ref, "#/components/responses/")
	if !ok {
		return response
	}
	resolved, ok := d.Components.Responses[name]
	if !ok {
		return response
	}
	return resolved
}

func selectResponse(responses map[string]*Response) *Response {
	if len(responses) == 0 {
		return nil
	}
	priority := []string{"200", "201", "202", "203", "204", "default"}
	for _, key := range priority {
		if response, ok := responses[key]; ok {
			return response
		}
	}

	codes := make([]int, 0)
	for status := range responses {
		value, err := strconv.Atoi(status)
		if err == nil {
			codes = append(codes, value)
		}
	}
	if len(codes) > 0 {
		sort.Ints(codes)
		return responses[strconv.Itoa(codes[0])]
	}

	keys := make([]string, 0, len(responses))
	for key := range responses {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return responses[keys[0]]
}

func selectContentType(content map[string]*MediaType) (string, *MediaType) {
	if len(content) == 0 {
		return "", nil
	}
	priority := []string{"application/json", "text/event-stream", "*/*"}
	for _, key := range priority {
		if mediaType, ok := content[key]; ok {
			return key, mediaType
		}
	}

	keys := make([]string, 0, len(content))
	for key := range content {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	first := keys[0]
	return first, content[first]
}

func refName(ref string, prefix string) (string, bool) {
	if !strings.HasPrefix(ref, prefix) {
		return "", false
	}
	name := strings.TrimPrefix(ref, prefix)
	if name == "" {
		return "", false
	}
	return name, true
}

func normalizeMethod(method string) string {
	return strings.ToLower(strings.TrimSpace(method))
}

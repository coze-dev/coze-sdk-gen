package python

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type packageModelDefinition struct {
	SchemaName            string
	Name                  string
	BaseClasses           []string
	Schema                *openapi.Schema
	IsEnum                bool
	BeforeCode            []string
	PrependCode           []string
	Builders              []config.ModelBuilder
	BeforeValidators      []config.ModelValidator
	FieldOrder            []string
	RequiredFields        []string
	FieldTypes            map[string]string
	FieldDefaults         map[string]string
	EnumBase              string
	EnumValues            []config.ModelEnumValue
	ExtraFields           []config.ModelField
	ExtraCode             []string
	AllowMissingInSwagger bool
	ExcludeUnordered      bool
}

func packageSchemaAliases(doc *openapi.Document, meta PackageMeta) map[string]string {
	aliases := map[string]string{}
	if meta.Package == nil {
		return aliases
	}
	for _, model := range meta.Package.ModelSchemas {
		schemaName := strings.TrimSpace(model.Schema)
		modelName := inferConfiguredModelName(doc, meta.Package, model)
		if schemaName == "" || modelName == "" {
			continue
		}
		aliases[schemaName] = modelName
	}
	return aliases
}

func resolvePackageModelDefinitions(doc *openapi.Document, meta PackageMeta, bindings []OperationBinding) ([]packageModelDefinition, map[string]string) {
	if meta.Package == nil || doc == nil {
		return nil, nil
	}
	inferredAliases := map[string]string{}
	result := make([]packageModelDefinition, 0, len(meta.Package.ModelSchemas))
	includedSchemaNames := map[string]struct{}{}
	usedModelNames := map[string]struct{}{}
	modelSignatures := map[string]string{}
	for _, model := range meta.Package.ModelSchemas {
		definition, ok := resolveConfiguredModelDefinition(doc, meta.Package, model)
		if !ok {
			continue
		}
		result = append(result, definition)
		usedModelNames[definition.Name] = struct{}{}
		if definition.Schema != nil {
			modelSignatures[definition.Name] = schemaStructuralSignature(doc, definition.Schema)
		}
		schemaName := strings.TrimSpace(definition.SchemaName)
		if schemaName != "" {
			includedSchemaNames[schemaName] = struct{}{}
			inferredAliases[schemaName] = definition.Name
		}
	}
	if len(meta.Package.ModelSchemas) == 0 {
		autoSeeds := inferOperationRootModels(doc, meta.Package, bindings)
		for _, definition := range autoSeeds {
			result = append(result, definition)
			usedModelNames[definition.Name] = struct{}{}
			if definition.Schema != nil {
				modelSignatures[definition.Name] = schemaStructuralSignature(doc, definition.Schema)
			}
			schemaName := strings.TrimSpace(definition.SchemaName)
			if schemaName != "" {
				includedSchemaNames[schemaName] = struct{}{}
				inferredAliases[schemaName] = definition.Name
			}
		}
	}
	for i := 0; i < len(result); i++ {
		model := result[i]
		if model.Schema == nil {
			continue
		}
		referencedSchemaNames := collectModelSchemaRefs(doc, model)
		for _, schemaName := range referencedSchemaNames {
			schemaName = strings.TrimSpace(schemaName)
			if schemaName == "" || schemaName == strings.TrimSpace(model.SchemaName) {
				continue
			}
			if _, exists := includedSchemaNames[schemaName]; exists {
				continue
			}
			schema, exists := doc.Components.Schemas[schemaName]
			if !exists || schema == nil {
				continue
			}
			resolved := doc.ResolveSchema(schema)
			if resolved == nil {
				continue
			}
			includedSchemaNames[schemaName] = struct{}{}
			next := packageModelDefinition{
				SchemaName:    schemaName,
				Name:          inferModelNameFromSchema(meta.Package, schemaName, resolved),
				Schema:        resolved,
				IsEnum:        isSchemaEnum(resolved, nil),
				FieldTypes:    map[string]string{},
				FieldDefaults: map[string]string{},
			}
			if next.Name == "" {
				continue
			}
			nameSignature := schemaStructuralSignature(doc, next.Schema)
			if existingSignature, exists := modelSignatures[next.Name]; exists {
				if existingSignature == nameSignature {
					inferredAliases[schemaName] = next.Name
					continue
				}
				next.Name = nextAvailableModelName(next.Name, usedModelNames)
			}
			usedModelNames[next.Name] = struct{}{}
			modelSignatures[next.Name] = nameSignature
			result = append(result, next)
			inferredAliases[schemaName] = next.Name
		}
	}
	return orderModelDefinitionsByDependencies(doc, result), inferredAliases
}

func resolveConfiguredModelDefinition(doc *openapi.Document, pkg *config.Package, model config.ModelSchema) (packageModelDefinition, bool) {
	schemaName := strings.TrimSpace(model.Schema)
	modelName := inferConfiguredModelName(doc, pkg, model)
	if modelName == "" {
		return packageModelDefinition{}, false
	}
	fieldTypes := map[string]string{}
	for k, v := range model.FieldTypes {
		fieldTypes[k] = v
	}
	fieldDefaults := map[string]string{}
	for k, v := range model.FieldDefaults {
		fieldDefaults[k] = v
	}
	enumValues := append([]config.ModelEnumValue(nil), model.EnumValues...)
	definition := packageModelDefinition{
		SchemaName:            schemaName,
		Name:                  modelName,
		BaseClasses:           append([]string(nil), model.BaseClasses...),
		Schema:                nil,
		IsEnum:                false,
		BeforeCode:            append([]string(nil), model.BeforeCode...),
		PrependCode:           append([]string(nil), model.PrependCode...),
		Builders:              append([]config.ModelBuilder(nil), model.Builders...),
		BeforeValidators:      append([]config.ModelValidator(nil), model.BeforeValidators...),
		FieldOrder:            append([]string(nil), model.FieldOrder...),
		RequiredFields:        append([]string(nil), model.RequiredFields...),
		FieldTypes:            fieldTypes,
		FieldDefaults:         fieldDefaults,
		EnumBase:              strings.TrimSpace(model.EnumBase),
		EnumValues:            enumValues,
		ExtraFields:           append([]config.ModelField(nil), model.ExtraFields...),
		ExtraCode:             append([]string(nil), model.ExtraCode...),
		AllowMissingInSwagger: model.AllowMissingInSwagger,
		ExcludeUnordered:      model.ExcludeUnorderedFields,
	}

	if schemaName == "" {
		if !model.AllowMissingInSwagger {
			return packageModelDefinition{}, false
		}
		definition.IsEnum = isSchemaEnum(nil, enumValues)
		return definition, true
	}
	schema, ok := doc.Components.Schemas[schemaName]
	if !ok || schema == nil {
		if !model.AllowMissingInSwagger {
			return packageModelDefinition{}, false
		}
		definition.IsEnum = isSchemaEnum(nil, enumValues)
		return definition, true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved == nil {
		if !model.AllowMissingInSwagger {
			return packageModelDefinition{}, false
		}
		definition.IsEnum = isSchemaEnum(nil, enumValues)
		return definition, true
	}
	definition.Schema = resolved
	definition.IsEnum = isSchemaEnum(resolved, enumValues)
	return definition, true
}

func inferConfiguredModelName(doc *openapi.Document, pkg *config.Package, model config.ModelSchema) string {
	modelName := strings.TrimSpace(model.Name)
	if modelName != "" {
		return modelName
	}
	schemaName := strings.TrimSpace(model.Schema)
	if schemaName == "" {
		return ""
	}
	return inferModelNameFromSchema(pkg, schemaName, resolvedSchemaByName(doc, schemaName))
}

func inferModelNameFromSchema(pkg *config.Package, schemaName string, schema *openapi.Schema) string {
	candidate, addPackagePrefix := inferModelNameCandidate(schemaName, schema)
	if strings.TrimSpace(candidate) == "" {
		candidate = strings.TrimSpace(schemaName)
	}
	if addPackagePrefix {
		packagePrefix := packageModelPrefix(pkg)
		if packagePrefix != "" && !strings.HasPrefix(candidate, packagePrefix+"_") {
			candidate = packagePrefix + "_" + candidate
		}
	}
	return NormalizeClassName(candidate)
}

func inferModelNameCandidate(schemaName string, schema *openapi.Schema) (string, bool) {
	trimmed := strings.TrimSpace(schemaName)
	if trimmed == "" {
		return "", false
	}
	if schemaHasAllProperties(schema, "status", "item_info") {
		return "status_info", true
	}
	if schemaHasAllProperties(schema, "used", "total", "start_at", "end_at", "strategy") {
		return "item_info", true
	}
	if schemaHasAllProperties(schema, "user_level") && len(schema.Properties) == 1 {
		return "basic_info", true
	}

	candidate := trimmed
	isSynthetic := false
	if strings.HasPrefix(candidate, "properties_data_properties_") {
		candidate = strings.TrimPrefix(candidate, "properties_data_properties_")
		isSynthetic = true
	}
	if strings.HasPrefix(candidate, "properties_") {
		candidate = strings.TrimPrefix(candidate, "properties_")
		isSynthetic = true
	}
	if strings.HasSuffix(candidate, "_properties_item_info") {
		return "item_info", true
	}
	if strings.HasSuffix(candidate, "_basic_info") {
		return "basic_info", true
	}
	if strings.HasSuffix(candidate, "benefit_info_items") {
		return "info", true
	}
	if isSynthetic {
		candidate = strings.ReplaceAll(candidate, "_properties_", "_")
		candidate = strings.ReplaceAll(candidate, "_items_", "_")
		candidate = strings.Trim(candidate, "_")
		return candidate, true
	}
	return candidate, false
}

func resolvedSchemaByName(doc *openapi.Document, schemaName string) *openapi.Schema {
	if doc == nil {
		return nil
	}
	schema, ok := doc.Components.Schemas[strings.TrimSpace(schemaName)]
	if !ok || schema == nil {
		return nil
	}
	return doc.ResolveSchema(schema)
}

func schemaHasAllProperties(schema *openapi.Schema, propertyNames ...string) bool {
	if schema == nil || len(propertyNames) == 0 {
		return false
	}
	for _, name := range propertyNames {
		if schema.Properties == nil || schema.Properties[name] == nil {
			return false
		}
	}
	return true
}

func packageModelPrefix(pkg *config.Package) string {
	if pkg == nil {
		return ""
	}
	name := strings.Trim(ToSnake(strings.TrimSpace(pkg.Name)), "_")
	if name == "" {
		return ""
	}
	switch {
	case strings.HasSuffix(name, "ies") && len(name) > 3:
		return strings.TrimSuffix(name, "ies") + "y"
	case strings.HasSuffix(name, "sses") && len(name) > 4:
		return strings.TrimSuffix(name, "es")
	case strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") && len(name) > 1:
		return strings.TrimSuffix(name, "s")
	default:
		return name
	}
}

func inferOperationRootModels(doc *openapi.Document, pkg *config.Package, bindings []OperationBinding) []packageModelDefinition {
	if doc == nil || pkg == nil || len(bindings) == 0 {
		return nil
	}
	out := make([]packageModelDefinition, 0, len(bindings))
	seenSchema := map[string]struct{}{}
	emptyModelSet := map[string]struct{}{}
	for _, modelName := range pkg.EmptyModels {
		trimmed := strings.TrimSpace(modelName)
		if trimmed == "" {
			continue
		}
		emptyModelSet[trimmed] = struct{}{}
	}
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		modelName := strings.TrimSpace(binding.Mapping.ResponseType)
		if !isPythonClassName(modelName) {
			inferredName, ok := inferBindingResponseModelName(doc, pkg, binding)
			if !ok {
				continue
			}
			modelName = inferredName
		}
		if _, exists := emptyModelSet[modelName]; exists {
			continue
		}
		responseDataSchema, schemaName, ok := resolveMappingResponseDataSchema(doc, binding)
		if !ok {
			continue
		}
		schemaName = strings.TrimSpace(schemaName)
		if schemaName == "" {
			continue
		}
		if _, exists := seenSchema[schemaName]; exists {
			continue
		}
		schema := doc.ResolveSchema(responseDataSchema)
		if schema == nil {
			continue
		}
		seenSchema[schemaName] = struct{}{}
		out = append(out, packageModelDefinition{
			SchemaName:    schemaName,
			Name:          modelName,
			Schema:        schema,
			IsEnum:        isSchemaEnum(schema, nil),
			FieldTypes:    map[string]string{},
			FieldDefaults: map[string]string{},
		})
	}
	return out
}

func inferBindingResponseModelName(doc *openapi.Document, pkg *config.Package, binding OperationBinding) (string, bool) {
	responseDataSchema, schemaName, ok := resolveMappingResponseDataSchema(doc, binding)
	if !ok {
		return "", false
	}
	modelName := strings.TrimSpace(inferModelNameFromSchema(pkg, schemaName, responseDataSchema))
	if !isPythonClassName(modelName) {
		return "", false
	}
	return modelName, true
}

func resolveMappingResponseDataSchema(doc *openapi.Document, binding OperationBinding) (*openapi.Schema, string, bool) {
	if doc == nil {
		return nil, "", false
	}
	responseSchema := doc.ResolveSchema(binding.Details.ResponseSchema)
	if responseSchema == nil {
		return nil, "", false
	}
	dataField := "data"
	if binding.Mapping != nil {
		if configured := strings.TrimSpace(binding.Mapping.DataField); configured != "" {
			dataField = configured
		}
	}
	if dataField != "" && responseSchema.Properties != nil {
		if fieldSchema, ok := responseSchema.Properties[dataField]; ok && fieldSchema != nil {
			if name, ok := doc.SchemaName(fieldSchema); ok {
				return doc.ResolveSchema(fieldSchema), name, true
			}
			if resolved := doc.ResolveSchema(fieldSchema); resolved != nil {
				if name, ok := doc.SchemaName(resolved); ok {
					return resolved, name, true
				}
			}
		}
	}
	if name, ok := doc.SchemaName(responseSchema); ok {
		return responseSchema, name, true
	}
	return nil, "", false
}

func isPythonClassName(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	runes := []rune(value)
	if !(unicode.IsLetter(runes[0]) || runes[0] == '_') {
		return false
	}
	for _, r := range runes[1:] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

func nextAvailableModelName(base string, used map[string]struct{}) string {
	trimmed := strings.TrimSpace(base)
	if trimmed == "" {
		trimmed = "GeneratedModel"
	}
	if _, exists := used[trimmed]; !exists {
		return trimmed
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s%d", trimmed, i)
		if _, exists := used[candidate]; !exists {
			return candidate
		}
	}
}

func schemaStructuralSignature(doc *openapi.Document, schema *openapi.Schema) string {
	if doc == nil || schema == nil {
		return ""
	}
	visiting := map[*openapi.Schema]struct{}{}
	var encode func(*openapi.Schema) string
	encode = func(current *openapi.Schema) string {
		resolved := doc.ResolveSchema(current)
		if resolved == nil {
			return "nil"
		}
		if _, seen := visiting[resolved]; seen {
			return "cycle"
		}
		visiting[resolved] = struct{}{}
		defer delete(visiting, resolved)

		var buf strings.Builder
		buf.WriteString("type=")
		buf.WriteString(strings.TrimSpace(resolved.Type))
		buf.WriteString(";format=")
		buf.WriteString(strings.TrimSpace(resolved.Format))

		if len(resolved.Required) > 0 {
			required := append([]string(nil), resolved.Required...)
			sort.Strings(required)
			buf.WriteString(";required=")
			buf.WriteString(strings.Join(required, ","))
		}
		if len(resolved.Enum) > 0 {
			values := make([]string, 0, len(resolved.Enum))
			for _, item := range resolved.Enum {
				values = append(values, fmt.Sprintf("%v", item))
			}
			sort.Strings(values)
			buf.WriteString(";enum=")
			buf.WriteString(strings.Join(values, ","))
		}
		if resolved.Items != nil {
			buf.WriteString(";items=(")
			buf.WriteString(encode(resolved.Items))
			buf.WriteString(")")
		}

		propNames := make([]string, 0, len(resolved.Properties))
		for name := range resolved.Properties {
			propNames = append(propNames, name)
		}
		sort.Strings(propNames)
		for _, name := range propNames {
			buf.WriteString(";prop:")
			buf.WriteString(name)
			buf.WriteString("=")
			buf.WriteString(encode(resolved.Properties[name]))
		}

		if additional, ok := resolved.AdditionalProperties.(*openapi.Schema); ok {
			buf.WriteString(";additional=(")
			buf.WriteString(encode(additional))
			buf.WriteString(")")
		}

		appendComposed := func(tag string, list []*openapi.Schema) {
			if len(list) == 0 {
				return
			}
			signatures := make([]string, 0, len(list))
			for _, item := range list {
				signatures = append(signatures, encode(item))
			}
			sort.Strings(signatures)
			buf.WriteString(";")
			buf.WriteString(tag)
			buf.WriteString("=(")
			buf.WriteString(strings.Join(signatures, "|"))
			buf.WriteString(")")
		}
		appendComposed("allOf", resolved.AllOf)
		appendComposed("anyOf", resolved.AnyOf)
		appendComposed("oneOf", resolved.OneOf)

		return buf.String()
	}
	return encode(schema)
}

func isSchemaEnum(schema *openapi.Schema, explicitEnumValues []config.ModelEnumValue) bool {
	if len(explicitEnumValues) > 0 {
		return true
	}
	if schema == nil {
		return false
	}
	return len(schema.Enum) > 0 && (schema.Type == "string" || schema.Type == "integer" || schema.Type == "")
}

func collectModelSchemaRefs(doc *openapi.Document, model packageModelDefinition) []string {
	if doc == nil || model.Schema == nil {
		return nil
	}
	resolvedRoot := doc.ResolveSchema(model.Schema)
	if resolvedRoot == nil {
		return nil
	}
	rootName, _ := doc.SchemaName(resolvedRoot)

	refs := map[string]struct{}{}
	visited := map[*openapi.Schema]struct{}{}
	var walk func(*openapi.Schema)
	walk = func(current *openapi.Schema) {
		resolved := doc.ResolveSchema(current)
		if resolved == nil {
			return
		}
		if _, ok := visited[resolved]; ok {
			return
		}
		visited[resolved] = struct{}{}
		if name, ok := doc.SchemaName(current); ok {
			refs[name] = struct{}{}
		} else if name, ok := doc.SchemaName(resolved); ok {
			refs[name] = struct{}{}
		}
		for _, property := range resolved.Properties {
			walk(property)
		}
		walk(resolved.Items)
		for _, item := range resolved.AllOf {
			walk(item)
		}
		for _, item := range resolved.AnyOf {
			walk(item)
		}
		for _, item := range resolved.OneOf {
			walk(item)
		}
		if additional, ok := resolved.AdditionalProperties.(*openapi.Schema); ok {
			walk(additional)
		}
	}

	propertyNames := modelRenderedPropertyNames(resolvedRoot.Properties, model.FieldOrder, model.ExcludeUnordered)
	for _, propertyName := range propertyNames {
		property := resolvedRoot.Properties[propertyName]
		if property == nil {
			continue
		}
		override := strings.TrimSpace(model.FieldTypes[propertyName])
		if override != "" {
			continue
		}
		walk(property)
	}

	names := make([]string, 0, len(refs))
	for name := range refs {
		name = strings.TrimSpace(name)
		if name == "" || name == rootName {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func modelRenderedPropertyNames(properties map[string]*openapi.Schema, fieldOrder []string, excludeUnordered bool) []string {
	if len(properties) == 0 {
		return nil
	}
	fieldNames := make([]string, 0, len(properties))
	seen := map[string]struct{}{}
	for _, rawName := range fieldOrder {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		if _, ok := properties[name]; !ok {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		fieldNames = append(fieldNames, name)
	}
	if excludeUnordered {
		return fieldNames
	}
	remaining := make([]string, 0, len(properties))
	for name := range properties {
		if _, exists := seen[name]; exists {
			continue
		}
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)
	return append(fieldNames, remaining...)
}

func orderModelDefinitionsByDependencies(doc *openapi.Document, models []packageModelDefinition) []packageModelDefinition {
	if len(models) < 2 {
		return models
	}
	indexBySchema := map[string]int{}
	for i, model := range models {
		schemaName := strings.TrimSpace(model.SchemaName)
		if schemaName == "" {
			continue
		}
		indexBySchema[schemaName] = i
	}
	dependencies := make([][]int, len(models))
	for i, model := range models {
		if model.Schema == nil {
			continue
		}
		depNames := collectModelSchemaRefs(doc, model)
		depIndexes := make([]int, 0, len(depNames))
		seen := map[int]struct{}{}
		for _, depName := range depNames {
			depIdx, ok := indexBySchema[depName]
			if !ok {
				continue
			}
			if depIdx == i {
				continue
			}
			if _, exists := seen[depIdx]; exists {
				continue
			}
			seen[depIdx] = struct{}{}
			depIndexes = append(depIndexes, depIdx)
		}
		sort.Ints(depIndexes)
		dependencies[i] = depIndexes
	}

	state := make([]int, len(models))
	ordered := make([]packageModelDefinition, 0, len(models))
	var visit func(index int)
	visit = func(index int) {
		if state[index] == 2 {
			return
		}
		if state[index] == 1 {
			return
		}
		state[index] = 1
		for _, depIndex := range dependencies[index] {
			visit(depIndex)
		}
		state[index] = 2
		ordered = append(ordered, models[index])
	}
	for i := range models {
		visit(i)
	}
	return ordered
}

func renderPackageModelDefinitions(
	doc *openapi.Document,
	meta PackageMeta,
	models []packageModelDefinition,
	schemaAliases map[string]string,
	commentOverrides config.CommentOverrides,
) string {
	var buf bytes.Buffer
	modulePrefix := "cozepy." + meta.ModulePath

	for _, model := range models {
		classKey := modulePrefix + "." + model.Name
		if len(model.BeforeCode) > 0 {
			for _, block := range model.BeforeCode {
				AppendIndentedCode(&buf, block, 0)
				buf.WriteString("\n")
			}
		}
		if model.IsEnum {
			if model.EnumBase == "dynamic_str" {
				buf.WriteString(fmt.Sprintf("class %s(DynamicStrEnum):\n", model.Name))
			} else if model.EnumBase == "int" {
				buf.WriteString(fmt.Sprintf("class %s(IntEnum):\n", model.Name))
			} else if model.EnumBase == "int_enum" {
				buf.WriteString(fmt.Sprintf("class %s(int, Enum):\n", model.Name))
			} else {
				buf.WriteString(fmt.Sprintf("class %s(str, Enum):\n", model.Name))
			}
			if !modelHasCustomClassDocstring(model) {
				if docstring := strings.TrimSpace(schemaDescription(model.Schema)); docstring != "" {
					WriteClassDocstring(&buf, 1, docstring, "block")
				} else if overrideDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[classKey]); overrideDoc != "" {
					style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[classKey])
					WriteClassDocstring(&buf, 1, overrideDoc, style)
				}
			}
			enumItems := make([]config.ModelEnumValue, 0)
			if len(model.EnumValues) > 0 {
				enumItems = append(enumItems, model.EnumValues...)
			} else if model.Schema != nil && len(model.Schema.Enum) > 0 {
				for _, enumValue := range model.Schema.Enum {
					enumItems = append(enumItems, config.ModelEnumValue{
						Name:  EnumMemberName(fmt.Sprintf("%v", enumValue)),
						Value: enumValue,
					})
				}
			}
			if len(enumItems) == 0 {
				buf.WriteString("    pass\n\n")
				continue
			}
			for _, enumValue := range enumItems {
				memberName := strings.TrimSpace(enumValue.Name)
				if memberName == "" {
					memberName = EnumMemberName(fmt.Sprintf("%v", enumValue.Value))
				}
				inlineEnumComment := strings.TrimSpace(commentOverrides.InlineEnumMemberComment[classKey+"."+memberName])
				if inlineEnumComment != "" {
					inlineEnumComment = strings.TrimPrefix(inlineEnumComment, "#")
					inlineEnumComment = strings.TrimSpace(inlineEnumComment)
				}
				enumComment := LinesFromCommentOverride(commentOverrides.EnumMemberComments[classKey+"."+memberName])
				if len(enumComment) > 0 && inlineEnumComment == "" {
					WriteLineComments(&buf, 1, enumComment)
				}
				if inlineEnumComment != "" {
					buf.WriteString(fmt.Sprintf("    %s = %s  # %s\n", memberName, RenderEnumValueLiteral(enumValue.Value), inlineEnumComment))
				} else {
					buf.WriteString(fmt.Sprintf("    %s = %s\n", memberName, RenderEnumValueLiteral(enumValue.Value)))
				}
			}
			buf.WriteString("\n\n")
			continue
		}

		baseExpr := "CozeModel"
		if len(model.BaseClasses) > 0 {
			baseList := make([]string, 0, len(model.BaseClasses))
			for _, baseClass := range model.BaseClasses {
				trimmed := strings.TrimSpace(baseClass)
				if trimmed == "" {
					continue
				}
				baseList = append(baseList, trimmed)
			}
			if len(baseList) > 0 {
				baseExpr = strings.Join(baseList, ", ")
			}
		}
		buf.WriteString(fmt.Sprintf("class %s(%s):\n", model.Name, baseExpr))
		hasClassDocstring := false
		if !modelHasCustomClassDocstring(model) {
			if docstring := strings.TrimSpace(schemaDescription(model.Schema)); docstring != "" {
				WriteClassDocstring(&buf, 1, docstring, "block")
				hasClassDocstring = true
			} else if overrideDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[classKey]); overrideDoc != "" {
				style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[classKey])
				WriteClassDocstring(&buf, 1, overrideDoc, style)
				hasClassDocstring = true
			}
		}
		properties := map[string]*openapi.Schema{}
		if model.Schema != nil {
			properties = model.Schema.Properties
		}
		if len(properties) == 0 {
			if len(model.PrependCode) == 0 &&
				len(model.ExtraFields) == 0 &&
				len(model.ExtraCode) == 0 &&
				len(model.BeforeValidators) == 0 &&
				len(model.Builders) == 0 {
				if !hasClassDocstring {
					buf.WriteString("    pass\n")
				}
				buf.WriteString("\n")
				continue
			}
		}
		for _, block := range model.PrependCode {
			AppendIndentedCode(&buf, block, 1)
		}

		requiredSet := map[string]bool{}
		if model.Schema != nil {
			for _, requiredName := range model.Schema.Required {
				requiredSet[requiredName] = true
			}
		}
		clearRequired := false
		configRequired := make([]string, 0, len(model.RequiredFields))
		for _, requiredName := range model.RequiredFields {
			requiredName = strings.TrimSpace(requiredName)
			if requiredName == "" {
				continue
			}
			if requiredName == "__none__" {
				clearRequired = true
				continue
			}
			configRequired = append(configRequired, requiredName)
		}
		if clearRequired {
			requiredSet = map[string]bool{}
		}
		for _, requiredName := range configRequired {
			requiredSet[requiredName] = true
		}
		extraFieldByName := map[string]config.ModelField{}
		extraFieldNames := make([]string, 0, len(model.ExtraFields))
		for _, extraField := range model.ExtraFields {
			rawName := strings.TrimSpace(extraField.Name)
			if rawName == "" {
				continue
			}
			if _, exists := extraFieldByName[rawName]; exists {
				continue
			}
			extraFieldByName[rawName] = extraField
			extraFieldNames = append(extraFieldNames, rawName)
		}

		fieldNames := make([]string, 0, len(properties)+len(extraFieldNames))
		seenFields := map[string]bool{}
		for _, fieldName := range model.FieldOrder {
			if _, ok := properties[fieldName]; ok {
				fieldNames = append(fieldNames, fieldName)
				seenFields[fieldName] = true
				continue
			}
			if _, ok := extraFieldByName[fieldName]; ok {
				if _, exists := properties[fieldName]; exists {
					continue
				}
				fieldNames = append(fieldNames, fieldName)
				seenFields[fieldName] = true
			}
		}
		if !model.ExcludeUnordered {
			remaining := make([]string, 0, len(properties))
			for propertyName := range properties {
				if seenFields[propertyName] {
					continue
				}
				remaining = append(remaining, propertyName)
			}
			sort.Strings(remaining)
			fieldNames = append(fieldNames, remaining...)
		}
		for _, extraName := range extraFieldNames {
			if _, exists := properties[extraName]; exists {
				continue
			}
			if seenFields[extraName] {
				continue
			}
			fieldNames = append(fieldNames, extraName)
		}

		hasRenderedField := false
		for _, fieldName := range fieldNames {
			if propertySchema, ok := properties[fieldName]; ok {
				typeName := modelFieldType(model, fieldName, PythonTypeForSchemaWithAliases(doc, propertySchema, requiredSet[fieldName], schemaAliases))
				normalizedFieldName := NormalizePythonIdentifier(fieldName)
				inlineFieldComment := ""
				fieldComment := schemaCommentLines(doc, propertySchema)
				if len(fieldComment) == 0 {
					inlineFieldComment = strings.TrimSpace(commentOverrides.InlineFieldComments[classKey+"."+normalizedFieldName])
					if inlineFieldComment != "" {
						inlineFieldComment = strings.TrimPrefix(inlineFieldComment, "#")
						inlineFieldComment = strings.TrimSpace(inlineFieldComment)
					}
					fieldComment = LinesFromCommentOverride(commentOverrides.FieldComments[classKey+"."+normalizedFieldName])
				}
				if len(fieldComment) > 0 && inlineFieldComment == "" {
					WriteLineComments(&buf, 1, fieldComment)
				}
				if requiredSet[fieldName] {
					if inlineFieldComment != "" {
						buf.WriteString(fmt.Sprintf("    %s: %s  # %s\n", normalizedFieldName, typeName, inlineFieldComment))
					} else {
						buf.WriteString(fmt.Sprintf("    %s: %s\n", normalizedFieldName, typeName))
					}
					hasRenderedField = true
				} else {
					defaultValue := modelFieldDefault(model, fieldName)
					if defaultValue == "None" && !strings.HasPrefix(typeName, "Optional[") {
						typeName = "Optional[" + typeName + "]"
					} else if defaultValue != "None" {
						if !hasModelFieldTypeOverride(model, fieldName) {
							typeName = unwrapOptionalType(typeName)
						}
					}
					if inlineFieldComment != "" {
						buf.WriteString(fmt.Sprintf("    %s: %s = %s  # %s\n", normalizedFieldName, typeName, defaultValue, inlineFieldComment))
					} else {
						buf.WriteString(fmt.Sprintf("    %s: %s = %s\n", normalizedFieldName, typeName, defaultValue))
					}
					hasRenderedField = true
				}
				continue
			}

			extraField, ok := extraFieldByName[fieldName]
			if !ok {
				continue
			}
			if _, exists := properties[fieldName]; exists {
				continue
			}
			normalizedFieldName := NormalizePythonIdentifier(fieldName)
			typeName := strings.TrimSpace(extraField.Type)
			if typeName == "" {
				typeName = "Any"
			}
			inlineFieldComment := strings.TrimSpace(commentOverrides.InlineFieldComments[classKey+"."+normalizedFieldName])
			if inlineFieldComment != "" {
				inlineFieldComment = strings.TrimPrefix(inlineFieldComment, "#")
				inlineFieldComment = strings.TrimSpace(inlineFieldComment)
			}
			fieldComment := LinesFromCommentOverride(commentOverrides.FieldComments[classKey+"."+normalizedFieldName])
			if len(fieldComment) > 0 && inlineFieldComment == "" {
				WriteLineComments(&buf, 1, fieldComment)
			}
			alias := strings.TrimSpace(extraField.Alias)
			if extraField.Required {
				expr := ""
				if alias != "" {
					expr = fmt.Sprintf("Field(alias=%q)", alias)
				}
				if inlineFieldComment != "" {
					if expr != "" {
						buf.WriteString(fmt.Sprintf("    %s: %s = %s  # %s\n", normalizedFieldName, typeName, expr, inlineFieldComment))
					} else {
						buf.WriteString(fmt.Sprintf("    %s: %s  # %s\n", normalizedFieldName, typeName, inlineFieldComment))
					}
				} else {
					if expr != "" {
						buf.WriteString(fmt.Sprintf("    %s: %s = %s\n", normalizedFieldName, typeName, expr))
					} else {
						buf.WriteString(fmt.Sprintf("    %s: %s\n", normalizedFieldName, typeName))
					}
				}
				hasRenderedField = true
				continue
			}
			defaultValue := strings.TrimSpace(extraField.Default)
			if defaultValue == "" {
				defaultValue = "None"
			}
			if defaultValue == "None" && !strings.HasPrefix(typeName, "Optional[") {
				typeName = "Optional[" + typeName + "]"
			}
			expr := defaultValue
			if alias != "" {
				expr = fmt.Sprintf("Field(default=%s, alias=%q)", defaultValue, alias)
			}
			if inlineFieldComment != "" {
				buf.WriteString(fmt.Sprintf("    %s: %s = %s  # %s\n", normalizedFieldName, typeName, expr, inlineFieldComment))
			} else {
				buf.WriteString(fmt.Sprintf("    %s: %s = %s\n", normalizedFieldName, typeName, expr))
			}
			hasRenderedField = true
		}
		combinedExtraCode := make([]string, 0, len(model.ExtraCode)+3)
		combinedExtraCode = append(combinedExtraCode, autoModelExtraCodeForPagination(model, properties, extraFieldByName)...)
		combinedExtraCode = append(combinedExtraCode, autoModelExtraCodeForBeforeValidators(model)...)
		combinedExtraCode = append(combinedExtraCode, autoModelExtraCodeForBuilders(model)...)
		combinedExtraCode = append(combinedExtraCode, model.ExtraCode...)
		if hasRenderedField && len(combinedExtraCode) > 0 {
			buf.WriteString("\n")
		}
		for _, block := range combinedExtraCode {
			AppendIndentedCode(&buf, block, 1)
		}
		buf.WriteString("\n\n")
	}

	if meta.Package != nil && len(meta.Package.EmptyModels) > 0 {
		for i, modelName := range meta.Package.EmptyModels {
			name := strings.TrimSpace(modelName)
			if name == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", name))
			classKey := modulePrefix + "." + name
			if docstring := strings.TrimSpace(commentOverrides.ClassDocstrings[classKey]); docstring != "" {
				style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[classKey])
				WriteClassDocstring(&buf, 1, docstring, style)
			} else {
				buf.WriteString("    pass\n")
			}
			buf.WriteString("\n")
			if i < len(meta.Package.EmptyModels)-1 {
				buf.WriteString("\n")
			}
		}
	}

	return strings.TrimRight(buf.String(), "\n")
}

func modelFieldType(model packageModelDefinition, propertyName string, fallback string) string {
	if len(model.FieldTypes) == 0 {
		return fallback
	}
	if fieldType, ok := model.FieldTypes[propertyName]; ok && strings.TrimSpace(fieldType) != "" {
		return strings.TrimSpace(fieldType)
	}
	return fallback
}

func autoModelExtraCodeForBeforeValidators(model packageModelDefinition) []string {
	if len(model.BeforeValidators) == 0 {
		return nil
	}
	blocks := make([]string, 0, len(model.BeforeValidators))
	for _, validator := range model.BeforeValidators {
		fieldName := NormalizePythonIdentifier(strings.TrimSpace(validator.Field))
		if fieldName == "" {
			continue
		}
		methodName := NormalizePythonIdentifier(strings.TrimSpace(validator.Method))
		if methodName == "" {
			methodName = "normalize_" + fieldName
		}

		switch strings.TrimSpace(validator.Rule) {
		case "int_to_string":
			blocks = append(blocks, fmt.Sprintf(
				"@field_validator(%q, mode=\"before\")\n@classmethod\ndef %s(cls, v):\n    if isinstance(v, int):\n        return str(v)\n    return v",
				fieldName,
				methodName,
			))
		case "empty_string_to_zero":
			blocks = append(blocks, fmt.Sprintf(
				"@field_validator(%q, mode=\"before\")\n@classmethod\ndef %s(cls, v):\n    if v == \"\":\n        return 0\n    return v",
				fieldName,
				methodName,
			))
		}
	}
	return blocks
}

func autoModelExtraCodeForBuilders(model packageModelDefinition) []string {
	if len(model.Builders) == 0 {
		return nil
	}
	blocks := make([]string, 0, len(model.Builders))
	for _, builder := range model.Builders {
		methodName := NormalizePythonIdentifier(strings.TrimSpace(builder.Name))
		if methodName == "" {
			continue
		}

		params := make([]string, 0, len(builder.Params))
		for _, param := range builder.Params {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}
			params = append(params, param)
		}
		paramExpr := strings.Join(params, ", ")
		signature := methodName + "()"
		if paramExpr != "" {
			signature = methodName + "(" + paramExpr + ")"
		}

		returnType := strings.TrimSpace(builder.ReturnType)
		if returnType == "" {
			returnType = model.Name
		}

		assignments := make([]string, 0, len(builder.Args))
		for _, arg := range builder.Args {
			argName := NormalizePythonIdentifier(strings.TrimSpace(arg.Name))
			if argName == "" {
				continue
			}
			argExpr := strings.TrimSpace(arg.Expr)
			if argExpr == "" {
				argExpr = argName
			}
			assignments = append(assignments, fmt.Sprintf("%s=%s", argName, argExpr))
		}
		returnCall := model.Name + "()"
		if len(assignments) > 0 {
			returnCall = model.Name + "(" + strings.Join(assignments, ", ") + ")"
		}

		blocks = append(blocks, fmt.Sprintf(
			"@staticmethod\ndef %s -> %q:\n    return %s",
			signature,
			returnType,
			returnCall,
		))
	}
	return blocks
}

func autoModelExtraCodeForPagination(
	model packageModelDefinition,
	properties map[string]*openapi.Schema,
	extraFieldByName map[string]config.ModelField,
) []string {
	if len(model.ExtraCode) > 0 {
		return nil
	}

	itemType, ok := numberPagedResponseItemType(model.BaseClasses)
	if ok {
		itemsField, ok := pagedItemsField(model, properties, extraFieldByName, itemType)
		if !ok {
			return nil
		}
		totalExpr := "None"
		if totalField, ok := pagedTotalField(properties, extraFieldByName); ok {
			totalExpr = fmt.Sprintf("self.%s", totalField)
		}
		hasMoreExpr := "None"
		if hasMoreField, ok := pagedHasMoreField(properties, extraFieldByName); ok {
			hasMoreExpr = fmt.Sprintf("self.%s", hasMoreField)
		}
		return []string{
			fmt.Sprintf(
				"def get_total(self) -> Optional[int]:\n    return %s\n\ndef get_has_more(self) -> Optional[bool]:\n    return %s\n\ndef get_items(self) -> List[%s]:\n    return self.%s",
				totalExpr,
				hasMoreExpr,
				itemType,
				itemsField,
			),
		}
	}

	itemType, ok = tokenPagedResponseItemType(model.BaseClasses)
	if ok {
		itemsField, ok := pagedItemsField(model, properties, extraFieldByName, itemType)
		if !ok {
			return nil
		}
		nextTokenField, ok := pagedNextPageTokenField(properties, extraFieldByName)
		if !ok {
			return nil
		}
		hasMoreField, ok := pagedHasMoreField(properties, extraFieldByName)
		if !ok {
			return nil
		}
		return []string{
			fmt.Sprintf(
				"def get_next_page_token(self) -> Optional[str]:\n    return self.%s\n\ndef get_has_more(self) -> Optional[bool]:\n    return self.%s\n\ndef get_items(self) -> List[%s]:\n    return self.%s",
				nextTokenField,
				hasMoreField,
				itemType,
				itemsField,
			),
		}
	}

	itemType, ok = lastIDPagedResponseItemType(model.BaseClasses)
	if ok {
		itemsField, ok := pagedItemsField(model, properties, extraFieldByName, itemType)
		if !ok {
			return nil
		}
		firstIDField, ok := pagedFirstIDField(properties, extraFieldByName)
		if !ok {
			return nil
		}
		lastIDField, ok := pagedLastIDField(properties, extraFieldByName)
		if !ok {
			return nil
		}
		hasMoreField, ok := pagedHasMoreField(properties, extraFieldByName)
		if !ok {
			return nil
		}
		return []string{
			fmt.Sprintf(
				"def get_first_id(self) -> str:\n    return self.%s\n\ndef get_last_id(self) -> str:\n    return self.%s\n\ndef get_has_more(self) -> bool:\n    return self.%s\n\ndef get_items(self) -> List[%s]:\n    return self.%s",
				firstIDField,
				lastIDField,
				hasMoreField,
				itemType,
				itemsField,
			),
		}
	}

	return nil
}

func numberPagedResponseItemType(baseClasses []string) (string, bool) {
	return pagedResponseItemType(baseClasses, "NumberPagedResponse[")
}

func tokenPagedResponseItemType(baseClasses []string) (string, bool) {
	return pagedResponseItemType(baseClasses, "TokenPagedResponse[")
}

func lastIDPagedResponseItemType(baseClasses []string) (string, bool) {
	return pagedResponseItemType(baseClasses, "LastIDPagedResponse[")
}

func pagedResponseItemType(baseClasses []string, prefix string) (string, bool) {
	for _, rawBaseClass := range baseClasses {
		baseClass := strings.TrimSpace(rawBaseClass)
		if !strings.HasPrefix(baseClass, prefix) || !strings.HasSuffix(baseClass, "]") {
			continue
		}
		itemType := strings.TrimSuffix(strings.TrimPrefix(baseClass, prefix), "]")
		itemType = strings.TrimSpace(itemType)
		if itemType == "" {
			continue
		}
		return itemType, true
	}
	return "", false
}

func pagedItemsField(
	model packageModelDefinition,
	properties map[string]*openapi.Schema,
	extraFieldByName map[string]config.ModelField,
	itemType string,
) (string, bool) {
	if modelHasField("items", properties, extraFieldByName) {
		return "items", true
	}
	for _, field := range model.ExtraFields {
		if strings.TrimSpace(field.Name) == "" {
			continue
		}
		if isListFieldType(field.Type, itemType) {
			return strings.TrimSpace(field.Name), true
		}
	}
	for _, fieldName := range model.FieldOrder {
		typeName, ok := model.FieldTypes[fieldName]
		if !ok {
			continue
		}
		if isListFieldType(typeName, itemType) {
			return strings.TrimSpace(fieldName), true
		}
	}
	return "", false
}

func pagedTotalField(properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) (string, bool) {
	for _, candidate := range []string{"total", "total_count"} {
		if modelHasField(candidate, properties, extraFieldByName) {
			return candidate, true
		}
	}
	return "", false
}

func pagedHasMoreField(properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) (string, bool) {
	if modelHasField("has_more", properties, extraFieldByName) {
		return "has_more", true
	}
	return "", false
}

func pagedNextPageTokenField(properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) (string, bool) {
	if modelHasField("next_page_token", properties, extraFieldByName) {
		return "next_page_token", true
	}
	return "", false
}

func pagedFirstIDField(properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) (string, bool) {
	if modelHasField("first_id", properties, extraFieldByName) {
		return "first_id", true
	}
	return "", false
}

func pagedLastIDField(properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) (string, bool) {
	if modelHasField("last_id", properties, extraFieldByName) {
		return "last_id", true
	}
	return "", false
}

func isListFieldType(typeName, itemType string) bool {
	normalizedType := strings.ReplaceAll(strings.TrimSpace(typeName), " ", "")
	normalizedItemType := strings.ReplaceAll(strings.TrimSpace(itemType), " ", "")
	if normalizedType == "" || normalizedItemType == "" {
		return false
	}
	if normalizedType == "List["+normalizedItemType+"]" {
		return true
	}
	if normalizedType == "Optional[List["+normalizedItemType+"]]" {
		return true
	}
	return false
}

func modelHasField(name string, properties map[string]*openapi.Schema, extraFieldByName map[string]config.ModelField) bool {
	if _, ok := properties[name]; ok {
		return true
	}
	_, ok := extraFieldByName[name]
	return ok
}

func hasModelFieldTypeOverride(model packageModelDefinition, propertyName string) bool {
	if len(model.FieldTypes) == 0 {
		return false
	}
	fieldType, ok := model.FieldTypes[propertyName]
	return ok && strings.TrimSpace(fieldType) != ""
}

func modelFieldDefault(model packageModelDefinition, propertyName string) string {
	if len(model.FieldDefaults) == 0 {
		return "None"
	}
	if value, ok := model.FieldDefaults[propertyName]; ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return "None"
}

func unwrapOptionalType(typeName string) string {
	trimmed := strings.TrimSpace(typeName)
	if !strings.HasPrefix(trimmed, "Optional[") || !strings.HasSuffix(trimmed, "]") {
		return typeName
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "Optional["), "]")
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return typeName
	}
	return inner
}

func modelHasCustomClassDocstring(model packageModelDefinition) bool {
	return CodeBlocksHaveLeadingDocstring(model.PrependCode) || CodeBlocksHaveLeadingDocstring(model.ExtraCode)
}

func schemaDescription(schema *openapi.Schema) string {
	if schema == nil {
		return ""
	}
	return normalizeSwaggerDescription(schema.Description)
}

func schemaCommentLines(doc *openapi.Document, schema *openapi.Schema) []string {
	if schema == nil {
		return nil
	}
	description := strings.TrimSpace(schema.Description)
	if description == "" && doc != nil {
		if resolved := doc.ResolveSchema(schema); resolved != nil && resolved != schema {
			description = strings.TrimSpace(resolved.Description)
		}
	}
	return descriptionLines(description)
}

func descriptionLines(description string) []string {
	normalized := normalizeSwaggerDescription(description)
	if normalized == "" {
		return nil
	}
	text := strings.ReplaceAll(normalized, "\r\n", "\n")
	rawLines := strings.Split(text, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func normalizeSwaggerDescription(description string) string {
	text := strings.TrimSpace(description)
	if text == "" {
		return ""
	}
	if swaggerDescriptionLooksLikeRichText(text) {
		if extracted := extractSwaggerRichText(text); extracted != "" {
			return extracted
		}
		return ""
	}
	return text
}

func swaggerDescriptionLooksLikeRichText(raw string) bool {
	text := strings.TrimSpace(raw)
	return strings.HasPrefix(text, "{") && strings.Contains(text, "\"insert\"") && strings.Contains(text, "\"ops\"")
}

func extractSwaggerRichText(raw string) string {
	var node interface{}
	if err := json.Unmarshal([]byte(raw), &node); err != nil {
		return ""
	}
	fragments := make([]string, 0, 32)
	collectRichTextInserts(node, &fragments)
	var buf strings.Builder
	var lastRune rune
	hasLastRune := false
	hasContent := false
	appendNewline := func() {
		if !hasContent {
			return
		}
		if hasLastRune && lastRune == '\n' {
			return
		}
		buf.WriteByte('\n')
		lastRune = '\n'
		hasLastRune = true
	}
	appendText := func(text string) {
		if text == "" {
			return
		}
		firstRune, _ := utf8.DecodeRuneInString(text)
		if hasContent && hasLastRune && shouldInsertRichTextSpace(lastRune, firstRune) {
			buf.WriteByte(' ')
			lastRune = ' '
			hasLastRune = true
		}
		buf.WriteString(text)
		r, _ := utf8.DecodeLastRuneInString(text)
		lastRune = r
		hasLastRune = true
		hasContent = true
	}
	for _, fragment := range fragments {
		text := strings.ReplaceAll(fragment, "\r\n", "\n")
		text = strings.ReplaceAll(text, "\r", "\n")
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			if strings.Contains(text, "\n") {
				appendNewline()
			}
			continue
		}
		if trimmed == "*" {
			// Rich text exports use "*" as list/heading markers. Treat it as a line break.
			appendNewline()
			continue
		}
		appendText(text)
	}
	if !hasContent {
		return ""
	}

	rawText := strings.TrimSpace(buf.String())
	if rawText == "" {
		return ""
	}

	lines := strings.Split(rawText, "\n")
	normalized := make([]string, 0, len(lines))
	prevBlank := true
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			if !prevBlank && len(normalized) > 0 {
				normalized = append(normalized, "")
				prevBlank = true
			}
			continue
		}
		if strings.HasPrefix(line, "*") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
			if line == "" {
				if !prevBlank && len(normalized) > 0 {
					normalized = append(normalized, "")
					prevBlank = true
				}
				continue
			}
		}
		normalized = append(normalized, line)
		prevBlank = false
	}

	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}
	return strings.Join(normalized, "\n")
}

func shouldInsertRichTextSpace(prev rune, next rune) bool {
	if unicode.IsSpace(prev) || unicode.IsSpace(next) {
		return false
	}
	if prev == '\n' || next == '\n' {
		return false
	}
	// CJK text generally does not need synthetic spaces between fragments.
	if unicode.In(prev, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) {
		return false
	}
	if unicode.In(next, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul) {
		return false
	}
	if (unicode.IsLetter(prev) || unicode.IsDigit(prev)) &&
		(unicode.IsLetter(next) || unicode.IsDigit(next)) {
		return true
	}
	return false
}

func collectRichTextInserts(node interface{}, fragments *[]string) {
	switch typed := node.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := typed[key]
			if key == "insert" {
				if text, ok := value.(string); ok {
					*fragments = append(*fragments, text)
				}
			}
			collectRichTextInserts(value, fragments)
		}
	case []interface{}:
		for _, item := range typed {
			collectRichTextInserts(item, fragments)
		}
	}
}

func packageClientClassName(meta PackageMeta, async bool) string {
	if meta.Package != nil {
		if async && strings.TrimSpace(meta.Package.AsyncClientClass) != "" {
			return strings.TrimSpace(meta.Package.AsyncClientClass)
		}
		if !async && strings.TrimSpace(meta.Package.ClientClass) != "" {
			return strings.TrimSpace(meta.Package.ClientClass)
		}
	}
	base := PackageClassName(meta.Name)
	if async {
		return "Async" + base + "Client"
	}
	return base + "Client"
}

func childImportModule(meta PackageMeta, module string) string {
	module = strings.TrimSpace(module)
	if module == "" {
		return ""
	}
	if strings.HasPrefix(module, "cozepy.") {
		return module
	}
	if strings.HasPrefix(module, ".") {
		suffix := strings.TrimPrefix(module, ".")
		if suffix == "" {
			return "cozepy." + meta.ModulePath
		}
		return "cozepy." + meta.ModulePath + "." + suffix
	}
	return "cozepy." + strings.TrimPrefix(module, ".")
}

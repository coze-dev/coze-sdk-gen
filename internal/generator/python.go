package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type operationBinding struct {
	PackageName string
	MethodName  string
	Details     openapi.OperationDetails
	Mapping     *config.OperationMapping
	Order       int
}

type packageMeta struct {
	Name       string
	ModulePath string
	DirPath    string
	Package    *config.Package
}

type fileWriter struct {
	count   int
	written map[string]struct{}
}

var pythonSupportOverridePaths = map[string]struct{}{
	"cozepy/files/__init__.py":      {},
	"cozepy/knowledge/__init__.py":  {},
	"cozepy/variables/__init__.py":  {},
	"cozepy/workflows/__init__.py":  {},
	"cozepy/workspaces/__init__.py": {},
}

func GeneratePython(cfg *config.Config, doc *openapi.Document) (Result, error) {
	if cfg == nil {
		return Result{}, fmt.Errorf("config is required")
	}
	if doc == nil {
		return Result{}, fmt.Errorf("swagger document is required")
	}

	report := cfg.ValidateAgainstSwagger(doc)
	if report.HasErrors() {
		return Result{}, fmt.Errorf("config and swagger mismatch: %s", report.Error())
	}

	bindings := buildOperationBindings(cfg, doc)
	if len(bindings) == 0 {
		return Result{}, fmt.Errorf("no operations selected for generation")
	}

	packages := groupBindingsByPackage(bindings)
	packageMetas := buildPackageMeta(cfg, packages)

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{
		written: map[string]struct{}{},
	}
	if err := writePythonSDK(cfg.OutputSDK, doc, packages, packageMetas, writer); err != nil {
		return Result{}, err
	}

	return Result{
		GeneratedFiles: writer.count,
		GeneratedOps:   len(bindings),
	}, nil
}

func buildOperationBindings(cfg *config.Config, doc *openapi.Document) []operationBinding {
	allOps := doc.ListOperationDetails()
	bindings := make([]operationBinding, 0)

	for _, details := range allOps {
		if cfg.IsIgnored(details.Path, details.Method) {
			continue
		}

		mappings := cfg.FindOperationMappings(details.Path, details.Method)
		if cfg.API.GenerateOnlyMapped && len(mappings) == 0 {
			continue
		}

		if len(mappings) > 0 {
			for _, mapping := range mappings {
				mappingCopy := mapping
				for methodIndex, sdkMethod := range mapping.SDKMethods {
					pkgName, methodName, ok := config.ParseSDKMethod(sdkMethod)
					if !ok {
						continue
					}
					pkg, ok := cfg.ResolvePackage(details.Path, pkgName)
					if !ok {
						continue
					}
					order := len(bindings)
					if mappingCopy.Order > 0 {
						order = mappingCopy.Order + methodIndex
					}
					bindings = append(bindings, operationBinding{
						PackageName: normalizePackageName(pkg.Name),
						MethodName:  normalizeMethodName(methodName),
						Details:     details,
						Mapping:     &mappingCopy,
						Order:       order,
					})
				}
			}
			continue
		}

		pkg, ok := cfg.ResolvePackage(details.Path, "")
		if !ok {
			continue
		}
		bindings = append(bindings, operationBinding{
			PackageName: normalizePackageName(pkg.Name),
			MethodName:  defaultMethodName(details.OperationID, details.Path, details.Method),
			Details:     details,
			Order:       len(bindings),
		})
	}

	return deduplicateBindings(bindings)
}

func deduplicateBindings(bindings []operationBinding) []operationBinding {
	seen := map[string]int{}
	for i := range bindings {
		key := bindings[i].PackageName + ":" + bindings[i].MethodName
		seen[key]++
		if seen[key] > 1 {
			bindings[i].MethodName = fmt.Sprintf("%s_%d", bindings[i].MethodName, seen[key])
		}
	}
	return bindings
}

func groupBindingsByPackage(bindings []operationBinding) map[string][]operationBinding {
	pkgOps := map[string][]operationBinding{}
	for _, binding := range bindings {
		pkgOps[binding.PackageName] = append(pkgOps[binding.PackageName], binding)
	}
	for pkgName := range pkgOps {
		sort.Slice(pkgOps[pkgName], func(i, j int) bool {
			return pkgOps[pkgName][i].Order < pkgOps[pkgName][j].Order
		})
	}
	return pkgOps
}

func buildPackageMeta(cfg *config.Config, packages map[string][]operationBinding) map[string]packageMeta {
	metas := map[string]packageMeta{}
	for _, pkg := range cfg.API.Packages {
		pkgCopy := pkg
		name := normalizePackageName(pkg.Name)
		dir := normalizePackageDir(pkg.SourceDir, name)
		metas[name] = packageMeta{
			Name:       name,
			ModulePath: strings.ReplaceAll(dir, "/", "."),
			DirPath:    dir,
			Package:    &pkgCopy,
		}
	}
	for name := range packages {
		if _, ok := metas[name]; ok {
			continue
		}
		metas[name] = packageMeta{
			Name:       name,
			ModulePath: name,
			DirPath:    name,
		}
	}
	return metas
}

func writePythonSDK(
	outputDir string,
	doc *openapi.Document,
	packages map[string][]operationBinding,
	packageMetas map[string]packageMeta,
	writer *fileWriter,
) error {
	rootDir := filepath.Join(outputDir, "cozepy")
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return fmt.Errorf("create python package root %q: %w", rootDir, err)
	}

	pkgNames := make([]string, 0, len(packages))
	for pkgName := range packages {
		pkgNames = append(pkgNames, pkgName)
	}
	for pkgName, meta := range packageMetas {
		if _, ok := packages[pkgName]; ok {
			continue
		}
		if meta.Package == nil {
			continue
		}
		if len(meta.Package.ChildClients) > 0 {
			pkgNames = append(pkgNames, pkgName)
		}
	}
	sort.Strings(pkgNames)

	configPy, err := renderConfigPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "config.py"), configPy); err != nil {
		return err
	}
	utilPy, err := renderUtilPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "util.py"), utilPy); err != nil {
		return err
	}
	modelPy, err := renderModelPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "model.py"), modelPy); err != nil {
		return err
	}
	requestPy, err := renderRequestPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "request.py"), requestPy); err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "py.typed"), ""); err != nil {
		return err
	}

	for _, pkgName := range pkgNames {
		meta := packageMetas[pkgName]
		pkgDir := filepath.Join(rootDir, meta.DirPath)
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			return fmt.Errorf("create package directory %q: %w", pkgDir, err)
		}
		content := renderPackageModule(doc, meta, packages[pkgName])
		if err := writer.write(filepath.Join(pkgDir, "__init__.py"), content); err != nil {
			return err
		}
	}

	if err := writePythonSupportAssets(outputDir, writer); err != nil {
		return err
	}

	return nil
}

func (w *fileWriter) write(path string, content string) error {
	return w.writeBytes(path, []byte(content))
}

func (w *fileWriter) writeBytes(path string, content []byte) error {
	cleanPath := filepath.Clean(path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	w.count++
	w.written[cleanPath] = struct{}{}
	return nil
}

func writePythonSupportAssets(outputDir string, writer *fileWriter) error {
	supportFiles, err := listPythonSupportFiles()
	if err != nil {
		return err
	}
	for _, relPath := range supportFiles {
		targetPath := filepath.Join(outputDir, filepath.FromSlash(relPath))
		if _, exists := writer.written[filepath.Clean(targetPath)]; exists {
			if _, override := pythonSupportOverridePaths[filepath.ToSlash(relPath)]; !override {
				continue
			}
		}
		content, err := readPythonSupportFile(relPath)
		if err != nil {
			return err
		}
		if err := writer.writeBytes(targetPath, content); err != nil {
			return err
		}
	}
	return nil
}

func renderConfigPy() (string, error) {
	return renderPythonTemplate("config.py.tpl", map[string]any{})
}

func renderUtilPy() (string, error) {
	return renderPythonTemplate("util.py.tpl", map[string]any{})
}

func renderModelPy() (string, error) {
	return renderPythonTemplate("model.py.tpl", map[string]any{})
}

func renderRequestPy() (string, error) {
	return renderPythonTemplate("request.py.tpl", map[string]any{})
}

func renderPackageModule(doc *openapi.Document, meta packageMeta, bindings []operationBinding) string {
	var buf bytes.Buffer
	hasChildClients := meta.Package != nil && len(meta.Package.ChildClients) > 0
	hasTypedChildClients := false
	if hasChildClients {
		for _, child := range meta.Package.ChildClients {
			if !child.DisableTypeHints {
				hasTypedChildClients = true
				break
			}
		}
	}
	schemaAliases := packageSchemaAliases(meta)
	modelDefs := resolvePackageModelDefinitions(doc, meta)
	hasModelClasses := len(modelDefs) > 0 || (meta.Package != nil && len(meta.Package.EmptyModels) > 0)
	hasTokenPagination := packageHasTokenPagination(bindings)
	hasNumberPagination := packageHasNumberPagination(bindings)
	needAny, needDict := packageNeedsAnyDict(bindings)
	hasStandardEnumClasses := false
	hasDynamicEnumClasses := false
	for _, model := range modelDefs {
		if model.IsEnum {
			if model.EnumBase == "dynamic_str" {
				hasDynamicEnumClasses = true
			} else {
				hasStandardEnumClasses = true
			}
		}
	}

	if hasStandardEnumClasses {
		buf.WriteString("from enum import Enum\n")
	}
	typingImports := make([]string, 0)
	if hasTypedChildClients {
		typingImports = append(typingImports, "TYPE_CHECKING")
	}
	if needAny {
		typingImports = append(typingImports, "Any")
	}
	if needDict {
		typingImports = append(typingImports, "Dict")
	}
	if len(modelDefs) > 0 || hasTokenPagination || hasNumberPagination {
		typingImports = append(typingImports, "List")
	}
	if hasTypedChildClients || len(modelDefs) > 0 || hasTokenPagination || hasNumberPagination || len(bindings) > 0 {
		typingImports = append(typingImports, "Optional")
	}
	if len(typingImports) > 0 {
		buf.WriteString(fmt.Sprintf("from typing import %s\n\n", strings.Join(typingImports, ", ")))
	}
	modelImports := make([]string, 0)
	if hasTokenPagination {
		modelImports = append(modelImports, "AsyncTokenPaged")
	}
	if hasNumberPagination {
		modelImports = append(modelImports, "AsyncNumberPaged")
	}
	if hasDynamicEnumClasses {
		modelImports = append(modelImports, "DynamicStrEnum")
	}
	if hasModelClasses {
		modelImports = append(modelImports, "CozeModel")
	}
	if hasTokenPagination {
		modelImports = append(modelImports, "TokenPaged", "TokenPagedResponse")
	}
	if hasNumberPagination {
		modelImports = append(modelImports, "NumberPaged", "NumberPagedResponse")
	}
	if len(modelImports) > 0 {
		buf.WriteString(fmt.Sprintf("from cozepy.model import %s\n", strings.Join(modelImports, ", ")))
	}
	requestImports := []string{"Requester"}
	if hasTokenPagination || hasNumberPagination {
		requestImports = append([]string{"HTTPRequest"}, requestImports...)
	}
	buf.WriteString(fmt.Sprintf("from cozepy.request import %s\n", strings.Join(requestImports, ", ")))
	utilImports := []string{"remove_url_trailing_slash"}
	if len(bindings) > 0 {
		utilImports = append([]string{"dump_exclude_none"}, utilImports...)
	}
	buf.WriteString(fmt.Sprintf("from cozepy.util import %s\n", strings.Join(utilImports, ", ")))

	imports := collectTypeImports(doc, bindings)
	if len(imports) > 0 {
		buf.WriteString(fmt.Sprintf("from cozepy.types import %s\n", strings.Join(imports, ", ")))
	}

	if hasTypedChildClients {
		buf.WriteString("\nif TYPE_CHECKING:\n")
		for _, child := range meta.Package.ChildClients {
			if child.DisableTypeHints {
				continue
			}
			typeModule := strings.TrimSpace(child.Module)
			if !strings.HasPrefix(typeModule, ".") {
				typeModule = childTypeImportModule(meta, typeModule)
			}
			if typeModule == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("    from %s import %s, %s\n", typeModule, child.AsyncClass, child.SyncClass))
		}
	}
	buf.WriteString("\n\n")

	if hasModelClasses {
		buf.WriteString(renderPackageModelDefinitions(doc, meta, modelDefs, schemaAliases))
		buf.WriteString("\n")
	}
	if hasTokenPagination || hasNumberPagination {
		buf.WriteString(renderPagedResponseClasses(bindings))
		buf.WriteString("\n")
	}

	syncClass := packageClientClassName(meta, false)
	asyncClass := packageClientClassName(meta, true)

	buf.WriteString(fmt.Sprintf("class %s(object):\n", syncClass))
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n\n")
	if hasChildClients {
		for _, child := range meta.Package.ChildClients {
			attribute := normalizePythonIdentifier(child.Attribute)
			if child.DisableTypeHints {
				buf.WriteString(fmt.Sprintf("        self._%s = None\n", attribute))
			} else {
				buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.SyncClass))
			}
		}
		buf.WriteString("\n")
	}
	if hasChildClients {
		for _, child := range meta.Package.ChildClients {
			buf.WriteString(renderChildClientProperty(meta, child, false))
			buf.WriteString("\n")
		}
	}

	for _, binding := range bindings {
		buf.WriteString(renderOperationMethod(doc, binding, false))
		buf.WriteString("\n")
	}

	buf.WriteString(fmt.Sprintf("class %s(object):\n", asyncClass))
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n\n")
	if hasChildClients {
		for _, child := range meta.Package.ChildClients {
			attribute := normalizePythonIdentifier(child.Attribute)
			if child.DisableTypeHints {
				buf.WriteString(fmt.Sprintf("        self._%s = None\n", attribute))
			} else {
				buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.AsyncClass))
			}
		}
		buf.WriteString("\n")
	}
	if hasChildClients {
		for _, child := range meta.Package.ChildClients {
			buf.WriteString(renderChildClientProperty(meta, child, true))
			buf.WriteString("\n")
		}
	}

	for _, binding := range bindings {
		buf.WriteString(renderOperationMethod(doc, binding, true))
		buf.WriteString("\n")
	}

	return buf.String()
}

func collectTypeImports(doc *openapi.Document, bindings []operationBinding) []string {
	_ = doc
	_ = bindings
	return nil
}

func packageHasTokenPagination(bindings []operationBinding) bool {
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		if strings.TrimSpace(binding.Mapping.Pagination) == "token" {
			return true
		}
	}
	return false
}

func packageHasNumberPagination(bindings []operationBinding) bool {
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		if strings.TrimSpace(binding.Mapping.Pagination) == "number" {
			return true
		}
	}
	return false
}

func packageNeedsAnyDict(bindings []operationBinding) (bool, bool) {
	needAny := false
	needDict := false

	for _, binding := range bindings {
		mapping := binding.Mapping
		if mapping == nil || strings.TrimSpace(mapping.ResponseType) == "" {
			paginationMode := ""
			if mapping != nil {
				paginationMode = strings.TrimSpace(mapping.Pagination)
			}
			if paginationMode != "token" && paginationMode != "number" {
				needAny = true
				needDict = true
			}
		} else {
			responseType := strings.TrimSpace(mapping.ResponseType)
			if strings.Contains(responseType, "Any") {
				needAny = true
			}
			if strings.Contains(responseType, "Dict") {
				needDict = true
			}
		}

		if mapping == nil || !mapping.UseKwargsHeaders {
			needDict = true
		}

		if binding.Details.RequestBodySchema != nil {
			disableRequestBody := mapping != nil && mapping.DisableRequestBody
			hasBodyFields := mapping != nil && len(mapping.BodyFields) > 0
			if !disableRequestBody && !hasBodyFields {
				needAny = true
				needDict = true
			}
		}

		if mapping != nil {
			for _, field := range mapping.QueryFields {
				fieldType := strings.TrimSpace(field.Type)
				if strings.Contains(fieldType, "Any") {
					needAny = true
				}
				if strings.Contains(fieldType, "Dict") {
					needDict = true
				}
			}
		}
	}

	return needAny, needDict
}

func renderPagedResponseClasses(bindings []operationBinding) string {
	seen := map[string]struct{}{}
	ordered := make([]operationBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		paginationMode := strings.TrimSpace(binding.Mapping.Pagination)
		if paginationMode != "token" && paginationMode != "number" {
			continue
		}
		className := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
		if className == "" || itemType == "" {
			continue
		}
		if _, ok := seen[className]; ok {
			continue
		}
		seen[className] = struct{}{}
		ordered = append(ordered, binding)
	}

	var buf bytes.Buffer
	for _, binding := range ordered {
		paginationMode := strings.TrimSpace(binding.Mapping.Pagination)
		className := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
		itemsField := strings.TrimSpace(binding.Mapping.PaginationItemsField)
		if itemsField == "" {
			itemsField = "items"
		}
		if paginationMode == "token" {
			hasMoreField := strings.TrimSpace(binding.Mapping.PaginationHasMoreField)
			if hasMoreField == "" {
				hasMoreField = "has_more"
			}
			nextTokenField := strings.TrimSpace(binding.Mapping.PaginationNextTokenField)
			if nextTokenField == "" {
				nextTokenField = "next_page_token"
			}
			buf.WriteString(fmt.Sprintf("class %s(CozeModel, TokenPagedResponse[%s]):\n", className, itemType))
			buf.WriteString(fmt.Sprintf("    %s: List[%s]\n", itemsField, itemType))
			buf.WriteString(fmt.Sprintf("    %s: Optional[str] = None\n", nextTokenField))
			buf.WriteString(fmt.Sprintf("    %s: bool\n\n", hasMoreField))
			buf.WriteString("    def get_next_page_token(self) -> Optional[str]:\n")
			buf.WriteString(fmt.Sprintf("        return self.%s\n\n", nextTokenField))
			buf.WriteString("    def get_has_more(self) -> Optional[bool]:\n")
			buf.WriteString(fmt.Sprintf("        return self.%s\n\n", hasMoreField))
			buf.WriteString(fmt.Sprintf("    def get_items(self) -> List[%s]:\n", itemType))
			buf.WriteString(fmt.Sprintf("        return self.%s\n\n", itemsField))
			continue
		}

		totalField := strings.TrimSpace(binding.Mapping.PaginationTotalField)
		if totalField == "" {
			totalField = "total"
		}
		buf.WriteString(fmt.Sprintf("class %s(CozeModel, NumberPagedResponse[%s]):\n", className, itemType))
		buf.WriteString(fmt.Sprintf("    %s: int\n", totalField))
		buf.WriteString(fmt.Sprintf("    %s: List[%s]\n\n", itemsField, itemType))
		buf.WriteString("    def get_total(self) -> Optional[int]:\n")
		buf.WriteString(fmt.Sprintf("        return self.%s\n\n", totalField))
		buf.WriteString("    def get_has_more(self) -> Optional[bool]:\n")
		buf.WriteString("        return None\n\n")
		buf.WriteString(fmt.Sprintf("    def get_items(self) -> List[%s]:\n", itemType))
		buf.WriteString(fmt.Sprintf("        return self.%s\n\n", itemsField))
	}
	return strings.TrimRight(buf.String(), "\n")
}

type packageModelDefinition struct {
	SchemaName     string
	Name           string
	Schema         *openapi.Schema
	IsEnum         bool
	FieldOrder     []string
	RequiredFields []string
	EnumBase       string
}

func packageSchemaAliases(meta packageMeta) map[string]string {
	aliases := map[string]string{}
	if meta.Package == nil {
		return aliases
	}
	for _, model := range meta.Package.ModelSchemas {
		schemaName := strings.TrimSpace(model.Schema)
		modelName := strings.TrimSpace(model.Name)
		if schemaName == "" || modelName == "" {
			continue
		}
		aliases[schemaName] = modelName
	}
	return aliases
}

func resolvePackageModelDefinitions(doc *openapi.Document, meta packageMeta) []packageModelDefinition {
	if meta.Package == nil || doc == nil {
		return nil
	}
	result := make([]packageModelDefinition, 0, len(meta.Package.ModelSchemas))
	for _, model := range meta.Package.ModelSchemas {
		schemaName := strings.TrimSpace(model.Schema)
		modelName := strings.TrimSpace(model.Name)
		if schemaName == "" || modelName == "" {
			continue
		}
		schema, ok := doc.Components.Schemas[schemaName]
		if !ok || schema == nil {
			continue
		}
		resolved := doc.ResolveSchema(schema)
		if resolved == nil {
			continue
		}
		isEnum := len(resolved.Enum) > 0 && (resolved.Type == "string" || resolved.Type == "")
		result = append(result, packageModelDefinition{
			SchemaName:     schemaName,
			Name:           modelName,
			Schema:         resolved,
			IsEnum:         isEnum,
			FieldOrder:     append([]string(nil), model.FieldOrder...),
			RequiredFields: append([]string(nil), model.RequiredFields...),
			EnumBase:       strings.TrimSpace(model.EnumBase),
		})
	}
	return result
}

func renderPackageModelDefinitions(
	doc *openapi.Document,
	meta packageMeta,
	models []packageModelDefinition,
	schemaAliases map[string]string,
) string {
	var buf bytes.Buffer

	for _, model := range models {
		if model.IsEnum {
			if model.EnumBase == "dynamic_str" {
				buf.WriteString(fmt.Sprintf("class %s(DynamicStrEnum):\n", model.Name))
			} else {
				buf.WriteString(fmt.Sprintf("class %s(str, Enum):\n", model.Name))
			}
			if len(model.Schema.Enum) == 0 {
				buf.WriteString("    pass\n\n")
				continue
			}
			for _, enumValue := range model.Schema.Enum {
				value, ok := enumValue.(string)
				if !ok {
					continue
				}
				memberName := enumMemberName(value)
				buf.WriteString(fmt.Sprintf("    %s = %q\n", memberName, value))
			}
			buf.WriteString("\n")
			continue
		}

		buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", model.Name))
		properties := model.Schema.Properties
		if len(properties) == 0 {
			buf.WriteString("    pass\n\n")
			continue
		}

		requiredSet := map[string]bool{}
		for _, requiredName := range model.Schema.Required {
			requiredSet[requiredName] = true
		}
		for _, requiredName := range model.RequiredFields {
			requiredName = strings.TrimSpace(requiredName)
			if requiredName == "" {
				continue
			}
			requiredSet[requiredName] = true
		}
		propertyNames := make([]string, 0, len(properties))
		seenProperties := map[string]bool{}
		for _, propertyName := range model.FieldOrder {
			if _, ok := properties[propertyName]; !ok {
				continue
			}
			propertyNames = append(propertyNames, propertyName)
			seenProperties[propertyName] = true
		}
		remaining := make([]string, 0, len(properties))
		for propertyName := range properties {
			if seenProperties[propertyName] {
				continue
			}
			remaining = append(remaining, propertyName)
		}
		sort.Strings(remaining)
		propertyNames = append(propertyNames, remaining...)
		for _, propertyName := range propertyNames {
			propertySchema := properties[propertyName]
			typeName := pythonTypeForSchemaWithAliases(doc, propertySchema, requiredSet[propertyName], schemaAliases)
			fieldName := normalizePythonIdentifier(propertyName)
			if requiredSet[propertyName] {
				buf.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, typeName))
			} else {
				buf.WriteString(fmt.Sprintf("    %s: %s = None\n", fieldName, typeName))
			}
		}
		buf.WriteString("\n")
	}

	if meta.Package != nil && len(meta.Package.EmptyModels) > 0 {
		for _, modelName := range meta.Package.EmptyModels {
			name := strings.TrimSpace(modelName)
			if name == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", name))
			buf.WriteString("    pass\n\n")
		}
	}

	return strings.TrimRight(buf.String(), "\n")
}

func enumMemberName(value string) string {
	name := strings.TrimSpace(value)
	if name == "" {
		return "UNKNOWN"
	}
	name = strings.ToUpper(collapseUnderscore(toSnake(name)))
	name = strings.Trim(name, "_")
	if name == "" {
		name = "UNKNOWN"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "VALUE_" + name
	}
	return name
}

func packageClientClassName(meta packageMeta, async bool) string {
	if meta.Package != nil {
		if async && strings.TrimSpace(meta.Package.AsyncClientClass) != "" {
			return strings.TrimSpace(meta.Package.AsyncClientClass)
		}
		if !async && strings.TrimSpace(meta.Package.ClientClass) != "" {
			return strings.TrimSpace(meta.Package.ClientClass)
		}
	}
	base := packageClassName(meta.Name)
	if async {
		return "Async" + base + "Client"
	}
	return base + "Client"
}

func childTypeImportModule(meta packageMeta, module string) string {
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

func renderChildClientProperty(meta packageMeta, child config.ChildClient, async bool) string {
	attribute := normalizePythonIdentifier(child.Attribute)
	typeName := child.SyncClass
	if async {
		typeName = child.AsyncClass
	}
	module := strings.TrimSpace(child.Module)
	nilCheckIsNone := strings.TrimSpace(child.NilCheck) == "is_none"
	useKeywords := child.InitWithKeywords
	constructExpr := fmt.Sprintf("%s(self._base_url, self._requester)", typeName)
	if useKeywords {
		constructExpr = fmt.Sprintf("%s(base_url=self._base_url, requester=self._requester)", typeName)
	}

	var buf bytes.Buffer
	buf.WriteString("    @property\n")
	if child.DisableTypeHints {
		buf.WriteString(fmt.Sprintf("    def %s(self):\n", attribute))
	} else {
		buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", attribute, typeName))
	}
	if nilCheckIsNone {
		buf.WriteString(fmt.Sprintf("        if self._%s is None:\n", attribute))
	} else {
		buf.WriteString(fmt.Sprintf("        if not self._%s:\n", attribute))
	}

	if module == "" {
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else if strings.HasPrefix(module, ".") {
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", module, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else {
		absModule := childTypeImportModule(meta, module)
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", absModule, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	}
	buf.WriteString(fmt.Sprintf("        return self._%s\n", attribute))
	return buf.String()
}

type renderQueryField struct {
	RawName      string
	ArgName      string
	TypeName     string
	Required     bool
	DefaultValue string
}

func paginationOrderedFields(fields []renderQueryField, pageSizeField string, pageTokenOrNumField string) []renderQueryField {
	pageSizeField = strings.TrimSpace(pageSizeField)
	pageTokenOrNumField = strings.TrimSpace(pageTokenOrNumField)
	if pageSizeField == "" {
		pageSizeField = "page_size"
	}
	out := make([]renderQueryField, 0, len(fields))
	var pageSize *renderQueryField
	var pageTokenOrNum *renderQueryField
	for i := range fields {
		field := fields[i]
		switch field.RawName {
		case pageSizeField:
			pageSize = &field
		case pageTokenOrNumField:
			pageTokenOrNum = &field
		default:
			out = append(out, field)
		}
	}
	if pageSize != nil {
		out = append(out, *pageSize)
	}
	if pageTokenOrNum != nil {
		out = append(out, *pageTokenOrNum)
	}
	return out
}

func buildRenderQueryFields(
	doc *openapi.Document,
	details openapi.OperationDetails,
	mapping *config.OperationMapping,
	paramAliases map[string]string,
	argTypes map[string]string,
) []renderQueryField {
	fields := make([]renderQueryField, 0)
	if mapping != nil && len(mapping.QueryFields) > 0 {
		for _, field := range mapping.QueryFields {
			rawName := strings.TrimSpace(field.Name)
			if rawName == "" {
				continue
			}
			argName := operationArgName(rawName, paramAliases)
			typeName := strings.TrimSpace(field.Type)
			if typeName == "" {
				typeName = "Any"
			}
			if !field.Required && !strings.HasPrefix(typeName, "Optional[") {
				typeName = "Optional[" + typeName + "]"
			}
			fields = append(fields, renderQueryField{
				RawName:      rawName,
				ArgName:      argName,
				TypeName:     typeName,
				Required:     field.Required,
				DefaultValue: strings.TrimSpace(field.Default),
			})
		}
		return fields
	}

	for _, param := range details.QueryParameters {
		argName := operationArgName(param.Name, paramAliases)
		typeName := typeOverride(param.Name, param.Required, pythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		fields = append(fields, renderQueryField{
			RawName:      param.Name,
			ArgName:      argName,
			TypeName:     typeName,
			Required:     param.Required,
			DefaultValue: "",
		})
	}
	return fields
}

func renderOperationMethod(doc *openapi.Document, binding operationBinding, async bool) string {
	details := binding.Details
	requestMethod := strings.ToLower(strings.TrimSpace(details.Method))
	paginationMode := ""
	returnType, returnCast := returnTypeInfo(doc, details.ResponseSchema)
	requestBodyType, bodyRequired := requestBodyTypeInfo(doc, details.RequestBodySchema, details.RequestBody)
	useKwargsHeaders := binding.Mapping != nil && binding.Mapping.UseKwargsHeaders
	paramAliases := map[string]string{}
	argTypes := map[string]string{}
	if binding.Mapping != nil && len(binding.Mapping.ParamAliases) > 0 {
		paramAliases = binding.Mapping.ParamAliases
	}
	if binding.Mapping != nil && len(binding.Mapping.ArgTypes) > 0 {
		argTypes = binding.Mapping.ArgTypes
	}
	if binding.Mapping != nil {
		if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" {
			requestMethod = strings.ToLower(methodOverride)
		}
		paginationMode = strings.TrimSpace(binding.Mapping.Pagination)
		if paginationMode == "token" || paginationMode == "number" {
			itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
			if itemType != "" {
				if paginationMode == "token" {
					if async {
						returnType = fmt.Sprintf("AsyncTokenPaged[%s]", itemType)
					} else {
						returnType = fmt.Sprintf("TokenPaged[%s]", itemType)
					}
				} else {
					if async {
						returnType = fmt.Sprintf("AsyncNumberPaged[%s]", itemType)
					} else {
						returnType = fmt.Sprintf("NumberPaged[%s]", itemType)
					}
				}
			}
		}
		if mappedReturnType := strings.TrimSpace(binding.Mapping.ResponseType); mappedReturnType != "" {
			returnType = mappedReturnType
		}
		if mappedReturnCast := strings.TrimSpace(binding.Mapping.ResponseCast); mappedReturnCast != "" {
			returnCast = mappedReturnCast
		} else if strings.TrimSpace(binding.Mapping.ResponseType) != "" {
			returnCast = strings.TrimSpace(binding.Mapping.ResponseType)
		}
	}
	bodyFieldNames := make([]string, 0)
	if binding.Mapping != nil && len(binding.Mapping.BodyFields) > 0 {
		bodyFieldNames = append(bodyFieldNames, binding.Mapping.BodyFields...)
	}
	if binding.Mapping != nil && binding.Mapping.DisableRequestBody {
		requestBodyType = ""
		bodyRequired = false
		bodyFieldNames = nil
	}
	queryFields := buildRenderQueryFields(doc, details, binding.Mapping, paramAliases, argTypes)
	bodyRequiredSet := map[string]bool{}
	if details.RequestBodySchema != nil {
		for _, requiredName := range details.RequestBodySchema.Required {
			bodyRequiredSet[requiredName] = true
		}
	}
	if binding.Mapping != nil && len(binding.Mapping.BodyRequiredFields) > 0 {
		bodyRequiredSet = map[string]bool{}
		for _, requiredName := range binding.Mapping.BodyRequiredFields {
			bodyRequiredSet[requiredName] = true
		}
	}

	pathParamNameMap := map[string]string{}
	signatureArgs := make([]string, 0)
	for _, param := range details.PathParameters {
		name := operationArgName(param.Name, paramAliases)
		pathParamNameMap[param.Name] = name
		typeName := typeOverride(param.Name, true, pythonTypeForSchema(doc, param.Schema, true), argTypes)
		signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
	}
	for _, field := range queryFields {
		if field.Required && field.DefaultValue == "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", field.ArgName, field.TypeName))
		} else if field.DefaultValue != "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", field.ArgName, field.TypeName, field.DefaultValue))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", field.ArgName, field.TypeName))
		}
	}
	for _, param := range details.HeaderParameters {
		name := operationArgName(param.Name, paramAliases)
		typeName := typeOverride(param.Name, param.Required, pythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		if param.Required {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", name, typeName))
		}
	}

	if len(bodyFieldNames) > 0 {
		for _, bodyField := range bodyFieldNames {
			argName := operationArgName(bodyField, paramAliases)
			fieldSchema := bodyFieldSchema(doc, details.RequestBodySchema, bodyField)
			required := bodyRequiredSet[bodyField]
			typeName := typeOverride(bodyField, required, pythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
		}
	} else if requestBodyType != "" {
		if bodyRequired {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: %s", requestBodyType))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = None", requestBodyType))
		}
	}
	if useKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	} else {
		signatureArgs = append(signatureArgs, "headers: Optional[Dict[str, str]] = None")
	}

	methodKeyword := "def"
	requestCall := "self._requester.request"
	if async {
		methodKeyword = "async def"
		requestCall = "await self._requester.arequest"
	}

	var buf bytes.Buffer
	compactSignature := len(bodyFieldNames) == 0 && requestBodyType == ""
	if compactSignature {
		buf.WriteString(fmt.Sprintf("    %s %s(self, *, %s) -> %s:\n", methodKeyword, binding.MethodName, strings.Join(signatureArgs, ", "), returnType))
	} else {
		buf.WriteString(fmt.Sprintf("    %s %s(\n", methodKeyword, binding.MethodName))
		buf.WriteString("        self,\n")
		buf.WriteString("        *,\n")
		for _, arg := range signatureArgs {
			buf.WriteString(fmt.Sprintf("        %s,\n", arg))
		}
		buf.WriteString(fmt.Sprintf("    ) -> %s:\n", returnType))
	}
	summary := details.Summary
	if binding.Mapping != nil {
		pagination := strings.TrimSpace(binding.Mapping.Pagination)
		if pagination == "token" || pagination == "number" {
			summary = ""
		}
		if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" &&
			strings.ToLower(methodOverride) != strings.ToLower(details.Method) {
			summary = ""
		}
	}
	if summary != "" {
		buf.WriteString(fmt.Sprintf("        \"\"\"%s\"\"\"\n", escapeDocstring(summary)))
	}

	urlPath := details.Path
	for rawName, pyName := range pathParamNameMap {
		if rawName == pyName {
			continue
		}
		urlPath = strings.ReplaceAll(urlPath, "{"+rawName+"}", "{"+pyName+"}")
	}
	buf.WriteString(fmt.Sprintf("        url = f\"{self._base_url}%s\"\n", urlPath))

	if len(queryFields) > 0 && paginationMode != "token" && paginationMode != "number" {
		buf.WriteString("        params = dump_exclude_none(\n")
		buf.WriteString("            {\n")
		for _, field := range queryFields {
			buf.WriteString(fmt.Sprintf("                %q: %s,\n", field.RawName, field.ArgName))
		}
		buf.WriteString("            }\n")
		buf.WriteString("        )\n")
	}

	if useKwargsHeaders && ((paginationMode == "token" || paginationMode == "number") || len(details.HeaderParameters) > 0) {
		buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
	}

	if len(details.HeaderParameters) > 0 {
		buf.WriteString("        header_values = dict(headers or {})\n")
		for _, param := range details.HeaderParameters {
			name := operationArgName(param.Name, paramAliases)
			if param.Required {
				buf.WriteString(fmt.Sprintf("        header_values[%q] = str(%s)\n", param.Name, name))
			} else {
				buf.WriteString(fmt.Sprintf("        if %s is not None:\n", name))
				buf.WriteString(fmt.Sprintf("            header_values[%q] = str(%s)\n", param.Name, name))
			}
		}
		buf.WriteString("        headers = header_values\n")
	}

	if paginationMode == "token" && binding.Mapping != nil {
		upperMethod := strings.ToUpper(requestMethod)
		dataClass := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		pageTokenField := strings.TrimSpace(binding.Mapping.PaginationPageTokenField)
		if pageTokenField == "" {
			pageTokenField = "page_token"
		}
		pageSizeField := strings.TrimSpace(binding.Mapping.PaginationPageSizeField)
		if pageSizeField == "" {
			pageSizeField = "page_size"
		}
		tokenExpr := "\"\""
		sizeExpr := "20"
		for _, field := range queryFields {
			if field.RawName == pageTokenField {
				tokenExpr = field.ArgName
			}
			if field.RawName == pageSizeField {
				sizeExpr = field.ArgName
			}
		}
		if async {
			buf.WriteString("        async def request_maker(i_page_token: str, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", upperMethod))
			buf.WriteString("                url,\n")
			buf.WriteString("                params=dump_exclude_none(\n")
			buf.WriteString("                    {\n")
			for _, field := range paginationOrderedFields(queryFields, pageSizeField, pageTokenField) {
				valueExpr := field.ArgName
				if field.RawName == pageTokenField {
					valueExpr = "i_page_token"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("                        %q: %s,\n", field.RawName, valueExpr))
			}
			buf.WriteString("                    }\n")
			buf.WriteString("                ),\n")
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			buf.WriteString("                headers=headers,\n")
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return await AsyncTokenPaged.build(\n")
			buf.WriteString(fmt.Sprintf("            page_token=%s,\n", tokenExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		} else {
			buf.WriteString("        def request_maker(i_page_token: str, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return self._requester.make_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", upperMethod))
			buf.WriteString("                url,\n")
			buf.WriteString("                params=dump_exclude_none(\n")
			buf.WriteString("                    {\n")
			for _, field := range paginationOrderedFields(queryFields, pageSizeField, pageTokenField) {
				valueExpr := field.ArgName
				if field.RawName == pageTokenField {
					valueExpr = "i_page_token"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("                        %q: %s,\n", field.RawName, valueExpr))
			}
			buf.WriteString("                    }\n")
			buf.WriteString("                ),\n")
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			buf.WriteString("                headers=headers,\n")
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return TokenPaged(\n")
			buf.WriteString(fmt.Sprintf("            page_token=%s,\n", tokenExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		}
		return buf.String()
	}
	if paginationMode == "number" && binding.Mapping != nil {
		upperMethod := strings.ToUpper(requestMethod)
		dataClass := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		pageNumField := strings.TrimSpace(binding.Mapping.PaginationPageNumField)
		if pageNumField == "" {
			pageNumField = "page_num"
		}
		pageSizeField := strings.TrimSpace(binding.Mapping.PaginationPageSizeField)
		if pageSizeField == "" {
			pageSizeField = "page_size"
		}
		pageNumExpr := "1"
		sizeExpr := "20"
		for _, field := range queryFields {
			if field.RawName == pageNumField {
				pageNumExpr = field.ArgName
			}
			if field.RawName == pageSizeField {
				sizeExpr = field.ArgName
			}
		}
		if async {
			buf.WriteString("        async def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", upperMethod))
			buf.WriteString("                url,\n")
			buf.WriteString("                params=dump_exclude_none(\n")
			buf.WriteString("                    {\n")
			for _, field := range paginationOrderedFields(queryFields, pageSizeField, pageNumField) {
				valueExpr := field.ArgName
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("                        %q: %s,\n", field.RawName, valueExpr))
			}
			buf.WriteString("                    }\n")
			buf.WriteString("                ),\n")
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			buf.WriteString("                headers=headers,\n")
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return await AsyncNumberPaged.build(\n")
			buf.WriteString(fmt.Sprintf("            page_num=%s,\n", pageNumExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		} else {
			buf.WriteString("        def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return self._requester.make_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", upperMethod))
			buf.WriteString("                url,\n")
			buf.WriteString("                params=dump_exclude_none(\n")
			buf.WriteString("                    {\n")
			for _, field := range paginationOrderedFields(queryFields, pageSizeField, pageNumField) {
				valueExpr := field.ArgName
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("                        %q: %s,\n", field.RawName, valueExpr))
			}
			buf.WriteString("                    }\n")
			buf.WriteString("                ),\n")
			buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			buf.WriteString("                headers=headers,\n")
			buf.WriteString("                stream=False,\n")
			buf.WriteString("            )\n\n")
			buf.WriteString("        return NumberPaged(\n")
			buf.WriteString(fmt.Sprintf("            page_num=%s,\n", pageNumExpr))
			buf.WriteString(fmt.Sprintf("            page_size=%s,\n", sizeExpr))
			buf.WriteString("            requestor=self._requester,\n")
			buf.WriteString("            request_maker=request_maker,\n")
			buf.WriteString("        )\n")
		}
		return buf.String()
	}

	if len(bodyFieldNames) > 0 {
		buf.WriteString("        body = dump_exclude_none(\n")
		buf.WriteString("            {\n")
		for _, bodyField := range bodyFieldNames {
			argName := operationArgName(bodyField, paramAliases)
			buf.WriteString(fmt.Sprintf("                %q: %s,\n", bodyField, argName))
		}
		buf.WriteString("            }\n")
		buf.WriteString("        )\n")
	} else if requestBodyType != "" {
		buf.WriteString("        request_body: Any = None\n")
		if bodyRequired {
			buf.WriteString("        request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		} else {
			buf.WriteString("        if body is not None:\n")
			buf.WriteString("            request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		}
	}
	if useKwargsHeaders && paginationMode != "token" && paginationMode != "number" && len(details.HeaderParameters) == 0 {
		buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
	}

	castExpr := "None"
	if returnCast != "" {
		castExpr = returnCast
	}
	bodyArgExpr := ""
	if len(bodyFieldNames) > 0 {
		bodyArgExpr = "body"
	} else if requestBodyType != "" {
		bodyArgExpr = "request_body"
	}
	callArgs := []string{
		fmt.Sprintf("%q", requestMethod),
		"url",
		"False",
		fmt.Sprintf("cast=%s", castExpr),
	}
	if len(queryFields) > 0 {
		callArgs = append(callArgs, "params=params")
	}
	if bodyArgExpr != "" {
		callArgs = append(callArgs, fmt.Sprintf("body=%s", bodyArgExpr))
	}
	callArgs = append(callArgs, "headers=headers")
	buf.WriteString(fmt.Sprintf("        return %s(%s)\n", requestCall, strings.Join(callArgs, ", ")))

	return buf.String()
}

func operationArgName(raw string, aliases map[string]string) string {
	name := raw
	if len(aliases) > 0 {
		if alias, ok := aliases[raw]; ok && strings.TrimSpace(alias) != "" {
			name = alias
		}
	}
	return normalizePythonIdentifier(name)
}

func bodyFieldSchema(doc *openapi.Document, bodySchema *openapi.Schema, fieldName string) *openapi.Schema {
	if bodySchema == nil {
		return nil
	}
	schema := bodySchema
	if doc != nil {
		schema = doc.ResolveSchema(bodySchema)
	}
	if schema == nil || schema.Properties == nil {
		return nil
	}
	field := schema.Properties[fieldName]
	if field == nil {
		return nil
	}
	if doc != nil {
		return doc.ResolveSchema(field)
	}
	return field
}

func typeOverride(rawName string, required bool, defaultType string, overrides map[string]string) string {
	if len(overrides) == 0 {
		return defaultType
	}
	override := strings.TrimSpace(overrides[rawName])
	if override == "" {
		return defaultType
	}
	if required {
		return override
	}
	if strings.HasPrefix(override, "Optional[") {
		return override
	}
	return "Optional[" + override + "]"
}

func returnTypeInfo(doc *openapi.Document, schema *openapi.Schema) (string, string) {
	_ = doc
	_ = schema
	return "Dict[str, Any]", ""
}

func requestBodyTypeInfo(doc *openapi.Document, schema *openapi.Schema, body *openapi.RequestBody) (string, bool) {
	_ = doc
	_ = schema
	if schema == nil {
		return "", false
	}
	return "Dict[str, Any]", body != nil && body.Required
}

func schemaTypeName(doc *openapi.Document, schema *openapi.Schema) (string, bool) {
	if schema == nil {
		return "", false
	}
	if name, ok := doc.SchemaName(schema); ok {
		return normalizeClassName(name), true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved != nil && resolved != schema {
		if name, ok := doc.SchemaName(resolved); ok {
			return normalizeClassName(name), true
		}
	}
	return "", false
}

func pythonTypeForSchema(doc *openapi.Document, schema *openapi.Schema, required bool) string {
	return pythonTypeForSchemaWithAliases(doc, schema, required, nil)
}

func pythonTypeForSchemaWithAliases(
	doc *openapi.Document,
	schema *openapi.Schema,
	required bool,
	aliases map[string]string,
) string {
	base := pythonTypeForSchemaRequiredWithAliases(doc, schema, aliases)
	if required {
		return base
	}
	if strings.HasPrefix(base, "Optional[") {
		return base
	}
	return "Optional[" + base + "]"
}

func pythonTypeForSchemaRequiredWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) string {
	if schema == nil {
		return "Any"
	}
	if typeName, ok := schemaTypeNameWithAliases(doc, schema, aliases); ok {
		return typeName
	}

	resolved := doc.ResolveSchema(schema)
	if resolved == nil {
		return "Any"
	}
	if typeName, ok := schemaTypeNameWithAliases(doc, resolved, aliases); ok {
		return typeName
	}

	switch resolved.Type {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "List[" + pythonTypeForSchemaRequiredWithAliases(doc, resolved.Items, aliases) + "]"
	case "object":
		return "Dict[str, Any]"
	default:
		if len(resolved.Enum) > 0 {
			return "str"
		}
		return "Any"
	}
}

func schemaTypeNameWithAliases(doc *openapi.Document, schema *openapi.Schema, aliases map[string]string) (string, bool) {
	if schema == nil {
		return "", false
	}
	if name, ok := doc.SchemaName(schema); ok {
		if alias, exists := aliases[name]; exists && strings.TrimSpace(alias) != "" {
			return alias, true
		}
		return normalizeClassName(name), true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved != nil && resolved != schema {
		if name, ok := doc.SchemaName(resolved); ok {
			if alias, exists := aliases[name]; exists && strings.TrimSpace(alias) != "" {
				return alias, true
			}
			return normalizeClassName(name), true
		}
	}
	return "", false
}

func pythonTypeForSchemaRequired(doc *openapi.Document, schema *openapi.Schema) string {
	return pythonTypeForSchemaRequiredWithAliases(doc, schema, nil)
}

func normalizePackageName(name string) string {
	name = normalizePythonIdentifier(name)
	if name == "" {
		return "default"
	}
	return name
}

func normalizePackageDir(sourceDir string, fallback string) string {
	trimmed := strings.TrimSpace(sourceDir)
	if trimmed == "" {
		return fallback
	}
	trimmed = strings.ReplaceAll(trimmed, "\\", "/")
	trimmed = strings.TrimPrefix(trimmed, "./")
	trimmed = strings.TrimPrefix(trimmed, "cozepy/")
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" || trimmed == "." {
		return fallback
	}
	return trimmed
}

func packageClassName(pkgName string) string {
	return normalizeClassName(pkgName)
}

func defaultMethodName(operationID string, path string, method string) string {
	if operationID != "" {
		op := strings.TrimSpace(operationID)
		prefixes := []string{"OpenAPI", "OpenApi", "Openapi", "API", "Api"}
		for _, prefix := range prefixes {
			op = strings.TrimPrefix(op, prefix)
		}
		opSnake := toSnake(op)
		opSnake = strings.TrimPrefix(opSnake, "open_api_")
		if opSnake != "" {
			return normalizeMethodName(opSnake)
		}
	}

	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		return normalizeMethodName(toSnake(part))
	}
	return normalizeMethodName(method)
}

func normalizeMethodName(value string) string {
	name := normalizePythonIdentifier(toSnake(value))
	if name == "" {
		return "call"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		return "method_" + name
	}
	return name
}

func normalizeClassName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "GeneratedModel"
	}

	snake := toSnake(trimmed)
	if snake == "" {
		return "GeneratedModel"
	}
	parts := strings.Split(snake, "_")
	if len(parts) == 0 {
		return "GeneratedModel"
	}
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}

	name := strings.Join(parts, "")
	if name == "" {
		name = "GeneratedModel"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "Model" + name
	}
	return name
}

func normalizePythonIdentifier(value string) string {
	parts := splitIdentifier(value)
	if len(parts) == 0 {
		return ""
	}
	name := strings.ToLower(strings.Join(parts, "_"))
	name = collapseUnderscore(name)
	if name == "" {
		return ""
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "_" + name
	}
	if pythonReservedWords[name] {
		name = name + "_"
	}
	return name
}

func toSnake(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var out []rune
	prevLowerOrDigit := false
	for _, r := range value {
		if unicode.IsUpper(r) {
			if prevLowerOrDigit && len(out) > 0 && out[len(out)-1] != '_' {
				out = append(out, '_')
			}
			out = append(out, unicode.ToLower(r))
			prevLowerOrDigit = false
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out = append(out, unicode.ToLower(r))
			prevLowerOrDigit = unicode.IsLower(r) || unicode.IsDigit(r)
			continue
		}
		if len(out) > 0 && out[len(out)-1] != '_' {
			out = append(out, '_')
		}
		prevLowerOrDigit = false
	}
	return strings.Trim(collapseUnderscore(string(out)), "_")
}

func splitIdentifier(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parts := make([]string, 0)
	var current []rune
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current = append(current, r)
			continue
		}
		if len(current) > 0 {
			parts = append(parts, string(current))
			current = current[:0]
		}
	}
	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func collapseUnderscore(value string) string {
	var out []rune
	lastUnderscore := false
	for _, r := range value {
		if r == '_' {
			if lastUnderscore {
				continue
			}
			lastUnderscore = true
			out = append(out, r)
			continue
		}
		lastUnderscore = false
		out = append(out, r)
	}
	return string(out)
}

func escapeDocstring(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\"\"\"", "\"")
	return strings.TrimSpace(value)
}

var pythonReservedWords = map[string]bool{
	"false":    true,
	"none":     true,
	"true":     true,
	"and":      true,
	"as":       true,
	"assert":   true,
	"async":    true,
	"await":    true,
	"break":    true,
	"class":    true,
	"continue": true,
	"def":      true,
	"del":      true,
	"elif":     true,
	"else":     true,
	"except":   true,
	"finally":  true,
	"for":      true,
	"from":     true,
	"global":   true,
	"if":       true,
	"import":   true,
	"in":       true,
	"is":       true,
	"lambda":   true,
	"nonlocal": true,
	"not":      true,
	"or":       true,
	"pass":     true,
	"raise":    true,
	"return":   true,
	"try":      true,
	"while":    true,
	"with":     true,
	"yield":    true,
}

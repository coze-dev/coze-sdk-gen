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
	"github.com/coze-dev/coze-sdk-gen/internal/fsutil"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type operationBinding struct {
	PackageName string
	MethodName  string
	Details     openapi.OperationDetails
}

type packageMeta struct {
	Name       string
	ModulePath string
	DirPath    string
}

type fileWriter struct {
	count int
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
	schemaNames := collectSchemaNames(doc, bindings)

	if cfg.Compatibility.EnforceZeroDiff {
		return generatePythonWithLegacyCompatibility(cfg, doc, packages, packageMetas, schemaNames)
	}

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{}
	if err := writePythonSDK(cfg.OutputSDK, doc, packages, packageMetas, schemaNames, writer); err != nil {
		return Result{}, err
	}

	return Result{
		GeneratedFiles: writer.count,
		GeneratedOps:   len(bindings),
	}, nil
}

func generatePythonWithLegacyCompatibility(
	cfg *config.Config,
	doc *openapi.Document,
	packages map[string][]operationBinding,
	packageMetas map[string]packageMeta,
	schemaNames []string,
) (Result, error) {
	tmpDir, err := os.MkdirTemp("", "coze-sdk-gen-swagger-*")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary generation directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpWriter := &fileWriter{}
	if err := writePythonSDK(tmpDir, doc, packages, packageMetas, schemaNames, tmpWriter); err != nil {
		return Result{}, err
	}

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	excludes := make([]string, 0, len(cfg.Copy.Exclude)+1)
	excludes = append(excludes, cfg.Copy.Exclude...)
	excludes = append(excludes, ".git")

	copyResult, err := fsutil.CopySelected(cfg.SourceSDK, cfg.OutputSDK, cfg.Copy.Include, excludes)
	if err != nil {
		return Result{}, fmt.Errorf("copy legacy sdk for strict compatibility: %w", err)
	}

	return Result{
		GeneratedFiles: copyResult.CopiedFiles,
		GeneratedOps:   countBindings(packages),
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
				for _, sdkMethod := range mapping.SDKMethods {
					pkgName, methodName, ok := config.ParseSDKMethod(sdkMethod)
					if !ok {
						continue
					}
					pkg, ok := cfg.ResolvePackage(details.Path, pkgName)
					if !ok {
						continue
					}
					bindings = append(bindings, operationBinding{
						PackageName: normalizePackageName(pkg.Name),
						MethodName:  normalizeMethodName(methodName),
						Details:     details,
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
		})
	}

	return deduplicateBindings(bindings)
}

func deduplicateBindings(bindings []operationBinding) []operationBinding {
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].PackageName != bindings[j].PackageName {
			return bindings[i].PackageName < bindings[j].PackageName
		}
		if bindings[i].MethodName != bindings[j].MethodName {
			return bindings[i].MethodName < bindings[j].MethodName
		}
		if bindings[i].Details.Path != bindings[j].Details.Path {
			return bindings[i].Details.Path < bindings[j].Details.Path
		}
		return bindings[i].Details.Method < bindings[j].Details.Method
	})

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
			if pkgOps[pkgName][i].MethodName != pkgOps[pkgName][j].MethodName {
				return pkgOps[pkgName][i].MethodName < pkgOps[pkgName][j].MethodName
			}
			if pkgOps[pkgName][i].Details.Path != pkgOps[pkgName][j].Details.Path {
				return pkgOps[pkgName][i].Details.Path < pkgOps[pkgName][j].Details.Path
			}
			return pkgOps[pkgName][i].Details.Method < pkgOps[pkgName][j].Details.Method
		})
	}
	return pkgOps
}

func collectSchemaNames(doc *openapi.Document, bindings []operationBinding) []string {
	names := map[string]struct{}{}
	for _, binding := range bindings {
		refs := doc.CollectSchemaRefsFromOperation(binding.Details)
		for _, ref := range refs {
			names[ref] = struct{}{}
		}
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func countBindings(packages map[string][]operationBinding) int {
	total := 0
	for _, bindings := range packages {
		total += len(bindings)
	}
	return total
}

func buildPackageMeta(cfg *config.Config, packages map[string][]operationBinding) map[string]packageMeta {
	metas := map[string]packageMeta{}
	for _, pkg := range cfg.API.Packages {
		name := normalizePackageName(pkg.Name)
		dir := normalizePackageDir(pkg.SourceDir, name)
		metas[name] = packageMeta{
			Name:       name,
			ModulePath: strings.ReplaceAll(dir, "/", "."),
			DirPath:    dir,
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
	schemaNames []string,
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
	sort.Strings(pkgNames)

	if err := writer.write(filepath.Join(rootDir, "__init__.py"), renderRootInit()); err != nil {
		return err
	}
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
	if err := writer.write(filepath.Join(rootDir, "types.py"), renderTypesPy(doc, schemaNames)); err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "coze.py"), renderCozePy(pkgNames, packageMetas)); err != nil {
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

	return nil
}

func (w *fileWriter) write(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	w.count++
	return nil
}

func renderRootInit() string {
	return "from .coze import AsyncCoze, Coze\n"
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

func renderCozePy(packageNames []string, packageMetas map[string]packageMeta) string {
	var buf bytes.Buffer
	buf.WriteString("from typing import Optional\n\n")
	buf.WriteString("from .config import COZE_COM_BASE_URL\n")
	buf.WriteString("from .request import Requester\n")
	buf.WriteString("from .util import remove_url_trailing_slash\n")
	for _, pkgName := range packageNames {
		meta := packageMetas[pkgName]
		className := packageClassName(pkgName)
		buf.WriteString(fmt.Sprintf("from .%s import Async%sClient, %sClient\n", meta.ModulePath, className, className))
	}
	buf.WriteString("\n\n")

	buf.WriteString("class Coze(object):\n")
	buf.WriteString("    def __init__(\n")
	buf.WriteString("        self,\n")
	buf.WriteString("        token: Optional[str] = None,\n")
	buf.WriteString("        base_url: str = COZE_COM_BASE_URL,\n")
	buf.WriteString("        requester: Optional[Requester] = None,\n")
	buf.WriteString("    ):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester or Requester(token=token)\n")
	for _, pkgName := range packageNames {
		buf.WriteString(fmt.Sprintf("        self._%s: Optional[%sClient] = None\n", pkgName, packageClassName(pkgName)))
	}
	buf.WriteString("\n")
	for _, pkgName := range packageNames {
		className := packageClassName(pkgName)
		buf.WriteString("    @property\n")
		buf.WriteString(fmt.Sprintf("    def %s(self) -> %sClient:\n", pkgName, className))
		buf.WriteString(fmt.Sprintf("        if self._%s is None:\n", pkgName))
		buf.WriteString(fmt.Sprintf("            self._%s = %sClient(self._base_url, self._requester)\n", pkgName, className))
		buf.WriteString(fmt.Sprintf("        return self._%s\n\n", pkgName))
	}

	buf.WriteString("class AsyncCoze(object):\n")
	buf.WriteString("    def __init__(\n")
	buf.WriteString("        self,\n")
	buf.WriteString("        token: Optional[str] = None,\n")
	buf.WriteString("        base_url: str = COZE_COM_BASE_URL,\n")
	buf.WriteString("        requester: Optional[Requester] = None,\n")
	buf.WriteString("    ):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester or Requester(token=token)\n")
	for _, pkgName := range packageNames {
		buf.WriteString(fmt.Sprintf("        self._%s: Optional[Async%sClient] = None\n", pkgName, packageClassName(pkgName)))
	}
	buf.WriteString("\n")
	for _, pkgName := range packageNames {
		className := packageClassName(pkgName)
		buf.WriteString("    @property\n")
		buf.WriteString(fmt.Sprintf("    def %s(self) -> Async%sClient:\n", pkgName, className))
		buf.WriteString(fmt.Sprintf("        if self._%s is None:\n", pkgName))
		buf.WriteString(fmt.Sprintf("            self._%s = Async%sClient(self._base_url, self._requester)\n", pkgName, className))
		buf.WriteString(fmt.Sprintf("        return self._%s\n\n", pkgName))
	}

	return buf.String()
}

func renderTypesPy(doc *openapi.Document, schemaNames []string) string {
	var buf bytes.Buffer
	buf.WriteString("from typing import Any, Dict, List, Optional\n\n")
	buf.WriteString("from pydantic import Field\n\n")
	buf.WriteString("from .model import CozeModel\n\n\n")

	for _, schemaName := range schemaNames {
		schema, ok := doc.Components.Schemas[schemaName]
		if !ok || schema == nil {
			continue
		}

		className := normalizeClassName(schemaName)
		buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", className))
		buf.WriteString(fmt.Sprintf("    \"\"\"Generated from OpenAPI schema `%s`.\"\"\"\n", schemaName))

		resolved := doc.ResolveSchema(schema)
		if resolved == nil {
			buf.WriteString("    pass\n\n\n")
			continue
		}

		required := map[string]struct{}{}
		for _, field := range resolved.Required {
			required[field] = struct{}{}
		}

		propNames := make([]string, 0, len(resolved.Properties))
		for propName := range resolved.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		if len(propNames) == 0 {
			buf.WriteString("    pass\n\n\n")
			continue
		}

		for _, propName := range propNames {
			propSchema := resolved.Properties[propName]
			_, isRequired := required[propName]
			pyFieldName := normalizePythonIdentifier(propName)
			pyType := pythonTypeForSchema(doc, propSchema, isRequired)
			suffix := ""
			if !isRequired {
				suffix = " = None"
			}
			if pyFieldName != propName {
				if isRequired {
					buf.WriteString(fmt.Sprintf("    %s: %s = Field(alias=%q)\n", pyFieldName, pyType, propName))
				} else {
					buf.WriteString(fmt.Sprintf("    %s: %s = Field(default=None, alias=%q)\n", pyFieldName, pyType, propName))
				}
				continue
			}
			buf.WriteString(fmt.Sprintf("    %s: %s%s\n", pyFieldName, pyType, suffix))
		}
		buf.WriteString("\n\n")
	}

	if len(schemaNames) == 0 {
		buf.WriteString("class EmptyModel(CozeModel):\n")
		buf.WriteString("    pass\n")
	}

	return buf.String()
}

func renderPackageModule(doc *openapi.Document, meta packageMeta, bindings []operationBinding) string {
	var buf bytes.Buffer
	buf.WriteString("from typing import Any, Dict, Optional\n\n")
	buf.WriteString("from ..request import Requester\n")
	buf.WriteString("from ..util import dump_exclude_none, remove_url_trailing_slash\n")

	imports := collectTypeImports(doc, bindings)
	if len(imports) > 0 {
		buf.WriteString(fmt.Sprintf("from ..types import %s\n", strings.Join(imports, ", ")))
	}
	buf.WriteString("\n\n")

	className := packageClassName(meta.Name)
	buf.WriteString(fmt.Sprintf("class %sClient(object):\n", className))
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n\n")

	for _, binding := range bindings {
		buf.WriteString(renderOperationMethod(doc, binding, false))
		buf.WriteString("\n")
	}

	buf.WriteString(fmt.Sprintf("class Async%sClient(object):\n", className))
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n\n")

	for _, binding := range bindings {
		buf.WriteString(renderOperationMethod(doc, binding, true))
		buf.WriteString("\n")
	}

	return buf.String()
}

func collectTypeImports(doc *openapi.Document, bindings []operationBinding) []string {
	seen := map[string]struct{}{}
	for _, binding := range bindings {
		if name, ok := schemaTypeName(doc, binding.Details.RequestBodySchema); ok {
			seen[name] = struct{}{}
		}
		if name, ok := schemaTypeName(doc, binding.Details.ResponseSchema); ok {
			seen[name] = struct{}{}
		}
		for _, param := range binding.Details.Parameters {
			if name, ok := schemaTypeName(doc, param.Schema); ok {
				seen[name] = struct{}{}
			}
		}
	}

	imports := make([]string, 0, len(seen))
	for name := range seen {
		imports = append(imports, name)
	}
	sort.Strings(imports)
	return imports
}

func renderOperationMethod(doc *openapi.Document, binding operationBinding, async bool) string {
	details := binding.Details
	returnType, returnCast := returnTypeInfo(doc, details.ResponseSchema)
	requestBodyType, bodyRequired := requestBodyTypeInfo(doc, details.RequestBodySchema, details.RequestBody)

	pathParamNameMap := map[string]string{}
	signatureArgs := make([]string, 0)
	for _, param := range details.PathParameters {
		name := normalizePythonIdentifier(param.Name)
		pathParamNameMap[param.Name] = name
		typeName := pythonTypeForSchema(doc, param.Schema, true)
		signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
	}
	for _, param := range details.QueryParameters {
		name := normalizePythonIdentifier(param.Name)
		typeName := pythonTypeForSchema(doc, param.Schema, param.Required)
		if param.Required {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", name, typeName))
		}
	}
	for _, param := range details.HeaderParameters {
		name := normalizePythonIdentifier(param.Name)
		typeName := pythonTypeForSchema(doc, param.Schema, param.Required)
		if param.Required {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", name, typeName))
		}
	}

	if requestBodyType != "" {
		if bodyRequired {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: %s", requestBodyType))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = None", requestBodyType))
		}
	}
	signatureArgs = append(signatureArgs, "headers: Optional[Dict[str, str]] = None")

	methodKeyword := "def"
	requestCall := "self._requester.request"
	if async {
		methodKeyword = "async def"
		requestCall = "await self._requester.arequest"
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("    %s %s(\n", methodKeyword, binding.MethodName))
	buf.WriteString("        self,\n")
	buf.WriteString("        *,\n")
	for _, arg := range signatureArgs {
		buf.WriteString(fmt.Sprintf("        %s,\n", arg))
	}
	buf.WriteString(fmt.Sprintf("    ) -> %s:\n", returnType))
	if details.Summary != "" {
		buf.WriteString(fmt.Sprintf("        \"\"\"%s\"\"\"\n", escapeDocstring(details.Summary)))
	}

	urlPath := details.Path
	for rawName, pyName := range pathParamNameMap {
		if rawName == pyName {
			continue
		}
		urlPath = strings.ReplaceAll(urlPath, "{"+rawName+"}", "{"+pyName+"}")
	}
	buf.WriteString(fmt.Sprintf("        url = f\"{self._base_url}%s\"\n", urlPath))

	if len(details.QueryParameters) > 0 {
		buf.WriteString("        params = dump_exclude_none(\n")
		buf.WriteString("            {\n")
		for _, param := range details.QueryParameters {
			name := normalizePythonIdentifier(param.Name)
			buf.WriteString(fmt.Sprintf("                %q: %s,\n", param.Name, name))
		}
		buf.WriteString("            }\n")
		buf.WriteString("        )\n")
	} else {
		buf.WriteString("        params = None\n")
	}

	if len(details.HeaderParameters) > 0 {
		buf.WriteString("        header_values = dict(headers or {})\n")
		for _, param := range details.HeaderParameters {
			name := normalizePythonIdentifier(param.Name)
			if param.Required {
				buf.WriteString(fmt.Sprintf("        header_values[%q] = str(%s)\n", param.Name, name))
			} else {
				buf.WriteString(fmt.Sprintf("        if %s is not None:\n", name))
				buf.WriteString(fmt.Sprintf("            header_values[%q] = str(%s)\n", param.Name, name))
			}
		}
		buf.WriteString("        headers = header_values\n")
	}

	if requestBodyType != "" {
		buf.WriteString("        request_body: Any = None\n")
		if bodyRequired {
			buf.WriteString("        request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		} else {
			buf.WriteString("        if body is not None:\n")
			buf.WriteString("            request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		}
	} else {
		buf.WriteString("        request_body = None\n")
	}

	castExpr := "None"
	if returnCast != "" {
		castExpr = returnCast
	}
	buf.WriteString(fmt.Sprintf(
		"        return %s(%q, url, params=params, headers=headers, body=request_body, cast=%s)\n",
		requestCall,
		strings.ToLower(details.Method),
		castExpr,
	))

	return buf.String()
}

func returnTypeInfo(doc *openapi.Document, schema *openapi.Schema) (string, string) {
	if schema == nil {
		return "Any", ""
	}
	if typeName, ok := schemaTypeName(doc, schema); ok {
		return typeName, typeName
	}
	return pythonTypeForSchema(doc, schema, true), ""
}

func requestBodyTypeInfo(doc *openapi.Document, schema *openapi.Schema, body *openapi.RequestBody) (string, bool) {
	if schema == nil {
		return "", false
	}
	if typeName, ok := schemaTypeName(doc, schema); ok {
		return typeName, body != nil && body.Required
	}
	return pythonTypeForSchema(doc, schema, true), body != nil && body.Required
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
	base := pythonTypeForSchemaRequired(doc, schema)
	if required {
		return base
	}
	if strings.HasPrefix(base, "Optional[") {
		return base
	}
	return "Optional[" + base + "]"
}

func pythonTypeForSchemaRequired(doc *openapi.Document, schema *openapi.Schema) string {
	if schema == nil {
		return "Any"
	}
	if typeName, ok := schemaTypeName(doc, schema); ok {
		return typeName
	}

	resolved := doc.ResolveSchema(schema)
	if resolved == nil {
		return "Any"
	}
	if typeName, ok := schemaTypeName(doc, resolved); ok {
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
		return "List[" + pythonTypeForSchemaRequired(doc, resolved.Items) + "]"
	case "object":
		return "Dict[str, Any]"
	default:
		if len(resolved.Enum) > 0 {
			return "str"
		}
		return "Any"
	}
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

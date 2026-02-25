package python

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/generator/fsutil"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type OperationBinding struct {
	PackageName string
	MethodName  string
	Details     openapi.OperationDetails
	Mapping     *config.OperationMapping
	Order       int
}

type PackageMeta struct {
	Name         string
	ModulePath   string
	DirPath      string
	Package      *config.Package
	ChildClients []childClient
}

type childClient struct {
	Attribute  string
	Module     string
	SyncClass  string
	AsyncClass string
}

type fileWriter struct {
	count   int
	written map[string]struct{}
}

type Result struct {
	GeneratedFiles int
	GeneratedOps   int
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

	if err := fsutil.CleanOutputDirPreserveEntries(cfg.OutputSDK, cfg.DiffIgnorePathsForLanguage("python")); err != nil {
		return Result{}, fmt.Errorf("prepare output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{
		written: map[string]struct{}{},
	}
	if err := writePythonSDK(cfg, doc, packages, packageMetas, writer); err != nil {
		return Result{}, err
	}

	return Result{
		GeneratedFiles: writer.count,
		GeneratedOps:   len(bindings),
	}, nil
}

func writePythonSDK(
	cfg *config.Config,
	doc *openapi.Document,
	packages map[string][]OperationBinding,
	packageMetas map[string]PackageMeta,
	writer *fileWriter,
) error {
	outputDir := cfg.OutputSDK
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
		if packageHasConfiguredContent(meta.Package) || len(meta.ChildClients) > 0 {
			pkgNames = append(pkgNames, pkgName)
		}
	}
	sort.Strings(pkgNames)

	runtimeFiles := []struct {
		baseDir     string
		name        string
		render      func() (string, error)
		skipIfEmpty bool
	}{
		{baseDir: rootDir, name: "config.py", render: RenderConfigPy},
		{baseDir: rootDir, name: "util.py", render: RenderUtilPy},
		{baseDir: rootDir, name: "model.py", render: RenderModelPy},
		{baseDir: rootDir, name: "request.py", render: RenderRequestPy},
		{baseDir: rootDir, name: "log.py", render: RenderLogPy},
		{baseDir: rootDir, name: "exception.py", render: RenderExceptionPy},
		{baseDir: rootDir, name: "version.py", render: RenderVersionPy},
		{baseDir: outputDir, name: "pyproject.toml", render: RenderPyprojectToml},
		{
			baseDir: rootDir,
			name:    "py.typed",
			render: func() (string, error) {
				return "", nil
			},
		},
		{
			baseDir: rootDir,
			name:    "coze.py",
			render: func() (string, error) {
				return renderCozePy(cfg, packageMetas)
			},
			skipIfEmpty: true,
		},
	}
	for _, file := range runtimeFiles {
		content, err := file.render()
		if err != nil {
			return err
		}
		if file.skipIfEmpty && content == "" {
			continue
		}
		if err := writer.write(filepath.Join(file.baseDir, file.name), content); err != nil {
			return err
		}
	}

	for _, pkgName := range pkgNames {
		meta := packageMetas[pkgName]
		pkgDir := filepath.Join(rootDir, meta.DirPath)
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			return fmt.Errorf("create package directory %q: %w", pkgDir, err)
		}
		content := RenderPackageModuleWithComments(doc, meta, packages[pkgName], cfg.CommentOverrides)
		if err := writer.write(filepath.Join(pkgDir, "__init__.py"), content); err != nil {
			return err
		}
	}
	if err := writePythonSpecialAssets(rootDir, writer); err != nil {
		return err
	}

	return nil
}

func writePythonSpecialAssets(rootDir string, writer *fileWriter) error {
	// These modules are not represented by the OpenAPI schema (OAuth/websocket runtime wiring).
	// Keep them in an explicit, minimal whitelist and render through the generator pipeline.
	specialAssets := []struct {
		relPath string
		asset   string
	}{
		{relPath: "__init__.py", asset: "special/cozepy/__init__.py.tpl"},
		{relPath: "auth/__init__.py", asset: "special/cozepy/auth/__init__.py.tpl"},
		{relPath: "websockets/__init__.py", asset: "special/cozepy/websockets/__init__.py.tpl"},
		{relPath: "websockets/audio/__init__.py", asset: "special/cozepy/websockets/audio/__init__.py.tpl"},
		{relPath: "websockets/audio/speech/__init__.py", asset: "special/cozepy/websockets/audio/speech/__init__.py.tpl"},
		{relPath: "websockets/audio/transcriptions/__init__.py", asset: "special/cozepy/websockets/audio/transcriptions/__init__.py.tpl"},
		{relPath: "websockets/chat/__init__.py", asset: "special/cozepy/websockets/chat/__init__.py.tpl"},
		{relPath: "websockets/ws.py", asset: "special/cozepy/websockets/ws.py.tpl"},
	}
	for _, item := range specialAssets {
		content, err := RenderPythonRawAsset(item.asset)
		if err != nil {
			return err
		}
		if err := writer.write(filepath.Join(rootDir, item.relPath), content); err != nil {
			return err
		}
	}
	return nil
}

type rootService struct {
	Attribute  string
	ModuleDir  string
	SyncClass  string
	AsyncClass string
}

func renderCozePy(cfg *config.Config, packageMetas map[string]PackageMeta) (string, error) {
	if cfg == nil {
		return "", nil
	}
	services := collectRootServices(cfg, packageMetas)
	if len(services) == 0 {
		return "", nil
	}

	syncOrder := []string{
		"bots",
		"workspaces",
		"conversations",
		"chat",
		"files",
		"workflows",
		"knowledge",
		"datasets",
		"audio",
		"templates",
		"users",
		"websockets",
		"variables",
		"apps",
		"enterprises",
		"api_apps",
		"connectors",
		"folders",
	}
	asyncOrder := []string{
		"bots",
		"chat",
		"conversations",
		"files",
		"knowledge",
		"datasets",
		"workflows",
		"workspaces",
		"audio",
		"templates",
		"users",
		"websockets",
		"variables",
		"apps",
		"enterprises",
		"api_apps",
		"connectors",
		"folders",
	}
	syncFieldOrder := []string{
		"bots",
		"workspaces",
		"conversations",
		"chat",
		"connectors",
		"files",
		"workflows",
		"knowledge",
		"datasets",
		"audio",
		"templates",
		"users",
		"websockets",
		"variables",
		"apps",
		"enterprises",
		"api_apps",
		"folders",
	}
	asyncFieldOrder := []string{
		"bots",
		"chat",
		"connectors",
		"conversations",
		"files",
		"knowledge",
		"datasets",
		"workflows",
		"workspaces",
		"audio",
		"templates",
		"users",
		"websockets",
		"variables",
		"apps",
		"enterprises",
		"api_apps",
		"folders",
	}
	typeCheckingOrder := []string{
		"api_apps",
		"apps",
		"audio",
		"bots",
		"chat",
		"connectors",
		"conversations",
		"datasets",
		"enterprises",
		"files",
		"folders",
		"knowledge",
		"templates",
		"users",
		"variables",
		"websockets",
		"workflows",
		"workspaces",
	}

	syncServices := orderRootServices(services, syncOrder)
	asyncServices := orderRootServices(services, asyncOrder)
	syncFieldServices := orderRootServices(services, syncFieldOrder)
	asyncFieldServices := orderRootServices(services, asyncFieldOrder)
	typeCheckingServices := orderRootServices(services, typeCheckingOrder)

	var buf bytes.Buffer
	buf.WriteString("import warnings\n")
	buf.WriteString("from typing import TYPE_CHECKING, Optional\n\n")
	buf.WriteString("from cozepy.auth import Auth, SyncAuth\n")
	buf.WriteString("from cozepy.config import COZE_COM_BASE_URL\n")
	buf.WriteString("from cozepy.request import AsyncHTTPClient, Requester, SyncHTTPClient\n")
	buf.WriteString("from cozepy.util import remove_url_trailing_slash\n\n")
	buf.WriteString("if TYPE_CHECKING:\n")
	for _, svc := range typeCheckingServices {
		importNames := formatTypeImportPair(svc.SyncClass, svc.AsyncClass)
		if svc.Attribute == "knowledge" {
			buf.WriteString(fmt.Sprintf("    from .%s import %s  # deprecated\n", svc.ModuleDir, importNames))
			continue
		}
		buf.WriteString(fmt.Sprintf("    from .%s import %s\n", svc.ModuleDir, importNames))
	}
	buf.WriteString("\n\n")

	buf.WriteString("class Coze(object):\n")
	buf.WriteString("    def __init__(\n")
	buf.WriteString("        self,\n")
	buf.WriteString("        auth: Auth,\n")
	buf.WriteString("        base_url: str = COZE_COM_BASE_URL,\n")
	buf.WriteString("        http_client: Optional[SyncHTTPClient] = None,\n")
	buf.WriteString("    ):\n")
	buf.WriteString("        self._auth = auth\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = Requester(auth=auth, sync_client=http_client)\n\n")
	buf.WriteString("        # service client\n")
	for _, svc := range syncFieldServices {
		line := fmt.Sprintf("        self._%s: Optional[%s] = None\n", svc.Attribute, svc.SyncClass)
		if svc.Attribute == "knowledge" {
			line = strings.TrimRight(line, "\n") + "  # deprecated\n"
		}
		buf.WriteString(line)
	}
	buf.WriteString("\n")
	for _, svc := range syncServices {
		buf.WriteString("    @property\n")
		buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", svc.Attribute, svc.SyncClass))
		if svc.Attribute == "knowledge" {
			buf.WriteString("        warnings.warn(\n")
			buf.WriteString("            \"The 'coze.knowledge' module is deprecated and will be removed in a future version. \"\n")
			buf.WriteString("            \"Please use 'coze.datasets' instead.\",\n")
			buf.WriteString("            DeprecationWarning,\n")
			buf.WriteString("            stacklevel=2,\n")
			buf.WriteString("        )\n")
		}
		buf.WriteString(fmt.Sprintf("        if not self._%s:\n", svc.Attribute))
		importStmt := fmt.Sprintf("from .%s import %s", svc.ModuleDir, svc.SyncClass)
		if useAbsoluteServiceImport(svc.Attribute) {
			importStmt = fmt.Sprintf("from cozepy.%s import %s", svc.ModuleDir, svc.SyncClass)
		}
		buf.WriteString(fmt.Sprintf("            %s\n\n", importStmt))
		buf.WriteString(fmt.Sprintf("            self._%s = %s(self._base_url, self._requester)\n", svc.Attribute, svc.SyncClass))
		buf.WriteString(fmt.Sprintf("        return self._%s\n\n", svc.Attribute))
	}

	buf.WriteString("class AsyncCoze(object):\n")
	buf.WriteString("    def __init__(\n")
	buf.WriteString("        self,\n")
	buf.WriteString("        auth: Auth,\n")
	buf.WriteString("        base_url: str = COZE_COM_BASE_URL,\n")
	buf.WriteString("        http_client: Optional[AsyncHTTPClient] = None,\n")
	buf.WriteString("    ):\n")
	buf.WriteString("        self._auth = auth\n")
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        if isinstance(auth, SyncAuth):\n")
	buf.WriteString("            warnings.warn(\n")
	buf.WriteString("                \"The 'coze.SyncAuth' use for AsyncCoze is deprecated and will be removed in a future version. \"\n")
	buf.WriteString("                \"Please use 'coze.AsyncAuth' instead.\",\n")
	buf.WriteString("                DeprecationWarning,\n")
	buf.WriteString("                stacklevel=2,\n")
	buf.WriteString("            )\n\n")
	buf.WriteString("        self._requester = Requester(auth=auth, async_client=http_client)\n\n")
	buf.WriteString("        # service client\n")
	for _, svc := range asyncFieldServices {
		line := fmt.Sprintf("        self._%s: Optional[%s] = None\n", svc.Attribute, svc.AsyncClass)
		if svc.Attribute == "knowledge" {
			line = strings.TrimRight(line, "\n") + "  # deprecated\n"
		}
		buf.WriteString(line)
	}
	buf.WriteString("\n")
	for _, svc := range asyncServices {
		buf.WriteString("    @property\n")
		buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", svc.Attribute, svc.AsyncClass))
		if svc.Attribute == "knowledge" {
			buf.WriteString("        warnings.warn(\n")
			buf.WriteString("            \"The 'coze.knowledge' module is deprecated and will be removed in a future version. \"\n")
			buf.WriteString("            \"Please use 'coze.datasets' instead.\",\n")
			buf.WriteString("            DeprecationWarning,\n")
			buf.WriteString("            stacklevel=2,\n")
			buf.WriteString("        )\n")
		}
		buf.WriteString(fmt.Sprintf("        if not self._%s:\n", svc.Attribute))
		importStmt := fmt.Sprintf("from .%s import %s", svc.ModuleDir, svc.AsyncClass)
		if useAbsoluteServiceImport(svc.Attribute) {
			importStmt = fmt.Sprintf("from cozepy.%s import %s", svc.ModuleDir, svc.AsyncClass)
		}
		buf.WriteString(fmt.Sprintf("            %s\n\n", importStmt))
		buf.WriteString(fmt.Sprintf("            self._%s = %s(self._base_url, self._requester)\n", svc.Attribute, svc.AsyncClass))
		buf.WriteString(fmt.Sprintf("        return self._%s\n\n", svc.Attribute))
	}

	return strings.TrimRight(buf.String(), "\n") + "\n", nil
}

func collectRootServices(cfg *config.Config, packageMetas map[string]PackageMeta) []rootService {
	services := make([]rootService, 0)
	seen := map[string]struct{}{}
	for _, pkg := range cfg.API.Packages {
		name := NormalizePackageName(pkg.Name)
		meta, ok := packageMetas[name]
		if !ok {
			continue
		}
		dir := strings.TrimSpace(meta.DirPath)
		if dir == "" || strings.Contains(dir, "/") {
			continue
		}
		attr := NormalizePythonIdentifier(dir)
		if attr == "" {
			continue
		}
		if _, exists := seen[attr]; exists {
			continue
		}
		seen[attr] = struct{}{}
		services = append(services, rootService{
			Attribute:  attr,
			ModuleDir:  dir,
			SyncClass:  packageClientClassName(meta, false),
			AsyncClass: packageClientClassName(meta, true),
		})
	}
	return services
}

func orderRootServices(services []rootService, order []string) []rootService {
	if len(services) == 0 {
		return services
	}
	byAttr := map[string]rootService{}
	for _, svc := range services {
		byAttr[svc.Attribute] = svc
	}
	ordered := make([]rootService, 0, len(services))
	seen := map[string]struct{}{}
	for _, attr := range order {
		attr = strings.TrimSpace(attr)
		if attr == "" {
			continue
		}
		svc, ok := byAttr[attr]
		if !ok {
			continue
		}
		ordered = append(ordered, svc)
		seen[attr] = struct{}{}
	}
	remaining := make([]string, 0)
	for attr := range byAttr {
		if _, ok := seen[attr]; ok {
			continue
		}
		remaining = append(remaining, attr)
	}
	sort.Strings(remaining)
	for _, attr := range remaining {
		ordered = append(ordered, byAttr[attr])
	}
	return ordered
}

func useAbsoluteServiceImport(attr string) bool {
	switch strings.TrimSpace(attr) {
	case "bots", "chat":
		return true
	default:
		return false
	}
}

func packageHasConfiguredContent(pkg *config.Package) bool {
	if pkg == nil {
		return false
	}
	return len(pkg.ModelSchemas) > 0 ||
		len(pkg.EmptyModels) > 0 ||
		len(pkg.PreModelCode) > 0 ||
		len(pkg.TopLevelCode) > 0 ||
		len(pkg.SyncExtraMethods) > 0 ||
		len(pkg.AsyncExtraMethods) > 0
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

func RenderConfigPy() (string, error) {
	return RenderPythonTemplate("config.py.tpl", map[string]any{})
}

func RenderUtilPy() (string, error) {
	return RenderPythonTemplate("util.py.tpl", map[string]any{})
}

func RenderModelPy() (string, error) {
	return RenderPythonTemplate("model.py.tpl", map[string]any{})
}

func RenderRequestPy() (string, error) {
	return RenderPythonTemplate("request.py.tpl", map[string]any{})
}

func RenderLogPy() (string, error) {
	return RenderPythonTemplate("log.py.tpl", map[string]any{})
}

func RenderExceptionPy() (string, error) {
	return RenderPythonTemplate("exception.py.tpl", map[string]any{})
}

func RenderVersionPy() (string, error) {
	return RenderPythonTemplate("version.py.tpl", map[string]any{})
}

func RenderPyprojectToml() (string, error) {
	return RenderPythonTemplate("pyproject.toml.tpl", map[string]any{})
}

func RenderPackageModule(
	doc *openapi.Document,
	meta PackageMeta,
	bindings []OperationBinding,
) string {
	return RenderPackageModuleWithComments(doc, meta, bindings, config.CommentOverrides{})
}

func RenderPackageModuleWithComments(
	doc *openapi.Document,
	meta PackageMeta,
	bindings []OperationBinding,
	commentOverrides config.CommentOverrides,
) string {
	var buf bytes.Buffer
	hasChildClients := len(meta.ChildClients) > 0
	childClientsForType := []childClient{}
	childClientsForInit := []childClient{}
	childClientsForSync := []childClient{}
	childClientsForAsync := []childClient{}
	if hasChildClients {
		childClientsForType = append([]childClient(nil), meta.ChildClients...)
		childClientsForInit = append([]childClient(nil), meta.ChildClients...)
		childClientsForSync = append([]childClient(nil), meta.ChildClients...)
		childClientsForAsync = append([]childClient(nil), meta.ChildClients...)
	}
	modelDefs := resolvePackageModelDefinitions(doc, meta)
	schemaAliases := packageSchemaAliases(meta)
	for _, model := range modelDefs {
		schemaName := strings.TrimSpace(model.SchemaName)
		modelName := strings.TrimSpace(model.Name)
		if schemaName == "" || modelName == "" {
			continue
		}
		if _, exists := schemaAliases[schemaName]; exists {
			continue
		}
		schemaAliases[schemaName] = modelName
	}
	hasModelClasses := len(modelDefs) > 0 || (meta.Package != nil && len(meta.Package.EmptyModels) > 0)
	overridePaginationClasses := map[string]struct{}{}
	if meta.Package != nil {
		for _, className := range meta.Package.OverridePaginationClasses {
			trimmed := strings.TrimSpace(className)
			if trimmed == "" {
				continue
			}
			overridePaginationClasses[trimmed] = struct{}{}
		}
	}
	hasGeneratedTokenPagedResponse := false
	hasGeneratedNumberPagedResponse := false
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		paginationMode := strings.TrimSpace(binding.Mapping.Pagination)
		if !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) {
			continue
		}
		className := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		if className == "" {
			continue
		}
		if _, overridden := overridePaginationClasses[className]; overridden {
			continue
		}
		if isTokenPagination(paginationMode) {
			hasGeneratedTokenPagedResponse = true
		}
		if isNumberPagination(paginationMode) {
			hasGeneratedNumberPagedResponse = true
		}
	}
	needsCozeModelImport := false
	if meta.Package != nil && len(meta.Package.EmptyModels) > 0 {
		needsCozeModelImport = true
	}
	if !needsCozeModelImport {
		for _, model := range modelDefs {
			if !model.IsEnum {
				needsCozeModelImport = true
				break
			}
		}
	}
	if !needsCozeModelImport {
		needsCozeModelImport = hasGeneratedTokenPagedResponse || hasGeneratedNumberPagedResponse
	}
	hasTokenPagination := packageHasTokenPagination(bindings)
	hasNumberPagination := packageHasNumberPagination(bindings)
	needsTokenPagedResponseImport := false
	needsNumberPagedResponseImport := false
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		mode := strings.TrimSpace(binding.Mapping.Pagination)
		className := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		if className != "" {
			if _, overridden := overridePaginationClasses[className]; overridden {
				continue
			}
		}
		if isTokenPagination(mode) {
			needsTokenPagedResponseImport = true
		}
		if isNumberPagination(mode) {
			needsNumberPagedResponseImport = true
		}
	}
	needAny, needDict := packageNeedsAnyDict(doc, bindings, modelDefs)
	needsListResponseImport := packageNeedsListResponseImport(bindings)
	hasStandardEnumClasses := false
	hasIntEnumClasses := false
	hasDynamicEnumClasses := false
	for _, model := range modelDefs {
		if model.IsEnum {
			enumBase := strings.TrimSpace(model.EnumBase)
			if enumBase == "dynamic_str" {
				hasDynamicEnumClasses = true
			} else if enumBase == "int" {
				hasIntEnumClasses = true
			} else {
				hasStandardEnumClasses = true
			}
		}
	}

	disableAutoImports := meta.Package != nil && meta.Package.DisableAutoImports
	wroteNonCozepyExtraImport := false
	writeExtraImports := func() {
		if meta.Package == nil || len(meta.Package.ExtraImports) == 0 {
			return
		}
		for _, spec := range meta.Package.ExtraImports {
			moduleName := strings.TrimSpace(spec.Module)
			if moduleName == "" {
				continue
			}
			names := make([]string, 0, len(spec.Names))
			for _, name := range spec.Names {
				trimmed := strings.TrimSpace(name)
				if trimmed == "" {
					continue
				}
				names = append(names, trimmed)
			}
			if len(names) == 0 {
				continue
			}
			if !strings.HasPrefix(moduleName, "cozepy.") {
				wroteNonCozepyExtraImport = true
			}
			buf.WriteString(fmt.Sprintf("from %s import %s\n", moduleName, strings.Join(names, ", ")))
		}
	}
	if !disableAutoImports {
		if hasStandardEnumClasses || hasIntEnumClasses {
			buf.WriteString("from enum import Enum\n")
			if hasIntEnumClasses {
				buf.WriteString("from enum import IntEnum\n")
			}
		}
		typingImports := make([]string, 0)
		if hasChildClients {
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
		if hasChildClients || len(modelDefs) > 0 || hasTokenPagination || hasNumberPagination || len(bindings) > 0 {
			typingImports = append(typingImports, "Optional")
		}
		if len(typingImports) > 0 {
			buf.WriteString(fmt.Sprintf("from typing import %s\n\n", strings.Join(typingImports, ", ")))
		}
		writeExtraImports()
		if wroteNonCozepyExtraImport {
			buf.WriteString("\n")
		}
		modelImports := make([]string, 0)
		requestHTTPFromModel := meta.Package != nil && meta.Package.HTTPRequestFromModel
		needHTTPRequest := hasTokenPagination || hasNumberPagination
		if hasTokenPagination {
			modelImports = append(modelImports, "AsyncTokenPaged")
		}
		if hasNumberPagination {
			modelImports = append(modelImports, "AsyncNumberPaged")
		}
		if hasDynamicEnumClasses {
			modelImports = append(modelImports, "DynamicStrEnum")
		}
		if needsCozeModelImport {
			modelImports = append(modelImports, "CozeModel")
		}
		if hasTokenPagination {
			modelImports = append(modelImports, "TokenPaged")
			if needsTokenPagedResponseImport {
				modelImports = append(modelImports, "TokenPagedResponse")
			}
		}
		if hasNumberPagination {
			modelImports = append(modelImports, "NumberPaged")
			if needsNumberPagedResponseImport {
				modelImports = append(modelImports, "NumberPagedResponse")
			}
		}
		if needHTTPRequest && requestHTTPFromModel {
			modelImports = append(modelImports, "HTTPRequest")
		}
		if needsListResponseImport {
			modelImports = append(modelImports, "ListResponse")
		}
		modelImports = OrderedUniqueByPriority(modelImports, []string{
			"AsyncTokenPaged",
			"AsyncNumberPaged",
			"CozeModel",
			"DynamicStrEnum",
			"HTTPRequest",
			"ListResponse",
			"TokenPaged",
			"TokenPagedResponse",
			"NumberPaged",
			"NumberPagedResponse",
		})
		if len(modelImports) > 0 {
			buf.WriteString(fmt.Sprintf("from cozepy.model import %s\n", strings.Join(modelImports, ", ")))
		}
		requestImports := []string{"Requester"}
		if needHTTPRequest && !requestHTTPFromModel {
			requestImports = append([]string{"HTTPRequest"}, requestImports...)
		}
		buf.WriteString(fmt.Sprintf("from cozepy.request import %s\n", strings.Join(requestImports, ", ")))
		utilImports := []string{"remove_url_trailing_slash"}
		needDumpExcludeNone := false
		needRemoveNoneValues := false
		for _, binding := range bindings {
			if binding.Mapping != nil && strings.TrimSpace(binding.Mapping.DelegateTo) != "" {
				continue
			}
			queryBuilder := "dump_exclude_none"
			bodyBuilder := "dump_exclude_none"
			queryBuilderSync := ""
			queryBuilderAsync := ""
			if binding.Mapping != nil {
				queryBuilder = normalizeMapBuilder(binding.Mapping.QueryBuilder)
				if override := strings.TrimSpace(binding.Mapping.QueryBuilderSync); override != "" {
					queryBuilderSync = normalizeMapBuilder(override)
				}
				if override := strings.TrimSpace(binding.Mapping.QueryBuilderAsync); override != "" {
					queryBuilderAsync = normalizeMapBuilder(override)
				}
				bodyBuilder = normalizeMapBuilder(binding.Mapping.BodyBuilder)
			}
			hasQueryFields := len(binding.Details.QueryParameters) > 0
			if binding.Mapping != nil && len(binding.Mapping.QueryFields) > 0 {
				hasQueryFields = true
			}
			hasBodyMap := binding.Mapping != nil && (len(binding.Mapping.BodyFields) > 0 || len(binding.Mapping.BodyFixedValues) > 0)
			if hasQueryFields {
				builders := make([]string, 0, 2)
				if mappingGeneratesSync(binding.Mapping) {
					if queryBuilderSync != "" {
						builders = append(builders, queryBuilderSync)
					} else {
						builders = append(builders, queryBuilder)
					}
				}
				if mappingGeneratesAsync(binding.Mapping) {
					if queryBuilderAsync != "" {
						builders = append(builders, queryBuilderAsync)
					} else {
						builders = append(builders, queryBuilder)
					}
				}
				for _, builder := range builders {
					if builder == "dump_exclude_none" {
						needDumpExcludeNone = true
					}
					if builder == "remove_none_values" {
						needRemoveNoneValues = true
					}
				}
			}
			if hasBodyMap && (mappingGeneratesSync(binding.Mapping) || mappingGeneratesAsync(binding.Mapping)) {
				if bodyBuilder == "dump_exclude_none" {
					needDumpExcludeNone = true
				}
				if bodyBuilder == "remove_none_values" {
					needRemoveNoneValues = true
				}
			}
			if needDumpExcludeNone && needRemoveNoneValues {
				break
			}
		}
		if needDumpExcludeNone {
			utilImports = append([]string{"dump_exclude_none"}, utilImports...)
		}
		if needRemoveNoneValues {
			utilImports = append([]string{"remove_none_values"}, utilImports...)
		}
		utilImports = OrderedUniqueByPriority(utilImports, []string{
			"dump_exclude_none",
			"remove_none_values",
			"remove_url_trailing_slash",
		})
		buf.WriteString(fmt.Sprintf("from cozepy.util import %s\n", strings.Join(utilImports, ", ")))
	} else {
		writeExtraImports()
	}
	if meta.Package != nil && len(meta.Package.RawImports) > 0 {
		for _, rawImport := range meta.Package.RawImports {
			block := strings.TrimRight(rawImport, "\n")
			if strings.TrimSpace(block) == "" {
				buf.WriteString("\n")
				continue
			}
			buf.WriteString(block)
			buf.WriteString("\n")
		}
	}

	imports := collectTypeImports(doc, bindings)
	if len(imports) > 0 {
		buf.WriteString(fmt.Sprintf("from cozepy.types import %s\n", strings.Join(imports, ", ")))
	}

	if hasChildClients {
		buf.WriteString("\nif TYPE_CHECKING:\n")
		for _, child := range childClientsForType {
			typeModule := childImportModule(meta, child.Module)
			if typeModule == "" {
				continue
			}
			first, second := orderedTypeImportNames(child.SyncClass, child.AsyncClass)
			importLine := fmt.Sprintf("    from %s import %s, %s", typeModule, first, second)
			if len(importLine) > 120 {
				buf.WriteString(fmt.Sprintf("    from %s import (\n", typeModule))
				buf.WriteString(fmt.Sprintf("        %s,\n", first))
				buf.WriteString(fmt.Sprintf("        %s,\n", second))
				buf.WriteString("    )\n")
			} else {
				buf.WriteString(importLine + "\n")
			}
		}
	}
	if meta.Package != nil && len(meta.Package.PreModelCode) > 0 {
		buf.WriteString("\n")
	} else {
		buf.WriteString("\n\n")
	}

	if meta.Package != nil && len(meta.Package.PreModelCode) > 0 {
		for _, block := range meta.Package.PreModelCode {
			AppendIndentedCode(&buf, block, 0)
			buf.WriteString("\n")
		}
		if hasModelClasses {
			buf.WriteString("\n")
		}
	}

	if hasModelClasses {
		buf.WriteString(renderPackageModelDefinitions(doc, meta, modelDefs, schemaAliases, commentOverrides))
		buf.WriteString("\n\n\n")
	}
	if hasTokenPagination || hasNumberPagination {
		pagedResponseClasses := renderPagedResponseClasses(bindings, overridePaginationClasses)
		if strings.TrimSpace(pagedResponseClasses) != "" {
			buf.WriteString(pagedResponseClasses)
			buf.WriteString("\n")
		}
	}
	if meta.Package != nil && len(meta.Package.TopLevelCode) > 0 {
		for _, block := range meta.Package.TopLevelCode {
			AppendIndentedCode(&buf, block, 0)
			buf.WriteString("\n")
		}
	}

	syncClass := packageClientClassName(meta, false)
	asyncClass := packageClientClassName(meta, true)
	syncClassKey := "cozepy." + meta.ModulePath + "." + syncClass
	asyncClassKey := "cozepy." + meta.ModulePath + "." + asyncClass
	blankLineBeforeSyncInitCode := meta.Package != nil && meta.Package.BlankLineBeforeSyncInit
	blankLineBeforeAsyncInitCode := meta.Package != nil && meta.Package.BlankLineBeforeAsyncInit

	EnsureTrailingNewlines(&buf, 3)
	buf.WriteString(fmt.Sprintf("class %s(object):\n", syncClass))
	if classDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[syncClassKey]); classDoc != "" {
		style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[syncClassKey])
		WriteClassDocstring(&buf, 1, classDoc, style)
	}
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	if meta.Package != nil && len(meta.Package.SyncInitPreCode) > 0 {
		for _, block := range meta.Package.SyncInitPreCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n")
	if meta.Package != nil && len(meta.Package.SyncInitCode) > 0 {
		if blankLineBeforeSyncInitCode {
			buf.WriteString("\n")
		}
		for _, block := range meta.Package.SyncInitCode {
			AppendIndentedCode(&buf, block, 2)
			buf.WriteString("\n")
		}
	}
	if hasChildClients {
		for _, child := range childClientsForInit {
			attribute := NormalizePythonIdentifier(child.Attribute)
			buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.SyncClass))
		}
		buf.WriteString("\n")
	} else if meta.Package == nil || len(meta.Package.SyncInitCode) == 0 {
		buf.WriteString("\n")
	}
	syncMethodBlocks := make([]ClassMethodBlock, 0)
	if hasChildClients {
		for _, child := range childClientsForSync {
			attribute := NormalizePythonIdentifier(child.Attribute)
			syncMethodBlocks = append(syncMethodBlocks, ClassMethodBlock{
				Name:    attribute,
				Content: renderChildClientProperty(meta, child, false, syncClassKey, commentOverrides),
				IsChild: true,
			})
		}
	}
	for _, binding := range bindings {
		if !mappingGeneratesSync(binding.Mapping) {
			continue
		}
		syncMethodBlocks = append(syncMethodBlocks, ClassMethodBlock{
			Name:    binding.MethodName,
			Content: renderOperationMethodWithContext(doc, binding, false, "cozepy."+meta.ModulePath, syncClass, commentOverrides),
		})
	}
	if meta.Package != nil && len(meta.Package.SyncExtraMethods) > 0 {
		for _, block := range meta.Package.SyncExtraMethods {
			content := IndentCodeBlock(block, 1)
			content = applyMethodDocstringOverrides(content, syncClassKey, commentOverrides)
			syncMethodBlocks = append(syncMethodBlocks, ClassMethodBlock{
				Name:    DetectMethodBlockName(block),
				Content: content,
			})
		}
	}
	syncMethodBlocks = OrderClassMethodBlocks(syncMethodBlocks)
	for _, block := range syncMethodBlocks {
		buf.WriteString(strings.TrimRight(block.Content, "\n"))
		buf.WriteString("\n\n")
	}
	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf("class %s(object):\n", asyncClass))
	if classDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[asyncClassKey]); classDoc != "" {
		style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[asyncClassKey])
		WriteClassDocstring(&buf, 1, classDoc, style)
	}
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	if meta.Package != nil && len(meta.Package.AsyncInitPreCode) > 0 {
		for _, block := range meta.Package.AsyncInitPreCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n")
	if meta.Package != nil && len(meta.Package.AsyncInitCode) > 0 {
		if blankLineBeforeAsyncInitCode {
			buf.WriteString("\n")
		}
		for _, block := range meta.Package.AsyncInitCode {
			AppendIndentedCode(&buf, block, 2)
			buf.WriteString("\n")
		}
	}
	if hasChildClients {
		for _, child := range childClientsForInit {
			attribute := NormalizePythonIdentifier(child.Attribute)
			buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.AsyncClass))
		}
		buf.WriteString("\n")
	} else if meta.Package == nil || len(meta.Package.AsyncInitCode) == 0 {
		buf.WriteString("\n")
	}
	asyncMethodBlocks := make([]ClassMethodBlock, 0)
	if hasChildClients {
		for _, child := range childClientsForAsync {
			attribute := NormalizePythonIdentifier(child.Attribute)
			asyncMethodBlocks = append(asyncMethodBlocks, ClassMethodBlock{
				Name:    attribute,
				Content: renderChildClientProperty(meta, child, true, asyncClassKey, commentOverrides),
				IsChild: true,
			})
		}
	}
	for _, binding := range bindings {
		if !mappingGeneratesAsync(binding.Mapping) {
			continue
		}
		asyncMethodBlocks = append(asyncMethodBlocks, ClassMethodBlock{
			Name:    binding.MethodName,
			Content: renderOperationMethodWithContext(doc, binding, true, "cozepy."+meta.ModulePath, asyncClass, commentOverrides),
		})
	}
	if meta.Package != nil && len(meta.Package.AsyncExtraMethods) > 0 {
		for _, block := range meta.Package.AsyncExtraMethods {
			content := IndentCodeBlock(block, 1)
			content = applyMethodDocstringOverrides(content, asyncClassKey, commentOverrides)
			asyncMethodBlocks = append(asyncMethodBlocks, ClassMethodBlock{
				Name:    DetectMethodBlockName(block),
				Content: content,
			})
		}
	}
	asyncMethodBlocks = OrderClassMethodBlocks(asyncMethodBlocks)
	for _, block := range asyncMethodBlocks {
		buf.WriteString(strings.TrimRight(block.Content, "\n"))
		buf.WriteString("\n\n")
	}

	content := strings.TrimRight(buf.String(), "\n") + "\n"
	if !strings.Contains(content, "(str, Enum):") && !strings.Contains(content, "(int, Enum):") {
		content = strings.Replace(content, "from enum import Enum\n", "", 1)
	}
	if !strings.Contains(content, "(IntEnum):") {
		content = strings.Replace(content, "from enum import IntEnum\n", "", 1)
	}
	content = normalizeAutoInferredImports(content)
	return normalizeTypingOptionalImport(content)
}

func collectTypeImports(doc *openapi.Document, bindings []OperationBinding) []string {
	_ = doc
	_ = bindings
	return nil
}

func isTokenPagination(mode string) bool {
	return strings.TrimSpace(mode) == "token"
}

func isNumberPagination(mode string) bool {
	mode = strings.TrimSpace(mode)
	return mode == "number" || mode == "number_has_more"
}

func normalizeMapBuilder(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return "dump_exclude_none"
	}
	switch v {
	case "dump_exclude_none", "remove_none_values", "raw":
		return v
	default:
		return "dump_exclude_none"
	}
}

func formatTypeImportPair(syncClass, asyncClass string) string {
	first, second := orderedTypeImportNames(syncClass, asyncClass)
	if second == "" {
		return first
	}
	return first + ", " + second
}

func orderedTypeImportNames(syncClass, asyncClass string) (string, string) {
	syncClass = strings.TrimSpace(syncClass)
	asyncClass = strings.TrimSpace(asyncClass)
	if syncClass == "" {
		return asyncClass, ""
	}
	if asyncClass == "" {
		return syncClass, ""
	}
	if strings.Compare(asyncClass, syncClass) < 0 {
		return asyncClass, syncClass
	}
	return syncClass, asyncClass
}

func packageHasTokenPagination(bindings []OperationBinding) bool {
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		if isTokenPagination(binding.Mapping.Pagination) {
			return true
		}
	}
	return false
}

func packageHasNumberPagination(bindings []OperationBinding) bool {
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		if isNumberPagination(binding.Mapping.Pagination) {
			return true
		}
	}
	return false
}

func packageNeedsAnyDict(doc *openapi.Document, bindings []OperationBinding, modelDefs []packageModelDefinition) (bool, bool) {
	needAny := false
	needDict := false

	for _, binding := range bindings {
		mapping := binding.Mapping
		if mapping == nil || strings.TrimSpace(mapping.ResponseType) == "" {
			paginationMode := ""
			if mapping != nil {
				paginationMode = strings.TrimSpace(mapping.Pagination)
			}
			if !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) {
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
			for _, typeName := range mapping.ArgTypes {
				typeName = strings.TrimSpace(typeName)
				if strings.Contains(typeName, "Any") {
					needAny = true
				}
				if strings.Contains(typeName, "Dict") {
					needDict = true
				}
			}
		}
	}

	for _, model := range modelDefs {
		for _, fieldType := range model.FieldTypes {
			typeName := strings.TrimSpace(fieldType)
			if strings.Contains(typeName, "Any") {
				needAny = true
			}
			if strings.Contains(typeName, "Dict") {
				needDict = true
			}
		}
		if model.IsEnum || model.Schema == nil {
			for _, extraField := range model.ExtraFields {
				fieldType := strings.TrimSpace(extraField.Type)
				if strings.Contains(fieldType, "Any") {
					needAny = true
				}
				if strings.Contains(fieldType, "Dict") {
					needDict = true
				}
			}
			continue
		}
		for propertyName, propertySchema := range model.Schema.Properties {
			required := false
			for _, requiredName := range model.Schema.Required {
				if requiredName == propertyName {
					required = true
					break
				}
			}
			fieldType := PythonTypeForSchemaWithAliases(doc, propertySchema, required, nil)
			if strings.Contains(fieldType, "Any") {
				needAny = true
			}
			if strings.Contains(fieldType, "Dict") {
				needDict = true
			}
		}
		for _, extraField := range model.ExtraFields {
			fieldType := strings.TrimSpace(extraField.Type)
			if strings.Contains(fieldType, "Any") {
				needAny = true
			}
			if strings.Contains(fieldType, "Dict") {
				needDict = true
			}
		}
	}

	return needAny, needDict
}

func packageNeedsListResponseImport(bindings []OperationBinding) bool {
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		candidates := []string{
			binding.Mapping.ResponseType,
			binding.Mapping.AsyncResponseType,
			binding.Mapping.ResponseCast,
		}
		for _, candidate := range candidates {
			if strings.Contains(strings.TrimSpace(candidate), "ListResponse[") {
				return true
			}
		}
	}
	return false
}

func renderPagedResponseClasses(bindings []OperationBinding, overriddenClasses map[string]struct{}) string {
	seen := map[string]struct{}{}
	ordered := make([]OperationBinding, 0, len(bindings))
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		paginationMode := strings.TrimSpace(binding.Mapping.Pagination)
		if !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) {
			continue
		}
		className := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
		if className == "" || itemType == "" {
			continue
		}
		if _, overridden := overriddenClasses[className]; overridden {
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

		if paginationMode == "number_has_more" {
			hasMoreField := strings.TrimSpace(binding.Mapping.PaginationHasMoreField)
			if hasMoreField == "" {
				hasMoreField = "has_more"
			}
			buf.WriteString(fmt.Sprintf("class %s(CozeModel, NumberPagedResponse[%s]):\n", className, itemType))
			buf.WriteString(fmt.Sprintf("    %s: bool\n", hasMoreField))
			buf.WriteString(fmt.Sprintf("    %s: List[%s]\n\n", itemsField, itemType))
			buf.WriteString("    def get_total(self) -> Optional[int]:\n")
			buf.WriteString("        return None\n\n")
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

func normalizeAutoInferredImports(content string) string {
	sanitized := stripPythonStringsAndComments(content)
	ensureNamed := func(module string, symbol string, includeQuotedAnnotation bool) {
		if symbol == "" || module == "" {
			return
		}
		used := containsIdentifierOutsideImportSection(sanitized, symbol)
		if !used && includeQuotedAnnotation {
			used = containsQuotedAnnotationOutsideImportSection(content, symbol)
		}
		if !used {
			return
		}
		if declaresIdentifier(sanitized, symbol) {
			return
		}
		content = ensureNamedImport(content, module, symbol)
	}
	ensureModule := func(module string) {
		if module == "" {
			return
		}
		if !containsModuleReferenceOutsideImportSection(sanitized, module) {
			return
		}
		if declaresIdentifier(sanitized, module) {
			return
		}
		content = ensureModuleImport(content, module)
	}

	for _, module := range []string{"base64", "httpx", "json", "os", "time", "warnings"} {
		ensureModule(module)
	}

	knownFromImports := []struct {
		Module                  string
		Symbols                 []string
		IncludeQuotedAnnotation bool
	}{
		{Module: "typing", Symbols: []string{"Any", "AsyncIterator", "Dict", "IO", "List", "Optional", "TYPE_CHECKING", "Tuple", "Union", "overload"}},
		{Module: "typing_extensions", Symbols: []string{"Literal"}},
		{Module: "pathlib", Symbols: []string{"Path"}},
		{Module: "pydantic", Symbols: []string{"Field", "field_validator"}},
		{Module: "cozepy", Symbols: []string{"AudioFormat"}},
		{Module: "cozepy.audio.voiceprint_groups.features", Symbols: []string{"UserInfo"}},
		{Module: "cozepy.bots", Symbols: []string{"PublishStatus"}},
		{Module: "cozepy.chat", Symbols: []string{"ChatEvent", "ChatUsage", "Message", "MessageContentType", "MessageRole", "_chat_stream_handler"}},
		{Module: "cozepy.datasets.documents", Symbols: []string{"Document", "DocumentBase", "DocumentChunkStrategy", "DocumentUpdateRule"}},
		{Module: "cozepy.exception", Symbols: []string{"CozeAPIError"}},
		{Module: "cozepy.files", Symbols: []string{"FileTypes", "_try_fix_file"}},
		{Module: "cozepy.model", Symbols: []string{"AsyncIteratorHTTPResponse", "AsyncLastIDPaged", "AsyncNumberPaged", "AsyncStream", "AsyncTokenPaged", "CozeModel", "DynamicStrEnum", "FileHTTPResponse", "IteratorHTTPResponse", "LastIDPaged", "LastIDPagedResponse", "ListResponse", "NumberPaged", "NumberPagedResponse", "Stream", "TokenPaged", "TokenPagedResponse"}},
		{Module: "cozepy.request", Symbols: []string{"Requester"}},
		{Module: "cozepy.util", Symbols: []string{"base64_encode_string", "dump_exclude_none", "remove_none_values", "remove_url_trailing_slash"}},
		{Module: "cozepy.workspaces", Symbols: []string{"WorkspaceRoleType"}},
		{Module: ".versions", Symbols: []string{"WorkflowUserInfo"}},
	}
	for _, entry := range knownFromImports {
		for _, symbol := range entry.Symbols {
			ensureNamed(entry.Module, symbol, entry.IncludeQuotedAnnotation)
		}
	}
	feedbackSymbols := make([]string, 0, 2)
	if containsQuotedAnnotationOutsideImportSection(content, "AsyncMessagesFeedbackClient") {
		feedbackSymbols = append(feedbackSymbols, "AsyncMessagesFeedbackClient")
	}
	if containsQuotedAnnotationOutsideImportSection(content, "ConversationsMessagesFeedbackClient") {
		feedbackSymbols = append(feedbackSymbols, "ConversationsMessagesFeedbackClient")
	}
	if len(feedbackSymbols) > 0 {
		content = ensureTypeCheckingNamedImport(content, ".feedback", feedbackSymbols)
	}
	if containsIdentifierOutsideImportSection(sanitized, "HTTPRequest") &&
		!declaresIdentifier(sanitized, "HTTPRequest") &&
		!isSymbolImportedFromModule(content, "cozepy.model", "HTTPRequest") &&
		!isSymbolImportedFromModule(content, "cozepy.request", "HTTPRequest") {
		content = ensureNamedImport(content, "cozepy.model", "HTTPRequest")
	}
	return content
}

func containsIdentifierOutsideImportSection(content string, ident string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if lineHasIdentifier(line, ident) {
			return true
		}
	}
	return false
}

func containsQuotedAnnotationOutsideImportSection(content string, ident string) bool {
	if ident == "" {
		return false
	}
	singleQuoted := "'" + ident + "'"
	doubleQuoted := `"` + ident + `"`
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if !strings.Contains(line, ":") && !strings.Contains(line, "->") {
			continue
		}
		if strings.Contains(line, singleQuoted) || strings.Contains(line, doubleQuoted) {
			return true
		}
	}
	return false
}

func containsModuleReferenceOutsideImportSection(content string, module string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			continue
		}
		if lineHasModuleReference(line, module) {
			return true
		}
	}
	return false
}

func lineHasModuleReference(line string, module string) bool {
	if module == "" {
		return false
	}
	for start := 0; start < len(line); {
		idx := strings.Index(line[start:], module)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isIdentifierRune(line[idx-1])
		afterIdx := idx + len(module)
		if beforeOK && afterIdx < len(line) && line[afterIdx] == '.' {
			return true
		}
		start = idx + len(module)
	}
	return false
}

func stripPythonStringsAndComments(content string) string {
	const (
		scanCode = iota
		scanSingleQuote
		scanDoubleQuote
		scanTripleSingleQuote
		scanTripleDoubleQuote
		scanComment
	)

	state := scanCode
	escaped := false
	var b strings.Builder
	b.Grow(len(content))

	for i := 0; i < len(content); i++ {
		ch := content[i]

		switch state {
		case scanComment:
			if ch == '\n' {
				state = scanCode
				b.WriteByte('\n')
			} else {
				b.WriteByte(' ')
			}
		case scanSingleQuote:
			if ch == '\n' {
				state = scanCode
				escaped = false
				b.WriteByte('\n')
				continue
			}
			if escaped {
				escaped = false
				b.WriteByte(' ')
				continue
			}
			if ch == '\\' {
				escaped = true
				b.WriteByte(' ')
				continue
			}
			if ch == '\'' {
				state = scanCode
			}
			b.WriteByte(' ')
		case scanDoubleQuote:
			if ch == '\n' {
				state = scanCode
				escaped = false
				b.WriteByte('\n')
				continue
			}
			if escaped {
				escaped = false
				b.WriteByte(' ')
				continue
			}
			if ch == '\\' {
				escaped = true
				b.WriteByte(' ')
				continue
			}
			if ch == '"' {
				state = scanCode
			}
			b.WriteByte(' ')
		case scanTripleSingleQuote:
			if ch == '\n' {
				b.WriteByte('\n')
				continue
			}
			if ch == '\'' && i+2 < len(content) && content[i+1] == '\'' && content[i+2] == '\'' {
				b.WriteString("   ")
				i += 2
				state = scanCode
				continue
			}
			b.WriteByte(' ')
		case scanTripleDoubleQuote:
			if ch == '\n' {
				b.WriteByte('\n')
				continue
			}
			if ch == '"' && i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
				b.WriteString("   ")
				i += 2
				state = scanCode
				continue
			}
			b.WriteByte(' ')
		default:
			if ch == '#' {
				state = scanComment
				b.WriteByte(' ')
				continue
			}
			if ch == '\'' {
				if i+2 < len(content) && content[i+1] == '\'' && content[i+2] == '\'' {
					state = scanTripleSingleQuote
					b.WriteString("   ")
					i += 2
					continue
				}
				state = scanSingleQuote
				b.WriteByte(' ')
				continue
			}
			if ch == '"' {
				if i+2 < len(content) && content[i+1] == '"' && content[i+2] == '"' {
					state = scanTripleDoubleQuote
					b.WriteString("   ")
					i += 2
					continue
				}
				state = scanDoubleQuote
				b.WriteByte(' ')
				continue
			}
			b.WriteByte(ch)
		}
	}

	return b.String()
}

func isSymbolImportedFromModule(content string, module string, symbol string) bool {
	if module == "" || symbol == "" {
		return false
	}
	lines := strings.Split(content, "\n")
	importPrefix := "from " + module + " import "
	insertIdx := topImportInsertIndex(lines)
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == symbol {
				return true
			}
		}
	}
	return false
}

func ensureTypeCheckingNamedImport(content string, module string, symbols []string) string {
	if module == "" || len(symbols) == 0 {
		return content
	}
	ordered := make([]string, 0, len(symbols))
	seen := map[string]struct{}{}
	for _, symbol := range symbols {
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		ordered = append(ordered, symbol)
	}
	if len(ordered) == 0 {
		return content
	}

	content = ensureNamedImport(content, "typing", "TYPE_CHECKING")
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)

	blockIdx := -1
	for i := insertIdx; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line == trimmed && trimmed == "if TYPE_CHECKING:" {
			blockIdx = i
			break
		}
		if line == trimmed && trimmed != "" {
			break
		}
	}
	if blockIdx < 0 {
		importLine := fmt.Sprintf("    from %s import %s", module, strings.Join(ordered, ", "))
		prefix := append([]string{}, lines[:insertIdx]...)
		suffix := append([]string{}, lines[insertIdx:]...)
		if len(prefix) > 0 && strings.TrimSpace(prefix[len(prefix)-1]) != "" {
			prefix = append(prefix, "")
		}
		prefix = append(prefix, "if TYPE_CHECKING:", importLine, "")
		return strings.Join(append(prefix, suffix...), "\n")
	}

	blockEnd := blockIdx + 1
	for blockEnd < len(lines) {
		line := lines[blockEnd]
		trimmed := strings.TrimSpace(line)
		if line == trimmed && trimmed != "" {
			break
		}
		blockEnd++
	}

	importPrefix := "from " + module + " import "
	importLineIdx := -1
	for i := blockIdx + 1; i < blockEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		importLineIdx = i
		break
	}

	if importLineIdx >= 0 {
		trimmed := strings.TrimSpace(lines[importLineIdx])
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		updated := make([]string, 0, len(ordered)+4)
		localSeen := map[string]struct{}{}
		for _, name := range strings.Split(namesPart, ",") {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, exists := localSeen[name]; exists {
				continue
			}
			localSeen[name] = struct{}{}
			updated = append(updated, name)
		}
		for _, symbol := range ordered {
			if _, exists := localSeen[symbol]; exists {
				continue
			}
			localSeen[symbol] = struct{}{}
			updated = append(updated, symbol)
		}
		lines[importLineIdx] = "    " + importPrefix + strings.Join(updated, ", ")
		return strings.Join(lines, "\n")
	}

	importLine := "    " + importPrefix + strings.Join(ordered, ", ")
	prefix := append([]string{}, lines[:blockEnd]...)
	suffix := append([]string{}, lines[blockEnd:]...)
	prefix = append(prefix, importLine)
	return strings.Join(append(prefix, suffix...), "\n")
}

func lineHasIdentifier(line string, ident string) bool {
	if ident == "" {
		return false
	}
	for start := 0; start < len(line); {
		idx := strings.Index(line[start:], ident)
		if idx < 0 {
			return false
		}
		idx += start
		beforeOK := idx == 0 || !isIdentifierRune(line[idx-1])
		afterIdx := idx + len(ident)
		afterOK := afterIdx >= len(line) || !isIdentifierRune(line[afterIdx])
		if beforeOK && afterOK {
			return true
		}
		start = idx + len(ident)
	}
	return false
}

func isIdentifierRune(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func declaresIdentifier(content string, ident string) bool {
	if ident == "" {
		return false
	}
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "class "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, "def "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, "async def "+ident+"(") {
			return true
		}
		if strings.HasPrefix(trimmed, ident+" =") || strings.HasPrefix(trimmed, ident+":") {
			return true
		}
	}
	return false
}

func ensureNamedImport(content string, module string, symbol string) string {
	if module == "" || symbol == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)
	importPrefix := "from " + module + " import "
	moduleImportIndex := -1
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, importPrefix) {
			continue
		}
		moduleImportIndex = i
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == symbol {
				return content
			}
		}
	}

	if moduleImportIndex >= 0 {
		trimmed := strings.TrimSpace(lines[moduleImportIndex])
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, importPrefix))
		names := strings.Split(namesPart, ",")
		seen := map[string]struct{}{}
		updated := make([]string, 0, len(names)+1)
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, exists := seen[name]; exists {
				continue
			}
			seen[name] = struct{}{}
			updated = append(updated, name)
		}
		if _, exists := seen[symbol]; !exists {
			updated = append(updated, symbol)
		}
		formatted := formatNamedImportLines(module, updated, "")
		if len(formatted) == 1 {
			lines[moduleImportIndex] = formatted[0]
			return strings.Join(lines, "\n")
		}
		prefix := append([]string{}, lines[:moduleImportIndex]...)
		suffix := append([]string{}, lines[moduleImportIndex+1:]...)
		lines = append(prefix, append(formatted, suffix...)...)
		return strings.Join(lines, "\n")
	}

	line := importPrefix + symbol
	prefix := append([]string{}, lines[:insertIdx]...)
	suffix := append([]string{}, lines[insertIdx:]...)
	formatted := formatNamedImportLines(module, []string{symbol}, "")
	if len(formatted) == 0 {
		formatted = []string{line}
	}
	prefix = append(prefix, formatted...)
	return strings.Join(append(prefix, suffix...), "\n")
}

func ensureModuleImport(content string, module string) string {
	if module == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	insertIdx := topImportInsertIndex(lines)
	for i := 0; i < insertIdx; i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if line != trimmed {
			continue
		}
		if !strings.HasPrefix(trimmed, "import ") {
			continue
		}
		items := strings.Split(strings.TrimSpace(strings.TrimPrefix(trimmed, "import ")), ",")
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == module || strings.HasPrefix(item, module+" as ") {
				return content
			}
		}
	}
	prefix := append([]string{}, lines[:insertIdx]...)
	suffix := append([]string{}, lines[insertIdx:]...)
	prefix = append(prefix, "import "+module)
	return strings.Join(append(prefix, suffix...), "\n")
}

func shouldWrapNamedImport(module string, names []string) bool {
	if module == "cozepy.chat" {
		for _, name := range names {
			if strings.TrimSpace(name) == "_chat_stream_handler" {
				return true
			}
		}
	}
	if module == "cozepy.datasets.documents" && len(names) >= 4 {
		return true
	}
	return false
}

func formatNamedImportLines(module string, names []string, indent string) []string {
	trimmed := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		trimmed = append(trimmed, name)
	}
	if len(trimmed) == 0 {
		return nil
	}
	if !shouldWrapNamedImport(module, trimmed) {
		return []string{fmt.Sprintf("%sfrom %s import %s", indent, module, strings.Join(trimmed, ", "))}
	}
	lines := make([]string, 0, len(trimmed)+2)
	lines = append(lines, fmt.Sprintf("%sfrom %s import (", indent, module))
	for _, name := range trimmed {
		lines = append(lines, fmt.Sprintf("%s    %s,", indent, name))
	}
	lines = append(lines, fmt.Sprintf("%s)", indent))
	return lines
}

func topImportInsertIndex(lines []string) int {
	insertIdx := 0
	seenImport := false
	parenDepth := 0
	for insertIdx < len(lines) {
		trimmed := strings.TrimSpace(lines[insertIdx])
		if !seenImport {
			if trimmed == "" {
				insertIdx++
				continue
			}
			if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
				seenImport = true
				parenDepth += strings.Count(trimmed, "(")
				parenDepth -= strings.Count(trimmed, ")")
				insertIdx++
				continue
			}
			return insertIdx
		}

		if parenDepth > 0 {
			parenDepth += strings.Count(trimmed, "(")
			parenDepth -= strings.Count(trimmed, ")")
			insertIdx++
			continue
		}
		if trimmed == "" {
			insertIdx++
			continue
		}
		if strings.HasPrefix(trimmed, "from ") || strings.HasPrefix(trimmed, "import ") {
			parenDepth += strings.Count(trimmed, "(")
			parenDepth -= strings.Count(trimmed, ")")
			insertIdx++
			continue
		}
		return insertIdx
	}
	return insertIdx
}

func normalizeTypingOptionalImport(content string) string {
	lines := strings.Split(content, "\n")
	usesOptional := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "from typing import ") {
			continue
		}
		if strings.Contains(line, "Optional[") {
			usesOptional = true
			break
		}
	}

	isTypingNameUsed := func(name string) bool {
		switch name {
		case "Optional":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "Optional[") {
					return true
				}
			}
			return false
		case "List":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "List[") {
					return true
				}
			}
			return false
		case "Dict":
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "from typing import ") {
					continue
				}
				if strings.Contains(line, "Dict[") {
					return true
				}
			}
			return false
		case "TYPE_CHECKING":
			return containsIdentifierOutsideImportSection(content, "TYPE_CHECKING")
		default:
			return containsIdentifierOutsideImportSection(content, name)
		}
	}

	typingImportIndexes := make([]int, 0)
	optionalImported := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			continue
		}
		typingImportIndexes = append(typingImportIndexes, i)
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, "from typing import "))
		if namesPart == "" {
			continue
		}
		for _, name := range strings.Split(namesPart, ",") {
			if strings.TrimSpace(name) == "Optional" {
				optionalImported = true
				break
			}
		}
		if optionalImported {
			break
		}
	}

	normalizeTypingLine := func(line string, includeOptional bool) (string, bool) {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			return line, true
		}
		namesPart := strings.TrimSpace(strings.TrimPrefix(trimmed, "from typing import "))
		names := strings.Split(namesPart, ",")
		seen := map[string]struct{}{}
		kept := make([]string, 0, len(names)+1)
		for _, name := range names {
			typedName := strings.TrimSpace(name)
			if typedName == "" {
				continue
			}
			if typedName == "Optional" && !includeOptional {
				continue
			}
			if typedName != "Optional" && !isTypingNameUsed(typedName) {
				continue
			}
			if _, exists := seen[typedName]; exists {
				continue
			}
			seen[typedName] = struct{}{}
			kept = append(kept, typedName)
		}
		if includeOptional {
			if _, exists := seen["Optional"]; !exists {
				kept = append(kept, "Optional")
			}
		}
		if len(kept) == 0 {
			return "", false
		}
		return "from typing import " + strings.Join(kept, ", "), true
	}

	includeOptional := usesOptional
	pruned := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "from typing import ") {
			pruned = append(pruned, line)
			continue
		}
		updated, keep := normalizeTypingLine(line, includeOptional)
		if !keep {
			continue
		}
		pruned = append(pruned, updated)
	}
	if includeOptional && !optionalImported {
		if len(typingImportIndexes) > 0 {
			idx := typingImportIndexes[0]
			lines = pruned
			if idx < len(lines) {
				updated, keep := normalizeTypingLine(lines[idx], true)
				if keep {
					lines[idx] = updated
					return strings.Join(lines, "\n")
				}
			}
		}
		insertIdx := 0
		for insertIdx < len(pruned) {
			trimmed := strings.TrimSpace(pruned[insertIdx])
			if trimmed == "" {
				insertIdx++
				continue
			}
			if strings.HasPrefix(trimmed, "import ") || strings.HasPrefix(trimmed, "from ") {
				insertIdx++
				continue
			}
			break
		}
		prefix := append([]string{}, pruned[:insertIdx]...)
		suffix := append([]string{}, pruned[insertIdx:]...)
		prefix = append(prefix, "from typing import Optional")
		if insertIdx < len(pruned) && strings.TrimSpace(pruned[insertIdx]) != "" {
			prefix = append(prefix, "")
		}
		return strings.Join(append(prefix, suffix...), "\n")
	}
	return strings.Join(pruned, "\n")
}

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
	SeparateCommentedEnum *bool
}

func packageSchemaAliases(meta PackageMeta) map[string]string {
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

func resolvePackageModelDefinitions(doc *openapi.Document, meta PackageMeta) []packageModelDefinition {
	if meta.Package == nil || doc == nil {
		return nil
	}
	result := make([]packageModelDefinition, 0, len(meta.Package.ModelSchemas))
	includedSchemaNames := map[string]struct{}{}
	for _, model := range meta.Package.ModelSchemas {
		definition, ok := resolveConfiguredModelDefinition(doc, model)
		if !ok {
			continue
		}
		result = append(result, definition)
		schemaName := strings.TrimSpace(definition.SchemaName)
		if schemaName != "" {
			includedSchemaNames[schemaName] = struct{}{}
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
			result = append(result, packageModelDefinition{
				SchemaName:    schemaName,
				Name:          NormalizeClassName(schemaName),
				Schema:        resolved,
				IsEnum:        isSchemaEnum(resolved, nil),
				FieldTypes:    map[string]string{},
				FieldDefaults: map[string]string{},
			})
		}
	}
	return orderModelDefinitionsByDependencies(doc, result)
}

func resolveConfiguredModelDefinition(doc *openapi.Document, model config.ModelSchema) (packageModelDefinition, bool) {
	schemaName := strings.TrimSpace(model.Schema)
	modelName := strings.TrimSpace(model.Name)
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
		SeparateCommentedEnum: model.SeparateCommentedEnum,
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
	separateCommentedEnum := meta.Package != nil && meta.Package.SeparateCommentedEnum

	for _, model := range models {
		classKey := modulePrefix + "." + model.Name
		if len(model.BeforeCode) > 0 {
			for _, block := range model.BeforeCode {
				AppendIndentedCode(&buf, block, 0)
				buf.WriteString("\n")
			}
		}
		modelSeparateCommentedEnum := separateCommentedEnum
		if model.SeparateCommentedEnum != nil {
			modelSeparateCommentedEnum = *model.SeparateCommentedEnum
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
			for i, enumValue := range enumItems {
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
				if modelSeparateCommentedEnum && i < len(enumItems)-1 {
					nextHasComment := false
					nextName := strings.TrimSpace(enumItems[i+1].Name)
					if nextName == "" {
						nextName = EnumMemberName(fmt.Sprintf("%v", enumItems[i+1].Value))
					}
					nextInlineComment := strings.TrimSpace(commentOverrides.InlineEnumMemberComment[classKey+"."+nextName])
					nextComment := LinesFromCommentOverride(commentOverrides.EnumMemberComments[classKey+"."+nextName])
					if nextInlineComment != "" || len(nextComment) > 0 {
						nextHasComment = true
					}
					if len(enumComment) > 0 || inlineEnumComment != "" || nextHasComment {
						buf.WriteString("\n")
					}
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
	if strings.HasPrefix(text, "{") && strings.Contains(text, "\"insert\"") && strings.Contains(text, "\"ops\"") {
		if extracted := extractSwaggerRichText(text); extracted != "" {
			return extracted
		}
		return ""
	}
	return text
}

func extractSwaggerRichText(raw string) string {
	var node interface{}
	if err := json.Unmarshal([]byte(raw), &node); err != nil {
		return ""
	}
	fragments := make([]string, 0, 32)
	collectRichTextInserts(node, &fragments)
	parts := make([]string, 0, len(fragments))
	for _, fragment := range fragments {
		line := strings.TrimSpace(fragment)
		if line == "" || line == "*" {
			continue
		}
		parts = append(parts, line)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func collectRichTextInserts(node interface{}, fragments *[]string) {
	switch typed := node.(type) {
	case map[string]interface{}:
		for key, value := range typed {
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

type ClassMethodBlock struct {
	Name    string
	Content string
	IsChild bool
}

func mappingGeneratesSync(mapping *config.OperationMapping) bool {
	if mapping == nil {
		return true
	}
	return !mapping.AsyncOnly
}

func mappingGeneratesAsync(mapping *config.OperationMapping) bool {
	if mapping == nil {
		return true
	}
	return !mapping.SyncOnly
}

func applyMethodDocstringOverrides(block string, classKey string, commentOverrides config.CommentOverrides) string {
	trimmedBlock := strings.TrimRight(block, "\n")
	if strings.TrimSpace(trimmedBlock) == "" {
		return block
	}
	lines := strings.Split(trimmedBlock, "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		name, ok := ParseDefName(strings.TrimSpace(line))
		if !ok {
			out = append(out, line)
			continue
		}

		signatureEnd := i
		for signatureEnd+1 < len(lines) && !strings.HasSuffix(strings.TrimSpace(lines[signatureEnd]), ":") {
			signatureEnd++
		}
		for j := i; j <= signatureEnd && j < len(lines); j++ {
			out = append(out, lines[j])
		}

		methodKey := strings.TrimSpace(classKey) + "." + name
		rawDoc, exists := commentOverrides.MethodDocstrings[methodKey]
		if exists {
			docstring := strings.TrimSpace(rawDoc)
			if docstring != "" {
				nextNonEmpty := signatureEnd + 1
				for nextNonEmpty < len(lines) && strings.TrimSpace(lines[nextNonEmpty]) == "" {
					nextNonEmpty++
				}
				if nextNonEmpty >= len(lines) || !IsDocstringLine(strings.TrimSpace(lines[nextNonEmpty])) {
					indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
					docLines := RenderMethodDocstringLines(
						docstring,
						strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey]),
						indent+"    ",
					)
					out = append(out, docLines...)
				}
			}
		}
		i = signatureEnd
	}
	return strings.Join(out, "\n")
}

func OrderClassMethodBlocks(blocks []ClassMethodBlock) []ClassMethodBlock {
	if len(blocks) == 0 {
		return blocks
	}

	prioritizedMethodNames := []string{"stream", "create", "clone", "retrieve", "update", "delete", "list"}
	prioritizedBuckets := make(map[string][]ClassMethodBlock, len(prioritizedMethodNames))
	prioritizedSet := make(map[string]struct{}, len(prioritizedMethodNames))
	for _, name := range prioritizedMethodNames {
		prioritizedSet[name] = struct{}{}
	}

	childMethods := make([]ClassMethodBlock, 0)
	otherMethods := make([]ClassMethodBlock, 0)
	for _, block := range blocks {
		if block.IsChild {
			childMethods = append(childMethods, block)
			continue
		}
		name := strings.TrimSpace(block.Name)
		if _, ok := prioritizedSet[name]; ok {
			prioritizedBuckets[name] = append(prioritizedBuckets[name], block)
			continue
		}
		otherMethods = append(otherMethods, block)
	}

	ordered := make([]ClassMethodBlock, 0, len(blocks))
	ordered = append(ordered, childMethods...)
	for _, name := range prioritizedMethodNames {
		ordered = append(ordered, prioritizedBuckets[name]...)
	}
	ordered = append(ordered, otherMethods...)
	return ordered
}

func renderChildClientProperty(
	meta PackageMeta,
	child childClient,
	async bool,
	classKey string,
	commentOverrides config.CommentOverrides,
) string {
	attribute := NormalizePythonIdentifier(child.Attribute)
	typeName := child.SyncClass
	if async {
		typeName = child.AsyncClass
	}
	module := strings.TrimSpace(child.Module)
	constructExpr := fmt.Sprintf("%s(base_url=self._base_url, requester=self._requester)", typeName)

	var buf bytes.Buffer
	buf.WriteString("    @property\n")
	buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", attribute, typeName))
	methodKey := strings.TrimSpace(classKey) + "." + attribute
	if docstring, ok := commentOverrides.MethodDocstrings[methodKey]; ok {
		docstring = strings.TrimSpace(docstring)
		if docstring != "" {
			style := strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey])
			WriteMethodDocstring(&buf, 2, docstring, style)
		}
	}
	buf.WriteString(fmt.Sprintf("        if not self._%s:\n", attribute))

	if module == "" {
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else if strings.HasPrefix(module, ".") {
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", module, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	} else {
		absModule := childImportModule(meta, module)
		buf.WriteString(fmt.Sprintf("            from %s import %s\n\n", absModule, typeName))
		buf.WriteString(fmt.Sprintf("            self._%s = %s\n", attribute, constructExpr))
	}
	buf.WriteString(fmt.Sprintf("        return self._%s\n", attribute))
	return buf.String()
}

func buildRenderQueryFields(
	doc *openapi.Document,
	details openapi.OperationDetails,
	mapping *config.OperationMapping,
	paramAliases map[string]string,
	argTypes map[string]string,
) []RenderQueryField {
	fields := make([]RenderQueryField, 0)
	if mapping != nil && len(mapping.QueryFields) > 0 {
		for _, field := range mapping.QueryFields {
			rawName := strings.TrimSpace(field.Name)
			if rawName == "" {
				continue
			}
			argName := OperationArgName(rawName, paramAliases)
			valueExpr := argName
			if field.UseValue {
				if field.Required {
					valueExpr = fmt.Sprintf("%s.value", argName)
				} else {
					valueExpr = fmt.Sprintf("%s.value if %s else None", argName, argName)
				}
			}
			if mapping != nil && len(mapping.QueryFieldValues) > 0 {
				if override, ok := mapping.QueryFieldValues[rawName]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
			}
			typeName := strings.TrimSpace(field.Type)
			if typeName == "" {
				typeName = "Any"
			}
			if !field.Required && !strings.HasPrefix(typeName, "Optional[") {
				typeName = "Optional[" + typeName + "]"
			}
			fields = append(fields, RenderQueryField{
				RawName:      rawName,
				ArgName:      argName,
				ValueExpr:    valueExpr,
				TypeName:     typeName,
				Required:     field.Required,
				DefaultValue: strings.TrimSpace(field.Default),
			})
		}
		return fields
	}

	for _, param := range details.QueryParameters {
		argName := OperationArgName(param.Name, paramAliases)
		typeName := TypeOverride(param.Name, param.Required, PythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		valueExpr := argName
		if mapping != nil && len(mapping.QueryFieldValues) > 0 {
			if override, ok := mapping.QueryFieldValues[param.Name]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
		}
		fields = append(fields, RenderQueryField{
			RawName:      param.Name,
			ArgName:      argName,
			ValueExpr:    valueExpr,
			TypeName:     typeName,
			Required:     param.Required,
			DefaultValue: "",
		})
	}
	return fields
}

func shouldSuppressSwaggerMethodDocstring(binding OperationBinding) bool {
	if binding.Mapping == nil {
		return false
	}
	pagination := strings.TrimSpace(binding.Mapping.Pagination)
	if pagination == "token" || pagination == "number" {
		return true
	}
	if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" &&
		strings.ToLower(methodOverride) != strings.ToLower(binding.Details.Method) {
		return true
	}
	return false
}

func singleLineDescription(value string) string {
	lines := descriptionLines(value)
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, " ")
}

func buildSwaggerMethodDocstring(
	doc *openapi.Document,
	binding OperationBinding,
	details openapi.OperationDetails,
	pathParamNameMap map[string]string,
	queryFields []RenderQueryField,
	bodyFieldNames []string,
	requestBodyType string,
	returnType string,
	paramAliases map[string]string,
) string {
	if shouldSuppressSwaggerMethodDocstring(binding) {
		return ""
	}

	docLines := make([]string, 0, 16)
	summary := strings.TrimSpace(details.Summary)
	if summary != "" {
		docLines = append(docLines, summary)
	}
	description := strings.TrimSpace(details.Description)
	descriptionText := singleLineDescription(description)
	if descriptionText != "" && descriptionText != summary {
		if len(docLines) > 0 {
			docLines = append(docLines, "")
		}
		docLines = append(docLines, descriptionLines(description)...)
	}

	paramDocs := make([]string, 0, len(details.Parameters)+len(bodyFieldNames))
	for _, param := range details.PathParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := strings.TrimSpace(pathParamNameMap[param.Name])
		if argName == "" {
			argName = OperationArgName(param.Name, paramAliases)
		}
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	queryArgByRaw := make(map[string]string, len(queryFields))
	for _, field := range queryFields {
		queryArgByRaw[field.RawName] = field.ArgName
	}
	for _, param := range details.QueryParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := strings.TrimSpace(queryArgByRaw[param.Name])
		if argName == "" {
			argName = OperationArgName(param.Name, paramAliases)
		}
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	for _, param := range details.HeaderParameters {
		description := singleLineDescription(param.Description)
		if description == "" {
			continue
		}
		argName := OperationArgName(param.Name, paramAliases)
		paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
	}

	if len(bodyFieldNames) > 0 && details.RequestBodySchema != nil {
		seen := map[string]struct{}{}
		for _, bodyField := range bodyFieldNames {
			if _, exists := seen[bodyField]; exists {
				continue
			}
			seen[bodyField] = struct{}{}
			descriptionLines := schemaCommentLines(doc, BodyFieldSchema(doc, details.RequestBodySchema, bodyField))
			if len(descriptionLines) == 0 {
				continue
			}
			description := strings.Join(descriptionLines, " ")
			argName := OperationArgName(bodyField, paramAliases)
			paramDocs = append(paramDocs, fmt.Sprintf(":param %s: %s", argName, description))
		}
	} else if requestBodyType != "" && details.RequestBody != nil {
		description := singleLineDescription(details.RequestBody.Description)
		if description != "" {
			paramDocs = append(paramDocs, fmt.Sprintf(":param body: %s", description))
		}
	}

	if len(paramDocs) > 0 {
		if len(docLines) > 0 {
			docLines = append(docLines, "")
		}
		docLines = append(docLines, paramDocs...)
	}

	if strings.TrimSpace(returnType) != "" && strings.TrimSpace(returnType) != "None" && details.Response != nil {
		description := singleLineDescription(details.Response.Description)
		if description != "" {
			if len(docLines) > 0 {
				docLines = append(docLines, "")
			}
			docLines = append(docLines, fmt.Sprintf(":return: %s", description))
		}
	}

	return strings.TrimSpace(strings.Join(docLines, "\n"))
}

func RenderOperationMethod(doc *openapi.Document, binding OperationBinding, async bool) string {
	return renderOperationMethodWithContext(doc, binding, async, "", "", config.CommentOverrides{})
}

func RenderOperationMethodWithComments(
	doc *openapi.Document,
	binding OperationBinding,
	async bool,
) string {
	return renderOperationMethodWithContext(doc, binding, async, "", "", config.CommentOverrides{})
}

func renderOperationMethodWithContext(
	doc *openapi.Document,
	binding OperationBinding,
	async bool,
	modulePath string,
	className string,
	commentOverrides config.CommentOverrides,
) string {
	details := binding.Details
	requestMethod := strings.ToLower(strings.TrimSpace(details.Method))
	paginationMode := ""
	returnType, returnCast := ReturnTypeInfo(doc, details.ResponseSchema)
	requestBodyType, bodyRequired := RequestBodyTypeInfo(doc, details.RequestBodySchema, details.RequestBody)
	ignoreHeaderParams := binding.Mapping != nil && binding.Mapping.IgnoreHeaderParams
	streamKeyword := binding.Mapping != nil && binding.Mapping.StreamKeyword
	streamWrap := binding.Mapping != nil && binding.Mapping.StreamWrap
	asyncIncludeKwargs := async && binding.Mapping != nil && binding.Mapping.AsyncIncludeKwargs
	paginationCastBeforeHeaders := binding.Mapping != nil && binding.Mapping.PaginationCastBeforeHeaders
	headersBeforePaginationParams := !paginationCastBeforeHeaders
	delegateTo := ""
	delegateAsyncYield := false
	streamWrapHandler := ""
	streamWrapFields := []string{}
	streamWrapAsyncYield := false
	streamWrapSyncResponseVar := "response"
	streamWrapCompactAsyncReturn := false
	streamWrapCompactSyncReturn := false
	streamWrapBlankLineBeforeAsyncReturn := false
	bodyFieldValues := map[string]string{}
	paramAliases := map[string]string{}
	argTypes := map[string]string{}
	bodyCallExprOverride := ""
	headersExpr := ""
	paginationRequestArg := "params"
	if binding.Mapping != nil && len(binding.Mapping.ParamAliases) > 0 {
		paramAliases = binding.Mapping.ParamAliases
	}
	if binding.Mapping != nil && len(binding.Mapping.ArgTypes) > 0 {
		argTypes = binding.Mapping.ArgTypes
	}
	if binding.Mapping != nil {
		if methodOverride := strings.TrimSpace(binding.Mapping.HTTPMethodOverride); methodOverride != "" {
			requestMethod = methodOverride
		}
		paginationMode = strings.TrimSpace(binding.Mapping.Pagination)
		if isTokenPagination(paginationMode) || isNumberPagination(paginationMode) {
			itemType := strings.TrimSpace(binding.Mapping.PaginationItemType)
			if itemType != "" {
				if isTokenPagination(paginationMode) {
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
		if async {
			if mappedAsyncReturnType := strings.TrimSpace(binding.Mapping.AsyncResponseType); mappedAsyncReturnType != "" {
				returnType = mappedAsyncReturnType
			}
		}
		if mappedReturnCast := strings.TrimSpace(binding.Mapping.ResponseCast); mappedReturnCast != "" {
			returnCast = mappedReturnCast
		} else if strings.TrimSpace(binding.Mapping.ResponseType) != "" {
			returnCast = strings.TrimSpace(binding.Mapping.ResponseType)
		}
		delegateTo = strings.TrimSpace(binding.Mapping.DelegateTo)
		delegateAsyncYield = binding.Mapping.DelegateAsyncYield
		streamWrapHandler = strings.TrimSpace(binding.Mapping.StreamWrapHandler)
		if len(binding.Mapping.StreamWrapFields) > 0 {
			streamWrapFields = append(streamWrapFields, binding.Mapping.StreamWrapFields...)
		}
		streamWrapAsyncYield = binding.Mapping.StreamWrapAsyncYield
		if varName := strings.TrimSpace(binding.Mapping.StreamWrapSyncResponseVar); varName != "" {
			streamWrapSyncResponseVar = varName
		}
		streamWrapCompactAsyncReturn = binding.Mapping.StreamWrapCompactAsyncReturn
		streamWrapCompactSyncReturn = binding.Mapping.StreamWrapCompactSyncReturn
		streamWrapBlankLineBeforeAsyncReturn = binding.Mapping.StreamWrapBlankLineBeforeAsync
		bodyCallExprOverride = strings.TrimSpace(binding.Mapping.BodyCallExpr)
		headersExpr = strings.TrimSpace(binding.Mapping.HeadersExpr)
		if override := strings.TrimSpace(binding.Mapping.PaginationRequestArg); override != "" {
			paginationRequestArg = override
		}
		if len(binding.Mapping.BodyFieldValues) > 0 {
			for k, v := range binding.Mapping.BodyFieldValues {
				bodyFieldValues[k] = v
			}
		}
	}
	if ignoreHeaderParams {
		details.HeaderParameters = nil
	}
	dataField := ""
	requestStream := false
	queryBuilder := "dump_exclude_none"
	bodyBuilder := "dump_exclude_none"
	bodyAnnotation := ""
	compactSingleItemMaps := false
	if binding.Mapping != nil {
		dataField = strings.TrimSpace(binding.Mapping.DataField)
		requestStream = binding.Mapping.RequestStream
		queryBuilder = normalizeMapBuilder(binding.Mapping.QueryBuilder)
		if async {
			if override := strings.TrimSpace(binding.Mapping.QueryBuilderAsync); override != "" {
				queryBuilder = normalizeMapBuilder(override)
			}
		} else {
			if override := strings.TrimSpace(binding.Mapping.QueryBuilderSync); override != "" {
				queryBuilder = normalizeMapBuilder(override)
			}
		}
		bodyBuilder = normalizeMapBuilder(binding.Mapping.BodyBuilder)
		bodyAnnotation = strings.TrimSpace(binding.Mapping.BodyAnnotation)
		compactSingleItemMaps = binding.Mapping.CompactSingleItemMaps
		if async && binding.Mapping.CompactSingleItemMapsAsync {
			compactSingleItemMaps = true
		}
		if !async && binding.Mapping.CompactSingleItemMapsSync {
			compactSingleItemMaps = true
		}
	}
	bodyFieldNames := make([]string, 0)
	bodyFixedValues := map[string]string{}
	filesFieldNames := make([]string, 0)
	filesFieldValues := map[string]string{}
	filesExpr := ""
	filesBeforeBody := false
	if binding.Mapping != nil && len(binding.Mapping.BodyFields) > 0 {
		bodyFieldNames = append(bodyFieldNames, binding.Mapping.BodyFields...)
	}
	if binding.Mapping != nil && len(binding.Mapping.BodyFixedValues) > 0 {
		for k, v := range binding.Mapping.BodyFixedValues {
			bodyFixedValues[k] = v
		}
	}
	if binding.Mapping != nil && len(binding.Mapping.FilesFields) > 0 {
		filesFieldNames = append(filesFieldNames, binding.Mapping.FilesFields...)
	}
	if binding.Mapping != nil && len(binding.Mapping.FilesFieldValues) > 0 {
		for k, v := range binding.Mapping.FilesFieldValues {
			filesFieldValues[k] = v
		}
	}
	if binding.Mapping != nil {
		filesExpr = strings.TrimSpace(binding.Mapping.FilesExpr)
		filesBeforeBody = binding.Mapping.FilesBeforeBody
	}
	if binding.Mapping != nil && binding.Mapping.DisableRequestBody {
		requestBodyType = ""
		bodyRequired = false
		bodyFieldNames = nil
		bodyFixedValues = map[string]string{}
	}
	queryFields := buildRenderQueryFields(doc, details, binding.Mapping, paramAliases, argTypes)
	signatureQueryFields := queryFields
	if binding.Mapping != nil && len(binding.Mapping.SignatureQueryFields) > 0 {
		signatureQueryFields = OrderSignatureQueryFields(queryFields, binding.Mapping.SignatureQueryFields)
	}
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
	signatureArgNames := map[string]struct{}{}
	for _, param := range details.PathParameters {
		name := OperationArgName(param.Name, paramAliases)
		pathParamNameMap[param.Name] = name
		typeName := TypeOverride(param.Name, true, PythonTypeForSchema(doc, param.Schema, true), argTypes)
		signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		signatureArgNames[name] = struct{}{}
	}
	for _, field := range signatureQueryFields {
		defaultValue := strings.TrimSpace(field.DefaultValue)
		if override, ok := OperationArgDefault(binding.Mapping, field.RawName, field.ArgName, async); ok {
			defaultValue = override
		}
		if field.Required && defaultValue == "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", field.ArgName, field.TypeName))
		} else if defaultValue != "" {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", field.ArgName, field.TypeName, defaultValue))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", field.ArgName, field.TypeName))
		}
		signatureArgNames[field.ArgName] = struct{}{}
	}
	for _, param := range details.HeaderParameters {
		name := OperationArgName(param.Name, paramAliases)
		typeName := TypeOverride(param.Name, param.Required, PythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		if defaultValue, ok := OperationArgDefault(binding.Mapping, param.Name, name, async); ok {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", name, typeName, defaultValue))
		} else if param.Required {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", name, typeName))
		}
		signatureArgNames[name] = struct{}{}
	}

	if len(bodyFieldNames) > 0 {
		for _, bodyField := range bodyFieldNames {
			argName := OperationArgName(bodyField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := BodyFieldSchema(doc, details.RequestBodySchema, bodyField)
			required := bodyRequiredSet[bodyField]
			typeName := TypeOverride(bodyField, required, PythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := OperationArgDefault(binding.Mapping, bodyField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	} else if requestBodyType != "" {
		if defaultValue, ok := OperationArgDefault(binding.Mapping, "body", "body", async); ok {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = %s", requestBodyType, defaultValue))
		} else if bodyRequired {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: %s", requestBodyType))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = None", requestBodyType))
		}
	}
	if len(filesFieldNames) > 0 {
		for _, filesField := range filesFieldNames {
			argName := OperationArgName(filesField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := BodyFieldSchema(doc, details.RequestBodySchema, filesField)
			required := bodyRequiredSet[filesField]
			typeName := TypeOverride(filesField, required, PythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := OperationArgDefault(binding.Mapping, filesField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	}
	includeKwargsHeaders := true
	if includeKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	}
	if asyncIncludeKwargs && !includeKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	}
	signatureArgs = NormalizeSignatureArgs(signatureArgs)

	methodKeyword := "def"
	requestCall := "self._requester.request"
	if async {
		methodKeyword = "async def"
		requestCall = "await self._requester.arequest"
	}
	headersAssigned := false

	var buf bytes.Buffer
	returnAnnotation := fmt.Sprintf(" -> %s", returnType)
	nonKwargsSignatureArgCount := 0
	for _, argDecl := range signatureArgs {
		if IsKwargsSignatureArg(argDecl) {
			continue
		}
		nonKwargsSignatureArgCount++
	}
	compactSignature := nonKwargsSignatureArgCount <= 2
	if binding.Mapping != nil {
		if binding.Mapping.ForceMultilineSignature {
			compactSignature = false
		}
		if async {
			if binding.Mapping.ForceMultilineSignatureAsync {
				compactSignature = false
			}
		} else {
			if binding.Mapping.ForceMultilineSignatureSync {
				compactSignature = false
			}
		}
	}
	if compactSignature {
		if len(signatureArgs) == 0 {
			buf.WriteString(fmt.Sprintf("    %s %s(self)%s:\n", methodKeyword, binding.MethodName, returnAnnotation))
		} else if len(signatureArgs) == 1 && IsKwargsSignatureArg(signatureArgs[0]) {
			buf.WriteString(fmt.Sprintf("    %s %s(self, %s)%s:\n", methodKeyword, binding.MethodName, signatureArgs[0], returnAnnotation))
		} else {
			buf.WriteString(fmt.Sprintf("    %s %s(self, *, %s)%s:\n", methodKeyword, binding.MethodName, strings.Join(signatureArgs, ", "), returnAnnotation))
		}
	} else {
		buf.WriteString(fmt.Sprintf("    %s %s(\n", methodKeyword, binding.MethodName))
		buf.WriteString("        self,\n")
		buf.WriteString("        *,\n")
		for _, arg := range signatureArgs {
			buf.WriteString(fmt.Sprintf("        %s,\n", arg))
		}
		buf.WriteString(fmt.Sprintf("    )%s:\n", returnAnnotation))
	}
	if binding.Mapping != nil && len(binding.Mapping.PreDocstringCode) > 0 {
		for _, block := range binding.Mapping.PreDocstringCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	methodDocstring := buildSwaggerMethodDocstring(doc, binding, details, pathParamNameMap, queryFields, bodyFieldNames, requestBodyType, returnType, paramAliases)
	docstringStyle := "block"
	if methodDocstring == "" && modulePath != "" && className != "" {
		key := strings.TrimSpace(modulePath) + "." + strings.TrimSpace(className) + "." + binding.MethodName
		if raw, ok := commentOverrides.MethodDocstrings[key]; ok {
			methodDocstring = strings.TrimSpace(raw)
			if style := strings.TrimSpace(commentOverrides.MethodDocstringStyles[key]); style != "" {
				docstringStyle = style
			}
		}
	}
	if methodDocstring != "" {
		WriteMethodDocstring(&buf, 2, methodDocstring, docstringStyle)
	}
	if delegateTo != "" {
		callArgs := BuildDelegateCallArgs(signatureArgs, binding.Mapping, async)
		RenderDelegatedCall(&buf, delegateTo, callArgs, async, delegateAsyncYield)
		return buf.String()
	}
	if binding.Mapping != nil && binding.Mapping.BlankLineAfterDocstring {
		buf.WriteString("\n")
	}

	urlPath := details.Path
	for rawName, pyName := range pathParamNameMap {
		if rawName == pyName {
			continue
		}
		urlPath = strings.ReplaceAll(urlPath, "{"+rawName+"}", "{"+pyName+"}")
	}
	buf.WriteString(fmt.Sprintf("        url = f\"{self._base_url}%s\"\n", urlPath))

	if len(queryFields) > 0 && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) {
		if queryBuilder == "raw" {
			if compactSingleItemMaps && len(queryFields) == 1 {
				field := queryFields[0]
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				buf.WriteString(fmt.Sprintf("        params = {%q: %s}\n", field.RawName, valueExpr))
			} else {
				buf.WriteString("        params = {\n")
				itemIndent := "            "
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				buf.WriteString("        }\n")
			}
		} else {
			if compactSingleItemMaps && len(queryFields) == 1 {
				field := queryFields[0]
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				buf.WriteString(fmt.Sprintf("        params = %s({%q: %s})\n", queryBuilder, field.RawName, valueExpr))
			} else {
				buf.WriteString(fmt.Sprintf("        params = %s(\n", queryBuilder))
				buf.WriteString("            {\n")
				itemIndent := "                "
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				buf.WriteString("            }\n")
				buf.WriteString("        )\n")
			}
		}
	}

	if headersExpr == "" && includeKwargsHeaders && (isTokenPagination(paginationMode) || isNumberPagination(paginationMode) || len(details.HeaderParameters) > 0) {
		buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
	}

	if len(details.HeaderParameters) > 0 {
		buf.WriteString("        header_values = dict(headers or {})\n")
		for _, param := range details.HeaderParameters {
			name := OperationArgName(param.Name, paramAliases)
			if param.Required {
				buf.WriteString(fmt.Sprintf("        header_values[%q] = str(%s)\n", param.Name, name))
			} else {
				buf.WriteString(fmt.Sprintf("        if %s is not None:\n", name))
				buf.WriteString(fmt.Sprintf("            header_values[%q] = str(%s)\n", param.Name, name))
			}
		}
		buf.WriteString("        headers = header_values\n")
	}
	if headersExpr != "" {
		buf.WriteString(fmt.Sprintf("        headers = %s\n", headersExpr))
		headersAssigned = true
		if isTokenPagination(paginationMode) || isNumberPagination(paginationMode) {
			buf.WriteString("\n")
		}
	}
	if (isTokenPagination(paginationMode) || isNumberPagination(paginationMode)) && binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			AppendIndentedCode(&buf, block, 2)
		}
		if headersExpr == "" {
			buf.WriteString("\n")
		}
	}
	includePaginationHeaders := true
	paginationRequestMethod := strings.ToUpper(requestMethod)
	if binding.Mapping != nil {
		if override := strings.TrimSpace(binding.Mapping.PaginationHTTPMethod); override != "" {
			paginationRequestMethod = override
		}
	}

	if isTokenPagination(paginationMode) && binding.Mapping != nil {
		dataClass := strings.TrimSpace(binding.Mapping.PaginationDataClass)
		paginationParamsVariable := binding.Mapping.PaginationParamsVariable
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
		if tokenInitExpr := strings.TrimSpace(binding.Mapping.PaginationInitPageTokenExpr); tokenInitExpr != "" {
			tokenExpr = tokenInitExpr
		}
		EnsureTrailingNewlines(&buf, 2)
		if async {
			buf.WriteString("        async def request_maker(i_page_token: str, i_page_size: int) -> HTTPRequest:\n")
			if paginationParamsVariable {
				if queryBuilder == "raw" {
					buf.WriteString("            params = {\n")
				} else {
					buf.WriteString(fmt.Sprintf("            params = %s(\n", queryBuilder))
					buf.WriteString("                {\n")
				}
				itemIndent := "                    "
				if queryBuilder != "raw" {
					itemIndent = "                    "
				}
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					if field.RawName == pageTokenField {
						valueExpr = "i_page_token"
					}
					if field.RawName == pageSizeField {
						valueExpr = "i_page_size"
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				if queryBuilder == "raw" {
					buf.WriteString("            }\n")
				} else {
					buf.WriteString("                }\n")
					buf.WriteString("            )\n")
				}
			}
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if headersBeforePaginationParams && includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if paginationParamsVariable {
				buf.WriteString(fmt.Sprintf("                %s=params,\n", paginationRequestArg))
			} else {
				if queryBuilder == "raw" {
					buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
				} else {
					buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
					buf.WriteString("                    {\n")
				}
				itemIndent := "                        "
				if queryBuilder == "raw" {
					itemIndent = "                    "
				}
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					if field.RawName == pageTokenField {
						valueExpr = "i_page_token"
					}
					if field.RawName == pageSizeField {
						valueExpr = "i_page_size"
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				if queryBuilder == "raw" {
					buf.WriteString("                },\n")
				} else {
					buf.WriteString("                    }\n")
					buf.WriteString("                ),\n")
				}
			}
			if includePaginationHeaders {
				if paginationCastBeforeHeaders && !headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				} else {
					buf.WriteString("                headers=headers,\n")
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				}
			} else {
				buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			}
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
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
			if paginationParamsVariable {
				if queryBuilder == "raw" {
					buf.WriteString("            params = {\n")
				} else {
					buf.WriteString(fmt.Sprintf("            params = %s(\n", queryBuilder))
					buf.WriteString("                {\n")
				}
				itemIndent := "                    "
				if queryBuilder != "raw" {
					itemIndent = "                    "
				}
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					if field.RawName == pageTokenField {
						valueExpr = "i_page_token"
					}
					if field.RawName == pageSizeField {
						valueExpr = "i_page_size"
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				if queryBuilder == "raw" {
					buf.WriteString("            }\n")
				} else {
					buf.WriteString("                }\n")
					buf.WriteString("            )\n")
				}
			}
			buf.WriteString("            return self._requester.make_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if headersBeforePaginationParams && includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if paginationParamsVariable {
				buf.WriteString(fmt.Sprintf("                %s=params,\n", paginationRequestArg))
			} else {
				if queryBuilder == "raw" {
					buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
				} else {
					buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
					buf.WriteString("                    {\n")
				}
				itemIndent := "                        "
				if queryBuilder == "raw" {
					itemIndent = "                    "
				}
				for _, field := range queryFields {
					valueExpr := field.ValueExpr
					if strings.TrimSpace(valueExpr) == "" {
						valueExpr = field.ArgName
					}
					if field.RawName == pageTokenField {
						valueExpr = "i_page_token"
					}
					if field.RawName == pageSizeField {
						valueExpr = "i_page_size"
					}
					buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
				}
				if queryBuilder == "raw" {
					buf.WriteString("                },\n")
				} else {
					buf.WriteString("                    }\n")
					buf.WriteString("                ),\n")
				}
			}
			if includePaginationHeaders {
				if paginationCastBeforeHeaders && !headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				} else {
					buf.WriteString("                headers=headers,\n")
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				}
			} else {
				buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			}
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
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
	if isNumberPagination(paginationMode) && binding.Mapping != nil {
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
		EnsureTrailingNewlines(&buf, 2)
		if async {
			buf.WriteString("        async def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if headersBeforePaginationParams && includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			if includePaginationHeaders {
				if paginationCastBeforeHeaders && !headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				} else {
					buf.WriteString("                headers=headers,\n")
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				}
			} else {
				buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			}
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
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
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if headersBeforePaginationParams && includePaginationHeaders {
				buf.WriteString("                headers=headers,\n")
			}
			if queryBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("                %s={\n", paginationRequestArg))
			} else {
				buf.WriteString(fmt.Sprintf("                %s=%s(\n", paginationRequestArg, queryBuilder))
				buf.WriteString("                    {\n")
			}
			itemIndent := "                        "
			if queryBuilder == "raw" {
				itemIndent = "                    "
			}
			for _, field := range queryFields {
				valueExpr := field.ValueExpr
				if strings.TrimSpace(valueExpr) == "" {
					valueExpr = field.ArgName
				}
				if field.RawName == pageNumField {
					valueExpr = "i_page_num"
				}
				if field.RawName == pageSizeField {
					valueExpr = "i_page_size"
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, field.RawName, valueExpr))
			}
			if queryBuilder == "raw" {
				buf.WriteString("                },\n")
			} else {
				buf.WriteString("                    }\n")
				buf.WriteString("                ),\n")
			}
			if includePaginationHeaders {
				if paginationCastBeforeHeaders && !headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if headersBeforePaginationParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				} else {
					buf.WriteString("                headers=headers,\n")
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
				}
			} else {
				buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
			}
			if dataField != "" {
				buf.WriteString(fmt.Sprintf("                data_field=%q,\n", dataField))
			}
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

	if headersExpr == "" && includeKwargsHeaders && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) && len(details.HeaderParameters) == 0 {
		if binding.Mapping != nil && binding.Mapping.BlankLineAfterHeaders {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
		} else {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
		}
		headersAssigned = true
	}
	if binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			AppendIndentedCode(&buf, block, 2)
		}
	}
	bodyVarAssign := "body"
	if bodyAnnotation != "" {
		bodyVarAssign = fmt.Sprintf("body: %s", bodyAnnotation)
	}
	renderFilesAssignment := func() bool {
		if filesExpr != "" {
			buf.WriteString(fmt.Sprintf("        files = %s\n", filesExpr))
			return true
		}
		if len(filesFieldNames) == 0 {
			return false
		}
		if len(filesFieldNames) == 1 {
			fieldName := filesFieldNames[0]
			argName := OperationArgName(fieldName, paramAliases)
			valueExpr := argName
			if override, ok := filesFieldValues[fieldName]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
			buf.WriteString(fmt.Sprintf("        files = {%q: %s}\n", fieldName, valueExpr))
			return true
		}

		buf.WriteString("        files = {\n")
		for _, filesField := range filesFieldNames {
			argName := OperationArgName(filesField, paramAliases)
			valueExpr := argName
			if override, ok := filesFieldValues[filesField]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
			buf.WriteString(fmt.Sprintf("            %q: %s,\n", filesField, valueExpr))
		}
		buf.WriteString("        }\n")
		return true
	}

	if filesBeforeBody {
		renderFilesAssignment()
	}

	if len(bodyFieldNames) > 0 {
		totalBodyItems := len(bodyFieldNames) + len(bodyFixedValues)
		if compactSingleItemMaps && totalBodyItems == 1 {
			fieldName := ""
			valueExpr := ""
			if len(bodyFieldNames) == 1 {
				fieldName = bodyFieldNames[0]
				argName := OperationArgName(fieldName, paramAliases)
				valueExpr = argName
				if override, ok := bodyFieldValues[fieldName]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
			} else {
				fixedKeys := make([]string, 0, len(bodyFixedValues))
				for k := range bodyFixedValues {
					fixedKeys = append(fixedKeys, k)
				}
				sort.Strings(fixedKeys)
				fieldName = fixedKeys[0]
				valueExpr = bodyFixedValues[fieldName]
			}
			if bodyBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("        %s = {%q: %s}\n", bodyVarAssign, fieldName, valueExpr))
			} else {
				buf.WriteString(fmt.Sprintf("        %s = %s({%q: %s})\n", bodyVarAssign, bodyBuilder, fieldName, valueExpr))
			}
		} else if bodyBuilder == "raw" {
			buf.WriteString(fmt.Sprintf("        %s = {\n", bodyVarAssign))
			itemIndent := "            "
			for _, bodyField := range bodyFieldNames {
				argName := OperationArgName(bodyField, paramAliases)
				valueExpr := argName
				if override, ok := bodyFieldValues[bodyField]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, bodyField, valueExpr))
			}
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("        }\n")
		} else {
			buf.WriteString(fmt.Sprintf("        %s = %s(\n", bodyVarAssign, bodyBuilder))
			buf.WriteString("            {\n")
			itemIndent := "                "
			for _, bodyField := range bodyFieldNames {
				argName := OperationArgName(bodyField, paramAliases)
				valueExpr := argName
				if override, ok := bodyFieldValues[bodyField]; ok && strings.TrimSpace(override) != "" {
					valueExpr = strings.TrimSpace(override)
				}
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, bodyField, valueExpr))
			}
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("            }\n")
			buf.WriteString("        )\n")
		}
	} else if len(bodyFixedValues) > 0 {
		if compactSingleItemMaps && len(bodyFixedValues) == 1 {
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for k := range bodyFixedValues {
				fixedKeys = append(fixedKeys, k)
			}
			sort.Strings(fixedKeys)
			fieldName := fixedKeys[0]
			if bodyBuilder == "raw" {
				buf.WriteString(fmt.Sprintf("        %s = {%q: %s}\n", bodyVarAssign, fieldName, bodyFixedValues[fieldName]))
			} else {
				buf.WriteString(fmt.Sprintf("        %s = %s({%q: %s})\n", bodyVarAssign, bodyBuilder, fieldName, bodyFixedValues[fieldName]))
			}
		} else if bodyBuilder == "raw" {
			buf.WriteString(fmt.Sprintf("        %s = {\n", bodyVarAssign))
			itemIndent := "            "
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("        }\n")
		} else {
			buf.WriteString(fmt.Sprintf("        %s = %s(\n", bodyVarAssign, bodyBuilder))
			buf.WriteString("            {\n")
			itemIndent := "                "
			fixedKeys := make([]string, 0, len(bodyFixedValues))
			for fieldName := range bodyFixedValues {
				fixedKeys = append(fixedKeys, fieldName)
			}
			sort.Strings(fixedKeys)
			for _, fieldName := range fixedKeys {
				buf.WriteString(fmt.Sprintf("%s%q: %s,\n", itemIndent, fieldName, bodyFixedValues[fieldName]))
			}
			buf.WriteString("            }\n")
			buf.WriteString("        )\n")
		}
	} else if requestBodyType != "" {
		buf.WriteString("        request_body: Any = None\n")
		if bodyRequired {
			buf.WriteString("        request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		} else {
			buf.WriteString("        if body is not None:\n")
			buf.WriteString("            request_body = body.model_dump(exclude_none=True) if hasattr(body, \"model_dump\") else body\n")
		}
	}
	if !filesBeforeBody {
		renderFilesAssignment()
	}
	if headersExpr == "" && !headersAssigned && includeKwargsHeaders && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) && len(details.HeaderParameters) == 0 {
		needsBlankLine := len(queryFields) > 0 ||
			len(bodyFieldNames) > 0 ||
			len(bodyFixedValues) > 0 ||
			filesExpr != "" ||
			len(filesFieldNames) > 0 ||
			requestBodyType != "" ||
			strings.EqualFold(requestMethod, "delete") ||
			(binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0)
		if binding.Mapping != nil && binding.Mapping.NoBlankLineAfterHeaders {
			needsBlankLine = false
		}
		if needsBlankLine {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
		} else {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
		}
	}

	castExpr := "None"
	if returnCast != "" {
		castExpr = returnCast
	}
	bodyArgExpr := ""
	if len(bodyFieldNames) > 0 {
		bodyArgExpr = "body"
	} else if len(bodyFixedValues) > 0 {
		bodyArgExpr = "body"
	} else if requestBodyType != "" {
		bodyArgExpr = "request_body"
	}
	callArgs := []string{
		fmt.Sprintf("%q", requestMethod),
		"url",
	}
	type requestCallArg struct {
		Key  string
		Expr string
		Pos  bool
	}
	optionalArgs := make([]requestCallArg, 0, 8)
	streamLiteral := "False"
	if requestStream {
		streamLiteral = "True"
	}
	streamExpr := streamLiteral
	if streamKeyword {
		streamExpr = fmt.Sprintf("stream=%s", streamLiteral)
	}
	optionalArgs = append(optionalArgs, requestCallArg{Key: "stream", Expr: streamExpr, Pos: !streamKeyword})
	castExprValue := fmt.Sprintf("cast=%s", castExpr)
	optionalArgs = append(optionalArgs, requestCallArg{Key: "cast", Expr: castExprValue, Pos: false})
	if len(queryFields) > 0 {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "params", Expr: "params=params"})
	}
	hasHeadersArg := true
	if hasHeadersArg {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "headers", Expr: "headers=headers"})
	}
	if bodyArgExpr != "" {
		bodyExpr := bodyArgExpr
		if bodyCallExprOverride != "" {
			if strings.Contains(bodyCallExprOverride, "{body}") {
				bodyExpr = strings.ReplaceAll(bodyCallExprOverride, "{body}", bodyArgExpr)
			} else {
				bodyExpr = bodyCallExprOverride
			}
		}
		optionalArgs = append(optionalArgs, requestCallArg{Key: "body", Expr: fmt.Sprintf("body=%s", bodyExpr)})
	}
	if filesExpr != "" || len(filesFieldNames) > 0 {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "files", Expr: "files=files"})
	}
	if dataField != "" {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "data_field", Expr: fmt.Sprintf("data_field=%q", dataField)})
	}
	if binding.Mapping != nil && len(binding.Mapping.RequestCallArgOrder) > 0 {
		argByKey := map[string]requestCallArg{}
		defaultOrder := make([]string, 0, len(optionalArgs))
		for _, item := range optionalArgs {
			if _, exists := argByKey[item.Key]; exists {
				continue
			}
			argByKey[item.Key] = item
			defaultOrder = append(defaultOrder, item.Key)
		}
		orderedKeys := make([]string, 0, len(optionalArgs))
		seen := map[string]struct{}{}
		for _, rawKey := range binding.Mapping.RequestCallArgOrder {
			key := strings.TrimSpace(rawKey)
			if key == "" {
				continue
			}
			if _, exists := argByKey[key]; !exists {
				continue
			}
			if _, exists := seen[key]; exists {
				continue
			}
			orderedKeys = append(orderedKeys, key)
			seen[key] = struct{}{}
		}
		for _, key := range defaultOrder {
			if _, exists := seen[key]; exists {
				continue
			}
			orderedKeys = append(orderedKeys, key)
			seen[key] = struct{}{}
		}
		orderedItems := make([]requestCallArg, 0, len(orderedKeys))
		for _, key := range orderedKeys {
			orderedItems = append(orderedItems, argByKey[key])
		}
		for _, item := range orderedItems {
			if item.Pos {
				callArgs = append(callArgs, item.Expr)
			}
		}
		for _, item := range orderedItems {
			if !item.Pos {
				callArgs = append(callArgs, item.Expr)
			}
		}
	} else {
		for _, item := range optionalArgs {
			callArgs = append(callArgs, item.Expr)
		}
	}
	requestExpr := fmt.Sprintf("%s(%s)", requestCall, strings.Join(callArgs, ", "))
	forceMultilineRequestCall := binding.Mapping != nil && binding.Mapping.ForceMultilineRequestCall
	if binding.Mapping != nil {
		if async && binding.Mapping.ForceMultilineRequestCallAsync {
			forceMultilineRequestCall = true
		}
		if !async && binding.Mapping.ForceMultilineRequestCallSync {
			forceMultilineRequestCall = true
		}
	}
	if binding.Mapping != nil && binding.Mapping.BlankLineBeforeReturn {
		EnsureTrailingNewlines(&buf, 2)
	}
	if binding.Mapping != nil && binding.Mapping.ResponseUnwrapListFirst {
		buf.WriteString(fmt.Sprintf("        res = %s\n", requestExpr))
		buf.WriteString("        data = res.data[0]\n")
		buf.WriteString("        data._raw_response = res._raw_response\n")
		buf.WriteString("        return data\n")
		return buf.String()
	}
	if requestStream && streamWrap {
		fieldLiterals := make([]string, 0, len(streamWrapFields))
		for _, field := range streamWrapFields {
			trimmed := strings.TrimSpace(field)
			if trimmed == "" {
				continue
			}
			fieldLiterals = append(fieldLiterals, fmt.Sprintf("%q", trimmed))
		}
		if async {
			if forceMultilineRequestCall {
				buf.WriteString(fmt.Sprintf("        resp: AsyncIteratorHTTPResponse[str] = %s(\n", requestCall))
				for _, arg := range callArgs {
					buf.WriteString(fmt.Sprintf("            %s,\n", arg))
				}
				buf.WriteString("        )\n")
			} else {
				buf.WriteString(fmt.Sprintf("        resp: AsyncIteratorHTTPResponse[str] = %s\n", requestExpr))
			}
			if streamWrapAsyncYield {
				buf.WriteString("        async for item in AsyncStream(\n")
				buf.WriteString("            resp.data,\n")
				if len(fieldLiterals) > 0 {
					buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
				}
				if streamWrapHandler != "" {
					buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
				}
				buf.WriteString("            raw_response=resp._raw_response,\n")
				buf.WriteString("        ):\n")
				buf.WriteString("            yield item\n")
			} else {
				if streamWrapBlankLineBeforeAsyncReturn {
					buf.WriteString("\n")
				}
				if streamWrapCompactAsyncReturn {
					asyncStreamArgs := []string{"resp.data"}
					if len(fieldLiterals) > 0 {
						asyncStreamArgs = append(asyncStreamArgs, fmt.Sprintf("fields=[%s]", strings.Join(fieldLiterals, ", ")))
					}
					if streamWrapHandler != "" {
						asyncStreamArgs = append(asyncStreamArgs, fmt.Sprintf("handler=%s", streamWrapHandler))
					}
					asyncStreamArgs = append(asyncStreamArgs, "raw_response=resp._raw_response")
					buf.WriteString(fmt.Sprintf("        return AsyncStream(%s)\n", strings.Join(asyncStreamArgs, ", ")))
				} else {
					buf.WriteString("        return AsyncStream(\n")
					buf.WriteString("            resp.data,\n")
					if len(fieldLiterals) > 0 {
						buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
					}
					if streamWrapHandler != "" {
						buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
					}
					buf.WriteString("            raw_response=resp._raw_response,\n")
					buf.WriteString("        )\n")
				}
			}
		} else {
			if forceMultilineRequestCall {
				buf.WriteString(fmt.Sprintf("        %s: IteratorHTTPResponse[str] = %s(\n", streamWrapSyncResponseVar, requestCall))
				for _, arg := range callArgs {
					buf.WriteString(fmt.Sprintf("            %s,\n", arg))
				}
				buf.WriteString("        )\n")
			} else {
				buf.WriteString(fmt.Sprintf("        %s: IteratorHTTPResponse[str] = %s\n", streamWrapSyncResponseVar, requestExpr))
			}
			if streamWrapCompactSyncReturn {
				streamArgs := []string{
					fmt.Sprintf("%s._raw_response", streamWrapSyncResponseVar),
					fmt.Sprintf("%s.data", streamWrapSyncResponseVar),
				}
				if len(fieldLiterals) > 0 {
					streamArgs = append(streamArgs, fmt.Sprintf("fields=[%s]", strings.Join(fieldLiterals, ", ")))
				}
				if streamWrapHandler != "" {
					streamArgs = append(streamArgs, fmt.Sprintf("handler=%s", streamWrapHandler))
				}
				buf.WriteString(fmt.Sprintf("        return Stream(%s)\n", strings.Join(streamArgs, ", ")))
			} else {
				buf.WriteString("        return Stream(\n")
				buf.WriteString(fmt.Sprintf("            %s._raw_response,\n", streamWrapSyncResponseVar))
				buf.WriteString(fmt.Sprintf("            %s.data,\n", streamWrapSyncResponseVar))
				if len(fieldLiterals) > 0 {
					buf.WriteString(fmt.Sprintf("            fields=[%s],\n", strings.Join(fieldLiterals, ", ")))
				}
				if streamWrapHandler != "" {
					buf.WriteString(fmt.Sprintf("            handler=%s,\n", streamWrapHandler))
				}
				buf.WriteString("        )\n")
			}
		}
	} else {
		if forceMultilineRequestCall {
			buf.WriteString(fmt.Sprintf("        return %s(\n", requestCall))
			for _, arg := range callArgs {
				buf.WriteString(fmt.Sprintf("            %s,\n", arg))
			}
			buf.WriteString("        )\n")
		} else {
			buf.WriteString(fmt.Sprintf("        return %s\n", requestExpr))
		}
	}

	return buf.String()
}

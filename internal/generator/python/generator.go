package python

import (
	"bytes"
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
	Package     *config.Package
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
	rootInitContent, err := renderPythonRootInit(rootDir)
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "__init__.py"), rootInitContent); err != nil {
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
	modelDefs, inferredSchemaAliases := resolvePackageModelDefinitions(doc, meta, bindings)
	schemaAliases := packageSchemaAliases(doc, meta)
	for schemaName, modelName := range inferredSchemaAliases {
		if strings.TrimSpace(schemaName) == "" || strings.TrimSpace(modelName) == "" {
			continue
		}
		if _, exists := schemaAliases[schemaName]; exists {
			continue
		}
		schemaAliases[schemaName] = modelName
	}
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
			queryBuilder := "dump_exclude_none"
			bodyBuilder := "dump_exclude_none"
			if binding.Mapping != nil {
				queryBuilder = normalizeMapBuilder(binding.Mapping.QueryBuilder)
				bodyBuilder = normalizeMapBuilder(binding.Mapping.BodyBuilder)
			}
			hasQueryFields := len(binding.Details.QueryParameters) > 0
			if binding.Mapping != nil && len(binding.Mapping.QueryFields) > 0 {
				hasQueryFields = true
			}
			hasBodyMap := binding.Mapping != nil && (len(binding.Mapping.BodyFields) > 0 || len(binding.Mapping.BodyFixedValues) > 0)
			if hasQueryFields && (mappingGeneratesSync(binding.Mapping) || mappingGeneratesAsync(binding.Mapping)) {
				if queryBuilder == "dump_exclude_none" {
					needDumpExcludeNone = true
				}
				if queryBuilder == "remove_none_values" {
					needRemoveNoneValues = true
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
	syncExtraMethods := []string(nil)
	if meta.Package != nil {
		syncExtraMethods = meta.Package.SyncExtraMethods
	}
	syncMethodNames := collectClassMethodNames(bindings, syncExtraMethods, false)
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
			Name: binding.MethodName,
			Content: renderOperationMethodWithContext(
				doc,
				binding,
				false,
				"cozepy."+meta.ModulePath,
				syncClass,
				commentOverrides,
				syncMethodNames,
			),
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
	asyncExtraMethods := []string(nil)
	if meta.Package != nil {
		asyncExtraMethods = meta.Package.AsyncExtraMethods
	}
	asyncMethodNames := collectClassMethodNames(bindings, asyncExtraMethods, true)
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
			Name: binding.MethodName,
			Content: renderOperationMethodWithContext(
				doc,
				binding,
				true,
				"cozepy."+meta.ModulePath,
				asyncClass,
				commentOverrides,
				asyncMethodNames,
			),
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

func collectClassMethodNames(bindings []OperationBinding, extraMethods []string, async bool) map[string]struct{} {
	names := make(map[string]struct{}, len(bindings)+len(extraMethods))
	for _, binding := range bindings {
		if async {
			if !mappingGeneratesAsync(binding.Mapping) {
				continue
			}
		} else {
			if !mappingGeneratesSync(binding.Mapping) {
				continue
			}
		}
		methodName := NormalizeMethodName(binding.MethodName)
		if strings.TrimSpace(methodName) == "" {
			continue
		}
		names[methodName] = struct{}{}
	}
	for _, block := range extraMethods {
		methodName := NormalizeMethodName(DetectMethodBlockName(block))
		if strings.TrimSpace(methodName) == "" {
			continue
		}
		names[methodName] = struct{}{}
	}
	return names
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

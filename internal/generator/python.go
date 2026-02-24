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
	if err := writePythonSDK(cfg, doc, packages, packageMetas, writer); err != nil {
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
	existingOps := map[string]struct{}{}

	for _, details := range allOps {
		existingOps[strings.ToLower(details.Method)+" "+details.Path] = struct{}{}
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

	for _, mapping := range cfg.API.OperationMappings {
		if !mapping.AllowMissingInSwagger {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(mapping.Method)) + " " + strings.TrimSpace(mapping.Path)
		if _, ok := existingOps[key]; ok {
			continue
		}
		if cfg.IsIgnored(mapping.Path, mapping.Method) {
			continue
		}
		if len(mapping.SDKMethods) == 0 {
			continue
		}
		mappingCopy := mapping
		details := syntheticOperationDetails(mappingCopy)
		for methodIndex, sdkMethod := range mappingCopy.SDKMethods {
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

	return deduplicateBindings(bindings)
}

func syntheticOperationDetails(mapping config.OperationMapping) openapi.OperationDetails {
	details := openapi.OperationDetails{
		Path:   strings.TrimSpace(mapping.Path),
		Method: strings.ToLower(strings.TrimSpace(mapping.Method)),
	}
	path := details.Path
	for {
		start := strings.Index(path, "{")
		if start < 0 {
			break
		}
		endOffset := strings.Index(path[start:], "}")
		if endOffset <= 1 {
			break
		}
		end := start + endOffset
		paramName := strings.TrimSpace(path[start+1 : end])
		if paramName == "" {
			path = path[end+1:]
			continue
		}
		details.PathParameters = append(details.PathParameters, openapi.ParameterSpec{
			Name:     paramName,
			In:       "path",
			Required: true,
			Schema:   &openapi.Schema{Type: "string"},
		})
		path = path[end+1:]
	}
	return details
}

func deduplicateBindings(bindings []operationBinding) []operationBinding {
	syncSeen := map[string]int{}
	asyncSeen := map[string]int{}
	nextSuffix := map[string]int{}
	for i := range bindings {
		key := bindings[i].PackageName + ":" + bindings[i].MethodName
		conflict := false
		if mappingGeneratesSync(bindings[i].Mapping) {
			syncSeen[key]++
			if syncSeen[key] > 1 {
				conflict = true
			}
		}
		if mappingGeneratesAsync(bindings[i].Mapping) {
			asyncSeen[key]++
			if asyncSeen[key] > 1 {
				conflict = true
			}
		}
		if conflict {
			suffix := nextSuffix[key]
			if suffix == 0 {
				suffix = 2
			}
			bindings[i].MethodName = fmt.Sprintf("%s_%d", bindings[i].MethodName, suffix)
			nextSuffix[key] = suffix + 1
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
	cfg *config.Config,
	doc *openapi.Document,
	packages map[string][]operationBinding,
	packageMetas map[string]packageMeta,
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
		if packageHasConfiguredContent(meta.Package) {
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
	logPy, err := renderLogPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "log.py"), logPy); err != nil {
		return err
	}
	exceptionPy, err := renderExceptionPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "exception.py"), exceptionPy); err != nil {
		return err
	}
	versionPy, err := renderVersionPy()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "version.py"), versionPy); err != nil {
		return err
	}
	pyprojectToml, err := renderPyprojectToml()
	if err != nil {
		return err
	}
	if err := writer.write(filepath.Join(outputDir, "pyproject.toml"), pyprojectToml); err != nil {
		return err
	}
	if err := writer.write(filepath.Join(rootDir, "py.typed"), ""); err != nil {
		return err
	}
	cozePy, err := renderCozePy(cfg, packageMetas)
	if err != nil {
		return err
	}
	if cozePy != "" {
		if err := writer.write(filepath.Join(rootDir, "coze.py"), cozePy); err != nil {
			return err
		}
	}

	for _, pkgName := range pkgNames {
		meta := packageMetas[pkgName]
		pkgDir := filepath.Join(rootDir, meta.DirPath)
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			return fmt.Errorf("create package directory %q: %w", pkgDir, err)
		}
		content := renderPackageModule(doc, meta, packages[pkgName], cfg.CommentOverrides)
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
		{relPath: "__init__.py", asset: "special/cozepy/__init__.py.raw"},
		{relPath: "auth/__init__.py", asset: "special/cozepy/auth/__init__.py.raw"},
		{relPath: "websockets/__init__.py", asset: "special/cozepy/websockets/__init__.py.raw"},
		{relPath: "websockets/audio/__init__.py", asset: "special/cozepy/websockets/audio/__init__.py.raw"},
		{relPath: "websockets/audio/speech/__init__.py", asset: "special/cozepy/websockets/audio/speech/__init__.py.raw"},
		{relPath: "websockets/audio/transcriptions/__init__.py", asset: "special/cozepy/websockets/audio/transcriptions/__init__.py.raw"},
		{relPath: "websockets/chat/__init__.py", asset: "special/cozepy/websockets/chat/__init__.py.raw"},
		{relPath: "websockets/ws.py", asset: "special/cozepy/websockets/ws.py.raw"},
	}
	for _, item := range specialAssets {
		content, err := renderPythonRawAsset(item.asset)
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

func renderCozePy(cfg *config.Config, packageMetas map[string]packageMeta) (string, error) {
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
		importNames := svc.AsyncClass + ", " + svc.SyncClass
		if svc.Attribute == "api_apps" || svc.Attribute == "apps" {
			importNames = svc.SyncClass + ", " + svc.AsyncClass
		}
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

func collectRootServices(cfg *config.Config, packageMetas map[string]packageMeta) []rootService {
	services := make([]rootService, 0)
	seen := map[string]struct{}{}
	for _, pkg := range cfg.API.Packages {
		name := normalizePackageName(pkg.Name)
		meta, ok := packageMetas[name]
		if !ok {
			continue
		}
		dir := strings.TrimSpace(meta.DirPath)
		if dir == "" || strings.Contains(dir, "/") {
			continue
		}
		attr := normalizePythonIdentifier(dir)
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
	return len(pkg.ChildClients) > 0 ||
		len(pkg.ModelSchemas) > 0 ||
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

func renderLogPy() (string, error) {
	return renderPythonTemplate("log.py.tpl", map[string]any{})
}

func renderExceptionPy() (string, error) {
	return renderPythonTemplate("exception.py.tpl", map[string]any{})
}

func renderVersionPy() (string, error) {
	return renderPythonTemplate("version.py.tpl", map[string]any{})
}

func renderPyprojectToml() (string, error) {
	return renderPythonTemplate("pyproject.toml.tpl", map[string]any{})
}

func renderPackageModule(
	doc *openapi.Document,
	meta packageMeta,
	bindings []operationBinding,
	commentOverrides config.CommentOverrides,
) string {
	var buf bytes.Buffer
	hasChildClients := meta.Package != nil && len(meta.Package.ChildClients) > 0
	childClientsForType := []config.ChildClient{}
	childClientsForInit := []config.ChildClient{}
	childClientsForSync := []config.ChildClient{}
	childClientsForAsync := []config.ChildClient{}
	if hasChildClients {
		childClientsForType = append([]config.ChildClient(nil), meta.Package.ChildClients...)
		childClientsForInit = append([]config.ChildClient(nil), meta.Package.ChildClients...)
		childClientsForSync = append([]config.ChildClient(nil), meta.Package.ChildClients...)
		childClientsForAsync = append([]config.ChildClient(nil), meta.Package.ChildClients...)
		if len(meta.Package.TypeCheckingChildOrder) > 0 {
			childClientsForType = orderChildClients(childClientsForType, meta.Package.TypeCheckingChildOrder)
		}
		if len(meta.Package.InitChildOrder) > 0 {
			childClientsForInit = orderChildClients(childClientsForInit, meta.Package.InitChildOrder)
		}
		if len(meta.Package.SyncChildOrder) > 0 {
			childClientsForSync = orderChildClients(childClientsForSync, meta.Package.SyncChildOrder)
		}
		if len(meta.Package.AsyncChildOrder) > 0 {
			childClientsForAsync = orderChildClients(childClientsForAsync, meta.Package.AsyncChildOrder)
		}
	}
	hasTypedChildClients := false
	if hasChildClients {
		for _, child := range childClientsForType {
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
	needsTokenPagedResponseImport := false
	needsNumberPagedResponseImport := false
	for _, binding := range bindings {
		if binding.Mapping == nil {
			continue
		}
		mode := strings.TrimSpace(binding.Mapping.Pagination)
		if isTokenPagination(mode) && paginationInheritResponse(binding.Mapping) {
			needsTokenPagedResponseImport = true
		}
		if isNumberPagination(mode) && paginationInheritResponse(binding.Mapping) {
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
		if hasModelClasses {
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
		modelImports = orderedUniqueByPriority(modelImports, []string{
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
				builders := []string{queryBuilder}
				if queryBuilderSync != "" {
					builders = append(builders, queryBuilderSync)
				}
				if queryBuilderAsync != "" {
					builders = append(builders, queryBuilderAsync)
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
			if hasBodyMap {
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
		utilImports = orderedUniqueByPriority(utilImports, []string{
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

	if hasTypedChildClients {
		buf.WriteString("\nif TYPE_CHECKING:\n")
		for _, child := range childClientsForType {
			if child.DisableTypeHints {
				continue
			}
			typeModule := strings.TrimSpace(child.TypeImportModule)
			if typeModule == "" {
				typeModule = strings.TrimSpace(child.Module)
			}
			if !strings.HasPrefix(typeModule, ".") {
				typeModule = childTypeImportModule(meta, typeModule)
			}
			if typeModule == "" {
				continue
			}
			if child.TypeImportSyncFirst {
				buf.WriteString(fmt.Sprintf("    from %s import %s, %s\n", typeModule, child.SyncClass, child.AsyncClass))
			} else {
				buf.WriteString(fmt.Sprintf("    from %s import %s, %s\n", typeModule, child.AsyncClass, child.SyncClass))
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
			appendIndentedCode(&buf, block, 0)
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
		pagedResponseClasses := renderPagedResponseClasses(bindings, overridePaginationClasses)
		if strings.TrimSpace(pagedResponseClasses) != "" {
			buf.WriteString(pagedResponseClasses)
			buf.WriteString("\n")
		}
	}
	if meta.Package != nil && len(meta.Package.TopLevelCode) > 0 {
		for _, block := range meta.Package.TopLevelCode {
			appendIndentedCode(&buf, block, 0)
			buf.WriteString("\n")
		}
	}

	syncClass := packageClientClassName(meta, false)
	asyncClass := packageClientClassName(meta, true)
	syncClassKey := "cozepy." + meta.ModulePath + "." + syncClass
	asyncClassKey := "cozepy." + meta.ModulePath + "." + asyncClass
	blankLineBeforeChildInits := meta.Package != nil && meta.Package.BlankLineBeforeChildInits
	blankLineBeforeSyncInitCode := meta.Package != nil && meta.Package.BlankLineBeforeSyncInit
	blankLineBeforeAsyncInitCode := meta.Package != nil && meta.Package.BlankLineBeforeAsyncInit

	ensureTrailingNewlines(&buf, 3)
	buf.WriteString(fmt.Sprintf("class %s(object):\n", syncClass))
	if classDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[syncClassKey]); classDoc != "" {
		style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[syncClassKey])
		writeClassDocstring(&buf, 1, classDoc, style)
	}
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	if meta.Package != nil && len(meta.Package.SyncInitPreCode) > 0 {
		for _, block := range meta.Package.SyncInitPreCode {
			appendIndentedCode(&buf, block, 2)
		}
	}
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n")
	if meta.Package != nil && len(meta.Package.SyncInitCode) > 0 {
		if blankLineBeforeSyncInitCode {
			buf.WriteString("\n")
		}
		for _, block := range meta.Package.SyncInitCode {
			appendIndentedCode(&buf, block, 2)
			buf.WriteString("\n")
		}
	}
	if hasChildClients {
		if blankLineBeforeChildInits {
			buf.WriteString("\n")
		}
		for _, child := range childClientsForInit {
			attribute := normalizePythonIdentifier(child.Attribute)
			if child.DisableTypeHints {
				buf.WriteString(fmt.Sprintf("        self._%s = None\n", attribute))
			} else {
				buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.SyncClass))
			}
		}
		buf.WriteString("\n")
	} else if meta.Package == nil || len(meta.Package.SyncInitCode) == 0 {
		buf.WriteString("\n")
	}
	syncMethodBlocks := make([]classMethodBlock, 0)
	if hasChildClients {
		for _, child := range childClientsForSync {
			attribute := normalizePythonIdentifier(child.Attribute)
			syncMethodBlocks = append(syncMethodBlocks, classMethodBlock{
				Name:    attribute,
				Content: renderChildClientProperty(meta, child, false, syncClassKey, commentOverrides),
			})
		}
	}
	for _, binding := range bindings {
		if !mappingGeneratesSync(binding.Mapping) {
			continue
		}
		syncMethodBlocks = append(syncMethodBlocks, classMethodBlock{
			Name:    binding.MethodName,
			Content: renderOperationMethodWithComments(doc, binding, false, "cozepy."+meta.ModulePath, syncClass, commentOverrides),
		})
	}
	if meta.Package != nil && len(meta.Package.SyncExtraMethods) > 0 {
		for _, block := range meta.Package.SyncExtraMethods {
			content := indentCodeBlock(block, 1)
			content = applyMethodDocstringOverrides(content, syncClassKey, commentOverrides)
			syncMethodBlocks = append(syncMethodBlocks, classMethodBlock{
				Name:    detectMethodBlockName(block),
				Content: content,
			})
		}
	}
	if meta.Package != nil && len(meta.Package.SyncMethodOrder) > 0 {
		syncMethodBlocks = orderClassMethodBlocks(syncMethodBlocks, meta.Package.SyncMethodOrder)
	}
	for _, block := range syncMethodBlocks {
		buf.WriteString(strings.TrimRight(block.Content, "\n"))
		buf.WriteString("\n\n")
	}
	buf.WriteString("\n")

	buf.WriteString(fmt.Sprintf("class %s(object):\n", asyncClass))
	if classDoc := strings.TrimSpace(commentOverrides.ClassDocstrings[asyncClassKey]); classDoc != "" {
		style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[asyncClassKey])
		writeClassDocstring(&buf, 1, classDoc, style)
	}
	buf.WriteString("    def __init__(self, base_url: str, requester: Requester):\n")
	if meta.Package != nil && len(meta.Package.AsyncInitPreCode) > 0 {
		for _, block := range meta.Package.AsyncInitPreCode {
			appendIndentedCode(&buf, block, 2)
		}
	}
	buf.WriteString("        self._base_url = remove_url_trailing_slash(base_url)\n")
	buf.WriteString("        self._requester = requester\n")
	if meta.Package != nil && len(meta.Package.AsyncInitCode) > 0 {
		if blankLineBeforeAsyncInitCode {
			buf.WriteString("\n")
		}
		for _, block := range meta.Package.AsyncInitCode {
			appendIndentedCode(&buf, block, 2)
			buf.WriteString("\n")
		}
	}
	if hasChildClients {
		if blankLineBeforeChildInits {
			buf.WriteString("\n")
		}
		for _, child := range childClientsForInit {
			attribute := normalizePythonIdentifier(child.Attribute)
			if child.DisableTypeHints {
				buf.WriteString(fmt.Sprintf("        self._%s = None\n", attribute))
			} else {
				buf.WriteString(fmt.Sprintf("        self._%s: Optional[%s] = None\n", attribute, child.AsyncClass))
			}
		}
		buf.WriteString("\n")
	} else if meta.Package == nil || len(meta.Package.AsyncInitCode) == 0 {
		buf.WriteString("\n")
	}
	asyncMethodBlocks := make([]classMethodBlock, 0)
	if hasChildClients {
		for _, child := range childClientsForAsync {
			attribute := normalizePythonIdentifier(child.Attribute)
			asyncMethodBlocks = append(asyncMethodBlocks, classMethodBlock{
				Name:    attribute,
				Content: renderChildClientProperty(meta, child, true, asyncClassKey, commentOverrides),
			})
		}
	}
	for _, binding := range bindings {
		if !mappingGeneratesAsync(binding.Mapping) {
			continue
		}
		asyncMethodBlocks = append(asyncMethodBlocks, classMethodBlock{
			Name:    binding.MethodName,
			Content: renderOperationMethodWithComments(doc, binding, true, "cozepy."+meta.ModulePath, asyncClass, commentOverrides),
		})
	}
	if meta.Package != nil && len(meta.Package.AsyncExtraMethods) > 0 {
		for _, block := range meta.Package.AsyncExtraMethods {
			content := indentCodeBlock(block, 1)
			content = applyMethodDocstringOverrides(content, asyncClassKey, commentOverrides)
			asyncMethodBlocks = append(asyncMethodBlocks, classMethodBlock{
				Name:    detectMethodBlockName(block),
				Content: content,
			})
		}
	}
	if meta.Package != nil && len(meta.Package.AsyncMethodOrder) > 0 {
		asyncMethodBlocks = orderClassMethodBlocks(asyncMethodBlocks, meta.Package.AsyncMethodOrder)
	}
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
	return content
}

func collectTypeImports(doc *openapi.Document, bindings []operationBinding) []string {
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

func paginationInheritResponse(mapping *config.OperationMapping) bool {
	if mapping == nil || mapping.PaginationInheritResponse == nil {
		return true
	}
	return *mapping.PaginationInheritResponse
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

func packageHasTokenPagination(bindings []operationBinding) bool {
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

func packageHasNumberPagination(bindings []operationBinding) bool {
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

func packageNeedsAnyDict(doc *openapi.Document, bindings []operationBinding, modelDefs []packageModelDefinition) (bool, bool) {
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

		if mapping == nil || (!mapping.UseKwargsHeaders && !mapping.DisableHeadersArg) {
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
			fieldType := pythonTypeForSchemaWithAliases(doc, propertySchema, required, nil)
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

func packageNeedsListResponseImport(bindings []operationBinding) bool {
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

func renderPagedResponseClasses(bindings []operationBinding, overriddenClasses map[string]struct{}) string {
	seen := map[string]struct{}{}
	ordered := make([]operationBinding, 0, len(bindings))
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
			if paginationInheritResponse(binding.Mapping) {
				buf.WriteString(fmt.Sprintf("class %s(CozeModel, TokenPagedResponse[%s]):\n", className, itemType))
			} else {
				buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", className))
			}
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
			if paginationInheritResponse(binding.Mapping) {
				buf.WriteString(fmt.Sprintf("class %s(CozeModel, NumberPagedResponse[%s]):\n", className, itemType))
			} else {
				buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", className))
			}
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
		if paginationInheritResponse(binding.Mapping) {
			buf.WriteString(fmt.Sprintf("class %s(CozeModel, NumberPagedResponse[%s]):\n", className, itemType))
		} else {
			buf.WriteString(fmt.Sprintf("class %s(CozeModel):\n", className))
		}
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
	SchemaName            string
	Name                  string
	BaseClasses           []string
	Schema                *openapi.Schema
	IsEnum                bool
	BeforeCode            []string
	PrependCode           []string
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
	SeparateRequired      *bool
	SeparateCommented     *bool
	BlankLineBeforeFields []string
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
		if modelName == "" {
			continue
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

		if schemaName == "" {
			if !model.AllowMissingInSwagger {
				continue
			}
			isEnum := len(enumValues) > 0
			result = append(result, packageModelDefinition{
				SchemaName:            schemaName,
				Name:                  modelName,
				BaseClasses:           append([]string(nil), model.BaseClasses...),
				Schema:                nil,
				IsEnum:                isEnum,
				BeforeCode:            append([]string(nil), model.BeforeCode...),
				PrependCode:           append([]string(nil), model.PrependCode...),
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
				SeparateRequired:      model.SeparateRequiredOptional,
				SeparateCommented:     model.SeparateCommentedFields,
				BlankLineBeforeFields: append([]string(nil), model.BlankLineBeforeFields...),
			})
			continue
		}
		schema, ok := doc.Components.Schemas[schemaName]
		if !ok || schema == nil {
			if !model.AllowMissingInSwagger {
				continue
			}
			isEnum := len(enumValues) > 0
			result = append(result, packageModelDefinition{
				SchemaName:            schemaName,
				Name:                  modelName,
				BaseClasses:           append([]string(nil), model.BaseClasses...),
				Schema:                nil,
				IsEnum:                isEnum,
				BeforeCode:            append([]string(nil), model.BeforeCode...),
				PrependCode:           append([]string(nil), model.PrependCode...),
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
				SeparateRequired:      model.SeparateRequiredOptional,
				SeparateCommented:     model.SeparateCommentedFields,
				BlankLineBeforeFields: append([]string(nil), model.BlankLineBeforeFields...),
			})
			continue
		}
		resolved := doc.ResolveSchema(schema)
		if resolved == nil {
			continue
		}
		isEnum := len(enumValues) > 0 || (len(resolved.Enum) > 0 && (resolved.Type == "string" || resolved.Type == "integer" || resolved.Type == ""))
		result = append(result, packageModelDefinition{
			SchemaName:            schemaName,
			Name:                  modelName,
			BaseClasses:           append([]string(nil), model.BaseClasses...),
			Schema:                resolved,
			IsEnum:                isEnum,
			BeforeCode:            append([]string(nil), model.BeforeCode...),
			PrependCode:           append([]string(nil), model.PrependCode...),
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
			SeparateRequired:      model.SeparateRequiredOptional,
			SeparateCommented:     model.SeparateCommentedFields,
			BlankLineBeforeFields: append([]string(nil), model.BlankLineBeforeFields...),
		})
	}
	return result
}

func renderPackageModelDefinitions(
	doc *openapi.Document,
	meta packageMeta,
	models []packageModelDefinition,
	schemaAliases map[string]string,
	commentOverrides config.CommentOverrides,
) string {
	var buf bytes.Buffer
	modulePrefix := "cozepy." + meta.ModulePath
	separateCommentedEnum := meta.Package != nil && meta.Package.SeparateCommentedEnum
	separateRequiredOptional := meta.Package != nil && meta.Package.SeparateRequiredOptional
	separateCommentedFields := meta.Package != nil && meta.Package.SeparateCommentedFields

	for _, model := range models {
		classKey := modulePrefix + "." + model.Name
		if len(model.BeforeCode) > 0 {
			for _, block := range model.BeforeCode {
				appendIndentedCode(&buf, block, 0)
				buf.WriteString("\n")
			}
		}
		modelSeparateCommentedEnum := separateCommentedEnum
		if model.SeparateCommentedEnum != nil {
			modelSeparateCommentedEnum = *model.SeparateCommentedEnum
		}
		modelSeparateRequiredOptional := separateRequiredOptional
		if model.SeparateRequired != nil {
			modelSeparateRequiredOptional = *model.SeparateRequired
		}
		modelSeparateCommentedFields := separateCommentedFields
		if model.SeparateCommented != nil {
			modelSeparateCommentedFields = *model.SeparateCommented
		}
		blankLineBeforeFieldSet := map[string]struct{}{}
		for _, rawFieldName := range model.BlankLineBeforeFields {
			fieldName := strings.TrimSpace(rawFieldName)
			if fieldName == "" {
				continue
			}
			blankLineBeforeFieldSet[fieldName] = struct{}{}
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
			if docstring := strings.TrimSpace(commentOverrides.ClassDocstrings[classKey]); docstring != "" && !modelHasCustomClassDocstring(model) {
				style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[classKey])
				writeClassDocstring(&buf, 1, docstring, style)
			}
			enumItems := make([]config.ModelEnumValue, 0)
			if len(model.EnumValues) > 0 {
				enumItems = append(enumItems, model.EnumValues...)
			} else if model.Schema != nil && len(model.Schema.Enum) > 0 {
				for _, enumValue := range model.Schema.Enum {
					enumItems = append(enumItems, config.ModelEnumValue{
						Name:  enumMemberName(fmt.Sprintf("%v", enumValue)),
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
					memberName = enumMemberName(fmt.Sprintf("%v", enumValue.Value))
				}
				inlineEnumComment := strings.TrimSpace(commentOverrides.InlineEnumMemberComment[classKey+"."+memberName])
				if inlineEnumComment != "" {
					inlineEnumComment = strings.TrimPrefix(inlineEnumComment, "#")
					inlineEnumComment = strings.TrimSpace(inlineEnumComment)
				}
				enumComment := linesFromCommentOverride(commentOverrides.EnumMemberComments[classKey+"."+memberName])
				if len(enumComment) > 0 && inlineEnumComment == "" {
					writeLineComments(&buf, 1, enumComment)
				}
				if inlineEnumComment != "" {
					buf.WriteString(fmt.Sprintf("    %s = %s  # %s\n", memberName, renderEnumValueLiteral(enumValue.Value), inlineEnumComment))
				} else {
					buf.WriteString(fmt.Sprintf("    %s = %s\n", memberName, renderEnumValueLiteral(enumValue.Value)))
				}
				if modelSeparateCommentedEnum && i < len(enumItems)-1 {
					nextHasComment := false
					nextName := strings.TrimSpace(enumItems[i+1].Name)
					if nextName == "" {
						nextName = enumMemberName(fmt.Sprintf("%v", enumItems[i+1].Value))
					}
					nextInlineComment := strings.TrimSpace(commentOverrides.InlineEnumMemberComment[classKey+"."+nextName])
					nextComment := linesFromCommentOverride(commentOverrides.EnumMemberComments[classKey+"."+nextName])
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
		if docstring := strings.TrimSpace(commentOverrides.ClassDocstrings[classKey]); docstring != "" && !modelHasCustomClassDocstring(model) {
			style := strings.TrimSpace(commentOverrides.ClassDocstringStyles[classKey])
			writeClassDocstring(&buf, 1, docstring, style)
		}
		properties := map[string]*openapi.Schema{}
		if model.Schema != nil {
			properties = model.Schema.Properties
		}
		if len(properties) == 0 {
			if len(model.PrependCode) == 0 && len(model.ExtraFields) == 0 && len(model.ExtraCode) == 0 {
				buf.WriteString("    pass\n\n")
				continue
			}
		}
		for _, block := range model.PrependCode {
			appendIndentedCode(&buf, block, 1)
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

		prevHasField := false
		prevRequired := false
		for _, fieldName := range fieldNames {
			if propertySchema, ok := properties[fieldName]; ok {
				currentRequired := requiredSet[fieldName]
				if _, ok := blankLineBeforeFieldSet[fieldName]; ok && prevHasField {
					buf.WriteString("\n")
				}
				if modelSeparateRequiredOptional && prevHasField && prevRequired && !currentRequired {
					buf.WriteString("\n")
				}
				typeName := modelFieldType(model, fieldName, pythonTypeForSchemaWithAliases(doc, propertySchema, requiredSet[fieldName], schemaAliases))
				normalizedFieldName := normalizePythonIdentifier(fieldName)
				inlineFieldComment := strings.TrimSpace(commentOverrides.InlineFieldComments[classKey+"."+normalizedFieldName])
				if inlineFieldComment != "" {
					inlineFieldComment = strings.TrimPrefix(inlineFieldComment, "#")
					inlineFieldComment = strings.TrimSpace(inlineFieldComment)
				}
				fieldComment := linesFromCommentOverride(commentOverrides.FieldComments[classKey+"."+normalizedFieldName])
				if modelSeparateCommentedFields && prevHasField && len(fieldComment) > 0 {
					buf.WriteString("\n")
				}
				if len(fieldComment) > 0 && inlineFieldComment == "" {
					writeLineComments(&buf, 1, fieldComment)
				}
				if requiredSet[fieldName] {
					if inlineFieldComment != "" {
						buf.WriteString(fmt.Sprintf("    %s: %s  # %s\n", normalizedFieldName, typeName, inlineFieldComment))
					} else {
						buf.WriteString(fmt.Sprintf("    %s: %s\n", normalizedFieldName, typeName))
					}
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
				}
				prevHasField = true
				prevRequired = currentRequired
				continue
			}

			extraField, ok := extraFieldByName[fieldName]
			if !ok {
				continue
			}
			if _, exists := properties[fieldName]; exists {
				continue
			}
			currentRequired := extraField.Required
			if _, ok := blankLineBeforeFieldSet[fieldName]; ok && prevHasField {
				buf.WriteString("\n")
			}
			if modelSeparateRequiredOptional && prevHasField && prevRequired && !currentRequired {
				buf.WriteString("\n")
			}
			normalizedFieldName := normalizePythonIdentifier(fieldName)
			typeName := strings.TrimSpace(extraField.Type)
			if typeName == "" {
				typeName = "Any"
			}
			inlineFieldComment := strings.TrimSpace(commentOverrides.InlineFieldComments[classKey+"."+normalizedFieldName])
			if inlineFieldComment != "" {
				inlineFieldComment = strings.TrimPrefix(inlineFieldComment, "#")
				inlineFieldComment = strings.TrimSpace(inlineFieldComment)
			}
			fieldComment := linesFromCommentOverride(commentOverrides.FieldComments[classKey+"."+normalizedFieldName])
			if modelSeparateCommentedFields && prevHasField && len(fieldComment) > 0 {
				buf.WriteString("\n")
			}
			if len(fieldComment) > 0 && inlineFieldComment == "" {
				writeLineComments(&buf, 1, fieldComment)
			}
			if extraField.Required {
				if inlineFieldComment != "" {
					buf.WriteString(fmt.Sprintf("    %s: %s  # %s\n", normalizedFieldName, typeName, inlineFieldComment))
				} else {
					buf.WriteString(fmt.Sprintf("    %s: %s\n", normalizedFieldName, typeName))
				}
				prevHasField = true
				prevRequired = currentRequired
				continue
			}
			defaultValue := strings.TrimSpace(extraField.Default)
			if defaultValue == "" {
				defaultValue = "None"
			}
			if defaultValue == "None" && !strings.HasPrefix(typeName, "Optional[") {
				typeName = "Optional[" + typeName + "]"
			}
			if inlineFieldComment != "" {
				buf.WriteString(fmt.Sprintf("    %s: %s = %s  # %s\n", normalizedFieldName, typeName, defaultValue, inlineFieldComment))
			} else {
				buf.WriteString(fmt.Sprintf("    %s: %s = %s\n", normalizedFieldName, typeName, defaultValue))
			}
			prevHasField = true
			prevRequired = currentRequired
		}
		if prevHasField && len(model.ExtraCode) > 0 {
			buf.WriteString("\n")
		}
		for _, block := range model.ExtraCode {
			appendIndentedCode(&buf, block, 1)
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
			buf.WriteString("    pass\n\n")
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

func appendIndentedCode(buf *bytes.Buffer, code string, indentLevel int) {
	block := strings.Trim(code, "\n")
	if strings.TrimSpace(block) == "" {
		return
	}
	lines := strings.Split(block, "\n")
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent < 0 {
		return
	}
	prefix := strings.Repeat("    ", indentLevel)
	for _, line := range lines {
		cleanLine := strings.TrimRight(line, "\r")
		if minIndent > 0 && len(cleanLine) >= minIndent {
			cleanLine = cleanLine[minIndent:]
		}
		if strings.TrimSpace(cleanLine) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(prefix)
		buf.WriteString(cleanLine)
		buf.WriteString("\n")
	}
}

func ensureTrailingNewlines(buf *bytes.Buffer, newlineCount int) {
	if newlineCount < 0 {
		newlineCount = 0
	}
	content := strings.TrimRight(buf.String(), "\n")
	buf.Reset()
	buf.WriteString(content)
	if newlineCount > 0 {
		buf.WriteString(strings.Repeat("\n", newlineCount))
	}
}

func linesFromCommentOverride(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			line = strings.TrimPrefix(line, "#")
		}
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func codeBlocksHaveLeadingDocstring(blocks []string) bool {
	for _, raw := range blocks {
		block := strings.TrimSpace(raw)
		if block == "" {
			continue
		}
		return strings.HasPrefix(block, "\"\"\"") || strings.HasPrefix(block, "'''")
	}
	return false
}

func modelHasCustomClassDocstring(model packageModelDefinition) bool {
	return codeBlocksHaveLeadingDocstring(model.PrependCode) || codeBlocksHaveLeadingDocstring(model.ExtraCode)
}

func writeLineComments(buf *bytes.Buffer, indentLevel int, lines []string) {
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	for _, line := range lines {
		buf.WriteString(indent)
		buf.WriteString("# ")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
}

func normalizedDocstringLines(docstring string) []string {
	text := strings.ReplaceAll(docstring, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	out := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		out = append(out, strings.TrimRight(line, "\r"))
	}
	return out
}

func writeClassDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	lines := normalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	if style == "inline" && len(lines) == 1 {
		buf.WriteString(indent)
		buf.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", lines[0]))
		buf.WriteString("\n")
		return
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
	buf.WriteString("\n")
}

func writeMethodDocstring(buf *bytes.Buffer, indentLevel int, docstring string, style string) {
	lines := normalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return
	}
	indent := strings.Repeat("    ", indentLevel)
	if style == "block" {
		buf.WriteString(indent)
		buf.WriteString("\"\"\"\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				buf.WriteString("\n")
				continue
			}
			buf.WriteString(indent)
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		buf.WriteString(indent)
		buf.WriteString("\"\"\"\n")
		return
	}
	if len(lines) == 1 {
		buf.WriteString(indent)
		buf.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", escapeDocstring(lines[0])))
		return
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"")
	buf.WriteString(lines[0])
	buf.WriteString("\n")
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			buf.WriteString("\n")
			continue
		}
		buf.WriteString(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(indent)
	buf.WriteString("\"\"\"\n")
}

func renderEnumValueLiteral(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", value))
	}
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

type classMethodBlock struct {
	Name    string
	Content string
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

func detectMethodBlockName(block string) string {
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if name, ok := parseDefName(trimmed); ok {
			return name
		}
	}
	return ""
}

func parseDefName(trimmedLine string) (string, bool) {
	defLine := strings.TrimSpace(trimmedLine)
	if strings.HasPrefix(defLine, "async def ") {
		defLine = strings.TrimSpace(strings.TrimPrefix(defLine, "async def "))
	} else if strings.HasPrefix(defLine, "def ") {
		defLine = strings.TrimSpace(strings.TrimPrefix(defLine, "def "))
	} else {
		return "", false
	}
	name := strings.TrimSpace(strings.SplitN(defLine, "(", 2)[0])
	if name == "" {
		return "", false
	}
	return name, true
}

func isDocstringLine(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "\"\"\"") || strings.HasPrefix(trimmedLine, "'''")
}

func renderMethodDocstringLines(docstring string, style string, indent string) []string {
	lines := normalizedDocstringLines(docstring)
	if len(lines) == 0 {
		return nil
	}
	out := make([]string, 0, len(lines)+3)
	if style == "block" {
		out = append(out, indent+"\"\"\"")
		for _, line := range lines {
			if strings.TrimSpace(line) == "" {
				out = append(out, "")
				continue
			}
			out = append(out, indent+line)
		}
		out = append(out, indent+"\"\"\"")
		return out
	}
	if len(lines) == 1 {
		out = append(out, indent+fmt.Sprintf("\"\"\"%s\"\"\"", escapeDocstring(lines[0])))
		return out
	}
	out = append(out, indent+"\"\"\""+lines[0])
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" {
			out = append(out, "")
			continue
		}
		out = append(out, indent+line)
	}
	out = append(out, indent+"\"\"\"")
	return out
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
		name, ok := parseDefName(strings.TrimSpace(line))
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
				if nextNonEmpty >= len(lines) || !isDocstringLine(strings.TrimSpace(lines[nextNonEmpty])) {
					indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
					docLines := renderMethodDocstringLines(
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

func orderClassMethodBlocks(blocks []classMethodBlock, orderedNames []string) []classMethodBlock {
	if len(blocks) == 0 || len(orderedNames) == 0 {
		return blocks
	}
	used := make([]bool, len(blocks))
	ordered := make([]classMethodBlock, 0, len(blocks))
	for _, rawName := range orderedNames {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		for i, block := range blocks {
			if used[i] {
				continue
			}
			if block.Name != name {
				continue
			}
			ordered = append(ordered, block)
			used[i] = true
			break
		}
	}
	for i, block := range blocks {
		if used[i] {
			continue
		}
		ordered = append(ordered, block)
	}
	return ordered
}

func orderChildClients(children []config.ChildClient, orderedAttrs []string) []config.ChildClient {
	if len(children) == 0 || len(orderedAttrs) == 0 {
		return children
	}
	used := make([]bool, len(children))
	ordered := make([]config.ChildClient, 0, len(children))
	for _, rawAttr := range orderedAttrs {
		attr := strings.TrimSpace(rawAttr)
		if attr == "" {
			continue
		}
		for i, child := range children {
			if used[i] {
				continue
			}
			if strings.TrimSpace(child.Attribute) != attr {
				continue
			}
			ordered = append(ordered, child)
			used[i] = true
			break
		}
	}
	for i, child := range children {
		if used[i] {
			continue
		}
		ordered = append(ordered, child)
	}
	return ordered
}

func indentCodeBlock(block string, level int) string {
	var buf bytes.Buffer
	appendIndentedCode(&buf, block, level)
	return buf.String()
}

func renderChildClientProperty(
	meta packageMeta,
	child config.ChildClient,
	async bool,
	classKey string,
	commentOverrides config.CommentOverrides,
) string {
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
	if child.MultilineSignature {
		if child.DisableTypeHints {
			buf.WriteString(fmt.Sprintf("    def %s(\n", attribute))
			buf.WriteString("        self,\n")
			buf.WriteString("    ):\n")
		} else {
			buf.WriteString(fmt.Sprintf("    def %s(\n", attribute))
			buf.WriteString("        self,\n")
			buf.WriteString(fmt.Sprintf("    ) -> \"%s\":\n", typeName))
		}
	} else {
		if child.DisableTypeHints {
			buf.WriteString(fmt.Sprintf("    def %s(self):\n", attribute))
		} else {
			buf.WriteString(fmt.Sprintf("    def %s(self) -> \"%s\":\n", attribute, typeName))
		}
	}
	methodKey := strings.TrimSpace(classKey) + "." + attribute
	if docstring, ok := commentOverrides.MethodDocstrings[methodKey]; ok {
		docstring = strings.TrimSpace(docstring)
		if docstring != "" {
			style := strings.TrimSpace(commentOverrides.MethodDocstringStyles[methodKey])
			writeMethodDocstring(&buf, 2, docstring, style)
		}
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
	ValueExpr    string
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
			fields = append(fields, renderQueryField{
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
		argName := operationArgName(param.Name, paramAliases)
		typeName := typeOverride(param.Name, param.Required, pythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		valueExpr := argName
		if mapping != nil && len(mapping.QueryFieldValues) > 0 {
			if override, ok := mapping.QueryFieldValues[param.Name]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
		}
		fields = append(fields, renderQueryField{
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

func orderSignatureQueryFields(fields []renderQueryField, orderedRawNames []string) []renderQueryField {
	if len(fields) == 0 || len(orderedRawNames) == 0 {
		return fields
	}
	fieldByName := make(map[string]renderQueryField, len(fields))
	for _, field := range fields {
		fieldByName[field.RawName] = field
	}
	result := make([]renderQueryField, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, rawName := range orderedRawNames {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		field, ok := fieldByName[name]
		if !ok {
			continue
		}
		result = append(result, field)
		seen[name] = struct{}{}
	}
	for _, field := range fields {
		if _, ok := seen[field.RawName]; ok {
			continue
		}
		result = append(result, field)
	}
	return result
}

func signatureArgName(argDecl string) string {
	trimmed := strings.TrimSpace(argDecl)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "**") {
		name := strings.TrimSpace(strings.TrimPrefix(trimmed, "**"))
		name = strings.TrimSpace(strings.SplitN(name, ":", 2)[0])
		name = strings.TrimSpace(strings.SplitN(name, "=", 2)[0])
		return name
	}
	name := strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[0])
	name = strings.TrimSpace(strings.SplitN(name, "=", 2)[0])
	return name
}

func isKwargsSignatureArg(argDecl string) bool {
	return strings.HasPrefix(strings.TrimSpace(argDecl), "**")
}

func orderSignatureArgs(signatureArgs []string, orderedNames []string) []string {
	if len(signatureArgs) == 0 || len(orderedNames) == 0 {
		return signatureArgs
	}
	argByName := make(map[string]string, len(signatureArgs))
	argOrder := make([]string, 0, len(signatureArgs))
	for _, argDecl := range signatureArgs {
		name := signatureArgName(argDecl)
		if name == "" {
			continue
		}
		if _, exists := argByName[name]; exists {
			continue
		}
		argByName[name] = argDecl
		argOrder = append(argOrder, name)
	}
	result := make([]string, 0, len(signatureArgs))
	seen := map[string]struct{}{}
	for _, rawName := range orderedNames {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		argDecl, ok := argByName[name]
		if !ok {
			continue
		}
		result = append(result, argDecl)
		seen[name] = struct{}{}
	}
	for _, name := range argOrder {
		if _, ok := seen[name]; ok {
			continue
		}
		result = append(result, argByName[name])
	}
	return result
}

func normalizeSignatureArgs(signatureArgs []string) []string {
	if len(signatureArgs) <= 1 {
		return signatureArgs
	}
	normal := make([]string, 0, len(signatureArgs))
	kwargs := make([]string, 0, 1)
	for _, argDecl := range signatureArgs {
		if isKwargsSignatureArg(argDecl) {
			kwargs = append(kwargs, argDecl)
			continue
		}
		normal = append(normal, argDecl)
	}
	return append(normal, kwargs...)
}

func orderedUniqueByPriority(values []string, priority []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := make(map[string]struct{}, len(values))
	ordered := make([]string, 0, len(values))
	for _, name := range priority {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		for _, value := range values {
			if strings.TrimSpace(value) != trimmed {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				break
			}
			ordered = append(ordered, trimmed)
			seen[trimmed] = struct{}{}
			break
		}
	}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		ordered = append(ordered, trimmed)
		seen[trimmed] = struct{}{}
	}
	return ordered
}

func operationArgDefault(mapping *config.OperationMapping, rawName string, argName string, async bool) (string, bool) {
	if mapping == nil {
		return "", false
	}
	defaultMaps := make([]map[string]string, 0, 3)
	if async && len(mapping.ArgDefaultsAsync) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaultsAsync)
	}
	if !async && len(mapping.ArgDefaultsSync) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaultsSync)
	}
	if len(mapping.ArgDefaults) > 0 {
		defaultMaps = append(defaultMaps, mapping.ArgDefaults)
	}
	if len(defaultMaps) == 0 {
		return "", false
	}
	if argName = strings.TrimSpace(argName); argName != "" {
		for _, defaults := range defaultMaps {
			if value, ok := defaults[argName]; ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value), true
			}
		}
	}
	if rawName = strings.TrimSpace(rawName); rawName != "" {
		for _, defaults := range defaultMaps {
			if value, ok := defaults[rawName]; ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value), true
			}
		}
	}
	return "", false
}

func buildDelegateCallArgs(signatureArgs []string, mapping *config.OperationMapping, async bool) []string {
	if mapping != nil {
		explicitArgs := mapping.DelegateCallArgs
		if async && len(mapping.AsyncDelegateCallArgs) > 0 {
			explicitArgs = mapping.AsyncDelegateCallArgs
		}
		if len(explicitArgs) > 0 {
			args := make([]string, 0, len(explicitArgs))
			for _, arg := range explicitArgs {
				trimmed := strings.TrimSpace(arg)
				if trimmed == "" {
					continue
				}
				args = append(args, trimmed)
			}
			return args
		}
	}
	args := make([]string, 0, len(signatureArgs))
	for _, argDecl := range signatureArgs {
		trimmed := strings.TrimSpace(argDecl)
		if trimmed == "" {
			continue
		}
		if trimmed == "*" {
			continue
		}
		name := signatureArgName(trimmed)
		if name == "" {
			continue
		}
		if name == "self" {
			continue
		}
		if isKwargsSignatureArg(trimmed) {
			args = append(args, fmt.Sprintf("**%s", name))
			continue
		}
		args = append(args, fmt.Sprintf("%s=%s", name, name))
	}
	return args
}

func renderDelegatedCall(buf *bytes.Buffer, target string, args []string, async bool, asyncYield bool) {
	target = strings.TrimSpace(target)
	if target == "" {
		return
	}
	qualifier := fmt.Sprintf("self.%s", target)
	if len(args) == 0 {
		if async && asyncYield {
			buf.WriteString(fmt.Sprintf("        async for item in await %s():\n", qualifier))
			buf.WriteString("            yield item\n")
			return
		}
		if async {
			buf.WriteString(fmt.Sprintf("        return await %s()\n", qualifier))
			return
		}
		buf.WriteString(fmt.Sprintf("        return %s()\n", qualifier))
		return
	}

	if async && asyncYield {
		buf.WriteString(fmt.Sprintf("        async for item in await %s(\n", qualifier))
		for _, arg := range args {
			trimmed := strings.TrimSpace(arg)
			if trimmed == "" {
				continue
			}
			buf.WriteString(fmt.Sprintf("            %s,\n", trimmed))
		}
		buf.WriteString("        ):\n")
		buf.WriteString("            yield item\n")
		return
	}
	if async {
		buf.WriteString(fmt.Sprintf("        return await %s(\n", qualifier))
		for _, arg := range args {
			buf.WriteString(fmt.Sprintf("            %s,\n", arg))
		}
		buf.WriteString("        )\n")
		return
	}
	buf.WriteString(fmt.Sprintf("        return %s(\n", qualifier))
	for _, arg := range args {
		buf.WriteString(fmt.Sprintf("            %s,\n", arg))
	}
	buf.WriteString("        )\n")
}

func renderOperationMethod(doc *openapi.Document, binding operationBinding, async bool) string {
	return renderOperationMethodWithComments(doc, binding, async, "", "", config.CommentOverrides{})
}

func renderOperationMethodWithComments(
	doc *openapi.Document,
	binding operationBinding,
	async bool,
	modulePath string,
	className string,
	commentOverrides config.CommentOverrides,
) string {
	details := binding.Details
	requestMethod := strings.ToLower(strings.TrimSpace(details.Method))
	paginationMode := ""
	returnType, returnCast := returnTypeInfo(doc, details.ResponseSchema)
	requestBodyType, bodyRequired := requestBodyTypeInfo(doc, details.RequestBodySchema, details.RequestBody)
	useKwargsHeaders := binding.Mapping != nil && binding.Mapping.UseKwargsHeaders
	disableHeadersArg := binding.Mapping != nil && binding.Mapping.DisableHeadersArg
	ignoreHeaderParams := binding.Mapping != nil && binding.Mapping.IgnoreHeaderParams
	castKeyword := binding.Mapping != nil && binding.Mapping.CastKeyword
	streamKeyword := binding.Mapping != nil && binding.Mapping.StreamKeyword
	streamWrap := binding.Mapping != nil && binding.Mapping.StreamWrap
	headersBeforeBody := binding.Mapping != nil && binding.Mapping.HeadersBeforeBody
	omitReturnType := binding.Mapping != nil && binding.Mapping.OmitReturnType
	asyncIncludeKwargs := async && binding.Mapping != nil && binding.Mapping.AsyncIncludeKwargs
	paginationHeadersBeforeParams := binding.Mapping != nil && binding.Mapping.PaginationHeadersBeforeParams
	paginationCastBeforeHeaders := binding.Mapping != nil && binding.Mapping.PaginationCastBeforeHeaders
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
		signatureQueryFields = orderSignatureQueryFields(queryFields, binding.Mapping.SignatureQueryFields)
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
		name := operationArgName(param.Name, paramAliases)
		pathParamNameMap[param.Name] = name
		typeName := typeOverride(param.Name, true, pythonTypeForSchema(doc, param.Schema, true), argTypes)
		signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", name, typeName))
		signatureArgNames[name] = struct{}{}
	}
	for _, field := range signatureQueryFields {
		defaultValue := strings.TrimSpace(field.DefaultValue)
		if override, ok := operationArgDefault(binding.Mapping, field.RawName, field.ArgName, async); ok {
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
		name := operationArgName(param.Name, paramAliases)
		typeName := typeOverride(param.Name, param.Required, pythonTypeForSchema(doc, param.Schema, param.Required), argTypes)
		if defaultValue, ok := operationArgDefault(binding.Mapping, param.Name, name, async); ok {
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
			argName := operationArgName(bodyField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := bodyFieldSchema(doc, details.RequestBodySchema, bodyField)
			required := bodyRequiredSet[bodyField]
			typeName := typeOverride(bodyField, required, pythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := operationArgDefault(binding.Mapping, bodyField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	} else if requestBodyType != "" {
		if defaultValue, ok := operationArgDefault(binding.Mapping, "body", "body", async); ok {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = %s", requestBodyType, defaultValue))
		} else if bodyRequired {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: %s", requestBodyType))
		} else {
			signatureArgs = append(signatureArgs, fmt.Sprintf("body: Optional[%s] = None", requestBodyType))
		}
	}
	if len(filesFieldNames) > 0 {
		for _, filesField := range filesFieldNames {
			argName := operationArgName(filesField, paramAliases)
			if _, exists := signatureArgNames[argName]; exists {
				continue
			}
			fieldSchema := bodyFieldSchema(doc, details.RequestBodySchema, filesField)
			required := bodyRequiredSet[filesField]
			typeName := typeOverride(filesField, required, pythonTypeForSchema(doc, fieldSchema, required), argTypes)
			if defaultValue, ok := operationArgDefault(binding.Mapping, filesField, argName, async); ok {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = %s", argName, typeName, defaultValue))
			} else if required {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s", argName, typeName))
			} else {
				signatureArgs = append(signatureArgs, fmt.Sprintf("%s: %s = None", argName, typeName))
			}
			signatureArgNames[argName] = struct{}{}
		}
	}
	includeKwargsHeaders := useKwargsHeaders && !disableHeadersArg
	includeExplicitHeadersArg := !disableHeadersArg && !includeKwargsHeaders
	if includeKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	} else if includeExplicitHeadersArg {
		signatureArgs = append(signatureArgs, "headers: Optional[Dict[str, str]] = None")
	}
	if asyncIncludeKwargs && !includeKwargsHeaders {
		signatureArgs = append(signatureArgs, "**kwargs")
	}
	if binding.Mapping != nil && len(binding.Mapping.SignatureArgs) > 0 {
		signatureArgs = orderSignatureArgs(signatureArgs, binding.Mapping.SignatureArgs)
	}
	signatureArgs = normalizeSignatureArgs(signatureArgs)

	methodKeyword := "def"
	requestCall := "self._requester.request"
	if async {
		methodKeyword = "async def"
		requestCall = "await self._requester.arequest"
	}
	headersAssigned := false

	var buf bytes.Buffer
	returnAnnotation := ""
	if !omitReturnType {
		returnAnnotation = fmt.Sprintf(" -> %s", returnType)
	}
	compactSignature := len(bodyFieldNames) == 0 && requestBodyType == "" && len(signatureArgs) <= 2
	if binding.Mapping != nil {
		if binding.Mapping.ForceMultilineSignature {
			compactSignature = false
		}
		if binding.Mapping.ForceCompactSignature {
			compactSignature = true
		}
		if async {
			if binding.Mapping.ForceMultilineSignatureAsync {
				compactSignature = false
			}
			if binding.Mapping.ForceCompactSignatureAsync {
				compactSignature = true
			}
		} else {
			if binding.Mapping.ForceMultilineSignatureSync {
				compactSignature = false
			}
			if binding.Mapping.ForceCompactSignatureSync {
				compactSignature = true
			}
		}
	}
	if compactSignature {
		if len(signatureArgs) == 0 {
			buf.WriteString(fmt.Sprintf("    %s %s(self)%s:\n", methodKeyword, binding.MethodName, returnAnnotation))
		} else if len(signatureArgs) == 1 && isKwargsSignatureArg(signatureArgs[0]) {
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
	overrideDocstring := ""
	overrideDocstringStyle := ""
	overrideDocstringExists := false
	if modulePath != "" && className != "" {
		key := strings.TrimSpace(modulePath) + "." + strings.TrimSpace(className) + "." + binding.MethodName
		rawDocstring, ok := commentOverrides.MethodDocstrings[key]
		if ok {
			overrideDocstringExists = true
			overrideDocstring = strings.TrimSpace(rawDocstring)
		}
		overrideDocstringStyle = strings.TrimSpace(commentOverrides.MethodDocstringStyles[key])
	}
	if binding.Mapping != nil && len(binding.Mapping.PreDocstringCode) > 0 {
		for _, block := range binding.Mapping.PreDocstringCode {
			appendIndentedCode(&buf, block, 2)
		}
	}
	if overrideDocstringExists {
		if overrideDocstring == "" {
			// Explicitly disabled via comment overrides.
		} else {
			writeMethodDocstring(&buf, 2, overrideDocstring, overrideDocstringStyle)
		}
	} else {
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
		summary = strings.TrimSpace(summary)
		if summary != "" {
			buf.WriteString(fmt.Sprintf("        \"\"\"%s\"\"\"\n", escapeDocstring(summary)))
		}
	}
	if delegateTo != "" {
		callArgs := buildDelegateCallArgs(signatureArgs, binding.Mapping, async)
		renderDelegatedCall(&buf, delegateTo, callArgs, async, delegateAsyncYield)
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
		if disableHeadersArg {
			buf.WriteString("        header_values = dict()\n")
		} else {
			buf.WriteString("        header_values = dict(headers or {})\n")
		}
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
	if headersExpr != "" {
		buf.WriteString(fmt.Sprintf("        headers = %s\n", headersExpr))
		headersAssigned = true
		if isTokenPagination(paginationMode) || isNumberPagination(paginationMode) {
			buf.WriteString("\n")
		}
	}
	if (isTokenPagination(paginationMode) || isNumberPagination(paginationMode)) && binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			appendIndentedCode(&buf, block, 2)
		}
		if headersExpr == "" {
			buf.WriteString("\n")
		}
	}
	includePaginationHeaders := headersExpr != "" || !disableHeadersArg || len(details.HeaderParameters) > 0
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
		ensureTrailingNewlines(&buf, 2)
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
			if paginationHeadersBeforeParams && includePaginationHeaders {
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
				if paginationCastBeforeHeaders && !paginationHeadersBeforeParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if paginationHeadersBeforeParams {
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
			if paginationHeadersBeforeParams && includePaginationHeaders {
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
				if paginationCastBeforeHeaders && !paginationHeadersBeforeParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if paginationHeadersBeforeParams {
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
		ensureTrailingNewlines(&buf, 2)
		if async {
			buf.WriteString("        async def request_maker(i_page_num: int, i_page_size: int) -> HTTPRequest:\n")
			buf.WriteString("            return await self._requester.amake_request(\n")
			buf.WriteString(fmt.Sprintf("                %q,\n", paginationRequestMethod))
			buf.WriteString("                url,\n")
			if paginationHeadersBeforeParams && includePaginationHeaders {
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
				if paginationCastBeforeHeaders && !paginationHeadersBeforeParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if paginationHeadersBeforeParams {
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
			if paginationHeadersBeforeParams && includePaginationHeaders {
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
				if paginationCastBeforeHeaders && !paginationHeadersBeforeParams {
					buf.WriteString(fmt.Sprintf("                cast=%s,\n", dataClass))
					buf.WriteString("                headers=headers,\n")
				} else if paginationHeadersBeforeParams {
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

	if headersExpr == "" && headersBeforeBody && includeKwargsHeaders && !isTokenPagination(paginationMode) && !isNumberPagination(paginationMode) && len(details.HeaderParameters) == 0 {
		if binding.Mapping != nil && binding.Mapping.BlankLineAfterHeaders {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n\n")
		} else {
			buf.WriteString("        headers: Optional[dict] = kwargs.get(\"headers\")\n")
		}
		headersAssigned = true
	}
	if binding.Mapping != nil && len(binding.Mapping.PreBodyCode) > 0 {
		for _, block := range binding.Mapping.PreBodyCode {
			appendIndentedCode(&buf, block, 2)
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
			argName := operationArgName(fieldName, paramAliases)
			valueExpr := argName
			if override, ok := filesFieldValues[fieldName]; ok && strings.TrimSpace(override) != "" {
				valueExpr = strings.TrimSpace(override)
			}
			buf.WriteString(fmt.Sprintf("        files = {%q: %s}\n", fieldName, valueExpr))
			return true
		}

		buf.WriteString("        files = {\n")
		for _, filesField := range filesFieldNames {
			argName := operationArgName(filesField, paramAliases)
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
				argName := operationArgName(fieldName, paramAliases)
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
				argName := operationArgName(bodyField, paramAliases)
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
				argName := operationArgName(bodyField, paramAliases)
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
	castExprValue := castExpr
	if castKeyword {
		castExprValue = fmt.Sprintf("cast=%s", castExpr)
	}
	optionalArgs = append(optionalArgs, requestCallArg{Key: "cast", Expr: castExprValue, Pos: !castKeyword})
	if len(queryFields) > 0 {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "params", Expr: "params=params"})
	}
	hasHeadersArg := headersExpr != "" || !disableHeadersArg || len(details.HeaderParameters) > 0
	if headersBeforeBody && hasHeadersArg {
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
	if !headersBeforeBody && hasHeadersArg {
		optionalArgs = append(optionalArgs, requestCallArg{Key: "headers", Expr: "headers=headers"})
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
		ensureTrailingNewlines(&buf, 2)
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
		if len(aliases) > 0 {
			return "", false
		}
		return normalizeClassName(name), true
	}
	resolved := doc.ResolveSchema(schema)
	if resolved != nil && resolved != schema {
		if name, ok := doc.SchemaName(resolved); ok {
			if alias, exists := aliases[name]; exists && strings.TrimSpace(alias) != "" {
				return alias, true
			}
			if len(aliases) > 0 {
				return "", false
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
	raw := strings.TrimSpace(value)
	privateMethod := strings.HasPrefix(raw, "_")
	name := normalizePythonIdentifier(toSnake(raw))
	if name == "" {
		if privateMethod {
			return "_call"
		}
		return "call"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "method_" + name
	}
	if privateMethod && !strings.HasPrefix(name, "_") {
		name = "_" + name
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

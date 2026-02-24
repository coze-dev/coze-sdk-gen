package generator

import (
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

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{written: map[string]struct{}{}}
	if err := writePythonSDK(cfg.OutputSDK, writer); err != nil {
		return Result{}, err
	}

	return Result{GeneratedFiles: writer.count, GeneratedOps: len(bindings)}, nil
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

func writePythonSDK(outputDir string, writer *fileWriter) error {
	rootDir := filepath.Join(outputDir, "cozepy")
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return fmt.Errorf("create python package root %q: %w", rootDir, err)
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
	if err := writer.write(filepath.Join(rootDir, "py.typed"), ""); err != nil {
		return err
	}

	if err := writePythonStaticAssets(outputDir, writer); err != nil {
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

func writePythonStaticAssets(outputDir string, writer *fileWriter) error {
	staticFiles, err := listPythonStaticFiles()
	if err != nil {
		return err
	}
	for _, relPath := range staticFiles {
		content, err := readPythonStaticFile(relPath)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(outputDir, filepath.FromSlash(relPath))
		if _, exists := writer.written[filepath.Clean(targetPath)]; exists {
			continue
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

func normalizePackageName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "default"
	}
	return normalizePythonIdentifier(toSnake(name))
}

func defaultMethodName(operationID string, path string, method string) string {
	base := strings.TrimSpace(operationID)
	if base != "" {
		lower := strings.ToLower(base)
		for _, prefix := range []string{"openapi", "open_api", "coze"} {
			if strings.HasPrefix(lower, prefix) {
				base = base[len(prefix):]
				break
			}
		}
		if strings.TrimSpace(base) != "" {
			return normalizeMethodName(base)
		}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	fallback := strings.ToLower(method)
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if part == "" || strings.HasPrefix(part, "{") {
			continue
		}
		fallback = part
		break
	}
	return normalizeMethodName(fallback)
}

func normalizeMethodName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "call"
	}
	return normalizePythonIdentifier(toSnake(value))
}

func normalizePythonIdentifier(value string) string {
	if value == "" {
		return "field"
	}
	value = strings.ToLower(value)
	if value[0] >= '0' && value[0] <= '9' {
		value = "_" + value
	}
	keywords := map[string]struct{}{
		"false": {}, "none": {}, "true": {},
		"and": {}, "as": {}, "assert": {}, "async": {}, "await": {},
		"break": {}, "class": {}, "continue": {}, "def": {}, "del": {},
		"elif": {}, "else": {}, "except": {}, "finally": {}, "for": {},
		"from": {}, "global": {}, "if": {}, "import": {}, "in": {},
		"is": {}, "lambda": {}, "nonlocal": {}, "not": {}, "or": {},
		"pass": {}, "raise": {}, "return": {}, "try": {}, "while": {},
		"with": {}, "yield": {},
	}
	if _, ok := keywords[value]; ok {
		return value + "_"
	}
	return value
}

func toSnake(value string) string {
	parts := splitIdentifier(value)
	if len(parts) == 0 {
		return ""
	}
	for i := range parts {
		parts[i] = strings.ToLower(parts[i])
	}
	return collapseUnderscore(strings.Join(parts, "_"))
}

func splitIdentifier(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	normalized := make([]rune, 0, len(value))
	for _, r := range value {
		switch {
		case r == '_' || r == '-' || r == '/' || r == ' ' || r == '.':
			normalized = append(normalized, '_')
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			normalized = append(normalized, r)
		default:
			normalized = append(normalized, '_')
		}
	}

	segments := strings.Split(string(normalized), "_")
	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		start := 0
		runes := []rune(segment)
		for i := 1; i < len(runes); i++ {
			if unicode.IsUpper(runes[i]) && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
				parts = append(parts, string(runes[start:i]))
				start = i
			}
		}
		parts = append(parts, string(runes[start:]))
	}
	return parts
}

func collapseUnderscore(value string) string {
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	value = strings.Trim(value, "_")
	if value == "" {
		return "field"
	}
	return value
}

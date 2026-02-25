package gogen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/generator/fsutil"
	pygen "github.com/coze-dev/coze-sdk-gen/internal/generator/python"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type goOperationBinding struct {
	MethodName string
	Path       string
	Method     string
	Summary    string
}

type Result struct {
	GeneratedFiles int
	GeneratedOps   int
}

type fileWriter struct {
	count   int
	written map[string]struct{}
}

func GenerateGo(cfg *config.Config, doc *openapi.Document) (Result, error) {
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

	bindings := buildGoOperationBindings(cfg, doc)
	if len(bindings) == 0 {
		return Result{}, fmt.Errorf("no operations selected for generation")
	}

	if err := fsutil.CleanOutputDirPreserveEntries(cfg.OutputSDK, cfg.DiffIgnorePathsForLanguage("go")); err != nil {
		return Result{}, fmt.Errorf("prepare output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{
		written: map[string]struct{}{},
	}
	if err := writeGoRuntimeScaffolding(cfg.OutputSDK, writer); err != nil {
		return Result{}, err
	}
	if err := writeGoAPIModules(cfg, doc, writer); err != nil {
		return Result{}, err
	}
	if err := writeGoExtraAssets(cfg.OutputSDK, writer); err != nil {
		return Result{}, err
	}

	return Result{
		GeneratedFiles: writer.count,
		GeneratedOps:   len(bindings),
	}, nil
}

func buildGoOperationBindings(cfg *config.Config, doc *openapi.Document) []goOperationBinding {
	allOps := doc.ListOperationDetails()
	bindings := make([]goOperationBinding, 0, len(allOps))
	seen := map[string]int{}

	for _, details := range allOps {
		if cfg.IsIgnored(details.Path, details.Method) {
			continue
		}
		base := strings.TrimSpace(details.OperationID)
		if base == "" {
			base = pygen.DefaultMethodName(details.OperationID, details.Path, details.Method)
		}
		name := normalizeGoExportedIdentifier(base)
		if name == "" {
			name = "Operation"
		}

		count := seen[name]
		seen[name] = count + 1
		if count > 0 {
			name = fmt.Sprintf("%s%d", name, count+1)
		}

		bindings = append(bindings, goOperationBinding{
			MethodName: name,
			Path:       details.Path,
			Method:     strings.ToUpper(details.Method),
			Summary:    goOperationSummary(details),
		})
	}
	return bindings
}

func goOperationSummary(details openapi.OperationDetails) string {
	summary := strings.TrimSpace(details.Summary)
	if summary == "" {
		summary = strings.TrimSpace(details.Description)
	}
	return oneLineText(summary)
}

func writeGoRuntimeScaffolding(outputDir string, writer *fileWriter) error {
	textAssets := map[string]string{
		".gitignore":      "gitignore.tpl",
		"codecov.yml":     "codecov.yml.tpl",
		"go.mod":          "go.mod.tpl",
		"go.sum":          "go.sum.tpl",
		"LICENSE":         "LICENSE.tpl",
		"CONTRIBUTING.md": "CONTRIBUTING.md.tpl",
	}
	for target, asset := range textAssets {
		content, err := renderGoRuntimeAsset(asset)
		if err != nil {
			return err
		}
		if target == ".gitignore" || target == "codecov.yml" || target == "LICENSE" {
			content = strings.TrimSuffix(content, "\n")
		}
		if err := writer.write(filepath.Join(outputDir, target), content); err != nil {
			return err
		}
	}

	goAssets := map[string]string{
		"audio.go":                         "audio.go.tpl",
		"auth.go":                          "auth.go.tpl",
		"auth_token.go":                    "auth_token.go.tpl",
		"base_model.go":                    "base_model.go.tpl",
		"client.go":                        "client.go.tpl",
		"common.go":                        "common.go.tpl",
		"const.go":                         "const.go.tpl",
		"error.go":                         "error.go.tpl",
		"enterprises.go":                   "enterprises.go.tpl",
		"logger.go":                        "logger.go.tpl",
		"pagination.go":                    "pagination.go.tpl",
		"request.go":                       "request.go.tpl",
		"stores.go":                        "stores.go.tpl",
		"stream_reader.go":                 "stream_reader.go.tpl",
		"user_agent.go":                    "user_agent.go.tpl",
		"utils.go":                         "utils.go.tpl",
		"websocket.go":                     "websocket.go.tpl",
		"websocket_audio.go":               "websocket_audio.go.tpl",
		"websocket_audio_speech_client.go": "websocket_audio_speech_client.go.tpl",
		"websocket_audio_speech.go":        "websocket_audio_speech.go.tpl",
		"websocket_audio_transcription_client.go": "websocket_audio_transcription_client.go.tpl",
		"websocket_audio_transcription.go":        "websocket_audio_transcription.go.tpl",
		"websocket_chat_client.go":                "websocket_chat_client.go.tpl",
		"websocket_chat.go":                       "websocket_chat.go.tpl",
		"websocket_client.go":                     "websocket_client.go.tpl",
		"websocket_event.go":                      "websocket_event.go.tpl",
		"websocket_event_type.go":                 "websocket_event_type.go.tpl",
		"websocket_wait.go":                       "websocket_wait.go.tpl",
	}
	for target, asset := range goAssets {
		content, err := renderGoRuntimeAsset(asset)
		if err != nil {
			return err
		}
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("format go runtime file %q: %w", target, err)
		}
		if err := writer.write(filepath.Join(outputDir, target), string(formatted)); err != nil {
			return err
		}
	}
	return nil
}

func writeGoExtraAssets(outputDir string, writer *fileWriter) error {
	assets, err := listGoExtraAssets()
	if err != nil {
		return err
	}
	for _, rel := range assets {
		if shouldSkipGoExtraAsset(rel) {
			continue
		}
		content, err := renderGoExtraAsset(rel)
		if err != nil {
			return err
		}
		target := rel
		if strings.HasSuffix(target, ".tpl") {
			target = strings.TrimSuffix(target, ".tpl")
		}
		if err := writer.writeBytes(filepath.Join(outputDir, target), content); err != nil {
			return err
		}
	}
	return nil
}

func writeGoAPIModules(cfg *config.Config, doc *openapi.Document, writer *fileWriter) error {
	for _, renderer := range listGoAPIModuleRenderers() {
		content, err := renderer.Render(cfg, doc)
		if err != nil {
			return err
		}
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("format generated go api module %q: %w", renderer.FileName, err)
		}
		if err := writer.write(filepath.Join(cfg.OutputSDK, renderer.FileName), string(formatted)); err != nil {
			return err
		}
	}
	return nil
}

func findGoOperationPath(cfg *config.Config, doc *openapi.Document, sdkMethod string, method string, fallback string) (string, error) {
	if cfg != nil {
		for _, mapping := range cfg.API.OperationMappings {
			if !strings.EqualFold(strings.TrimSpace(mapping.Method), strings.TrimSpace(method)) {
				continue
			}
			for _, m := range mapping.SDKMethods {
				if strings.TrimSpace(m) == strings.TrimSpace(sdkMethod) {
					return strings.TrimSpace(mapping.Path), nil
				}
			}
		}
	}
	if doc != nil && doc.HasOperation(method, fallback) {
		return fallback, nil
	}
	return "", fmt.Errorf(
		"resolve go operation path failed for sdk_method=%q method=%q fallback=%q",
		strings.TrimSpace(sdkMethod),
		strings.ToUpper(strings.TrimSpace(method)),
		strings.TrimSpace(fallback),
	)
}

func convertCurlyPathToColon(path string) string {
	converted := path
	for {
		start := strings.Index(converted, "{")
		if start < 0 {
			break
		}
		endOffset := strings.Index(converted[start:], "}")
		if endOffset <= 1 {
			break
		}
		end := start + endOffset
		param := strings.TrimSpace(converted[start+1 : end])
		if param == "" {
			break
		}
		converted = converted[:start] + ":" + param + converted[end+1:]
	}
	return converted
}

func normalizeGoExportedIdentifier(value string) string {
	parts := splitIdentifierWords(value)
	if len(parts) == 0 {
		return ""
	}
	var buf strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		first := unicode.ToUpper(runes[0])
		buf.WriteRune(first)
		if len(runes) > 1 {
			buf.WriteString(string(runes[1:]))
		}
	}
	result := buf.String()
	if result == "" {
		return ""
	}
	firstRune := []rune(result)[0]
	if unicode.IsDigit(firstRune) {
		return "Op" + result
	}
	return result
}

func splitIdentifierWords(value string) []string {
	if value == "" {
		return nil
	}
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	words := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		words = append(words, field)
	}
	return words
}

func oneLineText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func (w *fileWriter) write(path string, content string) error {
	return w.writeBytes(path, []byte(content))
}

func (w *fileWriter) writeBytes(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %q: %w", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write file %q: %w", path, err)
	}
	w.count++
	w.written[path] = struct{}{}
	return nil
}

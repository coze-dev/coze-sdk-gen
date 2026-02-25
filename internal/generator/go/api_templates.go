package gogen

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type goAPIModuleRenderer struct {
	FileName string
	Render   func(cfg *config.Config, doc *openapi.Document) (string, error)
}

type goSwaggerModuleSpec struct {
	FileName        string
	PackageName     string
	TypeName        string
	ConstructorName string
	CoreFieldName   string
	Children        []goSwaggerModuleChild
}

type goSwaggerModuleChild struct {
	FieldName       string
	TypeName        string
	ConstructorName string
}

type goSwaggerOperationBinding struct {
	MethodName string
	HTTPMethod string
	Path       string
	Summary    string
	IsFile     bool
	Priority   int
}

var goInlineAPIModuleRenderers = []goAPIModuleRenderer{
	{FileName: "apps.go", Render: renderGoAppsModule},
	{FileName: "audio_live.go", Render: renderGoAudioLiveModule},
	{FileName: "audio_speech.go", Render: renderGoAudioSpeechModule},
	{FileName: "audio_transcription.go", Render: renderGoAudioTranscriptionsModule},
	{FileName: "chats_messages.go", Render: renderGoChatsMessagesModule},
	{FileName: "files.go", Render: renderGoFilesModule},
	{FileName: "templates.go", Render: renderGoTemplatesModule},
	{FileName: "users.go", Render: renderGoUsersModule},
	{FileName: "workflows_chat.go", Render: renderGoWorkflowsChatModule},
}

var goSwaggerAPIModuleSpecs = []goSwaggerModuleSpec{}

var goGeneratedAPIModuleFiles = buildGoGeneratedAPIModuleFiles()

func buildGoGeneratedAPIModuleFiles() map[string]struct{} {
	files := make(map[string]struct{}, len(goInlineAPIModuleRenderers)+len(goSwaggerAPIModuleSpecs))
	for _, renderer := range goInlineAPIModuleRenderers {
		files[renderer.FileName] = struct{}{}
	}
	for _, spec := range goSwaggerAPIModuleSpecs {
		files[spec.FileName] = struct{}{}
	}
	return files
}

func listGoAPIModuleRenderers() []goAPIModuleRenderer {
	renderers := make([]goAPIModuleRenderer, 0, len(goInlineAPIModuleRenderers)+len(goSwaggerAPIModuleSpecs))
	renderers = append(renderers, goInlineAPIModuleRenderers...)

	specs := append([]goSwaggerModuleSpec(nil), goSwaggerAPIModuleSpecs...)
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].FileName < specs[j].FileName
	})
	for _, spec := range specs {
		specCopy := spec
		renderers = append(renderers, goAPIModuleRenderer{
			FileName: specCopy.FileName,
			Render: func(cfg *config.Config, doc *openapi.Document) (string, error) {
				bindings := buildGoSwaggerOperationBindings(cfg, doc, specCopy.PackageName)
				return renderGoSwaggerModule(specCopy, bindings), nil
			},
		})
	}
	return renderers
}

func shouldSkipGoExtraAsset(rel string) bool {
	target := strings.TrimSuffix(rel, ".tpl")
	if target == "README.md" || target == "main" || strings.HasPrefix(target, ".github/") {
		return true
	}
	_, ok := goGeneratedAPIModuleFiles[target]
	return ok
}

func buildGoSwaggerOperationBindings(cfg *config.Config, doc *openapi.Document, packageName string) []goSwaggerOperationBinding {
	if cfg == nil {
		return nil
	}
	packageName = strings.TrimSpace(packageName)
	if packageName == "" {
		return nil
	}

	bindings := make([]goSwaggerOperationBinding, 0)
	for _, mapping := range cfg.API.OperationMappings {
		if cfg.IsIgnored(mapping.Path, mapping.Method) {
			continue
		}
		details, hasDetails := resolveGoOperationDetails(doc, mapping)
		for methodIndex, sdkMethod := range mapping.SDKMethods {
			pkg, method, ok := parseGoSDKMethod(cfg, mapping.Path, sdkMethod)
			if !ok || pkg != packageName {
				continue
			}
			goMethod := normalizeGoExportedIdentifier(method)
			if goMethod == "" {
				continue
			}
			httpMethod := strings.TrimSpace(mapping.HTTPMethodOverride)
			if httpMethod == "" {
				httpMethod = strings.TrimSpace(mapping.Method)
			}
			if httpMethod == "" {
				httpMethod = http.MethodGet
			}

			isFile := false
			summary := ""
			if hasDetails {
				summary = goOperationSummary(details)
				contentType := strings.ToLower(strings.TrimSpace(details.RequestBodyContentType))
				isFile = strings.Contains(contentType, "multipart/form-data")
			}

			priority := len(bindings)
			if mapping.Priority > 0 {
				priority = mapping.Priority + methodIndex
			}
			bindings = append(bindings, goSwaggerOperationBinding{
				MethodName: goMethod,
				HTTPMethod: strings.ToUpper(httpMethod),
				Path:       strings.TrimSpace(mapping.Path),
				Summary:    summary,
				IsFile:     isFile,
				Priority:   priority,
			})
		}
	}

	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].Priority == bindings[j].Priority {
			if bindings[i].MethodName == bindings[j].MethodName {
				if bindings[i].HTTPMethod == bindings[j].HTTPMethod {
					return bindings[i].Path < bindings[j].Path
				}
				return bindings[i].HTTPMethod < bindings[j].HTTPMethod
			}
			return bindings[i].MethodName < bindings[j].MethodName
		}
		return bindings[i].Priority < bindings[j].Priority
	})

	seen := map[string]int{}
	for i := range bindings {
		name := bindings[i].MethodName
		count := seen[name]
		if count > 0 {
			bindings[i].MethodName = fmt.Sprintf("%s%d", name, count+1)
		}
		seen[name] = count + 1
	}
	return bindings
}

func resolveGoOperationDetails(doc *openapi.Document, mapping config.OperationMapping) (openapi.OperationDetails, bool) {
	if doc == nil {
		return openapi.OperationDetails{}, false
	}
	details, ok := doc.OperationDetails(strings.TrimSpace(mapping.Path), strings.TrimSpace(mapping.Method))
	if !ok || details == nil {
		return openapi.OperationDetails{}, false
	}
	return *details, true
}

func parseGoSDKMethod(cfg *config.Config, path string, sdkMethod string) (string, string, bool) {
	sdkMethod = strings.TrimSpace(sdkMethod)
	if sdkMethod == "" {
		return "", "", false
	}
	parts := strings.Split(sdkMethod, ".")
	switch len(parts) {
	case 1:
		method := strings.TrimSpace(parts[0])
		if method == "" || cfg == nil {
			return "", "", false
		}
		pkg, ok := cfg.ResolvePackage(path, "")
		if !ok {
			return "", "", false
		}
		return strings.TrimSpace(pkg.Name), method, true
	case 2:
		pkg := strings.TrimSpace(parts[0])
		method := strings.TrimSpace(parts[1])
		if pkg == "" || method == "" {
			return "", "", false
		}
		return pkg, method, true
	case 3:
		if strings.TrimSpace(parts[0]) != "go" {
			return "", "", false
		}
		pkg := strings.TrimSpace(parts[1])
		method := strings.TrimSpace(parts[2])
		if pkg == "" || method == "" {
			return "", "", false
		}
		return pkg, method, true
	default:
		return "", "", false
	}
}

func renderGoSwaggerModule(spec goSwaggerModuleSpec, bindings []goSwaggerOperationBinding) string {
	coreField := strings.TrimSpace(spec.CoreFieldName)
	if coreField == "" {
		coreField = "core"
	}

	var buf bytes.Buffer
	buf.WriteString("package coze\n\n")
	if len(bindings) > 0 {
		buf.WriteString("import (\n")
		buf.WriteString("\t\"context\"\n")
		buf.WriteString("\t\"net/http\"\n")
		buf.WriteString(")\n\n")
	}

	for _, binding := range bindings {
		if summary := strings.TrimSpace(binding.Summary); summary != "" {
			buf.WriteString(fmt.Sprintf("// %s %s\n", binding.MethodName, summary))
		}
		buf.WriteString(fmt.Sprintf("func (r *%s) %s(ctx context.Context, req *SwaggerOperationRequest) (*SwaggerOperationResponse, error) {\n", spec.TypeName, binding.MethodName))
		buf.WriteString("\tif req == nil {\n")
		buf.WriteString("\t\treq = &SwaggerOperationRequest{}\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\trequest := &RawRequestReq{\n")
		buf.WriteString(fmt.Sprintf("\t\tMethod: %s,\n", goHTTPMethodConstant(binding.HTTPMethod)))
		buf.WriteString(fmt.Sprintf("\t\tURL:    buildSwaggerOperationURL(%q, req.PathParams, req.QueryParams),\n", binding.Path))
		buf.WriteString("\t\tBody:   req.Body,\n")
		if binding.IsFile {
			buf.WriteString("\t\tIsFile: true,\n")
		}
		buf.WriteString("\t}\n")
		buf.WriteString("\tresponse := new(SwaggerOperationResponse)\n")
		buf.WriteString(fmt.Sprintf("\terr := r.%s.rawRequest(ctx, request, response)\n", coreField))
		buf.WriteString("\treturn response, err\n")
		buf.WriteString("}\n\n")
	}

	buf.WriteString(fmt.Sprintf("type %s struct {\n", spec.TypeName))
	buf.WriteString(fmt.Sprintf("\t%s *core\n", coreField))
	for _, child := range spec.Children {
		buf.WriteString(fmt.Sprintf("\t%s *%s\n", child.FieldName, child.TypeName))
	}
	buf.WriteString("}\n\n")

	buf.WriteString(fmt.Sprintf("func %s(core *core) *%s {\n", spec.ConstructorName, spec.TypeName))
	if len(spec.Children) == 0 {
		buf.WriteString(fmt.Sprintf("\treturn &%s{%s: core}\n", spec.TypeName, coreField))
		buf.WriteString("}\n")
		return buf.String()
	}

	buf.WriteString(fmt.Sprintf("\treturn &%s{\n", spec.TypeName))
	buf.WriteString(fmt.Sprintf("\t\t%s: core,\n", coreField))
	for _, child := range spec.Children {
		buf.WriteString(fmt.Sprintf("\t\t%s: %s(core),\n", child.FieldName, child.ConstructorName))
	}
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
	return buf.String()
}

func goHTTPMethodConstant(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "http.MethodGet"
	case http.MethodPost:
		return "http.MethodPost"
	case http.MethodPut:
		return "http.MethodPut"
	case http.MethodDelete:
		return "http.MethodDelete"
	case http.MethodPatch:
		return "http.MethodPatch"
	case http.MethodHead:
		return "http.MethodHead"
	case http.MethodOptions:
		return "http.MethodOptions"
	case http.MethodTrace:
		return "http.MethodTrace"
	default:
		return fmt.Sprintf("%q", strings.ToUpper(strings.TrimSpace(method)))
	}
}

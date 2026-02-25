package gogen

import (
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestRenderGoInlineTemplateModuleUsesMappedPath(t *testing.T) {
	module := mustInlineTemplate(t, "apps.go")
	cfg := &config.Config{
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v9/custom/apps",
					Method:     "get",
					SDKMethods: []string{"apps.list"},
				},
			},
		},
	}

	content, err := renderGoInlineTemplateModule(cfg, nil, module)
	if err != nil {
		t.Fatalf("renderGoInlineTemplateModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v9/custom/apps\",") {
		t.Fatalf("expected mapped path in rendered content, got:\n%s", content)
	}
}

func TestRenderGoInlineTemplateModuleFallsBackToSwaggerPath(t *testing.T) {
	module := mustInlineTemplate(t, "users.go")
	doc := mustParseOpenAPIDoc(t, `
openapi: 3.0.0
paths:
  /v1/users/me:
    get:
      responses:
        '200':
          description: ok
`)

	content, err := renderGoInlineTemplateModule(&config.Config{}, doc, module)
	if err != nil {
		t.Fatalf("renderGoInlineTemplateModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v1/users/me\",") {
		t.Fatalf("expected swagger fallback path in rendered content, got:\n%s", content)
	}
}

func TestRenderGoInlineTemplateModuleConvertsCurlyPath(t *testing.T) {
	module := mustInlineTemplate(t, "templates.go")
	cfg := &config.Config{
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v2/templates/{template_id}/duplicate",
					Method:     "post",
					SDKMethods: []string{"templates.duplicate"},
				},
			},
		},
	}

	content, err := renderGoInlineTemplateModule(cfg, nil, module)
	if err != nil {
		t.Fatalf("renderGoInlineTemplateModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v2/templates/:template_id/duplicate\",") {
		t.Fatalf("expected curly path to be converted, got:\n%s", content)
	}
}

func mustInlineTemplate(t *testing.T, fileName string) goInlineAPIModuleTemplate {
	t.Helper()
	for _, module := range goInlineAPIModuleTemplates {
		if module.FileName == fileName {
			return module
		}
	}
	t.Fatalf("inline template %q not found", fileName)
	return goInlineAPIModuleTemplate{}
}

func mustParseOpenAPIDoc(t *testing.T, content string) *openapi.Document {
	t.Helper()
	doc, err := openapi.Parse([]byte(content))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}
	return doc
}

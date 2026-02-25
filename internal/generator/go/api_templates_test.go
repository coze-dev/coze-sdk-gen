package gogen

import (
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestListGoAPIModuleRenderersUsesSwaggerSpecs(t *testing.T) {
	renderers := listGoAPIModuleRenderers()
	if len(renderers) != len(goSwaggerAPIModuleSpecs) {
		t.Fatalf("expected %d swagger renderers, got %d", len(goSwaggerAPIModuleSpecs), len(renderers))
	}

	expected := make(map[string]struct{}, len(goSwaggerAPIModuleSpecs))
	for _, spec := range goSwaggerAPIModuleSpecs {
		expected[spec.FileName] = struct{}{}
	}
	for _, renderer := range renderers {
		if _, ok := expected[renderer.FileName]; !ok {
			t.Fatalf("unexpected renderer file %q", renderer.FileName)
		}
	}
}

func TestBuildGoSwaggerOperationBindingsSupportsGoPrefixedMethods(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v1/files/upload",
					Method:     "post",
					Order:      10,
					SDKMethods: []string{"go.files.upload"},
				},
				{
					Path:       "/v1/files/retrieve",
					Method:     "get",
					Order:      20,
					SDKMethods: []string{"go.files.retrieve"},
				},
			},
		},
	}
	doc := mustParseOpenAPIDoc(t, `
openapi: 3.0.0
paths:
  /v1/files/upload:
    post:
      summary: upload file
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
      responses:
        '200':
          description: ok
  /v1/files/retrieve:
    get:
      summary: retrieve file
      responses:
        '200':
          description: ok
`)

	bindings := buildGoSwaggerOperationBindings(cfg, doc, "files")
	if len(bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(bindings))
	}
	if bindings[0].MethodName != "Upload" || !bindings[0].IsFile || bindings[0].Path != "/v1/files/upload" {
		t.Fatalf("unexpected first binding: %+v", bindings[0])
	}
	if bindings[1].MethodName != "Retrieve" || bindings[1].IsFile || bindings[1].Path != "/v1/files/retrieve" {
		t.Fatalf("unexpected second binding: %+v", bindings[1])
	}
}

func TestRenderGoSwaggerModuleRendersGenericOperationMethod(t *testing.T) {
	content := renderGoSwaggerModule(goSwaggerModuleSpec{
		FileName:        "apps.go",
		PackageName:     "apps",
		TypeName:        "apps",
		ConstructorName: "newApps",
	}, []goSwaggerOperationBinding{
		{
			MethodName: "List",
			HTTPMethod: "GET",
			Path:       "/v1/apps",
			Summary:    "List apps",
		},
	})

	if !strings.Contains(content, "func (r *apps) List(ctx context.Context, req *SwaggerOperationRequest) (*SwaggerOperationResponse, error)") {
		t.Fatalf("expected generic swagger method signature, got:\n%s", content)
	}
	if !strings.Contains(content, "URL:    buildSwaggerOperationURL(\"/v1/apps\", req.PathParams, req.QueryParams),") {
		t.Fatalf("expected swagger URL helper usage, got:\n%s", content)
	}
	if !strings.Contains(content, "func newApps(core *core) *apps {") {
		t.Fatalf("expected constructor in rendered module, got:\n%s", content)
	}
}

func mustParseOpenAPIDoc(t *testing.T, content string) *openapi.Document {
	t.Helper()
	doc, err := openapi.Parse([]byte(content))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}
	return doc
}

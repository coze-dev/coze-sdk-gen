package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestLoadConfigAndValidate(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "generator.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Language != "python" {
		t.Fatalf("unexpected language: %q", cfg.Language)
	}
	if cfg.OutputSDK == "" {
		t.Fatalf("output should not be empty")
	}
	if len(cfg.API.OperationMappings) != 3 {
		t.Fatalf("unexpected operation mappings count: %d", len(cfg.API.OperationMappings))
	}
}

func TestParseInvalidConfig(t *testing.T) {
	_, err := Parse([]byte("language: go\noutput_sdk: b"))
	if err == nil {
		t.Fatal("expected Parse() to fail with unsupported language")
	}
	if !strings.Contains(err.Error(), "only python is supported") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAgainstSwagger(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "generator.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	doc, err := openapi.Load(filepath.Join("..", "openapi", "testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}

	report := cfg.ValidateAgainstSwagger(doc)
	if report.HasErrors() {
		t.Fatalf("expected no validation errors, got: %s", report.Error())
	}
}

func TestValidateAgainstSwaggerHasErrors(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "generator.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.API.OperationMappings = append(cfg.API.OperationMappings, OperationMapping{
		Path:       "/v1/not-exist",
		Method:     "post",
		SDKMethods: []string{"demo.method"},
	})
	cfg.API.Packages[0].PathPrefixes = append(cfg.API.Packages[0].PathPrefixes, "/v2/none")

	doc, err := openapi.Load(filepath.Join("..", "openapi", "testdata", "swagger_fragment.yaml"))
	if err != nil {
		t.Fatalf("openapi.Load() error = %v", err)
	}

	report := cfg.ValidateAgainstSwagger(doc)
	if !report.HasErrors() {
		t.Fatal("expected validation report to have errors")
	}
	if !strings.Contains(report.Error(), "missing operations") {
		t.Fatalf("unexpected report error: %s", report.Error())
	}
	if !strings.Contains(report.Error(), "path prefixes") {
		t.Fatalf("unexpected report error: %s", report.Error())
	}
}

func TestValidateAgainstNilSwagger(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "generator.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	report := cfg.ValidateAgainstSwagger(nil)
	if !report.HasErrors() {
		t.Fatal("expected report to include missing operations for nil swagger")
	}
	if len(report.MissingOperations) == 0 {
		t.Fatal("expected missing operations")
	}
}

func TestDefaultsApplied(t *testing.T) {
	cfg, err := Parse([]byte("language: python\noutput_sdk: out\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.API.FieldAliases == nil {
		t.Fatal("expected field aliases map to be initialized")
	}
}

func TestValidateConfigFailures(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name: "duplicate package",
			content: `
language: python
output_sdk: out
api:
  packages:
    - name: chat
      source_dir: a
    - name: chat
      source_dir: b
`,
		},
		{
			name: "invalid package prefix",
			content: `
language: python
output_sdk: out
api:
  packages:
    - name: chat
      source_dir: a
      path_prefixes:
        - v3/chat
`,
		},
		{
			name: "missing sdk methods",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v3/chat
      method: post
`,
		},
		{
			name: "invalid method",
			content: `
language: python
output_sdk: out
api:
  ignore_operations:
    - path: /v3/chat
      method: bad
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Parse([]byte(tc.content)); err == nil {
				t.Fatalf("expected Parse() to fail for case %s", tc.name)
			}
		})
	}
}

func TestValidateAgainstSwaggerAllowMissing(t *testing.T) {
	cfg := &Config{
		Language:  "python",
		OutputSDK: "out",
		API: APIConfig{
			Packages: []Package{
				{
					Name:                  "users",
					SourceDir:             "cozepy/users",
					PathPrefixes:          []string{"/v1/users"},
					AllowMissingInSwagger: true,
				},
			},
			OperationMappings: []OperationMapping{
				{
					Path:                  "/v1/users/me",
					Method:                "get",
					SDKMethods:            []string{"users.me"},
					AllowMissingInSwagger: true,
				},
			},
		},
	}
	doc, err := openapi.Parse([]byte("paths:\n  /v3/chat:\n    post: {operationId: OpenApiChat}\n"))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}

	report := cfg.ValidateAgainstSwagger(doc)
	if report.HasErrors() {
		t.Fatalf("expected no errors when allow_missing_in_swagger is set, got %s", report.Error())
	}
}

func TestConfigHelpers(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "generator.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.IsIgnored("/v1/workspaces/{workspace_id}", "get") {
		t.Fatal("expected operation to be ignored")
	}
	if cfg.IsIgnored("/v1/workspaces/{workspace_id}", "post") {
		t.Fatal("did not expect post to be ignored")
	}

	mappings := cfg.FindOperationMappings("/v3/chat", "POST")
	if len(mappings) != 1 {
		t.Fatalf("expected one mapping, got %d", len(mappings))
	}

	pkg, ok := cfg.ResolvePackage("/v3/chat/message/list", "")
	if !ok || pkg.Name != "chat" {
		t.Fatalf("unexpected package resolution: %+v, ok=%v", pkg, ok)
	}

	pkg, ok = cfg.ResolvePackage("/v3/chat/message/list", "workflows")
	if !ok || pkg.Name != "workflows" {
		t.Fatalf("expected preferred package workflows, got %+v", pkg)
	}

	if _, ok := cfg.ResolvePackage("/v2/not-exists", ""); ok {
		t.Fatal("did not expect package for unknown path")
	}
}

func TestParseSDKMethod(t *testing.T) {
	pkg, method, ok := ParseSDKMethod("chat.stream")
	if !ok || pkg != "chat" || method != "stream" {
		t.Fatalf("unexpected parsed sdk method: pkg=%q method=%q ok=%v", pkg, method, ok)
	}

	pkg, method, ok = ParseSDKMethod("create")
	if !ok || pkg != "" || method != "create" {
		t.Fatalf("unexpected parsed sdk method without package: pkg=%q method=%q ok=%v", pkg, method, ok)
	}

	if _, _, ok := ParseSDKMethod("a.b.c"); ok {
		t.Fatal("expected invalid sdk method with more than one dot")
	}
	if _, _, ok := ParseSDKMethod(".create"); ok {
		t.Fatal("expected invalid sdk method")
	}
}

package config

import (
	"os"
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
	if len(cfg.API.OperationMappings) != 3 {
		t.Fatalf("unexpected operation mappings count: %d", len(cfg.API.OperationMappings))
	}
}

func TestValidateRuntimeLanguage(t *testing.T) {
	cfg, err := Parse([]byte("api: {}\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	cfg.Language = "ruby"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected Validate() to fail with unsupported language")
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
	cfg, err := Parse([]byte("api: {}\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.API.FieldAliases == nil {
		t.Fatal("expected field aliases map to be initialized")
	}
	if cfg.CommentOverrides.ClassDocstrings == nil {
		t.Fatal("expected comment overrides maps to be initialized")
	}
	if cfg.CommentOverrides.ClassDocstringStyles == nil {
		t.Fatal("expected class docstring style map to be initialized")
	}
	if cfg.CommentOverrides.InlineFieldComments == nil {
		t.Fatal("expected inline field comments map to be initialized")
	}
	if cfg.CommentOverrides.InlineEnumMemberComment == nil {
		t.Fatal("expected inline enum comments map to be initialized")
	}
}

func TestParseIgnoresRuntimeOptionsInYAML(t *testing.T) {
	cfg, err := Parse([]byte("language: go\noutput_sdk: out\napi: {}\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.Language != "" {
		t.Fatalf("expected language to be ignored from yaml, got %q", cfg.Language)
	}
	if cfg.OutputSDK != "" {
		t.Fatalf("expected output_sdk to be ignored from yaml, got %q", cfg.OutputSDK)
	}
}

func TestLoadCommentOverrides(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "generator.yaml")
	commentsPath := filepath.Join(dir, "comments.yaml")
	configYAML := `
language: python
output_sdk: out
comment_overrides_file: comments.yaml
api:
  packages:
    - name: chat
      source_dir: cozepy/chat
`
	commentsYAML := `
class_docstrings:
  cozepy.chat.Chat: Chat doc
class_docstring_styles:
  cozepy.chat.Chat: block
method_docstrings:
  cozepy.chat.ChatClient.create: Create chat
method_docstring_styles:
  cozepy.chat.ChatClient.create: block
field_comments:
  cozepy.chat.Chat.id:
    - chat id
inline_field_comments:
  cozepy.chat.Chat.name: inline field
enum_member_comments:
  cozepy.chat.ChatStatus.CREATED:
    - created status
inline_enum_member_comments:
  cozepy.chat.ChatStatus.UNKNOWN: unknown status
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config error: %v", err)
	}
	if err := os.WriteFile(commentsPath, []byte(commentsYAML), 0o644); err != nil {
		t.Fatalf("write comments error: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := cfg.CommentOverrides.ClassDocstrings["cozepy.chat.Chat"]; got != "Chat doc" {
		t.Fatalf("unexpected class docstring override: %q", got)
	}
	if got := cfg.CommentOverrides.ClassDocstringStyles["cozepy.chat.Chat"]; got != "block" {
		t.Fatalf("unexpected class docstring style override: %q", got)
	}
	if got := cfg.CommentOverrides.MethodDocstrings["cozepy.chat.ChatClient.create"]; got != "Create chat" {
		t.Fatalf("unexpected method docstring override: %q", got)
	}
	if got := cfg.CommentOverrides.MethodDocstringStyles["cozepy.chat.ChatClient.create"]; got != "block" {
		t.Fatalf("unexpected method docstring style override: %q", got)
	}
	if got := cfg.CommentOverrides.FieldComments["cozepy.chat.Chat.id"]; len(got) != 1 || got[0] != "chat id" {
		t.Fatalf("unexpected field comments override: %#v", got)
	}
	if got := cfg.CommentOverrides.EnumMemberComments["cozepy.chat.ChatStatus.CREATED"]; len(got) != 1 || got[0] != "created status" {
		t.Fatalf("unexpected enum comments override: %#v", got)
	}
	if got := cfg.CommentOverrides.InlineFieldComments["cozepy.chat.Chat.name"]; got != "inline field" {
		t.Fatalf("unexpected inline field comment override: %q", got)
	}
	if got := cfg.CommentOverrides.InlineEnumMemberComment["cozepy.chat.ChatStatus.UNKNOWN"]; got != "unknown status" {
		t.Fatalf("unexpected inline enum comment override: %q", got)
	}
}

func TestLoadCommentOverridesMissingFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "generator.yaml")
	configYAML := `
language: python
output_sdk: out
comment_overrides_file: missing-comments.yaml
api:
  packages:
    - name: chat
      source_dir: cozepy/chat
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0o644); err != nil {
		t.Fatalf("write config error: %v", err)
	}
	if _, err := Load(configPath); err == nil {
		t.Fatal("expected Load() to fail for missing comment overrides file")
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
			name: "empty sync method order",
			content: `
language: python
output_sdk: out
api:
  packages:
    - name: chat
      source_dir: a
      sync_method_order:
        - ""
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
		{
			name: "both sync and async only",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v3/chat
      method: post
      sdk_methods:
        - chat.create
      sync_only: true
      async_only: true
`,
		},
		{
			name: "delegate args without target",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v3/chat
      method: post
      sdk_methods:
        - chat.create
      delegate_call_args:
        - bot_id=bot_id
`,
		},
		{
			name: "async delegate args without target",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v3/chat
      method: post
      sdk_methods:
        - chat.stream
      async_delegate_call_args:
        - bot_id=bot_id
`,
		},
		{
			name: "conflicting pagination order flags",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v1/apps
      method: get
      sdk_methods:
        - apps.list
      pagination: number
      pagination_data_class: _PrivateListAppsData
      pagination_item_type: App
      pagination_headers_before_params: true
      pagination_cast_before_headers: true
`,
		},
		{
			name: "empty pre body code",
			content: `
language: python
output_sdk: out
api:
  operation_mappings:
    - path: /v1/bot/publish
      method: post
      sdk_methods:
        - bots.publish
      pre_body_code:
        - ""
`,
		},
		{
			name: "empty model base class",
			content: `
language: python
output_sdk: out
api:
  packages:
    - name: datasets
      source_dir: a
      model_schemas:
        - name: DemoModel
          allow_missing_in_swagger: true
          base_classes:
            - ""
`,
		},
		{
			name: "empty async child order",
			content: `
language: python
output_sdk: out
api:
  packages:
    - name: audio
      source_dir: a
      async_child_order:
        - ""
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

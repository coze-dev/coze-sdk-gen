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
	if cfg.SourceSDK == "" || cfg.OutputSDK == "" {
		t.Fatalf("source/output should not be empty")
	}
	if len(cfg.API.OperationMappings) != 3 {
		t.Fatalf("unexpected operation mappings count: %d", len(cfg.API.OperationMappings))
	}
}

func TestParseInvalidConfig(t *testing.T) {
	_, err := Parse([]byte("language: go\nsource_sdk: a\noutput_sdk: b"))
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
	cfg, err := Parse([]byte("language: python\nsource_sdk: src\noutput_sdk: out\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(cfg.Copy.Include) == 0 {
		t.Fatal("expected default copy.include to be set")
	}
	if cfg.API.FieldAliases == nil {
		t.Fatal("expected field aliases map to be initialized")
	}
}

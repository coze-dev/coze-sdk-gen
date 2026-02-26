package python

import (
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestShouldGenerateImplicitRequestBody(t *testing.T) {
	t.Run("nil schema", func(t *testing.T) {
		if shouldGenerateImplicitRequestBody("post", nil, nil) {
			t.Fatal("expected false for nil schema")
		}
	})

	t.Run("empty object schema", func(t *testing.T) {
		schema := &openapi.Schema{Type: "object"}
		if shouldGenerateImplicitRequestBody("post", nil, schema) {
			t.Fatal("expected false for empty object schema")
		}
	})

	t.Run("schema with properties", func(t *testing.T) {
		schema := &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
		}
		if !shouldGenerateImplicitRequestBody("post", nil, schema) {
			t.Fatal("expected true when schema has properties")
		}
	})

	t.Run("primitive schema", func(t *testing.T) {
		schema := &openapi.Schema{Type: "string"}
		if !shouldGenerateImplicitRequestBody("post", nil, schema) {
			t.Fatal("expected true for primitive request body schema")
		}
	})

	t.Run("mapping body fields suppresses implicit body", func(t *testing.T) {
		schema := &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
		}
		mapping := &config.OperationMapping{
			BodyFields: []string{"name"},
		}
		if shouldGenerateImplicitRequestBody("post", mapping, schema) {
			t.Fatal("expected false when body_fields is configured")
		}
	})

	t.Run("mapping files fields suppresses implicit body", func(t *testing.T) {
		schema := &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"file": {Type: "string"},
			},
		}
		mapping := &config.OperationMapping{
			FilesFields: []string{"file"},
		}
		if shouldGenerateImplicitRequestBody("post", mapping, schema) {
			t.Fatal("expected false when files_fields is configured")
		}
	})

	t.Run("safe methods suppress implicit body", func(t *testing.T) {
		schema := &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
		}
		if shouldGenerateImplicitRequestBody("get", nil, schema) {
			t.Fatal("expected false for get method")
		}
	})

	t.Run("query-only mapping still allows implicit body on post", func(t *testing.T) {
		schema := &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"ignored": {Type: "string"},
			},
		}
		mapping := &config.OperationMapping{
			Method: "post",
			QueryFields: []config.OperationField{
				{Name: "cursor", Type: "str", Required: false},
			},
		}
		if !shouldGenerateImplicitRequestBody("post", mapping, schema) {
			t.Fatal("expected true for post query-only mapping with non-empty schema")
		}
	})
}

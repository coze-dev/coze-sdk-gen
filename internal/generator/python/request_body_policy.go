package python

import (
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func shouldGenerateImplicitRequestBody(defaultMethod string, mapping *config.OperationMapping, requestBodySchema *openapi.Schema) bool {
	if requestBodySchema == nil {
		return false
	}
	requestMethod := strings.ToLower(strings.TrimSpace(defaultMethod))
	if mapping != nil {
		requestMethod = strings.ToLower(strings.TrimSpace(mapping.Method))
		if override := strings.TrimSpace(mapping.HTTPMethodOverride); override != "" {
			requestMethod = strings.ToLower(override)
		}
	}
	if isSafeWithoutBodyMethod(requestMethod) {
		return false
	}
	if mapping != nil && len(mapping.QueryFields) > 0 {
		return false
	}
	if hasExplicitPayloadConfig(mapping) {
		return false
	}
	return !isEffectivelyEmptyObjectSchema(requestBodySchema)
}

func hasExplicitPayloadConfig(mapping *config.OperationMapping) bool {
	if mapping == nil {
		return false
	}
	if len(mapping.BodyFields) > 0 || len(mapping.BodyFixedValues) > 0 {
		return true
	}
	if len(mapping.FilesFields) > 0 {
		return true
	}
	return len(mapping.BodyFieldValues) > 0
}

func isEffectivelyEmptyObjectSchema(schema *openapi.Schema) bool {
	if schema == nil {
		return false
	}
	if len(schema.Properties) > 0 || len(schema.Required) > 0 {
		return false
	}
	if schema.Items != nil || len(schema.AllOf) > 0 || len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		return false
	}
	if len(schema.Enum) > 0 {
		return false
	}
	if schema.AdditionalProperties != nil {
		if allow, ok := schema.AdditionalProperties.(bool); !ok || allow {
			return false
		}
	}
	schemaType := strings.ToLower(strings.TrimSpace(schema.Type))
	return schemaType == "" || schemaType == "object"
}

func isSafeWithoutBodyMethod(method string) bool {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "get", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}

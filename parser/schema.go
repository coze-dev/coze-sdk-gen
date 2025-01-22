package parser

// SchemaType represents the type of a schema
type SchemaType string

const (
	SchemaTypeString  SchemaType = "string"
	SchemaTypeInteger SchemaType = "integer"
	SchemaTypeNumber  SchemaType = "number"
	SchemaTypeBoolean SchemaType = "boolean"
	SchemaTypeArray   SchemaType = "array"
	SchemaTypeObject  SchemaType = "object"
)

// Schema represents a type definition
type Schema struct {
	Type             SchemaType
	Ref              string // Reference to another type (if any)
	Description      string
	Required         []string
	Properties       map[string]Schema // For object types
	Items            *Schema           // For array types
	Format           string            // For additional type information (e.g. int32, int64)
	Enum             []interface{}     // For enum types
	IsResponseSchema bool              // Whether this schema is a response schema
}

// RequestBody represents a request body in an operation
type RequestBody struct {
	Required bool
	Schema   Schema
}

// MediaType represents a media type in a request or response
type MediaType struct {
	Schema Schema
}

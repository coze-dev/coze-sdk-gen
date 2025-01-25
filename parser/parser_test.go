package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_ParseOpenAPI(t *testing.T) {
	// Read OpenAPI file
	yamlContent, err := os.ReadFile("../openapi.yaml")
	require.NoError(t, err)

	// Create parser
	parser, err := NewParser(nil)
	require.NoError(t, err)

	// Parse OpenAPI
	modules, err := parser.ParseOpenAPI(yamlContent)
	require.NoError(t, err)

	// Get bots module
	botsModule, ok := modules["bots"]
	require.True(t, ok, "bots module not found")

	// Convert to JSON
	// jsonData, err := json.MarshalIndent(modules, "", "  ")
	jsonData, err := json.MarshalIndent(botsModule, "", "  ")
	require.NoError(t, err)

	// Write to file
	outputPath := filepath.Join(".", "bots_module.json")
	err = os.WriteFile(outputPath, jsonData, 0o644)
	require.NoError(t, err)
}

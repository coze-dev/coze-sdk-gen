package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"code.byted.org/gopkg/logs"
)

func TestGeneratePythonSDK(t *testing.T) {
	// Read the YAML file from current directory
	yamlPath := filepath.Join(".", "openapi.yaml")
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("Failed to read YAML file: %v", err)
	}

	// Generate Python code
	handler := &GeneratePythonSDKHandler{}
	files, oerr := handler.GeneratePythonSDK(context.Background(), yamlContent)
	if oerr != nil {
		t.Fatalf("Failed to generate Python SDK: %v", oerr)
	}

	// Create base directory
	baseDir := "/home/luoyangze.ptrl/python/coze-dev/coze-py/cozepy/"
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Write each generated file
	for dir, content := range files {
		outputPath := filepath.Join(baseDir, dir, "generated_sdk.py")

		// Create subdirectory if needed
		err = os.MkdirAll(filepath.Dir(outputPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", dir, err)
		}

		err = os.WriteFile(outputPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", dir, err)
		}
		logs.Info("Successfully generated Python file at: %s", outputPath)
	}
}

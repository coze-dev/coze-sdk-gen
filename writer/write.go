package writer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func WriteOutput(ctx context.Context, files map[string]string, outputPath string) error {
	// Create base directory
	err := os.MkdirAll(outputPath, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Write each generated file
	for dir, content := range files {
		// Convert module name (with dots) to directory path
		dirPath := strings.ReplaceAll(dir, ".", string(os.PathSeparator))
		outputFilePath := filepath.Join(outputPath, dirPath, "__init__.py")

		// Create subdirectory if needed
		err = os.MkdirAll(filepath.Dir(outputFilePath), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", dir, err)
		}

		err = os.WriteFile(outputFilePath, []byte(content), 0o644)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %v", dir, err)
		}
		log.Printf("Successfully generated Python file at: %s", outputFilePath)
	}

	fmt.Println("SDK generation completed successfully!")
	return nil
}

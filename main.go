package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Define command line flags
	lang := flag.String("lang", "", "SDK language to generate (currently only supports 'python')")
	outputPath := flag.String("output", "", "Output directory path for the generated SDK")
	flag.Parse()

	// Validate required flags
	if *lang == "" || *outputPath == "" {
		fmt.Println("Error: both -lang and -output flags are required")
		flag.Usage()
		os.Exit(1)
	}

	// Validate language support
	if *lang != "python" {
		fmt.Printf("Error: unsupported language %q (currently only supports 'python')\n", *lang)
		os.Exit(1)
	}

	// Read the YAML file from current directory
	yamlPath := filepath.Join(".", "openapi.yaml")
	yamlContent, err := os.ReadFile(yamlPath)
	if err != nil {
		fmt.Printf("Failed to read YAML file: %v\n", err)
		os.Exit(1)
	}

	// Generate SDK code
	handler := &GeneratePythonSDKHandler{}
	files, err := handler.GeneratePythonSDK(context.Background(), yamlContent)
	if err != nil {
		fmt.Printf("Failed to generate Python SDK: %v\n", err)
		os.Exit(1)
	}

	// Create base directory
	err = os.MkdirAll(*outputPath, 0755)
	if err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Write each generated file
	for dir, content := range files {
		outputFilePath := filepath.Join(*outputPath, dir, "generated_sdk.py")

		// Create subdirectory if needed
		err = os.MkdirAll(filepath.Dir(outputFilePath), 0755)
		if err != nil {
			fmt.Printf("Failed to create directory for %s: %v\n", dir, err)
			os.Exit(1)
		}

		err = os.WriteFile(outputFilePath, []byte(content), 0644)
		if err != nil {
			fmt.Printf("Failed to write file %s: %v\n", dir, err)
			os.Exit(1)
		}
		log.Printf("Successfully generated Python file at: %s", outputFilePath)
	}

	fmt.Println("SDK generation completed successfully!")
}

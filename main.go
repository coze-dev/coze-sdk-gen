package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	lang       string
	outputPath string
)

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "", "SDK language to generate")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path for the generated SDK")

	// Mark flags as required
	rootCmd.MarkFlagRequired("lang")
	rootCmd.MarkFlagRequired("output")

	// Add validation for lang flag
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Validate language support
		supportedLangs := map[string]bool{"python": true}
		if !supportedLangs[lang] {
			return fmt.Errorf("unsupported language %q (currently only supports 'python')", lang)
		}
		return nil
	}
}

var rootCmd = &cobra.Command{
	Use:   "coze-sdk-gen",
	Short: "Generate SDK from OpenAPI specification",
	Long: `A generator tool that creates SDK from OpenAPI specification.
Currently supports generating Python SDK.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read the YAML file from current directory
		yamlPath := filepath.Join(".", "openapi.yaml")
		yamlContent, err := os.ReadFile(yamlPath)
		if err != nil {
			return fmt.Errorf("failed to read YAML file: %v", err)
		}

		// Generate SDK code
		handler := &GeneratePythonSDKHandler{}
		files, err := handler.GeneratePythonSDK(context.Background(), yamlContent)
		if err != nil {
			return fmt.Errorf("failed to generate Python SDK: %v", err)
		}

		// Create base directory
		err = os.MkdirAll(outputPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		// Write each generated file
		for dir, content := range files {
			outputFilePath := filepath.Join(outputPath, dir, "generated_sdk.py")

			// Create subdirectory if needed
			err = os.MkdirAll(filepath.Dir(outputFilePath), 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory for %s: %v", dir, err)
			}

			err = os.WriteFile(outputFilePath, []byte(content), 0644)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %v", dir, err)
			}
			log.Printf("Successfully generated Python file at: %s", outputFilePath)
		}

		fmt.Println("SDK generation completed successfully!")
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

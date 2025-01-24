package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/generator/python"
	"github.com/spf13/cobra"
)

var (
	lang       string
	outputPath string
	module     string
)

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "", "SDK language to generate")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory path for the generated SDK")
	rootCmd.Flags().StringVarP(&module, "module", "m", "", "Specific module to generate")

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
	Use:   "coze-sdk-gen <openapi.yaml>",
	Short: "Generate SDK from OpenAPI specification",
	Long: `A generator tool that creates SDK from OpenAPI specification.
Currently supports generating Python SDK.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read the YAML file
		yamlPath := args[0]
		yamlContent, err := os.ReadFile(yamlPath)
		if err != nil {
			return fmt.Errorf("failed to read YAML file: %v", err)
		}

		// Generate SDK code based on language
		var files map[string]string
		switch lang {
		case "python":
			generator := python.Generator{}
			files, err = generator.Generate(context.Background(), yamlContent)
			if err != nil {
				return fmt.Errorf("failed to generate Python SDK: %v", err)
			}
		default:
			return fmt.Errorf("unsupported language %q", lang)
		}

		// Filter files by module if specified
		if module != "" {
			filteredFiles := make(map[string]string)
			for dir, content := range files {
				if dir == module {
					filteredFiles[dir] = content
				}
			}
			files = filteredFiles
		}

		// Create base directory
		err = os.MkdirAll(outputPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		// Write each generated file
		for dir, content := range files {
			// Convert module name (with dots) to directory path
			dirPath := strings.ReplaceAll(dir, ".", string(os.PathSeparator))
			outputFilePath := filepath.Join(outputPath, dirPath, "__init__.py")

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

		// Run ruff format on the generated files
		ruffCmd := exec.Command("poetry", "run", "ruff", "format", ".")
		ruffCmd.Dir = outputPath
		ruffOutput, err := ruffCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Warning: Failed to run ruff format: %v\nOutput: %s\n", err, ruffOutput)
		} else {
			fmt.Println("Successfully formatted code with ruff!")
		}

		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

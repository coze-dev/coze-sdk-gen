package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coze-dev/coze-sdk-gen/formater"
	"github.com/coze-dev/coze-sdk-gen/generator"
	"github.com/coze-dev/coze-sdk-gen/writer"
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
		files, err := generator.Generate(context.Background(), lang, yamlContent, module)
		if err != nil {
			return err
		}

		// Create directory and files
		if err = writer.WriteOutput(context.Background(), files, outputPath); err != nil {
			return err
		}

		// Run format on the generated files
		if err := formater.Format(context.Background(), lang, outputPath); err != nil {
			return err
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

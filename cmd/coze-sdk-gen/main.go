package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/generator"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
	"github.com/coze-dev/coze-sdk-gen/internal/version"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "coze-sdk-gen: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("coze-sdk-gen", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	showVersion := fs.Bool("version", false, "print version")
	configPath := fs.String("config", "config/generator.yaml", "path to generator config file")
	swaggerPath := fs.String("swagger", "coze-openapi.yaml", "path to OpenAPI swagger yaml file")
	languageArg := fs.String("language", "", "target language (python/go), required")
	outputArg := fs.String("output-sdk", "", "output sdk directory, required")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		_, err := fmt.Fprintln(stdout, version.String())
		return err
	}
	lang := strings.ToLower(strings.TrimSpace(*languageArg))
	if lang == "" {
		return fmt.Errorf("--language is required")
	}
	if lang != "python" && lang != "go" {
		return fmt.Errorf("unsupported language %q, supported languages: python, go", lang)
	}
	output := strings.TrimSpace(*outputArg)
	if output == "" {
		return fmt.Errorf("--output-sdk is required")
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	cfg.Language = lang
	cfg.OutputSDK = output
	if err := cfg.Validate(); err != nil {
		return err
	}

	doc, err := openapi.Load(*swaggerPath)
	if err != nil {
		return err
	}

	result, err := generator.Run(cfg, doc)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(
		stdout,
		"language=%s generated_files=%d generated_ops=%d output=%s\n",
		cfg.Language,
		result.GeneratedFiles,
		result.GeneratedOps,
		cfg.OutputSDK,
	)
	return err
}

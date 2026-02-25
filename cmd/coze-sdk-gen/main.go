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
	languageOverride := fs.String("language", "", "override language in config (python/go)")
	outputOverride := fs.String("output-sdk", "", "override output sdk directory in config")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		_, err := fmt.Fprintln(stdout, version.String())
		return err
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		return err
	}
	if *languageOverride != "" {
		cfg.Language = strings.ToLower(strings.TrimSpace(*languageOverride))
	}
	if *outputOverride != "" {
		cfg.OutputSDK = *outputOverride
	}
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

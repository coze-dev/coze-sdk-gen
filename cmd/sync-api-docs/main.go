package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/coze-dev/coze-sdk-gen/internal/apidocs"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "sync-api-docs: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("sync-api-docs", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	indexURL := fs.String("index-url", apidocs.DefaultIndexURL, "llms index url")
	section := fs.String("section", apidocs.DefaultSection, "llms section name")
	markdownDir := fs.String("markdown-dir", apidocs.DefaultMarkdownDir, "output dir for api markdown files")
	swaggerDir := fs.String("swagger-dir", apidocs.DefaultSwaggerDir, "output dir for swagger yaml files")
	timeout := fs.Duration("timeout", apidocs.DefaultTimeout, "http timeout")

	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := &http.Client{Timeout: *timeout}
	result, err := apidocs.Sync(ctx, apidocs.Options{
		IndexURL:          *indexURL,
		Section:           *section,
		MarkdownOutputDir: *markdownDir,
		SwaggerOutputDir:  *swaggerDir,
		HTTPClient:        client,
	})
	if err != nil {
		return err
	}

	fmt.Printf(
		"links=%d fetched=%d generated=%d skipped=%d markdown_dir=%s swagger_dir=%s\n",
		result.TotalLinks,
		result.FetchedDocs,
		result.GeneratedFiles,
		result.SkippedDocs,
		result.MarkdownDir,
		result.SwaggerDir,
	)
	return nil
}

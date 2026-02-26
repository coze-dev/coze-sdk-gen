package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/coze-dev/coze-sdk-gen/internal/apidocsync"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "coze-api-doc-sync: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("coze-api-doc-sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	llmsURL := fs.String("llms-url", "https://docs.coze.cn/llms.txt", "llms index URL")
	section := fs.String("section", "developer_guides", "llms section name")
	outputRoot := fs.String("output-root", "docs", "output root directory")
	markdownSubdir := fs.String("markdown-subdir", "api-markdown", "markdown output subdirectory")
	swaggerSubdir := fs.String("swagger-subdir", "api-swagger", "swagger output subdirectory")
	httpTimeout := fs.Duration("http-timeout", 30*time.Second, "HTTP timeout")

	if err := fs.Parse(args); err != nil {
		return err
	}

	_, err := apidocsync.Run(context.Background(), stdout, apidocsync.Options{
		LLMSURL:        *llmsURL,
		Section:        *section,
		OutputRoot:     *outputRoot,
		MarkdownSubdir: *markdownSubdir,
		SwaggerSubdir:  *swaggerSubdir,
		HTTPTimeout:    *httpTimeout,
	})
	return err
}

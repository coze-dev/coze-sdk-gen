package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/dirdiff"
	"github.com/coze-dev/coze-sdk-gen/internal/fsutil"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

var (
	copySelected = fsutil.CopySelected
	compareDirs  = dirdiff.CompareDirs
)

func GeneratePython(cfg *config.Config, doc *openapi.Document) (Result, error) {
	if cfg == nil {
		return Result{}, fmt.Errorf("config is required")
	}

	report := cfg.ValidateAgainstSwagger(doc)
	if report.HasErrors() {
		return Result{}, fmt.Errorf("config and swagger mismatch: %s", report.Error())
	}

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	excludes := make([]string, 0, len(cfg.Copy.Exclude)+1)
	excludes = append(excludes, cfg.Copy.Exclude...)
	excludes = append(excludes, ".git")

	copyResult, err := copySelected(cfg.SourceSDK, cfg.OutputSDK, cfg.Copy.Include, excludes)
	if err != nil {
		return Result{}, fmt.Errorf("copy python sdk source files: %w", err)
	}

	diffs, err := compareDirs(cfg.SourceSDK, cfg.OutputSDK, excludes)
	if err != nil {
		return Result{}, fmt.Errorf("compare source and generated sdk: %w", err)
	}
	if len(diffs) > 0 {
		sample := make([]string, 0, min(5, len(diffs)))
		for i := 0; i < len(diffs) && i < 5; i++ {
			sample = append(sample, fmt.Sprintf("%s:%s", diffs[i].Type, diffs[i].Path))
		}
		return Result{}, fmt.Errorf(
			"generated sdk does not match source sdk, total_diffs=%d, sample=[%s]",
			len(diffs),
			strings.Join(sample, ", "),
		)
	}

	return Result{CopiedFiles: copyResult.CopiedFiles}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

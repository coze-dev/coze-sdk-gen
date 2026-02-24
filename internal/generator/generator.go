package generator

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type Result struct {
	CopiedFiles int
}

func Run(cfg *config.Config, doc *openapi.Document) (Result, error) {
	if cfg == nil {
		return Result{}, fmt.Errorf("config is required")
	}

	switch strings.ToLower(cfg.Language) {
	case "python":
		return GeneratePython(cfg, doc)
	case "go":
		return Result{}, fmt.Errorf("language %q is not implemented yet", cfg.Language)
	default:
		return Result{}, fmt.Errorf("unsupported language %q", cfg.Language)
	}
}

package generator

import (
	"fmt"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	gogen "github.com/coze-dev/coze-sdk-gen/internal/generator/go"
	pygen "github.com/coze-dev/coze-sdk-gen/internal/generator/python"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type Result struct {
	GeneratedFiles int
	GeneratedOps   int
}

func Run(cfg *config.Config, doc *openapi.Document) (Result, error) {
	if cfg == nil {
		return Result{}, fmt.Errorf("config is required")
	}

	switch strings.ToLower(cfg.Language) {
	case "python":
		return GeneratePython(cfg, doc)
	case "go":
		return GenerateGo(cfg, doc)
	default:
		return Result{}, fmt.Errorf("unsupported language %q", cfg.Language)
	}
}

func GeneratePython(cfg *config.Config, doc *openapi.Document) (Result, error) {
	result, err := pygen.GeneratePython(cfg, doc)
	if err != nil {
		return Result{}, err
	}
	return Result{
		GeneratedFiles: result.GeneratedFiles,
		GeneratedOps:   result.GeneratedOps,
	}, nil
}

func GenerateGo(cfg *config.Config, doc *openapi.Document) (Result, error) {
	result, err := gogen.GenerateGo(cfg, doc)
	if err != nil {
		return Result{}, err
	}
	return Result{
		GeneratedFiles: result.GeneratedFiles,
		GeneratedOps:   result.GeneratedOps,
	}, nil
}

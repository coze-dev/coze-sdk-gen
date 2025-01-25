package generator

import (
	"context"
	"fmt"

	"github.com/coze-dev/coze-sdk-gen/consts"
	"github.com/coze-dev/coze-sdk-gen/generator/python"
)

func Generate(ctx context.Context, lang string, yamlContent []byte, module string) (map[string]string, error) {
	var files map[string]string
	var err error

	switch lang {
	case consts.Python:
		generator := python.Generator{}
		files, err = generator.Generate(ctx, yamlContent)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Python SDK: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported language %q", lang)
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

	return files, nil
}

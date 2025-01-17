package python

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"text/template"

	"github.com/coze-dev/coze-sdk-gen/parser"
)

//go:embed templates/sdk.tmpl
var templateFS embed.FS

// Generator handles Python SDK generation
type Generator struct{}

// Generate generates Python SDK code from parsed OpenAPI data
func (g Generator) Generate(ctx context.Context, yamlContent []byte) (map[string]string, error) {
	p := parser.Parser{}
	modules, _, err := p.ParseOpenAPI(ctx, yamlContent)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI failed: %w", err)
	}

	// Generate code for each module
	files := make(map[string]string)

	// Read template
	tmpl, err := template.New("python").Parse(g.getTemplate())
	if err != nil {
		return nil, fmt.Errorf("parse template failed: %w", err)
	}

	// Generate module files
	for moduleName, module := range modules {
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]interface{}{
			"ModuleName": moduleName,
			"Operations": module.Operations,
			"Classes":    module.Classes,
		})
		if err != nil {
			return nil, fmt.Errorf("execute template failed: %w", err)
		}
		files[fmt.Sprintf("%s", moduleName)] = buf.String()
	}

	return files, nil
}

func (g Generator) getTemplate() string {
	// Read template from embedded file
	templateContent, err := fs.ReadFile(templateFS, "templates/sdk.tmpl")
	if err != nil {
		return ""
	}
	return string(templateContent)
}

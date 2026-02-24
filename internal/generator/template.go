package generator

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"text/template"
)

//go:embed templates/python/*.tpl
var pythonTemplateFS embed.FS

func renderPythonTemplate(templateName string, data any) (string, error) {
	templatePath := path.Join("templates", "python", templateName)
	tplContent, err := pythonTemplateFS.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read python template %q: %w", templatePath, err)
	}

	tpl, err := template.New(templateName).Parse(string(tplContent))
	if err != nil {
		return "", fmt.Errorf("parse python template %q: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute python template %q: %w", templateName, err)
	}

	return buf.String(), nil
}

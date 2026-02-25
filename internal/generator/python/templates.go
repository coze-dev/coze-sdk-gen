package python

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"text/template"
)

//go:embed all:templates
var pythonTemplateFS embed.FS

func RenderPythonTemplate(templateName string, data any) (string, error) {
	templatePath := path.Join("templates", templateName)
	tplContent, err := loadPythonAsset(templatePath)
	if err != nil {
		return "", err
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

func RenderPythonRawAsset(assetName string) (string, error) {
	assetPath := path.Join("templates", assetName)
	content, err := loadPythonAsset(assetPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func loadPythonAsset(assetPath string) ([]byte, error) {
	content, err := pythonTemplateFS.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("read python asset %q: %w", assetPath, err)
	}
	return content, nil
}

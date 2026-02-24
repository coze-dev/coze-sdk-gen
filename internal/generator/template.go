package generator

import (
	"bytes"
	"embed"
	"fmt"
	"path"
	"text/template"
)

//go:embed templates/python/*.tpl all:templates/python/special
var pythonTemplateFS embed.FS

func renderPythonTemplate(templateName string, data any) (string, error) {
	templatePath := path.Join("templates", "python", templateName)
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

func renderPythonRawAsset(assetName string) (string, error) {
	assetPath := path.Join("templates", "python", assetName)
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

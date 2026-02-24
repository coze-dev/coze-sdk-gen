package generator

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/python/*.tpl all:templates/python/legacy_static
var pythonAssetsFS embed.FS

func renderPythonTemplate(templateName string, data any) (string, error) {
	templatePath := path.Join("templates", "python", templateName)
	tplContent, err := pythonAssetsFS.ReadFile(templatePath)
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

func listPythonStaticFiles() ([]string, error) {
	const root = "templates/python/legacy_static"
	files := make([]string, 0)
	err := fs.WalkDir(pythonAssetsFS, root, func(current string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(current, root+"/")
		if rel == current {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk python static files: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func readPythonStaticFile(pathName string) ([]byte, error) {
	fullPath := path.Join("templates", "python", "legacy_static", pathName)
	content, err := pythonAssetsFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read python static file %q: %w", fullPath, err)
	}
	return content, nil
}

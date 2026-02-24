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

//go:embed templates/python/*.tpl all:templates/python/support_static
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

func listPythonSupportFiles() ([]string, error) {
	const root = "templates/python/support_static"
	files := make([]string, 0)
	err := fs.WalkDir(pythonTemplateFS, root, func(current string, d fs.DirEntry, walkErr error) error {
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
		return nil, fmt.Errorf("walk python support files: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func readPythonSupportFile(pathName string) ([]byte, error) {
	fullPath := path.Join("templates", "python", "support_static", pathName)
	content, err := pythonTemplateFS.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read python support file %q: %w", fullPath, err)
	}
	return content, nil
}

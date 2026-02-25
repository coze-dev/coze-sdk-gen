package gogen

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"
)

//go:embed all:templates/go_runtime
var goRuntimeFS embed.FS

//go:embed all:templates/go_api
var goAPIFS embed.FS

//go:embed all:templates/go_extra
var goExtraFS embed.FS

func renderGoRuntimeAsset(assetName string) (string, error) {
	assetPath := path.Join("templates", "go_runtime", assetName)
	content, err := goRuntimeFS.ReadFile(assetPath)
	if err != nil {
		return "", fmt.Errorf("read go runtime asset %q: %w", assetPath, err)
	}
	return string(content), nil
}

func listGoExtraAssets() ([]string, error) {
	root := path.Join("templates", "go_extra")
	assets := make([]string, 0)
	if err := fs.WalkDir(goExtraFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, root+"/")
		if rel == p {
			return nil
		}
		assets = append(assets, rel)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk go extra assets: %w", err)
	}
	return assets, nil
}

func renderGoExtraAsset(assetName string) ([]byte, error) {
	assetPath := path.Join("templates", "go_extra", assetName)
	content, err := goExtraFS.ReadFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("read go extra asset %q: %w", assetPath, err)
	}
	return content, nil
}

func renderGoAPIAsset(assetName string, data any) (string, error) {
	assetPath := path.Join("templates", "go_api", assetName)
	content, err := goAPIFS.ReadFile(assetPath)
	if err != nil {
		return "", fmt.Errorf("read go api asset %q: %w", assetPath, err)
	}

	tmpl, err := template.New(assetName).Option("missingkey=error").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parse go api template %q: %w", assetPath, err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("render go api template %q: %w", assetPath, err)
	}
	return out.String(), nil
}

package gogen

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed all:templates/go_runtime
var goRuntimeFS embed.FS

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

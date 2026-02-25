package generator

import (
	"embed"
	"fmt"
	"path"
)

//go:embed all:templates/go_runtime
var goRuntimeFS embed.FS

func renderGoRuntimeAsset(assetName string) (string, error) {
	assetPath := path.Join("templates", "go_runtime", assetName)
	content, err := goRuntimeFS.ReadFile(assetPath)
	if err != nil {
		return "", fmt.Errorf("read go runtime asset %q: %w", assetPath, err)
	}
	return string(content), nil
}

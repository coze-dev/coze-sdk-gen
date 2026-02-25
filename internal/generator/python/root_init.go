package python

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var pythonTopLevelClassPattern = regexp.MustCompile(`(?m)^class\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:\(|:)`)

var pythonRootInitExtraExports = map[string][]string{
	"auth": {
		"load_oauth_app_from_config",
	},
	"config": {
		"COZE_CN_BASE_URL",
		"COZE_COM_BASE_URL",
		"DEFAULT_CONNECTION_LIMITS",
		"DEFAULT_TIMEOUT",
	},
	"log": {
		"setup_logging",
	},
	"version": {
		"VERSION",
	},
}

func renderPythonRootInit(rootDir string) (string, error) {
	moduleClassExports, err := collectPythonRootInitClassExports(rootDir)
	if err != nil {
		return "", err
	}
	for module, names := range pythonRootInitExtraExports {
		moduleClassExports[module] = append(moduleClassExports[module], names...)
	}

	modules := make([]string, 0, len(moduleClassExports))
	for module, names := range moduleClassExports {
		cleanNames := uniqueSortedNonEmptyStrings(names)
		if len(cleanNames) == 0 {
			continue
		}
		moduleClassExports[module] = cleanNames
		modules = append(modules, module)
	}
	sort.Strings(modules)

	exportOwner := map[string]string{}
	allExports := make([]string, 0)
	for _, module := range modules {
		for _, name := range moduleClassExports[module] {
			if owner, exists := exportOwner[name]; exists && owner != module {
				return "", fmt.Errorf("duplicate root export %q from modules %q and %q", name, owner, module)
			}
			exportOwner[name] = module
			allExports = append(allExports, name)
		}
	}
	sort.Strings(allExports)

	var buf bytes.Buffer
	for _, module := range modules {
		names := moduleClassExports[module]
		if len(names) == 1 {
			buf.WriteString(fmt.Sprintf("from .%s import %s\n", module, names[0]))
			continue
		}
		buf.WriteString(fmt.Sprintf("from .%s import (\n", module))
		for _, name := range names {
			buf.WriteString(fmt.Sprintf("    %s,\n", name))
		}
		buf.WriteString(")\n")
	}
	if len(modules) > 0 {
		buf.WriteString("\n")
	}
	buf.WriteString("__all__ = [\n")
	for _, name := range allExports {
		buf.WriteString(fmt.Sprintf("    %q,\n", name))
	}
	buf.WriteString("]\n")
	return buf.String(), nil
}

func collectPythonRootInitClassExports(rootDir string) (map[string][]string, error) {
	exports := map[string][]string{}
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".py" {
			return nil
		}
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return fmt.Errorf("get relative path for %q: %w", path, err)
		}
		relPath = filepath.Clean(relPath)
		if relPath == "__init__.py" {
			return nil
		}
		modulePath, err := pythonRootImportModuleFromRelativePath(relPath)
		if err != nil {
			return err
		}
		classNames, err := collectPublicTopLevelClassesFromFile(path)
		if err != nil {
			return err
		}
		if len(classNames) == 0 {
			return nil
		}
		exports[modulePath] = append(exports[modulePath], classNames...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("collect python root init exports: %w", err)
	}
	return exports, nil
}

func pythonRootImportModuleFromRelativePath(relPath string) (string, error) {
	clean := filepath.ToSlash(filepath.Clean(relPath))
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return "", fmt.Errorf("path %q escapes root package", relPath)
	}
	switch {
	case strings.HasSuffix(clean, "/__init__.py"):
		clean = strings.TrimSuffix(clean, "/__init__.py")
	case strings.HasSuffix(clean, ".py"):
		clean = strings.TrimSuffix(clean, ".py")
	default:
		return "", fmt.Errorf("path %q is not a python module", relPath)
	}
	clean = strings.Trim(clean, "/")
	if clean == "" {
		return "", fmt.Errorf("empty module path from %q", relPath)
	}
	return strings.ReplaceAll(clean, "/", "."), nil
}

func collectPublicTopLevelClassesFromFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read python module %q: %w", path, err)
	}
	return collectPublicTopLevelClassesFromSource(string(content)), nil
}

func collectPublicTopLevelClassesFromSource(content string) []string {
	matches := pythonTopLevelClassPattern.FindAllStringSubmatch(content, -1)
	classes := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		className := strings.TrimSpace(match[1])
		if className == "" || strings.HasPrefix(className, "_") {
			continue
		}
		if strings.HasSuffix(className, "Client") {
			continue
		}
		if _, exists := seen[className]; exists {
			continue
		}
		seen[className] = struct{}{}
		classes = append(classes, className)
	}
	return classes
}

func uniqueSortedNonEmptyStrings(values []string) []string {
	seen := map[string]struct{}{}
	unique := make([]string, 0, len(values))
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" {
			continue
		}
		if _, exists := seen[clean]; exists {
			continue
		}
		seen[clean] = struct{}{}
		unique = append(unique, clean)
	}
	sort.Strings(unique)
	return unique
}

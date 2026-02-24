package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Language  string    `yaml:"language"`
	SourceSDK string    `yaml:"source_sdk"`
	OutputSDK string    `yaml:"output_sdk"`
	Copy      Copy      `yaml:"copy"`
	API       APIConfig `yaml:"api"`
}

type Copy struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

type APIConfig struct {
	Packages          []Package                    `yaml:"packages"`
	OperationMappings []OperationMapping           `yaml:"operation_mappings"`
	IgnoreOperations  []OperationRef               `yaml:"ignore_operations"`
	FieldAliases      map[string]map[string]string `yaml:"field_aliases"`
}

type Package struct {
	Name                  string   `yaml:"name"`
	SourceDir             string   `yaml:"source_dir"`
	PathPrefixes          []string `yaml:"path_prefixes"`
	AllowMissingInSwagger bool     `yaml:"allow_missing_in_swagger"`
}

type OperationMapping struct {
	Path                  string   `yaml:"path"`
	Method                string   `yaml:"method"`
	SDKMethods            []string `yaml:"sdk_methods"`
	AllowMissingInSwagger bool     `yaml:"allow_missing_in_swagger"`
}

type OperationRef struct {
	Path                  string `yaml:"path"`
	Method                string `yaml:"method"`
	AllowMissingInSwagger bool   `yaml:"allow_missing_in_swagger"`
}

type ValidationReport struct {
	MissingOperations []OperationRef
	UnmatchedPrefixes []string
}

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}
	return Parse(content)
}

func Parse(content []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if len(c.Copy.Include) == 0 {
		c.Copy.Include = []string{
			"cozepy",
			"examples",
			"tests",
			"README.md",
			"pyproject.toml",
			"poetry.lock",
			"LICENSE",
			"codecov.yml",
			".gitignore",
			".pre-commit-config.yaml",
			".github",
		}
	}
	if c.API.FieldAliases == nil {
		c.API.FieldAliases = map[string]map[string]string{}
	}
}

func (c *Config) Validate() error {
	if c.Language == "" {
		return fmt.Errorf("language is required")
	}
	if strings.ToLower(c.Language) != "python" {
		return fmt.Errorf("unsupported language %q, only python is supported", c.Language)
	}
	if c.SourceSDK == "" {
		return fmt.Errorf("source_sdk is required")
	}
	if c.OutputSDK == "" {
		return fmt.Errorf("output_sdk is required")
	}
	if len(c.Copy.Include) == 0 {
		return fmt.Errorf("copy.include should not be empty")
	}

	seenPackageName := map[string]struct{}{}
	for i, pkg := range c.API.Packages {
		if pkg.Name == "" {
			return fmt.Errorf("api.packages[%d].name is required", i)
		}
		if _, dup := seenPackageName[pkg.Name]; dup {
			return fmt.Errorf("duplicate package name %q", pkg.Name)
		}
		seenPackageName[pkg.Name] = struct{}{}

		if pkg.SourceDir == "" {
			return fmt.Errorf("api.packages[%d].source_dir is required", i)
		}
		for j, prefix := range pkg.PathPrefixes {
			if prefix == "" || !strings.HasPrefix(prefix, "/") {
				return fmt.Errorf("api.packages[%d].path_prefixes[%d] must start with '/'", i, j)
			}
		}
	}

	for i, mapping := range c.API.OperationMappings {
		if err := validateOperationRef(mapping.Path, mapping.Method, fmt.Sprintf("api.operation_mappings[%d]", i)); err != nil {
			return err
		}
		if len(mapping.SDKMethods) == 0 {
			return fmt.Errorf("api.operation_mappings[%d].sdk_methods should not be empty", i)
		}
	}

	for i, ref := range c.API.IgnoreOperations {
		if err := validateOperationRef(ref.Path, ref.Method, fmt.Sprintf("api.ignore_operations[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ValidateAgainstSwagger(doc *openapi.Document) ValidationReport {
	report := ValidationReport{
		MissingOperations: make([]OperationRef, 0),
		UnmatchedPrefixes: make([]string, 0),
	}

	if doc == nil {
		for _, op := range c.API.OperationMappings {
			report.MissingOperations = append(report.MissingOperations, OperationRef{Path: op.Path, Method: normalizeMethod(op.Method)})
		}
		for _, op := range c.API.IgnoreOperations {
			report.MissingOperations = append(report.MissingOperations, OperationRef{Path: op.Path, Method: normalizeMethod(op.Method)})
		}
		return report
	}

	seenMissing := map[string]struct{}{}
	appendMissing := func(path, method string) {
		key := path + "#" + method
		if _, ok := seenMissing[key]; ok {
			return
		}
		seenMissing[key] = struct{}{}
		report.MissingOperations = append(report.MissingOperations, OperationRef{Path: path, Method: method})
	}

	for _, op := range c.API.OperationMappings {
		method := normalizeMethod(op.Method)
		if op.AllowMissingInSwagger {
			continue
		}
		if !doc.HasOperation(method, op.Path) {
			appendMissing(op.Path, method)
		}
	}

	for _, op := range c.API.IgnoreOperations {
		method := normalizeMethod(op.Method)
		if op.AllowMissingInSwagger {
			continue
		}
		if !doc.HasOperation(method, op.Path) {
			appendMissing(op.Path, method)
		}
	}

	for _, pkg := range c.API.Packages {
		if pkg.AllowMissingInSwagger {
			continue
		}
		for _, prefix := range pkg.PathPrefixes {
			if len(doc.PathsWithPrefix(prefix)) == 0 {
				report.UnmatchedPrefixes = append(report.UnmatchedPrefixes, prefix)
			}
		}
	}

	sort.Slice(report.MissingOperations, func(i, j int) bool {
		if report.MissingOperations[i].Path == report.MissingOperations[j].Path {
			return report.MissingOperations[i].Method < report.MissingOperations[j].Method
		}
		return report.MissingOperations[i].Path < report.MissingOperations[j].Path
	})
	sort.Strings(report.UnmatchedPrefixes)

	return report
}

func (r ValidationReport) HasErrors() bool {
	return len(r.MissingOperations) > 0 || len(r.UnmatchedPrefixes) > 0
}

func (r ValidationReport) Error() string {
	parts := make([]string, 0)
	if len(r.MissingOperations) > 0 {
		items := make([]string, 0, len(r.MissingOperations))
		for _, op := range r.MissingOperations {
			items = append(items, op.Method+" "+op.Path)
		}
		parts = append(parts, "missing operations in swagger: "+strings.Join(items, ", "))
	}
	if len(r.UnmatchedPrefixes) > 0 {
		parts = append(parts, "path prefixes not found in swagger: "+strings.Join(r.UnmatchedPrefixes, ", "))
	}
	return strings.Join(parts, "; ")
}

func validateOperationRef(path, method, field string) error {
	if path == "" || !strings.HasPrefix(path, "/") {
		return fmt.Errorf("%s.path must start with '/'", field)
	}
	normalizedMethod := normalizeMethod(method)
	switch normalizedMethod {
	case "get", "put", "post", "delete", "patch", "options", "head", "trace":
		return nil
	default:
		return fmt.Errorf("%s.method is invalid: %q", field, method)
	}
}

func normalizeMethod(method string) string {
	return strings.ToLower(strings.TrimSpace(method))
}

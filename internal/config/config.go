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
	OutputSDK string    `yaml:"output_sdk"`
	API       APIConfig `yaml:"api"`
}

type APIConfig struct {
	Packages           []Package                    `yaml:"packages"`
	OperationMappings  []OperationMapping           `yaml:"operation_mappings"`
	IgnoreOperations   []OperationRef               `yaml:"ignore_operations"`
	GenerateOnlyMapped bool                         `yaml:"generate_only_mapped"`
	FieldAliases       map[string]map[string]string `yaml:"field_aliases"`
}

type Package struct {
	Name                  string        `yaml:"name"`
	SourceDir             string        `yaml:"source_dir"`
	PathPrefixes          []string      `yaml:"path_prefixes"`
	AllowMissingInSwagger bool          `yaml:"allow_missing_in_swagger"`
	ClientClass           string        `yaml:"client_class"`
	AsyncClientClass      string        `yaml:"async_client_class"`
	ChildClients          []ChildClient `yaml:"child_clients"`
	ModelSchemas          []ModelSchema `yaml:"model_schemas"`
	EmptyModels           []string      `yaml:"empty_models"`
}

type ChildClient struct {
	Attribute        string `yaml:"attribute"`
	Module           string `yaml:"module"`
	SyncClass        string `yaml:"sync_class"`
	AsyncClass       string `yaml:"async_class"`
	NilCheck         string `yaml:"nil_check"`
	InitWithKeywords bool   `yaml:"init_with_keywords"`
	DisableTypeHints bool   `yaml:"disable_type_hints"`
}

type ModelSchema struct {
	Schema         string   `yaml:"schema"`
	Name           string   `yaml:"name"`
	FieldOrder     []string `yaml:"field_order"`
	RequiredFields []string `yaml:"required_fields"`
	EnumBase       string   `yaml:"enum_base"`
}

type OperationMapping struct {
	Path                     string            `yaml:"path"`
	Method                   string            `yaml:"method"`
	Order                    int               `yaml:"order"`
	SDKMethods               []string          `yaml:"sdk_methods"`
	AllowMissingInSwagger    bool              `yaml:"allow_missing_in_swagger"`
	HTTPMethodOverride       string            `yaml:"http_method_override"`
	DisableRequestBody       bool              `yaml:"disable_request_body"`
	BodyFields               []string          `yaml:"body_fields"`
	BodyRequiredFields       []string          `yaml:"body_required_fields"`
	UseKwargsHeaders         bool              `yaml:"use_kwargs_headers"`
	ParamAliases             map[string]string `yaml:"param_aliases"`
	ArgTypes                 map[string]string `yaml:"arg_types"`
	ResponseType             string            `yaml:"response_type"`
	ResponseCast             string            `yaml:"response_cast"`
	QueryFields              []OperationField  `yaml:"query_fields"`
	Pagination               string            `yaml:"pagination"`
	PaginationDataClass      string            `yaml:"pagination_data_class"`
	PaginationItemType       string            `yaml:"pagination_item_type"`
	PaginationItemsField     string            `yaml:"pagination_items_field"`
	PaginationTotalField     string            `yaml:"pagination_total_field"`
	PaginationHasMoreField   string            `yaml:"pagination_has_more_field"`
	PaginationNextTokenField string            `yaml:"pagination_next_token_field"`
	PaginationPageNumField   string            `yaml:"pagination_page_num_field"`
	PaginationPageSizeField  string            `yaml:"pagination_page_size_field"`
	PaginationPageTokenField string            `yaml:"pagination_page_token_field"`
}

type OperationField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default"`
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
	if c.OutputSDK == "" {
		return fmt.Errorf("output_sdk is required")
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
		for j, child := range pkg.ChildClients {
			if strings.TrimSpace(child.Attribute) == "" {
				return fmt.Errorf("api.packages[%d].child_clients[%d].attribute is required", i, j)
			}
			if strings.TrimSpace(child.Module) == "" {
				return fmt.Errorf("api.packages[%d].child_clients[%d].module is required", i, j)
			}
			if strings.TrimSpace(child.SyncClass) == "" || strings.TrimSpace(child.AsyncClass) == "" {
				return fmt.Errorf("api.packages[%d].child_clients[%d].sync_class and async_class are required", i, j)
			}
			if strings.TrimSpace(child.NilCheck) != "" {
				switch strings.TrimSpace(child.NilCheck) {
				case "truthy", "is_none":
				default:
					return fmt.Errorf("api.packages[%d].child_clients[%d].nil_check must be one of: truthy, is_none", i, j)
				}
			}
		}
		for j, model := range pkg.ModelSchemas {
			if strings.TrimSpace(model.Schema) == "" || strings.TrimSpace(model.Name) == "" {
				return fmt.Errorf("api.packages[%d].model_schemas[%d].schema and name are required", i, j)
			}
			if strings.TrimSpace(model.EnumBase) != "" {
				switch strings.TrimSpace(model.EnumBase) {
				case "dynamic_str":
				default:
					return fmt.Errorf("api.packages[%d].model_schemas[%d].enum_base must be 'dynamic_str' when set", i, j)
				}
			}
		}
	}

	for i, mapping := range c.API.OperationMappings {
		if err := validateOperationRef(mapping.Path, mapping.Method, fmt.Sprintf("api.operation_mappings[%d]", i)); err != nil {
			return err
		}
		if strings.TrimSpace(mapping.HTTPMethodOverride) != "" {
			if err := validateOperationRef(mapping.Path, mapping.HTTPMethodOverride, fmt.Sprintf("api.operation_mappings[%d].http_method_override", i)); err != nil {
				return err
			}
		}
		if len(mapping.SDKMethods) == 0 {
			return fmt.Errorf("api.operation_mappings[%d].sdk_methods should not be empty", i)
		}
		for j, field := range mapping.QueryFields {
			if strings.TrimSpace(field.Name) == "" {
				return fmt.Errorf("api.operation_mappings[%d].query_fields[%d].name is required", i, j)
			}
		}
		if strings.TrimSpace(mapping.Pagination) != "" {
			pagination := strings.TrimSpace(mapping.Pagination)
			if pagination != "token" && pagination != "number" {
				return fmt.Errorf("api.operation_mappings[%d].pagination must be 'token' or 'number' when set", i)
			}
			if strings.TrimSpace(mapping.PaginationDataClass) == "" || strings.TrimSpace(mapping.PaginationItemType) == "" {
				return fmt.Errorf("api.operation_mappings[%d].pagination_data_class and pagination_item_type are required for pagination", i)
			}
		}
	}

	for i, ref := range c.API.IgnoreOperations {
		if err := validateOperationRef(ref.Path, ref.Method, fmt.Sprintf("api.ignore_operations[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) IsIgnored(path string, method string) bool {
	method = normalizeMethod(method)
	for _, ref := range c.API.IgnoreOperations {
		if ref.Path == path && normalizeMethod(ref.Method) == method {
			return true
		}
	}
	return false
}

func (c *Config) FindOperationMappings(path string, method string) []OperationMapping {
	method = normalizeMethod(method)
	result := make([]OperationMapping, 0)
	for _, mapping := range c.API.OperationMappings {
		if mapping.Path == path && normalizeMethod(mapping.Method) == method {
			result = append(result, mapping)
		}
	}
	return result
}

func (c *Config) ResolvePackage(path string, preferred string) (Package, bool) {
	if preferred != "" {
		for _, pkg := range c.API.Packages {
			if pkg.Name == preferred {
				return pkg, true
			}
		}
	}

	var (
		found      bool
		best       Package
		bestPrefix string
	)
	for _, pkg := range c.API.Packages {
		for _, prefix := range pkg.PathPrefixes {
			if strings.HasPrefix(path, prefix) {
				if !found || len(prefix) > len(bestPrefix) {
					best = pkg
					bestPrefix = prefix
					found = true
				}
			}
		}
	}
	return best, found
}

func ParseSDKMethod(value string) (string, string, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", false
	}
	parts := strings.Split(value, ".")
	if len(parts) == 1 {
		return "", parts[0], true
	}
	if len(parts) == 2 {
		if parts[0] == "" || parts[1] == "" {
			return "", "", false
		}
		return parts[0], parts[1], true
	}
	return "", "", false
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

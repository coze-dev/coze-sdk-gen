package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Language             string           `yaml:"-"`
	OutputSDK            string           `yaml:"-"`
	CommentOverridesFile string           `yaml:"comment_overrides_file"`
	Diff                 DiffConfig       `yaml:"diff"`
	API                  APIConfig        `yaml:"api"`
	CommentOverrides     CommentOverrides `yaml:"-"`
}

type DiffConfig struct {
	IgnorePathsByLanguage map[string][]string `yaml:"ignore_paths_by_language"`
}

var defaultDiffIgnorePathsByLanguage = map[string][]string{
	"python": {
		".git",
		".github",
		".gitignore",
		".pre-commit-config.yaml",
		".vscode",
		"CONTRIBUTING.md",
		"LICENSE",
		"README.md",
		"codecov.yml",
		"examples",
		"poetry.lock",
		"__pycache__",
		"tests",
	},
	"go": {
		".git",
		".github",
		"README.md",
		"*_test.go",
	},
}

type CommentOverrides struct {
	ClassDocstrings         map[string]string   `yaml:"class_docstrings"`
	ClassDocstringStyles    map[string]string   `yaml:"class_docstring_styles"`
	MethodDocstrings        map[string]string   `yaml:"method_docstrings"`
	MethodDocstringStyles   map[string]string   `yaml:"method_docstring_styles"`
	FieldComments           map[string][]string `yaml:"field_comments"`
	InlineFieldComments     map[string]string   `yaml:"inline_field_comments"`
	EnumMemberComments      map[string][]string `yaml:"enum_member_comments"`
	InlineEnumMemberComment map[string]string   `yaml:"inline_enum_member_comments"`
}

type APIConfig struct {
	Packages           []Package                    `yaml:"packages"`
	OperationMappings  []OperationMapping           `yaml:"operation_mappings"`
	IgnoreAPIs         []OperationRef               `yaml:"ignore_apis"`
	GenerateOnlyMapped bool                         `yaml:"generate_only_mapped"`
	FieldAliases       map[string]map[string]string `yaml:"field_aliases"`
}

type Package struct {
	Name                      string        `yaml:"name"`
	SourceDir                 string        `yaml:"source_dir"`
	PathPrefixes              []string      `yaml:"path_prefixes"`
	AllowMissingInSwagger     bool          `yaml:"allow_missing_in_swagger"`
	DisableAutoImports        bool          `yaml:"disable_auto_imports"`
	SeparateCommentedEnum     bool          `yaml:"separate_commented_enum_members"`
	HTTPRequestFromModel      bool          `yaml:"http_request_from_model"`
	ClientClass               string        `yaml:"client_class"`
	AsyncClientClass          string        `yaml:"async_client_class"`
	ChildClients              []ChildClient `yaml:"child_clients"`
	ExtraImports              []ImportSpec  `yaml:"extra_imports"`
	RawImports                []string      `yaml:"raw_imports"`
	ModelSchemas              []ModelSchema `yaml:"model_schemas"`
	EmptyModels               []string      `yaml:"empty_models"`
	PreModelCode              []string      `yaml:"pre_model_code"`
	TopLevelCode              []string      `yaml:"top_level_code"`
	SyncInitPreCode           []string      `yaml:"sync_init_pre_code"`
	AsyncInitPreCode          []string      `yaml:"async_init_pre_code"`
	BlankLineBeforeSyncInit   bool          `yaml:"blank_line_before_sync_init_code"`
	BlankLineBeforeAsyncInit  bool          `yaml:"blank_line_before_async_init_code"`
	SyncInitCode              []string      `yaml:"sync_init_code"`
	AsyncInitCode             []string      `yaml:"async_init_code"`
	SyncExtraMethods          []string      `yaml:"sync_extra_methods"`
	AsyncExtraMethods         []string      `yaml:"async_extra_methods"`
	OverridePaginationClasses []string      `yaml:"override_pagination_classes"`
}

type ChildClient struct {
	Attribute        string `yaml:"attribute"`
	Module           string `yaml:"module"`
	SyncClass        string `yaml:"sync_class"`
	AsyncClass       string `yaml:"async_class"`
	DisableTypeHints bool   `yaml:"disable_type_hints"`
}

type ModelSchema struct {
	Schema                 string            `yaml:"schema"`
	Name                   string            `yaml:"name"`
	BaseClasses            []string          `yaml:"base_classes"`
	BeforeCode             []string          `yaml:"before_code"`
	PrependCode            []string          `yaml:"prepend_code"`
	Builders               []ModelBuilder    `yaml:"builders"`
	BeforeValidators       []ModelValidator  `yaml:"before_validators"`
	SeparateCommentedEnum  *bool             `yaml:"separate_commented_enum_members"`
	FieldOrder             []string          `yaml:"field_order"`
	RequiredFields         []string          `yaml:"required_fields"`
	FieldTypes             map[string]string `yaml:"field_types"`
	FieldDefaults          map[string]string `yaml:"field_defaults"`
	ExcludeUnorderedFields bool              `yaml:"exclude_unordered_fields"`
	EnumBase               string            `yaml:"enum_base"`
	EnumValues             []ModelEnumValue  `yaml:"enum_values"`
	ExtraFields            []ModelField      `yaml:"extra_fields"`
	ExtraCode              []string          `yaml:"extra_code"`
	AllowMissingInSwagger  bool              `yaml:"allow_missing_in_swagger"`
}

type ModelField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Alias    string `yaml:"alias"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default"`
}

type ModelValidator struct {
	Field  string `yaml:"field"`
	Rule   string `yaml:"rule"`
	Method string `yaml:"method"`
}

type ModelBuilder struct {
	Name       string            `yaml:"name"`
	Params     []string          `yaml:"params"`
	ReturnType string            `yaml:"return_type"`
	Args       []ModelBuilderArg `yaml:"args"`
}

type ModelBuilderArg struct {
	Name string `yaml:"name"`
	Expr string `yaml:"expr"`
}

type ModelEnumValue struct {
	Name  string      `yaml:"name"`
	Value interface{} `yaml:"value"`
}

type ImportSpec struct {
	Module string   `yaml:"module"`
	Names  []string `yaml:"names"`
}

type OperationMapping struct {
	Path                           string            `yaml:"path"`
	Method                         string            `yaml:"method"`
	Order                          int               `yaml:"order"`
	SDKMethods                     []string          `yaml:"sdk_methods"`
	DelegateTo                     string            `yaml:"delegate_to"`
	DelegateCallArgs               []string          `yaml:"delegate_call_args"`
	AsyncDelegateCallArgs          []string          `yaml:"async_delegate_call_args"`
	DelegateAsyncYield             bool              `yaml:"delegate_async_yield"`
	SyncOnly                       bool              `yaml:"sync_only"`
	AsyncOnly                      bool              `yaml:"async_only"`
	AllowMissingInSwagger          bool              `yaml:"allow_missing_in_swagger"`
	HTTPMethodOverride             string            `yaml:"http_method_override"`
	DisableRequestBody             bool              `yaml:"disable_request_body"`
	BodyFields                     []string          `yaml:"body_fields"`
	BodyFixedValues                map[string]string `yaml:"body_fixed_values"`
	BodyBuilder                    string            `yaml:"body_builder"`
	FilesFields                    []string          `yaml:"files_fields"`
	FilesFieldValues               map[string]string `yaml:"files_field_values"`
	FilesExpr                      string            `yaml:"files_expr"`
	FilesBeforeBody                bool              `yaml:"files_before_body"`
	BodyAnnotation                 string            `yaml:"body_annotation"`
	PreDocstringCode               []string          `yaml:"pre_docstring_code"`
	PreBodyCode                    []string          `yaml:"pre_body_code"`
	BodyRequiredFields             []string          `yaml:"body_required_fields"`
	ParamAliases                   map[string]string `yaml:"param_aliases"`
	ArgTypes                       map[string]string `yaml:"arg_types"`
	ResponseType                   string            `yaml:"response_type"`
	AsyncResponseType              string            `yaml:"async_response_type"`
	ResponseCast                   string            `yaml:"response_cast"`
	StreamKeyword                  bool              `yaml:"stream_keyword"`
	QueryFields                    []OperationField  `yaml:"query_fields"`
	QueryFieldValues               map[string]string `yaml:"query_field_values"`
	SignatureQueryFields           []string          `yaml:"signature_query_fields"`
	ArgDefaults                    map[string]string `yaml:"arg_defaults"`
	ArgDefaultsSync                map[string]string `yaml:"arg_defaults_sync"`
	ArgDefaultsAsync               map[string]string `yaml:"arg_defaults_async"`
	Pagination                     string            `yaml:"pagination"`
	PaginationDataClass            string            `yaml:"pagination_data_class"`
	PaginationItemType             string            `yaml:"pagination_item_type"`
	PaginationItemsField           string            `yaml:"pagination_items_field"`
	PaginationTotalField           string            `yaml:"pagination_total_field"`
	PaginationHasMoreField         string            `yaml:"pagination_has_more_field"`
	PaginationNextTokenField       string            `yaml:"pagination_next_token_field"`
	PaginationPageNumField         string            `yaml:"pagination_page_num_field"`
	PaginationPageSizeField        string            `yaml:"pagination_page_size_field"`
	PaginationPageTokenField       string            `yaml:"pagination_page_token_field"`
	PaginationCastBeforeHeaders    bool              `yaml:"pagination_cast_before_headers"`
	AsyncIncludeKwargs             bool              `yaml:"async_include_kwargs"`
	IgnoreHeaderParams             bool              `yaml:"ignore_header_params"`
	DataField                      string            `yaml:"data_field"`
	ResponseUnwrapListFirst        bool              `yaml:"response_unwrap_list_first"`
	RequestStream                  bool              `yaml:"request_stream"`
	StreamWrap                     bool              `yaml:"stream_wrap"`
	StreamWrapHandler              string            `yaml:"stream_wrap_handler"`
	StreamWrapFields               []string          `yaml:"stream_wrap_fields"`
	StreamWrapAsyncYield           bool              `yaml:"stream_wrap_async_yield"`
	StreamWrapSyncResponseVar      string            `yaml:"stream_wrap_sync_response_var"`
	StreamWrapCompactAsyncReturn   bool              `yaml:"stream_wrap_compact_async_return"`
	StreamWrapCompactSyncReturn    bool              `yaml:"stream_wrap_compact_sync_return"`
	StreamWrapBlankLineBeforeAsync bool              `yaml:"stream_wrap_blank_line_before_return_async"`
	QueryBuilder                   string            `yaml:"query_builder"`
	QueryBuilderSync               string            `yaml:"query_builder_sync"`
	QueryBuilderAsync              string            `yaml:"query_builder_async"`
	BodyCallExpr                   string            `yaml:"body_call_expr"`
	BodyFieldValues                map[string]string `yaml:"body_field_values"`
	HeadersExpr                    string            `yaml:"headers_expr"`
	CompactSingleItemMaps          bool              `yaml:"compact_single_item_maps"`
	CompactSingleItemMapsSync      bool              `yaml:"compact_single_item_maps_sync"`
	CompactSingleItemMapsAsync     bool              `yaml:"compact_single_item_maps_async"`
	BlankLineAfterHeaders          bool              `yaml:"blank_line_after_headers"`
	NoBlankLineAfterHeaders        bool              `yaml:"no_blank_line_after_headers"`
	BlankLineAfterDocstring        bool              `yaml:"blank_line_after_docstring"`
	BlankLineBeforeReturn          bool              `yaml:"blank_line_before_return"`
	ForceMultilineSignature        bool              `yaml:"force_multiline_signature"`
	ForceMultilineSignatureSync    bool              `yaml:"force_multiline_signature_sync"`
	ForceMultilineSignatureAsync   bool              `yaml:"force_multiline_signature_async"`
	ForceMultilineRequestCall      bool              `yaml:"force_multiline_request_call"`
	ForceMultilineRequestCallSync  bool              `yaml:"force_multiline_request_call_sync"`
	ForceMultilineRequestCallAsync bool              `yaml:"force_multiline_request_call_async"`
	RequestCallArgOrder            []string          `yaml:"request_call_arg_order"`
	PaginationHTTPMethod           string            `yaml:"pagination_http_method"`
	PaginationRequestArg           string            `yaml:"pagination_request_arg"`
	PaginationInitPageTokenExpr    string            `yaml:"pagination_init_page_token_expr"`
	PaginationParamsVariable       bool              `yaml:"pagination_params_variable"`
}

type OperationField struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
	Default  string `yaml:"default"`
	UseValue bool   `yaml:"use_value"`
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
	cfg, err := Parse(content)
	if err != nil {
		return nil, err
	}
	if err := cfg.loadCommentOverrides(path); err != nil {
		return nil, err
	}
	return cfg, nil
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
	normalizedByLanguage := map[string][]string{}
	for rawLang, paths := range c.Diff.IgnorePathsByLanguage {
		lang := normalizeLanguage(rawLang)
		if lang == "" {
			continue
		}
		normalizedByLanguage[lang] = append(normalizedByLanguage[lang], paths...)
	}
	c.Diff.IgnorePathsByLanguage = normalizedByLanguage

	for lang, defaults := range defaultDiffIgnorePathsByLanguage {
		paths := c.Diff.IgnorePathsByLanguage[lang]
		if len(paths) == 0 {
			paths = append([]string(nil), defaults...)
		}
		paths = normalizeDiffPaths(paths)
		if !containsPath(paths, ".git") {
			paths = append([]string{".git"}, paths...)
		}
		c.Diff.IgnorePathsByLanguage[lang] = paths
	}
	if c.API.FieldAliases == nil {
		c.API.FieldAliases = map[string]map[string]string{}
	}
	c.CommentOverrides.ensureMaps()
	for i := range c.API.OperationMappings {
		if strings.TrimSpace(c.API.OperationMappings[i].QueryBuilder) == "" {
			c.API.OperationMappings[i].QueryBuilder = "dump_exclude_none"
		}
		if strings.TrimSpace(c.API.OperationMappings[i].BodyBuilder) == "" {
			c.API.OperationMappings[i].BodyBuilder = "dump_exclude_none"
		}
	}
}

func (c *Config) loadCommentOverrides(configPath string) error {
	c.CommentOverrides.ensureMaps()
	overridesPath := strings.TrimSpace(c.CommentOverridesFile)
	if overridesPath == "" {
		return nil
	}
	if !filepath.IsAbs(overridesPath) {
		overridesPath = filepath.Join(filepath.Dir(configPath), overridesPath)
	}
	content, err := os.ReadFile(overridesPath)
	if err != nil {
		return fmt.Errorf("read comment_overrides_file %q: %w", overridesPath, err)
	}
	var overrides CommentOverrides
	if err := yaml.Unmarshal(content, &overrides); err != nil {
		return fmt.Errorf("parse comment_overrides_file %q: %w", overridesPath, err)
	}
	overrides.ensureMaps()
	c.CommentOverrides = overrides
	return nil
}

func (c *CommentOverrides) ensureMaps() {
	if c.ClassDocstrings == nil {
		c.ClassDocstrings = map[string]string{}
	}
	if c.ClassDocstringStyles == nil {
		c.ClassDocstringStyles = map[string]string{}
	}
	if c.MethodDocstrings == nil {
		c.MethodDocstrings = map[string]string{}
	}
	if c.MethodDocstringStyles == nil {
		c.MethodDocstringStyles = map[string]string{}
	}
	if c.FieldComments == nil {
		c.FieldComments = map[string][]string{}
	}
	if c.InlineFieldComments == nil {
		c.InlineFieldComments = map[string]string{}
	}
	if c.EnumMemberComments == nil {
		c.EnumMemberComments = map[string][]string{}
	}
	if c.InlineEnumMemberComment == nil {
		c.InlineEnumMemberComment = map[string]string{}
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.Language) != "" {
		lang := strings.ToLower(strings.TrimSpace(c.Language))
		if lang != "python" && lang != "go" {
			return fmt.Errorf("unsupported language %q, supported languages: python, go", c.Language)
		}
	}

	for lang, paths := range c.Diff.IgnorePathsByLanguage {
		normalizedLang := normalizeLanguage(lang)
		if normalizedLang != "python" && normalizedLang != "go" {
			return fmt.Errorf("diff.ignore_paths_by_language.%s is unsupported, supported languages: python, go", lang)
		}
		for i, path := range paths {
			trimmed := strings.TrimSpace(path)
			if trimmed == "" {
				return fmt.Errorf("diff.ignore_paths_by_language.%s[%d] should not be empty", normalizedLang, i)
			}
			if trimmed == "." || trimmed == ".." {
				return fmt.Errorf("diff.ignore_paths_by_language.%s[%d] is invalid: %q", normalizedLang, i, path)
			}
		}
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
		}
		for j, imp := range pkg.ExtraImports {
			if strings.TrimSpace(imp.Module) == "" {
				return fmt.Errorf("api.packages[%d].extra_imports[%d].module is required", i, j)
			}
			if len(imp.Names) == 0 {
				return fmt.Errorf("api.packages[%d].extra_imports[%d].names should not be empty", i, j)
			}
			for k, name := range imp.Names {
				if strings.TrimSpace(name) == "" {
					return fmt.Errorf("api.packages[%d].extra_imports[%d].names[%d] is required", i, j, k)
				}
			}
		}
		for j, rawImport := range pkg.RawImports {
			if strings.TrimSpace(rawImport) == "" {
				return fmt.Errorf("api.packages[%d].raw_imports[%d] should not be empty", i, j)
			}
		}
		for j, block := range pkg.TopLevelCode {
			if strings.TrimSpace(block) == "" {
				return fmt.Errorf("api.packages[%d].top_level_code[%d] should not be empty", i, j)
			}
		}
		for j, block := range pkg.SyncExtraMethods {
			if strings.TrimSpace(block) == "" {
				return fmt.Errorf("api.packages[%d].sync_extra_methods[%d] should not be empty", i, j)
			}
		}
		for j, block := range pkg.AsyncExtraMethods {
			if strings.TrimSpace(block) == "" {
				return fmt.Errorf("api.packages[%d].async_extra_methods[%d] should not be empty", i, j)
			}
		}
		for j, className := range pkg.OverridePaginationClasses {
			if strings.TrimSpace(className) == "" {
				return fmt.Errorf("api.packages[%d].override_pagination_classes[%d] should not be empty", i, j)
			}
		}
		for j, model := range pkg.ModelSchemas {
			if strings.TrimSpace(model.Name) == "" {
				return fmt.Errorf("api.packages[%d].model_schemas[%d].name is required", i, j)
			}
			if strings.TrimSpace(model.Schema) == "" && !model.AllowMissingInSwagger {
				return fmt.Errorf("api.packages[%d].model_schemas[%d].schema is required when allow_missing_in_swagger is false", i, j)
			}
			if strings.TrimSpace(model.EnumBase) != "" {
				switch strings.TrimSpace(model.EnumBase) {
				case "dynamic_str", "int", "int_enum":
				default:
					return fmt.Errorf("api.packages[%d].model_schemas[%d].enum_base must be 'dynamic_str', 'int' or 'int_enum' when set", i, j)
				}
			}
			for fieldName, fieldType := range model.FieldTypes {
				if strings.TrimSpace(fieldName) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].field_types has empty key", i, j)
				}
				if strings.TrimSpace(fieldType) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].field_types[%q] is empty", i, j, fieldName)
				}
			}
			for k, extra := range model.ExtraFields {
				if strings.TrimSpace(extra.Name) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].extra_fields[%d].name is required", i, j, k)
				}
				if strings.TrimSpace(extra.Type) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].extra_fields[%d].type is required", i, j, k)
				}
			}
			for k, block := range model.ExtraCode {
				if strings.TrimSpace(block) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].extra_code[%d] should not be empty", i, j, k)
				}
			}
			for k, block := range model.PrependCode {
				if strings.TrimSpace(block) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].prepend_code[%d] should not be empty", i, j, k)
				}
			}
			for k, baseClass := range model.BaseClasses {
				if strings.TrimSpace(baseClass) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].base_classes[%d] should not be empty", i, j, k)
				}
			}
			for k, enumValue := range model.EnumValues {
				if strings.TrimSpace(enumValue.Name) == "" {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].enum_values[%d].name is required", i, j, k)
				}
				if enumValue.Value == nil {
					return fmt.Errorf("api.packages[%d].model_schemas[%d].enum_values[%d].value is required", i, j, k)
				}
			}
		}
	}

	for i, mapping := range c.API.OperationMappings {
		if err := validateOperationRef(mapping.Path, mapping.Method, fmt.Sprintf("api.operation_mappings[%d]", i)); err != nil {
			return err
		}
		if mapping.SyncOnly && mapping.AsyncOnly {
			return fmt.Errorf("api.operation_mappings[%d] cannot set both sync_only and async_only", i)
		}
		if strings.TrimSpace(mapping.HTTPMethodOverride) != "" {
			if err := validateOperationRef(mapping.Path, mapping.HTTPMethodOverride, fmt.Sprintf("api.operation_mappings[%d].http_method_override", i)); err != nil {
				return err
			}
		}
		if len(mapping.SDKMethods) == 0 {
			return fmt.Errorf("api.operation_mappings[%d].sdk_methods should not be empty", i)
		}
		if strings.TrimSpace(mapping.DelegateTo) == "" && len(mapping.DelegateCallArgs) > 0 {
			return fmt.Errorf("api.operation_mappings[%d].delegate_call_args requires delegate_to", i)
		}
		if strings.TrimSpace(mapping.DelegateTo) == "" && len(mapping.AsyncDelegateCallArgs) > 0 {
			return fmt.Errorf("api.operation_mappings[%d].async_delegate_call_args requires delegate_to", i)
		}
		if strings.TrimSpace(mapping.DelegateTo) == "" && mapping.DelegateAsyncYield {
			return fmt.Errorf("api.operation_mappings[%d].delegate_async_yield requires delegate_to", i)
		}
		for j, arg := range mapping.DelegateCallArgs {
			if strings.TrimSpace(arg) == "" {
				return fmt.Errorf("api.operation_mappings[%d].delegate_call_args[%d] should not be empty", i, j)
			}
		}
		for j, arg := range mapping.AsyncDelegateCallArgs {
			if strings.TrimSpace(arg) == "" {
				return fmt.Errorf("api.operation_mappings[%d].async_delegate_call_args[%d] should not be empty", i, j)
			}
		}
		for j, field := range mapping.QueryFields {
			if strings.TrimSpace(field.Name) == "" {
				return fmt.Errorf("api.operation_mappings[%d].query_fields[%d].name is required", i, j)
			}
		}
		for fieldName, fieldValue := range mapping.QueryFieldValues {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].query_field_values has empty key", i)
			}
			if strings.TrimSpace(fieldValue) == "" {
				return fmt.Errorf("api.operation_mappings[%d].query_field_values[%q] is empty", i, fieldName)
			}
		}
		for j, fieldName := range mapping.SignatureQueryFields {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].signature_query_fields[%d] should not be empty", i, j)
			}
		}
		for fieldName, fieldValue := range mapping.ArgDefaults {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].arg_defaults has empty key", i)
			}
			if strings.TrimSpace(fieldValue) == "" {
				return fmt.Errorf("api.operation_mappings[%d].arg_defaults[%q] is empty", i, fieldName)
			}
		}
		for fieldName, fieldValue := range mapping.BodyFixedValues {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].body_fixed_values has empty key", i)
			}
			if strings.TrimSpace(fieldValue) == "" {
				return fmt.Errorf("api.operation_mappings[%d].body_fixed_values[%q] is empty", i, fieldName)
			}
		}
		for fieldName, fieldValue := range mapping.BodyFieldValues {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].body_field_values has empty key", i)
			}
			if strings.TrimSpace(fieldValue) == "" {
				return fmt.Errorf("api.operation_mappings[%d].body_field_values[%q] is empty", i, fieldName)
			}
		}
		for j, fileField := range mapping.FilesFields {
			if strings.TrimSpace(fileField) == "" {
				return fmt.Errorf("api.operation_mappings[%d].files_fields[%d] should not be empty", i, j)
			}
		}
		for fieldName, fieldValue := range mapping.FilesFieldValues {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].files_field_values has empty key", i)
			}
			if strings.TrimSpace(fieldValue) == "" {
				return fmt.Errorf("api.operation_mappings[%d].files_field_values[%q] is empty", i, fieldName)
			}
		}
		for j, codeBlock := range mapping.PreBodyCode {
			if strings.TrimSpace(codeBlock) == "" {
				return fmt.Errorf("api.operation_mappings[%d].pre_body_code[%d] should not be empty", i, j)
			}
		}
		for j, fieldName := range mapping.StreamWrapFields {
			if strings.TrimSpace(fieldName) == "" {
				return fmt.Errorf("api.operation_mappings[%d].stream_wrap_fields[%d] should not be empty", i, j)
			}
		}
		if mapping.StreamWrap && !mapping.RequestStream {
			return fmt.Errorf("api.operation_mappings[%d].stream_wrap requires request_stream=true", i)
		}
		for j, argName := range mapping.RequestCallArgOrder {
			switch strings.TrimSpace(argName) {
			case "stream", "cast", "params", "headers", "body", "files", "data_field":
			default:
				return fmt.Errorf("api.operation_mappings[%d].request_call_arg_order[%d] must be one of: stream, cast, params, headers, body, files, data_field", i, j)
			}
		}
		switch strings.TrimSpace(mapping.QueryBuilder) {
		case "", "dump_exclude_none", "remove_none_values", "raw":
		default:
			return fmt.Errorf("api.operation_mappings[%d].query_builder must be one of: dump_exclude_none, remove_none_values, raw", i)
		}
		switch strings.TrimSpace(mapping.BodyBuilder) {
		case "", "dump_exclude_none", "remove_none_values", "raw":
		default:
			return fmt.Errorf("api.operation_mappings[%d].body_builder must be one of: dump_exclude_none, remove_none_values, raw", i)
		}
		if strings.TrimSpace(mapping.Pagination) != "" {
			pagination := strings.TrimSpace(mapping.Pagination)
			if pagination != "token" && pagination != "number" && pagination != "number_has_more" {
				return fmt.Errorf("api.operation_mappings[%d].pagination must be 'token', 'number', or 'number_has_more' when set", i)
			}
			if strings.TrimSpace(mapping.PaginationDataClass) == "" || strings.TrimSpace(mapping.PaginationItemType) == "" {
				return fmt.Errorf("api.operation_mappings[%d].pagination_data_class and pagination_item_type are required for pagination", i)
			}
		}
	}

	for i, ref := range c.API.IgnoreAPIs {
		if err := validateOperationRef(ref.Path, ref.Method, fmt.Sprintf("api.ignore_apis[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

func containsPath(paths []string, target string) bool {
	target = strings.TrimSpace(target)
	for _, path := range paths {
		if strings.TrimSpace(path) == target {
			return true
		}
	}
	return false
}

func normalizeDiffPaths(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func normalizeLanguage(language string) string {
	return strings.ToLower(strings.TrimSpace(language))
}

func (c *Config) DiffIgnorePathsForLanguage(language string) []string {
	if c == nil {
		return []string{".git"}
	}
	lang := normalizeLanguage(language)
	if paths := c.Diff.IgnorePathsByLanguage[lang]; len(paths) > 0 {
		return append([]string(nil), paths...)
	}
	if defaults := defaultDiffIgnorePathsByLanguage[lang]; len(defaults) > 0 {
		return append([]string(nil), defaults...)
	}
	return []string{".git"}
}

func (c *Config) IsIgnored(path string, method string) bool {
	method = normalizeMethod(method)
	for _, ref := range c.API.IgnoreAPIs {
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
		for _, op := range c.API.IgnoreAPIs {
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

	for _, op := range c.API.IgnoreAPIs {
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

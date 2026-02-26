package generator

import (
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	pygen "github.com/coze-dev/coze-sdk-gen/internal/generator/python"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestRenderOperationMethodUsesSwaggerComments(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:        "/v1/demo/{demo_id}",
		Method:      "get",
		Summary:     "Retrieve demo",
		Description: "Retrieve detail for one demo.",
		PathParameters: []openapi.ParameterSpec{
			{
				Name:        "demo_id",
				In:          "path",
				Description: "Demo identifier",
				Required:    true,
				Schema:      &openapi.Schema{Type: "string"},
			},
		},
		QueryParameters: []openapi.ParameterSpec{
			{
				Name:        "expand",
				In:          "query",
				Description: "Include extra fields",
				Required:    false,
				Schema:      &openapi.Schema{Type: "boolean"},
			},
		},
		Response:       &openapi.Response{Description: "Demo object"},
		ResponseSchema: &openapi.Schema{Type: "object"},
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "retrieve",
		Details:     details,
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(code, "Retrieve demo") {
		t.Fatalf("expected summary in docstring:\n%s", code)
	}
	if !strings.Contains(code, "Retrieve detail for one demo.") {
		t.Fatalf("expected description in docstring:\n%s", code)
	}
	if !strings.Contains(code, ":param demo_id: Demo identifier") {
		t.Fatalf("expected path param description in docstring:\n%s", code)
	}
	if !strings.Contains(code, ":param expand: Include extra fields") {
		t.Fatalf("expected query param description in docstring:\n%s", code)
	}
	if !strings.Contains(code, ":return: Demo object") {
		t.Fatalf("expected return description in docstring:\n%s", code)
	}
}

func TestRenderOperationMethodSanitizesRichTextDescription(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:        "/v1/demo",
		Method:      "post",
		Description: `{"0":{"ops":[{"insert":"first line"},{"insert":"second line"}],"zoneType":"Z"}}`,
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "create",
		Details:     details,
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(code, "first line second line") {
		t.Fatalf("expected rich text payload converted to plain text:\n%s", code)
	}
	if strings.Contains(code, "\"ops\"") {
		t.Fatalf("did not expect raw rich text JSON in docstring:\n%s", code)
	}
}

func TestRenderPackageModuleUsesSwaggerFieldDescriptions(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"DemoModel": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"id": {
							Type:        "string",
							Description: "Demo identifier.",
						},
						"name": {
							Type:        "string",
							Description: "Demo display name.",
						},
					},
					Required: []string{"id"},
				},
			},
		},
	}
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Schema: "DemoModel",
					Name:   "DemoModel",
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "# Demo identifier.\n    id: str") {
		t.Fatalf("expected required field description comment:\n%s", code)
	}
	if !strings.Contains(code, "# Demo display name.\n    name: Optional[str] = None") {
		t.Fatalf("expected optional field description comment:\n%s", code)
	}
}

func TestRenderPackageModuleFieldCommentsPreferSwaggerFallbackOverrides(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"DemoModel": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"id": {
							Type:        "string",
							Description: "Swagger id description.",
						},
						"fallback_field": {
							Type: "string",
						},
					},
					Required: []string{"id"},
				},
			},
		},
	}
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Schema: "DemoModel",
					Name:   "DemoModel",
					FieldOrder: []string{
						"id",
						"fallback_field",
						"legacy_only",
					},
					ExtraFields: []config.ModelField{
						{Name: "legacy_only", Type: "str", Required: true},
					},
				},
			},
		},
	}
	commentOverrides := config.CommentOverrides{
		FieldComments: map[string][]string{
			"cozepy.demo.DemoModel.id":             {"Override id description."},
			"cozepy.demo.DemoModel.fallback_field": {"Override fallback description."},
			"cozepy.demo.DemoModel.legacy_only":    {"Legacy field description."},
		},
	}

	code := pygen.RenderPackageModuleWithComments(doc, meta, nil, commentOverrides)
	if !strings.Contains(code, "# Swagger id description.\n    id: str") {
		t.Fatalf("expected swagger description for id field:\n%s", code)
	}
	if strings.Contains(code, "Override id description.") {
		t.Fatalf("did not expect id override to replace swagger description:\n%s", code)
	}
	if !strings.Contains(code, "# Override fallback description.\n    fallback_field: Optional[str] = None") {
		t.Fatalf("expected fallback override description for swagger-missing field:\n%s", code)
	}
	if !strings.Contains(code, "# Legacy field description.\n    legacy_only: str") {
		t.Fatalf("expected fallback override description for non-swagger extra field:\n%s", code)
	}
}

func TestRenderPackageModuleMethodDocstringPreferSwaggerFallbackOverrides(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ClientClass:      "DemoClient",
			AsyncClientClass: "AsyncDemoClient",
		},
	}
	bindings := []pygen.OperationBinding{
		{
			PackageName: "demo",
			MethodName:  "create",
			Details: openapi.OperationDetails{
				Path:   "/v1/demo",
				Method: "post",
			},
		},
		{
			PackageName: "demo",
			MethodName:  "retrieve",
			Details: openapi.OperationDetails{
				Path:    "/v1/demo/{id}",
				Method:  "get",
				Summary: "Retrieve from swagger.",
				PathParameters: []openapi.ParameterSpec{
					{Name: "id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
				},
			},
		},
	}
	commentOverrides := config.CommentOverrides{
		MethodDocstrings: map[string]string{
			"cozepy.demo.DemoClient.create":   "Create from overrides.",
			"cozepy.demo.DemoClient.retrieve": "Retrieve from overrides.",
		},
		MethodDocstringStyles: map[string]string{
			"cozepy.demo.DemoClient.create":   "block",
			"cozepy.demo.DemoClient.retrieve": "block",
		},
	}

	code := pygen.RenderPackageModuleWithComments(doc, meta, bindings, commentOverrides)
	if !strings.Contains(code, "Create from overrides.") {
		t.Fatalf("expected override docstring when swagger comment is missing:\n%s", code)
	}
	if !strings.Contains(code, "Retrieve from swagger.") {
		t.Fatalf("expected swagger docstring to be used when available:\n%s", code)
	}
	if strings.Contains(code, "Retrieve from overrides.") {
		t.Fatalf("did not expect override docstring to replace swagger docstring:\n%s", code)
	}
}

func TestRenderPackageModuleMethodDocstringUsesRichTextOverrides(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ClientClass: "DemoClient",
		},
	}
	bindings := []pygen.OperationBinding{
		{
			PackageName: "demo",
			MethodName:  "create",
			Details: openapi.OperationDetails{
				Path:        "/v1/demo",
				Method:      "post",
				Description: `{"0":{"ops":[{"insert":"first line"},{"insert":"second line"}],"zoneType":"Z"}}`,
			},
		},
	}
	commentOverrides := config.CommentOverrides{
		RichTextMethodDocstrings: map[string]string{
			"cozepy.demo.DemoClient.create":      "Use rich text override.",
			"cozepy.demo.AsyncDemoClient.create": "Use rich text override.",
		},
		MethodDocstringStyles: map[string]string{
			"cozepy.demo.DemoClient.create":      "block",
			"cozepy.demo.AsyncDemoClient.create": "block",
		},
	}

	code := pygen.RenderPackageModuleWithComments(doc, meta, bindings, commentOverrides)
	if !strings.Contains(code, "Use rich text override.") {
		t.Fatalf("expected rich text override docstring:\n%s", code)
	}
	if strings.Contains(code, "first line second line") {
		t.Fatalf("did not expect raw swagger rich text docstring when override exists:\n%s", code)
	}
}

func TestRenderOperationMethodBlankLineAfterHeaders(t *testing.T) {
	doc := mustParseSwagger(t)
	details := openapi.OperationDetails{
		Path:   "/v1/demo/{id}",
		Method: "post",
		PathParameters: []openapi.ParameterSpec{
			{Name: "id", In: "path", Required: true, Schema: &openapi.Schema{Type: "string"}},
		},
		RequestBodySchema: &openapi.Schema{
			Type: "object",
			Properties: map[string]*openapi.Schema{
				"name": {Type: "string"},
			},
			Required: []string{"name"},
		},
	}
	binding := pygen.OperationBinding{
		PackageName: "demo",
		MethodName:  "update",
		Details:     details,
		Mapping: &config.OperationMapping{
			BodyBuilder:           "raw",
			BodyFields:            []string{"name"},
			BlankLineAfterHeaders: true,
			ArgTypes: map[string]string{
				"id":   "str",
				"name": "str",
			},
		},
	}

	code := pygen.RenderOperationMethod(doc, binding, false)
	if !strings.Contains(code, "headers: Optional[dict] = kwargs.get(\"headers\")\n\n        body = {") {
		t.Fatalf("expected blank line between headers assignment and body when blank_line_after_headers=true:\n%s", code)
	}
}

func TestRenderPackageModuleIntEnumMixinBase(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "AuditStatus",
					AllowMissingInSwagger: true,
					EnumBase:              "int_enum",
					EnumValues: []config.ModelEnumValue{
						{Name: "PENDING", Value: 1},
						{Name: "APPROVED", Value: 2},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "from enum import Enum") {
		t.Fatalf("expected Enum import for int_enum base:\n%s", code)
	}
	if strings.Contains(code, "from enum import IntEnum") {
		t.Fatalf("did not expect IntEnum import for int_enum base:\n%s", code)
	}
	if !strings.Contains(code, "class AuditStatus(int, Enum):") {
		t.Fatalf("expected int+Enum class base for int_enum model:\n%s", code)
	}
}

func TestRenderPackageModuleAutoAddsReferencedSchemasForSameName(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"OpenAPIVoiceData": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"support_emotions": {
							Type:  "array",
							Items: &openapi.Schema{Ref: "#/components/schemas/EmotionInfo"},
						},
						"state": {Ref: "#/components/schemas/OpenAPIVoiceState"},
					},
				},
				"EmotionInfo": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"emotion_scale_interval": {Ref: "#/components/schemas/Interval"},
					},
				},
				"Interval": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"max": {Type: "number"},
					},
				},
				"OpenAPIVoiceState": {
					Type: "string",
					Enum: []interface{}{"init"},
				},
			},
		},
	}
	meta := pygen.PackageMeta{
		Name:       "audio_voices",
		ModulePath: "audio.voices",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "VoiceState",
					AllowMissingInSwagger: true,
					EnumBase:              "dynamic_str",
					EnumValues: []config.ModelEnumValue{
						{Name: "INIT", Value: "init"},
					},
				},
				{
					Schema: "OpenAPIVoiceData",
					Name:   "Voice",
					FieldTypes: map[string]string{
						"state": "VoiceState",
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "support_emotions: Optional[List[EmotionInfo]] = None") {
		t.Fatalf("expected support_emotions to keep model type, got:\n%s", code)
	}
	if strings.Contains(code, "List[Dict[str, Any]]") {
		t.Fatalf("did not expect support_emotions to fallback to Dict type, got:\n%s", code)
	}
	idxInterval := strings.Index(code, "class Interval(CozeModel):")
	idxEmotionInfo := strings.Index(code, "class EmotionInfo(CozeModel):")
	idxVoice := strings.Index(code, "class Voice(CozeModel):")
	if idxInterval < 0 || idxEmotionInfo < 0 || idxVoice < 0 {
		t.Fatalf("expected Interval/EmotionInfo/Voice classes, got:\n%s", code)
	}
	if !(idxInterval < idxEmotionInfo && idxEmotionInfo < idxVoice) {
		t.Fatalf("expected dependency classes before Voice, got:\n%s", code)
	}
	if strings.Contains(code, "class OpenApiVoiceState(") {
		t.Fatalf("did not expect auto-generated OpenApiVoiceState when field type override is set, got:\n%s", code)
	}
}

func TestRenderPackageModulePrunesUnorderedSchemaDependencies(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"BotInfo": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"used_info":    {Ref: "#/components/schemas/UsedInfo"},
						"unused_info":  {Ref: "#/components/schemas/ShortcutCommandInfo"},
						"display_name": {Type: "string"},
					},
				},
				"UsedInfo": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"id": {Type: "string"},
					},
				},
				"ShortcutCommandInfo": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"id": {Type: "string"},
					},
				},
			},
		},
	}
	meta := pygen.PackageMeta{
		Name:       "bots",
		ModulePath: "bots",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Schema:                 "BotInfo",
					Name:                   "Bot",
					ExcludeUnorderedFields: true,
					FieldOrder:             []string{"used_info", "display_name"},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "used_info: Optional[UsedInfo] = None") {
		t.Fatalf("expected used_info typed with referenced UsedInfo model, got:\n%s", code)
	}
	if strings.Contains(code, "class ShortcutCommandInfo(CozeModel):") {
		t.Fatalf("did not expect pruned dependency ShortcutCommandInfo to be generated, got:\n%s", code)
	}
	if !strings.Contains(code, "class UsedInfo(CozeModel):") {
		t.Fatalf("expected used dependency UsedInfo to be generated, got:\n%s", code)
	}
}

func TestRenderPackageModuleAutoNumberPagedResponseMethods(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "_PrivateListWorkflowData",
					AllowMissingInSwagger: true,
					BaseClasses:           []string{"CozeModel", "NumberPagedResponse[WorkflowBasic]"},
					ExtraFields: []config.ModelField{
						{Name: "items", Type: "List[WorkflowBasic]", Required: true},
						{Name: "has_more", Type: "bool", Required: true},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "class _PrivateListWorkflowData(CozeModel, NumberPagedResponse[WorkflowBasic]):") {
		t.Fatalf("expected paged model class definition:\n%s", code)
	}
	if !strings.Contains(code, "def get_total(self) -> Optional[int]:\n        return None") {
		t.Fatalf("expected auto generated get_total for number_has_more paged model:\n%s", code)
	}
	if !strings.Contains(code, "def get_has_more(self) -> Optional[bool]:\n        return self.has_more") {
		t.Fatalf("expected auto generated get_has_more for number_has_more paged model:\n%s", code)
	}
	if !strings.Contains(code, "def get_items(self) -> List[WorkflowBasic]:\n        return self.items") {
		t.Fatalf("expected auto generated get_items for number_has_more paged model:\n%s", code)
	}
}

func TestRenderPackageModuleAutoNumberPagedResponseMethodsWithCustomFieldNames(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "_PrivateListDatasetsData",
					AllowMissingInSwagger: true,
					BaseClasses:           []string{"CozeModel", "NumberPagedResponse[Dataset]"},
					ExtraFields: []config.ModelField{
						{Name: "total_count", Type: "int", Required: true},
						{Name: "dataset_list", Type: "List[Dataset]", Required: true},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "def get_total(self) -> Optional[int]:\n        return self.total_count") {
		t.Fatalf("expected auto generated get_total from total_count:\n%s", code)
	}
	if !strings.Contains(code, "def get_has_more(self) -> Optional[bool]:\n        return None") {
		t.Fatalf("expected auto generated get_has_more=None when has_more is absent:\n%s", code)
	}
	if !strings.Contains(code, "def get_items(self) -> List[Dataset]:\n        return self.dataset_list") {
		t.Fatalf("expected auto generated get_items from dataset_list:\n%s", code)
	}
}

func TestRenderPackageModuleAutoTokenAndLastIDPagedResponseMethods(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "_PrivateListWorkflowVersionData",
					AllowMissingInSwagger: true,
					BaseClasses:           []string{"CozeModel", "TokenPagedResponse[WorkflowVersionInfo]"},
					ExtraFields: []config.ModelField{
						{Name: "items", Type: "List[WorkflowVersionInfo]", Required: true},
						{Name: "has_more", Type: "bool", Required: true},
						{Name: "next_page_token", Type: "Optional[str]", Required: false, Default: "None"},
					},
				},
				{
					Name:                  "_PrivateListMessageResp",
					AllowMissingInSwagger: true,
					BaseClasses:           []string{"CozeModel", "LastIDPagedResponse[Message]"},
					ExtraFields: []config.ModelField{
						{Name: "first_id", Type: "str", Required: true},
						{Name: "last_id", Type: "str", Required: true},
						{Name: "has_more", Type: "bool", Required: true},
						{Name: "items", Type: "List[Message]", Required: true},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "def get_next_page_token(self) -> Optional[str]:\n        return self.next_page_token") {
		t.Fatalf("expected auto generated get_next_page_token for TokenPagedResponse:\n%s", code)
	}
	if !strings.Contains(code, "def get_items(self) -> List[WorkflowVersionInfo]:\n        return self.items") {
		t.Fatalf("expected auto generated get_items for TokenPagedResponse:\n%s", code)
	}
	if !strings.Contains(code, "def get_first_id(self) -> str:\n        return self.first_id") {
		t.Fatalf("expected auto generated get_first_id for LastIDPagedResponse:\n%s", code)
	}
	if !strings.Contains(code, "def get_last_id(self) -> str:\n        return self.last_id") {
		t.Fatalf("expected auto generated get_last_id for LastIDPagedResponse:\n%s", code)
	}
	if !strings.Contains(code, "def get_has_more(self) -> bool:\n        return self.has_more") {
		t.Fatalf("expected auto generated get_has_more for LastIDPagedResponse:\n%s", code)
	}
}

func TestRenderPackageModuleEmptyModelWithSwaggerDocstringOmitsPass(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"DeleteConversationResp": {
					Type:        "object",
					Description: "删除会话的响应结构体\n不包含任何字段，仅用于表示操作成功",
				},
			},
		},
	}
	meta := pygen.PackageMeta{
		Name:       "conversations",
		ModulePath: "conversations",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Schema: "DeleteConversationResp",
					Name:   "DeleteConversationResp",
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	classStart := strings.Index(code, "class DeleteConversationResp(CozeModel):")
	if classStart < 0 {
		t.Fatalf("expected empty model class definition:\n%s", code)
	}
	if !strings.Contains(code, "删除会话的响应结构体") {
		t.Fatalf("expected swagger class docstring content:\n%s", code)
	}
	nextClass := strings.Index(code[classStart+1:], "\nclass ")
	classBlock := code[classStart:]
	if nextClass >= 0 {
		classBlock = code[classStart : classStart+1+nextClass]
	}
	if strings.Contains(classBlock, "\n    pass\n") {
		t.Fatalf("did not expect pass when empty model already has override docstring:\n%s", classBlock)
	}
}

func TestRenderPackageModuleExtraFieldAlias(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "TranslateConfig",
					AllowMissingInSwagger: true,
					ExtraFields: []config.ModelField{
						{Name: "from_", Type: "Optional[str]", Alias: "from", Required: true},
						{Name: "to", Type: "Optional[str]", Required: true},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, `from_: Optional[str] = Field(alias="from")`) {
		t.Fatalf("expected alias field rendering for required extra field:\n%s", code)
	}
	if !strings.Contains(code, "to: Optional[str]") {
		t.Fatalf("expected normal extra field rendering without alias:\n%s", code)
	}
}

func TestRenderPackageModuleBeforeValidators(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "SimpleBot",
					AllowMissingInSwagger: true,
					ExtraFields: []config.ModelField{
						{Name: "publish_time", Type: "str", Required: true},
					},
					BeforeValidators: []config.ModelValidator{
						{Field: "publish_time", Rule: "int_to_string", Method: "convert_to_string"},
					},
				},
				{
					Name:                  "WorkflowRunHistory",
					AllowMissingInSwagger: true,
					ExtraFields: []config.ModelField{
						{Name: "error_code", Type: "int", Required: true},
					},
					BeforeValidators: []config.ModelValidator{
						{Field: "error_code", Rule: "empty_string_to_zero", Method: "error_code_empty_str_to_zero"},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, `@field_validator("publish_time", mode="before")`) {
		t.Fatalf("expected int_to_string before validator:\n%s", code)
	}
	if !strings.Contains(code, "def convert_to_string(cls, v):") {
		t.Fatalf("expected named int_to_string validator method:\n%s", code)
	}
	if !strings.Contains(code, "if isinstance(v, int):") {
		t.Fatalf("expected int_to_string conversion body:\n%s", code)
	}
	if !strings.Contains(code, `@field_validator("error_code", mode="before")`) {
		t.Fatalf("expected empty_string_to_zero before validator:\n%s", code)
	}
	if !strings.Contains(code, "def error_code_empty_str_to_zero(cls, v):") {
		t.Fatalf("expected named empty_string_to_zero validator method:\n%s", code)
	}
	if !strings.Contains(code, `if v == "":`) {
		t.Fatalf("expected empty_string_to_zero conversion body:\n%s", code)
	}
}

func TestRenderPackageModuleBuilders(t *testing.T) {
	doc := mustParseSwagger(t)
	meta := pygen.PackageMeta{
		Name:       "demo",
		ModulePath: "demo",
		Package: &config.Package{
			ModelSchemas: []config.ModelSchema{
				{
					Name:                  "DocumentUpdateRule",
					AllowMissingInSwagger: true,
					Builders: []config.ModelBuilder{
						{
							Name:       "build_no_auto_update",
							ReturnType: "DocumentUpdateRule",
							Args: []config.ModelBuilderArg{
								{Name: "update_type", Expr: "DocumentUpdateType.NO_AUTO_UPDATE"},
								{Name: "update_interval", Expr: "24"},
							},
						},
						{
							Name:       "build_auto_update",
							ReturnType: "DocumentUpdateRule",
							Params:     []string{"interval: int"},
							Args: []config.ModelBuilderArg{
								{Name: "update_type", Expr: "DocumentUpdateType.AUTO_UPDATE"},
								{Name: "update_interval", Expr: "interval"},
							},
						},
					},
				},
			},
		},
	}

	code := pygen.RenderPackageModule(doc, meta, nil)
	if !strings.Contains(code, "@staticmethod\n    def build_no_auto_update() -> \"DocumentUpdateRule\":") {
		t.Fatalf("expected builder staticmethod for no_auto_update:\n%s", code)
	}
	if !strings.Contains(code, "return DocumentUpdateRule(update_type=DocumentUpdateType.NO_AUTO_UPDATE, update_interval=24)") {
		t.Fatalf("expected builder return expression for no_auto_update:\n%s", code)
	}
	if !strings.Contains(code, "@staticmethod\n    def build_auto_update(interval: int) -> \"DocumentUpdateRule\":") {
		t.Fatalf("expected builder staticmethod for auto_update:\n%s", code)
	}
	if !strings.Contains(code, "return DocumentUpdateRule(update_type=DocumentUpdateType.AUTO_UPDATE, update_interval=interval)") {
		t.Fatalf("expected builder return expression for auto_update:\n%s", code)
	}
}

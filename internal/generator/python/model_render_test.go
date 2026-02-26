package python

import (
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestInferConfiguredModelNameForSyntheticBenefitSchemas(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"properties_data_properties_basic_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"user_level": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_basic_properties_item_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"used":     {Type: "number"},
						"total":    {Type: "number"},
						"start_at": {Type: "integer"},
						"end_at":   {Type: "integer"},
						"strategy": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_basic": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"status":    {Type: "string"},
						"item_info": {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_basic_properties_item_info"},
					},
				},
			},
		},
	}
	pkg := &config.Package{Name: "benefits"}

	cases := []struct {
		schema string
		want   string
	}{
		{schema: "properties_data_properties_basic_info", want: "BenefitBasicInfo"},
		{schema: "properties_data_properties_benefit_info_items_properties_basic_properties_item_info", want: "BenefitItemInfo"},
		{schema: "properties_data_properties_benefit_info_items_properties_basic", want: "BenefitStatusInfo"},
	}

	for _, tc := range cases {
		got := inferConfiguredModelName(doc, pkg, config.ModelSchema{Schema: tc.schema})
		if got != tc.want {
			t.Fatalf("inferConfiguredModelName(%q) = %q, want %q", tc.schema, got, tc.want)
		}
	}
}

func TestInferConfiguredModelNameNoPrefixForRegularSchema(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"OpenApiChatResp": {Type: "object"},
			},
		},
	}

	got := inferConfiguredModelName(doc, &config.Package{Name: "chat"}, config.ModelSchema{Schema: "OpenApiChatResp"})
	if got != "OpenApiChatResp" {
		t.Fatalf("inferConfiguredModelName() = %q, want %q", got, "OpenApiChatResp")
	}
}

func TestPackageSchemaAliasesUsesInferredNames(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"properties_data_properties_basic_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"user_level": {Type: "string"},
					},
				},
			},
		},
	}
	meta := PackageMeta{
		Package: &config.Package{
			Name: "benefits",
			ModelSchemas: []config.ModelSchema{
				{Schema: "properties_data_properties_basic_info"},
			},
		},
	}

	aliases := packageSchemaAliases(doc, meta)
	if got := aliases["properties_data_properties_basic_info"]; got != "BenefitBasicInfo" {
		t.Fatalf("packageSchemaAliases() = %q, want %q", got, "BenefitBasicInfo")
	}
}

func TestResolvePackageModelDefinitionsWithoutConfiguredModels(t *testing.T) {
	doc := &openapi.Document{
		Components: openapi.Components{
			Schemas: map[string]*openapi.Schema{
				"properties_data": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"basic_info":   {Ref: "#/components/schemas/properties_data_properties_basic_info"},
						"benefit_info": {Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/properties_data_properties_benefit_info_items"}},
					},
				},
				"properties_data_properties_basic_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"user_level": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"basic":        {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_basic"},
						"effective":    {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_effective"},
						"extra":        {Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_extra_items"}},
						"resource_id":  {Type: "string"},
						"benefit_type": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_basic": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"status":    {Type: "string"},
						"item_info": {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_basic_properties_item_info"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_basic_properties_item_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"used":     {Type: "number"},
						"total":    {Type: "number"},
						"start_at": {Type: "integer"},
						"end_at":   {Type: "integer"},
						"strategy": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_effective": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"status":    {Type: "string"},
						"item_info": {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_effective_properties_item_info"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_effective_properties_item_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"used":     {Type: "number"},
						"total":    {Type: "number"},
						"start_at": {Type: "integer"},
						"end_at":   {Type: "integer"},
						"strategy": {Type: "string"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_extra_items": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"status":    {Type: "string"},
						"item_info": {Ref: "#/components/schemas/properties_data_properties_benefit_info_items_properties_extra_items_properties_item_info"},
					},
				},
				"properties_data_properties_benefit_info_items_properties_extra_items_properties_item_info": {
					Type: "object",
					Properties: map[string]*openapi.Schema{
						"used":     {Type: "number"},
						"total":    {Type: "number"},
						"start_at": {Type: "integer"},
						"end_at":   {Type: "integer"},
						"strategy": {Type: "string"},
					},
				},
			},
		},
	}

	meta := PackageMeta{
		Package: &config.Package{
			Name:      "benefits",
			SourceDir: "cozepy/benefits",
		},
	}
	bindings := []OperationBinding{
		{
			PackageName: "benefits",
			Details: openapi.OperationDetails{
				ResponseSchema: &openapi.Schema{Ref: "#/components/schemas/properties_data"},
			},
			Mapping: &config.OperationMapping{
				ResponseType: "BenefitOverview",
			},
		},
	}

	models, aliases := resolvePackageModelDefinitions(doc, meta, bindings)
	if len(models) == 0 {
		t.Fatal("expected inferred model definitions")
	}
	nameCount := map[string]int{}
	for _, model := range models {
		nameCount[model.Name]++
	}

	requiredNames := []string{"BenefitOverview", "BenefitInfo", "BenefitBasicInfo", "BenefitStatusInfo", "BenefitItemInfo"}
	for _, want := range requiredNames {
		if nameCount[want] != 1 {
			t.Fatalf("expected %s to be generated once, got %d (all=%v)", want, nameCount[want], nameCount)
		}
	}

	if got := aliases["properties_data_properties_benefit_info_items_properties_effective"]; got != "BenefitStatusInfo" {
		t.Fatalf("unexpected alias for effective schema: %q", got)
	}
	if got := aliases["properties_data_properties_benefit_info_items_properties_extra_items"]; got != "BenefitStatusInfo" {
		t.Fatalf("unexpected alias for extra schema: %q", got)
	}
	if got := aliases["properties_data_properties_benefit_info_items_properties_effective_properties_item_info"]; got != "BenefitItemInfo" {
		t.Fatalf("unexpected alias for effective.item_info schema: %q", got)
	}
}

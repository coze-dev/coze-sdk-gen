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

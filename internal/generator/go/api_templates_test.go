package gogen

import (
	"strings"
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

func TestRenderGoAppsModuleUsesMappedPath(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v9/custom/apps",
					Method:     "get",
					SDKMethods: []string{"apps.list"},
				},
			},
		},
	}

	content, err := renderGoAppsModule(cfg, nil)
	if err != nil {
		t.Fatalf("renderGoAppsModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v9/custom/apps\",") {
		t.Fatalf("expected mapped path in rendered content, got:\n%s", content)
	}
}

func TestRenderGoUsersModuleFallsBackToSwaggerPath(t *testing.T) {
	doc := mustParseOpenAPIDoc(t, `
openapi: 3.0.0
paths:
  /v1/users/me:
    get:
      responses:
        '200':
          description: ok
`)

	content, err := renderGoUsersModule(&config.Config{}, doc)
	if err != nil {
		t.Fatalf("renderGoUsersModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v1/users/me\",") {
		t.Fatalf("expected swagger fallback path in rendered content, got:\n%s", content)
	}
}

func TestRenderGoTemplatesModuleConvertsCurlyPath(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			OperationMappings: []config.OperationMapping{
				{
					Path:       "/v2/templates/{template_id}/duplicate",
					Method:     "post",
					SDKMethods: []string{"templates.duplicate"},
				},
			},
		},
	}

	content, err := renderGoTemplatesModule(cfg, nil)
	if err != nil {
		t.Fatalf("renderGoTemplatesModule() error = %v", err)
	}
	if !strings.Contains(content, "URL:    \"/v2/templates/:template_id/duplicate\",") {
		t.Fatalf("expected curly path to be converted, got:\n%s", content)
	}
}

func TestListGoAPIModuleRenderersIncludesInlineRenderers(t *testing.T) {
	renderers := listGoAPIModuleRenderers()
	if len(renderers) < len(goInlineAPIModuleRenderers) {
		t.Fatalf("expected at least %d renderers, got %d", len(goInlineAPIModuleRenderers), len(renderers))
	}

	expected := map[string]struct{}{
		"apps.go":                {},
		"audio_live.go":          {},
		"audio_speech.go":        {},
		"audio_transcription.go": {},
		"chats_messages.go":      {},
		"files.go":               {},
		"templates.go":           {},
		"users.go":               {},
		"workflows_chat.go":      {},
	}
	seen := map[string]struct{}{}
	for _, r := range renderers {
		seen[r.FileName] = struct{}{}
	}
	for fileName := range expected {
		if _, ok := seen[fileName]; !ok {
			t.Fatalf("expected renderer for %s", fileName)
		}
	}
}

func TestRenderGoAudioModuleUsesSwaggerAndConfig(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Packages: []config.Package{
				{Name: "audio_rooms"},
				{
					Name: "audio_speech",
					ModelSchemas: []config.ModelSchema{
						{
							Name: "AudioFormat",
							EnumValues: []config.ModelEnumValue{
								{Name: "PCM", Value: "pcm"},
							},
						},
						{
							Name: "LanguageCode",
							EnumValues: []config.ModelEnumValue{
								{Name: "ZH", Value: "zh"},
							},
						},
					},
				},
			},
		},
	}
	doc := mustParseOpenAPIDoc(t, `
openapi: 3.0.0
paths:
  /v1/audio/speech:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                response_format:
                  type: string
                  enum: [mp3, wav]
      responses:
        '200':
          description: ok
  /v1/audio/voices/clone:
    post:
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
              properties:
                audio_format:
                  type: string
                  enum: [m4a]
      responses:
        '200':
          description: ok
`)

	content, err := renderGoAudioModule(cfg, doc)
	if err != nil {
		t.Fatalf("renderGoAudioModule() error = %v", err)
	}
	for _, expected := range []string{
		"AudioFormatPCM AudioFormat = \"pcm\"",
		"AudioFormatMP3 AudioFormat = \"mp3\"",
		"AudioFormatWAV AudioFormat = \"wav\"",
		"AudioFormatM4A AudioFormat = \"m4a\"",
		"LanguageCodeZH LanguageCode = \"zh\"",
		"Rooms *audioRooms",
		"Speech *audioSpeech",
		"Rooms: newAudioRooms(core)",
		"Speech: newAudioSpeech(core)",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("rendered audio module missing %q:\n%s", expected, content)
		}
	}
	if strings.Contains(content, "Voices *audioVoices") {
		t.Fatalf("did not expect voices child when package is not configured:\n%s", content)
	}
}

func mustParseOpenAPIDoc(t *testing.T, content string) *openapi.Document {
	t.Helper()
	doc, err := openapi.Parse([]byte(content))
	if err != nil {
		t.Fatalf("openapi.Parse() error = %v", err)
	}
	return doc
}

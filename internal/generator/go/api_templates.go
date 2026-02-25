package gogen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type goAPIModuleRenderer struct {
	FileName string
	Render   func(cfg *config.Config, doc *openapi.Document) (string, error)
}

type goTemplateModuleSpec struct {
	Asset            string
	PathReplacements []goPathReplacementSpec
}

type goPathReplacementSpec struct {
	Placeholder      string
	SDKMethod        string
	Method           string
	FallbackPath     string
	ConvertCurlyPath bool
}

var goInlineAPIModuleRenderers = []goAPIModuleRenderer{
	{FileName: "apps.go", Render: renderGoAppsModule},
	{FileName: "audio_live.go", Render: renderGoAudioLiveModule},
	{FileName: "audio_speech.go", Render: renderGoAudioSpeechModule},
	{FileName: "audio_transcription.go", Render: renderGoAudioTranscriptionsModule},
	{FileName: "chats_messages.go", Render: renderGoChatsMessagesModule},
	{FileName: "files.go", Render: renderGoFilesModule},
	{FileName: "templates.go", Render: renderGoTemplatesModule},
	{FileName: "users.go", Render: renderGoUsersModule},
	{FileName: "workflows_chat.go", Render: renderGoWorkflowsChatModule},
}

var goTemplateAPIModuleSpecs = map[string]goTemplateModuleSpec{
	"audio_rooms.go": {
		Asset: "audio_rooms.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/audio/rooms",
				SDKMethod:    "audio_rooms.create",
				Method:       "post",
				FallbackPath: "/v1/audio/rooms",
			},
		},
	},
	"audio_voiceprint_groups.go": {
		Asset: "audio_voiceprint_groups.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/audio/voiceprint_groups",
				SDKMethod:    "audio_voiceprint_groups.create",
				Method:       "post",
				FallbackPath: "/v1/audio/voiceprint_groups",
			},
			{
				Placeholder:      "/v1/audio/voiceprint_groups/:group_id",
				SDKMethod:        "audio_voiceprint_groups.update",
				Method:           "put",
				FallbackPath:     "/v1/audio/voiceprint_groups/{group_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"audio_voiceprint_groups_features.go": {
		Asset: "audio_voiceprint_groups_features.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/audio/voiceprint_groups/:group_id/features",
				SDKMethod:        "audio_voiceprint_groups_features.create",
				Method:           "post",
				FallbackPath:     "/v1/audio/voiceprint_groups/{group_id}/features",
				ConvertCurlyPath: true,
			},
			{
				Placeholder:      "/v1/audio/voiceprint_groups/:group_id/features/:feature_id",
				SDKMethod:        "audio_voiceprint_groups_features.update",
				Method:           "put",
				FallbackPath:     "/v1/audio/voiceprint_groups/{group_id}/features/{feature_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"audio_voices.go": {
		Asset: "audio_voices.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/audio/voices/clone",
				SDKMethod:    "audio_voices.clone",
				Method:       "post",
				FallbackPath: "/v1/audio/voices/clone",
			},
			{
				Placeholder:  "/v1/audio/voices",
				SDKMethod:    "audio_voices.list",
				Method:       "get",
				FallbackPath: "/v1/audio/voices",
			},
		},
	},
	"bots.go": {
		Asset: "bots.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/bot/create",
				SDKMethod:    "bots.create",
				Method:       "post",
				FallbackPath: "/v1/bot/create",
			},
			{
				Placeholder:  "/v1/bot/update",
				SDKMethod:    "bots.update",
				Method:       "post",
				FallbackPath: "/v1/bot/update",
			},
			{
				Placeholder:  "/v1/bot/publish",
				SDKMethod:    "bots.publish",
				Method:       "post",
				FallbackPath: "/v1/bot/publish",
			},
			{
				Placeholder:  "/v1/space/published_bots_list",
				SDKMethod:    "bots._list_v1",
				Method:       "get",
				FallbackPath: "/v1/space/published_bots_list",
			},
			{
				Placeholder:  "/v1/bot/get_online_info",
				SDKMethod:    "bots._retrieve_v1",
				Method:       "get",
				FallbackPath: "/v1/bot/get_online_info",
			},
			{
				Placeholder:      "/v1/bots/:bot_id",
				SDKMethod:        "bots._retrieve_v2",
				Method:           "get",
				FallbackPath:     "/v1/bots/{bot_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"chats.go": {
		Asset: "chats.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v3/chat",
				SDKMethod:    "chat.create",
				Method:       "post",
				FallbackPath: "/v3/chat",
			},
			{
				Placeholder:  "/v3/chat/cancel",
				SDKMethod:    "chat.cancel",
				Method:       "post",
				FallbackPath: "/v3/chat/cancel",
			},
			{
				Placeholder:  "/v3/chat/retrieve",
				SDKMethod:    "chat.retrieve",
				Method:       "get",
				FallbackPath: "/v3/chat/retrieve",
			},
			{
				Placeholder:  "/v3/chat/submit_tool_outputs",
				SDKMethod:    "chat.submit_tool_outputs",
				Method:       "post",
				FallbackPath: "/v3/chat/submit_tool_outputs",
			},
		},
	},
	"conversations.go": {
		Asset: "conversations.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/conversations",
				SDKMethod:    "conversations.list",
				Method:       "get",
				FallbackPath: "/v1/conversations",
			},
			{
				Placeholder:  "/v1/conversation/create",
				SDKMethod:    "conversations.create",
				Method:       "post",
				FallbackPath: "/v1/conversation/create",
			},
			{
				Placeholder:  "/v1/conversation/retrieve",
				SDKMethod:    "conversations.retrieve",
				Method:       "get",
				FallbackPath: "/v1/conversation/retrieve",
			},
			{
				Placeholder:      "/v1/conversations/:conversation_id/clear",
				SDKMethod:        "conversations.clear",
				Method:           "post",
				FallbackPath:     "/v1/conversations/{conversation_id}/clear",
				ConvertCurlyPath: true,
			},
		},
	},
	"conversations_messages.go": {
		Asset: "conversations_messages.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/conversation/message/list",
				SDKMethod:    "conversations_message.list",
				Method:       "post",
				FallbackPath: "/v1/conversation/message/list",
			},
			{
				Placeholder:  "/v1/conversation/message/create",
				SDKMethod:    "conversations_message.create",
				Method:       "post",
				FallbackPath: "/v1/conversation/message/create",
			},
			{
				Placeholder:  "/v1/conversation/message/retrieve",
				SDKMethod:    "conversations_message.retrieve",
				Method:       "get",
				FallbackPath: "/v1/conversation/message/retrieve",
			},
			{
				Placeholder:  "/v1/conversation/message/modify",
				SDKMethod:    "conversations_message.update",
				Method:       "post",
				FallbackPath: "/v1/conversation/message/modify",
			},
			{
				Placeholder:  "/v1/conversation/message/delete",
				SDKMethod:    "conversations_message.delete",
				Method:       "post",
				FallbackPath: "/v1/conversation/message/delete",
			},
		},
	},
	"conversations_messages_feedback.go": {
		Asset: "conversations_messages_feedback.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/conversations/:conversation_id/messages/:message_id/feedback",
				SDKMethod:        "conversations_message_feedback.create",
				Method:           "post",
				FallbackPath:     "/v1/conversations/{conversation_id}/messages/{message_id}/feedback",
				ConvertCurlyPath: true,
			},
		},
	},
	"datasets.go": {
		Asset: "datasets.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/datasets",
				SDKMethod:    "datasets.create",
				Method:       "post",
				FallbackPath: "/v1/datasets",
			},
			{
				Placeholder:      "/v1/datasets/:dataset_id",
				SDKMethod:        "datasets.update",
				Method:           "put",
				FallbackPath:     "/v1/datasets/{dataset_id}",
				ConvertCurlyPath: true,
			},
			{
				Placeholder:      "/v1/datasets/:dataset_id/process",
				SDKMethod:        "datasets.process",
				Method:           "post",
				FallbackPath:     "/v1/datasets/{dataset_id}/process",
				ConvertCurlyPath: true,
			},
		},
	},
	"datasets_documents.go": {
		Asset: "datasets_documents.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/open_api/knowledge/document/create",
				SDKMethod:    "datasets_documents.create",
				Method:       "post",
				FallbackPath: "/open_api/knowledge/document/create",
			},
			{
				Placeholder:  "/open_api/knowledge/document/update",
				SDKMethod:    "datasets_documents.update",
				Method:       "post",
				FallbackPath: "/open_api/knowledge/document/update",
			},
			{
				Placeholder:  "/open_api/knowledge/document/delete",
				SDKMethod:    "datasets_documents.delete",
				Method:       "post",
				FallbackPath: "/open_api/knowledge/document/delete",
			},
			{
				Placeholder:  "/open_api/knowledge/document/list",
				SDKMethod:    "datasets_documents.list",
				Method:       "post",
				FallbackPath: "/open_api/knowledge/document/list",
			},
		},
	},
	"datasets_images.go": {
		Asset: "datasets_images.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/datasets/:dataset_id/images/:document_id",
				SDKMethod:        "datasets_images.update",
				Method:           "put",
				FallbackPath:     "/v1/datasets/{dataset_id}/images/{document_id}",
				ConvertCurlyPath: true,
			},
			{
				Placeholder:      "/v1/datasets/:dataset_id/images",
				SDKMethod:        "datasets_images.list",
				Method:           "get",
				FallbackPath:     "/v1/datasets/{dataset_id}/images",
				ConvertCurlyPath: true,
			},
		},
	},
	"enterprises_members.go": {
		Asset: "enterprises_members.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/enterprises/:enterprise_id/members",
				SDKMethod:        "enterprises_members.create",
				Method:           "post",
				FallbackPath:     "/v1/enterprises/{enterprise_id}/members",
				ConvertCurlyPath: true,
			},
			{
				Placeholder:      "/v1/enterprises/:enterprise_id/members/:user_id",
				SDKMethod:        "enterprises_members.update",
				Method:           "put",
				FallbackPath:     "/v1/enterprises/{enterprise_id}/members/{user_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"folders.go": {
		Asset: "folders.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/folders",
				SDKMethod:    "folders.list",
				Method:       "get",
				FallbackPath: "/v1/folders",
			},
			{
				Placeholder:      "/v1/folders/:folder_id",
				SDKMethod:        "folders.retrieve",
				Method:           "get",
				FallbackPath:     "/v1/folders/{folder_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"stores_plugins.go": {
		Asset: "stores_plugins.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/stores/plugins",
				SDKMethod:    "go.stores_plugins.list",
				Method:       "get",
				FallbackPath: "/v1/stores/plugins",
			},
		},
	},
	"variables.go": {
		Asset: "variables.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/variables",
				SDKMethod:    "variables.retrieve",
				Method:       "get",
				FallbackPath: "/v1/variables",
			},
		},
	},
	"workflows.go": {
		Asset: "workflows.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/workflows",
				SDKMethod:    "workflows.list",
				Method:       "get",
				FallbackPath: "/v1/workflows",
			},
		},
	},
	"workflows_runs.go": {
		Asset: "workflows_runs.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/workflow/run",
				SDKMethod:    "workflows_runs.create",
				Method:       "post",
				FallbackPath: "/v1/workflow/run",
			},
			{
				Placeholder:  "/v1/workflow/stream_resume",
				SDKMethod:    "workflows_runs.resume",
				Method:       "post",
				FallbackPath: "/v1/workflow/stream_resume",
			},
			{
				Placeholder:  "/v1/workflow/stream_run",
				SDKMethod:    "workflows_runs.stream",
				Method:       "post",
				FallbackPath: "/v1/workflow/stream_run",
			},
		},
	},
	"workflows_runs_histories.go": {
		Asset: "workflows_runs_histories.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/workflows/:workflow_id/run_histories/:execute_id",
				SDKMethod:        "workflows_runs_run_histories.retrieve",
				Method:           "get",
				FallbackPath:     "/v1/workflows/{workflow_id}/run_histories/{execute_id}",
				ConvertCurlyPath: true,
			},
		},
	},
	"workflows_runs_histories_execute_nodes.go": {
		Asset: "workflows_runs_histories_execute_nodes.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/workflows/:workflow_id/run_histories/:execute_id/execute_nodes/:node_execute_uuid",
				SDKMethod:        "workflows_runs_run_histories_execute_nodes.retrieve",
				Method:           "get",
				FallbackPath:     "/v1/workflows/{workflow_id}/run_histories/{execute_id}/execute_nodes/{node_execute_uuid}",
				ConvertCurlyPath: true,
			},
		},
	},
	"workspaces.go": {
		Asset: "workspaces.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:  "/v1/workspaces",
				SDKMethod:    "go.workspaces.list",
				Method:       "get",
				FallbackPath: "/v1/workspaces",
			},
		},
	},
	"workspaces_members.go": {
		Asset: "workspaces_members.go.tpl",
		PathReplacements: []goPathReplacementSpec{
			{
				Placeholder:      "/v1/workspaces/:workspace_id/members",
				SDKMethod:        "workspaces_members.list",
				Method:           "get",
				FallbackPath:     "/v1/workspaces/{workspace_id}/members",
				ConvertCurlyPath: true,
			},
		},
	},
}

var goGeneratedAPIModuleFiles = buildGoGeneratedAPIModuleFiles()

func buildGoGeneratedAPIModuleFiles() map[string]struct{} {
	files := make(map[string]struct{}, len(goInlineAPIModuleRenderers)+len(goTemplateAPIModuleSpecs))
	for _, renderer := range goInlineAPIModuleRenderers {
		files[renderer.FileName] = struct{}{}
	}
	for fileName := range goTemplateAPIModuleSpecs {
		files[fileName] = struct{}{}
	}
	return files
}

func listGoAPIModuleRenderers() []goAPIModuleRenderer {
	renderers := make([]goAPIModuleRenderer, 0, len(goInlineAPIModuleRenderers)+len(goTemplateAPIModuleSpecs))
	renderers = append(renderers, goInlineAPIModuleRenderers...)

	fileNames := make([]string, 0, len(goTemplateAPIModuleSpecs))
	for fileName := range goTemplateAPIModuleSpecs {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)
	for _, fileName := range fileNames {
		spec := goTemplateAPIModuleSpecs[fileName]
		specCopy := spec
		renderers = append(renderers, goAPIModuleRenderer{
			FileName: fileName,
			Render: func(cfg *config.Config, doc *openapi.Document) (string, error) {
				return renderGoAPITemplateModule(cfg, doc, specCopy)
			},
		})
	}
	return renderers
}

func shouldSkipGoExtraAsset(rel string) bool {
	target := strings.TrimSuffix(rel, ".tpl")
	if target == "README.md" || strings.HasPrefix(target, ".github/") {
		return true
	}
	_, ok := goGeneratedAPIModuleFiles[target]
	return ok
}

func renderGoAPITemplateModule(cfg *config.Config, doc *openapi.Document, spec goTemplateModuleSpec) (string, error) {
	content, err := renderGoExtraAsset(spec.Asset)
	if err != nil {
		return "", err
	}
	rendered := string(content)
	for _, replacement := range spec.PathReplacements {
		resolvedPath, err := findGoOperationPath(cfg, doc, replacement.SDKMethod, replacement.Method, replacement.FallbackPath)
		if err != nil {
			return "", err
		}
		if replacement.ConvertCurlyPath {
			resolvedPath = convertCurlyPathToColon(resolvedPath)
		}

		quotedPlaceholder := fmt.Sprintf("%q", replacement.Placeholder)
		quotedResolvedPath := fmt.Sprintf("%q", resolvedPath)
		if strings.Contains(rendered, quotedPlaceholder) {
			rendered = strings.ReplaceAll(rendered, quotedPlaceholder, quotedResolvedPath)
			continue
		}
		if strings.Contains(rendered, replacement.Placeholder) {
			rendered = strings.ReplaceAll(rendered, replacement.Placeholder, resolvedPath)
			continue
		}
		return "", fmt.Errorf("go api template %q missing path placeholder %q", spec.Asset, replacement.Placeholder)
	}
	return rendered, nil
}

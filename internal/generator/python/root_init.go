package python

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
)

type rootInitGroup struct {
	Module string
	Names  []string
}

var baseRootInitGroups = []rootInitGroup{
	{Module: ".api_apps", Names: []string{"APIApp", "AppType", "DeleteAPIAppsResp", "UpdateAPIAppsResp"}},
	{Module: ".api_apps.events", Names: []string{"CreateAPIAppsEventsResp", "DeleteAPIAppsEventsResp"}},
	{Module: ".apps", Names: []string{"SimpleApp"}},
	{Module: ".audio.live", Names: []string{"LiveInfo", "LiveType", "StreamInfo"}},
	{Module: ".audio.rooms", Names: []string{"CreateRoomResp"}},
	{Module: ".audio.speech", Names: []string{"AudioFormat"}},
	{Module: ".audio.transcriptions", Names: []string{"CreateTranscriptionsResp"}},
	{
		Module: ".audio.voiceprint_groups",
		Names: []string{
			"CreateVoicePrintGroupResp",
			"DeleteVoicePrintGroupResp",
			"FeatureScore",
			"SpeakerIdentifyResp",
			"UpdateVoicePrintGroupResp",
			"VoicePrintGroup",
		},
	},
	{
		Module: ".audio.voiceprint_groups.features",
		Names: []string{
			"CreateVoicePrintGroupFeatureResp",
			"DeleteVoicePrintGroupFeatureResp",
			"UpdateVoicePrintGroupFeatureResp",
			"VoicePrintGroupFeature",
		},
	},
	{Module: ".audio.voices", Names: []string{"Voice", "VoiceModelType", "VoiceState"}},
	{
		Module: ".auth",
		Names: []string{
			"AsyncAuth",
			"AsyncDeviceOAuthApp",
			"AsyncJWTAuth",
			"AsyncJWTOAuthApp",
			"AsyncPKCEOAuthApp",
			"AsyncTokenAuth",
			"AsyncWebOAuthApp",
			"Auth",
			"DeviceAuthCode",
			"DeviceOAuthApp",
			"JWTAuth",
			"JWTOAuthApp",
			"OAuthApp",
			"OAuthToken",
			"PKCEOAuthApp",
			"Scope",
			"SyncAuth",
			"TokenAuth",
			"WebOAuthApp",
			"load_oauth_app_from_config",
		},
	},
	{
		Module: ".bots",
		Names: []string{
			"BackgroundImageInfo",
			"Bot",
			"BotBackgroundImageInfo",
			"BotKnowledge",
			"BotModelInfo",
			"BotOnboardingInfo",
			"BotPluginAPIInfo",
			"BotPluginInfo",
			"BotPromptInfo",
			"BotSuggestReplyInfo",
			"BotVariable",
			"BotVoiceInfo",
			"BotWorkflowInfo",
			"CanvasPosition",
			"GradientPosition",
			"PluginIDList",
			"PublishStatus",
			"SimpleBot",
			"SuggestReplyMode",
			"UpdateBotResp",
			"UserInputType",
			"VariableChannel",
			"VariableType",
			"WorkflowIDList",
		},
	},
	{
		Module: ".chat",
		Names: []string{
			"Chat",
			"ChatError",
			"ChatEvent",
			"ChatEventType",
			"ChatPoll",
			"ChatRequiredAction",
			"ChatRequiredActionType",
			"ChatStatus",
			"ChatSubmitToolOutputs",
			"ChatToolCall",
			"ChatToolCallFunction",
			"ChatToolCallType",
			"ChatUsage",
			"Message",
			"MessageContentType",
			"MessageObjectString",
			"MessageObjectStringType",
			"MessageRole",
			"MessageType",
			"ToolOutput",
		},
	},
	{
		Module: ".config",
		Names: []string{
			"COZE_CN_BASE_URL",
			"COZE_COM_BASE_URL",
			"DEFAULT_CONNECTION_LIMITS",
			"DEFAULT_TIMEOUT",
		},
	},
	{Module: ".conversations", Names: []string{"Conversation", "Section"}},
	{
		Module: ".conversations.message.feedback",
		Names: []string{
			"CreateConversationMessageFeedbackResp",
			"DeleteConversationMessageFeedbackResp",
			"FeedbackType",
		},
	},
	{Module: ".coze", Names: []string{"AsyncCoze", "Coze"}},
	{Module: ".datasets", Names: []string{"CreateDatasetResp", "Dataset", "DatasetStatus", "DocumentProgress"}},
	{
		Module: ".datasets.documents",
		Names: []string{
			"Document",
			"DocumentBase",
			"DocumentChunkStrategy",
			"DocumentFormatType",
			"DocumentSourceInfo",
			"DocumentSourceType",
			"DocumentStatus",
			"DocumentUpdateRule",
			"DocumentUpdateType",
		},
	},
	{Module: ".datasets.images", Names: []string{"Photo"}},
	{Module: ".enterprises.members", Names: []string{"EnterpriseMember", "EnterpriseMemberRole"}},
	{
		Module: ".exception",
		Names: []string{
			"CozeAPIError",
			"CozeError",
			"CozeInvalidEventError",
			"CozePKCEAuthError",
			"CozePKCEAuthErrorType",
		},
	},
	{Module: ".files", Names: []string{"File"}},
	{Module: ".folders", Names: []string{"FolderType", "SimpleFolder"}},
	{Module: ".log", Names: []string{"setup_logging"}},
	{
		Module: ".model",
		Names: []string{
			"AsyncLastIDPaged",
			"AsyncNumberPaged",
			"AsyncPagedBase",
			"AsyncStream",
			"FileHTTPResponse",
			"LastIDPaged",
			"LastIDPagedResponse",
			"ListResponse",
			"NumberPaged",
			"NumberPagedResponse",
			"Stream",
		},
	},
	{Module: ".request", Names: []string{"AsyncHTTPClient", "SyncHTTPClient"}},
	{Module: ".templates", Names: []string{"TemplateDuplicateResp", "TemplateEntityType"}},
	{Module: ".users", Names: []string{"User"}},
	{Module: ".variables", Names: []string{"UpdateVariableResp", "VariableValue"}},
	{Module: ".version", Names: []string{"VERSION"}},
	{
		Module: ".websockets.audio.speech",
		Names: []string{
			"AsyncWebsocketsAudioSpeechClient",
			"AsyncWebsocketsAudioSpeechEventHandler",
			"InputTextBufferAppendEvent",
			"InputTextBufferCompletedEvent",
			"InputTextBufferCompleteEvent",
			"SpeechAudioCompletedEvent",
			"SpeechAudioUpdateEvent",
			"SpeechCreatedEvent",
			"SpeechUpdatedEvent",
			"SpeechUpdateEvent",
			"WebsocketsAudioSpeechClient",
			"WebsocketsAudioSpeechEventHandler",
		},
	},
	{
		Module: ".websockets.audio.transcriptions",
		Names: []string{
			"AsyncWebsocketsAudioTranscriptionsClient",
			"AsyncWebsocketsAudioTranscriptionsEventHandler",
			"InputAudioBufferAppendEvent",
			"InputAudioBufferClearedEvent",
			"InputAudioBufferClearEvent",
			"InputAudioBufferCompletedEvent",
			"InputAudioBufferCompleteEvent",
			"TranscriptionsCreatedEvent",
			"TranscriptionsMessageCompletedEvent",
			"TranscriptionsMessageUpdateEvent",
			"TranscriptionsUpdatedEvent",
			"TranscriptionsUpdateEvent",
			"WebsocketsAudioTranscriptionsClient",
			"WebsocketsAudioTranscriptionsEventHandler",
		},
	},
	{
		Module: ".websockets.chat",
		Names: []string{
			"AsyncWebsocketsChatClient",
			"AsyncWebsocketsChatEventHandler",
			"ChatCreatedEvent",
			"ChatUpdatedEvent",
			"ChatUpdateEvent",
			"ConversationAudioCompletedEvent",
			"ConversationAudioDeltaEvent",
			"ConversationAudioSentenceStartEvent",
			"ConversationAudioTranscriptCompletedEvent",
			"ConversationAudioTranscriptUpdateEvent",
			"ConversationChatCanceledEvent",
			"ConversationChatCancelEvent",
			"ConversationChatCompletedEvent",
			"ConversationChatCreatedEvent",
			"ConversationChatFailedEvent",
			"ConversationChatInProgressEvent",
			"ConversationChatRequiresActionEvent",
			"ConversationChatSubmitToolOutputsEvent",
			"ConversationClear",
			"ConversationClearedEvent",
			"ConversationMessageCompletedEvent",
			"ConversationMessageCreateEvent",
			"ConversationMessageDeltaEvent",
			"InputAudioBufferSpeechStartedEvent",
			"InputAudioBufferSpeechStoppedEvent",
			"InputTextGenerateAudioEvent",
			"WebsocketsChatClient",
			"WebsocketsChatEventHandler",
		},
	},
	{
		Module: ".websockets.ws",
		Names: []string{
			"InputAudio",
			"LimitConfig",
			"OpusConfig",
			"OutputAudio",
			"PCMConfig",
			"WebsocketsErrorEvent",
			"WebsocketsEvent",
			"WebsocketsEventType",
		},
	},
	{Module: ".workflows", Names: []string{"WorkflowBasic", "WorkflowMode"}},
	{
		Module: ".workflows.runs",
		Names: []string{
			"WorkflowEvent",
			"WorkflowEventError",
			"WorkflowEventInterrupt",
			"WorkflowEventInterruptData",
			"WorkflowEventMessage",
			"WorkflowEventType",
			"WorkflowRunResult",
		},
	},
	{
		Module: ".workflows.runs.run_histories",
		Names: []string{
			"WorkflowExecuteStatus",
			"WorkflowRunHistory",
			"WorkflowRunHistoryNodeExecuteStatus",
			"WorkflowRunMode",
		},
	},
	{Module: ".workflows.runs.run_histories.execute_nodes", Names: []string{"WorkflowNodeExecuteHistory"}},
	{Module: ".workflows.versions", Names: []string{"WorkflowUserInfo", "WorkflowVersionInfo"}},
	{Module: ".workspaces", Names: []string{"Workspace", "WorkspaceRoleType", "WorkspaceType"}},
	{
		Module: ".workspaces.members",
		Names: []string{
			"CreateWorkspaceMemberResp",
			"DeleteWorkspaceMemberResp",
			"WorkspaceMember",
		},
	},
}

func renderPythonRootInit(cfg *config.Config) (string, error) {
	collaboratorNames := collectAppsCollaboratorExports(cfg)
	importGroups := make([]rootInitGroup, 0, len(baseRootInitGroups)+1)
	exportGroups := make([]rootInitGroup, 0, len(baseRootInitGroups)+1)

	for _, group := range baseRootInitGroups {
		importGroups = append(importGroups, group)
		if group.Module != ".version" {
			exportGroups = append(exportGroups, group)
		}
		if group.Module == ".apps" && len(collaboratorNames) > 0 {
			collaboratorGroup := rootInitGroup{
				Module: ".apps.collaborators",
				Names:  collaboratorNames,
			}
			importGroups = append(importGroups, collaboratorGroup)
			exportGroups = append(exportGroups, collaboratorGroup)
		}
	}

	var buf strings.Builder
	for _, group := range importGroups {
		importLine := renderRootInitImport(group.Module, group.Names)
		if strings.TrimSpace(importLine) == "" {
			continue
		}
		buf.WriteString(importLine)
		buf.WriteString("\n")
	}
	buf.WriteString("\n")
	buf.WriteString("__all__ = [\n")
	buf.WriteString("    \"VERSION\",\n")
	seenAll := map[string]struct{}{}
	for _, group := range exportGroups {
		groupNames := make([]string, 0, len(group.Names))
		for _, rawName := range group.Names {
			name := strings.TrimSpace(rawName)
			if name == "" {
				continue
			}
			if _, ok := seenAll[name]; ok {
				continue
			}
			seenAll[name] = struct{}{}
			groupNames = append(groupNames, name)
		}
		if len(groupNames) == 0 {
			continue
		}
		buf.WriteString(fmt.Sprintf("    # %s\n", strings.TrimPrefix(group.Module, ".")))
		for _, name := range groupNames {
			buf.WriteString(fmt.Sprintf("    %q,\n", name))
		}
	}
	buf.WriteString("]\n")
	return buf.String(), nil
}

func renderRootInitImport(module string, names []string) string {
	if strings.TrimSpace(module) == "" {
		return ""
	}
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		n := strings.TrimSpace(name)
		if n == "" {
			continue
		}
		cleaned = append(cleaned, n)
	}
	if len(cleaned) == 0 {
		return ""
	}
	if len(cleaned) <= 4 {
		return fmt.Sprintf("from %s import %s", module, strings.Join(cleaned, ", "))
	}
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("from %s import (\n", module))
	for _, name := range cleaned {
		buf.WriteString(fmt.Sprintf("    %s,\n", name))
	}
	buf.WriteString(")")
	return buf.String()
}

func collectAppsCollaboratorExports(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	seen := map[string]struct{}{}
	names := make([]string, 0, 8)
	appendName := func(value string) {
		name := strings.TrimSpace(value)
		if !isPublicPythonName(name) {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}

	for _, pkg := range cfg.API.Packages {
		sourceDir := strings.TrimSpace(pkg.SourceDir)
		if sourceDir != "cozepy/apps/collaborators" {
			continue
		}
		for _, schema := range pkg.ModelSchemas {
			appendName(schema.Name)
		}
		for _, modelName := range pkg.EmptyModels {
			appendName(modelName)
		}
	}
	sort.Strings(names)
	return names
}

func isPublicPythonName(name string) bool {
	if strings.TrimSpace(name) == "" {
		return false
	}
	runes := []rune(name)
	if len(runes) == 0 || !unicode.IsUpper(runes[0]) {
		return false
	}
	for _, r := range runes {
		if r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

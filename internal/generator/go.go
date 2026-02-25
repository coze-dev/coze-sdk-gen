package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
	"github.com/coze-dev/coze-sdk-gen/internal/openapi"
)

type goOperationBinding struct {
	MethodName string
	Path       string
	Method     string
	Summary    string
}

func GenerateGo(cfg *config.Config, doc *openapi.Document) (Result, error) {
	if cfg == nil {
		return Result{}, fmt.Errorf("config is required")
	}
	if doc == nil {
		return Result{}, fmt.Errorf("swagger document is required")
	}

	bindings := buildGoOperationBindings(cfg, doc)
	if len(bindings) == 0 {
		return Result{}, fmt.Errorf("no operations selected for generation")
	}

	if err := os.RemoveAll(cfg.OutputSDK); err != nil {
		return Result{}, fmt.Errorf("clean output directory %q: %w", cfg.OutputSDK, err)
	}
	if err := os.MkdirAll(cfg.OutputSDK, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory %q: %w", cfg.OutputSDK, err)
	}

	writer := &fileWriter{
		written: map[string]struct{}{},
	}
	if err := writeGoRuntimeScaffolding(cfg.OutputSDK, writer); err != nil {
		return Result{}, err
	}
	if err := writeGoAPIModules(cfg, doc, writer); err != nil {
		return Result{}, err
	}
	if err := writeGoExtraAssets(cfg.OutputSDK, writer); err != nil {
		return Result{}, err
	}

	return Result{
		GeneratedFiles: writer.count,
		GeneratedOps:   len(bindings),
	}, nil
}

func buildGoOperationBindings(cfg *config.Config, doc *openapi.Document) []goOperationBinding {
	allOps := doc.ListOperationDetails()
	bindings := make([]goOperationBinding, 0, len(allOps))
	seen := map[string]int{}

	for _, details := range allOps {
		if cfg.IsIgnored(details.Path, details.Method) {
			continue
		}
		base := strings.TrimSpace(details.OperationID)
		if base == "" {
			base = defaultMethodName(details.OperationID, details.Path, details.Method)
		}
		name := normalizeGoExportedIdentifier(base)
		if name == "" {
			name = "Operation"
		}

		count := seen[name]
		seen[name] = count + 1
		if count > 0 {
			name = fmt.Sprintf("%s%d", name, count+1)
		}

		bindings = append(bindings, goOperationBinding{
			MethodName: name,
			Path:       details.Path,
			Method:     strings.ToUpper(details.Method),
			Summary:    oneLineText(details.Summary),
		})
	}
	return bindings
}

func writeGoRuntimeScaffolding(outputDir string, writer *fileWriter) error {
	textAssets := map[string]string{
		".gitignore":      "gitignore.raw",
		"codecov.yml":     "codecov.yml.raw",
		"go.mod":          "go.mod.raw",
		"go.sum":          "go.sum.raw",
		"LICENSE":         "LICENSE.raw",
		"CONTRIBUTING.md": "CONTRIBUTING.md.raw",
	}
	for target, asset := range textAssets {
		content, err := renderGoRuntimeAsset(asset)
		if err != nil {
			return err
		}
		if target == ".gitignore" || target == "codecov.yml" || target == "LICENSE" {
			content = strings.TrimSuffix(content, "\n")
		}
		if err := writer.write(filepath.Join(outputDir, target), content); err != nil {
			return err
		}
	}

	goAssets := map[string]string{
		"audio.go":                         "audio.go.raw",
		"auth.go":                          "auth.go.raw",
		"auth_token.go":                    "auth_token.go.raw",
		"base_model.go":                    "base_model.go.raw",
		"client.go":                        "client.go.raw",
		"common.go":                        "common.go.raw",
		"const.go":                         "const.go.raw",
		"error.go":                         "error.go.raw",
		"enterprises.go":                   "enterprises.go.raw",
		"logger.go":                        "logger.go.raw",
		"pagination.go":                    "pagination.go.raw",
		"request.go":                       "request.go.raw",
		"stores.go":                        "stores.go.raw",
		"stream_reader.go":                 "stream_reader.go.raw",
		"user_agent.go":                    "user_agent.go.raw",
		"utils.go":                         "utils.go.raw",
		"websocket.go":                     "websocket.go.raw",
		"websocket_audio.go":               "websocket_audio.go.raw",
		"websocket_audio_speech_client.go": "websocket_audio_speech_client.go.raw",
		"websocket_audio_speech.go":        "websocket_audio_speech.go.raw",
		"websocket_audio_transcription_client.go": "websocket_audio_transcription_client.go.raw",
		"websocket_audio_transcription.go":        "websocket_audio_transcription.go.raw",
		"websocket_chat_client.go":                "websocket_chat_client.go.raw",
		"websocket_chat.go":                       "websocket_chat.go.raw",
		"websocket_client.go":                     "websocket_client.go.raw",
		"websocket_event.go":                      "websocket_event.go.raw",
		"websocket_event_type.go":                 "websocket_event_type.go.raw",
		"websocket_wait.go":                       "websocket_wait.go.raw",
	}
	for target, asset := range goAssets {
		content, err := renderGoRuntimeAsset(asset)
		if err != nil {
			return err
		}
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("format go runtime file %q: %w", target, err)
		}
		if err := writer.write(filepath.Join(outputDir, target), string(formatted)); err != nil {
			return err
		}
	}
	return nil
}

func writeGoExtraAssets(outputDir string, writer *fileWriter) error {
	assets, err := listGoExtraAssets()
	if err != nil {
		return err
	}
	for _, rel := range assets {
		if shouldSkipGoExtraAsset(rel) {
			continue
		}
		content, err := renderGoExtraAsset(rel)
		if err != nil {
			return err
		}
		target := rel
		if strings.HasSuffix(target, ".raw") {
			target = strings.TrimSuffix(target, ".raw")
		}
		if err := writer.writeBytes(filepath.Join(outputDir, target), content); err != nil {
			return err
		}
	}
	return nil
}

func writeGoAPIModules(cfg *config.Config, doc *openapi.Document, writer *fileWriter) error {
	for _, renderer := range listGoAPIModuleRenderers() {
		content, err := renderer.Render(cfg, doc)
		if err != nil {
			return err
		}
		formatted, err := format.Source([]byte(content))
		if err != nil {
			return fmt.Errorf("format generated go api module %q: %w", renderer.FileName, err)
		}
		if err := writer.write(filepath.Join(cfg.OutputSDK, renderer.FileName), string(formatted)); err != nil {
			return err
		}
	}
	return nil
}

func renderGoAppsModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	listPath := findGoOperationPath(cfg, doc, "apps.list", "get", "/v1/apps")
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

func (r *apps) List(ctx context.Context, req *ListAppReq) (NumberPaged[SimpleApp], error) {
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	return NewNumberPaged(
		func(request *pageRequest) (*pageResponse[SimpleApp], error) {
			resp := new(listAppResp)
			err := r.core.rawRequest(ctx, &RawRequestReq{
				Method: http.MethodGet,
				URL:    %q,
				Body:   req.toReq(request),
			}, resp)
			if err != nil {
				return nil, err
			}
			return &pageResponse[SimpleApp]{
				response: resp.HTTPResponse,
				HasMore:  len(resp.Data.Items) >= request.PageSize,
				Data:     resp.Data.Items,
				LogID:    resp.HTTPResponse.LogID(),
			}, nil
		}, req.PageSize, req.PageNum)
}

type ListAppReq struct {
	WorkspaceID   string         `+"`query:\"workspace_id\" json:\"-\"`"+`
	PublishStatus *PublishStatus `+"`query:\"publish_status\" json:\"-\"`"+`
	ConnectorID   *string        `+"`query:\"connector_id\" json:\"-\"`"+`
	PageNum       int            `+"`query:\"page_num\" json:\"-\"`"+`
	PageSize      int            `+"`query:\"page_size\" json:\"-\"`"+`
}

type SimpleApp struct {
	ID          string `+"`json:\"id,omitempty\"`"+`
	Name        string `+"`json:\"name,omitempty\"`"+`
	Description string `+"`json:\"description,omitempty\"`"+`
	IconURL     string `+"`json:\"icon_url,omitempty\"`"+`
	IsPublished bool   `+"`json:\"is_published,omitempty\"`"+`
	OwnerUserID string `+"`json:\"owner_user_id,omitempty\"`"+`
	UpdatedAt   int    `+"`json:\"updated_at,omitempty\"`"+`
	PublishedAt *int   `+"`json:\"published_at,omitempty\"`"+`
}

type ListAppResp struct {
	Total int          `+"`json:\"total\"`"+`
	Items []*SimpleApp `+"`json:\"items\"`"+`
}

type listAppResp struct {
	baseResponse
	Data *ListAppResp `+"`json:\"data\"`"+`
}

func (r ListAppReq) toReq(request *pageRequest) *ListAppReq {
	return &ListAppReq{
		WorkspaceID:   r.WorkspaceID,
		PublishStatus: r.PublishStatus,
		ConnectorID:   r.ConnectorID,
		PageNum:       request.PageNum,
		PageSize:      request.PageSize,
	}
}

type apps struct {
	core *core
}

func newApps(core *core) *apps {
	return &apps{
		core: core,
	}
}
`, listPath), nil
}

func renderGoAudioLiveModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	retrievePath := findGoOperationPath(cfg, doc, "audio_live.retrieve", "get", "/v1/audio/live/{live_id}")
	retrievePath = convertCurlyPathToColon(retrievePath)
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

// Retrieve retrieves live stream information
func (r *audioLive) Retrieve(ctx context.Context, req *RetrieveAudioLiveReq) (*LiveInfo, error) {
	request := &RawRequestReq{
		Method: http.MethodGet,
		URL:    %q,
		Body:   req,
	}
	response := new(retrieveAudioLiveResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.Data, err
}

// LiveType represents the type of live stream
type LiveType string

const (
	LiveTypeOrigin      LiveType = "origin"
	LiveTypeTranslation LiveType = "translation"
)

func (l LiveType) String() string {
	return string(l)
}

func (l LiveType) Ptr() *LiveType {
	return &l
}

// StreamInfo represents information about a stream
type StreamInfo struct {
	StreamID string   `+"`json:\"stream_id\"`"+`
	Name     string   `+"`json:\"name\"`"+`
	LiveType LiveType `+"`json:\"live_type\"`"+`
}

// LiveInfo represents information about a live session
type LiveInfo struct {
	baseModel
	AppID       string        `+"`json:\"app_id\"`"+`
	StreamInfos []*StreamInfo `+"`json:\"stream_infos\"`"+`
}

// RetrieveAudioLiveReq represents the request for retrieving live information
type RetrieveAudioLiveReq struct {
	LiveID string `+"`path:\"live_id\" json:\"-\"`"+`
}

type retrieveAudioLiveResp struct {
	baseResponse
	Data *LiveInfo `+"`json:\"data\"`"+`
}

// audioLive provides operations for live audio streams
type audioLive struct {
	core *core
}

func newAudioLive(core *core) *audioLive {
	return &audioLive{core: core}
}
`, retrievePath), nil
}

func renderGoAudioSpeechModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	createPath := findGoOperationPath(cfg, doc, "audio_speech.create", "post", "/v1/audio/speech")
	return fmt.Sprintf(`package coze

import (
	"context"
	"io"
	"net/http"
	"os"
)

func (r *audioSpeech) Create(ctx context.Context, req *CreateAudioSpeechReq) (*CreateAudioSpeechResp, error) {
	request := &RawRequestReq{
		Method: http.MethodPost,
		URL:    %q,
		Body:   req,
	}
	response := new(createAudioSpeechResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.Data, err
}

// CreateAudioSpeechReq represents the request for creating speech
type CreateAudioSpeechReq struct {
	Input          string       `+"`json:\"input\"`"+`
	VoiceID        string       `+"`json:\"voice_id\"`"+`
	ResponseFormat *AudioFormat `+"`json:\"response_format\"`"+`
	Speed          *float32     `+"`json:\"speed\"`"+`
	SampleRate     *int         `+"`json:\"sample_rate\"`"+`
	LoudnessRate   *int         `+"`json:\"loudness_rate\"`"+`
	Emotion        *string      `+"`json:\"emotion\"`"+`
	EmotionScale   *float32     `+"`json:\"emotion_scale\"`"+`
}

// CreateAudioSpeechResp represents the response for creating speech
type CreateAudioSpeechResp struct {
	baseModel
	Data io.ReadCloser
}

type createAudioSpeechResp struct {
	baseResponse
	Data *CreateAudioSpeechResp
}

func (r *createAudioSpeechResp) SetReader(file io.ReadCloser) {
	if r.Data == nil {
		r.Data = &CreateAudioSpeechResp{}
	}
	r.Data.Data = file
}

func (c *CreateAudioSpeechResp) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	defer c.Data.Close()

	_, err = io.Copy(file, c.Data)
	return err
}

type audioSpeech struct {
	core *core
}

func newAudioSpeech(core *core) *audioSpeech {
	return &audioSpeech{core: core}
}
`, createPath), nil
}

func renderGoAudioTranscriptionsModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	createPath := findGoOperationPath(cfg, doc, "audio_transcriptions.create", "post", "/v1/audio/transcriptions")
	return fmt.Sprintf(`package coze

import (
	"context"
	"io"
	"net/http"
)

func (r *audioTranscriptions) Create(ctx context.Context, req *AudioSpeechTranscriptionsReq) (*CreateAudioTranscriptionsResp, error) {
	request := &RawRequestReq{
		Method: http.MethodPost,
		URL:    %q,
		Body:   req,
		IsFile: true,
	}
	response := new(createAudioTranscriptionsResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.CreateAudioTranscriptionsResp, err
}

type AudioSpeechTranscriptionsReq struct {
	Filename string    `+"`json:\"filename\"`"+`
	Audio    io.Reader `+"`json:\"file\"`"+`
}

type createAudioTranscriptionsResp struct {
	baseResponse
	*CreateAudioTranscriptionsResp
}

type CreateAudioTranscriptionsResp struct {
	baseModel
	Data AudioTranscriptionsData `+"`json:\"data\"`"+`
}

type AudioTranscriptionsData struct {
	Text string `+"`json:\"text\"`"+`
}

type audioTranscriptions struct {
	core *core
}

func newAudioTranscriptions(core *core) *audioTranscriptions {
	return &audioTranscriptions{core: core}
}
`, createPath), nil
}

func renderGoChatsMessagesModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	listPath := findGoOperationPath(cfg, doc, "chat_messages.list", "get", "/v3/chat/message/list")
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

func (r *chatMessages) List(ctx context.Context, req *ListChatsMessagesReq) (*ListChatsMessagesResp, error) {
	request := &RawRequestReq{
		Method: http.MethodGet,
		URL:    %q,
		Body:   req,
	}
	response := new(listChatsMessagesResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.ListChatsMessagesResp, err
}

// ListChatsMessagesReq represents the request to list messages
type ListChatsMessagesReq struct {
	// The Conversation ID can be viewed in the 'conversation_id' field of the Response when
	// initiating a conversation through the Chat API.
	ConversationID string `+"`query:\"conversation_id\" json:\"-\"`"+`

	// The Chat ID can be viewed in the 'id' field of the Response when initiating a chat through the
	// Chat API. If it is a streaming response, check the 'id' field in the chat event of the Response.
	ChatID string `+"`query:\"chat_id\" json:\"-\"`"+`
}

type ListChatsMessagesResp struct {
	baseModel
	Messages []*Message `+"`json:\"data\"`"+`
}

type listChatsMessagesResp struct {
	baseResponse
	*ListChatsMessagesResp
}

type chatMessages struct {
	core *core
}

func newChatMessages(core *core) *chatMessages {
	return &chatMessages{core: core}
}
`, listPath), nil
}

func renderGoFilesModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	uploadPath := findGoOperationPath(cfg, doc, "files.upload", "post", "/v1/files/upload")
	retrievePath := findGoOperationPath(cfg, doc, "files.retrieve", "get", "/v1/files/retrieve")
	return fmt.Sprintf(`package coze

import (
	"context"
	"io"
	"net/http"
)

func (r *files) Upload(ctx context.Context, req *UploadFilesReq) (*UploadFilesResp, error) {
	request := &RawRequestReq{
		Method: http.MethodPost,
		URL:    %q,
		Body:   req,
		IsFile: true,
	}
	response := new(uploadFilesResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.Data, err
}

func (r *files) Retrieve(ctx context.Context, req *RetrieveFilesReq) (*RetrieveFilesResp, error) {
	request := &RawRequestReq{
		Method: http.MethodGet,
		URL:    %q,
		Body:   req,
	}
	response := new(retrieveFilesResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.Data, err
}

// FileInfo represents information about a file
type FileInfo struct {
	// The ID of the uploaded file.
	ID string `+"`json:\"id\"`"+`

	// The total byte size of the file.
	Bytes int `+"`json:\"bytes\"`"+`

	// The upload time of the file, in the format of a 10-digit Unix timestamp in seconds (s).
	CreatedAt int `+"`json:\"created_at\"`"+`

	// The name of the file.
	FileName string `+"`json:\"file_name\"`"+`
}

type FileTypes interface {
	io.Reader
	Name() string
}

type implFileInterface struct {
	io.Reader
	fileName string
}

func (r *implFileInterface) Name() string {
	return r.fileName
}

type UploadFilesReq struct {
	File FileTypes `+"`json:\"file\"`"+`
}

func NewUploadFile(reader io.Reader, fileName string) FileTypes {
	return &implFileInterface{
		Reader:   reader,
		fileName: fileName,
	}
}

// RetrieveFilesReq represents request for retrieving file
type RetrieveFilesReq struct {
	FileID string `+"`query:\"file_id\" json:\"-\"`"+`
}

// UploadFilesResp represents response for uploading file
type UploadFilesResp struct {
	baseModel
	FileInfo
}

// RetrieveFilesResp represents response for retrieving file
type RetrieveFilesResp struct {
	baseModel
	FileInfo
}

type uploadFilesResp struct {
	baseResponse
	Data *UploadFilesResp `+"`json:\"data\"`"+`
}

type retrieveFilesResp struct {
	baseResponse
	Data *RetrieveFilesResp `+"`json:\"data\"`"+`
}

type files struct {
	core *core
}

func newFiles(core *core) *files {
	return &files{core: core}
}
`, uploadPath, retrievePath), nil
}

func renderGoTemplatesModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	duplicatePath := findGoOperationPath(cfg, doc, "templates.duplicate", "post", "/v1/templates/{template_id}/duplicate")
	duplicatePath = convertCurlyPathToColon(duplicatePath)
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

// Duplicate creates a copy of an existing template
func (r *templates) Duplicate(ctx context.Context, templateID string, req *DuplicateTemplateReq) (*TemplateDuplicateResp, error) {
	if req == nil {
		req = &DuplicateTemplateReq{}
	}
	req.TemplateID = templateID
	request := &RawRequestReq{
		Method: http.MethodPost,
		URL:    %q,
		Body:   req,
	}
	response := new(templateDuplicateResp)
	err := r.core.rawRequest(ctx, request, response)
	return response.Data, err
}

// TemplateEntityType represents the type of template entity
type TemplateEntityType string

const (
	// TemplateEntityTypeAgent represents an agent template
	TemplateEntityTypeAgent TemplateEntityType = "agent"
)

// TemplateDuplicateResp represents the response from duplicating a template
type TemplateDuplicateResp struct {
	baseModel
	EntityID   string             `+"`json:\"entity_id\"`"+`
	EntityType TemplateEntityType `+"`json:\"entity_type\"`"+`
}

// DuplicateTemplateReq represents the request to duplicate a template
type DuplicateTemplateReq struct {
	TemplateID  string  `+"`path:\"template_id\" json:\"-\"`"+`
	WorkspaceID string  `+"`json:\"workspace_id,omitempty\"`"+`
	Name        *string `+"`json:\"name,omitempty\"`"+`
}

// templateDuplicateResp represents response for creating document
type templateDuplicateResp struct {
	baseResponse
	Data *TemplateDuplicateResp `+"`json:\"data\"`"+`
}

type templates struct {
	core *core
}

func newTemplates(core *core) *templates {
	return &templates{core: core}
}
`, duplicatePath), nil
}

func renderGoUsersModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	mePath := findGoOperationPath(cfg, doc, "users.me", "get", "/v1/users/me")
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

// Me retrieves the current user's information
func (r *users) Me(ctx context.Context) (*User, error) {
	request := &RawRequestReq{
		Method: http.MethodGet,
		URL:    %q,
		Body:   new(GetUserMeReq),
	}
	response := new(meResp)
	err := r.client.rawRequest(ctx, request, response)
	return response.User, err
}

type GetUserMeReq struct{}

// User represents a Coze user
type User struct {
	baseModel
	UserID    string `+"`json:\"user_id\"`"+`
	UserName  string `+"`json:\"user_name\"`"+`
	NickName  string `+"`json:\"nick_name\"`"+`
	AvatarURL string `+"`json:\"avatar_url\"`"+`
}

type meResp struct {
	baseResponse
	User *User `+"`json:\"data\"`"+`
}

type users struct {
	client *core
}

func newUsers(core *core) *users {
	return &users{
		client: core,
	}
}
`, mePath), nil
}

func renderGoWorkflowsChatModule(cfg *config.Config, doc *openapi.Document) (string, error) {
	streamPath := findGoOperationPath(cfg, doc, "workflows_chat.stream", "post", "/v1/workflows/chat")
	return fmt.Sprintf(`package coze

import (
	"context"
	"net/http"
)

func (r *workflowsChat) Stream(ctx context.Context, req *WorkflowsChatStreamReq) (Stream[ChatEvent], error) {
	request := &RawRequestReq{
		Method: http.MethodPost,
		URL:    %q,
		Body:   req,
	}
	response := new(createChatsResp)
	err := r.client.rawRequest(ctx, request, response)
	return newStream(ctx, r.client, response.HTTPResponse, parseChatEvent), err
}

// WorkflowsChatStreamReq 表示工作流聊天流式请求
type WorkflowsChatStreamReq struct {
	WorkflowID         string            `+"`json:\"workflow_id\"`"+`               // 工作流ID
	AdditionalMessages []*Message        `+"`json:\"additional_messages\"`"+`       // 额外的消息信息
	Parameters         map[string]any    `+"`json:\"parameters,omitempty\"`"+`      // 工作流参数
	AppID              *string           `+"`json:\"app_id,omitempty\"`"+`          // 应用ID
	BotID              *string           `+"`json:\"bot_id,omitempty\"`"+`          // 机器人ID
	ConversationID     *string           `+"`json:\"conversation_id,omitempty\"`"+` // 会话ID
	Ext                map[string]string `+"`json:\"ext,omitempty\"`"+`             // 扩展信息
}

type workflowsChat struct {
	client *core
}

func newWorkflowsChat(core *core) *workflowsChat {
	return &workflowsChat{
		client: core,
	}
}
`, streamPath), nil
}

func findGoOperationPath(cfg *config.Config, doc *openapi.Document, sdkMethod string, method string, fallback string) string {
	if cfg != nil {
		for _, mapping := range cfg.API.OperationMappings {
			if !strings.EqualFold(strings.TrimSpace(mapping.Method), strings.TrimSpace(method)) {
				continue
			}
			for _, m := range mapping.SDKMethods {
				if strings.TrimSpace(m) == strings.TrimSpace(sdkMethod) {
					return strings.TrimSpace(mapping.Path)
				}
			}
		}
	}
	if doc != nil && doc.HasOperation(method, fallback) {
		return fallback
	}
	return fallback
}

func convertCurlyPathToColon(path string) string {
	converted := path
	for {
		start := strings.Index(converted, "{")
		if start < 0 {
			break
		}
		endOffset := strings.Index(converted[start:], "}")
		if endOffset <= 1 {
			break
		}
		end := start + endOffset
		param := strings.TrimSpace(converted[start+1 : end])
		if param == "" {
			break
		}
		converted = converted[:start] + ":" + param + converted[end+1:]
	}
	return converted
}

func normalizeGoExportedIdentifier(value string) string {
	parts := splitIdentifierWords(value)
	if len(parts) == 0 {
		return ""
	}
	var buf strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		first := unicode.ToUpper(runes[0])
		buf.WriteRune(first)
		if len(runes) > 1 {
			buf.WriteString(string(runes[1:]))
		}
	}
	result := buf.String()
	if result == "" {
		return ""
	}
	firstRune := []rune(result)[0]
	if unicode.IsDigit(firstRune) {
		return "Op" + result
	}
	return result
}

func splitIdentifierWords(value string) []string {
	if value == "" {
		return nil
	}
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	words := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		words = append(words, field)
	}
	return words
}

func oneLineText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

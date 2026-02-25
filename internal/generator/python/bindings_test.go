package python

import (
	"testing"

	"github.com/coze-dev/coze-sdk-gen/internal/config"
)

func TestBuildPackageMetaInfersDirectChildClients(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Packages: []config.Package{
				{
					Name:      "workflows",
					SourceDir: "cozepy/workflows",
				},
				{
					Name:             "workflows_runs",
					SourceDir:        "cozepy/workflows/runs",
					ClientClass:      "WorkflowsRunsClient",
					AsyncClientClass: "AsyncWorkflowsRunsClient",
				},
				{
					Name:             "workflows_chat",
					SourceDir:        "cozepy/workflows/chat",
					ClientClass:      "WorkflowsChatClient",
					AsyncClientClass: "AsyncWorkflowsChatClient",
				},
			},
		},
	}

	metas := buildPackageMeta(cfg, nil)
	parent, ok := metas["workflows"]
	if !ok || parent.Package == nil {
		t.Fatalf("missing workflows package meta: %+v", parent)
	}

	childByAttr := map[string]config.ChildClient{}
	for _, child := range parent.Package.ChildClients {
		childByAttr[child.Attribute] = child
	}
	runs, ok := childByAttr["runs"]
	if !ok {
		t.Fatalf("expected inferred runs child client, got %+v", parent.Package.ChildClients)
	}
	if runs.Module != ".runs" || runs.SyncClass != "WorkflowsRunsClient" || runs.AsyncClass != "AsyncWorkflowsRunsClient" {
		t.Fatalf("unexpected inferred runs child config: %+v", runs)
	}
}

func TestBuildPackageMetaInferredChildDoesNotOverrideExplicit(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Packages: []config.Package{
				{
					Name:      "workflows",
					SourceDir: "cozepy/workflows",
					ChildClients: []config.ChildClient{
						{
							Attribute:  "runs",
							Module:     ".custom_runs",
							SyncClass:  "CustomRunsClient",
							AsyncClass: "AsyncCustomRunsClient",
						},
					},
				},
				{
					Name:             "workflows_runs",
					SourceDir:        "cozepy/workflows/runs",
					ClientClass:      "WorkflowsRunsClient",
					AsyncClientClass: "AsyncWorkflowsRunsClient",
				},
			},
		},
	}

	metas := buildPackageMeta(cfg, nil)
	parent, ok := metas["workflows"]
	if !ok || parent.Package == nil {
		t.Fatalf("missing workflows package meta: %+v", parent)
	}

	runsCount := 0
	for _, child := range parent.Package.ChildClients {
		if child.Attribute != "runs" {
			continue
		}
		runsCount++
		if child.Module != ".custom_runs" || child.SyncClass != "CustomRunsClient" || child.AsyncClass != "AsyncCustomRunsClient" {
			t.Fatalf("expected explicit runs child to be kept, got %+v", child)
		}
	}
	if runsCount != 1 {
		t.Fatalf("expected exactly one runs child client, got %d from %+v", runsCount, parent.Package.ChildClients)
	}
}

func TestBuildPackageMetaInferredChildSkipsExtraMethodName(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Packages: []config.Package{
				{
					Name:      "conversations_message",
					SourceDir: "cozepy/conversations/message",
					SyncExtraMethods: []string{
						`@property
def feedback(self):
    return self._feedback`,
					},
					AsyncExtraMethods: []string{
						`@property
def feedback(self):
    return self._feedback`,
					},
				},
				{
					Name:             "conversations_message_feedback",
					SourceDir:        "cozepy/conversations/message/feedback",
					ClientClass:      "ConversationsMessagesFeedbackClient",
					AsyncClientClass: "AsyncMessagesFeedbackClient",
				},
			},
		},
	}

	metas := buildPackageMeta(cfg, nil)
	parent, ok := metas["conversations_message"]
	if !ok || parent.Package == nil {
		t.Fatalf("missing conversations_message package meta: %+v", parent)
	}
	for _, child := range parent.Package.ChildClients {
		if child.Attribute == "feedback" {
			t.Fatalf("expected inferred feedback child to be skipped due extra method, got %+v", parent.Package.ChildClients)
		}
	}
}

func TestBuildPackageMetaInferredChildUsesPluralLexicon(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			Packages: []config.Package{
				{
					Name:      "chat",
					SourceDir: "cozepy/chat",
					ChildClients: []config.ChildClient{
						{
							Attribute:  "messages",
							Module:     ".message",
							SyncClass:  "ChatMessagesClient",
							AsyncClass: "AsyncChatMessagesClient",
						},
					},
				},
				{
					Name:             "chat_message",
					SourceDir:        "cozepy/chat/message",
					ClientClass:      "ChatMessagesClient",
					AsyncClientClass: "AsyncChatMessagesClient",
				},
			},
		},
	}

	metas := buildPackageMeta(cfg, nil)
	parent, ok := metas["chat"]
	if !ok || parent.Package == nil {
		t.Fatalf("missing chat package meta: %+v", parent)
	}

	messageAttrCount := 0
	messagesAttrCount := 0
	for _, child := range parent.Package.ChildClients {
		switch child.Attribute {
		case "message":
			messageAttrCount++
		case "messages":
			messagesAttrCount++
		}
	}
	if messageAttrCount != 0 {
		t.Fatalf("did not expect singular message child attribute, got %+v", parent.Package.ChildClients)
	}
	if messagesAttrCount != 1 {
		t.Fatalf("expected one messages child attribute, got %+v", parent.Package.ChildClients)
	}
}

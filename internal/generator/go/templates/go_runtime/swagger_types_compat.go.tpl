package coze

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
)

// PublishStatus is kept for compatibility with existing request signatures.
type PublishStatus string

// ToolOutput is used by websocket event payloads.
type ToolOutput struct {
	ToolCallID string `json:"tool_call_id,omitempty"`
	Output     string `json:"output,omitempty"`
}

// Chat is used by websocket event payloads.
type Chat struct {
	ID             string `json:"id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	Status         string `json:"status,omitempty"`
}

// ChatEvent is a generic stream event for chat stream APIs.
type ChatEvent struct {
	Event string          `json:"event,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

// WorkflowEvent is a generic stream event for workflow stream APIs.
type WorkflowEvent struct {
	Event string          `json:"event,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func parseChatEvent(_ context.Context, _ *core, line []byte, _ *bufio.Reader) (*ChatEvent, bool, error) {
	payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
	if len(payload) == 0 {
		return nil, false, nil
	}
	if bytes.Equal(payload, []byte("[DONE]")) {
		return nil, true, nil
	}
	event := &ChatEvent{}
	if err := json.Unmarshal(payload, event); err != nil {
		event.Data = append([]byte(nil), payload...)
	}
	return event, false, nil
}

func parseWorkflowEvent(_ context.Context, _ *core, line []byte, _ *bufio.Reader) (*WorkflowEvent, bool, error) {
	payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
	if len(payload) == 0 {
		return nil, false, nil
	}
	if bytes.Equal(payload, []byte("[DONE]")) {
		return nil, true, nil
	}
	event := &WorkflowEvent{}
	if err := json.Unmarshal(payload, event); err != nil {
		event.Data = append([]byte(nil), payload...)
	}
	return event, false, nil
}

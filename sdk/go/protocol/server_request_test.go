package protocol

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestServerRequestDecodeKnownCommandApproval(t *testing.T) {
	request := ServerRequest{
		Method: "item/commandExecution/requestApproval",
		Params: json.RawMessage(`{
			"threadId":"thread-1",
			"turnId":"turn-1",
			"itemId":"item-1",
			"command":"git status"
		}`),
	}

	payload, err := request.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*CommandExecutionRequestApprovalParams)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || typed.Command == nil || *typed.Command != "git status" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestServerRequestDecodeKnownCurrentSchemaMethods(t *testing.T) {
	tests := []struct {
		method string
		params string
		want   any
	}{
		{
			method: "account/chatgptAuthTokens/refresh",
			params: `{"previousAccountId":"acct-1","reason":"unauthorized"}`,
			want:   &ChatgptAuthTokensRefreshParams{},
		},
		{
			method: "applyPatchApproval",
			params: `{"callId":"call-1","conversationId":"thread-1","fileChanges":{"main.go":{"type":"update","unified_diff":"@@\\n"}}}`,
			want:   &ApplyPatchApprovalParams{},
		},
		{
			method: "attestation/generate",
			params: `{}`,
			want:   &AttestationGenerateParams{},
		},
		{
			method: "execCommandApproval",
			params: `{"callId":"call-1","conversationId":"thread-1","command":["git","status"],"cwd":"/tmp","parsedCmd":[]}`,
			want:   &ExecCommandApprovalParams{},
		},
		{
			method: "item/tool/call",
			params: `{"threadId":"thread-1","turnId":"turn-1","callId":"call-1","tool":"demo","arguments":{"x":1}}`,
			want:   &DynamicToolCallParams{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			request := ServerRequest{Method: tt.method, Params: json.RawMessage(tt.params)}
			payload, err := request.DecodeKnown()
			if err != nil {
				t.Fatalf("DecodeKnown returned error: %v", err)
			}
			if fmt.Sprintf("%T", payload) != fmt.Sprintf("%T", tt.want) {
				t.Fatalf("unexpected payload type: %T", payload)
			}
		})
	}
}

func TestServerRequestDecodeKnownUnknownMethod(t *testing.T) {
	request := ServerRequest{
		Method: "unknown/request",
		Params: json.RawMessage(`{}`),
	}
	if _, err := request.DecodeKnown(); err == nil {
		t.Fatal("expected unknown server request error")
	}
}

func TestServerRequestDecodeKnownPermissionsApproval(t *testing.T) {
	request := ServerRequest{
		Method: "item/permissions/requestApproval",
		Params: json.RawMessage(`{
			"threadId":"thread-1",
			"turnId":"turn-1",
			"itemId":"item-1",
			"permissions":{
				"network":{"enabled":true},
				"fileSystem":{"read":["/tmp/read"],"write":["/tmp/write"]}
			}
		}`),
	}

	payload, err := request.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*PermissionsRequestApprovalParams)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.Permissions.Network == nil || typed.Permissions.Network.Enabled == nil || !*typed.Permissions.Network.Enabled {
		t.Fatalf("expected typed network permissions, got %+v", typed.Permissions)
	}
	if typed.Permissions.FileSystem == nil || len(typed.Permissions.FileSystem.Read) != 1 || typed.Permissions.FileSystem.Read[0] != "/tmp/read" {
		t.Fatalf("expected typed filesystem permissions, got %+v", typed.Permissions)
	}
}

func TestServerRequestDecodeKnownToolRequestUserInput(t *testing.T) {
	request := ServerRequest{
		Method: "item/tool/requestUserInput",
		Params: json.RawMessage(`{
			"threadId":"thread-1",
			"turnId":"turn-1",
			"itemId":"item-1",
			"questions":[{"id":"q1","header":"Name","question":"Who are you?"}]
		}`),
	}

	payload, err := request.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*ToolRequestUserInputParams)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || len(typed.Questions) != 1 || typed.Questions[0].ID != "q1" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestServerRequestDecodeKnownMcpServerElicitation(t *testing.T) {
	request := ServerRequest{
		Method: "mcpServer/elicitation/request",
		Params: json.RawMessage(`{
			"threadId":"thread-1",
			"turnId":"turn-1",
			"serverName":"demo",
			"mode":"url",
			"message":"Open the browser",
			"url":"https://example.com/auth",
			"elicitationId":"elic-1"
		}`),
	}

	payload, err := request.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*McpServerElicitationRequestParams)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ServerName != "demo" || typed.Mode != "url" || typed.URL == nil || *typed.URL != "https://example.com/auth" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

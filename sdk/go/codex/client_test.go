package codex

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	irpc "github.com/openai/codex/sdk/go/internal/jsonrpc"
	"github.com/openai/codex/sdk/go/protocol"
)

func TestValidateInitializeParsesUserAgent(t *testing.T) {
	resp := protocol.InitializeResponse{UserAgent: "codex-cli/1.2.3"}
	if err := validateInitialize(&resp); err != nil {
		t.Fatalf("validateInitialize returned error: %v", err)
	}
	if resp.ServerInfo == nil {
		t.Fatal("expected server info to be populated")
	}
	if resp.ServerInfo.Name != "codex-cli" || resp.ServerInfo.Version != "1.2.3" {
		t.Fatalf("unexpected server info: %+v", resp.ServerInfo)
	}
}

func TestValidateInitializeRequiresMetadata(t *testing.T) {
	err := validateInitialize(&protocol.InitializeResponse{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestFinalAssistantResponsePrefersFinalAnswer(t *testing.T) {
	finalPhase := protocol.MessagePhaseFinalAnswer
	items := []protocol.ThreadItem{
		{Type: "agentMessage", Text: "draft"},
		{Type: "agentMessage", Text: "final", Phase: &finalPhase},
	}
	if got := finalAssistantResponse(items); got != "final" {
		t.Fatalf("finalAssistantResponse = %q, want %q", got, "final")
	}
}

func TestNormalizeRunInput(t *testing.T) {
	items, err := normalizeRunInput([]InputItem{
		TextInput{Text: "hello"},
		LocalImageInput{Path: "./ui.png"},
	})
	if err != nil {
		t.Fatalf("normalizeRunInput returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Type != "text" || items[1].Type != "localImage" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestDefaultApprovalHandlerAcceptsKnownMethods(t *testing.T) {
	handler := effectiveApprovalHandler(nil)
	result, err := handler(context.Background(), "item/commandExecution/requestApproval", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("approval handler returned error: %v", err)
	}
	response, ok := result.(protocol.CommandExecutionRequestApprovalResponse)
	if !ok || response.Decision != protocol.CommandExecutionApprovalDecisionAccept {
		t.Fatalf("unexpected approval response: %#v", result)
	}
}

func TestIsRetryableError(t *testing.T) {
	retryable := &irpc.Error{Code: -32001, Message: "Server overloaded; retry later."}
	if !IsRetryableError(retryable) {
		t.Fatal("expected overload error to be retryable")
	}
	if IsRetryableError(errors.New("boom")) {
		t.Fatal("plain errors should not be retryable")
	}
}

func TestClientNotificationsIncludeThreadScopedEvents(t *testing.T) {
	client := &Client{
		threadSubs: make(map[string][]chan protocol.Notification),
		turnSubs:   make(map[string][]chan protocol.Notification),
		closed:     make(chan struct{}),
	}

	thread := &Thread{client: client, ID: "thread-1"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, errs := thread.Notifications(ctx)
	client.dispatch(protocol.Notification{
		Method: "thread/status/changed",
		Params: json.RawMessage(`{"threadId":"thread-1","status":"loaded"}`),
	})

	select {
	case event := <-events:
		if event.Method != "thread/status/changed" {
			t.Fatalf("unexpected event method: %s", event.Method)
		}
	case err := <-errs:
		t.Fatalf("unexpected stream error: %v", err)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for thread-scoped event")
	}
}

func TestTurnHandleRunReturnsCanonicalCompletedTurn(t *testing.T) {
	client := &Client{
		threadSubs: make(map[string][]chan protocol.Notification),
		turnSubs:   make(map[string][]chan protocol.Notification),
		closed:     make(chan struct{}),
	}

	handle := &TurnHandle{client: client, ThreadID: "thread-1", TurnID: "turn-1"}

	go func() {
		client.dispatch(protocol.Notification{
			Method: "item/completed",
			Params: json.RawMessage(`{
				"threadId":"thread-1",
				"turnId":"turn-1",
				"item":{"id":"item-1","type":"agentMessage","text":"hello"}
			}`),
		})
		client.dispatch(protocol.Notification{
			Method: "turn/completed",
			Params: json.RawMessage(`{
				"threadId":"thread-1",
				"turn":{
					"id":"turn-1",
					"status":"completed",
					"items":[{"id":"server-item","type":"agentMessage","text":"canonical"}]
				}
			}`),
		})
	}()

	turn, err := handle.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if turn.ID != "turn-1" || turn.Status != protocol.TurnStatusCompleted {
		t.Fatalf("unexpected completed turn: %+v", turn)
	}
	if len(turn.Items) != 1 || turn.Items[0].ID != "server-item" {
		t.Fatalf("expected canonical server turn items, got %+v", turn.Items)
	}
}

func TestDeliverResponseMatchesIntegerRequestID(t *testing.T) {
	client := &Client{
		pending:    make(map[string]chan responseResult),
		threadSubs: make(map[string][]chan protocol.Notification),
		turnSubs:   make(map[string][]chan protocol.Notification),
		closed:     make(chan struct{}),
	}

	waiter := make(chan responseResult, 1)
	client.pending["i:7"] = waiter

	client.deliverResponse(irpc.NewIntegerRequestID(7), json.RawMessage(`{"ok":true}`), nil)

	select {
	case result := <-waiter:
		if string(result.result) != `{"ok":true}` {
			t.Fatalf("unexpected delivered result: %s", result.result)
		}
	default:
		t.Fatal("expected waiter to receive integer-id response")
	}
}

func TestDecodeServerRequest(t *testing.T) {
	payload, err := DecodeServerRequest(
		"item/fileChange/requestApproval",
		json.RawMessage(`{"threadId":"thread-1","turnId":"turn-1","itemId":"item-1","reason":"review changes"}`),
	)
	if err != nil {
		t.Fatalf("DecodeServerRequest returned error: %v", err)
	}

	typed, ok := payload.(*protocol.FileChangeRequestApprovalParams)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || typed.ItemId != "item-1" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestDefaultServerRequestResponseToolInput(t *testing.T) {
	response := DefaultServerRequestResponse("item/tool/requestUserInput")
	typed, ok := response.(protocol.ToolRequestUserInputResponse)
	if !ok {
		t.Fatalf("unexpected response type: %T", response)
	}
	if len(typed.Answers) != 0 {
		t.Fatalf("expected empty answers map, got %+v", typed.Answers)
	}
}

func TestDefaultServerRequestResponseMcpServerElicitation(t *testing.T) {
	response := DefaultServerRequestResponse("mcpServer/elicitation/request")
	typed, ok := response.(protocol.McpServerElicitationRequestResponse)
	if !ok {
		t.Fatalf("unexpected response type: %T", response)
	}
	if typed.Action != protocol.McpServerElicitationActionCancel {
		t.Fatalf("expected cancel action, got %q", typed.Action)
	}
}

func TestDefaultServerRequestResponsePermissions(t *testing.T) {
	response := DefaultServerRequestResponse("item/permissions/requestApproval")
	typed, ok := response.(protocol.PermissionsRequestApprovalResponse)
	if !ok {
		t.Fatalf("unexpected response type: %T", response)
	}
	if typed.Scope != protocol.PermissionGrantScopeTurn {
		t.Fatalf("expected turn scope, got %q", typed.Scope)
	}
	if typed.Permissions.Network != nil || typed.Permissions.FileSystem != nil {
		t.Fatalf("expected empty granted permission profile, got %+v", typed.Permissions)
	}
}

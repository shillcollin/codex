package protocol

import (
	"encoding/json"
	"testing"
)

func TestNotificationThreadIDHandlesNestedThreadObject(t *testing.T) {
	notification := Notification{
		Method: "thread/started",
		Params: json.RawMessage(`{"thread":{"id":"thread-1"}}`),
	}
	if got := notification.ThreadID(); got != "thread-1" {
		t.Fatalf("ThreadID = %q, want %q", got, "thread-1")
	}
}

func TestNotificationDecodeKnown(t *testing.T) {
	notification := Notification{
		Method: "turn/started",
		Params: json.RawMessage(`{"threadId":"thread-1","turn":{"id":"turn-1","status":"running"}}`),
	}

	payload, err := notification.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*TurnStartedNotification)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || typed.Turn.ID != "turn-1" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestNotificationDecodeKnownRealtimeTranscript(t *testing.T) {
	notification := Notification{
		Method: "thread/realtime/transcript/delta",
		Params: json.RawMessage(`{"threadId":"thread-1","role":"assistant","delta":"hello"}`),
	}

	payload, err := notification.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*ThreadRealtimeTranscriptDeltaNotification)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || typed.Role != "assistant" || typed.Delta != "hello" {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestNotificationDecodeKnownThreadGoalUpdated(t *testing.T) {
	notification := Notification{
		Method: "thread/goal/updated",
		Params: json.RawMessage(`{
			"threadId":"thread-1",
			"turnId":"turn-1",
			"goal":{
				"threadId":"thread-1",
				"objective":"ship the Go SDK",
				"status":"active",
				"tokenBudget":1000,
				"tokensUsed":12,
				"timeUsedSeconds":3,
				"createdAt":1,
				"updatedAt":2
			}
		}`),
	}

	payload, err := notification.DecodeKnown()
	if err != nil {
		t.Fatalf("DecodeKnown returned error: %v", err)
	}

	typed, ok := payload.(*ThreadGoalUpdatedNotification)
	if !ok {
		t.Fatalf("unexpected payload type: %T", payload)
	}
	if typed.ThreadId != "thread-1" || typed.Goal.Objective != "ship the Go SDK" || typed.Goal.Status != ThreadGoalStatusActive {
		t.Fatalf("unexpected decoded payload: %+v", typed)
	}
}

func TestNotificationDecodeKnownRejectsUnknownMethod(t *testing.T) {
	notification := Notification{
		Method: "not/a-real-notification",
		Params: json.RawMessage(`{}`),
	}
	if _, err := notification.DecodeKnown(); err == nil {
		t.Fatal("expected unknown notification error")
	}
}

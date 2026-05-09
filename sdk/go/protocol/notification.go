package protocol

import (
	"encoding/json"
	"fmt"
)

type Notification struct {
	Method string
	Params json.RawMessage
}

type NotificationPayload interface{}

var knownNotificationPayloads = map[string]func() NotificationPayload{
	"account/login/completed":    func() NotificationPayload { return &AccountLoginCompletedNotification{} },
	"account/rateLimits/updated": func() NotificationPayload { return &AccountRateLimitsUpdatedNotification{} },
	"account/updated":            func() NotificationPayload { return &AccountUpdatedNotification{} },
	"app/list/updated":           func() NotificationPayload { return &AppListUpdatedNotification{} },
	"command/exec/outputDelta":   func() NotificationPayload { return &CommandExecOutputDeltaNotification{} },
	"configWarning":              func() NotificationPayload { return &ConfigWarningNotification{} },
	"deprecationNotice":          func() NotificationPayload { return &DeprecationNoticeNotification{} },
	"error":                      func() NotificationPayload { return &ErrorNotification{} },
	"externalAgentConfig/import/completed": func() NotificationPayload {
		return new(ExternalAgentConfigImportCompletedNotification)
	},
	"fs/changed":                        func() NotificationPayload { return &FsChangedNotification{} },
	"fuzzyFileSearch/sessionCompleted":  func() NotificationPayload { return &FuzzyFileSearchSessionCompletedNotification{} },
	"fuzzyFileSearch/sessionUpdated":    func() NotificationPayload { return &FuzzyFileSearchSessionUpdatedNotification{} },
	"guardianWarning":                   func() NotificationPayload { return &GuardianWarningNotification{} },
	"hook/completed":                    func() NotificationPayload { return &HookCompletedNotification{} },
	"hook/started":                      func() NotificationPayload { return &HookStartedNotification{} },
	"item/agentMessage/delta":           func() NotificationPayload { return &AgentMessageDeltaNotification{} },
	"item/autoApprovalReview/completed": func() NotificationPayload { return &ItemGuardianApprovalReviewCompletedNotification{} },
	"item/autoApprovalReview/started":   func() NotificationPayload { return &ItemGuardianApprovalReviewStartedNotification{} },
	"item/commandExecution/outputDelta": func() NotificationPayload { return &CommandExecutionOutputDeltaNotification{} },
	"item/commandExecution/terminalInteraction": func() NotificationPayload {
		return &TerminalInteractionNotification{}
	},
	"item/completed":                  func() NotificationPayload { return &ItemCompletedNotification{} },
	"item/fileChange/outputDelta":     func() NotificationPayload { return &FileChangeOutputDeltaNotification{} },
	"item/fileChange/patchUpdated":    func() NotificationPayload { return &FileChangePatchUpdatedNotification{} },
	"item/mcpToolCall/progress":       func() NotificationPayload { return &McpToolCallProgressNotification{} },
	"item/plan/delta":                 func() NotificationPayload { return &PlanDeltaNotification{} },
	"item/reasoning/summaryPartAdded": func() NotificationPayload { return &ReasoningSummaryPartAddedNotification{} },
	"item/reasoning/summaryTextDelta": func() NotificationPayload { return &ReasoningSummaryTextDeltaNotification{} },
	"item/reasoning/textDelta":        func() NotificationPayload { return &ReasoningTextDeltaNotification{} },
	"item/started":                    func() NotificationPayload { return &ItemStartedNotification{} },
	"mcpServer/oauthLogin/completed":  func() NotificationPayload { return &McpServerOauthLoginCompletedNotification{} },
	"mcpServer/startupStatus/updated": func() NotificationPayload { return &McpServerStatusUpdatedNotification{} },
	"model/rerouted":                  func() NotificationPayload { return &ModelReroutedNotification{} },
	"model/verification":              func() NotificationPayload { return &ModelVerificationNotification{} },
	"process/exited":                  func() NotificationPayload { return &ProcessExitedNotification{} },
	"process/outputDelta":             func() NotificationPayload { return &ProcessOutputDeltaNotification{} },
	"remoteControl/status/changed":    func() NotificationPayload { return &RemoteControlStatusChangedNotification{} },
	"serverRequest/resolved":          func() NotificationPayload { return &ServerRequestResolvedNotification{} },
	"skills/changed":                  func() NotificationPayload { return &SkillsChangedNotification{} },
	"thread/archived":                 func() NotificationPayload { return &ThreadArchivedNotification{} },
	"thread/closed":                   func() NotificationPayload { return &ThreadClosedNotification{} },
	"thread/compacted":                func() NotificationPayload { return &ContextCompactedNotification{} },
	"thread/goal/cleared":             func() NotificationPayload { return &ThreadGoalClearedNotification{} },
	"thread/goal/updated":             func() NotificationPayload { return &ThreadGoalUpdatedNotification{} },
	"thread/name/updated":             func() NotificationPayload { return &ThreadNameUpdatedNotification{} },
	"thread/realtime/closed":          func() NotificationPayload { return &ThreadRealtimeClosedNotification{} },
	"thread/realtime/error":           func() NotificationPayload { return &ThreadRealtimeErrorNotification{} },
	"thread/realtime/itemAdded":       func() NotificationPayload { return &ThreadRealtimeItemAddedNotification{} },
	"thread/realtime/outputAudio/delta": func() NotificationPayload {
		return &ThreadRealtimeOutputAudioDeltaNotification{}
	},
	"thread/realtime/sdp":     func() NotificationPayload { return &ThreadRealtimeSdpNotification{} },
	"thread/realtime/started": func() NotificationPayload { return &ThreadRealtimeStartedNotification{} },
	"thread/realtime/transcript/delta": func() NotificationPayload {
		return &ThreadRealtimeTranscriptDeltaNotification{}
	},
	"thread/realtime/transcript/done": func() NotificationPayload {
		return &ThreadRealtimeTranscriptDoneNotification{}
	},
	"thread/started":            func() NotificationPayload { return &ThreadStartedNotification{} },
	"thread/status/changed":     func() NotificationPayload { return &ThreadStatusChangedNotification{} },
	"thread/tokenUsage/updated": func() NotificationPayload { return &ThreadTokenUsageUpdatedNotification{} },
	"thread/unarchived":         func() NotificationPayload { return &ThreadUnarchivedNotification{} },
	"turn/completed":            func() NotificationPayload { return &TurnCompletedNotification{} },
	"turn/diff/updated":         func() NotificationPayload { return &TurnDiffUpdatedNotification{} },
	"turn/plan/updated":         func() NotificationPayload { return &TurnPlanUpdatedNotification{} },
	"turn/started":              func() NotificationPayload { return &TurnStartedNotification{} },
	"warning":                   func() NotificationPayload { return &WarningNotification{} },
	"windows/worldWritableWarning": func() NotificationPayload {
		return &WindowsWorldWritableWarningNotification{}
	},
	"windowsSandbox/setupCompleted": func() NotificationPayload { return &WindowsSandboxSetupCompletedNotification{} },
}

func (n Notification) Decode(v any) error {
	return json.Unmarshal(n.Params, v)
}

func (n Notification) DecodeKnown() (NotificationPayload, error) {
	factory := knownNotificationPayloads[n.Method]
	if factory == nil {
		return nil, fmt.Errorf("unknown notification method: %s", n.Method)
	}
	payload := factory()
	if err := n.Decode(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (n Notification) IsKnown() bool {
	_, ok := knownNotificationPayloads[n.Method]
	return ok
}

func (n Notification) ThreadID() string {
	var envelope struct {
		ThreadID string `json:"threadId"`
	}
	if err := json.Unmarshal(n.Params, &envelope); err == nil && envelope.ThreadID != "" {
		return envelope.ThreadID
	}

	var nested struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	_ = json.Unmarshal(n.Params, &nested)
	return nested.Thread.ID
}

func (n Notification) TurnID() string {
	var direct struct {
		TurnID string `json:"turnId"`
	}
	if err := json.Unmarshal(n.Params, &direct); err == nil && direct.TurnID != "" {
		return direct.TurnID
	}

	var nested struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	_ = json.Unmarshal(n.Params, &nested)
	return nested.Turn.ID
}

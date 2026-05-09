package protocol

import (
	"encoding/json"
	"fmt"
)

type ServerRequest struct {
	Method string
	Params json.RawMessage
}

type ServerRequestPayload interface{}

type ChatgptAuthTokensRefreshReason string

const (
	ChatgptAuthTokensRefreshReasonUnauthorized ChatgptAuthTokensRefreshReason = "unauthorized"
)

type ChatgptAuthTokensRefreshParams struct {
	PreviousAccountId *string                        `json:"previousAccountId,omitempty"`
	Reason            ChatgptAuthTokensRefreshReason `json:"reason"`
}

type ChatgptAuthTokensRefreshResponse struct {
	AccessToken      string  `json:"accessToken"`
	ChatgptAccountId string  `json:"chatgptAccountId"`
	ChatgptPlanType  *string `json:"chatgptPlanType,omitempty"`
}

type ApplyPatchApprovalParams struct {
	CallId         string                     `json:"callId"`
	ConversationId string                     `json:"conversationId"`
	FileChanges    map[string]json.RawMessage `json:"fileChanges"`
	GrantRoot      *string                    `json:"grantRoot,omitempty"`
	Reason         *string                    `json:"reason,omitempty"`
}

type ReviewDecision = json.RawMessage

type ApplyPatchApprovalResponse struct {
	Decision ReviewDecision `json:"decision"`
}

type ExecCommandApprovalParams struct {
	ApprovalId     *string           `json:"approvalId,omitempty"`
	CallId         string            `json:"callId"`
	Command        []string          `json:"command"`
	ConversationId string            `json:"conversationId"`
	Cwd            string            `json:"cwd"`
	ParsedCmd      []json.RawMessage `json:"parsedCmd"`
	Reason         *string           `json:"reason,omitempty"`
}

type ExecCommandApprovalResponse struct {
	Decision ReviewDecision `json:"decision"`
}

type AttestationGenerateParams struct{}

type AttestationGenerateResponse struct {
	Token string `json:"token"`
}

type CommandExecutionApprovalDecision string

const (
	CommandExecutionApprovalDecisionAccept           CommandExecutionApprovalDecision = "accept"
	CommandExecutionApprovalDecisionAcceptForSession CommandExecutionApprovalDecision = "acceptForSession"
	CommandExecutionApprovalDecisionDecline          CommandExecutionApprovalDecision = "decline"
	CommandExecutionApprovalDecisionCancel           CommandExecutionApprovalDecision = "cancel"
)

type FileChangeApprovalDecision string

const (
	FileChangeApprovalDecisionAccept           FileChangeApprovalDecision = "accept"
	FileChangeApprovalDecisionAcceptForSession FileChangeApprovalDecision = "acceptForSession"
	FileChangeApprovalDecisionDecline          FileChangeApprovalDecision = "decline"
	FileChangeApprovalDecisionCancel           FileChangeApprovalDecision = "cancel"
)

type PermissionGrantScope string

const (
	PermissionGrantScopeTurn    PermissionGrantScope = "turn"
	PermissionGrantScopeSession PermissionGrantScope = "session"
)

type AdditionalPermissionProfile struct {
	Network    *AdditionalNetworkPermissions    `json:"network,omitempty"`
	FileSystem *AdditionalFileSystemPermissions `json:"fileSystem,omitempty"`
}

type GrantedPermissionProfile struct {
	Network    *AdditionalNetworkPermissions    `json:"network,omitempty"`
	FileSystem *AdditionalFileSystemPermissions `json:"fileSystem,omitempty"`
}

type CommandExecutionRequestApprovalParams struct {
	ThreadId                        string                       `json:"threadId"`
	TurnId                          string                       `json:"turnId"`
	ItemId                          string                       `json:"itemId"`
	ApprovalId                      *string                      `json:"approvalId,omitempty"`
	Reason                          *string                      `json:"reason,omitempty"`
	NetworkApprovalContext          json.RawMessage              `json:"networkApprovalContext,omitempty"`
	Command                         *string                      `json:"command,omitempty"`
	Cwd                             *string                      `json:"cwd,omitempty"`
	CommandActions                  []json.RawMessage            `json:"commandActions,omitempty"`
	AdditionalPermissions           *AdditionalPermissionProfile `json:"additionalPermissions,omitempty"`
	ProposedExecpolicyAmendment     json.RawMessage              `json:"proposedExecpolicyAmendment,omitempty"`
	ProposedNetworkPolicyAmendments []json.RawMessage            `json:"proposedNetworkPolicyAmendments,omitempty"`
	AvailableDecisions              []json.RawMessage            `json:"availableDecisions,omitempty"`
}

type CommandExecutionRequestApprovalResponse struct {
	Decision CommandExecutionApprovalDecision `json:"decision"`
}

type FileChangeRequestApprovalParams struct {
	ThreadId  string  `json:"threadId"`
	TurnId    string  `json:"turnId"`
	ItemId    string  `json:"itemId"`
	Reason    *string `json:"reason,omitempty"`
	GrantRoot *string `json:"grantRoot,omitempty"`
}

type FileChangeRequestApprovalResponse struct {
	Decision FileChangeApprovalDecision `json:"decision"`
}

type PermissionsRequestApprovalParams struct {
	ThreadId    string                   `json:"threadId"`
	TurnId      string                   `json:"turnId"`
	ItemId      string                   `json:"itemId"`
	Reason      *string                  `json:"reason,omitempty"`
	Permissions RequestPermissionProfile `json:"permissions"`
}

type PermissionsRequestApprovalResponse struct {
	Permissions GrantedPermissionProfile `json:"permissions"`
	Scope       PermissionGrantScope     `json:"scope,omitempty"`
}

type ToolRequestUserInputOption struct {
	Label       string `json:"label"`
	Description string `json:"description"`
}

type ToolRequestUserInputQuestion struct {
	ID       string                       `json:"id"`
	Header   string                       `json:"header"`
	Question string                       `json:"question"`
	IsOther  bool                         `json:"isOther,omitempty"`
	IsSecret bool                         `json:"isSecret,omitempty"`
	Options  []ToolRequestUserInputOption `json:"options,omitempty"`
}

type ToolRequestUserInputParams struct {
	ThreadId  string                         `json:"threadId"`
	TurnId    string                         `json:"turnId"`
	ItemId    string                         `json:"itemId"`
	Questions []ToolRequestUserInputQuestion `json:"questions"`
}

type ToolRequestUserInputAnswer struct {
	Answers []string `json:"answers"`
}

type ToolRequestUserInputResponse struct {
	Answers map[string]ToolRequestUserInputAnswer `json:"answers"`
}

type McpServerElicitationAction string

const (
	McpServerElicitationActionAccept  McpServerElicitationAction = "accept"
	McpServerElicitationActionDecline McpServerElicitationAction = "decline"
	McpServerElicitationActionCancel  McpServerElicitationAction = "cancel"
)

type McpServerElicitationRequestParams struct {
	ThreadId        string          `json:"threadId"`
	TurnId          *string         `json:"turnId,omitempty"`
	ServerName      string          `json:"serverName"`
	Mode            string          `json:"mode"`
	Message         string          `json:"message"`
	RequestedSchema json.RawMessage `json:"requestedSchema,omitempty"`
	URL             *string         `json:"url,omitempty"`
	ElicitationId   *string         `json:"elicitationId,omitempty"`
	Meta            json.RawMessage `json:"_meta,omitempty"`
}

type McpServerElicitationRequestResponse struct {
	Action  McpServerElicitationAction `json:"action"`
	Content json.RawMessage            `json:"content,omitempty"`
	Meta    json.RawMessage            `json:"_meta,omitempty"`
}

type DynamicToolCallParams struct {
	Arguments json.RawMessage `json:"arguments"`
	CallId    string          `json:"callId"`
	ThreadId  string          `json:"threadId"`
	Tool      string          `json:"tool"`
	TurnId    string          `json:"turnId"`
}

type DynamicToolCallResponse struct {
	ContentItems []json.RawMessage `json:"contentItems"`
	Success      bool              `json:"success"`
}

var knownServerRequestPayloads = map[string]func() ServerRequestPayload{
	"account/chatgptAuthTokens/refresh":     func() ServerRequestPayload { return &ChatgptAuthTokensRefreshParams{} },
	"applyPatchApproval":                    func() ServerRequestPayload { return &ApplyPatchApprovalParams{} },
	"attestation/generate":                  func() ServerRequestPayload { return &AttestationGenerateParams{} },
	"execCommandApproval":                   func() ServerRequestPayload { return &ExecCommandApprovalParams{} },
	"item/commandExecution/requestApproval": func() ServerRequestPayload { return &CommandExecutionRequestApprovalParams{} },
	"item/fileChange/requestApproval":       func() ServerRequestPayload { return &FileChangeRequestApprovalParams{} },
	"item/permissions/requestApproval":      func() ServerRequestPayload { return &PermissionsRequestApprovalParams{} },
	"item/tool/call":                        func() ServerRequestPayload { return &DynamicToolCallParams{} },
	"item/tool/requestUserInput":            func() ServerRequestPayload { return &ToolRequestUserInputParams{} },
	"mcpServer/elicitation/request":         func() ServerRequestPayload { return &McpServerElicitationRequestParams{} },
}

func (r ServerRequest) Decode(v any) error {
	return json.Unmarshal(r.Params, v)
}

func (r ServerRequest) DecodeKnown() (ServerRequestPayload, error) {
	factory := knownServerRequestPayloads[r.Method]
	if factory == nil {
		return nil, fmt.Errorf("unknown server request method: %s", r.Method)
	}
	payload := factory()
	if err := r.Decode(payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func (r ServerRequest) IsKnown() bool {
	_, ok := knownServerRequestPayloads[r.Method]
	return ok
}

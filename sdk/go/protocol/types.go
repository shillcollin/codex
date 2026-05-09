package protocol

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ClientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version,omitempty"`
}

type InitializeCapabilities struct {
	ExperimentalAPI          bool     `json:"experimentalApi,omitempty"`
	OptOutNotificationMethod []string `json:"optOutNotificationMethods,omitempty"`
}

type InitializeParams struct {
	ClientInfo   ClientInfo              `json:"clientInfo"`
	Capabilities *InitializeCapabilities `json:"capabilities,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	UserAgent      string      `json:"userAgent,omitempty"`
	CodexHome      string      `json:"codexHome,omitempty"`
	PlatformFamily string      `json:"platformFamily,omitempty"`
	PlatformOS     string      `json:"platformOs,omitempty"`
	ServerInfo     *ServerInfo `json:"serverInfo,omitempty"`
}

type RequestId struct {
	stringValue  *string
	integerValue *int64
}

func NewStringRequestId(value string) RequestId {
	return RequestId{stringValue: &value}
}

func NewIntegerRequestId(value int64) RequestId {
	return RequestId{integerValue: &value}
}

func (id RequestId) String() string {
	if id.stringValue != nil {
		return *id.stringValue
	}
	if id.integerValue != nil {
		return strconv.FormatInt(*id.integerValue, 10)
	}
	return ""
}

func (id RequestId) Key() string {
	if id.stringValue != nil {
		return "s:" + *id.stringValue
	}
	if id.integerValue != nil {
		return "i:" + strconv.FormatInt(*id.integerValue, 10)
	}
	return ""
}

func (id RequestId) MarshalJSON() ([]byte, error) {
	if id.stringValue != nil {
		return json.Marshal(*id.stringValue)
	}
	if id.integerValue != nil {
		return json.Marshal(*id.integerValue)
	}
	return []byte("null"), nil
}

func (id *RequestId) UnmarshalJSON(data []byte) error {
	var stringValue string
	if err := json.Unmarshal(data, &stringValue); err == nil {
		*id = NewStringRequestId(stringValue)
		return nil
	}
	var integerValue int64
	if err := json.Unmarshal(data, &integerValue); err == nil {
		*id = NewIntegerRequestId(integerValue)
		return nil
	}
	return fmt.Errorf("unsupported request id: %s", string(data))
}

type ResourceContent struct {
	URI      string          `json:"uri"`
	MimeType *string         `json:"mimeType,omitempty"`
	Text     *string         `json:"text,omitempty"`
	Blob     *string         `json:"blob,omitempty"`
	Meta     json.RawMessage `json:"_meta,omitempty"`
}

func (r ResourceContent) MarshalJSON() ([]byte, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}
	type alias ResourceContent
	return json.Marshal(alias(r))
}

func (r *ResourceContent) UnmarshalJSON(data []byte) error {
	type alias ResourceContent
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	resource := ResourceContent(decoded)
	if err := resource.validate(); err != nil {
		return err
	}
	*r = resource
	return nil
}

func (r ResourceContent) validate() error {
	switch {
	case r.URI == "":
		return fmt.Errorf("resource content missing uri")
	case r.Text != nil && r.Blob != nil:
		return fmt.Errorf("resource content cannot contain both text and blob")
	case r.Text == nil && r.Blob == nil:
		return fmt.Errorf("resource content must contain text or blob")
	default:
		return nil
	}
}

type AskForApproval string

const (
	AskForApprovalUntrusted AskForApproval = "untrusted"
	AskForApprovalNever     AskForApproval = "never"
	AskForApprovalOnFailure AskForApproval = "on-failure"
	AskForApprovalOnFail    AskForApproval = AskForApprovalOnFailure
	AskForApprovalOnRequest AskForApproval = "on-request"
)

type ApprovalsReviewer string

type Personality string

const (
	PersonalityFriendly  Personality = "friendly"
	PersonalityPragmatic Personality = "pragmatic"
	PersonalityNone      Personality = "none"
)

type SandboxMode string

const (
	SandboxModeReadOnly         SandboxMode = "read-only"
	SandboxModeWorkspaceWrite   SandboxMode = "workspace-write"
	SandboxModeDangerFullAccess SandboxMode = "danger-full-access"
)

type ReasoningEffort string

type ReasoningSummary string

type ServiceTier string

type SandboxPolicy = json.RawMessage

type ThreadStatus string

const (
	ThreadStatusNotLoaded ThreadStatus = "notLoaded"
	ThreadStatusLoading   ThreadStatus = "loading"
	ThreadStatusLoaded    ThreadStatus = "loaded"
)

type ThreadSortKey string

type ThreadSourceKind string

type UserInput struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
	Name string `json:"name,omitempty"`
}

type Thread struct {
	ID            string       `json:"id"`
	Preview       string       `json:"preview,omitempty"`
	ModelProvider string       `json:"modelProvider,omitempty"`
	CreatedAt     int64        `json:"createdAt,omitempty"`
	Archived      bool         `json:"archived,omitempty"`
	Ephemeral     bool         `json:"ephemeral,omitempty"`
	Path          *string      `json:"path,omitempty"`
	Status        ThreadStatus `json:"status,omitempty"`
	ForkedFromID  string       `json:"forkedFromId,omitempty"`
	Turns         []Turn       `json:"turns,omitempty"`
}

type ThreadStartParams struct {
	ApprovalPolicy        *AskForApproval    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	Config                map[string]any     `json:"config,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	Ephemeral             *bool              `json:"ephemeral,omitempty"`
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	Personality           *Personality       `json:"personality,omitempty"`
	Sandbox               *SandboxMode       `json:"sandbox,omitempty"`
	ServiceName           *string            `json:"serviceName,omitempty"`
	ServiceTier           *ServiceTier       `json:"serviceTier,omitempty"`
}

type ThreadResumeParams struct {
	ApprovalPolicy        *AskForApproval    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	Config                map[string]any     `json:"config,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	Personality           *Personality       `json:"personality,omitempty"`
	Sandbox               *SandboxMode       `json:"sandbox,omitempty"`
	ServiceTier           *ServiceTier       `json:"serviceTier,omitempty"`
}

type ThreadForkParams struct {
	ApprovalPolicy        *AskForApproval    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer     *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	BaseInstructions      *string            `json:"baseInstructions,omitempty"`
	Config                map[string]any     `json:"config,omitempty"`
	Cwd                   *string            `json:"cwd,omitempty"`
	DeveloperInstructions *string            `json:"developerInstructions,omitempty"`
	Ephemeral             *bool              `json:"ephemeral,omitempty"`
	Model                 *string            `json:"model,omitempty"`
	ModelProvider         *string            `json:"modelProvider,omitempty"`
	Sandbox               *SandboxMode       `json:"sandbox,omitempty"`
	ServiceTier           *ServiceTier       `json:"serviceTier,omitempty"`
}

type ThreadListParams struct {
	Archived       *bool              `json:"archived,omitempty"`
	Cursor         *string            `json:"cursor,omitempty"`
	Cwd            *string            `json:"cwd,omitempty"`
	Limit          *uint32            `json:"limit,omitempty"`
	ModelProviders []string           `json:"modelProviders,omitempty"`
	SearchTerm     *string            `json:"searchTerm,omitempty"`
	SortKey        *ThreadSortKey     `json:"sortKey,omitempty"`
	SourceKinds    []ThreadSourceKind `json:"sourceKinds,omitempty"`
}

type ThreadStartResponse struct {
	Thread Thread `json:"thread"`
}

type ThreadResumeResponse struct {
	Thread Thread `json:"thread"`
}

type ThreadForkResponse struct {
	Thread Thread `json:"thread"`
}

type ThreadListResponse struct {
	Data       []Thread `json:"data"`
	NextCursor *string  `json:"nextCursor,omitempty"`
}

type ThreadReadResponse struct {
	Thread Thread `json:"thread"`
}

type ThreadArchiveResponse struct{}

type ThreadUnarchiveResponse struct {
	Thread Thread `json:"thread"`
}

type ThreadSetNameResponse struct{}

type ThreadCompactStartResponse struct{}

type Model struct {
	ID      string          `json:"id"`
	Title   string          `json:"title,omitempty"`
	Hidden  bool            `json:"hidden,omitempty"`
	RawJSON json.RawMessage `json:"-"`
}

func (m *Model) UnmarshalJSON(data []byte) error {
	type alias Model
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = Model(decoded)
	m.RawJSON = append(m.RawJSON[:0], data...)
	return nil
}

type ModelListResponse struct {
	Data []Model `json:"data"`
}

type TurnStatus string

const (
	TurnStatusCompleted   TurnStatus = "completed"
	TurnStatusFailed      TurnStatus = "failed"
	TurnStatusInterrupted TurnStatus = "interrupted"
	TurnStatusRunning     TurnStatus = "running"
)

type MessagePhase string

const (
	MessagePhaseFinalAnswer MessagePhase = "final_answer"
)

type ThreadItem struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type,omitempty"`
	Text    string          `json:"text,omitempty"`
	Phase   *MessagePhase   `json:"phase,omitempty"`
	RawJSON json.RawMessage `json:"-"`
}

func (i *ThreadItem) UnmarshalJSON(data []byte) error {
	type alias ThreadItem
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*i = ThreadItem(decoded)
	i.RawJSON = append(i.RawJSON[:0], data...)
	return nil
}

type TurnError struct {
	Message string `json:"message,omitempty"`
}

type Turn struct {
	ID      string          `json:"id"`
	Items   []ThreadItem    `json:"items,omitempty"`
	Status  TurnStatus      `json:"status,omitempty"`
	Error   *TurnError      `json:"error,omitempty"`
	RawJSON json.RawMessage `json:"-"`
}

func (t *Turn) UnmarshalJSON(data []byte) error {
	type alias Turn
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*t = Turn(decoded)
	t.RawJSON = append(t.RawJSON[:0], data...)
	return nil
}

type TurnStartParams struct {
	ThreadID          string             `json:"threadId"`
	Input             []UserInput        `json:"input"`
	ApprovalPolicy    *AskForApproval    `json:"approvalPolicy,omitempty"`
	ApprovalsReviewer *ApprovalsReviewer `json:"approvalsReviewer,omitempty"`
	Cwd               *string            `json:"cwd,omitempty"`
	Effort            *ReasoningEffort   `json:"effort,omitempty"`
	Model             *string            `json:"model,omitempty"`
	OutputSchema      map[string]any     `json:"outputSchema,omitempty"`
	Personality       *Personality       `json:"personality,omitempty"`
	SandboxPolicy     SandboxPolicy      `json:"sandboxPolicy,omitempty"`
	ServiceTier       *ServiceTier       `json:"serviceTier,omitempty"`
	Summary           *ReasoningSummary  `json:"summary,omitempty"`
}

type TurnStartResponse struct {
	Turn Turn `json:"turn"`
}

type TurnInterruptResponse struct{}

type TurnSteerResponse struct{}

type ThreadTokenUsageSample struct {
	CachedInputTokens     int64 `json:"cachedInputTokens,omitempty"`
	InputTokens           int64 `json:"inputTokens,omitempty"`
	OutputTokens          int64 `json:"outputTokens,omitempty"`
	ReasoningOutputTokens int64 `json:"reasoningOutputTokens,omitempty"`
	TotalTokens           int64 `json:"totalTokens,omitempty"`
}

type ThreadTokenUsage struct {
	Last  *ThreadTokenUsageSample `json:"last,omitempty"`
	Total *ThreadTokenUsageSample `json:"total,omitempty"`
}

type AgentMessageDeltaNotification struct {
	Delta    string `json:"delta"`
	ItemID   string `json:"itemId,omitempty"`
	ThreadID string `json:"threadId,omitempty"`
	TurnID   string `json:"turnId,omitempty"`
}

type ItemCompletedNotification struct {
	Item     ThreadItem `json:"item"`
	ThreadID string     `json:"threadId,omitempty"`
	TurnID   string     `json:"turnId,omitempty"`
}

type ThreadTokenUsageUpdatedNotification struct {
	ThreadID   string           `json:"threadId,omitempty"`
	TurnID     string           `json:"turnId,omitempty"`
	TokenUsage ThreadTokenUsage `json:"tokenUsage"`
}

type TurnCompletedNotification struct {
	ThreadID string `json:"threadId,omitempty"`
	Turn     Turn   `json:"turn"`
}

type ThreadInjectItemsResponse struct{}

type ThreadRealtimeTransportType string

const (
	ThreadRealtimeTransportTypeWebsocket ThreadRealtimeTransportType = "websocket"
	ThreadRealtimeTransportTypeWebrtc    ThreadRealtimeTransportType = "webrtc"
)

type ThreadRealtimeStartTransport struct {
	Type ThreadRealtimeTransportType `json:"type"`
	Sdp  *string                     `json:"sdp,omitempty"`
}

func NewThreadRealtimeWebsocketTransport() ThreadRealtimeStartTransport {
	return ThreadRealtimeStartTransport{Type: ThreadRealtimeTransportTypeWebsocket}
}

func NewThreadRealtimeWebrtcTransport(sdp string) ThreadRealtimeStartTransport {
	return ThreadRealtimeStartTransport{
		Type: ThreadRealtimeTransportTypeWebrtc,
		Sdp:  &sdp,
	}
}

func (t ThreadRealtimeStartTransport) MarshalJSON() ([]byte, error) {
	if err := t.validate(); err != nil {
		return nil, err
	}
	type alias ThreadRealtimeStartTransport
	return json.Marshal(alias(t))
}

func (t *ThreadRealtimeStartTransport) UnmarshalJSON(data []byte) error {
	type alias ThreadRealtimeStartTransport
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	transport := ThreadRealtimeStartTransport(decoded)
	if err := transport.validate(); err != nil {
		return err
	}
	*t = transport
	return nil
}

func (t ThreadRealtimeStartTransport) validate() error {
	switch t.Type {
	case ThreadRealtimeTransportTypeWebsocket:
		if t.Sdp != nil {
			return fmt.Errorf("websocket realtime transport cannot include sdp")
		}
		return nil
	case ThreadRealtimeTransportTypeWebrtc:
		if t.Sdp == nil || *t.Sdp == "" {
			return fmt.Errorf("webrtc realtime transport requires sdp")
		}
		return nil
	default:
		return fmt.Errorf("unsupported realtime transport type: %q", t.Type)
	}
}

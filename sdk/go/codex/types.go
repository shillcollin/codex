package codex

import (
	"context"
	"encoding/json"

	"github.com/openai/codex/sdk/go/protocol"
)

type ApprovalHandler func(ctx context.Context, method string, params json.RawMessage) (any, error)

type Config struct {
	CodexBin        string
	ConfigOverrides []string
	Cwd             string
	Env             map[string]string
	ClientName      string
	ClientTitle     string
	ClientVersion   string
	ExperimentalAPI bool
	ApprovalHandler ApprovalHandler
}

type RunOptions struct {
	ApprovalPolicy    *protocol.AskForApproval
	ApprovalsReviewer *protocol.ApprovalsReviewer
	Cwd               *string
	Effort            *protocol.ReasoningEffort
	Model             *string
	OutputSchema      map[string]any
	Personality       *protocol.Personality
	SandboxPolicy     protocol.SandboxPolicy
	ServiceTier       *protocol.ServiceTier
	Summary           *protocol.ReasoningSummary
}

type RunResult struct {
	FinalResponse string
	Items         []protocol.ThreadItem
	Usage         *protocol.ThreadTokenUsage
}

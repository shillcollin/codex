package codex

import (
	"context"
	"encoding/json"

	"github.com/openai/codex/sdk/go/protocol"
)

func DecodeServerRequest(method string, params json.RawMessage) (protocol.ServerRequestPayload, error) {
	return protocol.ServerRequest{Method: method, Params: params}.DecodeKnown()
}

func DefaultServerRequestResponse(method string) any {
	switch method {
	case "account/chatgptAuthTokens/refresh":
		return protocol.ChatgptAuthTokensRefreshResponse{}
	case "applyPatchApproval":
		return protocol.ApplyPatchApprovalResponse{
			Decision: json.RawMessage(`"approved"`),
		}
	case "execCommandApproval":
		return protocol.ExecCommandApprovalResponse{
			Decision: json.RawMessage(`"approved"`),
		}
	case "item/commandExecution/requestApproval":
		return protocol.CommandExecutionRequestApprovalResponse{
			Decision: protocol.CommandExecutionApprovalDecisionAccept,
		}
	case "item/fileChange/requestApproval":
		return protocol.FileChangeRequestApprovalResponse{
			Decision: protocol.FileChangeApprovalDecisionAccept,
		}
	case "item/permissions/requestApproval":
		return protocol.PermissionsRequestApprovalResponse{
			Permissions: protocol.GrantedPermissionProfile{},
			Scope:       protocol.PermissionGrantScopeTurn,
		}
	case "item/tool/call":
		return protocol.DynamicToolCallResponse{
			ContentItems: []json.RawMessage{},
			Success:      false,
		}
	case "item/tool/requestUserInput":
		return protocol.ToolRequestUserInputResponse{
			Answers: map[string]protocol.ToolRequestUserInputAnswer{},
		}
	case "mcpServer/elicitation/request":
		return protocol.McpServerElicitationRequestResponse{
			Action: protocol.McpServerElicitationActionCancel,
		}
	default:
		return map[string]any{}
	}
}

func defaultApprovalHandler(_ context.Context, method string, _ json.RawMessage) (any, error) {
	return DefaultServerRequestResponse(method), nil
}

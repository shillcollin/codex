# Codex Go SDK Reference

This is the reference for the experimental Go SDK for `codex app-server` JSON-RPC v2 over stdio.

Use `README.md` for a quickstart. Use this file when you need the API shape, lifecycle rules, protocol model guidance, or extension points.

## Packages

The SDK exposes two public packages:

- `github.com/openai/codex/sdk/go/codex`
  Ergonomic client API for starting Codex, creating threads, running turns, streaming notifications, and handling server requests.

- `github.com/openai/codex/sdk/go/protocol`
  Wire models for app-server v2. Most definitions are generated from the app-server JSON schema; a small set of union-heavy or convenience types is hand-shaped.

Packages under `internal/` are implementation details and should not be imported by consumers.

## Client Lifecycle

Create a client with `codex.NewClient(ctx, codex.Config{})`.

```go
client, err := codex.NewClient(ctx, codex.Config{})
if err != nil {
    return err
}
defer client.Close()
```

`NewClient` starts a `codex app-server` child process over stdio, sends `initialize`, validates server metadata, and then sends `initialized`.

`Client.Close()` closes the transport, unblocks pending requests, and closes notification subscriptions.

`Client.Metadata()` returns the `protocol.InitializeResponse` received during initialization.

### Config

`codex.Config` fields:

- `CodexBin`: path to a Codex binary. Empty uses the SDK runtime resolver.
- `ConfigOverrides`: config overrides passed to the Codex process.
- `Cwd`: working directory for the Codex process.
- `Env`: additional environment variables for the Codex process.
- `ClientName`, `ClientTitle`, `ClientVersion`: metadata sent in `initialize`.
- `ExperimentalAPI`: sets initialize capability `experimentalApi`.
- `ApprovalHandler`: callback for app-server requests that require a client response.

## Models

Use `Client.Models(ctx, includeHidden)` to call `model/list`.

```go
models, err := client.Models(ctx, false)
```

The response type is `*protocol.ModelListResponse`.

## Threads

Start a thread:

```go
thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{
    Cwd: ptr("/path/to/workspace"),
})
```

Client-level thread helpers:

- `ThreadStart(ctx, protocol.ThreadStartParams) (*codex.Thread, error)`
- `ThreadResume(ctx, threadID string, params protocol.ThreadResumeParams) (*codex.Thread, error)`
- `ThreadFork(ctx, threadID string, params protocol.ThreadForkParams) (*codex.Thread, error)`
- `ThreadList(ctx, params protocol.ThreadListParams) (*protocol.ThreadListResponse, error)`
- `ThreadArchive(ctx, threadID string) error`
- `ThreadUnarchive(ctx, threadID string) (*codex.Thread, error)`

Thread-level helpers:

- `Thread.Read(ctx, includeTurns bool) (*protocol.ThreadReadResponse, error)`
- `Thread.SetName(ctx, name string) error`
- `Thread.Compact(ctx) error`
- `Thread.Notifications(ctx) (<-chan protocol.Notification, <-chan error)`

`Thread` contains the public `ID string` field.

## Turns

For a synchronous run, use `Thread.Run`.

```go
result, err := thread.Run(ctx, "Summarize this repository.", codex.RunOptions{})
if err != nil {
    return err
}
fmt.Println(result.FinalResponse)
```

For streaming or control, start a turn with `Thread.Turn`.

```go
handle, err := thread.Turn(ctx, "Run the tests.", codex.RunOptions{})
if err != nil {
    return err
}

events, errs := handle.Stream(ctx)
for event := range events {
    fmt.Println(event.Method)
}
if err := <-errs; err != nil {
    return err
}
```

Turn helpers:

- `Thread.Run(ctx, input any, opts codex.RunOptions) (*codex.RunResult, error)`
- `Thread.Turn(ctx, input any, opts codex.RunOptions) (*codex.TurnHandle, error)`
- `TurnHandle.Stream(ctx) (<-chan protocol.Notification, <-chan error)`
- `TurnHandle.Run(ctx) (*protocol.Turn, error)`
- `TurnHandle.Steer(ctx, input any) error`
- `TurnHandle.Interrupt(ctx) error`

`TurnHandle` contains public `ThreadID string` and `TurnID string` fields.

`TurnHandle.Stream` stops after the matching `turn/completed` notification.

## Inputs

Turn input accepts:

- `string`
- one `codex.InputItem`
- `[]codex.InputItem`

Input item types:

- `codex.TextInput{Text: "..."}`
- `codex.ImageInput{URL: "https://..."}`
- `codex.LocalImageInput{Path: "/absolute/or/relative/path.png"}`
- `codex.SkillInput{Name: "...", Path: "..."}`
- `codex.MentionInput{Name: "...", Path: "..."}`

Example:

```go
items := []codex.InputItem{
    codex.TextInput{Text: "What is in this image?"},
    codex.LocalImageInput{Path: "./screenshot.png"},
}

result, err := thread.Run(ctx, items, codex.RunOptions{})
```

## Run Options

`codex.RunOptions` maps onto `protocol.TurnStartParams`.

Fields:

- `ApprovalPolicy *protocol.AskForApproval`
- `ApprovalsReviewer *protocol.ApprovalsReviewer`
- `Cwd *string`
- `Effort *protocol.ReasoningEffort`
- `Model *string`
- `OutputSchema map[string]any`
- `Personality *protocol.Personality`
- `SandboxPolicy protocol.SandboxPolicy`
- `ServiceTier *protocol.ServiceTier`
- `Summary *protocol.ReasoningSummary`

`Thread.Run` returns `codex.RunResult`:

- `FinalResponse string`: final assistant message when detectable.
- `Items []protocol.ThreadItem`: completed items observed during the turn.
- `Usage *protocol.ThreadTokenUsage`: latest token usage notification for the turn, if any.

## Notifications

Client-level notifications:

```go
events, errs := client.Notifications(ctx)
```

Thread-level notifications:

```go
events, errs := thread.Notifications(ctx)
```

Notification helpers:

- `Notification.Decode(v any) error`
- `Notification.DecodeKnown() (protocol.NotificationPayload, error)`
- `Notification.IsKnown() bool`
- `Notification.ThreadID() string`
- `Notification.TurnID() string`

Example:

```go
payload, err := event.DecodeKnown()
if err != nil {
    // Unknown notification method or decode failure.
}
switch typed := payload.(type) {
case *protocol.ThreadGoalUpdatedNotification:
    fmt.Println(typed.Goal.Objective)
case *protocol.TurnCompletedNotification:
    fmt.Println(typed.Turn.Status)
}
```

The Go SDK keeps an explicit notification registry in `protocol/notification.go`. Regenerate protocol types after schema changes, then update this registry when new server notification methods are added.

## Server Requests And Approvals

App-server may send JSON-RPC requests to the client. The SDK routes those to `codex.Config.ApprovalHandler`.

```go
client, err := codex.NewClient(ctx, codex.Config{
    ApprovalHandler: func(ctx context.Context, method string, params json.RawMessage) (any, error) {
        payload, err := codex.DecodeServerRequest(method, params)
        if err != nil {
            return nil, err
        }

        switch req := payload.(type) {
        case *protocol.CommandExecutionRequestApprovalParams:
            _ = req
            return protocol.CommandExecutionRequestApprovalResponse{
                Decision: protocol.CommandExecutionApprovalDecisionAccept,
            }, nil
        default:
            return codex.DefaultServerRequestResponse(method), nil
        }
    },
})
```

Helpers:

- `codex.DecodeServerRequest(method, params)`
- `codex.DefaultServerRequestResponse(method)`

Known server requests include:

- `account/chatgptAuthTokens/refresh`
- `applyPatchApproval`
- `attestation/generate`
- `execCommandApproval`
- `item/commandExecution/requestApproval`
- `item/fileChange/requestApproval`
- `item/permissions/requestApproval`
- `item/tool/call`
- `item/tool/requestUserInput`
- `mcpServer/elicitation/request`

The default handler accepts common command and file-change approvals, returns empty answers for `requestUserInput`, cancels MCP elicitation, and returns unsuccessful dynamic tool responses. Production callers should provide an explicit handler for policy-sensitive workflows.

## Browser Use And MCP Elicitation

Browser Use is surfaced through Codex's plugin/MCP paths, not as a dedicated Go package API.

The SDK can participate in Browser Use flows through:

- `item/tool/call` server requests, decoded as `*protocol.DynamicToolCallParams`.
- `mcpServer/elicitation/request`, decoded as `*protocol.McpServerElicitationRequestParams`.
- regular notifications emitted by the app-server while Browser Use tools run.

If you need Browser Use to do real work, install or enable the relevant bundled plugin in the Codex environment and provide an `ApprovalHandler` that makes deliberate decisions for `item/tool/call` and MCP elicitation requests. The SDK default response for `item/tool/call` is `Success: false`.

## Goals

The current upstream app-server supports persisted thread goals behind the `goals` feature and exposes goal notifications.

Generated protocol types include:

- `protocol.ThreadGoal`
- `protocol.ThreadGoalStatus`
- `protocol.ThreadGoalUpdatedNotification`
- `protocol.ThreadGoalClearedNotification`

The notification registry decodes:

- `thread/goal/updated`
- `thread/goal/cleared`

Current SDK gap: the ergonomic `codex.Client` does not yet expose first-class helpers for `thread/goal/set`, `thread/goal/get`, or `thread/goal/clear`, and the generated Go protocol package does not currently include public `ThreadGoalSetParams`, `ThreadGoalGetParams`, or `ThreadGoalClearParams` types. Add those before treating goals as a complete Go SDK feature.

## Errors And Retry

JSON-RPC errors are returned as Go errors. Use `codex.IsRetryableError(err)` to identify retryable app-server overload and transient errors.

See `codex/retry.go` and `examples/error_handling_and_retry` for retry behavior.

## Protocol Generation

Regenerate protocol bindings after app-server protocol changes:

```bash
cd sdk/go
go run ./cmd/generate
```

Generated output:

- `protocol/generated.go`

Hand-maintained protocol files:

- `protocol/types.go`
- `protocol/notification.go`
- `protocol/server_request.go`

After generation, run:

```bash
cd sdk/go
go test ./...
```

Also check schema-method coverage when app-server adds server notification or request methods. The tests in `protocol/notification_test.go`, `protocol/server_request_test.go`, and `protocol/fallbacks_test.go` are designed to catch common drift.

## Compatibility

This SDK targets Codex app-server v2 and is tied to the checked-in schema at:

```text
codex-rs/app-server-protocol/schema/json/codex_app_server_protocol.v2.schemas.json
```

The SDK is experimental. Breaking changes are acceptable while the fork is pre-launch, but the reference should stay honest about generated types, hand-shaped types, and missing ergonomic helpers.

## Testing

Run all Go SDK tests:

```bash
cd sdk/go
go test ./...
```

Live integration tests are gated:

```bash
cd sdk/go
RUN_REAL_CODEX_GO_TESTS=1 go test ./codex
```

Integration tests require a usable Codex runtime and local environment capable of launching `codex app-server`.

## Examples

Examples live under `sdk/go/examples`:

- `quickstart`
- `turn_run`
- `turn_stream_events`
- `thread_lifecycle_and_controls`
- `existing_thread`
- `image_and_text`
- `local_image_and_text`
- `models_and_metadata`
- `turn_params_kitchen_sink`
- `error_handling_and_retry`
- `cli_mini_app`


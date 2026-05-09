# API Reference

## Packages

- `codex`: ergonomic client API
- `protocol`: request, response, notification, and model types

## `codex.Client`

- `NewClient(ctx, cfg)`
- `Close()`
- `Metadata()`
- `Notifications(ctx)`
- `Models(ctx, includeHidden)`
- `ThreadStart(ctx, params)`
- `ThreadResume(ctx, threadID, params)`
- `ThreadFork(ctx, threadID, params)`
- `ThreadList(ctx, params)`
- `ThreadArchive(ctx, threadID)`
- `ThreadUnarchive(ctx, threadID)`

## `codex.Thread`

- `Run(ctx, input, opts)`
- `Turn(ctx, input, opts)`
- `Notifications(ctx)`
- `Read(ctx, includeTurns)`
- `SetName(ctx, name)`
- `Compact(ctx)`

## `codex.TurnHandle`

- `Stream(ctx)`
- `Run(ctx)`
- `Steer(ctx, input)`
- `Interrupt(ctx)`

## Input helpers

- `codex.TextInput`
- `codex.ImageInput`
- `codex.LocalImageInput`
- `codex.SkillInput`
- `codex.MentionInput`

## Notification helpers

- `protocol.Notification.Decode(v)`
- `protocol.Notification.DecodeKnown()`
- `protocol.Notification.IsKnown()`
- `protocol.Notification.ThreadID()`
- `protocol.Notification.TurnID()`

## Server request helpers

- `protocol.ServerRequest.Decode(v)`
- `protocol.ServerRequest.DecodeKnown()`
- `protocol.ServerRequest.IsKnown()`
- `codex.DecodeServerRequest(method, params)`
- `codex.DefaultServerRequestResponse(method)`
- `protocol.RequestPermissionProfile`
- `protocol.AdditionalPermissionProfile`
- `protocol.GrantedPermissionProfile`

## Retry helpers

- `codex.IsRetryableError`
- `codex.RetryOnOverload`

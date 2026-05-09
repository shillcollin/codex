# Getting Started

The Go SDK drives `codex app-server` over stdio, so you need a local `codex` binary on `PATH` or an explicit `codex.Config{CodexBin: ...}` override.

## First program

1. Create a client with `codex.NewClient`.
2. Start or resume a thread.
3. Use `Thread.Run` for the common case or `Thread.Turn` when you need streaming or turn controls.

```go
ctx := context.Background()
client, err := codex.NewClient(ctx, codex.Config{})
if err != nil {
    return err
}
defer client.Close()

thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{})
if err != nil {
    return err
}

result, err := thread.Run(ctx, "Summarize the current repository state.", codex.RunOptions{})
if err != nil {
    return err
}
fmt.Println(result.FinalResponse)
```

## Streaming

Use `Thread.Turn` to get a `TurnHandle`, then read events from `TurnHandle.Stream`.

```go
handle, err := thread.Turn(ctx, "Diagnose the failing test.", codex.RunOptions{})
if err != nil {
    return err
}

events, errs := handle.Stream(ctx)
for event := range events {
    payload, err := event.DecodeKnown()
    if err != nil {
        fmt.Println(event.Method)
        continue
    }
    fmt.Printf("%s: %T\n", event.Method, payload)
}
if err := <-errs; err != nil {
    return err
}
```

Use `Client.Notifications` or `Thread.Notifications` when you want lifecycle events that are not scoped to a single turn, such as `thread/started` or `thread/status/changed`.

For server-initiated approval and user-input requests, use `codex.DecodeServerRequest(...)` or `protocol.ServerRequest.DecodeKnown()` to get a typed payload before deciding how to respond. The common approval flows now expose typed permission/profile structs rather than raw JSON blobs.

# FAQ

## Why does the SDK use `codex app-server` instead of `codex exec`?

The Go SDK targets the richer app-server JSON-RPC v2 surface, which includes thread lifecycle APIs, typed notifications, approval callbacks, and a stable protocol schema.

## Are all protocol types fully hand-written?

No. The SDK keeps the canonical app-server schema in a checked-in generated file and layers a small set of hand-shaped types on top for the highest-value public surface.

## Does the SDK support websocket transport?

Not in the initial release. The first implementation focuses on stdio because it is the default and most stable transport described by the app-server documentation.

## How do I run the live integration tests?

Set `RUN_REAL_CODEX_GO_TESTS=1` before `go test ./...` in `sdk/go/`. Those tests require a working local `codex` runtime and suitable credentials/configuration.

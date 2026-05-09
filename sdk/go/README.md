# Codex Go SDK

Experimental Go SDK for `codex app-server` JSON-RPC v2 over stdio.

The SDK is organized into two public packages:

- `github.com/openai/codex/sdk/go/codex` for the ergonomic client API
- `github.com/openai/codex/sdk/go/protocol` for generated and hand-shaped wire models

For the full API and protocol reference, see [`SDK_REFERENCE.md`](SDK_REFERENCE.md).

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/codex/sdk/go/codex"
	"github.com/openai/codex/sdk/go/protocol"
)

func main() {
	ctx := context.Background()

	client, err := codex.NewClient(ctx, codex.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{})
	if err != nil {
		log.Fatal(err)
	}

	result, err := thread.Run(ctx, "Say hello in one sentence.", codex.RunOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result.FinalResponse)
}
```

## Development

Regenerate the checked-in protocol bindings:

```bash
cd sdk/go
go run ./cmd/generate
```

Run the package tests:

```bash
cd sdk/go
go test ./...
```

Live integration coverage is gated behind `RUN_REAL_CODEX_GO_TESTS=1`.

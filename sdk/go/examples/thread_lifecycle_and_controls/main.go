package main

import (
	"context"
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

	if err := thread.SetName(ctx, "Lifecycle Demo"); err != nil {
		log.Fatal(err)
	}

	handle, err := thread.Turn(ctx, "Start a longer-running task.", codex.RunOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if err := handle.Interrupt(ctx); err != nil {
		log.Fatal(err)
	}

	if err := thread.Compact(ctx); err != nil {
		log.Fatal(err)
	}
}

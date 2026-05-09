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

	resumed, err := client.ThreadResume(ctx, thread.ID, protocol.ThreadResumeParams{})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := resumed.Run(ctx, "Continue the conversation.", codex.RunOptions{}); err != nil {
		log.Fatal(err)
	}
}

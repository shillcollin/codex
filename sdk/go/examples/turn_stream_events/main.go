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

	handle, err := thread.Turn(ctx, "Diagnose the current build.", codex.RunOptions{})
	if err != nil {
		log.Fatal(err)
	}

	events, errs := handle.Stream(ctx)
	for event := range events {
		fmt.Println(event.Method)
	}
	if err := <-errs; err != nil {
		log.Fatal(err)
	}
}

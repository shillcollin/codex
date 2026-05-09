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

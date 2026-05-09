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

	_, err = thread.Run(ctx, []codex.InputItem{
		codex.TextInput{Text: "Describe this image."},
		codex.ImageInput{URL: "https://example.com/ui.png"},
	}, codex.RunOptions{})
	if err != nil {
		log.Fatal(err)
	}
}

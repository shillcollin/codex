package main

import (
	"context"
	"log"

	"github.com/openai/codex/sdk/go/codex"
	"github.com/openai/codex/sdk/go/protocol"
)

func main() {
	ctx := context.Background()
	client, err := codex.NewClient(ctx, codex.Config{ExperimentalAPI: true})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{})
	if err != nil {
		log.Fatal(err)
	}

	model := "gpt-5"
	personality := protocol.PersonalityFriendly
	_, err = thread.Run(ctx, "Summarize this repository as JSON.", codex.RunOptions{
		Model:       &model,
		Personality: &personality,
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"summary": map[string]any{"type": "string"},
			},
			"required":             []string{"summary"},
			"additionalProperties": false,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

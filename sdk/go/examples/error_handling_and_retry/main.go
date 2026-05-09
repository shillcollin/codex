package main

import (
	"context"
	"log"

	"github.com/openai/codex/sdk/go/codex"
)

func main() {
	ctx := context.Background()

	_, err := codex.RetryOnOverload(ctx, 3, func() (string, error) {
		client, err := codex.NewClient(ctx, codex.Config{})
		if err != nil {
			return "", err
		}
		defer client.Close()

		models, err := client.Models(ctx, true)
		if err != nil {
			return "", err
		}
		if len(models.Data) == 0 {
			return "", nil
		}
		return models.Data[0].ID, nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

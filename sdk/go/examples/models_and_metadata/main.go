package main

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/codex/sdk/go/codex"
)

func main() {
	ctx := context.Background()
	client, err := codex.NewClient(ctx, codex.Config{})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	models, err := client.Models(ctx, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(client.Metadata().UserAgent)
	fmt.Printf("models=%d\n", len(models.Data))
}

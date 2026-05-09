package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

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

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		if scanner.Text() == "/exit" {
			return
		}
		result, err := thread.Run(ctx, scanner.Text(), codex.RunOptions{})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result.FinalResponse)
	}
}

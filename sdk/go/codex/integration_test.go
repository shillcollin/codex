package codex

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/openai/codex/sdk/go/protocol"
)

func TestRealInitializeAndModelList(t *testing.T) {
	if os.Getenv("RUN_REAL_CODEX_GO_TESTS") != "1" {
		t.Skip("set RUN_REAL_CODEX_GO_TESTS=1 to run real Codex integration coverage")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, Config{ExperimentalAPI: true})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	models, err := client.Models(ctx, true)
	if err != nil {
		t.Fatalf("Models returned error: %v", err)
	}
	if len(models.Data) == 0 {
		t.Fatal("expected at least one model")
	}
	if client.Metadata().UserAgent == "" {
		t.Fatal("expected initialize metadata")
	}
}

func TestRealThreadRunSmoke(t *testing.T) {
	if os.Getenv("RUN_REAL_CODEX_GO_TESTS") != "1" {
		t.Skip("set RUN_REAL_CODEX_GO_TESTS=1 to run real Codex integration coverage")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, Config{ExperimentalAPI: true})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{})
	if err != nil {
		t.Fatalf("ThreadStart returned error: %v", err)
	}

	result, err := thread.Run(ctx, "Say hello in one short sentence.", RunOptions{})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected at least one completed item")
	}
}

func TestRealThreadNotificationsAndCanonicalTurnRun(t *testing.T) {
	if os.Getenv("RUN_REAL_CODEX_GO_TESTS") != "1" {
		t.Skip("set RUN_REAL_CODEX_GO_TESTS=1 to run real Codex integration coverage")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := NewClient(ctx, Config{ExperimentalAPI: true})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer client.Close()

	events, errs := client.Notifications(ctx)

	thread, err := client.ThreadStart(ctx, protocol.ThreadStartParams{})
	if err != nil {
		t.Fatalf("ThreadStart returned error: %v", err)
	}

	var sawThreadStarted bool
waitForStarted:
	for {
		select {
		case event := <-events:
			if event.Method == "thread/started" && event.ThreadID() == thread.ID {
				sawThreadStarted = true
				break waitForStarted
			}
		case err := <-errs:
			if err != nil {
				t.Fatalf("Notifications returned error: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("timed out waiting for thread/started notification")
		}
	}
	if !sawThreadStarted {
		t.Fatal("expected thread/started notification")
	}

	handle, err := thread.Turn(ctx, "Say hello in one short sentence.", RunOptions{})
	if err != nil {
		t.Fatalf("Turn returned error: %v", err)
	}
	turn, err := handle.Run(ctx)
	if err != nil {
		t.Fatalf("TurnHandle.Run returned error: %v", err)
	}
	if turn.ID == "" {
		t.Fatal("expected canonical completed turn id")
	}
}

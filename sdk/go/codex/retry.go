package codex

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	irpc "github.com/openai/codex/sdk/go/internal/jsonrpc"
)

func IsRetryableError(err error) bool {
	var rpcErr *irpc.Error
	if !errors.As(err, &rpcErr) {
		return false
	}
	return rpcErr.Code == -32001
}

func RetryOnOverload[T any](ctx context.Context, attempts int, fn func() (T, error)) (T, error) {
	var zero T
	delay := 250 * time.Millisecond
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := fn()
		if err == nil {
			return value, nil
		}
		if attempt == attempts || !IsRetryableError(err) {
			return zero, err
		}
		jitter := time.Duration(rand.Int64N(int64(delay / 2)))
		timer := time.NewTimer(delay + jitter)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
		if delay < 2*time.Second {
			delay *= 2
		}
	}
	return zero, context.Canceled
}

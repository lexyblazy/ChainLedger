package rpc

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimiterAcquireBurstThenTimeout(t *testing.T) {
	l := NewRateLimiter(1, 2)

	ctx := context.Background()
	if err := l.Acquire(ctx); err != nil {
		t.Fatalf("first acquire returned error: %v", err)
	}
	if err := l.Acquire(ctx); err != nil {
		t.Fatalf("second acquire returned error: %v", err)
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := l.Acquire(timeoutCtx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("third acquire error = %v, want deadline exceeded", err)
	}
}

func TestRateLimiterAcquireCancelledContext(t *testing.T) {
	l := NewRateLimiter(1, 0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := l.Acquire(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Acquire error = %v, want context canceled", err)
	}
}

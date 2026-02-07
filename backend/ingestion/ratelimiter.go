package ingestion

import (
	"context"
	"time"
)

type RPCLimiter struct {
	tokens chan struct{}
}

func NewRPCLimiter(rps int, burst int) *RPCLimiter {
	l := &RPCLimiter{tokens: make(chan struct{}, burst)}

	// fill the channel with the burst size
	for range burst {
		l.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(rps)

	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			select {
			case l.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return l
}

func (l *RPCLimiter) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}

package rpc

import (
	"context"
	"time"
)

type RateLimiter struct {
	tokens chan struct{}
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	l := &RateLimiter{tokens: make(chan struct{}, burst)}

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

func (l *RateLimiter) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}

package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lexyblazy/chainledger/config"
)

func TestClientCallSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":{"ok":true},"id":1}`))
	}))
	defer ts.Close()

	c := &Client{url: ts.URL, client: ts.Client()}

	result, err := c.Call(context.Background(), "eth_blockNumber", []any{})
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}

	var payload map[string]bool
	if err := json.Unmarshal(result, &payload); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if !payload["ok"] {
		t.Fatalf("unexpected result payload: %+v", payload)
	}
}

func TestClientCallHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := &Client{url: ts.URL, client: ts.Client()}

	if _, err := c.Call(context.Background(), "x", nil); err == nil {
		t.Fatal("expected non-200 response to return error")
	}
}

func TestClientCallRPCError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","error":{"code":-32000,"message":"boom"},"id":1}`))
	}))
	defer ts.Close()

	c := &Client{url: ts.URL, client: ts.Client()}

	if _, err := c.Call(context.Background(), "x", nil); err == nil {
		t.Fatal("expected rpc error field to return error")
	}
}

func TestCallRpcWithRetryEventuallySucceeds(t *testing.T) {
	c := &Client{
		limiter: NewRateLimiter(1000, 1),
		config: &config.RPCRateLimitConfig{MaxRetryCount: 3, RetryDelayMs: 1},
	}

	var attempts int32
	rpcCall := func() (json.RawMessage, error) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			return nil, errors.New("temporary")
		}
		return json.RawMessage(`"ok"`), nil
	}

	result, err := c.CallRpcWithRetry(context.Background(), rpcCall)
	if err != nil {
		t.Fatalf("CallRpcWithRetry returned error: %v", err)
	}
	if string(result) != `"ok"` {
		t.Fatalf("result = %s, want \"ok\"", string(result))
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestCallRpcWithRetryHonorsContext(t *testing.T) {
	c := &Client{
		limiter: NewRateLimiter(1, 0),
		config: &config.RPCRateLimitConfig{MaxRetryCount: 1, RetryDelayMs: 1},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := c.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return json.RawMessage(`"unused"`), nil
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
}

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"chainledger/config"
)

type Client struct {
	url     string
	client  *http.Client
	limiter *RateLimiter
	config  *config.RPCRateLimitConfig
}

func New(networkConfig *config.NetworkConfig) *Client {
	limiter := NewRateLimiter(networkConfig.RPCRateLimit.RPS, networkConfig.RPCRateLimit.Burst)

	return &Client{
		url: networkConfig.RPCUrl,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		limiter: limiter,
		config:  &networkConfig.RPCRateLimit,
	}
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      int             `json:"id"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (c *Client) Call(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {

	reqBody := rpcRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rpc http status %d", resp.StatusCode)
	}

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func (c *Client) CallRpcWithRetry(ctx context.Context, rpcCall func() (json.RawMessage, error)) (json.RawMessage, error) {

	var lastError error
	var result json.RawMessage

	delay := time.Duration(c.config.RetryDelayMs) * time.Millisecond

	for i := 0; i < c.config.MaxRetryCount; i++ {

		for {
			err := c.limiter.Acquire(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}

				// Otherwise wait and retry token acquisition.
				time.Sleep(time.Duration(50 * time.Millisecond))
				continue
			}

			break
		}

		result, lastError = rpcCall()

		// successful call, break the loop
		if lastError == nil {
			break
		}

		if i == c.config.MaxRetryCount-1 {
			break
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}

	}

	return result, lastError
}

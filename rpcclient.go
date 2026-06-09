package hivego

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// rpcRequest, rpcResponse and rpcError are the JSON-RPC 2.0 envelope types.
// They replace the former github.com/cfoxon/jsonrpc2client dependency, which was
// dropped because it (a) routed every call through fasthttp's global client and
// so offered no way to cancel an in-flight request — see callRaw/callBatch and
// HG-H7 — (b) swallowed transport errors in its batch path, and (c) carried a
// racy, unused worker-pool batch fan-out. net/http with a per-request context
// gives real cancellation, transparent gzip, and connection reuse for free.
type rpcRequest struct {
	JsonRpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type rpcResponse struct {
	JsonRpc string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	Id      int             `json:"id"`
}

// rpcHTTPClient is shared so connections are pooled and reused across calls (the
// old per-call client construction reused nothing). We never set Accept-Encoding
// ourselves: net/http negotiates gzip and transparently decompresses the body.
// No Client.Timeout is set — each call bounds itself with a context deadline
// (h.rpcTimeout()), which also covers DNS, dial and TLS, not just the response.
var rpcHTTPClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     90 * time.Second,
	},
}

// postJSON POSTs body to endpoint bounded by ctx and returns the raw response
// body. Cancelling ctx aborts the in-flight request and frees the goroutine —
// the real HG-H7 fix, which the previous callWithTimeout wrapper could not do
// (it only stopped the caller waiting while the request leaked in the
// background).
func postJSON(ctx context.Context, endpoint string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rpcHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("http %d from %s: %s", resp.StatusCode, endpoint, string(respBody))
	}
	return respBody, nil
}

// callRaw performs a single JSON-RPC call against endpoint and decodes the
// response envelope. The call is bounded by h.rpcTimeout().
func (h *HiveRpcNode) callRaw(endpoint string, req *rpcRequest) (*rpcResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.rpcTimeout())
	defer cancel()

	respBody, err := postJSON(ctx, endpoint, body)
	if err != nil {
		return nil, err
	}

	var out rpcResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("undecodable response from %s: %w", endpoint, err)
	}
	return &out, nil
}

// callBatch sends reqs as a single JSON-RPC batch (a JSON array) against
// endpoint and returns the raw array body for the caller to parse. Unlike the
// former library it issues exactly one HTTP request and never drops or reorders
// sub-responses. The call is bounded by h.rpcTimeout().
func (h *HiveRpcNode) callBatch(endpoint string, reqs []*rpcRequest) ([]byte, error) {
	body, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.rpcTimeout())
	defer cancel()

	return postJSON(ctx, endpoint, body)
}

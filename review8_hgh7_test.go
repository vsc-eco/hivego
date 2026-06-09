package hivego

// review8 HG-H7 — the RPC client had no read/write timeout, so a node that
// accepts the connection but never replies blocked the calling (broadcast)
// goroutine forever. Each call now wraps its HTTP request in a context deadline
// (h.rpcTimeout()), so a hung node is abandoned and the request is cancelled.

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReview8_HGH7_CallTimesOutOnHungNode(t *testing.T) {
	// A server that accepts the connection but never replies — the HG-H7 scenario.
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-release // hang until the test releases it
	}))
	// LIFO defers: release the handler first so srv.Close() doesn't wait on it.
	defer srv.Close()
	defer close(release)

	node := NewHiveRpcWithOpts([]string{srv.URL}, 1, 1)
	node.RpcTimeout = 100 * time.Millisecond

	start := time.Now()
	_, err := node.callRaw(srv.URL, &rpcRequest{Method: "x", JsonRpc: "2.0", Id: 1})
	if err == nil {
		t.Fatal("HG-H7: a hung node must produce a timeout error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("HG-H7: caller should unblock near the timeout (~100ms), waited %s", elapsed)
	}
}

func TestReview8_HGH7_CallRawSucceeds(t *testing.T) {
	// A prompt, well-formed response passes through untouched.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"ok":true}}`))
	}))
	defer srv.Close()

	node := NewHiveRpcWithOpts([]string{srv.URL}, 1, 1)
	resp, err := node.callRaw(srv.URL, &rpcRequest{Method: "x", JsonRpc: "2.0", Id: 1})
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected rpc error: %v", resp.Error)
	}
	if len(resp.Result) == 0 {
		t.Fatal("expected a non-empty result")
	}
}

func TestReview8_HGH7_DefaultTimeoutApplied(t *testing.T) {
	node := NewHiveRpcWithOpts([]string{"https://api.hive.blog"}, 1, 1)
	if node.rpcTimeout() <= 0 {
		t.Fatal("constructed node must have a positive rpc timeout")
	}
	bare := &HiveRpcNode{}
	if bare.rpcTimeout() != defaultRpcTimeout {
		t.Fatalf("zero-value node must fall back to defaultRpcTimeout, got %s", bare.rpcTimeout())
	}
}

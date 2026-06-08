package hivego

// review8 HG-H7 — the RPC client had no read/write timeout, so a node that
// accepts the connection but never replies blocked the calling (broadcast)
// goroutine forever. callWithTimeout bounds each attempt.

import (
	"errors"
	"testing"
	"time"
)

func TestReview8_HGH7_CallWithTimeoutUnblocksOnHang(t *testing.T) {
	// A call that returns promptly passes its value through untouched.
	v, err := callWithTimeout(time.Second, func() (int, error) { return 42, nil })
	if err != nil || v != 42 {
		t.Fatalf("fast call should pass through, got v=%d err=%v", v, err)
	}

	// An underlying error is propagated.
	sentinel := errors.New("boom")
	if _, err := callWithTimeout(time.Second, func() (int, error) { return 0, sentinel }); !errors.Is(err, sentinel) {
		t.Fatalf("error should propagate, got %v", err)
	}

	// A hung call (never returns within the timeout) returns a timeout error
	// rather than blocking the caller forever — the core HG-H7 fix.
	start := time.Now()
	_, err = callWithTimeout(50*time.Millisecond, func() (int, error) {
		time.Sleep(5 * time.Second) // simulate a node that accepts but never replies
		return 1, nil
	})
	if err == nil {
		t.Fatal("HG-H7: a hung call must time out")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("HG-H7: caller should unblock at the timeout (~50ms), waited %s", elapsed)
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

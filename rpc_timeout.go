package hivego

import (
	"fmt"
	"time"
)

// defaultRpcTimeout bounds a single node RPC round-trip. HG-H7: the underlying
// jsonrpc2client/fasthttp client is created without any read/write deadline, so
// a node that accepts the TCP connection but never replies blocks the calling
// goroutine forever — and since broadcast runs on that goroutine, one hung Hive
// node stalls broadcasting indefinitely with no failover. We bound each attempt
// here so the loop can record the failure and rotate to the next node.
const defaultRpcTimeout = 30 * time.Second

// rpcTimeout returns the configured per-call timeout, or the default when unset.
func (h *HiveRpcNode) rpcTimeout() time.Duration {
	if h.RpcTimeout > 0 {
		return h.RpcTimeout
	}
	return defaultRpcTimeout
}

// callWithTimeout runs fn and returns its result, or a timeout error if it does
// not complete within timeout. The result channel is buffered so the worker
// goroutine never blocks delivering a late result (it exits when fn finally
// returns); this unblocks the CALLER on a hung node, which is the HG-H7 fix.
func callWithTimeout[T any](timeout time.Duration, fn func() (T, error)) (T, error) {
	type outcome struct {
		val T
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		val, err := fn()
		ch <- outcome{val, err}
	}()
	select {
	case o := <-ch:
		return o.val, o.err
	case <-time.After(timeout):
		var zero T
		return zero, fmt.Errorf("rpc call timed out after %s", timeout)
	}
}

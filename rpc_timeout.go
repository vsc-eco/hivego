package hivego

import "time"

// defaultRpcTimeout bounds a single node RPC round-trip (HG-H7). Each call wraps
// its HTTP request in a context with this deadline (see callRaw/callBatch), so a
// node that accepts the connection but never replies no longer blocks the calling
// goroutine forever — the context is cancelled, the request is aborted, and the
// failover loop rotates to the next node.
const defaultRpcTimeout = 30 * time.Second

// rpcTimeout returns the configured per-call timeout, or the default when unset
// (so struct-literal construction is still protected).
func (h *HiveRpcNode) rpcTimeout() time.Duration {
	if h.RpcTimeout > 0 {
		return h.RpcTimeout
	}
	return defaultRpcTimeout
}

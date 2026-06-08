package hivego

// review8 HG-H3 — rpcExecBatchFast ignored per-response JSON-RPC errors (it only
// checked len==0), so a node-side failure inside a batch (rejected broadcast,
// duplicate/already-exists tx) was reported as success. batchResponseError now
// surfaces those.

import (
	"bytes"
	"testing"
)

func TestReview8_HGH3_BatchResponseErrorDetected(t *testing.T) {
	// A successful batch body (array, no error members) is accepted.
	ok := []byte(`[{"id":0,"result":{"block":1}},{"id":1,"result":"ok"}]`)
	if err := batchResponseError(ok); err != nil {
		t.Fatalf("clean batch should pass, got %v", err)
	}

	// A batch where one sub-response carries a JSON-RPC error must be rejected.
	// Pre-fix this body was non-empty, so rpcExecBatchFast reported success.
	withErr := []byte(`[{"id":0,"result":"ok"},{"id":1,"error":{"code":-32000,"message":"Assert Exception:false: Transaction already exists"}}]`)
	err := batchResponseError(withErr)
	if err == nil {
		t.Fatal("HG-H3: a per-response error in the batch must be surfaced, got nil")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("already exists")) {
		t.Fatalf("error should carry the node message, got %v", err)
	}

	// A single (non-array) error object is also handled.
	single := []byte(`{"id":1,"error":{"code":-32603,"message":"Internal error"}}`)
	if err := batchResponseError(single); err == nil {
		t.Fatal("HG-H3: single-object error response must be surfaced")
	}

	// Empty body is an error (mirrors the prior len==0 check).
	if err := batchResponseError([]byte("  ")); err == nil {
		t.Fatal("empty body must be an error")
	}
}

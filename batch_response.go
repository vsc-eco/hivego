package hivego

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// batchResponseError inspects a raw JSON-RPC batch HTTP body and returns a
// non-nil error if any sub-response carries a JSON-RPC `error` member. HG-H3:
// rpcExecBatchFast previously treated any non-empty body as success, so a
// per-request failure inside the batch (e.g. a broadcast that the node rejected,
// or an already-exists/duplicate-tx error) was silently reported as a successful
// call — the single-call path (rpcExec) already rejects resp.Error, the batch
// path did not. Each body is a JSON-RPC batch response, normally a JSON array of
// response objects, but a single object is tolerated for robustness.
func batchResponseError(body []byte) error {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return errors.New("empty response body")
	}

	type jr2Error struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	type jr2Response struct {
		Error *jr2Error       `json:"error"`
		ID    json.RawMessage `json:"id"`
	}

	var responses []jr2Response
	if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &responses); err != nil {
			return fmt.Errorf("undecodable batch response: %w", err)
		}
	} else {
		var single jr2Response
		if err := json.Unmarshal(trimmed, &single); err != nil {
			return fmt.Errorf("undecodable batch response: %w", err)
		}
		responses = append(responses, single)
	}

	for _, r := range responses {
		if r.Error != nil {
			return fmt.Errorf("json-rpc error in batch response (id %s): code=%d %s",
				string(r.ID), r.Error.Code, r.Error.Message)
		}
	}
	return nil
}

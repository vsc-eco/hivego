package hivego

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v2"
)

type HiveTransaction struct {
	RefBlockNum    uint16           `json:"ref_block_num"`
	RefBlockPrefix uint32           `json:"ref_block_prefix"`
	Expiration     string           `json:"expiration"`
	Operations     []HiveOperation  `json:"-"`
	OperationsJs   [][2]interface{} `json:"operations"`
	Extensions     []string         `json:"extensions"`
	Signatures     []string         `json:"signatures"`
}

func (t *HiveTransaction) GenerateTrxId() (string, error) {
	tB, err := SerializeTx(*t)
	if err != nil {
		return "", err
	}
	digest := HashTx(tB)

	return hex.EncodeToString(digest)[0:40], nil
}

// ValidateExpiration reports an error if the transaction's Expiration is
// malformed or not strictly in the future relative to `now` (compared in UTC,
// the timezone Hive uses for the expiration field).
//
// review7 HG-M16: Sign() is a pure signing primitive and is deliberately left
// time-independent — a tx must be re-signable/reproducible (fixed-vector tests
// sign a tx dated 2016), so Sign() cannot reject an expired tx without breaking
// deterministic signing. Callers that build a tx for broadcast should call
// ValidateExpiration(time.Now().UTC()) first, so they don't sign and broadcast
// a tx the node will only reject as expired.
func (t *HiveTransaction) ValidateExpiration(now time.Time) error {
	exp, err := time.Parse("2006-01-02T15:04:05", t.Expiration)
	if err != nil {
		return fmt.Errorf("invalid expiration %q: %w", t.Expiration, err)
	}
	if !exp.After(now.UTC()) {
		return fmt.Errorf("transaction expiration %s is not in the future (now %s)",
			t.Expiration, now.UTC().Format("2006-01-02T15:04:05"))
	}
	return nil
}

func (t *HiveTransaction) Sign(keyPair KeyPair, chainId ...string) (string, error) {
	message, err := SerializeTx(*t)

	if err != nil {
		return "", err
	}

	digest, err := HashTxForSig(message, chainId...)
	if err != nil {
		return "", err
	}
	sig, err := secp256k1.SignCompact(keyPair.PrivateKey, digest, true)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

func (t *HiveTransaction) AddSig(sig string) {
	t.Signatures = append(t.Signatures, sig)
}

func (t *HiveTransaction) prepareJson() {
	var opsContainer [][2]interface{}
	for _, op := range t.Operations {
		var opContainer [2]interface{}
		opContainer[0] = op.OpName()
		opContainer[1] = op
		opsContainer = append(opsContainer, opContainer)
	}
	if t.Extensions == nil {
		t.Extensions = []string{}
	}
	t.OperationsJs = opsContainer
}

func (h *HiveRpcNode) Broadcast(ops []HiveOperation, wif *string) (string, error) {
	signingData, err := h.GetSigningData()
	if err != nil {
		return "", err
	}
	tx := HiveTransaction{
		RefBlockNum:    signingData.refBlockNum,
		RefBlockPrefix: signingData.refBlockPrefix,
		Expiration:     signingData.expiration,
		Operations:     ops,
	}

	message, err := SerializeTx(tx)

	if err != nil {
		return "", err
	}

	digest, err := HashTxForSig(message, h.ChainID)
	if err != nil {
		return "", err
	}

	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}
	sig, err := SignDigest(digest, wif)
	if err != nil {
		return "", err
	}

	tx.Signatures = append(tx.Signatures, hex.EncodeToString(sig))

	tx.prepareJson()

	var params []interface{}
	params = append(params, tx)
	if !h.NoBroadcast {
		q := hrpcQuery{"condenser_api.broadcast_transaction", params}
		res, err := h.rpcExec(q)
		if err != nil {
			return string(res), err
		}
	}

	return txId, nil
}

func (h *HiveRpcNode) BroadcastRaw(tx HiveTransaction) (string, error) {
	if len(tx.Signatures) == 0 {
		return "", fmt.Errorf("transaction is not signed")
	}

	tx.prepareJson()
	var params []interface{}
	params = append(params, tx)
	if !h.NoBroadcast {
		q := hrpcQuery{"condenser_api.broadcast_transaction", params}
		res, err := h.rpcExec(q)
		if err != nil {
			return string(res), err
		}
	}
	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}
	return txId, nil
}

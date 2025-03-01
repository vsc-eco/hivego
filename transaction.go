package hivego

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v2"
	"github.com/vsc-eco/hivego/utils"
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

func (t *HiveTransaction) Sign(keyPair KeyPair) (string, error) {
	message, err := SerializeTx(*t)

	if err != nil {
		return "", err
	}

	digest := HashTxForSig(message)

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

type TransactionQueryParams struct {
	TransactionId     string `json:"id"`
	IncludeReversible bool   `json:"include_reversible"`
}

func (h *HiveRpcNode) GetTransaction(txId string, includeReversible bool) (HiveTransaction, error) {
	var ht HiveTransaction
	q := hrpcQuery{method: CondenserApiGetTransaction, params: TransactionQueryParams{TransactionId: txId, IncludeReversible: includeReversible}}
	res, err := h.CallRaw(q)
	if err != nil {
		return ht, err
	}
	err = utils.Recast(res.Result, &ht)
	if err != nil {
		return ht, err
	}
	return ht, nil
}

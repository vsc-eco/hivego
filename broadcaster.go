package hivego

import (
	"encoding/hex"
	"fmt"
)

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

	digest := HashTxForSig(message)

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
		q := hrpcQuery{CondenserApiBroadcastTransaction, params}
		res, err := h.CallRaw(q)
		if err != nil {
			return res.Error.Message, err
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
		q := hrpcQuery{CondenserApiBroadcastTransaction, params}
		res, err := h.CallRaw(q)
		if err != nil {
			return res.Error.Message, err
		}
	}
	txId, err := tx.GenerateTrxId()
	if err != nil {
		return "", err
	}
	return txId, nil
}

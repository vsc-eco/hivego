package hivego

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/decred/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v2"
)

type signingDataFromChain struct {
	refBlockNum    uint16
	refBlockPrefix uint32
	expiration     string
}

func (h *HiveRpcNode) GetSigningData() (signingDataFromChain, error) {
	propsB, err := h.GetDynamicGlobalProps()
	if err != nil {
		return signingDataFromChain{}, err
	}

	var props globalProps
	err = json.Unmarshal(propsB, &props)
	if err != nil {
		return signingDataFromChain{}, err
	}

	refBlockNum := uint16(props.HeadBlockNumber & 0xffff)
	hbidB, err := hex.DecodeString(props.HeadBlockId)
	if err != nil {
		return signingDataFromChain{}, err
	}
	refBlockPrefix := binary.LittleEndian.Uint32(hbidB[4:])

	exp, err := time.Parse("2006-01-02T15:04:05", props.Time)
	if err != nil {
		return signingDataFromChain{}, err
	}
	exp = exp.Add(30 * time.Second)
	expStr := exp.Format("2006-01-02T15:04:05")

	signingData := signingDataFromChain{refBlockNum, refBlockPrefix, expStr}

	return signingData, nil
}

func HashTxForSig(tx []byte, chainID ...string) ([]byte, error) {
	var message bytes.Buffer

	// Use custom chain ID if provided, otherwise use default
	if len(chainID) > 0 && chainID[0] != "" {
		cid, err := hex.DecodeString(chainID[0])
		if err != nil {
			// HG-H6: never silently drop an invalid chain id. The old code
			// discarded this error, leaving cid empty, so the signing digest was
			// sha256(tx) with NO chain prefix — identical to the digest for a
			// different/empty chain. That collision makes a signature produced for
			// one chain replayable on another. Fail closed instead.
			return nil, fmt.Errorf("invalid chain id %q: %w", chainID[0], err)
		}
		message.Write(cid)
	} else {
		message.Write(getHiveChainId())
	}

	message.Write(tx)

	digest := sha256.New()
	digest.Write(message.Bytes())
	return digest.Sum(nil), nil
}

func HashTx(tx []byte) []byte {
	var message bytes.Buffer
	message.Write(tx)

	digest := sha256.New()
	digest.Write(message.Bytes())
	return digest.Sum(nil)
}

func SignDigest(digest []byte, wif *string) ([]byte, error) {
	keyPair, err := KeyPairFromWif(*wif)

	if err != nil {
		return nil, err
	}

	return secp256k1.SignCompact(keyPair.PrivateKey, digest, true)
}

func GphBase58CheckDecode(input string) ([]byte, [1]byte, error) {
	decoded := base58.Decode(input)
	if len(decoded) < 6 {
		return nil, [1]byte{0}, errors.New("invalid format: version and/or checksum bytes missing")
	}
	version := [1]byte{decoded[0]}
	dataLen := len(decoded) - 4
	decodedChecksum := decoded[dataLen:]
	calculatedChecksum := checksum(decoded[:dataLen])
	if !bytes.Equal(decodedChecksum, calculatedChecksum[:]) {
		return nil, [1]byte{0}, errors.New("checksum error")
	}
	payload := decoded[1:dataLen]
	return payload, version, nil
}

func GphBase58Encode(input []byte, version [1]byte) string {
	// review7 HG-M9: the version byte must be part of the encoded output, not
	// just the checksum pre-image — GphBase58CheckDecode reads decoded[0] as the
	// version and re-derives the checksum over (version || payload). The old
	// form emitted base58(input || checksum), dropping the version prefix, so a
	// GphBase58CheckDecode round-trip mis-read the first payload byte as the
	// version and failed the checksum.
	payload := append([]byte{version[0]}, input...)
	checksum := checksum(payload)
	encoded := append(payload, checksum[:4]...)
	return base58.Encode(encoded)
}

func checksum(input []byte) [4]byte {
	var calculatedChecksum [4]byte
	intermediateHash := sha256.Sum256(input)
	finalHash := sha256.Sum256(intermediateHash[:])
	copy(calculatedChecksum[:], finalHash[:])
	return calculatedChecksum
}

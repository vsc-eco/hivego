package hivego

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"time"

	"github.com/decred/base58"
	"github.com/decred/dcrd/dcrec/secp256k1/v2"
	"github.com/vsc-eco/hivego/utils"
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
	//err = json.Unmarshal(propsB, &props)
	err = utils.Recast(propsB.Result, &props)
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

func HashTxForSig(tx []byte) []byte {
	var message bytes.Buffer
	message.Write(getHiveChainId())
	message.Write(tx)

	digest := sha256.New()
	digest.Write(message.Bytes())
	return digest.Sum(nil)
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
	checksum := checksum(append([]byte{version[0]}, input...))
	encoded := append(input, checksum[:4]...)
	return base58.Encode(encoded)
}

func checksum(input []byte) [4]byte {
	var calculatedChecksum [4]byte
	intermediateHash := sha256.Sum256(input)
	finalHash := sha256.Sum256(intermediateHash[:])
	copy(calculatedChecksum[:], finalHash[:])
	return calculatedChecksum
}

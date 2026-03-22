package hivego

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

func opIdB(opName string) byte {
	id := getHiveOpId(opName)
	return byte(id)
}

func refBlockNumB(refBlockNumber uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, refBlockNumber)
	return buf
}

func refBlockPrefixB(refBlockPrefix uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, refBlockPrefix)
	return buf
}

func expTimeB(expTime string) ([]byte, error) {

	exp, err := time.Parse("2006-01-02T15:04:05", expTime)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(exp.Unix()))
	return buf, nil
}

func countOpsB(ops []HiveOperation) []byte {
	b := make([]byte, 5)
	l := binary.PutUvarint(b, uint64(len(ops)))
	return b[0:l]
}

func extensionsB() byte {
	return byte(0x00)
}

func appendVString(s string, b *bytes.Buffer) *bytes.Buffer {
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(s)))
	b.Write(vBuf[0:vLen])

	b.WriteString(s)
	return b
}

func appendVStringArray(a []string, b *bytes.Buffer) *bytes.Buffer {
	vBuf := make([]byte, 5)
	vLen := binary.PutUvarint(vBuf, uint64(len(a)))
	b.Write(vBuf[0:vLen])
	for _, s := range a {
		appendVString(s, b)
	}
	return b
}

func appendVAsset(asset string, b *bytes.Buffer) error {
	parts := strings.Split(asset, " ")
	if len(parts) != 2 {
		return errors.New("invalid asset format: " + asset)
	}

	amountStr, symbol := parts[0], parts[1]
	precision := 3
	if symbol == "VESTS" {
		precision = 6
	}

	var nai uint32
	switch symbol {
	case "HIVE":
		fallthrough
	case "TESTS":
		nai = ((99999999 + 2) << 5) | 3
	case "HBD":
		fallthrough
	case "TBD":
		nai = ((99999999 + 1) << 5) | 3
	case "VESTS":
		nai = ((99999999 + 3) << 5) | 6
	default:
		return errors.New("invalid asset symbol")
	}

	// Handle decimal parsing without floating points
	parts = strings.Split(amountStr, ".")
	if len(parts) > 2 {
		return errors.New("invalid amount format: " + amountStr)
	}

	// Pad or truncate decimal part to required precision
	decimalPart := ""
	if len(parts) > 1 {
		decimalPart = parts[1]
	}
	if len(decimalPart) > precision {
		decimalPart = decimalPart[:precision]
	} else {
		decimalPart += strings.Repeat("0", precision-len(decimalPart))
	}

	// Combine whole and decimal parts
	fullNumber := parts[0] + decimalPart
	amount, err := strconv.ParseInt(fullNumber, 10, 64)
	if err != nil {
		return err
	}

	if amount < 0 {
		return errors.New("negative amount not allowed")
	}

	// Write the amount as int64
	err = binary.Write(b, binary.LittleEndian, amount)
	if err != nil {
		return err
	}

	// Write the nai
	naiBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(naiBytes, nai)
	b.Write(naiBytes)

	return nil
}

func SerializeTx(tx HiveTransaction) ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(refBlockNumB(tx.RefBlockNum))
	buf.Write(refBlockPrefixB(tx.RefBlockPrefix))
	expTime, err := expTimeB(tx.Expiration)
	if err != nil {
		return nil, err
	}
	buf.Write(expTime)

	opsB, err := serializeOps(tx.Operations)
	if err != nil {
		return nil, err
	}
	buf.Write(opsB)
	buf.Write([]byte{extensionsB()})
	return buf.Bytes(), nil
}

func serializeOps(ops []HiveOperation) ([]byte, error) {
	var opsBuf bytes.Buffer
	opsBuf.Write(countOpsB(ops))
	for _, op := range ops {
		b, err := op.SerializeOp()
		if err != nil {
			return nil, err
		}
		opsBuf.Write(b)
	}
	return opsBuf.Bytes(), nil
}

func (o voteOperation) SerializeOp() ([]byte, error) {
	var voteBuf bytes.Buffer
	voteBuf.Write([]byte{opIdB(o.OpName())})
	appendVString(o.Voter, &voteBuf)
	appendVString(o.Author, &voteBuf)
	appendVString(o.Permlink, &voteBuf)

	weightBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(weightBuf, uint16(o.Weight))
	voteBuf.Write(weightBuf)

	return voteBuf.Bytes(), nil
}

func (o CustomJsonOperation) SerializeOp() ([]byte, error) {
	var jBuf bytes.Buffer
	jBuf.Write([]byte{opIdB(o.OpName())})
	appendVStringArray(o.RequiredAuths, &jBuf)
	appendVStringArray(o.RequiredPostingAuths, &jBuf)
	appendVString(o.Id, &jBuf)
	appendVString(o.Json, &jBuf)

	return jBuf.Bytes(), nil
}

func (o ClaimRewardOperation) SerializeOp() ([]byte, error) {
	var claimBuf bytes.Buffer
	claimBuf.Write([]byte{opIdB(o.OpName())})
	appendVString(o.Account, &claimBuf)
	err := appendVAsset(o.RewardHIVE, &claimBuf)

	if err != nil {
		return nil, err
	}

	err = appendVAsset(o.RewardHBD, &claimBuf)

	if err != nil {
		return nil, err
	}

	err = appendVAsset(o.RewardVests, &claimBuf)

	if err != nil {
		return nil, err
	}

	return claimBuf.Bytes(), nil
}

func (o TransferOperation) SerializeOp() ([]byte, error) {
	var transferBuf bytes.Buffer
	transferBuf.Write([]byte{opIdB(o.OpName())})
	appendVString(o.From, &transferBuf)
	appendVString(o.To, &transferBuf)
	if err := appendVAsset(o.Amount, &transferBuf); err != nil {
		return nil, err
	}
	appendVString(o.Memo, &transferBuf)

	return transferBuf.Bytes(), nil
}

func (a AccountCreateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(a.OpName()))

	// fee
	err := appendVAsset(a.Fee, &buf)
	if err != nil {
		return nil, err
	}

	appendVString(a.Creator, &buf)
	appendVString(a.NewAccountName, &buf)
	if err = serializeAuthority(a.Owner, &buf); err != nil {
		return nil, err
	}
	if err = serializeAuthority(a.Active, &buf); err != nil {
		return nil, err
	}
	if err = serializeAuthority(a.Posting, &buf); err != nil {
		return nil, err
	}

	err = writePublicKey(a.MemoKey, &buf)
	if err != nil {
		return nil, err
	}
	appendVString(a.JsonMetadata, &buf)
	// extensions (empty) — required by Hive protocol, same as CreateClaimedAccountOperation
	buf.WriteByte(extensionsB())

	return buf.Bytes(), nil
}

func (a AccountUpdateOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer

	// operation ID
	buf.WriteByte(opIdB(a.OpName()))

	// account name
	appendVString(a.Account, &buf)

	// serialize optional authorities (owner, active, posting)
	// TODO: THIS IS UNTESTED
	if err := appendOptionalAuthority(a.Owner, &buf); err != nil {
		return nil, err
	}
	if err := appendOptionalAuthority(a.Active, &buf); err != nil {
		return nil, err
	}
	if err := appendOptionalAuthority(a.Posting, &buf); err != nil {
		return nil, err
	}

	// memo key
	//
	// The memo key is kept as a string argument for the sake of simplicity and
	// because it's intuative to the user. However, it must be serialized as a
	// public key. We decode the public key and compressed to 33 bytes to actually
	// be used.
	pubKey, err := DecodePublicKey(a.MemoKey)
	if err != nil {
		return nil, err
	}
	buf.Write(pubKey.SerializeCompressed())

	// JSON metadata
	appendVString(a.JsonMetadata, &buf)

	return buf.Bytes(), nil
}

func (o TransferToSavings) SerializeOp() ([]byte, error) {
	// OperationSerializers.transfer_to_savings = OperationDataSerializer(32, [
	// 	['from', StringSerializer],
	// 	['to', StringSerializer],
	// 	['amount', AssetSerializer],
	// 	['memo', StringSerializer]
	//   ])

	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	if err := appendVAsset(o.Amount, &buf); err != nil {
		return nil, err
	}
	appendVString(o.Memo, &buf)

	return buf.Bytes(), nil
}

func (o TransferFromSavings) SerializeOp() ([]byte, error) {
	// OperationSerializers.transfer_from_savings = OperationDataSerializer(33, [
	// 	['from', StringSerializer],
	// 	['request_id', UInt32Serializer],
	// 	['to', StringSerializer],
	// 	['amount', AssetSerializer],
	// 	['memo', StringSerializer]
	//   ])

	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.From, &buf)
	err := binary.Write(&buf, binary.LittleEndian, uint32(o.RequestId))
	if err != nil {
		return nil, err
	}
	appendVString(o.To, &buf)
	if err := appendVAsset(o.Amount, &buf); err != nil {
		return nil, err
	}
	appendVString(o.Memo, &buf)

	return buf.Bytes(), nil
}

func (o CancelTransferFromSavings) SerializeOp() ([]byte, error) {
	// OperationSerializers.cancel_transfer_from_savings = OperationDataSerializer(
	// 	34,
	// 	[
	// 	  ['from', StringSerializer],
	// 	  ['request_id', UInt32Serializer]
	// 	]
	//   )

	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.From, &buf)
	err := binary.Write(&buf, binary.LittleEndian, uint32(o.RequestId))
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (o TransferToVesting) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.From, &buf)
	appendVString(o.To, &buf)
	err := appendVAsset(o.Amount, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// WithdrawVesting (op 4): Account + VestingShares
// Follows same pattern as TransferToVesting
func (o WithdrawVestingOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.Account, &buf)
	err := appendVAsset(o.VestingShares, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DelegateVestingShares (op 40): Delegator + Delegatee + VestingShares
func (o DelegateVestingSharesOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.Delegator, &buf)
	appendVString(o.Delegatee, &buf)
	err := appendVAsset(o.VestingShares, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// AccountWitnessProxy (op 13): Account + Proxy
// Follows same pattern as voteOperation (simple string fields)
func (o AccountWitnessProxyOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.Account, &buf)
	appendVString(o.Proxy, &buf)
	return buf.Bytes(), nil
}

// CreateClaimedAccount (op 23): Same as AccountCreate but without Fee field
// Creator + NewAccountName + Owner + Active + Posting + MemoKey + JsonMetadata + Extensions
//
// NOTE(C2): The Extensions field is present in JSON but only serialized as a zero byte
// (empty extensions) in binary. Non-empty extensions are not supported by this serializer.
func (o CreateClaimedAccountOperation) SerializeOp() ([]byte, error) {
	if len(o.Extensions) > 0 {
		return nil, errors.New("non-empty extensions not supported in binary serialization")
	}
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))
	appendVString(o.Creator, &buf)
	appendVString(o.NewAccountName, &buf)
	if err := serializeAuthority(o.Owner, &buf); err != nil {
		return nil, err
	}
	if err := serializeAuthority(o.Active, &buf); err != nil {
		return nil, err
	}
	if err := serializeAuthority(o.Posting, &buf); err != nil {
		return nil, err
	}
	err := writePublicKey(o.MemoKey, &buf)
	if err != nil {
		return nil, err
	}
	appendVString(o.JsonMetadata, &buf)
	// extensions (empty)
	buf.WriteByte(extensionsB())
	return buf.Bytes(), nil
}

func (o ClaimAccountOperation) SerializeOp() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(opIdB(o.OpName()))

	appendVString(o.Creator, &buf)
	err := appendVAsset(o.Fee, &buf)
	if err != nil {
		return nil, err
	}
	buf.Write([]byte{extensionsB()})

	return buf.Bytes(), nil
}

// todo: UNTESTED
func appendOptionalAuthority(auth *Auths, buf *bytes.Buffer) error {
	if auth != nil {
		buf.WriteByte(1) // field is present, so we prepend a 1
		return serializeAuthority(*auth, buf)
	}
	buf.WriteByte(0) // field is absent, so we write a 0
	return nil
}

// todo: UNTESTED
// encodes a uint64 into a variable-length byte slice and writes it to w
func WriteUvarint(w io.Writer, x uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], x)
	if _, err := w.Write(buf[:n]); err != nil {
		return fmt.Errorf("failed to write Uvarint: %w", err)
	}
	return nil
}

// todo: UNTESTED
// encodes an int64 into a variable-length byte slice and writes it to w
func WriteVarint(w io.Writer, x int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], x)
	if _, err := w.Write(buf[:n]); err != nil {
		return fmt.Errorf("failed to write Varint: %w", err)
	}
	return nil
}

// todo: UNTESTED
func serializeAuthority(auth Auths, buf *bytes.Buffer) error {
	// write weight_threshold
	err := binary.Write(buf, binary.LittleEndian, uint32(auth.WeightThreshold))
	if err != nil {
		return fmt.Errorf("error writing weight_threshold: %w", err)
	}

	// write account_auths
	err = WriteUvarint(buf, uint64(len(auth.AccountAuths)))
	if err != nil {
		return fmt.Errorf("error writing account_auths length: %w", err)
	}
	for _, accountAuth := range auth.AccountAuths {
		appendVString(accountAuth[0].(string), buf)
		weight, err := toUint16(accountAuth[1])
		if err != nil {
			return fmt.Errorf("error converting account_auth weight: %w", err)
		}
		err = binary.Write(buf, binary.LittleEndian, weight)
		if err != nil {
			return fmt.Errorf("error writing account_auth weight: %w", err)
		}
	}

	// write key_auths
	err = WriteUvarint(buf, uint64(len(auth.KeyAuths)))
	if err != nil {
		return fmt.Errorf("error writing key_auths length: %w", err)
	}
	for _, keyAuth := range sortKeyAuth(auth.KeyAuths) {
		err = writePublicKey(keyAuth[0].(string), buf)
		if err != nil {
			return fmt.Errorf("error writing key_auth public key: %w", err)
		}
		weight, err := toUint16(keyAuth[1])
		if err != nil {
			return fmt.Errorf("error converting key_auth weight: %w", err)
		}
		err = binary.Write(buf, binary.LittleEndian, weight)
		if err != nil {
			return fmt.Errorf("error writing key_auth weight: %w", err)
		}
	}
	return nil
}

func writePublicKey(pub string, buf *bytes.Buffer) error {
	pk, err := DecodePublicKey(pub)
	if err != nil {
		return err
	}
	return binary.Write(buf, binary.LittleEndian, pk.SerializeCompressed())
}

// toUint16 safely converts an interface{} value to uint16.
// Handles both int (from Go code) and float64 (from JSON unmarshal).
func toUint16(v interface{}) (uint16, error) {
	switch val := v.(type) {
	case int:
		return uint16(val), nil
	case float64:
		return uint16(val), nil
	case int64:
		return uint16(val), nil
	default:
		return 0, fmt.Errorf("unsupported weight type: %T", v)
	}
}

func sortKeyAuth(auths [][2]interface{}) [][2]interface{} {
	copied := make([][2]interface{}, len(auths))
	copy(copied, auths)
	sort.Slice(copied, func(i, j int) bool {
		return copied[i][0].(string) < copied[j][0].(string)
	})
	return copied
}

package hivego

// Regression tests for the C6 hivego LOW-fix cluster (Magi MAINNET audit 2026-06-03).
//
// IMPORTANT (HG-M13): these tests are in `package hivego` (in-package) so they
// exercise THIS local library at github.com/vsc-eco/hivego. The pre-existing
// keys_test.go is `package hivego_test` and imports github.com/cfoxon/hivego
// (the WRONG fork) — its assertions do not cover the code we are fixing here.
//
// Each test is PROOF-INVERTED: it asserts the post-fix (correct) behavior, and
// MUST fail against the unfixed @508b8c3 source (the original behavior asserted
// by the LOW-proof pass). A regression test that passes without the fix is
// worthless (C2 lesson).

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/decred/base58"
)

// buildWif constructs a base58check string that GphBase58CheckDecode reads as
// (payload, version). It mirrors the decoder exactly:
//
//	decoded = base58.Decode(s)
//	version = decoded[0]; payload = decoded[1:len-4]
//	checksum(decoded[:len-4]) == decoded[len-4:]
//
// where checksum is the lib's double-sha256 first-4-bytes. NOTE: we cannot use
// the lib's own GphBase58Encode here — it is asymmetric with the decoder
// (it drops the version byte from the encoded output; that is the separate
// HG-M9 finding). So we encode the way the decoder expects.
func buildWif(version byte, payload []byte) string {
	body := append([]byte{version}, payload...)
	h1 := sha256.Sum256(body)
	h2 := sha256.Sum256(h1[:])
	full := append(body, h2[:4]...)
	return base58.Encode(full)
}

// ---------------------------------------------------------------------------
// HG-L-DECTRUNC — appendVAsset must ERROR on sub-precision decimals, not
// silently truncate. Pre-fix: '1.0009 HBD' serialized byte-identical to
// '1.000 HBD' with nil error (0.0009 HBD silently dropped).
// ---------------------------------------------------------------------------

func TestC6_DecTrunc_ErrorsOnSubPrecision(t *testing.T) {
	var buf bytes.Buffer
	err := appendVAsset("1.0009 HBD", &buf) // 4 decimals, HBD precision is 3
	if err == nil {
		t.Fatalf("expected error for sub-precision decimal, got nil (silent truncation) — buf=%x", buf.Bytes())
	}
	// On error nothing meaningful should have been appended past the amount;
	// the key guarantee is the error itself (caller aborts the broadcast).
}

func TestC6_DecTrunc_VESTSSubPrecisionErrors(t *testing.T) {
	var buf bytes.Buffer
	// VESTS precision is 6; 7 decimals must error.
	if err := appendVAsset("0.5000001 VESTS", &buf); err == nil {
		t.Fatalf("expected error for 7-decimal VESTS, got nil — buf=%x", buf.Bytes())
	}
}

// Byte-identity guard: exactly-precision and fewer-than-precision decimals
// must serialize EXACTLY as before the fix (these are L1 signing bytes).
func TestC6_DecTrunc_ValidInputsByteIdentical(t *testing.T) {
	cases := []struct {
		asset string
		want  []byte // amount(int64 LE) ++ nai(uint32 LE)
	}{
		// "1.000 HBD": amount=1000, nai=((99999999+1)<<5)|3
		{"1.000 HBD", appendAssetExpected(1000, ((99999999+1)<<5)|3)},
		// "1 HBD": padded to 1.000 => amount=1000
		{"1 HBD", appendAssetExpected(1000, ((99999999+1)<<5)|3)},
		// "1.5 HIVE": padded to 1.500 => amount=1500
		{"1.5 HIVE", appendAssetExpected(1500, ((99999999+2)<<5)|3)},
		// "0.500000 VESTS": amount=500000
		{"0.500000 VESTS", appendAssetExpected(500000, ((99999999+3)<<5)|6)},
		// "12.345 TESTS": amount=12345
		{"12.345 TESTS", appendAssetExpected(12345, ((99999999+2)<<5)|3)},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := appendVAsset(c.asset, &buf); err != nil {
			t.Fatalf("%q: unexpected error: %v", c.asset, err)
		}
		if !bytes.Equal(buf.Bytes(), c.want) {
			t.Errorf("%q: byte mismatch\n got %x\nwant %x", c.asset, buf.Bytes(), c.want)
		}
	}
}

func appendAssetExpected(amount int64, nai uint32) []byte {
	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, amount)
	naiBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(naiBytes, nai)
	b.Write(naiBytes)
	return b.Bytes()
}

// ---------------------------------------------------------------------------
// HG-L-WIFLEN — KeyPairFromWif must reject a decoded payload that is not
// exactly 32 bytes. Pre-fix: a 33-byte payload silently derived a (wrong) key.
// ---------------------------------------------------------------------------

func TestC6_WifLen_Valid32ByteWifAccepted(t *testing.T) {
	// Canonical valid Hive WIF (32-byte scalar). Must still succeed and derive
	// the SAME public key as before (no behavior change for valid input).
	kp, err := KeyPairFromWif("5JUvJcF6rQvFbZLtDFagreKCYWWcHpHApy7sbRHZ6PeZYNftLh6")
	if err != nil {
		t.Fatalf("valid WIF rejected: %v", err)
	}
	wantPub := []byte{3, 106, 48, 22, 243, 45, 96, 255, 51, 197, 8, 179, 85, 147, 131, 32, 165, 214, 76, 64, 90, 168, 63, 67, 124, 7, 139, 26, 114, 145, 144, 94, 153}
	if !bytes.Equal(kp.PublicKey.SerializeCompressed(), wantPub) {
		t.Errorf("valid WIF derived wrong pubkey: got %x want %x", kp.PublicKey.SerializeCompressed(), wantPub)
	}
}

func TestC6_WifLen_OversizedPayloadRejected(t *testing.T) {
	// Build a WIF-looking string whose decoded payload (after stripping version
	// byte + checksum) is 33 bytes instead of 32. Pre-fix this was silently
	// accepted and produced a distinct key; post-fix it must error.
	payload33 := make([]byte, 33)
	for i := range payload33 {
		payload33[i] = byte(i + 1)
	}
	wif := buildWif(0x80, payload33) // 0x80 = mainnet WIF version byte

	if _, err := KeyPairFromWif(wif); err == nil {
		t.Fatalf("expected error for 33-byte decoded payload, got nil (wrong key derived silently)")
	}

	// Sanity: the 32-byte version of the same data must be ACCEPTED (proves the
	// rejection is specifically the length gate, not a malformed-WIF reject).
	payload32 := payload33[:32]
	wif32 := buildWif(0x80, payload32)
	if _, err := KeyPairFromWif(wif32); err != nil {
		t.Fatalf("32-byte payload should be accepted, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HG-L-ADDSIG — AddSig must not append a duplicate signature. Pre-fix: the
// same sig added twice yielded len==2.
// ---------------------------------------------------------------------------

func TestC6_AddSig_DeduplicatesIdenticalSig(t *testing.T) {
	tx := HiveTransaction{}
	const sig = "1f4e2c...deadbeefsignaturehex"
	tx.AddSig(sig)
	tx.AddSig(sig)
	if len(tx.Signatures) != 1 {
		t.Fatalf("expected 1 sig after adding the same sig twice, got %d: %v", len(tx.Signatures), tx.Signatures)
	}
}

func TestC6_AddSig_DistinctSigsAllAppended(t *testing.T) {
	// Distinct sigs must all be retained, in order (no behavior change).
	tx := HiveTransaction{}
	tx.AddSig("aaa")
	tx.AddSig("bbb")
	tx.AddSig("ccc")
	tx.AddSig("aaa") // duplicate of first -> dropped
	want := []string{"aaa", "bbb", "ccc"}
	if len(tx.Signatures) != len(want) {
		t.Fatalf("expected %d sigs, got %d: %v", len(want), len(tx.Signatures), tx.Signatures)
	}
	for i := range want {
		if tx.Signatures[i] != want[i] {
			t.Errorf("sig[%d]=%q want %q", i, tx.Signatures[i], want[i])
		}
	}
}

// ---------------------------------------------------------------------------
// HG-L-VSTR255 — appendVStringArray must ERROR when len > 255 instead of
// truncating the 1-byte count (256 -> 0x00). CustomJsonOperation.SerializeOp
// must surface that error. Pre-fix: 256 required_auths -> count prefix 0x00,
// nil error.
// ---------------------------------------------------------------------------

func TestC6_VStr255_ErrorsOn256Elements(t *testing.T) {
	var buf bytes.Buffer
	a := make([]string, 256)
	for i := range a {
		a[i] = "x"
	}
	if err := appendVStringArray(a, &buf); err == nil {
		t.Fatalf("expected error for 256-element array, got nil (count byte would wrap to 0x00)")
	}
}

func TestC6_VStr255_CustomJsonSerializeSurfacesError(t *testing.T) {
	auths := make([]string, 256)
	for i := range auths {
		auths[i] = "acct"
	}
	op := CustomJsonOperation{
		RequiredAuths:        auths,
		RequiredPostingAuths: []string{},
		Id:                   "test",
		Json:                 "{}",
		opText:               "custom_json",
	}
	b, err := op.SerializeOp()
	if err == nil {
		t.Fatalf("CustomJsonOperation.SerializeOp must return error for 256 required_auths; got nil, buf=%x", b)
	}
}

// Byte-identity guard: a 255-element array (the max) and small arrays must
// serialize exactly as the 1-byte-count format expects.
func TestC6_VStr255_BoundaryAndSmallByteIdentical(t *testing.T) {
	// 255 elements -> count byte 0xff, no error.
	var buf255 bytes.Buffer
	a255 := make([]string, 255)
	for i := range a255 {
		a255[i] = "x"
	}
	if err := appendVStringArray(a255, &buf255); err != nil {
		t.Fatalf("255-element array must NOT error: %v", err)
	}
	if got := buf255.Bytes()[0]; got != 0xff {
		t.Errorf("255-element count byte = 0x%02x, want 0xff", got)
	}

	// Small case identical to the existing TestAppendVStringArray vector.
	var buf bytes.Buffer
	if err := appendVStringArray([]string{"xeroc", "piston"}, &buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte{2, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("small array byte mismatch: got %x want %x", buf.Bytes(), want)
	}
}

// ---------------------------------------------------------------------------
// HG-L-REQID — RequestId is now uint32; a negative request_id can no longer be
// stored (compile-time). The serialized 4 LE bytes must equal the true uint32.
// Pre-fix: RequestId int + uint32(-1) -> 0xffffffff silently.
//
// The original wrap test (RequestId: -1) no longer COMPILES with a uint32 field
// — that is the regression gate (the bad value is unrepresentable). We instead
// assert correct serialization across the full uint32 range, including the
// boundary that used to be reached only via negative wrap.
// ---------------------------------------------------------------------------

func TestC6_ReqId_SerializesUint32Correctly(t *testing.T) {
	cases := []uint32{0, 1, 7320629, 4294967295} // incl. max uint32 (==0xffffffff)
	for _, id := range cases {
		op := TransferFromSavings{
			Amount:    "1.000 HBD",
			From:      "alice",
			To:        "alice",
			Memo:      "",
			RequestId: id,
		}
		b, err := op.SerializeOp()
		if err != nil {
			t.Fatalf("RequestId=%d: serialize error: %v", id, err)
		}
		// Layout: opId(1) ++ vstring(from) ++ uint32(request_id LE) ++ ...
		// from = "alice" -> 1 (len) + 5 bytes = 6 bytes; request_id starts at
		// offset 1+6 = 7.
		off := 1 + 1 + len("alice")
		gotID := binary.LittleEndian.Uint32(b[off : off+4])
		if gotID != id {
			t.Errorf("RequestId=%d serialized as %d (bytes %x)", id, gotID, b[off:off+4])
		}
	}
}

func TestC6_ReqId_CancelSerializesUint32Correctly(t *testing.T) {
	op := CancelTransferFromSavings{From: "alice", RequestId: 4294967295}
	b, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}
	off := 1 + 1 + len("alice")
	gotID := binary.LittleEndian.Uint32(b[off : off+4])
	if gotID != 4294967295 {
		t.Errorf("CancelTransferFromSavings RequestId serialized as %d, want 4294967295", gotID)
	}
}

// ---------------------------------------------------------------------------
// AF-1 (HG-L-DECTRUNC follow-up) — the 3 transfer SerializeOp methods that
// bare-call appendVAsset (TransferOperation, TransferToSavings,
// TransferFromSavings) must SURFACE the new appendVAsset error instead of
// discarding it. Pre-AF1 (the original C6 fix): a sub-precision Amount made
// appendVAsset return an error that these three methods threw away, so they
// emitted a MALFORMED signed L1 tx with a nil error. These tests are
// PROOF-INVERTED: each must FAIL against the pre-AF1 (current C6) code (which
// returns nil error) and PASS only once the error is wired through.
//
// The byte-identity guards prove valid (in-precision / shorter) Amounts still
// serialize EXACTLY as before — no change to L1 signing bytes for good input.
// ---------------------------------------------------------------------------

func TestAF1_TransferOperation_SubPrecisionErrors(t *testing.T) {
	op := TransferOperation{From: "alice", To: "bob", Amount: "1.0009 HBD", Memo: ""}
	b, err := op.SerializeOp()
	if err == nil {
		t.Fatalf("TransferOperation.SerializeOp must error on sub-precision Amount; got nil, malformed buf=%x", b)
	}
}

func TestAF1_TransferToSavings_SubPrecisionErrors(t *testing.T) {
	op := TransferToSavings{From: "alice", To: "bob", Amount: "1.0009 HBD", Memo: ""}
	b, err := op.SerializeOp()
	if err == nil {
		t.Fatalf("TransferToSavings.SerializeOp must error on sub-precision Amount; got nil, malformed buf=%x", b)
	}
}

func TestAF1_TransferFromSavings_SubPrecisionErrors(t *testing.T) {
	op := TransferFromSavings{From: "alice", To: "bob", Amount: "1.0009 HBD", Memo: "", RequestId: 7}
	b, err := op.SerializeOp()
	if err == nil {
		t.Fatalf("TransferFromSavings.SerializeOp must error on sub-precision Amount; got nil, malformed buf=%x", b)
	}
}

// Byte-identity: valid Amount serialization is unchanged. We reconstruct the
// expected buffer from the SAME serializer primitives (appendVString /
// appendVAsset) the methods use, which on VALID input behave identically
// pre- and post-AF1 (AF-1 only adds an error branch, never alters the bytes).

func TestAF1_TransferOperation_ValidByteIdentical(t *testing.T) {
	op := TransferOperation{From: "alice", To: "bob", Amount: "1.500 HBD", Memo: "hello"}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("valid TransferOperation unexpectedly errored: %v", err)
	}

	var want bytes.Buffer
	want.Write([]byte{opIdB(op.OpName())})
	appendVString(op.From, &want)
	appendVString(op.To, &want)
	if err := appendVAsset(op.Amount, &want); err != nil {
		t.Fatalf("building expected vector: %v", err)
	}
	appendVString(op.Memo, &want)

	if !bytes.Equal(got, want.Bytes()) {
		t.Errorf("TransferOperation byte mismatch\n got %x\nwant %x", got, want.Bytes())
	}
}

func TestAF1_TransferToSavings_ValidByteIdentical(t *testing.T) {
	op := TransferToSavings{From: "alice", To: "bob", Amount: "2.000 HIVE", Memo: "save"}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("valid TransferToSavings unexpectedly errored: %v", err)
	}

	var want bytes.Buffer
	want.WriteByte(opIdB(op.OpName()))
	appendVString(op.From, &want)
	appendVString(op.To, &want)
	if err := appendVAsset(op.Amount, &want); err != nil {
		t.Fatalf("building expected vector: %v", err)
	}
	appendVString(op.Memo, &want)

	if !bytes.Equal(got, want.Bytes()) {
		t.Errorf("TransferToSavings byte mismatch\n got %x\nwant %x", got, want.Bytes())
	}
}

func TestAF1_TransferFromSavings_ValidByteIdentical(t *testing.T) {
	op := TransferFromSavings{From: "alice", To: "bob", Amount: "0.123 HBD", Memo: "m", RequestId: 42}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("valid TransferFromSavings unexpectedly errored: %v", err)
	}

	var want bytes.Buffer
	want.WriteByte(opIdB(op.OpName()))
	appendVString(op.From, &want)
	if err := binary.Write(&want, binary.LittleEndian, uint32(op.RequestId)); err != nil {
		t.Fatalf("building expected vector (request_id): %v", err)
	}
	appendVString(op.To, &want)
	if err := appendVAsset(op.Amount, &want); err != nil {
		t.Fatalf("building expected vector (amount): %v", err)
	}
	appendVString(op.Memo, &want)

	if !bytes.Equal(got, want.Bytes()) {
		t.Errorf("TransferFromSavings byte mismatch\n got %x\nwant %x", got, want.Bytes())
	}
}

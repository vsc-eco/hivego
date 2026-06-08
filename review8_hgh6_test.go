package hivego

// review8 HG-H6 — HashTxForSig silently dropped an invalid chain id (cid, _ :=
// hex.DecodeString), leaving the signing digest with NO chain prefix (== a
// different/empty chain's digest), enabling cross-chain signature replay. It now
// returns an error instead.

import (
	"bytes"
	"testing"
)

func TestReview8_HGH6_HashTxForSigRejectsInvalidChainID(t *testing.T) {
	tx := []byte("some-serialized-tx")

	// Valid custom chain id: digest is prefixed with that chain id, so it differs
	// from the default-chain digest.
	validCID := "fedcba9800000000000000000000000000000000000000000000000000000000"
	dValid, err := HashTxForSig(tx, validCID)
	if err != nil {
		t.Fatalf("valid chain id should not error: %v", err)
	}
	dDefault, err := HashTxForSig(tx)
	if err != nil {
		t.Fatalf("default chain id should not error: %v", err)
	}
	if bytes.Equal(dValid, dDefault) {
		t.Fatal("valid custom chain id must change the digest")
	}

	// Invalid hex chain id: pre-fix this silently produced a NO-prefix digest,
	// enabling cross-chain replay. It must now be a hard error.
	if _, err := HashTxForSig(tx, "not-hex"); err == nil {
		t.Fatal("HG-H6: invalid chain id must return an error, not a silent no-prefix digest")
	}
	// Odd-length hex is also invalid.
	if _, err := HashTxForSig(tx, "abc"); err == nil {
		t.Fatal("HG-H6: odd-length hex chain id must return an error")
	}
}

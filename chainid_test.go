package hivego

import (
	"encoding/hex"
	"testing"
)

func TestCustomChainID(t *testing.T) {
	// Test that custom chain ID is used when provided
	tx := getTestVoteTx()
	customChainID := "1234567800000000000000000000000000000000000000000000000000000000"

	// Test signing with custom chain ID
	message, err := SerializeTx(tx)
	if err != nil {
		t.Fatal("Failed to serialize transaction:", err)
	}

	// Create a test key pair
	wif := "5JuMt237G3m3BaT7zH4YdoycUtbw4AEPy6DLdCrKAnFGAtXyQ1W"
	keyPair, err := KeyPairFromWif(wif)
	if err != nil {
		t.Fatal("Failed to create key pair:", err)
	}

	// Sign with custom chain ID
	sig, err := tx.Sign(*keyPair, customChainID)
	if err != nil {
		t.Fatal("Failed to sign transaction:", err)
	}

	// Verify signature is not empty
	if sig == "" {
		t.Error("Signature should not be empty")
	}

	// Test that the signature is different from default chain ID signature
	sigDefault, err := tx.Sign(*keyPair)
	if err != nil {
		t.Fatal("Failed to sign transaction with default chain ID:", err)
	}

	// Signatures should be different
	if sig == sigDefault {
		t.Error("Signatures with custom and default chain IDs should be different")
	}

	// Test HashTxForSig with custom chain ID
	digestCustom, err := HashTxForSig(message, customChainID)
	if err != nil {
		t.Fatalf("HashTxForSig(custom) returned error: %v", err)
	}
	digestDefault, err := HashTxForSig(message) // Uses default chain ID
	if err != nil {
		t.Fatalf("HashTxForSig(default) returned error: %v", err)
	}

	// Digests should be different
	if hex.EncodeToString(digestCustom) == hex.EncodeToString(digestDefault) {
		t.Error("Digests with custom and default chain IDs should be different")
	}
}

func TestHiveRpcNodeChainID(t *testing.T) {
	// Test that HiveRpcNode uses its ChainID property
	node := NewHiveRpcWithOpts([]string{"https://api.hive.blog"}, 1, 1)
	customChainID := "fedcba9800000000000000000000000000000000000000000000000000000000"
	node.ChainID = customChainID

	// The node should have the custom chain ID
	if node.ChainID != customChainID {
		t.Errorf("Expected node ChainID to be %s, got %s", customChainID, node.ChainID)
	}
}

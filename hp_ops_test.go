package hivego

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// ============================================================
// TEST DATA
// ============================================================

func getTestWithdrawVesting() WithdrawVestingOperation {
	return WithdrawVestingOperation{
		Account:       "magi.test1",
		VestingShares: "100.000000 VESTS",
	}
}

func getTestDelegateVestingShares() DelegateVestingSharesOperation {
	return DelegateVestingSharesOperation{
		Delegator:     "magi-v-abcd1234",
		Delegatee:     "lordbutterfly",
		VestingShares: "50000.000000 VESTS",
	}
}

func getTestDelegateVestingSharesZero() DelegateVestingSharesOperation {
	return DelegateVestingSharesOperation{
		Delegator:     "magi-v-abcd1234",
		Delegatee:     "lordbutterfly",
		VestingShares: "0.000000 VESTS",
	}
}

func getTestAccountWitnessProxy() AccountWitnessProxyOperation {
	return AccountWitnessProxyOperation{
		Account: "magi-v-abcd1234",
		Proxy:   "lordbutterfly",
	}
}

func getTestAccountWitnessProxyClear() AccountWitnessProxyOperation {
	return AccountWitnessProxyOperation{
		Account: "magi-v-abcd1234",
		Proxy:   "",
	}
}

func getTestCreateClaimedAccount() CreateClaimedAccountOperation {
	return CreateClaimedAccountOperation{
		Creator:        "vsc.gateway",
		NewAccountName: "magi-v-abcd1234",
		Owner: Auths{
			WeightThreshold: 2,
			KeyAuths:        [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}},
		},
		Active: Auths{
			WeightThreshold: 2,
			KeyAuths:        [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}},
		},
		Posting: Auths{
			WeightThreshold: 1,
			KeyAuths:        [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}},
		},
		MemoKey:      "STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow",
		JsonMetadata: "",
		Extensions:   []interface{}{},
	}
}

// ============================================================
// OP ID TESTS — verify op IDs are correct
// ============================================================

func TestWithdrawVestingOpId(t *testing.T) {
	got := opIdB("withdraw_vesting")
	if got != 4 {
		t.Errorf("WithdrawVesting op ID: expected 4, got %d", got)
	}
}

func TestDelegateVestingSharesOpId(t *testing.T) {
	got := opIdB("delegate_vesting_shares")
	if got != 40 {
		t.Errorf("DelegateVestingShares op ID: expected 40, got %d", got)
	}
}

func TestAccountWitnessProxyOpId(t *testing.T) {
	got := opIdB("account_witness_proxy")
	if got != 13 {
		t.Errorf("AccountWitnessProxy op ID: expected 13, got %d", got)
	}
}

func TestCreateClaimedAccountOpId(t *testing.T) {
	got := opIdB("create_claimed_account")
	if got != 23 {
		t.Errorf("CreateClaimedAccount op ID: expected 23, got %d", got)
	}
}

// ============================================================
// OPNAME TESTS — verify OpName() returns correct string
// ============================================================

func TestWithdrawVestingOpName(t *testing.T) {
	op := getTestWithdrawVesting()
	if op.OpName() != "withdraw_vesting" {
		t.Errorf("expected 'withdraw_vesting', got '%s'", op.OpName())
	}
}

func TestDelegateVestingSharesOpName(t *testing.T) {
	op := getTestDelegateVestingShares()
	if op.OpName() != "delegate_vesting_shares" {
		t.Errorf("expected 'delegate_vesting_shares', got '%s'", op.OpName())
	}
}

func TestAccountWitnessProxyOpName(t *testing.T) {
	op := getTestAccountWitnessProxy()
	if op.OpName() != "account_witness_proxy" {
		t.Errorf("expected 'account_witness_proxy', got '%s'", op.OpName())
	}
}

func TestCreateClaimedAccountOpName(t *testing.T) {
	op := getTestCreateClaimedAccount()
	if op.OpName() != "create_claimed_account" {
		t.Errorf("expected 'create_claimed_account', got '%s'", op.OpName())
	}
}

// ============================================================
// SERIALIZATION TESTS — verify binary output matches expected
// ============================================================

func TestSerializeWithdrawVesting(t *testing.T) {
	op := getTestWithdrawVesting()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// Hardcoded expected bytes: op ID 4 + vstring("magi.test1") + asset(100.000000 VESTS)
	expected := []byte{4, 10, 109, 97, 103, 105, 46, 116, 101, 115, 116, 49, 0, 225, 245, 5, 0, 0, 0, 0, 70, 32, 188, 190}
	if !bytes.Equal(got, expected) {
		t.Errorf("WithdrawVesting serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}

	// Verify op ID byte is first
	if got[0] != 4 {
		t.Errorf("first byte should be op ID 4, got %d", got[0])
	}
}

func TestSerializeDelegateVestingShares(t *testing.T) {
	op := getTestDelegateVestingShares()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// Hardcoded expected bytes: op ID 40 + vstring("magi-v-abcd1234") + vstring("lordbutterfly") + asset(50000.000000 VESTS)
	expected := []byte{40, 15, 109, 97, 103, 105, 45, 118, 45, 97, 98, 99, 100, 49, 50, 51, 52, 13, 108, 111, 114, 100, 98, 117, 116, 116, 101, 114, 102, 108, 121, 0, 116, 59, 164, 11, 0, 0, 0, 70, 32, 188, 190}
	if !bytes.Equal(got, expected) {
		t.Errorf("DelegateVestingShares serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}

	if got[0] != 40 {
		t.Errorf("first byte should be op ID 40, got %d", got[0])
	}
}

func TestSerializeDelegateVestingSharesZero(t *testing.T) {
	// Zero delegation = undelegation on Hive
	op := getTestDelegateVestingSharesZero()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// Hardcoded expected bytes: op ID 40 + vstring("magi-v-abcd1234") + vstring("lordbutterfly") + asset(0.000000 VESTS)
	expected := []byte{40, 15, 109, 97, 103, 105, 45, 118, 45, 97, 98, 99, 100, 49, 50, 51, 52, 13, 108, 111, 114, 100, 98, 117, 116, 116, 101, 114, 102, 108, 121, 0, 0, 0, 0, 0, 0, 0, 0, 70, 32, 188, 190}
	if !bytes.Equal(got, expected) {
		t.Errorf("DelegateVestingShares zero serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}

	// Verify the amount bytes are all zero (int64 = 0)
	// The amount is the last 12 bytes: 8 bytes int64 + 4 bytes NAI
	amountStart := len(got) - 12
	amountBytes := got[amountStart : amountStart+8]
	amount := binary.LittleEndian.Uint64(amountBytes)
	if amount != 0 {
		t.Errorf("expected zero VESTS amount, got %d", amount)
	}
}

func TestSerializeAccountWitnessProxy(t *testing.T) {
	op := getTestAccountWitnessProxy()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// Hardcoded expected bytes: op ID 13 + vstring("magi-v-abcd1234") + vstring("lordbutterfly")
	expected := []byte{13, 15, 109, 97, 103, 105, 45, 118, 45, 97, 98, 99, 100, 49, 50, 51, 52, 13, 108, 111, 114, 100, 98, 117, 116, 116, 101, 114, 102, 108, 121}
	if !bytes.Equal(got, expected) {
		t.Errorf("AccountWitnessProxy serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}

	if got[0] != 13 {
		t.Errorf("first byte should be op ID 13, got %d", got[0])
	}
}

func TestSerializeAccountWitnessProxyClear(t *testing.T) {
	// Empty proxy string = clear proxy on Hive
	op := getTestAccountWitnessProxyClear()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// Hardcoded expected bytes: op ID 13 + vstring("magi-v-abcd1234") + vstring("")
	expected := []byte{13, 15, 109, 97, 103, 105, 45, 118, 45, 97, 98, 99, 100, 49, 50, 51, 52, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("AccountWitnessProxy clear serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}
}

func TestSerializeCreateClaimedAccount(t *testing.T) {
	op := getTestCreateClaimedAccount()
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("SerializeOp failed: %v", err)
	}

	// A9: Hardcoded expected bytes for CreateClaimedAccount
	expected := []byte{23, 11, 118, 115, 99, 46, 103, 97, 116, 101, 119, 97, 121, 15, 109, 97, 103, 105, 45, 118, 45, 97, 98, 99, 100, 49, 50, 51, 52, 2, 0, 0, 0, 0, 1, 2, 10, 101, 192, 10, 6, 132, 7, 65, 238, 81, 177, 178, 164, 187, 202, 162, 70, 26, 198, 248, 227, 102, 116, 96, 8, 245, 232, 159, 143, 49, 25, 233, 1, 0, 2, 0, 0, 0, 0, 1, 2, 10, 101, 192, 10, 6, 132, 7, 65, 238, 81, 177, 178, 164, 187, 202, 162, 70, 26, 198, 248, 227, 102, 116, 96, 8, 245, 232, 159, 143, 49, 25, 233, 1, 0, 1, 0, 0, 0, 0, 1, 2, 10, 101, 192, 10, 6, 132, 7, 65, 238, 81, 177, 178, 164, 187, 202, 162, 70, 26, 198, 248, 227, 102, 116, 96, 8, 245, 232, 159, 143, 49, 25, 233, 1, 0, 2, 10, 101, 192, 10, 6, 132, 7, 65, 238, 81, 177, 178, 164, 187, 202, 162, 70, 26, 198, 248, 227, 102, 116, 96, 8, 245, 232, 159, 143, 49, 25, 233, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Errorf("CreateClaimedAccount serialization mismatch\nexpected: %v\ngot:      %v", expected, got)
	}
}

// ============================================================
// INTERFACE COMPLIANCE — verify all ops implement HiveOperation
// ============================================================

func TestHiveOperationInterface(t *testing.T) {
	// These must compile. If they don't, the types don't implement HiveOperation.
	var _ HiveOperation = WithdrawVestingOperation{}
	var _ HiveOperation = DelegateVestingSharesOperation{}
	var _ HiveOperation = AccountWitnessProxyOperation{}
	var _ HiveOperation = CreateClaimedAccountOperation{}
}

// ============================================================
// TRANSACTION INTEGRATION — verify ops work inside a full tx
// ============================================================

func TestSerializeTxWithWithdrawVesting(t *testing.T) {
	op := getTestWithdrawVesting()
	tx := getTestTx([]HiveOperation{op})
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with WithdrawVesting failed: %v", err)
	}
	if len(got) < 20 {
		t.Errorf("serialized tx too short: %d bytes", len(got))
	}
}

func TestSerializeTxWithDelegateVestingShares(t *testing.T) {
	op := getTestDelegateVestingShares()
	tx := getTestTx([]HiveOperation{op})
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with DelegateVestingShares failed: %v", err)
	}
	if len(got) < 20 {
		t.Errorf("serialized tx too short: %d bytes", len(got))
	}
}

func TestSerializeTxWithAccountWitnessProxy(t *testing.T) {
	op := getTestAccountWitnessProxy()
	tx := getTestTx([]HiveOperation{op})
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with AccountWitnessProxy failed: %v", err)
	}
	if len(got) < 20 {
		t.Errorf("serialized tx too short: %d bytes", len(got))
	}
}

func TestSerializeTxWithCreateClaimedAccount(t *testing.T) {
	op := getTestCreateClaimedAccount()
	tx := getTestTx([]HiveOperation{op})
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with CreateClaimedAccount failed: %v", err)
	}
	if len(got) < 20 {
		t.Errorf("serialized tx too short: %d bytes", len(got))
	}
}

// ============================================================
// MULTI-OP TRANSACTION — verify all 4 new ops serialize together
// ============================================================

func TestSerializeTxWithAllNewOps(t *testing.T) {
	ops := []HiveOperation{
		getTestWithdrawVesting(),
		getTestDelegateVestingShares(),
		getTestAccountWitnessProxy(),
		getTestCreateClaimedAccount(),
	}
	tx := getTestTx(ops)
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with all new ops failed: %v", err)
	}
	if len(got) < 50 {
		t.Errorf("serialized multi-op tx too short: %d bytes", len(got))
	}
}

// ============================================================
// MIXED OLD+NEW OPS — verify new ops don't break existing ones
// ============================================================

func TestSerializeTxMixedOldAndNewOps(t *testing.T) {
	ops := []HiveOperation{
		getTestTransferOp(),           // existing: Transfer
		getTestPowerUp(),              // existing: TransferToVesting
		getTestWithdrawVesting(),      // new: WithdrawVesting
		getTestDelegateVestingShares(), // new: DelegateVestingShares
		getTestAccountWitnessProxy(),  // new: AccountWitnessProxy
		getTestCreateClaimedAccount(), // new: CreateClaimedAccount
	}
	tx := getTestTx(ops)
	got, err := SerializeTx(tx)
	if err != nil {
		t.Fatalf("SerializeTx with mixed ops failed: %v", err)
	}
	if len(got) < 80 {
		t.Errorf("serialized mixed tx too short: %d bytes", len(got))
	}
}

// ============================================================
// REGRESSION — verify existing ops still serialize correctly
// ============================================================

func TestRegressionTransferToVesting(t *testing.T) {
	got, err := getTestPowerUp().SerializeOp()
	if err != nil {
		t.Fatalf("TransferToVesting serialization failed: %v", err)
	}
	expected := []byte{3, 9, 105, 110, 105, 116, 109, 105, 110, 101, 114, 10, 109, 97, 103, 105, 46, 116, 101, 115, 116, 49, 160, 134, 1, 0, 0, 0, 0, 0, 35, 32, 188, 190}
	if !bytes.Equal(got, expected) {
		t.Errorf("TransferToVesting REGRESSION: serialization changed!\nexpected: %v\ngot:      %v", expected, got)
	}
}

func TestRegressionTransfer(t *testing.T) {
	got, err := getTestTransferOp().SerializeOp()
	if err != nil {
		t.Fatalf("Transfer serialization failed: %v", err)
	}
	expected := []byte{2, 10, 116, 105, 98, 102, 111, 120, 46, 118,
		115, 99, 11, 118, 115, 99, 46, 103, 97, 116,
		101, 119, 97, 121, 232, 3, 0, 0, 0, 0,
		0, 0, 35, 32, 188, 190, 9, 116, 111, 61,
		116, 105, 98, 102, 111, 120}
	if !bytes.Equal(got, expected) {
		t.Errorf("Transfer REGRESSION: serialization changed!\nexpected: %v\ngot:      %v", expected, got)
	}
}

func TestRegressionClaimAccount(t *testing.T) {
	got, err := getTestClaimAcc().SerializeOp()
	if err != nil {
		t.Fatalf("ClaimAccount serialization failed: %v", err)
	}
	expected := []byte{
		22, 10, 116, 101, 99, 104, 99, 111,
		100, 101, 114, 120, 0, 0, 0, 0,
		0, 0, 0, 0, 35, 32, 188, 190,
		0,
	}
	if !bytes.Equal(got, expected) {
		t.Errorf("ClaimAccount REGRESSION: serialization changed!\nexpected: %v\ngot:      %v", expected, got)
	}
}

func TestRegressionAccountUpdate(t *testing.T) {
	got, err := getTestAccountUpdateOp().SerializeOp()
	if err != nil {
		t.Fatalf("AccountUpdate serialization failed: %v", err)
	}
	expected := []byte{10, 12, 115, 110, 105, 112, 101, 114, 100, 117, 101, 108, 49, 55, 0, 0, 0, 2, 248, 203, 193, 109, 141, 110, 237, 126, 105, 254, 86, 201, 65, 157, 81, 189, 244, 224, 193, 227, 202, 141, 140, 24, 154, 173, 150, 112, 27, 195, 12, 77, 13, 123, 34, 102, 111, 111, 34, 58, 34, 98, 97, 114, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Errorf("AccountUpdate REGRESSION: serialization changed!\nexpected: %v\ngot:      %v", expected, got)
	}
}

func TestRegressionVote(t *testing.T) {
	got, err := getTestVoteOp().SerializeOp()
	if err != nil {
		t.Fatalf("Vote serialization failed: %v", err)
	}
	expected := []byte{0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39}
	if !bytes.Equal(got, expected) {
		t.Errorf("Vote REGRESSION: serialization changed!\nexpected: %v\ngot:      %v", expected, got)
	}
}

// ============================================================
// A3: Verify all known op names return expected non-zero IDs
// (except "vote" which is legitimately 0)
// ============================================================

func TestAllKnownOpIdsAreCorrect(t *testing.T) {
	knownOps := map[string]byte{
		"vote":                         0,
		"transfer":                     2,
		"transfer_to_vesting":          3,
		"withdraw_vesting":             4,
		"account_create":               9,
		"account_update":               10,
		"account_witness_proxy":        13,
		"custom_json":                  18,
		"claim_account":                22,
		"create_claimed_account":       23,
		"transfer_to_savings":          32,
		"transfer_from_savings":        33,
		"cancel_transfer_from_savings": 34,
		"delegate_vesting_shares":      40,
	}
	for name, expected := range knownOps {
		got := opIdB(name)
		if got != expected {
			t.Errorf("opIdB(%q) = %d, want %d", name, got, expected)
		}
	}
}

func TestUnknownOpIdReturnsZero(t *testing.T) {
	// WARNING: unknown op names return 0 (same as "vote"). This is a known
	// limitation documented in A3. Any new operation MUST be added to
	// getHiveOpIds() before use. This test documents the behavior.
	got := opIdB("totally_bogus_op")
	if got != 0 {
		t.Errorf("opIdB for unknown op should return 0, got %d", got)
	}
}

// ============================================================
// C3: Negative/error-path tests for hivego serialization ops
// ============================================================

func TestWithdrawVesting_EmptyAccount(t *testing.T) {
	op := WithdrawVestingOperation{
		Account:       "",
		VestingShares: "100.000000 VESTS",
	}
	got, err := op.SerializeOp()
	// Empty account is a valid string (0-length vstring). SerializeOp does not validate.
	// This test documents that no panic occurs.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output even with empty account")
	}
}

func TestDelegateVestingShares_InvalidVESTS(t *testing.T) {
	op := DelegateVestingSharesOperation{
		Delegator:     "magi-v-abcd1234",
		Delegatee:     "lordbutterfly",
		VestingShares: "not-a-number VESTS",
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Error("expected error for invalid VESTS amount string, got nil")
	}
}

func TestAccountWitnessProxy_EmptyAccount(t *testing.T) {
	op := AccountWitnessProxyOperation{
		Account: "",
		Proxy:   "someproxy",
	}
	got, err := op.SerializeOp()
	// Empty account is not validated by SerializeOp — this documents no panic.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output even with empty account")
	}
}

func TestCreateClaimedAccount_EmptyCreator(t *testing.T) {
	op := CreateClaimedAccountOperation{
		Creator:        "",
		NewAccountName: "testaccount",
		Owner:          Auths{WeightThreshold: 1, KeyAuths: [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}}},
		Active:         Auths{WeightThreshold: 1, KeyAuths: [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}}},
		Posting:        Auths{WeightThreshold: 1, KeyAuths: [][2]interface{}{{"STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow", 1}}},
		MemoKey:        "STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow",
		JsonMetadata:   "",
		Extensions:     []interface{}{},
	}
	got, err := op.SerializeOp()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) == 0 {
		t.Error("expected non-empty output even with empty creator")
	}
}

func TestCreateClaimedAccount_NonEmptyExtensionsRejected(t *testing.T) {
	op := CreateClaimedAccountOperation{
		Creator:        "vsc.gateway",
		NewAccountName: "testaccount",
		Owner:          Auths{WeightThreshold: 1},
		Active:         Auths{WeightThreshold: 1},
		Posting:        Auths{WeightThreshold: 1},
		MemoKey:        "STM4y4wdy4eNBVBzXAXEp5SSrXEQMqBstDu6TvMGN1aUz19zAruow",
		JsonMetadata:   "",
		Extensions:     []interface{}{"something"},
	}
	_, err := op.SerializeOp()
	if err == nil {
		t.Error("expected error for non-empty extensions, got nil")
	}
}

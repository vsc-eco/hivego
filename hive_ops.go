package hivego

import (
	"encoding/hex"
	"fmt"
)

type HiveOperation interface {
	SerializeOp() ([]byte, error)
	OpName() string
}

type voteOperation struct {
	Voter    string `json:"voter"`
	Author   string `json:"author"`
	Permlink string `json:"permlink"`
	Weight   int16  `json:"weight"`
	opText   string
}

func (o voteOperation) OpName() string {
	return "vote"
}

func (h *HiveRpcNode) VotePost(voter string, author string, permlink string, weight int, wif *string) (string, error) {
	vote := voteOperation{voter, author, permlink, int16(weight), "vote"}

	return h.Broadcast([]HiveOperation{vote}, wif)
}

type TransferFromSavings struct {
	Amount    string `json:"amount"`
	From      string `json:"from"`
	To        string `json:"to"`
	Memo      string `json:"memo"`
	RequestId int    `json:"request_id"`
}

func (o TransferFromSavings) OpName() string {
	return "transfer_from_savings"
}

type TransferToSavings struct {
	Amount string `json:"amount"`
	From   string `json:"from"`
	To     string `json:"to"`
	Memo   string `json:"memo"`
}

func (o TransferToSavings) OpName() string {
	return "transfer_to_savings"
}

type CancelTransferFromSavings struct {
	From      string `json:"from"`
	RequestId int    `json:"request_id"`
}

func (o CancelTransferFromSavings) OpName() string {
	return "cancel_transfer_from_savings"
}

type TransferToVesting struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

func (o TransferToVesting) OpName() string {
	return "transfer_to_vesting"
}

type Auths struct {
	WeightThreshold int              `json:"weight_threshold"`
	AccountAuths    [][2]interface{} `json:"account_auths"` // tuple (string, int)
	KeyAuths        [][2]interface{} `json:"key_auths"`     // tuple (string, int)
}

type AccountCreateOperation struct {
	Fee            string `json:"fee"`
	Creator        string `json:"creator"`
	NewAccountName string `json:"new_account_name"`
	Owner          Auths  `json:"owner"`
	Active         Auths  `json:"active"`
	Posting        Auths  `json:"posting"`
	MemoKey        string `json:"memo_key"`
	JsonMetadata   string `json:"json_metadata"`
}

func (o AccountCreateOperation) OpName() string {
	return "account_create"
}

// ref: https://developers.hive.io/apidefinitions/#broadcast_ops_account_update
type AccountUpdateOperation struct {
	Account string `json:"account"`

	// optional: auths
	Owner   *Auths `json:"owner"`
	Active  *Auths `json:"active"`
	Posting *Auths `json:"posting"`

	MemoKey      string `json:"memo_key"`
	JsonMetadata string `json:"json_metadata"`

	// special (not serialized, used to determine operation ID number)
	opText string
}

func (o AccountUpdateOperation) OpName() string {
	return "account_update"
}

// Broadcast Account update operation
func (h *HiveRpcNode) UpdateAccount(
	account string,
	owner *Auths,
	active *Auths,
	posting *Auths,
	jsonMetadata string,
	memoKey string,
	wif *string,
) (string, error) {

	if owner != nil || active != nil || posting != nil {
		return "", fmt.Errorf("owner, active, posting are not supported or tested yet")
	}

	op := AccountUpdateOperation{
		Account:      account,
		Owner:        owner,
		Active:       active,
		Posting:      posting,
		MemoKey:      memoKey,
		JsonMetadata: jsonMetadata,
	}

	return h.Broadcast([]HiveOperation{op}, wif)
}

type CustomJsonOperation struct {
	RequiredAuths        []string `json:"required_auths"`
	RequiredPostingAuths []string `json:"required_posting_auths"`
	Id                   string   `json:"id"`
	Json                 string   `json:"json"`
	opText               string
}

func (o CustomJsonOperation) OpName() string {
	return "custom_json"
}

func (h *HiveRpcNode) BroadcastJson(reqAuth []string, reqPostAuth []string, id string, cj string, wif *string) (string, error) {
	op := CustomJsonOperation{reqAuth, reqPostAuth, id, cj, "custom_json"}
	return h.Broadcast([]HiveOperation{op}, wif)
}

type ClaimRewardOperation struct {
	Account     string `json:"account"`
	RewardHBD   string `json:"reward_hbd"`
	RewardHIVE  string `json:"reward_hive"`
	RewardVests string `json:"reward_vests"`
	opText      string
}

func (o ClaimRewardOperation) OpName() string {
	return o.opText
}

type ClaimAccountOperation struct {
	Fee     string `json:"fee"`
	Creator string `json:"creator"`
}

func (o ClaimAccountOperation) OpName() string {
	return "claim_account"
}

func (h *HiveRpcNode) ClaimRewards(Account string, wif *string) (string, error) {
	accountData, err := h.GetAccount([]string{Account})

	if err != nil {
		return "", err
	}

	for _, accounts := range accountData {
		claim := ClaimRewardOperation{Account, accounts.RewardHbdBalance, accounts.RewardHiveBalance, accounts.RewardVestingBalance, "claim_reward_balance"}
		broadcast, err := h.Broadcast([]HiveOperation{claim}, wif)
		return broadcast, err
	}

	return "", nil

}

// WithdrawVesting (op 4): Power down — converts VESTS back to HIVE over 13 weeks
// ref: https://developers.hive.io/apidefinitions/#broadcast_ops_withdraw_vesting
type WithdrawVestingOperation struct {
	Account       string `json:"account"`
	VestingShares string `json:"vesting_shares"`
}

func (o WithdrawVestingOperation) OpName() string {
	return "withdraw_vesting"
}

// DelegateVestingShares (op 40): Delegate HP to another account for curation
// Set VestingShares to "0.000000 VESTS" to undelegate
// ref: https://developers.hive.io/apidefinitions/#broadcast_ops_delegate_vesting_shares
type DelegateVestingSharesOperation struct {
	Delegator     string `json:"delegator"`
	Delegatee     string `json:"delegatee"`
	VestingShares string `json:"vesting_shares"`
}

func (o DelegateVestingSharesOperation) OpName() string {
	return "delegate_vesting_shares"
}

// AccountWitnessProxy (op 13): Set a proxy for witness and DHF voting
// Set Proxy to "" to clear the proxy
// ref: https://developers.hive.io/apidefinitions/#broadcast_ops_account_witness_proxy
type AccountWitnessProxyOperation struct {
	Account string `json:"account"`
	Proxy   string `json:"proxy"`
}

func (o AccountWitnessProxyOperation) OpName() string {
	return "account_witness_proxy"
}

// CreateClaimedAccount (op 23): Create account using a previously claimed account token
// Same as AccountCreateOperation but fee is zero (uses claim token instead)
// ref: https://developers.hive.io/apidefinitions/#broadcast_ops_create_claimed_account
type CreateClaimedAccountOperation struct {
	Creator        string        `json:"creator"`
	NewAccountName string        `json:"new_account_name"`
	Owner          Auths         `json:"owner"`
	Active         Auths         `json:"active"`
	Posting        Auths         `json:"posting"`
	MemoKey        string        `json:"memo_key"`
	JsonMetadata   string        `json:"json_metadata"`
	Extensions     []interface{} `json:"extensions"`
}

func (o CreateClaimedAccountOperation) OpName() string {
	return "create_claimed_account"
}

type TransferOperation struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Memo   string `json:"memo"`
}

func (o TransferOperation) OpName() string {
	return "transfer"
}

func (h *HiveRpcNode) Transfer(from string, to string, amount string, memo string, wif *string) (string, error) {
	transfer := TransferOperation{from, to, amount, memo}

	return h.Broadcast([]HiveOperation{transfer}, wif)
}

func getHiveChainId() []byte {
	cid, _ := hex.DecodeString("beeab0de00000000000000000000000000000000000000000000000000000000")
	return cid
}

func getHiveOpId(op string) uint64 {
	op = op + "_operation"
	hiveOpsIds := getHiveOpIds()
	return hiveOpsIds[op]
}

func getHiveOpIds() map[string]uint64 {
	hiveOpsIds := make(map[string]uint64)
	hiveOpsIds["vote_operation"] = 0
	hiveOpsIds["comment_operation"] = 1

	hiveOpsIds["transfer_operation"] = 2
	hiveOpsIds["transfer_to_vesting_operation"] = 3
	hiveOpsIds["withdraw_vesting_operation"] = 4

	hiveOpsIds["limit_order_create_operation"] = 5
	hiveOpsIds["limit_order_cancel_operation"] = 6

	hiveOpsIds["feed_publish_operation"] = 7
	hiveOpsIds["convert_operation"] = 8

	hiveOpsIds["account_create_operation"] = 9
	hiveOpsIds["account_update_operation"] = 10

	hiveOpsIds["witness_update_operation"] = 11
	hiveOpsIds["account_witness_vote_operation"] = 12
	hiveOpsIds["account_witness_proxy_operation"] = 13

	hiveOpsIds["pow_operation"] = 14

	hiveOpsIds["custom_operation"] = 15

	hiveOpsIds["report_over_production_operation"] = 16

	hiveOpsIds["delete_comment_operation"] = 17
	hiveOpsIds["custom_json_operation"] = 18
	hiveOpsIds["comment_options_operation"] = 19
	hiveOpsIds["set_withdraw_vesting_route_operation"] = 20
	hiveOpsIds["limit_order_create2_operation"] = 21
	hiveOpsIds["claim_account_operation"] = 22
	hiveOpsIds["create_claimed_account_operation"] = 23
	hiveOpsIds["request_account_recovery_operation"] = 24
	hiveOpsIds["recover_account_operation"] = 25
	hiveOpsIds["change_recovery_account_operation"] = 26
	hiveOpsIds["escrow_transfer_operation"] = 27
	hiveOpsIds["escrow_dispute_operation"] = 28
	hiveOpsIds["escrow_release_operation"] = 29
	hiveOpsIds["pow2_operation"] = 30
	hiveOpsIds["escrow_approve_operation"] = 31
	hiveOpsIds["transfer_to_savings_operation"] = 32
	hiveOpsIds["transfer_from_savings_operation"] = 33
	hiveOpsIds["cancel_transfer_from_savings_operation"] = 34
	hiveOpsIds["custom_binary_operation"] = 35
	hiveOpsIds["decline_voting_rights_operation"] = 36
	hiveOpsIds["reset_account_operation"] = 37
	hiveOpsIds["set_reset_account_operation"] = 38
	hiveOpsIds["claim_reward_balance_operation"] = 39
	hiveOpsIds["delegate_vesting_shares_operation"] = 40
	hiveOpsIds["account_create_with_delegation_operation"] = 41
	hiveOpsIds["witness_set_properties_operation"] = 42
	hiveOpsIds["account_update2_operation"] = 43
	hiveOpsIds["create_proposal_operation"] = 44
	hiveOpsIds["update_proposal_votes_operation"] = 45
	hiveOpsIds["remove_proposal_operation"] = 46
	hiveOpsIds["update_proposal_operation"] = 47
	hiveOpsIds["collateralized_convert_operation"] = 48
	hiveOpsIds["recurrent_transfer_operation"] = 49

	return hiveOpsIds
}

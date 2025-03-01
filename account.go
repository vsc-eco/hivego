package hivego

import (
	"time"

	"github.com/vsc-eco/hivego/utils"
)

type CustomTime time.Time

const customTimeLayout = "2006-01-02T15:04:05"

type Authority struct {
	AccountAuths    [][]interface{} `json:"account_auths"`
	KeyAuths        [][]interface{} `json:"key_auths"`
	WeightThreshold int             `json:"weight_threshold"`
}

type RC struct {
	CurrentMana    int64 `json:"current_mana"`
	LastUpdateTime int64 `json:"last_update_time"`
}

type AccountData struct {
	ID                            int64         `json:"id"`
	Name                          string        `json:"name"`
	Owner                         Authority     `json:"owner"`
	Active                        Authority     `json:"active"`
	Posting                       Authority     `json:"posting"`
	MemoKey                       string        `json:"memo_key"`
	JSONMetadata                  string        `json:"json_metadata"`
	Proxy                         string        `json:"proxy"`
	LastOwnerUpdate               CustomTime    `json:"last_owner_update"`
	LastAccountUpdate             CustomTime    `json:"last_account_update"`
	Created                       CustomTime    `json:"created"`
	Mined                         bool          `json:"mined"`
	RecoveryAccount               string        `json:"recovery_account"`
	LastAccountRecovery           CustomTime    `json:"last_account_recovery"`
	CommentCount                  int32         `json:"comment_count"`
	LifetimeVoteCount             int32         `json:"lifetime_vote_count"`
	PostCount                     int32         `json:"post_count"`
	CanVote                       bool          `json:"can_vote"`
	VotingPower                   int16         `json:"voting_power"`
	LastVoteTime                  CustomTime    `json:"last_vote_time"`
	Balance                       string        `json:"balance"`
	SavingsBalance                string        `json:"savings_balance"`
	HbdBalance                    string        `json:"hbd_balance"`
	HbdSeconds                    string        `json:"hbd_seconds"`
	HbdSecondsLastUpdate          CustomTime    `json:"hbd_seconds_last_update"`
	HbdLastInterestPayment        CustomTime    `json:"hbd_last_interest_payment"`
	SavingsHbdBalance             string        `json:"savings_hbd_balance"`
	SavingsHbdSeconds             string        `json:"savings_hbd_seconds"`
	SavingsHbdLastUpdate          CustomTime    `json:"savings_hbd_last_update"`
	SavingsHbdLastInterestPayment CustomTime    `json:"savings_hbd_last_interest_payment"`
	SavingsWithdrawRequests       int32         `json:"savings_withdraw_requests"`
	RewardHbdBalance              string        `json:"reward_hbd_balance"`
	RewardHiveBalance             string        `json:"reward_hive_balance"`
	RewardVestingBalance          string        `json:"reward_vesting_balance"`
	RewardVestingHive             string        `json:"reward_vesting_hive"`
	VestingShares                 string        `json:"vesting_shares"`
	DelegatedVestingShares        string        `json:"delegated_vesting_shares"`
	ReceivedVestingShares         string        `json:"received_vesting_shares"`
	VestingWithdrawRate           string        `json:"vesting_withdraw_rate"`
	NextVestingWithdrawal         CustomTime    `json:"next_vesting_withdrawal"`
	Withdrawn                     int64         `json:"withdrawn"`
	ToWithdraw                    int64         `json:"to_withdraw"`
	WithdrawRoutes                int32         `json:"withdraw_routes"`
	CurationRewards               int64         `json:"curation_rewards"`
	PostingRewards                int64         `json:"posting_rewards"`
	ProxiedVsfVotes               []int64       `json:"proxied_vsf_votes"`
	WitnessesVotedFor             int32         `json:"witnesses_voted_for"`
	LastPost                      CustomTime    `json:"last_post"`
	LastRootPost                  CustomTime    `json:"last_root_post"`
	AverageBandwidth              string        `json:"average_bandwidth"`
	LifetimeBandwidth             string        `json:"lifetime_bandwidth"`
	LastBandwidthUpdate           CustomTime    `json:"last_bandwidth_update"`
	PostVotingPower               string        `json:"post_voting_power"`
	Reputation                    int64         `json:"reputation"`
	PostBandwidth                 int64         `json:"post_bandwidth"`
	PendingClaimedAccounts        int32         `json:"pending_claimed_accounts"`
	PendingTransfers              int32         `json:"pending_transfers"`
	PreviousOwnerUpdate           CustomTime    `json:"previous_owner_update"`
	TransferHistory               []interface{} `json:"transfer_history"`
	MarketHistory                 []interface{} `json:"market_history"`
	PostHistory                   []interface{} `json:"post_history"`
	VoteHistory                   []interface{} `json:"vote_history"`
	OtherHistory                  []interface{} `json:"other_history"`
	WitnessVotes                  []string      `json:"witness_votes"`
	TagsUsage                     []interface{} `json:"tags_usage"`
	GuestBloggers                 []interface{} `json:"guest_bloggers"`
	DelayedVotes                  []interface{} `json:"delayed_votes"`
	VotingManabar                 RC            `json:"voting_manabar"`
	DownvoteManabar               RC            `json:"downvote_manabar"`
	GovernanceVoteExpirationTs    CustomTime    `json:"governance_vote_expiration_ts"`
	PostingJSONMetadata           string        `json:"posting_json_metadata"`
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = str[1 : len(str)-1]
	t, err := time.Parse(customTimeLayout, str)
	if err != nil {
		return err
	}
	*ct = CustomTime(t)
	return nil
}

func (ct CustomTime) ToTime() time.Time {
	return time.Time(ct)
}

func (h *HiveRpcNode) GetAccount(accountNames []string) ([]AccountData, error) {
	params := [][]string{accountNames}
	var query = hrpcQuery{
		method: CondenserApiGetAccounts,
		params: params,
	}

	res, err := h.CallRaw(query)
	if err != nil {
		return nil, err
	}
	var accountData []AccountData
	err = utils.Recast(res.Result, &accountData)

	if err != nil {
		return nil, err
	}
	return accountData, nil
}

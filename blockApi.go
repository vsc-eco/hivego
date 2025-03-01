package hivego

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"time"

	"github.com/vsc-eco/hivego/utils"
)

type globalProps struct {
	HeadBlockNumber uint64 `json:"head_block_number"`
	HeadBlockId     string `json:"head_block_id"`
	Time            string `json:"time"`
}

type getBlockRangeQueryParams struct {
	StartingBlockNum uint64 `json:"starting_block_num"`
	Count            uint64 `json:"count"`
}

type getBlockQueryParams struct {
	BlockNum uint64 `json:"block_num"`
}

type getVirtualOpsQueryParams struct {
	BlockNum          uint64 `json:"block_num"`
	OnlyVirtual       bool   `json:"only_virtual"`
	IncludeReversible bool   `json:"include_reversible"`
}

const (
	failureWaitTime = 2500 * time.Millisecond
	retryWaitTime   = 1000 * time.Millisecond
)

type Block struct {
	BlockNumber           uint64
	BlockID               string        `json:"block_id"`
	Previous              string        `json:"previous"`
	Timestamp             string        `json:"timestamp"`
	Witness               string        `json:"witness"`
	TransactionMerkleRoot string        `json:"transaction_merkle_root"`
	Transactions          []Transaction `json:"transactions"`
	Extensions            []interface{} `json:"extensions"`
	SigningKey            string        `json:"signing_key"`
	TransactionIds        []string      `json:"transaction_ids"`
	WitnessSignature      string        `json:"witness_signature"`
}
type Transaction struct {
	Expiration           string        `json:"expiration"`
	Extensions           []interface{} `json:"extensions"`
	Operations           []Operation   `json:"operations"`
	RefBlockNum          uint16        `json:"ref_block_num"`
	RefBlockPrefix       uint32        `json:"ref_block_prefix"`
	Signatures           []string      `json:"signatures"`
	RequiredAuths        []string      `json:"required_auths,omitempty"`
	RequiredPostingAuths []string      `json:"required_posting_auths,omitempty"`
}
type Operation struct {
	Type  string                 `json:"type"`
	Value map[string]interface{} `json:"value"`
}

type VirtualOp struct {
	Block uint64 `json:"block"`
	Op    struct {
		Type  string                 `json:"type"`
		Value map[string]interface{} `json:"value"`
	} `json:"op"`
	OpInTrx     int    `json:"op_in_trx"`
	TrxId       string `json:"trx_id"`
	TrxInBlock  int    `json:"trx_in_block"`
	VirtualOp   bool   `json:"virtual_op"`
	OperationId int    `json:"operation_id"`
	Timestamp   string `json:"timestamp"`
}

type operationTypes struct {
	Vote                        string
	Comment                     string
	Transfer                    string
	TransferToVesting           string
	WithdrawVesting             string
	LimitOrderCreate            string
	LimitOrderCancel            string
	FeedPublish                 string
	Convert                     string
	AccountCreate               string
	AccountUpdate               string
	WitnessUpdate               string
	AccountWitnessVote          string
	AccountWitnessProxy         string
	Pow                         string
	Custom                      string
	ReportOverProduction        string
	DeleteComment               string
	CustomJson                  string
	CommentOptions              string
	SetWithdrawVestingRoute     string
	LimitOrderCreate2           string
	ClaimAccount                string
	CreateClaimedAccount        string
	RequestAccountRecovery      string
	RecoverAccount              string
	ChangeRecoveryAccount       string
	EscrowTransfer              string
	EscrowDispute               string
	EscrowRelease               string
	Pow2                        string
	EscrowApprove               string
	TransferToSavings           string
	TransferFromSavings         string
	CancelTransferFromSavings   string
	CustomBinary                string
	DeclineVotingRights         string
	ResetAccount                string
	SetResetAccount             string
	ClaimRewardBalance          string
	DelegateVestingShares       string
	AccountCreateWithDelegation string
	WitnessSetProperties        string
	AccountUpdate2              string
	CreateProposal              string
	UpdateProposalVotes         string
	RemoveProposal              string
	UpdateProposal              string
	CollateralizedConvert       string
	RecurrentTransfer           string
}

var OperationType = operationTypes{
	Vote:                        "vote_operation",
	Comment:                     "comment_operation",
	Transfer:                    "transfer_operation",
	TransferToVesting:           "transfer_to_vesting_operation",
	WithdrawVesting:             "withdraw_vesting_operation",
	LimitOrderCreate:            "limit_order_create_operation",
	LimitOrderCancel:            "limit_order_cancel_operation",
	FeedPublish:                 "feed_publish_operation",
	Convert:                     "convert_operation",
	AccountCreate:               "account_create_operation",
	AccountUpdate:               "account_update_operation",
	WitnessUpdate:               "witness_update_operation",
	AccountWitnessVote:          "account_witness_vote_operation",
	AccountWitnessProxy:         "account_witness_proxy_operation",
	Pow:                         "pow_operation",
	Custom:                      "custom_operation",
	ReportOverProduction:        "report_over_production_operation",
	DeleteComment:               "delete_comment_operation",
	CustomJson:                  "custom_json_operation",
	CommentOptions:              "comment_options_operation",
	SetWithdrawVestingRoute:     "set_withdraw_vesting_route_operation",
	LimitOrderCreate2:           "limit_order_create2_operation",
	ClaimAccount:                "claim_account_operation",
	CreateClaimedAccount:        "create_claimed_account_operation",
	RequestAccountRecovery:      "request_account_recovery_operation",
	RecoverAccount:              "recover_account_operation",
	ChangeRecoveryAccount:       "change_recovery_account_operation",
	EscrowTransfer:              "escrow_transfer_operation",
	EscrowDispute:               "escrow_dispute_operation",
	EscrowRelease:               "escrow_release_operation",
	Pow2:                        "pow2_operation",
	EscrowApprove:               "escrow_approve_operation",
	TransferToSavings:           "transfer_to_savings_operation",
	TransferFromSavings:         "transfer_from_savings_operation",
	CancelTransferFromSavings:   "cancel_transfer_from_savings_operation",
	CustomBinary:                "custom_binary_operation",
	DeclineVotingRights:         "decline_voting_rights_operation",
	ResetAccount:                "reset_account_operation",
	SetResetAccount:             "set_reset_account_operation",
	ClaimRewardBalance:          "claim_reward_balance_operation",
	DelegateVestingShares:       "delegate_vesting_shares_operation",
	AccountCreateWithDelegation: "account_create_with_delegation_operation",
	WitnessSetProperties:        "witness_set_properties_operation",
	AccountUpdate2:              "account_update2_operation",
	CreateProposal:              "create_proposal_operation",
	UpdateProposalVotes:         "update_proposal_votes_operation",
	RemoveProposal:              "remove_proposal_operation",
	UpdateProposal:              "update_proposal_operation",
	CollateralizedConvert:       "collateralized_convert_operation",
	RecurrentTransfer:           "recurrent_transfer_operation",
}

func (h *HiveRpcNode) GetGlobalProps() (globalProps, error) {
	var props globalProps
	r, err := h.GetDynamicGlobalProps()
	if err != nil {
		return props, err
	}
	err = utils.Recast(r.Result, &props)
	return props, err
}

func (h *HiveRpcNode) GetDynamicGlobalProps() (*utils.RPCResponse, error) {
	q := hrpcQuery{method: CondenserApiGetDynamicGlobalProperties, params: []string{}}
	return h.CallRaw(q)
}

func (h *HiveRpcNode) GetBlockRange(startBlock uint64, count uint64) ([]Block, error) {
	return h.fetchBlockInRange(startBlock, count)
}

func (h *HiveRpcNode) GetBlock(blockNum uint64) (Block, error) {
	blocks, err := h.fetchBlock([]getBlockQueryParams{{BlockNum: blockNum}})
	if err != nil || len(blocks) == 0 {
		return Block{}, err
	}
	return blocks[0], nil
}

func (h *HiveRpcNode) StreamBlocks() (<-chan Block, error) {
	blockChan := make(chan Block)

	go func() {

		res, err := h.GetDynamicGlobalProps()

		if err != nil {
			log.Fatalf("Failed to fetch dynamic global properties: %v", err)
			close(blockChan)
			return
		}

		var props globalProps
		err = utils.Recast(res.Result, &props)
		if err != nil {
			log.Fatalf("Failed to unmarshal dynamic global properties: %v", err)
			close(blockChan)
			return
		}

		currentBlock := props.HeadBlockNumber

		for {
			blockData, err := h.GetBlock(currentBlock)

			if err != nil || blockData.BlockID == "" {
				log.Printf("Error fetching block %d: %v\n. Retrying in 3 seconds...", currentBlock, err)
				time.Sleep(failureWaitTime)
				continue
			}

			blockChan <- blockData
			currentBlock++
			time.Sleep(retryWaitTime)
		}
	}()
	// Start a separate goroutine to log the contents of blockChan
	go func() {
		for block := range blockChan {
			log.Printf("Received block: %+v", block.BlockNumber) // Log the received block
		}
	}()

	return blockChan, nil
}

func (h *HiveRpcNode) FetchVirtualOps(blockHeight uint64, onlyVirtual bool, IncludeReversible bool) ([]VirtualOp, error) {
	params := getVirtualOpsQueryParams{BlockNum: blockHeight, OnlyVirtual: IncludeReversible, IncludeReversible: IncludeReversible}
	query := hrpcQuery{method: CondenserApiGetOpsInBlock, params: params}
	queries := []hrpcQuery{query}

	res, err := h.CallBatchRaw(queries)

	if err != nil {
		return nil, err
	}

	var virtualOpResponses []struct {
		ID      int    `json:"id"`
		JsonRPC string `json:"jsonrpc"`
		Result  struct {
			Ops []struct {
				Op struct {
					Type  string                 `json:"type"`
					Value map[string]interface{} `json:"value"`
					// Value interface{} `json:"value"`
				} `json:"op"`
				Block       uint64 `json:"block"`
				OpInTrx     int    `json:"op_in_trx"`
				TrxId       string `json:"trx_id"`
				TrxInBlock  int    `json:"trx_in_block"`
				VirtualOp   bool   `json:"virtual_op"`
				OperationId int    `json:"operation_id"`
				Timestamp   string `json:"timestamp"`
			} `json:"ops"`
		} `json:"result"`
	}

	//err = json.Unmarshal(res[0], &virtualOpResponses)
	err = utils.Recast(res, &virtualOpResponses)
	if err != nil {
		return nil, err
	}

	var virtualOps []VirtualOp
	for _, virtualOpResponse := range virtualOpResponses {
		for _, op := range virtualOpResponse.Result.Ops {
			virtualOps = append(virtualOps, VirtualOp{
				Block: op.Block,
				Op: struct {
					Type  string                 `json:"type"`
					Value map[string]interface{} `json:"value"`
				}{
					Type:  op.Op.Type,
					Value: op.Op.Value,
				},
				OpInTrx:     op.OpInTrx,
				TrxId:       op.TrxId,
				TrxInBlock:  op.TrxInBlock,
				VirtualOp:   op.VirtualOp,
				OperationId: op.OperationId,
				Timestamp:   op.Timestamp,
			})
		}
	}

	return virtualOps, nil
}

func (h *HiveRpcNode) fetchBlockInRange(startBlock, count uint64) ([]Block, error) {
	params := getBlockRangeQueryParams{StartingBlockNum: startBlock, Count: count}
	query := hrpcQuery{method: BlockApiGetBlockRange, params: params}
	queries := []hrpcQuery{query}

	res, err := h.CallBatchRaw(queries)
	if err != nil {
		return nil, err
	}

	var blockRangeResponses []struct {
		ID      int    `json:"id"`
		JsonRPC string `json:"jsonrpc"`
		Result  struct {
			Blocks []Block `json:"blocks"`
		} `json:"result"`
	}

	err = utils.Recast(res, &blockRangeResponses)
	if err != nil {
		return nil, err
	}

	var blocks []Block
	for _, blockRangeResponse := range blockRangeResponses {
		blocks = append(blocks, blockRangeResponse.Result.Blocks...)
	}

	return processedBlocks(blocks), nil
}

func (h *HiveRpcNode) fetchBlock(params []getBlockQueryParams) ([]Block, error) {
	var queries []hrpcQuery
	for _, param := range params {
		query := hrpcQuery{method: BlockApiGetBlock, params: param}
		queries = append(queries, query)
	}

	res, err := h.CallBatchRaw(queries)
	if err != nil {
		return nil, err
	}

	var blockResponses []struct {
		ID      int    `json:"id"`
		JsonRPC string `json:"jsonrpc"`
		Result  struct {
			Block Block `json:"block"`
		} `json:"result"`
	}

	err = utils.Recast(res, &blockResponses)

	if err != nil {
		return nil, err
	}

	var blocks []Block
	for _, blockResponse := range blockResponses {
		blocks = append(blocks, blockResponse.Result.Block)
	}

	return processedBlocks(blocks), nil
}

func processedBlocks(blocks []Block) []Block {
	var processedBlocks []Block

	for _, block := range blocks {

		if block.BlockID != "" {
			blockInt, _ := hex.DecodeString(block.BlockID[0:8])

			// Ensure blockInt has 8 bytes, pad with zeros if necessary
			if len(blockInt) < 8 {
				paddedBlockInt := make([]byte, 8)
				copy(paddedBlockInt[8-len(blockInt):], blockInt)
				block.BlockNumber = binary.BigEndian.Uint64(paddedBlockInt)
			} else {
				block.BlockNumber = binary.BigEndian.Uint64(blockInt)
			}

			processedBlocks = append(processedBlocks, block)
		}
	}
	return processedBlocks
}

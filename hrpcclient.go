package hivego

import (
	"errors"
	"log"
	"sync/atomic"

	"github.com/cfoxon/jsonrpc2client"
)

const precisionFactor = 10000 // basis points for precision on rolling average

type NodeStats struct {
	successCount atomic.Uint32
	failureCount atomic.Uint32
	rollingAvg   atomic.Uint32
}

type HiveRpcNode struct {
	addresses    []string
	currentIndex atomic.Uint32
	nodeStats    []NodeStats
	MaxConn      int
	MaxBatch     int
	NoBroadcast  bool
	ChainID      string
}

type globalProps struct {
	HeadBlockNumber int    `json:"head_block_number"`
	HeadBlockId     string `json:"head_block_id"`
	Time            string `json:"time"`
}

type hrpcQuery struct {
	method string
	params interface{}
}

func NewHiveRpc(addrs []string) *HiveRpcNode {
	return NewHiveRpcWithOpts(addrs, 1, 1)
}

func NewHiveRpcWithOpts(addrs []string, maxConn int, maxBatch int) *HiveRpcNode {
	nodeStats := make([]NodeStats, len(addrs))
	return &HiveRpcNode{
		addresses:    addrs,
		currentIndex: atomic.Uint32{},
		nodeStats:    nodeStats,
		MaxConn:      maxConn,
		MaxBatch:     maxBatch,
	}
}

func (h *HiveRpcNode) GetDynamicGlobalProps() ([]byte, error) {
	q := hrpcQuery{method: "condenser_api.get_dynamic_global_properties", params: []string{}}
	res, err := h.rpcExec(q)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *HiveRpcNode) rpcExec(query hrpcQuery) ([]byte, error) {
	numNodes := uint32(len(h.addresses))
	var lastError error

	for i := uint32(0); i < numNodes; i++ {
		index := h.currentIndex.Add(1) % numNodes
		endpoint := h.addresses[index]

		rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)
		jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: 1, Params: query.params}
		resp, err := rpcClient.CallRaw(jr2query)
		if err != nil {
			if enableLogging {
				log.Printf(
					"rpcExec failed for endpoint %s (index %d), method %s: %v",
					endpoint,
					index,
					query.method,
					err,
				)
			}
			h.nodeStats[index].failureCount.Add(1)
			h.updateRollingAvg(index)
			if enableLogging {
				h.logFailureCounts()
				nextIndex := (h.currentIndex.Load() + i + 1) % uint32(numNodes)
				h.logSwitchingNode(index, nextIndex, numNodes)
			}
			lastError = err
			continue
		}

		if resp.Error != nil {
			if enableLogging {
				log.Printf(
					"rpcExec received error response from endpoint %s (index %d), method %s: %v",
					endpoint,
					index,
					query.method,
					resp.Error,
				)
			}
			h.nodeStats[index].failureCount.Add(1)
			h.updateRollingAvg(index)
			if enableLogging {
				h.logFailureCounts()
				nextIndex := (h.currentIndex.Load() + i + 1) % numNodes
				h.logSwitchingNode(index, nextIndex, numNodes)
			}
			lastError = errors.New(resp.Error.Message)
			continue
		}

		// Check for bad data: if result is empty, consider it bad
		if len(resp.Result) == 0 {
			if enableLogging {
				log.Printf(
					"rpcExec received empty result from endpoint %s (index %d), method %s",
					endpoint,
					index,
					query.method,
				)
			}
			h.nodeStats[index].failureCount.Add(1)
			h.updateRollingAvg(index)
			if enableLogging {
				h.logFailureCounts()
				nextIndex := (h.currentIndex.Load() + i + 1) % numNodes
				h.logSwitchingNode(index, nextIndex, numNodes)
			}
			lastError = errors.New("empty result received from node")
			continue
		}

		// Success
		h.nodeStats[index].successCount.Add(1)
		h.updateRollingAvg(index)
		h.currentIndex.Store(index) // Set to last successful node
		return resp.Result, nil
	}

	if lastError != nil {
		return nil, lastError
	}
	return nil, errors.New("all API nodes failed")
}

func (h *HiveRpcNode) updateRollingAvg(index uint32) {
	successes := h.nodeStats[index].successCount.Load()
	total := successes + h.nodeStats[index].failureCount.Load()
	if total > 0 {
		h.nodeStats[index].rollingAvg.Store(uint32(float64(successes) / float64(total) * precisionFactor))
	}
}

func (h *HiveRpcNode) logFailureCounts() {
	log.Printf("DEBUG: API Node Failure Counts:")
	log.Printf("| Node | failureCount |")
	for i, addr := range h.addresses {
		log.Printf("| %s | %d |", addr, h.nodeStats[i].failureCount.Load())
	}
}

func (h *HiveRpcNode) logSwitchingNode(currentIndex uint32, nextIndex uint32, numNodes uint32) {
	nextEndpoint := h.addresses[nextIndex]
	log.Printf("DEBUG: Switching to node: %s (index %d)", nextEndpoint, nextIndex)
}

func (h *HiveRpcNode) rpcExecBatchFast(queries []hrpcQuery) ([][]byte, error) {
	numNodes := uint32(len(h.addresses))
	var lastError error

	for i := uint32(0); i < numNodes; i++ {
		index := (h.currentIndex.Load() + i) % numNodes
		endpoint := h.addresses[index]

		rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)

		var jr2queries jsonrpc2client.RPCRequests
		for j, query := range queries {
			jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: j, Params: query.params}
			jr2queries = append(jr2queries, jr2query)
		}

		resps, err := rpcClient.CallBatchFast(jr2queries)
		if err != nil {
			if enableLogging {
				log.Printf("rpcExecBatchFast failed for endpoint %s (index %d): %v", endpoint, index, err)
			}
			h.nodeStats[index].failureCount.Add(1)
			h.updateRollingAvg(index)
			if enableLogging {
				h.logFailureCounts()
				nextIndex := (h.currentIndex.Load() + i + 1) % numNodes
				h.logSwitchingNode(index, nextIndex, numNodes)
			}
			lastError = err
			continue
		}

		// Check if any response has error or bad data
		hasError := false
		for _, respBytes := range resps {
			if len(respBytes) == 0 {
				hasError = true
				break
			}
		}
		if hasError {
			if enableLogging {
				log.Printf("rpcExecBatchFast received empty response(s) from endpoint %s (index %d)", endpoint, index)
			}
			h.nodeStats[index].failureCount.Add(1)
			h.updateRollingAvg(index)
			if enableLogging {
				h.logFailureCounts()
				nextIndex := (h.currentIndex.Load() + i + 1) % numNodes
				h.logSwitchingNode(index, nextIndex, numNodes)
			}
			lastError = errors.New("empty response(s) received from node")
			continue
		}

		// Success
		h.nodeStats[index].successCount.Add(1)
		h.updateRollingAvg(index)
		h.currentIndex.Store(index)

		var batchResult [][]byte
		batchResult = append(batchResult, resps...)
		return batchResult, nil
	}

	if lastError != nil {
		return nil, lastError
	}
	return nil, errors.New("all API nodes failed")
}

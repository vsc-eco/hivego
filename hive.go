package hivego

import (
	"context"
	"log"

	"github.com/vsc-eco/hivego/utils"
)

const USE_RANDOM_URLS = "random"

// this hive client is may use loadbalancing over public api nodes,
// just set url=USE_RANDOM_URLS
type HiveRpcNode struct {
	id          int
	url         string
	opts        *utils.RPCClientOpts
	rpc         utils.RPCClient
	NoBroadcast bool
}

type hrpcQuery struct {
	method ApiMethod
	params interface{}
}

func NewHiveClient(id int, url string) *HiveRpcNode {

	h := &HiveRpcNode{
		id:   id,
		opts: nil,
		url:  url,
	}

	return h
}

func NewHiveClientWithOps(id int, url string, opts *utils.RPCClientOpts) *HiveRpcNode {
	h := NewHiveClient(id, url)
	h.opts = opts
	return h
}

// with this method if url set to USE_RANDOM_URLS the rpc get random url everytime callraw or callbatchraw is called
// in this way the client loadbalance over public client node urls
func (h *HiveRpcNode) setRpc() {
	//set random url from public node urls
	if h.url == USE_RANDOM_URLS {
		h.url = utils.GetRandomApiUrlFromPublicNode()
	}
	if h.opts != nil {
		h.rpc = utils.NewClientWithOpts(h.url, h.opts)
	}
	h.rpc = utils.NewClient(h.url)

	log.Default().Println("used node url:>>>>", h.url)

}
func (h *HiveRpcNode) CallRaw(query hrpcQuery) (*utils.RPCResponse, error) {
	h.setRpc()
	request := utils.NewRequestWithID(h.id, string(query.method), query.params)
	return h.rpc.CallRaw(context.Background(), request)
}

func (h *HiveRpcNode) CallBatchRaw(queries []hrpcQuery) (utils.RPCResponses, error) {
	h.setRpc()
	var requests utils.RPCRequests
	for i, query := range queries {
		request := utils.NewRequestWithID(i, string(query.method), query.params)
		requests = append(requests, request)
	}
	return h.rpc.CallBatchRaw(context.Background(), requests)
}

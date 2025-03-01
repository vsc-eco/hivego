package main

import (
	"log"

	"github.com/vsc-eco/hivego"
	"github.com/vsc-eco/hivego/utils"
)

func main() {
	proxyURL := "http or https proxy_url_here"
	hc_proxy, _ := utils.NewHTTPClientWithProxy(proxyURL)
	opts := utils.RPCClientOpts{
		//CustomHeaders:      map[string]string{"Content-Type": "application/json"},
		HTTPClient:         hc_proxy,
		AllowUnknownFields: true,
		DefaultRequestID:   1,
	}

	hrpc := hivego.NewHiveClientWithOps(1, "https://api.hive.blog", &opts)
	//hrpc := hivego.NewHiveClient(1, hivego.USE_RANDOM_URLS)
	r, err := hrpc.GetAccount([]string{"megabytee"})
	if err != nil {
		log.Default().Println(err)
	}
	log.Default().Println(r)
}

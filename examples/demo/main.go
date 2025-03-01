package main

import (
	"log"

	"github.com/vsc-eco/hivego"
)

// keep in mind that they are some public node api urls not working
// so if you want use USE_RANDOM_URLS just set working api urls on utils/nodeUrls.go
func main() {
	hrpc := hivego.NewHiveClient(1, "https://api.hive.blog")
	//hrpc := hivego.NewHiveClient(1, hivego.USE_RANDOM_URLS)
	r, err := hrpc.GetAccount([]string{"megabytee"})
	if err != nil {
		log.Default().Println(err)
	}
	log.Default().Println(r)
}

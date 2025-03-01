package hivego

import (
	"fmt"
	"testing"
)

func TestVirtualOps(t *testing.T) {
	rpc := NewHiveClient(1, "https://api.hive.blog")
	virtualOps, err := rpc.FetchVirtualOps(88386873, true, false)

	if err != nil {
		t.Errorf("%s", err.Error())
	}
	fmt.Println(virtualOps)
}

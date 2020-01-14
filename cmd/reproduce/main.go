package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
)

func main() {
	var (
		root      = "/Users/kirill/geth/data/geth/chaindata"
		namespace = "eth/db/chaindata"
		cache     = 2048
		handles   = 5120
		freezer   = root + "/ancient"
	)
	lvlDb, _ := rawdb.NewLevelDBDatabaseWithFreezer(root, cache, handles, freezer, namespace)
	db := state.NewDatabase(lvlDb)

	_, err := state.New(common.HexToHash("0x4d2c9335d8d5e6dbd1a09f8da3444b9aed9d3090acd67ac568ad84b6d4c4be21"), db)

	if err != nil {
		panic(err)
	}
}

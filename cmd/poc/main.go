package main

import (
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"log"
)

func main() {
	fetchChainConfig()
	_ = createStateProcessor()
}

func fetchChainConfig() *params.ChainConfig {
	config := params.MainnetChainConfig

	return config
}

func createStateProcessor() *core.StateProcessor {
	//file := "/Users/kirill/Library/Ethereum/geth/chaindata"
	file := "/Users/kirill/geth/data/geth/chaindata"

	db, _ := rawdb.NewLevelDBDatabase(file, 128, 128, "eth/db/chaindata/")
	chainConfig := fetchChainConfig()

	bc, err := core.NewBlockChain(db, nil, chainConfig, ethash.NewFaker(), vm.Config{}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return core.NewStateProcessor(chainConfig, bc, ethash.NewFaker())
}

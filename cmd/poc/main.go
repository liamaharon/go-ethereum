package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"log"
)

func main() {
	chainConfig := fetchChainConfig()
	db := createDb()
	bc := createBlockchain(db, chainConfig)
	createStateProcessor(chainConfig, bc)

	tryToCreateStateDb(db)
}

func fetchChainConfig() *params.ChainConfig {
	config := params.MainnetChainConfig

	return config
}

func createDb() ethdb.Database {
	//file := "/Users/kirill/geth/data/geth/chaindata"
	var (
		root      = "/Users/kirill/geth/data/geth/chaindata"
		namespace = "eth/db/chaindata"
		cache     = 2048
		handles   = 5120
		freezer   = root + "/ancient"
	)
	db, _ := rawdb.NewLevelDBDatabaseWithFreezer(root, cache, handles, freezer, namespace)
	//db, _ := rawdb.NewLevelDBDatabase(file, 128, 128, "eth/db/chaindata/")

	return db
}

func createBlockchain(db ethdb.Database, chainConfig *params.ChainConfig) *core.BlockChain {
	bc, err := core.NewBlockChain(db, nil, chainConfig, ethash.NewFaker(), vm.Config{}, nil)
	if err != nil {
		log.Fatal(err)
	}

	return bc
}

func createStateProcessor(config *params.ChainConfig, bc *core.BlockChain) *core.StateProcessor {
	return core.NewStateProcessor(config, bc, ethash.NewFaker())
}

func tryToCreateStateDb(ethDb ethdb.Database) *state.StateDB {
	db := state.NewDatabase(ethDb)
	state, err := state.New(common.HexToHash("0x6a80bf5706919c445a3ea54d0ecba776568c387c5d426762ab03fe9a172def05"), db)

	if err != nil {
		panic(err)
	}

	return state
}

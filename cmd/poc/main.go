package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"math/big"
)

var (
	stateRoothash = common.HexToHash("0x8447453d176455bfb1f9786034dd72079d90012f82e1751e3360c20de7465de0")
	blockHash     = common.HexToHash("0xfffbc04b0c5999b33313f2eea68e996da1d631bc15db9c0b34176eef54ddc2f9")
	blockNumber   = uint64(2006668)

	testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddr   = crypto.PubkeyToAddress(testKey.PublicKey)
)

func main() {
	chainConfig := fetchChainConfig()
	db := createDb()
	bc := createBlockchain(db, chainConfig)
	processor := createStateProcessor(chainConfig, bc)

	stateDb := tryToCreateStateDb(db)
	processBlock(db, processor, stateDb)
}

func fetchChainConfig() *params.ChainConfig {
	config := params.MainnetChainConfig

	return config
}

func createDb() ethdb.Database {
	//file := "/Users/kirill/geth/data/geth/chaindata"
	var (
		root      = "/Users/kirill/geth_archive/data/geth/chaindata"
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
	state, err := state.New(stateRoothash, db)

	if err != nil {
		panic(err)
	}

	return state
}

func processBlock(db ethdb.Database, processor *core.StateProcessor, state *state.StateDB) (types.Receipts, []*types.Log, uint64) {
	block := rawdb.ReadBlock(db, blockHash, blockNumber)
	tx := types.NewTransaction(uint64(0), testAddr, big.NewInt(1000), params.TxGas, big.NewInt(1), nil)
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)

	txs := make([]*types.Transaction, 1)
	txs[0] = signedTx

	headers := make([]*types.Header, 1)
	headers[0] = block.Header()

	newBlock := block.WithBody(txs, headers)

	receipts, logs, num, err := processor.Process(newBlock, state, vm.Config{})

	if err != nil {
		panic(err)
	}

	return receipts, logs, num
}

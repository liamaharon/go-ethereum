package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/internal/ethapi"
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
	//processor := createStateProcessor(chainConfig, bc)

	stateDb := tryToCreateStateDb(db)

	message := createTransferTokensMessage()
	block := getLatestBlock(db)

	vmctx := core.NewEVMContext(message, block.Header(), bc, nil)

	result, err := traceTx(message, vmctx, stateDb)
	if err != nil {
		panic(err)
	}

	println(result)
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

func createTransferTokensMessage() types.Message {
	tx := types.NewTransaction(uint64(0), testAddr, big.NewInt(0), uint64(100000), big.NewInt(0), []byte("0xa9059cbb0000000000000000000000008b24eb4e6aae906058242d83e51fb077370c472000000000000000000000000000000000000000000000000000000000000003e8"))
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)

	if err != nil {
		panic(err)
	}

	message, err := signedTx.AsMessage(types.HomesteadSigner{})
	if err != nil {
		panic(err)
	}

	return message
}

// This func uses hardcoded block number at the moment
func getLatestBlock(db ethdb.Database) *types.Block {
	block := rawdb.ReadBlock(db, blockHash, blockNumber)

	return block
}

//func processBlock(db ethdb.Database, processor *core.StateProcessor, state *state.StateDB) (types.Receipts, []*types.Log, uint64) {
//	block := rawdb.ReadBlock(db, blockHash, blockNumber)
//	tx := types.NewTransaction(uint64(0), testAddr, big.NewInt(0), uint64(100000), big.NewInt(0), []byte("0xa9059cbb0000000000000000000000008b24eb4e6aae906058242d83e51fb077370c472000000000000000000000000000000000000000000000000000000000000003e8"))
//	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)
//
//	txs := make([]*types.Transaction, 1)
//	txs[0] = signedTx
//
//	headers := make([]*types.Header, 1)
//	headers[0] = block.Header()
//
//	newBlock := block.WithBody(txs, headers)
//
//	receipts, logs, num, err := processor.Process(newBlock, state, vm.Config{})
//
//	if err != nil {
//		panic(err)
//	}
//
//	return receipts, logs, num
//}

// This is a almost 100% copy from https://github.com/ethereum/go-ethereum/blob/master/eth/api_tracer.go#L725
// transplanted here to make It independent func not a member or `PrivateDebugAPI` class
func traceTx(message core.Message, vmctx vm.Context, statedb *state.StateDB) (*ethapi.ExecutionResult, error) {
	// Assemble the structured logger or the JavaScript tracer
	var (
		tracer vm.Tracer
		err    error
	)

	tracer = vm.NewStructLogger(&vm.LogConfig{
		DisableMemory:  false,
		DisableStack:   false,
		DisableStorage: false,
		Debug:          false,
		Limit:          0,
	})
	// Run the transaction with tracing enabled.
	vmenv := vm.NewEVM(vmctx, statedb, fetchChainConfig(), vm.Config{Debug: true, Tracer: tracer})

	ret, gas, failed, err := core.ApplyMessage(vmenv, message, new(core.GasPool).AddGas(message.Gas()))
	if err != nil {
		return nil, fmt.Errorf("tracing failed: %v", err)
	}
	// Depending on the tracer type, format and return the output
	switch tracer := tracer.(type) {
	case *vm.StructLogger:
		return &ethapi.ExecutionResult{
			Gas:         gas,
			Failed:      failed,
			ReturnValue: fmt.Sprintf("%x", ret),
			StructLogs:  ethapi.FormatLogs(tracer.StructLogs()),
		}, nil

	default:
		panic(fmt.Sprintf("bad tracer type %T", tracer))
	}
}

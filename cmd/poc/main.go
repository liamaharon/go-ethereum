package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"log"
)

var (
	stateRoothash = common.HexToHash("0xfec00c4e91c882b6e4eb3f31c3caa1802465e9e0e6cb8b8140f58aea14b0c92d")
	blockHash     = common.HexToHash("0x4a4a3daae5365a057cf4b6e06cc9baf44926aa0b783849e61a9058ee22d3d6d5")
	blockNumber   = uint64(9467878)

	//testKey, _ = crypto.HexToECDSA(<PUT YOUR HEX PK HERE>)

	DAIContractAddress = common.HexToAddress("0x6b175474e89094c44da98b954eedeac495271d0f")
	selfAddress        = common.HexToAddress("0xa520ca9a99e3da0faa656ff5c0ea0756a69be58c")
	// emulating transfer one base unit of DAI
	rawTXString = "f8a58001830186a0946b175474e89094c44da98b954eedeac495271d0f80b844a9059cbb0000000000000000000000006f7799c642ba4cc4a30cdcb538bb8a26661e358700000000000000000000000000000000000000000000000000000000000000011ca0c7e768441b300b23185f068c72a55db0abde2d552d9237fbd1dc4ba72a892754a072abc013256c98d4b518c0b56ff14ba9fe4a2e7bf9eb76de5ea2dc18754424a0"
)

func main() {
	var tx *types.Transaction
	rawtx, err := hex.DecodeString(rawTXString)
	if err != nil {
		panic(err)
	}
	rlp.DecodeBytes(rawtx, &tx)
	message, err := tx.AsMessage(types.EIP155Signer{})
	if err != nil {
		panic(err)
	}

	chainConfig := fetchMainnetChainConfig()
	db := createDb()
	bc := createBlockchain(db, chainConfig)

	stateDb := tryToCreateStateDb(db)

	block := getLatestBlock(db)
	vmctx := core.NewEVMContext(message, block.Header(), bc, nil)

	result, err := traceTx(message, vmctx, stateDb)
	if err != nil {
		panic(err)
	}

	for _, log := range result.StructLogs {
		if log.Op == "LOG3" {
			println(log.Stack)
		}
	}
}

func fetchMainnetChainConfig() *params.ChainConfig {
	return params.MainnetChainConfig
}

func createDb() ethdb.Database {
	file := "/Users/kirill/geth/geth/chaindata"
	var (
		//root      = "/Users/kirill/geth_archive/data/geth/chaindata"
		root      = file
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

//func createStateProcessor(config *params.ChainConfig, bc *core.BlockChain) *core.StateProcessor {
//	return core.NewStateProcessor(config, bc, ethash.NewFaker())
//}

func tryToCreateStateDb(ethDb ethdb.Database) *state.StateDB {
	db := state.NewDatabase(ethDb)
	state, err := state.New(stateRoothash, db)

	if err != nil {
		panic(err)
	}

	return state
}

//func createTransferTokensMessage(config *params.ChainConfig) types.Message {
//	data := common.Hex2Bytes("a9059cbb0000000000000000000000006f7799c642ba4cc4a30cdcb538bb8a26661e35870000000000000000000000000000000000000000000000000000000000000001")
//
//	tx := types.NewTransaction(uint64(0), DAIContractAddress, big.NewInt(0), uint64(100000), nil, data)
//	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, testKey)
//
//	if err != nil {
//		panic(err)
//	}
//
//	message, err := signedTx.AsMessage(types.HomesteadSigner{})
//	if err != nil {
//		panic(err)
//	}
//
//	return message
//}

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
	vmenv := vm.NewEVM(vmctx, statedb, fetchMainnetChainConfig(), vm.Config{Debug: true, Tracer: tracer})

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

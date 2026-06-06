package ingestion

import (
	"context"
	"log"
	"strings"
	"sync"

	"encoding/json"

	"github.com/lexyblazy/chainledger/config"
	"github.com/lexyblazy/chainledger/db"
	"github.com/lexyblazy/chainledger/internal/hex"
	"github.com/lexyblazy/chainledger/rpc"
)

type NetworkReader struct {
	db     *db.DB
	rpcc   *rpc.Client
	config *config.NetworkConfig
}

func NewNetworkReader(db *db.DB, config *config.NetworkConfig) *NetworkReader {
	return &NetworkReader{
		db:     db,
		rpcc:   rpc.New(config),
		config: config,
	}
}

func (w *NetworkReader) getStatus(ctx context.Context) (NetworkStatus, error) {
	var status NetworkStatus
	var err error
	status.Name = w.config.Name
	status.ChainID = int64(w.config.ChainID)
	status.StartBlock = w.config.StartBlock

	var wg sync.WaitGroup

	wg.Add(2)

	// get best rpc block number
	go func() {
		defer wg.Done()
		status.BestRpcBlockNumber, err = w.getBestRpcBlockNumber(ctx)

		if err != nil {
			log.Println(w.config.Name, "Error getting best rpc block number", err)
		}
	}()

	// get blocks count and max block number from database
	go func() {
		defer wg.Done()

		blocksQuery := "SELECT COUNT(*), MAX(block_number) FROM blocks WHERE chain_id = $1"
		tokensQuery := "SELECT COUNT(*) FROM tokens WHERE chain_id = $1"
		addressesQuery := "SELECT COUNT(*) FROM address_registry WHERE chain_id = $1"

		var blocksCount int64
		var maxBlockNumber int64

		err = w.db.Pool().QueryRow(ctx, blocksQuery, w.config.ChainID).Scan(&blocksCount, &maxBlockNumber)
		if err != nil {
			log.Println(w.config.Name, "Error getting blocks count", err)
		} else {
			status.IngestedBlocksCount = blocksCount
			status.IngestedBlocksMaxBlockNumber = maxBlockNumber
		}
		var tokensCount int64
		err = w.db.Pool().QueryRow(ctx, tokensQuery, w.config.ChainID).Scan(&tokensCount)
		if err != nil {
			log.Println(w.config.Name, "Error getting tokens count", err)
		} else {

			status.TokensCount = tokensCount
		}

		var addressesCount int
		err = w.db.Pool().QueryRow(ctx, addressesQuery, w.config.ChainID).Scan(&addressesCount)
		if err != nil {
			log.Println(w.config.Name, "Error getting addresses count", err)
		} else {
			status.AddressesCount = addressesCount
		}

	}()

	wg.Wait()

	status.IngestionLagBlocksCount = status.BestRpcBlockNumber - status.IngestedBlocksMaxBlockNumber
	status.IngestionProgressPct = float64(status.IngestedBlocksCount) / float64(status.BestRpcBlockNumber-status.StartBlock) * 100

	return status, nil

}

func (w *NetworkReader) getBestRpcBlockNumber(ctx context.Context) (int64, error) {
	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_blockNumber", []interface{}{"latest", false})
	})
	if err != nil {
		return 0, err
	}

	var blockNumber string
	err = json.Unmarshal(rpcResult, &blockNumber)

	if err != nil {
		return 0, err
	}

	val, err := hex.DecodeUint64(blockNumber)

	if err != nil {
		return 0, err
	}

	return int64(val), nil
}

func (w *NetworkReader) fetchERC20TransferLogs(ctx context.Context, blockNumber int64) ([]rpc.Log, error) {

	//  Quick note about RPC provider preferences/compatibility:
	// Some RPC providers prefer the string representation of the block number (e.g. "0x1234567890") - e.g Alchemy
	// while others prefer the integer representation (e.g. 1234567890). e.g Chainstack
	// Anything otherwise results in a client error. Provider specific documentation should be consulted.

	// this is a temporary solution to handle the different RPC provider preferences.

	var fromBlock interface{}
	var toBlock interface{}

	if isChainstackRpcProvider(w.config.RPCUrl) {
		fromBlock = blockNumber
		toBlock = blockNumber
	} else {
		fromBlock = hex.IntToHex(blockNumber)
		toBlock = hex.IntToHex(blockNumber)
	}

	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_getLogs", []interface{}{map[string]interface{}{
			"fromBlock": fromBlock,
			"toBlock":   toBlock,
			"topics":    []string{w.config.ERC20TransferTopic},
		}})
	})

	if err != nil {
		return nil, err
	}

	var logs []rpc.Log
	err = json.Unmarshal(rpcResult, &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil

}

func (w *NetworkReader) fetchTokenMetadataDecimals(ctx context.Context, tokenAddress string) (int8, error) {
	rawResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_call", []interface{}{map[string]interface{}{
			"to":    tokenAddress,
			"input": w.config.TokenMetadata.RpcCalls["decimals"],
		}, "latest"})
	})

	if err != nil {
		log.Println(w.config.Name, "Error fetching token metadata decimals", tokenAddress, err)
		return 0, err
	}

	var decimals string

	err = json.Unmarshal(rawResult, &decimals)
	if err != nil {
		return 0, err
	}

	d, err := hex.DecodeUint8(decimals)
	if err != nil {
		return 0, err
	}

	return int8(d), nil
}

func (w *NetworkReader) fetchTokenMetadataName(ctx context.Context, tokenAddress string) (string, error) {
	rawResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_call", []interface{}{map[string]interface{}{
			"to":    tokenAddress,
			"input": w.config.TokenMetadata.RpcCalls["name"],
		}, "latest"})
	})

	if err != nil {
		log.Println(w.config.Name, "Error fetching token metadata name", tokenAddress, err)
		return "", err
	}

	var name string

	err = json.Unmarshal(rawResult, &name)
	if err != nil {
		return "", err
	}

	return hex.DecodeStringOrBytes32(name)
}

func (w *NetworkReader) fetchTokenMetadataSymbol(ctx context.Context, tokenAddress string) (string, error) {
	rawResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_call", []interface{}{map[string]interface{}{
			"to":    tokenAddress,
			"input": w.config.TokenMetadata.RpcCalls["symbol"],
		}, "latest"})
	})

	if err != nil {
		log.Println(w.config.Name, "Error fetching token metadata symbol", tokenAddress, err)
		return "", err
	}

	var symbol string

	err = json.Unmarshal(rawResult, &symbol)

	if err != nil {
		return "", err
	}

	return hex.DecodeStringOrBytes32(symbol)
}

func isAlchemyRpcProvider(url string) bool {
	return strings.Contains(url, "alchemy")
}

func isChainstackRpcProvider(url string) bool {
	return strings.Contains(url, "chainstack")
}

func (w *NetworkReader) fetchBlockWithTransactions(ctx context.Context, blockNumber int64) (rpc.Block, error) {
	//  Quick note about RPC provider preferences/compatibility:
	// Some RPC providers prefer the string representation of the block number (e.g. "0x1234567890") - e.g Alchemy
	// while others prefer the integer representation (e.g. 1234567890). e.g Chainstack
	// Anything otherwise results in a client error. Provider specific documentation should be consulted.

	// this is a temporary solution to handle the different RPC provider preferences.
	var params []interface{}
	if isChainstackRpcProvider(w.config.RPCUrl) {
		params = []interface{}{blockNumber, true}
	} else {
		params = []interface{}{hex.IntToHex(blockNumber), true}
	}

	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_getBlockByNumber", params)
	})
	if err != nil {
		return rpc.Block{}, err
	}

	var block rpc.Block
	err = json.Unmarshal(rpcResult, &block)
	if err != nil {
		return rpc.Block{}, err
	}

	return block, nil
}

func (w *NetworkReader) fetchBlockReceipts(ctx context.Context, blockNumber int64) ([]rpc.TransactionReceipt, error) {
	// same rpc provider notes as fetchBlockWithTransactions

	var params []interface{}
	if isChainstackRpcProvider(w.config.RPCUrl) {
		params = []interface{}{blockNumber}
	} else {
		params = []interface{}{hex.IntToHex(blockNumber)}
	}

	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_getBlockReceipts", params)
	})

	if err != nil {
		return nil, err
	}

	var receipts []rpc.TransactionReceipt
	err = json.Unmarshal(rpcResult, &receipts)
	if err != nil {
		return nil, err
	}

	return receipts, nil
}

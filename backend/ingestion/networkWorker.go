package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"polychain.capital/config"
	"polychain.capital/db"
)

type NetworkWorker struct {
	db         *db.DB
	rpcc       *RPCClient
	config     *config.NetworkConfig
	addressSet map[string]bool
}

func (w *NetworkWorker) start(ctx context.Context) error {
	log.Println("Network worker started for", w.config.Name, "chain")
	addresses, err := w.getAllAddresses(ctx)
	if err != nil {
		return err
	}

	for _, address := range addresses {
		w.addressSet[normalizeAddress(address.Address)] = true
	}

	retryDelay := time.Duration(w.config.SyncWorker.RetryDelayMs) * time.Millisecond
	maxDelay := time.Duration(w.config.SyncWorker.MaxDelayMs) * time.Millisecond

	for {

		var (
			block                     RPCBlock
			erc20TransferLogs         []RPCLog
			fetchBlockErr             error
			fetchERC20TransferLogsErr error
			commitBlockErr            error
		)

		bn, err := w.fetchLastProcessedBlock(ctx)
		nextBlock := bn + 1
		if err != nil {
			return err
		}
		// parallelize fetching of block transactions and erc20 transfer logs
		var wg sync.WaitGroup
		wg.Add(2)

		go func(blockNumber int64) {
			defer wg.Done()
			b, err := w.fetchBlockWithTransactions(ctx, blockNumber)
			if err != nil {
				fetchBlockErr = err
				return
			}
			b.Transactions = w.filterTransactions(b.Transactions)
			block = b
		}(nextBlock)

		go func(blockNumber int64) {
			defer wg.Done()
			l, err := w.fetchERC20TransferLogs(ctx, blockNumber)
			if err != nil {
				fetchERC20TransferLogsErr = err
				return
			}
			erc20TransferLogs = w.filterERC20TransferLogs(l)
		}(nextBlock)

		wg.Wait()

		isContextCanceled := errors.Is(fetchBlockErr, context.Canceled) || errors.Is(fetchERC20TransferLogsErr, context.Canceled)
		if isContextCanceled {
			return context.Canceled
		}

		if fetchBlockErr != nil {
			log.Println("error fetching block transactions", fetchBlockErr)
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		if fetchERC20TransferLogsErr != nil {
			log.Println("error fetching erc20 transfer logs", fetchERC20TransferLogsErr)
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		commitBlockErr = w.commitBlock(ctx, block, erc20TransferLogs)

		if errors.Is(commitBlockErr, context.Canceled) {
			return context.Canceled
		}

		if commitBlockErr != nil {
			log.Println("error committing block atomically", commitBlockErr)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxDelay)
			continue
		}

		log.Println("✅ Block", nextBlock, "committed successfully")
		// reset retry delay
		retryDelay = time.Duration(w.config.SyncWorker.RetryDelayMs) * time.Millisecond
	}

}

func (w *NetworkWorker) getAllAddresses(ctx context.Context) ([]db.AddressRegistryEntity, error) {
	rows, err := w.db.Pool().Query(ctx, "SELECT * FROM address_registry where chain_id = $1", w.config.ChainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entities := make([]db.AddressRegistryEntity, 0)
	for rows.Next() {
		var entity db.AddressRegistryEntity
		err := rows.Scan(&entity.Address, &entity.ChainID, &entity.EntityType, &entity.Label, &entity.CreatedAt)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func (w *NetworkWorker) fetchLastProcessedBlock(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := w.db.Pool().QueryRow(ctx, "SELECT COALESCE(MAX(block_number), $1 - 1) FROM blocks where chain_id = $2", w.config.StartBlock, w.config.ChainID).Scan(&blockNumber)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil

}

func (w *NetworkWorker) fetchERC20TransferLogs(ctx context.Context, block int64) ([]RPCLog, error) {

	rpcResult, err := w.rpcc.callRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.call(ctx, "eth_getLogs", []interface{}{map[string]interface{}{
			"fromBlock": block,
			"toBlock":   block,
			"topics":    []string{w.config.ERC20TransferTopic},
		}})
	})

	if err != nil {
		return nil, err
	}

	var logs []RPCLog
	err = json.Unmarshal(rpcResult, &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil

}

func (w *NetworkWorker) ingestERC20(ctx context.Context) {}

func (w *NetworkWorker) fetchBlockWithTransactions(ctx context.Context, blockNumber int64) (RPCBlock, error) {

	rpcResult, err := w.rpcc.callRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.call(ctx, "eth_getBlockByNumber", []interface{}{blockNumber, true})
	})
	if err != nil {
		return RPCBlock{}, err
	}

	var block RPCBlock
	err = json.Unmarshal(rpcResult, &block)
	if err != nil {
		return RPCBlock{}, err
	}

	return block, nil
}

func (w *NetworkWorker) filterTransactions(transactions []RPCTransaction) []RPCTransaction {
	filteredTransactions := make([]RPCTransaction, 0)

	for _, transaction := range transactions {

		if transaction.Value == "0x0" {
			continue
		}

		if w.addressSet[normalizeAddress(transaction.From)] || w.addressSet[normalizeAddress(transaction.To)] {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
}

func (w *NetworkWorker) filterERC20TransferLogs(logs []RPCLog) []RPCLog {
	filteredLogs := make([]RPCLog, 0)

	for _, l := range logs {
		method := l.Topics[0]

		if !isAddressEqual(method, w.config.ERC20TransferTopic) || len(l.Topics) != 3 {
			continue
		}

		from := l.Topics[1]
		to := l.Topics[2]

		if w.addressSet[normalizeAddress(from)] || w.addressSet[normalizeAddress(to)] {
			filteredLogs = append(filteredLogs, l)
		}
	}
	return filteredLogs
}

func (w *NetworkWorker) commitBlock(ctx context.Context, block RPCBlock, erc20TransferLogs []RPCLog) error {

	return w.db.RunInTx(ctx, func(tx pgx.Tx) error {

		var err error

		err = w.insertBlock(ctx, tx, block)
		if err != nil {
			return err
		}

		err = w.insertNativeTransfers(ctx, tx, block)

		if err != nil {
			return err
		}

		err = w.insertERC20Transfers(ctx, tx, block, erc20TransferLogs)

		if err != nil {
			return err
		}

		return nil
	})
}

func (w *NetworkWorker) insertNativeTransfers(ctx context.Context, tx pgx.Tx, block RPCBlock) error {

	if len(block.Transactions) == 0 {
		return nil
	}

	blockNumber, err := hexToUint64(block.Number)
	if err != nil {
		return err
	}

	blockTimestamp, err := hexToTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	query := "INSERT INTO native_transfers (tx_hash, block_number, block_timestamp, from_address, to_address, amount_raw, chain_id) VALUES "
	values := make([]interface{}, 0, len(block.Transactions))
	valuesPlaceholder := make([]string, 0, len(block.Transactions))
	counter := 0
	// insert into table (column1, column2, column3, ...) values (value1, value2, value3, ...), (value1, value2, value3, ...), ...
	for _, transaction := range block.Transactions {
		currentPlaceholder := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", counter+1, counter+2, counter+3, counter+4, counter+5, counter+6, counter+7)
		counter += 7

		values = append(values,
			transaction.Hash, blockNumber, blockTimestamp,
			normalizeAddress(transaction.From),
			normalizeAddress(transaction.To),
			transaction.Value, w.config.ChainID,
		)
		valuesPlaceholder = append(valuesPlaceholder, currentPlaceholder)
	}

	query += strings.Join(valuesPlaceholder, ", ") + " ON CONFLICT (tx_hash, chain_id) DO NOTHING"

	_, err = tx.Exec(ctx, query, values...)

	return err
}

func (w *NetworkWorker) insertBlock(ctx context.Context, tx pgx.Tx, block RPCBlock) error {
	blockNumber, err := hexToUint64(block.Number)
	if err != nil {
		return err
	}

	blockTimestamp, err := hexToTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	query := "INSERT INTO blocks (block_number, block_timestamp, chain_id) VALUES ($1, $2, $3) ON CONFLICT (block_number, chain_id) DO NOTHING"
	_, err = tx.Exec(ctx, query, int64(blockNumber), blockTimestamp, w.config.ChainID)

	return err
}

func (w *NetworkWorker) insertERC20Transfers(ctx context.Context, tx pgx.Tx, block RPCBlock, erc20TransferLogs []RPCLog) error {

	if len(erc20TransferLogs) == 0 {
		return nil
	}

	blockNumber, err := hexToUint64(block.Number)
	if err != nil {

		return err
	}

	blockTimestamp, err := hexToTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	query := "INSERT INTO erc20_transfers (tx_hash, log_index, block_number, block_timestamp, from_address, to_address, token_address, amount_raw, chain_id) VALUES "
	values := make([]interface{}, 0, len(erc20TransferLogs))
	valuesPlaceholder := make([]string, 0, len(erc20TransferLogs))
	counter := 0

	for _, l := range erc20TransferLogs {
		currentPlaceholder := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", counter+1, counter+2, counter+3, counter+4, counter+5, counter+6, counter+7, counter+8, counter+9)
		counter += 9

		logIndex, err := hexToUint64(l.LogIndex)
		if err != nil {
			return err
		}

		values = append(values,
			l.TxHash, logIndex,
			blockNumber, blockTimestamp,
			normalizeAddress(l.Topics[1]),
			normalizeAddress(l.Topics[2]),
			normalizeAddress(l.Address),
			l.Data, w.config.ChainID)

		valuesPlaceholder = append(valuesPlaceholder, currentPlaceholder)
	}

	query += strings.Join(valuesPlaceholder, ", ") + " ON CONFLICT (tx_hash, chain_id, log_index) DO NOTHING"
	_, err = tx.Exec(ctx, query, values...)

	return err
}

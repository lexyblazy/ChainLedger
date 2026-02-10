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
	addrUtil "polychain.capital/internal/address"
	"polychain.capital/internal/hex"
	"polychain.capital/rpc"
)

type NetworkWorker struct {
	db         *db.DB
	rpcc       *rpc.Client
	config     *config.NetworkConfig
	addressSet map[string]bool
}

func (w *NetworkWorker) tokenDiscovery(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := w.syncDiscoveredTokensInBatch(ctx)
		if err != nil {
			log.Println(w.config.Name, "Error syncing discovered tokens", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(w.config.TokenMetadata.DiscoveryIntervalSeconds)):
		}
	}

}

func (w *NetworkWorker) syncTokenMetadata(ctx context.Context) error {

	for {

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entities, err := w.fetchTokensForMetadataSyncInBatch(ctx)
		if err != nil {
			log.Println(w.config.Name, "Error fetching tokens for metadata sync", err)
			continue
		}

		err = w.fetchTokensMetadata(ctx, entities)
		if err != nil {
			log.Println(w.config.Name, "Error fetching tokens metadata", err)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second * time.Duration(w.config.TokenMetadata.MetadataFetchIntervalSeconds)):

		}
	}

}

func (w *NetworkWorker) populateAddressSet(ctx context.Context) error {
	addresses, err := w.getAllAddresses(ctx)
	if err != nil {
		return err
	}
	for _, address := range addresses {
		w.addressSet[addrUtil.Normalize(address)] = true
	}

	return nil
}

func (w *NetworkWorker) refreshAddressSet(ctx context.Context) error {
	interval := time.Second * time.Duration(w.config.SyncWorker.UpdateAddressSetIntervalSeconds)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			addresses, err := w.getNewlyAddedAddresses(ctx)
			if err != nil {
				log.Println(w.config.Name, "error getting newly added addresses", err)
				continue
			}
			for _, address := range addresses {
				w.addressSet[addrUtil.Normalize(address)] = true
			}
		}
	}
}

func (w *NetworkWorker) start(ctx context.Context) error {
	log.Println("🔗 Started", w.config.Name, "ingestion workflow")

	err := w.populateAddressSet(ctx)

	if err != nil {
		return err
	}

	if len(w.addressSet) == 0 {
		return errors.New("no addresses found")
	}

	retryDelay := time.Duration(w.config.SyncWorker.RetryDelayMs) * time.Millisecond
	maxDelay := time.Duration(w.config.SyncWorker.MaxDelayMs) * time.Millisecond

	// start syncing erc20 token metadata in the background
	// 1. do a token discovery i.e
	//  Find ERC-20 token contracts that have appeared in transfers, that we haven’t yet inserted into tokens,
	//  ordered by when we first saw them. Then save this batch to the tokens table.
	go w.tokenDiscovery(ctx)

	// 2. do a token metadata fetch i.e
	// Read directly from the tokens table in batch and fetch the metadata for each token from the rpc,
	// then update the token
	go w.syncTokenMetadata(ctx)

	// 3. refresh the address set every N seconds
	go w.refreshAddressSet(ctx)

	var (
		lastKnownHead     int64
		lastHeadCheckTime time.Time
	)

	for {

		var (
			block                     rpc.Block
			erc20TransferLogs         []rpc.Log
			fetchBlockErr             error
			fetchERC20TransferLogsErr error
			commitBlockErr            error
		)

		bn, err := w.fetchLastProcessedBlock(ctx)

		if errors.Is(err, context.Canceled) {
			return context.Canceled
		}

		if err != nil {
			log.Println(w.config.Name, "error fetching last processed block", err)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxDelay)
			continue
		}

		nextBlockNumber := bn + 1

		// do a check to ensure that there is sufficient block gap between the last known head and the best rpc block
		// and throttle the rpc calls to ensure sufficient gap between the rpc calls
		blockDelta := lastKnownHead - w.config.SyncWorker.BlockGap

		if lastKnownHead == 0 || nextBlockNumber > blockDelta {

			headPollInterval := time.Second * time.Duration(w.config.SyncWorker.BestRpcBlockPollIntervalSeconds)

			if time.Since(lastHeadCheckTime) < headPollInterval {
				time.Sleep(headPollInterval - time.Since(lastHeadCheckTime))
				continue
			}

			bestRpcBlockNumber, err := w.getBestRpcBlockNumber(ctx)

			if errors.Is(err, context.Canceled) {
				return context.Canceled
			}

			if err != nil {
				log.Println(w.config.Name, "error getting best rpc block number", err)
				time.Sleep(retryDelay)
				retryDelay = min(retryDelay*2, maxDelay)
				continue
			}

			lastKnownHead = bestRpcBlockNumber
			lastHeadCheckTime = time.Now()

			continue
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
		}(nextBlockNumber)

		go func(blockNumber int64) {
			defer wg.Done()
			l, err := w.fetchERC20TransferLogs(ctx, blockNumber)
			if err != nil {
				fetchERC20TransferLogsErr = err
				return
			}
			erc20TransferLogs = w.filterERC20TransferLogs(l)
		}(nextBlockNumber)

		wg.Wait()

		isContextCanceled := errors.Is(fetchBlockErr, context.Canceled) || errors.Is(fetchERC20TransferLogsErr, context.Canceled)
		if isContextCanceled {
			return context.Canceled
		}

		if fetchBlockErr != nil {
			log.Println(w.config.Name, "error fetching block transactions", fetchBlockErr)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxDelay)
			continue
		}

		if fetchERC20TransferLogsErr != nil {
			log.Println(w.config.Name, "error fetching erc20 transfer logs", fetchERC20TransferLogsErr)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxDelay)
			continue
		}

		commitBlockErr = w.commitBlock(ctx, block, erc20TransferLogs)

		if errors.Is(commitBlockErr, context.Canceled) {
			return context.Canceled
		}

		if commitBlockErr != nil {
			log.Println(w.config.Name, "error committing block atomically", commitBlockErr)
			time.Sleep(retryDelay)
			retryDelay = min(retryDelay*2, maxDelay)
			continue
		}

		if nextBlockNumber%w.config.SyncWorker.BalanceSnapshotIntervalBlocks == 0 {
			// take a snapshot of the balances
			err := w.takeBalanceSnapshot(ctx, block)
			if err != nil {
				log.Println(w.config.Name, "error taking balance snapshot", err)
			}
		}

		// reset retry delay
		retryDelay = time.Duration(w.config.SyncWorker.RetryDelayMs) * time.Millisecond
	}

}

func (w *NetworkWorker) takeBalanceSnapshot(ctx context.Context, block rpc.Block) error {

	query := `INSERT INTO balance_snapshots (
		chain_id, wallet_address, asset_type,
		asset_address, balance_raw,
		block_number, block_timestamp
	) SELECT chain_id, wallet_address, asset_type, asset_address,
		 	balance_raw, $1 as block_number, $2 as block_timestamp
		 FROM balances
	`

	blockNumber, err := hex.DecodeUint64(block.Number)
	if err != nil {
		return err
	}

	blockTimestamp, err := hex.DecodeTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	_, err = w.db.Pool().Exec(ctx, query, block.Number, blockTimestamp)
	if err != nil {
		log.Println(w.config.Name, "error taking balance snapshot", err)
		return err
	}

	log.Printf("✅ %s took balance snapshot at block %d", w.config.Name, blockNumber)
	return nil
}

func (w *NetworkWorker) getNewlyAddedAddresses(ctx context.Context) ([]string, error) {

	params := []interface{}{w.config.ChainID, w.config.SyncWorker.UpdateAddressSetIntervalSeconds}
	query := "SELECT address FROM address_registry where chain_id = $1 and created_at > now() - ($2 * interval '1 second')"
	rows, err := w.db.Pool().Query(ctx, query, params...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]string, 0)
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (w *NetworkWorker) getAllAddresses(ctx context.Context) ([]string, error) {
	rows, err := w.db.Pool().Query(ctx, "SELECT address FROM address_registry where chain_id = $1", w.config.ChainID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]string, 0)
	for rows.Next() {
		var address string
		err := rows.Scan(&address)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, address)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

func (w *NetworkWorker) fetchLastProcessedBlock(ctx context.Context) (int64, error) {
	var blockNumber int64
	err := w.db.Pool().QueryRow(ctx, "SELECT COALESCE(MAX(block_number), $1 - 1) FROM blocks where chain_id = $2", w.config.StartBlock, w.config.ChainID).Scan(&blockNumber)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil

}

func (w *NetworkWorker) fetchERC20TransferLogs(ctx context.Context, blockNumber int64) ([]rpc.Log, error) {

	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_getLogs", []interface{}{map[string]interface{}{
			"fromBlock": hex.IntToHex(blockNumber),
			"toBlock":   hex.IntToHex(blockNumber),
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

func (w *NetworkWorker) fetchBlockWithTransactions(ctx context.Context, blockNumber int64) (rpc.Block, error) {

	rpcResult, err := w.rpcc.CallRpcWithRetry(ctx, func() (json.RawMessage, error) {
		return w.rpcc.Call(ctx, "eth_getBlockByNumber", []interface{}{hex.IntToHex(blockNumber), true})
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

func (w *NetworkWorker) filterTransactions(transactions []rpc.Transaction) []rpc.Transaction {
	filteredTransactions := make([]rpc.Transaction, 0)

	for _, transaction := range transactions {

		if transaction.Value == "0x0" {
			continue
		}

		if w.addressSet[addrUtil.Normalize(transaction.From)] || w.addressSet[addrUtil.Normalize(transaction.To)] {
			filteredTransactions = append(filteredTransactions, transaction)
		}
	}
	return filteredTransactions
}

func (w *NetworkWorker) filterERC20TransferLogs(logs []rpc.Log) []rpc.Log {
	filteredLogs := make([]rpc.Log, 0)

	for _, l := range logs {
		method := l.Topics[0]

		if !addrUtil.IsEqual(method, w.config.ERC20TransferTopic) || len(l.Topics) != 3 {
			continue
		}

		from := l.Topics[1]
		to := l.Topics[2]

		if w.addressSet[addrUtil.Normalize(from)] || w.addressSet[addrUtil.Normalize(to)] {
			filteredLogs = append(filteredLogs, l)
		}
	}
	return filteredLogs
}

func (w *NetworkWorker) commitBlock(ctx context.Context, block rpc.Block, erc20TransferLogs []rpc.Log) error {

	return w.db.RunInTx(ctx, func(tx pgx.Tx) error {

		var err error

		err = w.insertBlock(ctx, tx, block)
		if err != nil {
			log.Println("error inserting block", err)
			return err
		}

		err = w.insertNativeTransfers(ctx, tx, block)

		if err != nil {
			log.Println("error inserting native transfers", err)
			return err
		}

		err = w.insertERC20Transfers(ctx, tx, block, erc20TransferLogs)

		if err != nil {
			log.Println("error inserting erc20 transfers", err)
			return err
		}

		err = w.updateBalances(ctx, tx, block)

		if err != nil {
			log.Println("error updating balances", err)
			return err
		}

		return nil
	})
}

func (w *NetworkWorker) updateBalances(ctx context.Context, tx pgx.Tx, block rpc.Block) error {
	query := `
INSERT INTO balances (
    wallet_address,
    chain_id,
    asset_type,
    asset_address,
    balance_raw
)
SELECT
    wallet_address,
    chain_id,
    asset_type,
    COALESCE(asset_address, 'native') as asset_address,
    SUM(amount_raw) AS delta
FROM asset_flows
WHERE block_number = $1
GROUP BY
    wallet_address,
    chain_id,
    asset_type,
    asset_address
ON CONFLICT (wallet_address, chain_id, asset_type, asset_address)
DO UPDATE
SET
    balance_raw = balances.balance_raw + EXCLUDED.balance_raw,
    updated_at = now();

	`
	blockNumber, err := hex.DecodeUint64(block.Number)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, blockNumber)

	if err != nil {
		return err
	}

	return nil
}

func (w *NetworkWorker) insertNativeTransfers(ctx context.Context, tx pgx.Tx, block rpc.Block) error {

	if len(block.Transactions) == 0 {
		return nil
	}

	blockNumber, err := hex.DecodeUint64(block.Number)
	if err != nil {
		return err
	}

	blockTimestamp, err := hex.DecodeTimestamp(block.Timestamp)
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
			addrUtil.Normalize(transaction.From),
			addrUtil.Normalize(transaction.To),
			transaction.Value, w.config.ChainID,
		)
		valuesPlaceholder = append(valuesPlaceholder, currentPlaceholder)
	}

	query += strings.Join(valuesPlaceholder, ", ") + " ON CONFLICT (tx_hash, chain_id) DO NOTHING"

	_, err = tx.Exec(ctx, query, values...)

	return err
}

func (w *NetworkWorker) insertBlock(ctx context.Context, tx pgx.Tx, block rpc.Block) error {
	blockNumber, err := hex.DecodeUint64(block.Number)
	if err != nil {
		return err
	}

	blockTimestamp, err := hex.DecodeTimestamp(block.Timestamp)
	if err != nil {
		return err
	}

	query := "INSERT INTO blocks (block_number, block_timestamp, chain_id) VALUES ($1, $2, $3) ON CONFLICT (block_number, chain_id) DO NOTHING"
	_, err = tx.Exec(ctx, query, int64(blockNumber), blockTimestamp, w.config.ChainID)

	return err
}

func (w *NetworkWorker) insertERC20Transfers(ctx context.Context, tx pgx.Tx, block rpc.Block, erc20TransferLogs []rpc.Log) error {

	if len(erc20TransferLogs) == 0 {
		return nil
	}

	blockNumber, err := hex.DecodeUint64(block.Number)
	if err != nil {

		return err
	}

	blockTimestamp, err := hex.DecodeTimestamp(block.Timestamp)
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

		logIndex, err := hex.DecodeUint64(l.LogIndex)
		if err != nil {
			return err
		}

		values = append(values,
			l.TxHash, logIndex,
			blockNumber, blockTimestamp,
			addrUtil.Normalize(l.Topics[1]),
			addrUtil.Normalize(l.Topics[2]),
			addrUtil.Normalize(l.Address),
			l.Data, w.config.ChainID)

		valuesPlaceholder = append(valuesPlaceholder, currentPlaceholder)
	}

	query += strings.Join(valuesPlaceholder, ", ") + " ON CONFLICT (tx_hash, chain_id, log_index) DO NOTHING"
	_, err = tx.Exec(ctx, query, values...)

	return err
}

func (w *NetworkWorker) syncDiscoveredTokensInBatch(ctx context.Context) error {
	query := `
	select chain_id, token_address, first_seen_block
	from (
		select
			et.chain_id,
			et.token_address,
			min(et.block_number) as first_seen_block
		from erc20_transfers et
		where et.chain_id = $1 and not exists (
			select 1
			from tokens t
			where t.chain_id = et.chain_id
			and t.token_address = et.token_address
		)
		group by et.chain_id, et.token_address
	) s
	order by first_seen_block
	limit $2;
	`
	rows, err := w.db.Pool().Query(ctx, query, w.config.ChainID, w.config.TokenDiscoveryBatchSize)

	if err != nil {
		return err
	}
	defer rows.Close()

	entities := make([]db.TokenEntity, 0)

	for rows.Next() {
		var entity db.TokenEntity
		err := rows.Scan(&entity.ChainID, &entity.TokenAddress, &entity.FirstSeenBlock)
		if err != nil {
			return err
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(entities) == 0 {
		return nil
	}

	query = "INSERT INTO tokens (chain_id, token_address, first_seen_block) VALUES "
	values := make([]interface{}, 0, len(entities))
	valuesPlaceholder := make([]string, 0, len(entities))
	counter := 0

	for _, entity := range entities {
		currentPlaceholder := fmt.Sprintf("($%d, $%d, $%d)", counter+1, counter+2, counter+3)
		counter += 3

		values = append(values, entity.ChainID, entity.TokenAddress, entity.FirstSeenBlock)
		valuesPlaceholder = append(valuesPlaceholder, currentPlaceholder)
	}

	query += strings.Join(valuesPlaceholder, ", ") + " ON CONFLICT (chain_id, token_address) DO NOTHING"
	_, err = w.db.Pool().Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func (w *NetworkWorker) fetchTokensMetadata(ctx context.Context, entities []db.TokenEntity) error {
	if len(entities) == 0 {
		return nil
	}

	sem := make(chan struct{}, w.config.TokenMetadata.RpcBatchSize)

	for _, entity := range entities {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sem <- struct{}{}:
		}

		go func(token db.TokenEntity) {
			defer func() { <-sem }()

			meta, err := w.fetchTokenMetadata(ctx, token)

			if meta.Name == nil || meta.Symbol == nil || meta.Decimals == nil {
				w.markTokenMetadataFetchFailed(ctx, token)
				return
			}

			err = w.updateTokenMetadata(ctx, token, meta)
			if err != nil {
				log.Println(w.config.Name, "Error updating token metadata", err)
			}
		}(entity)
	}

	return nil
}

func (w *NetworkWorker) fetchTokenMetadataDecimals(ctx context.Context, tokenAddress string) (int8, error) {
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

func (w *NetworkWorker) fetchTokenMetadataName(ctx context.Context, tokenAddress string) (string, error) {
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

func (w *NetworkWorker) fetchTokenMetadataSymbol(ctx context.Context, tokenAddress string) (string, error) {
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

func (w *NetworkWorker) fetchTokenMetadata(ctx context.Context, entity db.TokenEntity) (*TokenMetadata, error) {

	meta := &TokenMetadata{}
	tokenAddress := addrUtil.Format(entity.TokenAddress)

	// decimals (most important)
	decimals, err := w.fetchTokenMetadataDecimals(ctx, tokenAddress)
	if err == nil {
		meta.Decimals = &decimals
	}

	name, err := w.fetchTokenMetadataName(ctx, tokenAddress)
	if err == nil {
		meta.Name = &name
	}
	symbol, err := w.fetchTokenMetadataSymbol(ctx, tokenAddress)
	if err == nil {
		meta.Symbol = &symbol
	}

	return meta, nil
}

func (w *NetworkWorker) updateTokenMetadata(ctx context.Context, entity db.TokenEntity, meta *TokenMetadata) error {
	query := "UPDATE tokens SET symbol = $1, name = $2, decimals = $3, metadata_fetch_failed_at = null WHERE chain_id = $4 AND token_address = $5"
	_, err := w.db.Pool().Exec(ctx, query, meta.Symbol, meta.Name, meta.Decimals, entity.ChainID, entity.TokenAddress)
	if err != nil {
		return err
	}

	return nil
}

func (w *NetworkWorker) markTokenMetadataFetchFailed(ctx context.Context, entity db.TokenEntity) error {
	query := "UPDATE tokens SET metadata_fetch_failed_at = now() WHERE chain_id = $1 AND token_address = $2"
	_, err := w.db.Pool().Exec(ctx, query, entity.ChainID, entity.TokenAddress)
	if err != nil {
		return err
	}
	return nil
}

func (w *NetworkWorker) fetchTokensForMetadataSyncInBatch(ctx context.Context) ([]db.TokenEntity, error) {
	query := `
	select chain_id, token_address
	from tokens
	where chain_id = $1
	and (symbol is null or name is null or decimals is null)
	and (metadata_fetch_failed_at is null or metadata_fetch_failed_at < now() - make_interval(hours => $2))
	limit $3;
	`
	rows, err := w.db.Pool().Query(ctx, query, w.config.ChainID, w.config.TokenMetadata.FailedIntervalHours, w.config.TokenMetadata.DbBatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entities := make([]db.TokenEntity, 0)
	for rows.Next() {
		var entity db.TokenEntity
		err := rows.Scan(&entity.ChainID, &entity.TokenAddress)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entities, nil

}

func (w *NetworkWorker) getBestRpcBlockNumber(ctx context.Context) (int64, error) {
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

func (w *NetworkWorker) getStatus(ctx context.Context) (NetworkStatus, error) {
	var status NetworkStatus
	var err error
	status.Name = w.config.Name
	status.ChainID = int64(w.config.ChainID)
	status.StartBlock = w.config.StartBlock
	status.AddressesCount = len(w.addressSet)

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

	}()

	wg.Wait()

	status.IngestionLagBlocksCount = status.BestRpcBlockNumber - status.IngestedBlocksMaxBlockNumber
	status.IngestionProgressPct = float64(status.IngestedBlocksCount) / float64(status.BestRpcBlockNumber-status.StartBlock) * 100

	return status, nil

}

package server

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"polychain.capital/db"
	"polychain.capital/internal/hex"
)

func (s *Server) getOffsetPaginationParams(r *http.Request) (limit int, offset int, err error) {
	limit = s.config.Api.MaxResults
	offset = 0

	page := r.URL.Query().Get("page")
	if page == "" {
		return limit, offset, nil
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return limit, offset, err
	}

	offset = (pageInt - 1) * limit

	return limit, max(offset, 0), nil
}

func (s *Server) getCursorPaginationParams(r *http.Request, cursorName string) (limit int, cursor interface{}) {
	limit = s.config.Api.MaxResults

	return limit, r.URL.Query().Get(cursorName)
}

func (s *Server) getStatus(r *http.Request) (any, error) {
	networkStatuses, err := s.i.GetStatus(r.Context())
	var response GetStatusResponse
	response.Data.Networks = networkStatuses

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Server) getWallets(r *http.Request) (any, error) {
	limit, offset, err := s.getOffsetPaginationParams(r)
	chainID := r.URL.Query().Get("chain_id")

	var response GetWalletsResponse

	query := `SELECT address, chain_id, entity_type, label FROM address_registry`
	params := []interface{}{limit, offset}

	if chainID != "" {
		query += fmt.Sprintf(" WHERE chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainID)
	}

	query += ` LIMIT $1 OFFSET $2`

	rows, err := s.db.Pool().Query(r.Context(), query, params...)
	if err != nil {
		log.Println("error getting wallets", err)
		return nil, err
	}
	defer rows.Close()

	wallets := make([]db.AddressRegistryEntity, 0)

	for rows.Next() {
		var wallet db.AddressRegistryEntity
		err := rows.Scan(&wallet.Address, &wallet.ChainID, &wallet.EntityType, &wallet.Label)
		if err != nil {
			return nil, err
		}
		wallet.Address = hex.FormatAddress(wallet.Address)
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	response.Data.Wallets = wallets
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, nil
}

func (s *Server) getTokens(r *http.Request) (any, error) {
	limit, offset, err := s.getOffsetPaginationParams(r)
	var response GetTokensResponse

	if err != nil {
		return nil, err
	}

	query := `SELECT chain_id, token_address, symbol, name, decimals FROM tokens`
	params := []interface{}{limit, offset}

	chainID := r.URL.Query().Get("chain_id")
	if chainID != "" {
		query += fmt.Sprintf(" WHERE chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainID)
	}
	query += ` order by first_seen_block asc LIMIT $1 OFFSET $2`

	rows, err := s.db.Pool().Query(r.Context(), query, params...)
	if err != nil {
		log.Println("error getting tokens", err)
		return nil, err
	}
	defer rows.Close()

	tokens := make([]db.TokenEntity, 0)

	for rows.Next() {
		var token db.TokenEntity
		err := rows.Scan(&token.ChainID, &token.TokenAddress, &token.Symbol, &token.Name, &token.Decimals)
		if err != nil {
			return nil, err
		}
		token.TokenAddress = hex.FormatAddress(token.TokenAddress)
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	response.Data.Tokens = tokens
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, nil
}

func (s *Server) getWalletBalanceSnapshots(r *http.Request) (any, error) {

	var response GetWalletBalanceSnapshotsResponse
	address := r.PathValue("address")
	limit, cursorId := s.getCursorPaginationParams(r, "cursor_id")

	params := []interface{}{hex.NormalizeAddress(address), limit}

	chainID := r.URL.Query().Get("chain_id")
	tokenAddress := r.URL.Query().Get("token_address")

	query := `SELECT id, chain_id, wallet_address, asset_type, asset_address,
	  balance_raw, block_number, block_timestamp
	FROM balance_snapshots WHERE wallet_address = $1`

	if chainID != "" {
		query += fmt.Sprintf(" AND chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainID)
	}

	if hex.IsValidAddressLength(tokenAddress) {
		query += fmt.Sprintf(" AND asset_address = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, hex.NormalizeAddress(tokenAddress))
	} else if tokenAddress == "null" {
		// null is a special value for native assets
		query += " AND asset_address is null"
	}

	if cursorId != "" {
		query += fmt.Sprintf(" AND id < %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, cursorId)
	}

	query += ` order by id desc LIMIT $2`

	rows, err := s.db.Pool().Query(r.Context(), query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	balanceSnapshots := make([]db.BalanceSnapshotEntity, 0)
	for rows.Next() {
		var balanceSnapshot db.BalanceSnapshotEntity
		err := rows.Scan(&balanceSnapshot.ID, &balanceSnapshot.ChainID, &balanceSnapshot.WalletAddress, &balanceSnapshot.AssetType, &balanceSnapshot.AssetAddress, &balanceSnapshot.BalanceRaw, &balanceSnapshot.BlockNumber, &balanceSnapshot.BlockTimestamp)
		if err != nil {
			return nil, err
		}
		balanceSnapshot.WalletAddress = hex.FormatAddress(balanceSnapshot.WalletAddress)
		if balanceSnapshot.AssetAddress != nil {
			formattedAddress := hex.FormatAddress(*balanceSnapshot.AssetAddress)
			balanceSnapshot.AssetAddress = &formattedAddress
		}
		balanceSnapshots = append(balanceSnapshots, balanceSnapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	response.Data.BalanceSnapshots = balanceSnapshots
	response.Meta.CursorId = balanceSnapshots[len(balanceSnapshots)-1].ID
	response.Meta.Limit = limit

	return response, nil
}

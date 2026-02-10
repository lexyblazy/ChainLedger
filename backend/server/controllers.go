package server

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (s *Server) getStatus(r *http.Request) (any, int, error) {
	networkStatuses, err := s.i.GetStatus(r.Context())
	var response GetStatusResponse
	response.Data.Networks = networkStatuses

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return response, http.StatusOK, nil
}

func (s *Server) createWallet(r *http.Request) (any, int, error) {

	var body CreateWalletRequestBody
	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	if body.Address == "" || !hex.IsValidAddressLength(body.Address) {
		return nil, http.StatusBadRequest, errors.New("address is required and must be a valid address")
	}

	networkConfig := s.config.Networks[strconv.Itoa(body.ChainID)]
	if networkConfig.Name == "" {
		return nil, http.StatusBadRequest, errors.New("chain_id is not supported")
	}

	query := `INSERT INTO address_registry (address, chain_id, entity_type, label)
	 VALUES ($1, $2, $3, $4) ON CONFLICT (address, chain_id)
	 DO UPDATE SET entity_type = $3, label = $4, updated_at = now()`

	_, err = s.db.Pool().Exec(r.Context(), query, hex.NormalizeAddress(body.Address), body.ChainID, body.EntityType, body.Label)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return &SuccessResponse{
		Success: true,
	}, http.StatusOK, nil
}

func (s *Server) handleWallets(r *http.Request) (any, int, error) {

	if r.Method == http.MethodPost {
		return s.createWallet(r)
	}

	limit, offset, err := s.getOffsetPaginationParams(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	chainID := r.URL.Query().Get("chain_id")
	var response GetWalletsResponse

	query := `SELECT * FROM address_registry`
	params := []interface{}{limit, offset}

	if chainID != "" {
		query += fmt.Sprintf(" WHERE chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainID)
	}

	query += ` LIMIT $1 OFFSET $2`

	rows, err := s.db.Pool().Query(r.Context(), query, params...)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer rows.Close()

	wallets := make([]db.AddressRegistryEntity, 0)

	for rows.Next() {
		var wallet db.AddressRegistryEntity
		err := rows.Scan(&wallet.Address, &wallet.ChainID, &wallet.EntityType, &wallet.Label, &wallet.CreatedAt, &wallet.UpdatedAt)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		wallet.Address = hex.FormatAddress(wallet.Address)
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.Wallets = wallets
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

func (s *Server) getTokens(r *http.Request) (any, int, error) {
	limit, offset, err := s.getOffsetPaginationParams(r)

	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var response GetTokensResponse

	query := `SELECT chain_id, token_address, symbol, name, decimals FROM tokens`
	params := []interface{}{limit, offset}

	chainIdStr := r.URL.Query().Get("chain_id")
	if chainIdStr != "" {
		chainId, err := strconv.Atoi(chainIdStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}
		query += fmt.Sprintf(" WHERE chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainId)
	}

	query += ` order by first_seen_block asc LIMIT $1 OFFSET $2`

	rows, err := s.db.Pool().Query(r.Context(), query, params...)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	defer rows.Close()

	tokens := make([]db.TokenEntity, 0)

	for rows.Next() {
		var token db.TokenEntity
		err := rows.Scan(&token.ChainID, &token.TokenAddress, &token.Symbol, &token.Name, &token.Decimals)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		token.TokenAddress = hex.FormatAddress(token.TokenAddress)
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.Tokens = tokens
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

func (s *Server) getWalletBalance(r *http.Request) (any, int, error) {
	var response GetWalletBalanceResponse
	limit, offset, err := s.getOffsetPaginationParams(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	address := r.PathValue("address")
	tokenAddress := r.URL.Query().Get("token_address")
	params := []interface{}{hex.NormalizeAddress(address), limit, offset}
	query := `SELECT chain_id, wallet_address, asset_type, asset_address, balance_raw FROM balances WHERE wallet_address = $1`

	chainIdStr := r.URL.Query().Get("chain_id")
	if chainIdStr != "" {
		chainId, err := strconv.Atoi(chainIdStr)
		if err != nil {
			return nil, http.StatusBadRequest, err
		}

		query += fmt.Sprintf(" AND chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainId)
	}

	if hex.IsValidAddressLength(tokenAddress) {
		query += fmt.Sprintf(" AND asset_address = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, hex.NormalizeAddress(tokenAddress))
	} else if tokenAddress == "null" {
		query += " AND asset_address is null"
	}

	query += " ORDER BY asset_type,asset_address DESC LIMIT $2 OFFSET $3"

	rows, err := s.db.Pool().Query(r.Context(), query, params...)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer rows.Close()

	balances := make([]db.BalanceEntity, 0)
	for rows.Next() {
		var balance db.BalanceEntity
		err := rows.Scan(&balance.ChainID, &balance.WalletAddress, &balance.AssetType, &balance.AssetAddress, &balance.BalanceRaw)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		balance.WalletAddress = hex.FormatAddress(balance.WalletAddress)
		if balance.AssetAddress != nil {
			formattedAddress := hex.FormatAddress(*balance.AssetAddress)
			balance.AssetAddress = &formattedAddress
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.Balance = balances
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

func (s *Server) getWalletBalanceSnapshots(r *http.Request) (any, int, error) {

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
		return nil, http.StatusInternalServerError, err
	}
	defer rows.Close()

	balanceSnapshots := make([]db.BalanceSnapshotEntity, 0)
	for rows.Next() {
		var balanceSnapshot db.BalanceSnapshotEntity
		err := rows.Scan(&balanceSnapshot.ID, &balanceSnapshot.ChainID, &balanceSnapshot.WalletAddress, &balanceSnapshot.AssetType, &balanceSnapshot.AssetAddress, &balanceSnapshot.BalanceRaw, &balanceSnapshot.BlockNumber, &balanceSnapshot.BlockTimestamp)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		balanceSnapshot.WalletAddress = hex.FormatAddress(balanceSnapshot.WalletAddress)
		if balanceSnapshot.AssetAddress != nil {
			formattedAddress := hex.FormatAddress(*balanceSnapshot.AssetAddress)
			balanceSnapshot.AssetAddress = &formattedAddress
		}
		balanceSnapshots = append(balanceSnapshots, balanceSnapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.BalanceSnapshots = balanceSnapshots
	response.Meta.CursorId = balanceSnapshots[len(balanceSnapshots)-1].ID
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

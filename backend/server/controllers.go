package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"polychain.capital/db"
	addrUtil "polychain.capital/internal/address" // address utility functions
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

	if body.Address == "" || !addrUtil.IsValidLength(body.Address) {
		return nil, http.StatusBadRequest, errors.New("address is required and must be a valid address")
	}

	networkConfig := s.config.Networks[strconv.Itoa(body.ChainID)]
	if networkConfig.Name == "" {
		return nil, http.StatusBadRequest, errors.New("chain_id is not supported")
	}

	query := `INSERT INTO address_registry (address, chain_id, entity_type, label)
	 VALUES ($1, $2, $3, $4) ON CONFLICT (address, chain_id)
	 DO UPDATE SET entity_type = $3, label = $4, updated_at = now()`

	_, err = s.db.Pool().Exec(r.Context(), query, addrUtil.Normalize(body.Address), body.ChainID, body.EntityType, body.Label)

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
		wallet.Address = addrUtil.Format(wallet.Address)
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
		token.TokenAddress = addrUtil.Format(token.TokenAddress)
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

func (s *Server) getWalletBalanceSnapshots(r *http.Request) (any, int, error) {

	var response GetWalletBalanceSnapshotsResponse
	address := r.PathValue("address")
	limit, cursorId := s.getCursorPaginationParams(r, "cursor_id")

	params := []interface{}{addrUtil.Normalize(address), limit}

	chainID := r.URL.Query().Get("chain_id")
	tokenAddress := r.URL.Query().Get("token_address")

	query := `SELECT id, chain_id, wallet_address, asset_type, asset_address,
	  balance_raw, block_number, block_timestamp
	FROM balance_snapshots WHERE wallet_address = $1`

	if chainID != "" {
		query += fmt.Sprintf(" AND chain_id = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, chainID)
	}

	if addrUtil.IsNativeAsset(tokenAddress) {
		query += " AND asset_address = 'native'"
	} else if addrUtil.IsValidLength(tokenAddress) {
		query += fmt.Sprintf(" AND asset_address = %s", fmt.Sprintf("$%d", len(params)+1))
		params = append(params, addrUtil.Normalize(tokenAddress))
	} else {
		return nil, http.StatusBadRequest, errors.New("token_address is required and must be a valid address")
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
		balanceSnapshot.WalletAddress = addrUtil.Format(balanceSnapshot.WalletAddress)
		if balanceSnapshot.AssetAddress != "native" {
			balanceSnapshot.AssetAddress = addrUtil.Format(balanceSnapshot.AssetAddress)
		}
		balanceSnapshots = append(balanceSnapshots, balanceSnapshot)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.BalanceSnapshots = balanceSnapshots
	if len(balanceSnapshots) > 0 {
		response.Meta.CursorId = balanceSnapshots[len(balanceSnapshots)-1].ID
	} else {
		response.Meta.CursorId = 0
	}
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

func (s *Server) getWalletPortfolio(r *http.Request) (any, int, error) {
	var response GetWalletPortfolioResponse
	address := r.PathValue("address")
	chainIDStr := r.URL.Query().Get("chain_id")

	if chainIDStr == "" {
		return nil, http.StatusBadRequest, errors.New("chain_id is required")
	}

	chainID, err := strconv.Atoi(chainIDStr)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	query := `
	SELECT
  b.asset_type,
  b.asset_address,
  b.balance_raw,
  t.symbol,
  t.decimals
FROM balances b
LEFT JOIN tokens t
  ON t.chain_id = b.chain_id
 AND t.token_address = b.asset_address
WHERE b.wallet_address = $1
  AND b.chain_id = $2
ORDER BY
  CASE WHEN b.asset_type = 'native' THEN 0 ELSE 1 END,
  t.symbol NULLS LAST,
  b.asset_address LIMIT $3 OFFSET $4;`

	limit, offset, err := s.getOffsetPaginationParams(r)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	params := []interface{}{addrUtil.Normalize(address), chainID, limit, offset}
	rows, err := s.db.Pool().Query(r.Context(), query, params...)

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	defer rows.Close()

	records := make([]WalletPortfolio, 0)

	for rows.Next() {
		var p WalletPortfolio
		err := rows.Scan(&p.AssetType, &p.AssetAddress, &p.BalanceRaw, &p.Symbol, &p.Decimals)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}

		if addrUtil.IsValidLength(p.AssetAddress) {
			p.AssetAddress = addrUtil.Format(p.AssetAddress)
		}

		if p.Symbol == nil {
			nativeSymbol := s.config.Networks[strconv.Itoa(chainID)].Symbol
			p.Symbol = &nativeSymbol
		}

		if p.Decimals == nil {
			nativeDecimals := int8(s.config.Networks[strconv.Itoa(chainID)].Decimals)
			p.Decimals = &nativeDecimals
		}

		records = append(records, p)
	}

	if err := rows.Err(); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	response.Data.Portfolio = records
	response.Meta.Offset = offset
	response.Meta.Limit = limit

	return response, http.StatusOK, nil
}

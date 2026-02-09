package server

import (
	"log"
	"net/http"
	"strconv"

	"polychain.capital/db"
	"polychain.capital/internal/hex"
)

func (s *Server) getPaginationParams(r *http.Request) (limit int, offset int, err error) {
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

func (s *Server) getStatus(r *http.Request) (any, error) {
	networkStatuses, err := s.i.GetStatus(r.Context())

	if err != nil {
		return nil, err
	}

	return StatusResponse{Networks: networkStatuses}, nil
}

func (s *Server) getWallets(r *http.Request) (any, error) {
	limit, offset, err := s.getPaginationParams(r)

	query := `SELECT address, chain_id, entity_type, label FROM address_registry LIMIT $1 OFFSET $2`
	rows, err := s.db.Pool().Query(r.Context(), query, limit, offset)
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

	return wallets, nil
}

func (s *Server) getTokens(r *http.Request) (any, error) {
	limit, offset, err := s.getPaginationParams(r)

	if err != nil {
		return nil, err
	}

	query := `SELECT chain_id, token_address, symbol, name, decimals 
	FROM tokens order by first_seen_block asc LIMIT $1 OFFSET $2`

	rows, err := s.db.Pool().Query(r.Context(), query, limit, offset)
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

	return tokens, nil
}

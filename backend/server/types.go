package server

import (
	"github.com/lexyblazy/chainledger/db"
	"github.com/lexyblazy/chainledger/ingestion"
)

type GetStatusResponse struct {
	Data struct {
		Networks map[string]ingestion.NetworkStatus `json:"networks"`
	} `json:"data"`
}

type OffsetPaginationMeta struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
}

type GetWalletsResponse struct {
	Data struct {
		Wallets []db.AddressRegistryEntity `json:"wallets"`
	} `json:"data"`

	Meta OffsetPaginationMeta `json:"meta"`
}

type CreateWalletRequestBody struct {
	Address    string  `json:"address"`
	ChainID    int     `json:"chain_id"`
	EntityType *string `json:"entity_type"`
	Label      *string `json:"label"`
}

type GetTokensResponse struct {
	Data struct {
		Tokens []db.TokenEntity `json:"tokens"`
	} `json:"data"`

	Meta OffsetPaginationMeta `json:"meta"`
}

type TokenMetadata struct {
	Symbol   *string `json:"symbol"`
	Name     *string `json:"name"`
	Decimals *int8   `json:"decimals"`
}

type GetWalletBalanceSnapshotsResponse struct {
	Data struct {
		BalanceSnapshots []db.BalanceSnapshotEntity `json:"balance_snapshots"`
		TokenMetadata    TokenMetadata              `json:"token_metadata"`
	} `json:"data"`

	Meta struct {
		CursorId int64 `json:"cursor_id"`
		Limit    int   `json:"limit"`
		HasMore  bool  `json:"has_more"`
	} `json:"meta"`
}

type WalletPortfolio struct {
	AssetType    string  `json:"asset_type"`
	AssetAddress string  `json:"asset_address"`
	BalanceRaw   string  `json:"balance_raw"`
	Symbol       *string `json:"symbol"`
	Decimals     *int8   `json:"decimals"`
}

type GetWalletPortfolioResponse struct {
	Data struct {
		Portfolio []WalletPortfolio `json:"portfolio"`
	} `json:"data"`

	Meta OffsetPaginationMeta `json:"meta"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

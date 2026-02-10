package server

import (
	"polychain.capital/db"
	"polychain.capital/ingestion"
)

type GetStatusResponse struct {
	Data struct {
		Networks map[string]ingestion.NetworkStatus `json:"networks"`
	} `json:"data"`
}

type GetWalletsResponse struct {
	Data struct {
		Wallets []db.AddressRegistryEntity `json:"wallets"`
	} `json:"data"`

	Meta struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"meta"`
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

	Meta struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"meta"`
}

type GetWalletBalanceSnapshotsResponse struct {
	Data struct {
		BalanceSnapshots []db.BalanceSnapshotEntity `json:"balance_snapshots"`
	} `json:"data"`

	Meta struct {
		CursorId int64 `json:"cursor_id"`
		Limit    int   `json:"limit"`
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

	Meta struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"meta"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

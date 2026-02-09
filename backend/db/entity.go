package db

import "time"

type AddressRegistryEntity struct {
	Address    string    `json:"address"`
	ChainID    int64     `json:"chain_id"`
	EntityType string    `json:"entity_type"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"-"`
}

type TokenEntity struct {
	ChainID               int64      `json:"chain_id"`      // required
	TokenAddress          string     `json:"token_address"` // required
	Symbol                *string    `json:"symbol"`        // nullable
	Name                  *string    `json:"name"`          // nullable
	Decimals              *int8      `json:"decimals"`      // nullable
	FirstSeenBlock        int64      `json:"-"`             // required
	MetadataFetchFailedAt *time.Time `json:"-"`             // nullable
	CreatedAt             time.Time  `json:"-"`
	UpdatedAt             time.Time  `json:"-"`
}

type BalanceSnapshotEntity struct {
	ID             int64     `json:"id"`
	ChainID        int64     `json:"chain_id"`
	WalletAddress  string    `json:"wallet_address"`
	AssetType      string    `json:"asset_type"`
	AssetAddress   *string   `json:"asset_address"` // nullable
	BalanceRaw     string    `json:"balance_raw"`
	BlockNumber    int64     `json:"block_number"`
	BlockTimestamp time.Time `json:"block_timestamp"`
	CreatedAt      time.Time `json:"-"`
}

// This is a view and not a table
type BalanceEntity struct {
	ID            int64     `json:"id"`
	ChainID       int64     `json:"chain_id"`
	WalletAddress string    `json:"wallet_address"`
	AssetType     string    `json:"asset_type"`
	AssetAddress  *string   `json:"asset_address"` // nullable
	BalanceRaw    string    `json:"balance_raw"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

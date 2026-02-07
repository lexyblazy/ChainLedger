package db

import "time"

type AddressRegistryEntity struct {
	Address    string    `json:"address"`
	ChainID    int64     `json:"chain_id"`
	EntityType string    `json:"entity_type"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
}

type TokenEntity struct {
	ChainID               int64      `json:"chain_id"`                 // required
	TokenAddress          string     `json:"token_address"`            // required
	Symbol                *string    `json:"symbol"`                   // nullable
	Name                  *string    `json:"name"`                     // nullable
	Decimals              *int8      `json:"decimals"`                 // nullable
	FirstSeenBlock        int64      `json:"first_seen_block"`         // required
	MetadataFetchFailedAt *time.Time `json:"metadata_fetch_failed_at"` // nullable
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

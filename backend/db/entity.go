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
	ChainID               int64      `json:"chain_id"`                 // required
	TokenAddress          string     `json:"token_address"`            // required
	Symbol                *string    `json:"symbol"`                   // nullable
	Name                  *string    `json:"name"`                     // nullable
	Decimals              *int8      `json:"decimals"`                 // nullable
	FirstSeenBlock        int64      `json:"-"`         // required
	MetadataFetchFailedAt *time.Time `json:"-"` // nullable
	CreatedAt             time.Time  `json:"-"`	
	UpdatedAt             time.Time  `json:"-"`
}

package db

import "time"

type AddressRegistryEntity struct {
	Address    string    `json:"address"`
	ChainID    int64     `json:"chain_id"`
	EntityType string    `json:"entity_type"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
}

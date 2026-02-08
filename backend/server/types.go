package server

import "polychain.capital/ingestion"

type StatusResponse struct {
	Networks map[string]ingestion.NetworkStatus `json:"networks"`
}

package ingestion

type TokenMetadata struct {
	Symbol   *string `json:"symbol"`
	Name     *string `json:"name"`
	Decimals *int8   `json:"decimals"`
}

type NetworkStatus struct {
	Name                         string  `json:"name"`
	ChainID                      int64   `json:"chainId"`
	StartBlock                   int64   `json:"startBlock"`
	BestRpcBlockNumber           int64   `json:"bestRpcBlockNumber"`
	IngestedBlocksCount          int64   `json:"ingestedBlocksCount"`
	IngestedBlocksMaxBlockNumber int64   `json:"ingestedBlocksMaxBlockNumber"`
	TokensCount                  int64   `json:"tokensCount"`
	IngestionLagBlocksCount      int64   `json:"ingestionLagBlocksCount"`
	IngestionProgressPct         float64 `json:"ingestionProgressPct"`
}

package ingestion

type RPCBlock struct {
	Number       string           `json:"number"`
	Hash         string           `json:"hash"`
	ParentHash   string           `json:"parentHash"`
	Timestamp    string           `json:"timestamp"`
	Transactions []RPCTransaction `json:"transactions"`
}

type RPCTransaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"` // hex wei
}

type RPCLog struct {
	Address     string   `json:"address"` // token contract
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	TxHash      string   `json:"transactionHash"`
	LogIndex    string   `json:"logIndex"`
	BlockNumber string   `json:"blockNumber"`
}

type TokenMetadata struct {
	Symbol   *string `json:"symbol"`
	Name     *string `json:"name"`
	Decimals *int8   `json:"decimals"`
}

package rpc

type Block struct {
	Number       string           `json:"number"`
	Hash         string           `json:"hash"`
	ParentHash   string           `json:"parentHash"`
	Timestamp    string           `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"` // hex wei
}

type Log struct {
	Address     string   `json:"address"` // token contract
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	TxHash      string   `json:"transactionHash"`
	LogIndex    string   `json:"logIndex"`
	BlockNumber string   `json:"blockNumber"`
}
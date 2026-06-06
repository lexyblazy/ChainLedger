package rpc

type Block struct {
	Number       string        `json:"number"`
	Hash         string        `json:"hash"`
	ParentHash   string        `json:"parentHash"`
	Timestamp    string        `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	TransactionIndex string `json:"transactionIndex"`
	Hash             string `json:"hash"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"` // hex wei
}

type Log struct {
	Address          string   `json:"address"` // token contract
	Topics           []string `json:"topics"`
	Data             string   `json:"data"`
	TxHash           string   `json:"transactionHash"`
	LogIndex         string   `json:"logIndex"`
	BlockNumber      string   `json:"blockNumber"`
	TransactionIndex string   `json:"transactionIndex"`
}

type TransactionReceipt struct {
	Type                  string  `json:"type"`
	Status                string  `json:"status"`
	CumulativeGasUsed     string  `json:"cumulativeGasUsed"`
	Logs                  []Log   `json:"logs"`
	DepositNonce          string  `json:"depositNonce"`
	DepositReceiptVersion string  `json:"depositReceiptVersion"`
	LogsBloom             string  `json:"logsBloom"`
	TransactionHash       string  `json:"transactionHash"`
	TransactionIndex      string  `json:"transactionIndex"`
	BlockHash             string  `json:"blockHash"`
	BlockNumber           string  `json:"blockNumber"`
	GasUsed               string  `json:"gasUsed"`
	EffectiveGasPrice     string  `json:"effectiveGasPrice"`
	BlobGasUsed           string  `json:"blobGasUsed"`
	From                  string  `json:"from"`
	To                    string  `json:"to"`
	ContractAddress       *string `json:"contractAddress"`
	// These are only present for L2 blocks
	L1GasPrice           *string `json:"l1GasPrice"`
	L1GasUsed            *string `json:"l1GasUsed"`
	L1Fee                *string `json:"l1Fee"`
	L1BaseFeeScalar      *string `json:"l1BaseFeeScalar"`
	L1BlobBaseFee        *string `json:"l1BlobBaseFee"`
	L1BlobBaseFeeScalar  *string `json:"l1BlobBaseFeeScalar"`
	DaFootprintGasScalar *string `json:"daFootprintGasScalar"`
}

package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Networks               map[string]NetworkConfig `json:"networks"`
	Database               DatabaseConfig           `json:"database"`
	Api                    ApiConfig                `json:"api"`
	ShutdownTimeoutSeconds int                      `json:"shutdownTimeoutSeconds"`
	SyncBlocks             bool                     `json:"syncBlocks"`
}

type NativeTokenConfig struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type WorkerConfig struct {
	RetryDelayMs                    int   `json:"retryDelayMs"`
	MaxDelayMs                      int   `json:"maxDelayMs"`
	BalanceSnapshotIntervalBlocks   int64 `json:"balanceSnapshotIntervalBlocks"`   // at every N blocks we take a snapshot of the balances
	BalanceSnapshotIntervalHours    int64 `json:"balanceSnapshotIntervalHours"`    // at every N hours we take a snapshot of the balances
	UpdateAddressSetIntervalSeconds int   `json:"updateAddressSetIntervalSeconds"` // at every N seconds we update the address set
	BlockGap                        int64 `json:"blockGap"`                        // Gap between the bestRpcBlock and the current block we are processing.
	BestRpcBlockPollIntervalSeconds int   `json:"bestRpcBlockPollIntervalSeconds"` // Throttle the polling of the best rpc block to ensure sufficient gap between the rpc calls
}

type NetworkConfig struct {
	StartBlock              int64               `json:"startBlock"`
	ERC20TransferTopic      string              `json:"erc20TransferTopic"`
	RPCUrl                  string              `json:"rpcURL"`
	ChainID                 int                 `json:"chainId"`
	Name                    string              `json:"name"`
	NativeToken             NativeTokenConfig   `json:"nativeToken"`
	RPCRateLimit            RPCRateLimitConfig  `json:"rpcRateLimit"`
	SyncWorker              WorkerConfig        `json:"syncWorker"`
	TokenMetadata           TokenMetadataConfig `json:"tokenMetadata"`
	TokenDiscoveryBatchSize int                 `json:"tokenDiscoveryBatchSize"`
}

type TokenMetadataConfig struct {
	DbBatchSize                  int               `json:"dbBatchSize"`
	RpcBatchSize                 int               `json:"rpcBatchSize"`
	FailedIntervalHours          int               `json:"failedIntervalHours"`
	RpcCalls                     map[string]string `json:"rpcCalls"`
	DiscoveryIntervalSeconds     int               `json:"discoveryIntervalSeconds"`
	MetadataFetchIntervalSeconds int               `json:"metadataFetchIntervalSeconds"`
}

type ApiConfig struct {
	Port       string `json:"port"`
	MaxResults int    `json:"maxResults"`
}

type RPCRateLimitConfig struct {
	RPS           int `json:"rps"`
	Burst         int `json:"burst"`
	MaxRetryCount int `json:"maxRetryCount"`
	RetryDelayMs  int `json:"retryDelayMs"`
}

type DatabaseConfig struct {
	URL string `json:"url"`
	ChunkSize int `json:"chunkSize"` // to specifically avoid too many parameters error
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	err = json.NewDecoder(jsonFile).Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

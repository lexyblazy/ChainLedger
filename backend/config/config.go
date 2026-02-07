package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Networks    map[string]NetworkConfig `json:"networks"`
	DatabaseURL string                   `json:"databaseURL"`
	Api         ApiConfig                `json:"api"`
}

type WorkerConfig struct {
	RetryDelayMs int `json:"retryDelayMs"`
	MaxDelayMs   int `json:"maxDelayMs"`
}

type NetworkConfig struct {
	StartBlock              int                 `json:"startBlock"`
	ERC20TransferTopic      string              `json:"erc20TransferTopic"`
	RPCUrl                  string              `json:"rpcURL"`
	ChainID                 int                 `json:"chainId"`
	Name                    string              `json:"name"`
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
	Port string `json:"port"`
}

type RPCRateLimitConfig struct {
	RPS           int `json:"rps"`
	Burst         int `json:"burst"`
	MaxRetryCount int `json:"maxRetryCount"`
	RetryDelayMs  int `json:"retryDelayMs"`
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

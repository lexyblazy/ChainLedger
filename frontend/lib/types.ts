export interface StatusResponse {
  data: {
    networks: Record<string, NetworkStatus>;
  };
}

export interface NetworkStatus {
  name: string;
  chainId: number;
  startBlock: number;
  bestRpcBlockNumber: number;
  ingestedBlocksCount: number;
  ingestedBlocksMaxBlockNumber: number;
  tokensCount: number;
  ingestionLagBlocksCount: number;
  ingestionProgressPct: number;
  addressesCount: number;
}

interface OffsetPaginationMeta {
  offset: number;
  limit: number;
  total: number;
}

export interface WalletsResponse {
  data: {
    wallets: Wallet[];
  };
  meta: OffsetPaginationMeta;
}

export interface Wallet {
  address: string;
  chain_id: number;
  entity_type: string;
  label: string;
  created_at: string;
  updated_at: string;
}

export interface PortfolioResponse {
  data: {
    portfolio: Portfolio[];
  };
  meta: OffsetPaginationMeta;
}

export interface Portfolio {
  asset_type: string;
  asset_address: string;
  balance_raw: string;
  symbol: string;
  decimals: number;
}

-- Address Registry
CREATE TABLE address_registry (
    address         TEXT NOT NULL,
    chain_id        INTEGER NOT NULL,
    entity_type     TEXT,
    label           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at		TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    primary key (chain_id,address)
);

CREATE INDEX idx_address_registry_chain_id ON address_registry (chain_id);

CREATE INDEX idx_address_registry_entity_type ON address_registry (entity_type);

-- Blocks
CREATE TABLE blocks (
    block_number    BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    chain_id        INTEGER NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (block_number, chain_id)
);

-- ERC-20 Transfer Events
CREATE TABLE erc20_transfers (
    tx_hash         TEXT NOT NULL,              -- transaction hash
    log_index       INTEGER NOT NULL,            -- log index within tx
    block_number    BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    from_address    TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    token_address   TEXT NOT NULL,
    amount_raw      NUMERIC(78, 0) NOT NULL,     -- uint256-safe
    chain_id        INTEGER NOT NULL,             -- e.g. 1 = Ethereum mainnet
    ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tx_hash, chain_id, log_index)
);

CREATE INDEX idx_erc20_from_address ON erc20_transfers (from_address);
CREATE INDEX idx_erc20_to_address ON erc20_transfers (to_address);
CREATE INDEX idx_erc20_token_address ON erc20_transfers (token_address);
CREATE INDEX idx_erc20_block_number ON erc20_transfers (block_number);
CREATE INDEX idx_erc20_block_timestamp ON erc20_transfers (block_timestamp);

-- native transfers
CREATE TABLE native_transfers (
    tx_hash         TEXT NOT NULL,             -- unique per native transfer
    block_number    BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    from_address    TEXT NOT NULL,
    to_address      TEXT NOT NULL,
    amount_raw      NUMERIC(78, 0) NOT NULL,      -- wei
    chain_id        INTEGER NOT NULL,              -- e.g. 1 = Ethereum mainnet
    ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tx_hash, chain_id)
);

CREATE INDEX idx_native_from_address ON native_transfers (from_address);
CREATE INDEX idx_native_to_address ON native_transfers (to_address);
CREATE INDEX idx_native_block_number ON native_transfers (block_number);
CREATE INDEX idx_native_block_timestamp ON native_transfers (block_timestamp);

-- Token Metadata
CREATE TABLE tokens (
    chain_id                  INTEGER NOT NULL,
    token_address             TEXT NOT NULL,
    symbol                    TEXT,
    name                      TEXT,
    decimals                  INTEGER,
    first_seen_block          BIGINT,
    metadata_fetch_failed_at  TIMESTAMPTZ,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chain_id, token_address)
);

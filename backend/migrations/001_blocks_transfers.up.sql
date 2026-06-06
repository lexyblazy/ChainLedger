-- blocks
CREATE TABLE blocks (
    chain_id INTEGER not null,
    block_number BIGINT not null,
    block_hash TEXT not null,
    parent_hash TEXT not null,
    timestamp TIMESTAMPTZ not null,
    canonical BOOLEAN not null default true,
    created_at TIMESTAMPTZ not null default now(),
    updated_at TIMESTAMPTZ not null default now(),
    PRIMARY KEY (chain_id, block_number)
);

-- transfers
-- transfers will support both native and erc20 transfers
CREATE TABLE transfers (
    chain_id INTEGER not null,
    block_number BIGINT not null,
    block_timestamp TIMESTAMPTZ not null,
    tx_hash TEXT not null,
    source_type TEXT not null, -- tx_value | log 
    source_index BIGINT not null, -- transaction_index for native transfers, log_index for erc20 transfers
    from_address TEXT not null,
    to_address TEXT not null,
    amount_raw NUMERIC(78, 0) not null,
    canonical BOOLEAN not null default true,
    created_at TIMESTAMPTZ not null default now(),
    updated_at TIMESTAMPTZ not null default now(),
    token_address TEXT null, -- nullable for native transfers

    PRIMARY KEY (chain_id, tx_hash, source_type, source_index)
);

CREATE INDEX idx_transfers_chain_id_from_address ON transfers (chain_id, from_address);
CREATE INDEX idx_transfers_chain_id_to_address ON transfers (chain_id, to_address);
CREATE INDEX idx_transfers_chain_id_token_address ON transfers (chain_id, token_address);
CREATE INDEX idx_transfers_chain_id_tx_hash ON transfers (chain_id, tx_hash);

-- transaction receipts
CREATE TABLE transaction_receipts (
    chain_id INTEGER NOT NULL,
    block_number BIGINT NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,

    tx_hash TEXT NOT NULL,
    tx_index BIGINT NOT NULL,

    from_address TEXT NOT NULL,
    to_address TEXT NULL,

    status SMALLINT NOT NULL, -- 1 = success, 0 = reverted/failed
    canonical BOOLEAN NOT NULL DEFAULT true,
    
    raw JSONB NOT NULL, -- raw receipt data from the blockchain
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (chain_id, tx_hash)
);

CREATE INDEX idx_transaction_receipts_chain_id_from_address ON transaction_receipts (chain_id, from_address);
CREATE INDEX idx_transaction_receipts_chain_id_to_address ON transaction_receipts (chain_id, to_address);
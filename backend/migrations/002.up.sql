-- Add indexes to native and erc20 transfers for easy asset flows query
CREATE INDEX ON native_transfers (chain_id, from_address);
CREATE INDEX ON native_transfers (chain_id, to_address);

CREATE INDEX ON erc20_transfers (chain_id, token_address, from_address);
CREATE INDEX ON erc20_transfers (chain_id, token_address, to_address);


-- Asset flows
CREATE VIEW asset_flows AS
-- Native outflow (ETH, etc.)
SELECT
  nt.chain_id,
  nt.from_address AS wallet_address,
  'native'      AS asset_type,
  NULL          AS asset_address,
  -nt.amount_raw   AS amount_raw,
  nt.block_number as block_number
FROM native_transfers nt
JOIN address_registry ar 
    ON ar.address = nt.from_address 
    AND ar.chain_id = nt.chain_id

UNION ALL

-- Native inflows (ETH, etc.)
SELECT
  nt.chain_id,
  nt.to_address   AS wallet_address,
  'native'     AS asset_type,
  NULL         AS asset_address,
  nt.amount_raw   AS amount_raw,
  nt.block_number as block_number
FROM native_transfers nt
JOIN address_registry ar 
    ON ar.address = nt.to_address 
    AND ar.chain_id = nt.chain_id

UNION ALL

-- ERC-20 outflow
SELECT
  et.chain_id,
  et.from_address AS wallet_address,
  'erc20'      AS asset_type,
  et.token_address AS asset_address,
  -et.amount_raw  AS amount_raw,
  et.block_number as block_number
FROM erc20_transfers et
JOIN address_registry ar 
    ON ar.address = et.from_address 
    AND ar.chain_id = et.chain_id

UNION ALL

-- ERC-20 inflow
SELECT
  et.chain_id,
  et.to_address   AS wallet_address,
  'erc20'      AS asset_type,
  et.token_address AS asset_address,
  et.amount_raw   AS amount_raw,
  et.block_number as block_number
FROM erc20_transfers et
JOIN address_registry ar 
    ON ar.address = et.to_address 
    AND ar.chain_id = et.chain_id
;

-- balances table
CREATE TABLE balances (
    id                  SERIAL PRIMARY KEY,
    chain_id            BIGINT NOT NULL,
    wallet_address      TEXT   NOT NULL,
    asset_type          TEXT   NOT NULL,
    asset_address       TEXT,  -- NULL for native assets
    balance_raw         NUMERIC(78, 0) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (wallet_address, chain_id,asset_type, asset_address)
)

CREATE TABLE balance_snapshots (
    id                  SERIAL PRIMARY KEY,
    chain_id            BIGINT NOT NULL,
    wallet_address      TEXT   NOT NULL,

    asset_type          TEXT   NOT NULL,
    asset_address       TEXT,  -- NULL for native assets

    balance_raw         NUMERIC(78, 0) NOT NULL,

    block_number      BIGINT NOT NULL,
    block_timestamp  TIMESTAMPTZ NOT NULL,

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

   UNIQUE (
        chain_id,
        wallet_address,
        asset_type,
        asset_address,
        block_number
    )
);

-- Latest snapshot per address / asset
CREATE INDEX balance_snapshots_latest_idx ON balance_snapshots (
    chain_id,
    wallet_address,
    asset_type,
    asset_address,
    block_number DESC
);

-- Time-series queries (charts)
CREATE INDEX balance_snapshots_time_idx ON balance_snapshots (
    chain_id,
    wallet_address,
    block_timestamp
);

--
CREATE INDEX balance_snapshots_wallet_address_chain_id_id_desc_idx ON balance_snapshots (wallet_address, chain_id, id DESC);

CREATE INDEX balance_snapshots_wallet_chain_asset_id_desc
ON balance_snapshots (wallet_address, chain_id, asset_address, id DESC);
# Multichain Crypto Asset Ingestion & Portfolio Backend

A **fund-grade, multichain backend system** for ingesting on-chain crypto asset data, normalizing it for accounting and reporting workflows, and exposing portfolio and exposure APIs for internal use.

This project is a **targeted backend engineering application** aligned with Polychain’s Accounting and Reporting Systems, emphasizing correctness, determinism, and operational reliability over speculative features.

---

## Motivation

Crypto funds require:

* reliable extraction of on-chain activity
* consistent normalization across chains
* queryable portfolio state
* historical balance tracking
* clear separation between raw ledger data and derived accounting views

This system is designed to serve as a **data foundation for portfolio monitoring, vesting analysis, and counterparty exposure**, not a public blockchain explorer.

---

## Scope & Guarantees

**In scope**

* On-chain ingestion (EVM chains)
* Address-scoped indexing
* Deterministic balance computation
* Historical balance snapshots
* SQL-friendly reporting tables
* Multichain support by design

**Explicitly out of scope**

* Reorg-aware reconciliation
* Price feeds / USD valuation
* Trading or execution logic
* Authentication & permissions

---

## Supported Networks

* **Ethereum Mainnet**

  * Historical backfill from a fixed start block
* **Base**

  * Forward indexing near chain head

Additional EVM networks (e.g. Arbitrum, Optimism) can be added with configuration only.

---

## System Architecture

### Ingestion Layer

* One **network worker per chain**
* Sequential block ingestion
* Address-scoped filtering
* Head-lag safety window to avoid unstable blocks
* Rate-limit aware RPC access (token bucket + backoff)
* Idempotent database writes

The ingestion pipeline is designed to be:

* predictable
* restart-safe
* provider-agnostic
* cost-aware (especially on free RPC tiers)

---

### Storage & Normalization (Postgres)

The database is organized into **three conceptual layers**.

---

### 1. Canonical Ledger (Append-Only)

These tables represent raw on-chain facts and are never mutated.

* `blocks`
* `native_transfers`
* `erc20_transfers`

They form the immutable source of truth.

---

### 2. Derived State (Current Balances)

#### `balances`

* One row per `(chain_id, wallet_address, asset_type, asset_address)`
* Composite primary key (no surrogate ID)
* Represents **current portfolio state**
* Native assets use a sentinel `asset_address = 'native'`
* Fully rebuildable from ledger data

This table exists to make portfolio queries fast and simple.

---

### 3. Historical State (Snapshots)

#### `balance_snapshots`

* Append-only
* Periodic snapshots every N processed blocks
* Cursor-based pagination via `BIGSERIAL id`
* Designed for charts, reporting, and historical analysis

Snapshot cadence is configurable and block-based (not wall-clock-based).

---

### Supporting Tables

* `tokens` — ERC-20 metadata (symbol, decimals)
* `address_registry` — tracked addresses with optional labels (funds, protocols, counterparties)

---

## Ingestion Strategy

### Address-Scoped Indexing

Only transactions and logs involving tracked addresses are persisted.
This avoids unnecessary data growth and keeps the system aligned with fund-specific concerns.

---

### Head Lag Safety Window

The indexer maintains a configurable gap from the RPC head (e.g. 200–500 blocks):

* avoids indexing unstable blocks
* prevents RPC race conditions
* removes the need for reorg handling within scope

---

### RPC Rate Limiting

* Token-bucket rate limiter per network
* Designed to respect **account-wide RPC limits** (e.g. Alchemy free tier)
* Conservative tuning to avoid 429s
* Exponential backoff on transient failures

---

## API Surface

The HTTP API is intentionally minimal and read-only, built using Go’s standard library.

---

### Network Status

```http
GET /status
```

Returns ingestion progress and indexing state per network.

---

### Wallet Registry

```http
GET  /wallets
POST /wallets
```

* List tracked wallets
* Add or update tracked addresses dynamically
* New addresses join ingestion at the current block height

---

### Portfolio (Primary Read Model)

```http
GET /wallets/{address}/portfolio?chain_id=1
```

Returns the **current portfolio state** for a wallet on a specific chain, enriched with token metadata.

This is the canonical consumer-facing view of balances.

---

### Balance Snapshots

```http
GET /wallets/{address}/balance-snapshots?chain_id=1&limit=50&cursor=123
```

* Historical balances
* Cursor-based pagination
* Ordered deterministically
* Suitable for reporting and visualization

---

### Tokens

```http
GET /tokens
```

Returns known ERC-20 metadata.

---

## Configuration

All networks are configured declaratively. see [config.example.json](config.example.json)

---

## Performance Characteristics

* Go process memory: **~20–25 MB** under sustained ingestion
* Postgres memory stable after snapshot tuning
* Storage growth dominated by transfer tables
* Balances and snapshots remain compact and queryable

The system favors **predictable performance and correctness** over raw ingestion speed.

---

## Design Principles (Intentional)

* **Ledger first, state second**
  Raw data is immutable; derived state is rebuildable.

* **Block-based determinism**
  Snapshots and ingestion are block-driven, not time-driven.

* **Chain isolation**
  All data is explicitly scoped by `chain_id`.

* **Minimal dependencies**
  Go stdlib for HTTP; pgx for Postgres.

* **Operational realism**
  Rate limits, RPC quirks, and head instability are handled explicitly.

---

## What This Demonstrates (Polychain-Relevant)

* Fund-grade on-chain data ingestion
* Multichain normalization
* SQL-first reporting design
* Address-scoped portfolio accounting
* Operational reliability under RPC constraints
* Clear separation between events, state, and history

---

## Future Work (Out of Scope)

* Price feeds and valuation
* PnL and cost basis
* Reorg reconciliation
* Authentication / permissions
* Frontend dashboards (Next.js planned)

---

## License

MIT / Demonstration Use

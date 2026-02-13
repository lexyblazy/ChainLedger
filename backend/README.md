# Multichain Crypto Asset Ingestion & Portfolio Backend

A **fund-grade, multichain backend system** for ingesting on-chain asset activity, normalizing it into deterministic relational models, and exposing portfolio and historical balance APIs for internal reporting.

The system prioritizes **correctness, determinism, and operational reliability** over feature breadth.

---

## Motivation

Crypto funds require:

* Reliable extraction of on-chain activity
* Deterministic portfolio state
* Queryable historical balances
* Clear separation between raw ledger data and derived accounting views
* Infrastructure that tolerates RPC limits and instability

This backend is designed as a **data foundation for portfolio monitoring, reporting, and exposure analysis** — not a public explorer or trading engine.

---

## Scope & Guarantees

### In Scope

* EVM chain ingestion
* Address-scoped indexing
* Deterministic balance computation
* Historical balance snapshots
* SQL-optimized reporting schema
* Multichain support via configuration

### Explicitly Out of Scope

* Reorg-aware reconciliation
* Price feeds / USD valuation
* Trading or execution logic
* Authentication / authorization

The system favors architectural clarity over speculative features.

---

## Supported Networks

* **Ethereum Mainnet**

  * Historical backfill from fixed start block
* **Base**

  * Forward indexing near chain head

Additional EVM networks can be added through configuration only.

---

# System Architecture

The backend separates ingestion (write model) from read APIs (read model).

```
RPC Providers
        ↓
Network Worker (per chain)
        ↓
Postgres (Ledger + Derived State)
        ↓
Read API (stateless)
```

### Write Model (Ingestion)

* One worker per configured network
* Sequential block processing
* Address-scoped filtering
* Configurable head-lag safety window
* Rate-limit aware RPC client
* Idempotent writes
* Restart-safe

Each chain runs independently via configuration.

---

# Storage Model (Postgres)

The schema is intentionally layered.

---

## 1️⃣ Canonical Ledger (Append-Only)

Immutable on-chain facts:

* `blocks`
* `native_transfers`
* `erc20_transfers`

These tables are never mutated and serve as the source of truth.

---

## 2️⃣ Derived State (Current Portfolio)

### `balances`

* One row per `(chain_id, wallet_address, asset_type, asset_address)`
* Composite primary key
* Represents current portfolio state
* Native assets use sentinel `asset_address = 'native'`
* Fully rebuildable from ledger

Optimized for fast portfolio queries.

---

## 3️⃣ Historical State (Snapshots)

### `balance_snapshots`

* Append-only
* Block-based snapshot cadence
* Cursor pagination via `BIGSERIAL id`
* Deterministic ordering
* Suitable for reporting and visualization

Snapshots are block-driven, not wall-clock driven.

---

## Supporting Tables

* `tokens` — ERC-20 metadata
* `address_registry` — tracked addresses

---

# Ingestion Strategy

## Address-Scoped Indexing

Only transactions involving tracked addresses are persisted.

This prevents unnecessary data growth and aligns ingestion with fund-specific exposure tracking.

---

## Head-Lag Safety Window

The worker maintains a configurable gap from RPC head (e.g., 200–500 blocks):

* Avoids unstable blocks
* Eliminates reorg handling within defined scope
* Ensures deterministic state

---

## RPC Layer

* Token-bucket rate limiting per network
* Account-aware request budgeting
* Exponential backoff for transient failures
* Provider-agnostic design

The RPC client encapsulates transport, retry, and rate limiting concerns.

---

# API Surface (Read Model)

The API is intentionally minimal and read-only, implemented with Go’s standard library.

---

## Network Status

```http
GET /status
```

Returns ingestion progress and sync state per network.

---

## Wallet Registry

```http
GET  /wallets
POST /wallets
```

* List tracked wallets
* Add/update tracked addresses
* Newly added addresses begin indexing at current chain height

---

## Portfolio (Primary Read Model)

```http
GET /wallets/{address}/portfolio?chain_id=1
```

Returns current portfolio state for a wallet on a specific chain, enriched with token metadata.

This endpoint represents the canonical balance view.

---

## Balance Snapshots

```http
GET /wallets/{address}/balance-snapshots?chain_id=1&limit=50&cursor_id=123
```

* Historical balances
* Cursor-based pagination
* Deterministic ordering

---

## Tokens

```http
GET /tokens
```

Returns known ERC-20 metadata.

---

# Configuration

All network behavior is declarative via `config.json`.

Per-network configuration includes:

* RPC URL
* Rate limits
* Block gap
* Snapshot cadence
* Token discovery strategy

Adding a new EVM chain requires no schema changes.

---

# Performance Characteristics

The system is designed for predictable, bounded resource usage under sustained ingestion.

## Memory Profile

Observed under sustained ingestion:

* **Native execution (single chain):** ~25 MB RSS
* **Containerized execution:** ~50–60 MB RSS

Memory usage scales primarily with:

* Snapshot cadence (more frequent snapshots increase allocation pressure)
* Number of active chains
* Address registry size
* RPC batching behavior

Memory remains **stable over time** during extended runs, with no unbounded growth observed.

---

## CPU Usage

CPU usage is modest and workload-dependent:

* Low during head polling
* Spikes during block processing and snapshot writes
* Scales linearly with number of configured chains

In steady-state (near-head indexing), CPU remains low due to head-gap buffering.

---

## Database Characteristics

Storage growth is dominated by append-only ledger tables:

* `native_transfers`
* `erc20_transfers`

Derived tables (`balances`, `balance_snapshots`) remain compact relative to transfer volume.

Snapshot storage growth is controlled by:

* Block-based snapshot interval
* Address set size
* Chain activity level

All derived state is rebuildable from ledger data.

---

## Ingestion Throughput

Throughput depends on:

* RPC provider limits
* Configured rate limits
* Block density (transfer-heavy vs quiet blocks)

The system prioritizes:

* Deterministic ordering
* Idempotency
* RPC stability

over maximum indexing speed.

---

## Horizontal Characteristics

Each network runs in isolation.

Adding additional chains:

* Increases memory and CPU linearly
* Does not introduce shared contention across workers
* Does not affect determinism of other networks

---

## Operational Stability

The system has been tested under:

* Long-running ingestion
* Restart scenarios
* Snapshot-intensive configurations

Observed properties:

* Restart-safe behavior
* Stable memory footprint
* No unbounded resource growth
* Deterministic state reconstruction



---

# Design Principles

**Ledger First, State Second**
Raw on-chain data is immutable; derived state is rebuildable.

**Block-Based Determinism**
Ingestion and snapshots are driven by block height.

**Chain Isolation**
All data is scoped by `chain_id`.

**Minimal Dependencies**
Go stdlib + pgx; Postgres.

**Operational Realism**
RPC limits and provider instability are handled explicitly.

---

# What This Demonstrates

* Deterministic multichain ingestion
* Address-scoped portfolio accounting
* Clear separation between ledger, state, and history
* SQL-first reporting schema
* Operational resilience under constrained RPC environments
* Clean separation of write and read models

---

# Future Extensions (Out of Scope)

* Valuation & pricing
* Cost basis / PnL
* Reorg reconciliation
* Authentication layer
* Distributed worker scaling

---

# License

MIT / Demonstration Use


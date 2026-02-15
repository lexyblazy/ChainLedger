# ChainLedger v1 — Architecture Overview

## Philosophy

ChainLedger v1 is a multichain blockchain indexing and reporting backend designed to:

* Ingest native and ERC20 transfers
* Normalize wallet inflows/outflows
* Maintain up-to-date balances
* Provide portfolio reporting APIs
* Support dynamic wallet registration

The system prioritizes operational simplicity and deterministic aggregation, without implementing canonical reorg modeling or event-sourced ledger mechanics.

---

# 1. High-Level Architecture

```
Per-Chain Block Ingestion Workers
        ↓
Transfer Extraction (Native + ERC20)
        ↓
Normalized Asset Flow View
        ↓
Mutable Balance Accumulation
        ↓
Balance Snapshots (Raw Balances)
        ↓
Read-Only Reporting API
```

Core characteristics:

* Chain-scoped state via `chain_id`
* Deterministic relational modeling
* Incremental balance accumulation
* No canonical chain modeling
* No checkpoint-based replay system

---

# 2. Wallet Registration Layer

ChainLedger v1 supports dynamic wallet onboarding via API.

### Endpoint Behavior

* Wallets can be added at runtime.
* Wallet is stored in `address_registry`.
* Ingestion pipeline begins tracking transfers for registered wallets.

This enables:

* On-demand portfolio tracking
* User-driven wallet monitoring
* Incremental system growth

Wallet scope is enforced via joins against `address_registry`.

---

# 3. Block Ingestion Layer

Each supported EVM network runs an independent worker.

### Responsibilities

* Fetch new blocks sequentially
* Process transactions within each block
* Extract native and ERC20 transfer events
* Insert transfer records into DB

### Characteristics

* No block_hash stored
* No parent_hash stored
* No canonical flag
* No explicit reorg detection
* No rollback strategy

The system assumes chain stability.

---

# 4. Transfer Extraction Layer

Two primary base tables:

* `native_transfers`
* `erc20_transfers`

### Fields (Simplified)

* `chain_id`
* `block_number`
* `transaction_index`
* `log_index` (ERC20)
* `from_address`
* `to_address`
* `amount_raw`

Transfers are treated as immutable records once inserted.

Gas fees are not modeled explicitly.

---

# 5. Asset Flow Normalization

`asset_flows` is a SQL view that unifies inflow/outflow logic:

### Native Outflow

* Negative amount for sender

### Native Inflow

* Positive amount for receiver

### ERC20 Outflow

* Negative amount for sender

### ERC20 Inflow

* Positive amount for receiver

This view:

* Normalizes accounting semantics
* Filters only tracked wallets via `address_registry`
* Provides unified aggregation surface

Asset flows are computed dynamically (not materialized).

---

# 6. Balance Accumulation Layer

Balances are stored in a mutable `balances` table.

Balances are updated incrementally per block:

```sql
INSERT INTO balances (...)
SELECT SUM(amount_raw)
FROM asset_flows
WHERE block_number = $1
GROUP BY wallet, asset
ON CONFLICT DO UPDATE
SET balance_raw = balances.balance_raw + EXCLUDED.balance_raw;
```

### Properties

* Balances are mutable.
* State accumulates forward.
* Block provenance is not embedded in balances.
* No checkpoint anchoring.
* No deterministic replay model.

Balances represent “latest known state”.

---

# 7. Balance Snapshots

Snapshots capture raw balances at a point in time.

### Properties

* Derived from balances table
* Store raw balance only
* No pricing metadata
* Used for portfolio reporting
* Not used for replay or reconciliation

Snapshots are business-layer artifacts.

---

# 8. Historical Balance Capability

v1 does not support:

* Arbitrary block-based balance queries
* Time-travel projection
* Deterministic historical reconstruction

Only current accumulated balance is maintained.

Full rebuild requires summing `asset_flows` from genesis.

---

# 9. Finality and Reorg Handling

v1 includes:

* Configurable finality depth buffer (e.g., 500 blocks)

However:

* No parent-hash validation
* No canonical chain tracking
* No block-level rollback
* No canonical flags

Reorgs are not explicitly modeled.

System assumes practical chain stability.

---

# 10. Read Model / API Layer

The API is read-only for balances and reporting.

Supports:

* Portfolio retrieval
* Paginated queries
* Wallet-level balance lookup
* Dynamic wallet registration

Read model depends directly on `balances` and `balance_snapshots`.

---

# 11. Operational Characteristics

* Fully Dockerized environment
* Restart-safe ingestion
* RPC rate limiting
* Head-gap buffering
* Deterministic environment reproducibility
* Clear write/read separation

Operational stability prioritized over advanced ledger modeling.

---

# 12. Strengths of v1

* Clean separation between ingestion and API
* Deterministic relational normalization
* Multi-chain support
* Dynamic wallet onboarding
* Simple mental model
* Production-ready ingestion pipeline
* Operational clarity

---

# 13. Architectural Limitations

v1 does not implement:

* Canonical block modeling
* Reorg detection or rollback
* Immutable ledger delta layer
* Checkpoint-based projection
* Deterministic replay
* Finalized vs pending separation
* Gas accounting
* Block-hash-level validation

Balances are mutable state, not projections of immutable deltas.

---

# 14. v1 Design Intent

v1 validates:

* Transfer ingestion correctness
* Asset normalization logic
* Portfolio computation viability
* Operational deployment model

It serves as a functional indexing and reporting backend.

v2 introduces formal ledger semantics and event-sourced modeling.


# ChainLedger v2 — Architecture Overview

Status: Design Draft
Branch: feature/v2-ledger-architecture

## Philosophy

ChainLedger v2 evolves from a mutable balance model into a deterministic, event-sourced ledger engine designed for blockchain-based financial systems.

Core principles:

* Blocks are the root domain.
* Canonicality is explicit and versioned.
* State is derived, never mutated.
* Balances are projections of immutable deltas.
* Reorg safety is first-class.
* Checkpoints bound computation.
* Snapshots serve business/reporting needs.
* Cache is performance optimization only.

---

# 1. Layered Architecture

```
Blocks (Canonical Control)
        ↓
Transfers (Immutable On-Chain Facts)
        ↓
Ledger Deltas (Event-Sourced Accounting Layer)
        ↓
Balance Checkpoints (Projection Anchors)
        ↓
Projection Engine (Deterministic Computation)
        ↓
Balance Snapshots (Business / Reporting Artifacts)
        ↓
Cache Layer (Optional Performance Acceleration)
        ↓
API Layer
```

Each layer has a strict responsibility and does not leak concerns into adjacent layers.

---

# 2. Canonical Block Layer

The block layer governs reorg detection and canonical state tracking.

## Responsibilities

* Store block metadata
* Detect reorg via parent hash mismatch
* Maintain canonical chain history
* Provide finalized head reference

## Core Fields

* `block_number`
* `block_hash`
* `parent_hash`
* `timestamp`
* `canonical BOOLEAN`

## Reorg Handling

On parent hash mismatch:

1. Identify divergence block.
2. Mark blocks ≥ divergence as `canonical = false`.
3. Cascade canonical invalidation to dependent layers.

Blocks are the root of all domain state.

---

# 3. Transfer Layer (Immutable Facts)

Transfers represent raw on-chain activity.

Includes:

* Native transfers
* ERC20 transfers

## Required Fields

* `block_number`
* `block_hash`
* `transaction_hash`
* `transaction_index`
* `log_index`
* `from_address`
* `to_address`
* `amount_raw`
* `canonical BOOLEAN`

## Rules

* Never mutate transfer amounts.
* Only canonical flag changes during reorg.
* Fully ordered and replayable.

---

# 4. Ledger Delta Layer (Core Accounting Engine)

Ledger deltas convert transfer events into signed accounting entries.

Each transfer generates:

* One negative delta (sender)
* One positive delta (receiver)

## Required Fields

* `block_number`
* `block_hash`
* `transaction_index`
* `log_index`
* `wallet_address`
* `asset_type`
* `asset_address`
* `delta_raw` (signed)
* `canonical BOOLEAN`

## Properties

* Immutable
* Deterministic ordering
* Replayable from genesis
* Canonical-aware

Ledger deltas are the source of truth for balance computation.

---

# 5. Balance Checkpoints (Projection Anchors)

Checkpoints bound replay cost and prevent summing from genesis.

They are engine-level optimization constructs.

## Granularity

* Per wallet
* Per asset

## Creation Strategy

* Every N blocks (e.g., 1000)
* Only beyond finality depth
* Only from canonical blocks

## Purpose

* Anchor projection window
* Reduce replay cost
* Enable scalable historical queries

## Reorg Handling

If divergence block < latest checkpoint:

1. Delete checkpoints ≥ divergence block.
2. Recompute forward from last safe checkpoint.

Checkpoints are computational artifacts, not business artifacts.

---

# 6. Projection Engine (Deterministic Balance Computation)

Balances are projections of ledger deltas.

To compute balance at block B:

1. Find latest checkpoint ≤ B.
2. Sum canonical deltas where:

   * `block_number > checkpoint_block`
   * `block_number ≤ B`
3. Compute:

```
balance = checkpoint_balance + forward_delta_sum
```

## Ordering Guarantee

Replay must respect:

```
ORDER BY block_number, transaction_index, log_index
```

## Properties

* Deterministic
* Reorg-safe
* Replayable at any block

Balances are never treated as mutable state.

---

# 7. Finalized vs Pending State

ChainLedger v2 explicitly separates authoritative state from volatile state.

## Finalized Balance (Authoritative)

* Block ≤ finalized_head
* Canonical-only
* Checkpoint-anchored
* Reorg-safe
* Default API response

## Pending Delta (Advisory)

Computed on-demand:

```
SUM(delta_raw
    WHERE block_number > finalized_head
    AND canonical = true)
```

Characteristics:

* Volatile
* Not checkpointed
* May change on reorg
* Explicitly labeled in API responses

Finalized and pending values are never merged silently.

---

# 8. Balance Snapshots (Business / Reporting Layer)

Snapshots represent projected balances at specific reporting points.

They are not used in projection or replay.

## Purpose

* End-of-day portfolio view
* Monthly statements
* Institutional reporting
* Historical NAV tracking
* User-requested snapshots

## Snapshot Properties

* Created from finalized projection only
* Immutable once created
* May include pricing metadata
* Exposed via API
* Not used for engine computation

Snapshots are business artifacts, not infrastructure anchors.

---

# 9. Cache Layer (Optional)

`current_balances` or Redis cache serves as performance optimization.

## Properties

* Derived from projection engine
* Always includes `block_number`
* Safe to delete and rebuild
* Never authoritative

## Reorg Handling

On reorg beyond finalized head:

* Flush cache
* Recompute from checkpoints

Cache must not contain state that cannot be rebuilt.

---

# 10. Reorg Strategy Summary

On divergence detection:

1. Mark blocks ≥ divergence as non-canonical.
2. Mark transfers ≥ divergence as non-canonical.
3. Mark ledger deltas ≥ divergence as non-canonical.
4. Delete checkpoints ≥ divergence.
5. Reinsert canonical chain data.
6. Recompute projection forward.
7. Flush cache if necessary.

Finalized balances remain stable.
Pending absorbs volatility.

---

# 11. Improvements Over v1

v1:

* Mutable balance updates
* Full aggregation from genesis
* No checkpoint anchors
* Limited canonical modeling

v2:

* Event-sourced ledger architecture
* Immutable signed deltas
* Block-scoped checkpoints
* Deterministic replay
* Canonical flag modeling
* Finalized/pending separation
* Cache as acceleration layer
* Clear reporting layer separation

ChainLedger v2 transitions from an indexing backend to a deterministic ledger engine.


# Polychain-Style Multichain Accounting & Reporting System

A **fund-grade, multichain crypto accounting and reporting system** built as a targeted engineering application for **Polychain Capital**.

This project demonstrates how to ingest on-chain asset data across multiple networks, normalize it into accounting-friendly relational models, and expose portfolio and historical balance views suitable for **internal fund operations, reporting, and exposure analysis**.

---

## Why This Project Exists

Polychain’s accounting and asset operations require:

* reliable extraction of on-chain activity
* consistent normalization across chains
* deterministic portfolio state
* historical balance tracking for reporting
* SQL-friendly access for dashboards and analysis tools

This project is intentionally scoped to reflect those needs — **not** to build a public explorer or trading system.

---

## What This Demonstrates (Polychain-Relevant)

* Fund-grade on-chain data ingestion
* Multichain asset normalization (EVM)
* Address-scoped portfolio accounting
* Separation of ledger data vs derived state
* Historical balance snapshots for reporting
* Operational realism (RPC limits, head lag, retries)
* SQL-first design for internal dashboards (Retool / React)

---

## Repository Structure

```
/
├── backend/        # Go + Postgres ingestion & reporting API
│   ├── README.md   # Backend architecture, data model, and ingestion strategy
│   └── ...
│
├── frontend/       # Next.js reporting dashboard
│   ├── README.md   # Frontend architecture and UI scope
│   └── ...
│
└── README.md       # (this file)
```

Each component is intentionally decoupled and can be evaluated independently.

---

## System Overview

### Backend (Core of the Project)

The backend is the primary focus and mirrors how a **fund-grade accounting backend** would be structured.

Key characteristics:

* Multichain ingestion (Ethereum Mainnet, Base)
* One worker per network (chain-scoped)
* Sequential block processing
* Address-scoped indexing
* Deterministic balance computation
* Historical balance snapshots
* SQL-friendly reporting tables
* Read-only APIs for internal dashboards

Detailed design is documented in **`/backend/README.md`**.

---

### Frontend (Reporting Interface)

The frontend consumes backend APIs to present:

* wallet portfolios
* historical balance charts
* network ingestion status

It is intentionally thin and depends entirely on backend correctness.

Details live in **`/frontend/README.md`**.

---

## Design Philosophy

This project follows principles aligned with Polychain’s operational needs:

* **Ledger first, state second**
  Raw on-chain data is immutable; balances are derived.

* **Determinism over freshness**
  Block-based ingestion and snapshots, not time-based heuristics.

* **Chain isolation**
  All data is explicitly scoped by `chain_id`.

* **Operational realism**
  RPC rate limits, head lag, and provider quirks are handled explicitly.

* **Minimal dependencies**
  Go stdlib for HTTP, Postgres for reporting, pgx for DB access.

---

## Intended Use

This system is designed to serve as:

* a foundation for internal portfolio dashboards
* a backend for accounting and reporting tools
* a base layer for exposure and vesting analysis

It is **not** intended for:

* trading or execution
* price discovery
* public APIs
* real-time market data

---

## Status

* Backend ingestion & APIs: **complete**
* Multichain ingestion: **validated**
* Portfolio and snapshot reporting: **complete**
* Frontend dashboard: **in progress / planned**
* Additional chains and data sources can be added incrementally

---

## How to Review This Project

For Polychain reviewers:

1. Start with **`/backend/README.md`**
2. Review ingestion flow and schema design
3. Review portfolio and snapshot modeling
4. Skim frontend for reporting usage
5. Ignore polish — focus on correctness and architecture

---

## License

MIT / Demonstration Use


# Polychain-Style Multichain Accounting & Reporting System

A **fund-grade, multichain crypto accounting and reporting backend** designed to reflect the operational needs of **Polychain Capital**.

This project demonstrates how to ingest on-chain activity across multiple EVM networks, normalize it into deterministic relational models, and expose portfolio and historical balance views suitable for **internal fund operations, reporting, and exposure analysis**.

---

## Purpose

Polychain’s accounting and asset operations require:

* Reliable extraction of on-chain activity
* Deterministic portfolio state
* Historical balance tracking
* SQL-friendly access for reporting and dashboards
* Operational resilience under RPC constraints

This project intentionally models those constraints.
It is not a block explorer or trading system.

---

## What This Demonstrates

* Multichain EVM ingestion (Ethereum Mainnet, Base)
* Address-scoped indexing and portfolio accounting
* Separation of raw ledger data vs derived balance state
* Deterministic historical balance snapshots
* Operational realism (RPC rate limits, head lag, retries)
* Worker/API separation (write model vs read model)
* Dockerized, reproducible infrastructure
* Backup + restore workflow validation

---

## System Architecture

```
Frontend (Next.js)
        ↓
API Container (Go - read layer)
        ↓
Postgres
        ↑
Worker Container (Go - ingestion engine)
        ↑
RPC Providers (per chain)
```

Key characteristics:

* One worker per configured network (chain-scoped goroutines)
* Deterministic block-based ingestion
* Config-driven per-chain behavior
* SQL-first schema optimized for internal dashboards
* Stateless API layer
* Graceful shutdown and retry handling

---

## Repository Structure

```
/
├── backend/        # Go + Postgres ingestion & reporting API
│   ├── README.md   # Backend architecture and ingestion design
│   └── ...
│
├── frontend/       # Next.js reporting dashboard
│   ├── README.md   # Frontend architecture and UI scope
│   └── ...
│
├── INFRA.md        # Docker, deployment, backup, restore workflow
└── README.md       # (this file)
```

Each component is intentionally decoupled and can be evaluated independently.

---

## Backend Overview (Primary Focus)

The backend models a fund-grade accounting ingestion system:

* Multichain ingestion workers
* Per-chain configurable RPC limits and retry policies
* Block-gap finality buffer
* Idempotent processing
* Derived balance tables
* Historical balance snapshots
* Read-only reporting APIs

The backend implements a lightweight CQRS-style separation:

* Worker = write model (block ingestion, balance updates)
* API = read model (portfolio, snapshots, ingestion status)
* NetworkReader = abstraction over RPC + DB reads
* RPC client = transport layer (rate limiting + retry)

See `/backend/README.md` for detailed design.

---

## Frontend Overview

The frontend provides:

* Wallet portfolio views
* Historical balance charts
* Network ingestion status

It is intentionally thin and fully dependent on backend correctness.

See `/frontend/README.md`.

---

## Infrastructure

The system is fully containerized and reproducible.

On a fresh machine:

```bash
git clone <repo>
cd <repo>
cp .env.example .env
# provide config.json
./start-dev.sh
```

If this runs successfully, the system is fully reproducible.

Details on deployment, backups, and restore procedures are documented in `INFRA.md`.

---

## Design Principles

**Ledger First, State Second**
Raw on-chain events are immutable; balances are derived.

**Determinism Over Freshness**
Blocks are processed sequentially with a configurable finality gap.

**Chain Isolation**
All data is explicitly scoped by `chain_id`.

**Operational Realism**
RPC rate limits, retries, head lag, and provider instability are handled explicitly.

**Minimal Dependencies**
Go (stdlib + pgx), Postgres, Docker.

---

## Intended Scope

Designed for:

* Internal portfolio dashboards
* Accounting and reporting backends
* Exposure and historical analysis

Not designed for:

* Trading or execution
* Real-time price feeds
* Public APIs
* Market data aggregation

---

## Status

* Backend ingestion & APIs: **Complete**
* Multichain ingestion: **Validated**
* API/Worker separation: **Complete**
* Infrastructure reproducibility: **Validated**
* Backup & restore workflow: **Tested**
* Frontend reporting interface: **In progress**

---

## How to Review

For Polychain reviewers:

1. Start with `/backend/README.md`
2. Review ingestion flow and schema design
3. Review portfolio and snapshot modeling
4. Review infrastructure in `INFRA.md`
5. Skim frontend for reporting usage

Focus on correctness, determinism, and operational structure.

---

## License

MIT / Demonstration Use


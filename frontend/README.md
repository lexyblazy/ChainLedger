# Multichain Portfolio Reporting Frontend

A lightweight **internal reporting dashboard** built with Next.js to visualize portfolio state and historical balances derived from the backend ingestion system.

This frontend is intentionally thin and depends entirely on backend correctness.

It is not designed as a public explorer or retail interface.

---

# Purpose

The frontend demonstrates how the backend’s normalized data model can power:

* Portfolio overviews per wallet
* Historical balance visualization
* Network ingestion monitoring
* Internal accounting workflows

It acts as a read-only consumer of the API.

---

# Architectural Role

```
Frontend (Next.js)
        ↓
Backend API (Go - read layer)
        ↓
Postgres (derived state)
```

The frontend:

* Does not compute balances
* Does not derive state
* Does not interact with RPC providers
* Does not mutate data

All state originates from the backend.

---

# Tech Stack

* Next.js (App Router)
* React Query (data fetching + caching)
* Tailwind CSS
* Minimal UI components
* Fully Dockerized production build

The build is configured to consume:

```
NEXT_PUBLIC_API_URL
```

at build time.

---

# Core Features

## Portfolio View

Displays:

* Native asset balances
* ERC-20 balances
* Sorted positions
* Optional filtering of zero balances
* Pagination

Data source:

```
GET /wallets/{address}/portfolio?chain_id=...
```

---

## Historical Balance Snapshots

Displays:

* Historical balance progression
* Cursor-based pagination
* Block-derived balance state

Data source:

```
GET /wallets/{address}/balance-snapshots
```

---

## Network Status

Displays:

* Ingestion progress per chain
* Sync height
* Operational visibility

Data source:

```
GET /status
```

---

# Design Principles

**Backend-Driven State**
All balances and history are computed server-side.

**Deterministic Data**
Frontend renders derived state — it does not attempt reconciliation.

**Minimal Client Logic**
React Query is used for caching and pagination only.

**Operational Simplicity**
The frontend can be rebuilt and redeployed independently of ingestion.

---

# Running Locally

From the project root:

```bash
./scripts/start-dev.sh
```

The frontend will be available at:

```
http://localhost:3000
```

It requires the backend API to be running.

---

# Production Build

The frontend is built inside Docker using:

```
NEXT_PUBLIC_API_URL
```

The API URL must be correct at build time.

In production, it is typically proxied behind Caddy.

---

# Scope Limitations

The frontend does not include:

* Authentication
* Role-based access control
* Price feeds
* PnL computation
* Execution capabilities
* Public-facing polish

It exists to demonstrate how normalized on-chain accounting data can be surfaced in a clean reporting interface.

---

# Relationship to Backend

This UI is intentionally dependent on backend correctness.

If backend balance logic changes, the frontend reflects it automatically.

The frontend is not a source of truth.

---

# Status

* Portfolio view: Implemented
* Snapshot view: Implemented
* Network status: Implemented
* UI polish: Minimal (by design)


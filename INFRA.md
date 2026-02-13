# INFRA.md

## Polychain Multichain Ingestion + API Infrastructure

This document describes how to run, deploy, migrate, back up, and restore the Polychain multichain ingestion system.

The system is fully Dockerized and requires no local Go, Node, or Postgres installation.

---

# 1️⃣ System Architecture

```
Frontend (Next.js)
    ↓
API Container (Go - api mode)
    ↓
Postgres
    ↑
Worker Container (Go - worker mode)
    ↑
RPC Providers (per chain)
```

### Containers

* `polychain_db` → PostgreSQL database
* `polychain_api` → REST API (read layer)
* `polychain_worker` → Ingestion engine (write layer)
* `polychain_frontend` → Next.js production build
* `polychain_migrate` → One-shot DB migrations
* `polychain_backup` → Automated daily DB backups

---

# 2️⃣ Prerequisites

Required:

* Docker
* Docker Compose (plugin version)
* Git

Nothing else is required locally.

---

# 3️⃣ Environment Variables

Create a `.env` file in the project root.

Minimum required variables:

```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=yourpassword
POSTGRES_DB=polychain
DATABASE_URL=postgres://postgres:yourpassword@db:5432/polychain?sslmode=disable
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Notes:

* `DATABASE_URL` must reference `db` as hostname (Docker network).
* `NEXT_PUBLIC_API_URL` must point to the API service.

You may provide a `.env.example` in the repository.

---

# 4️⃣ Running in Development

Start everything:

```bash
./scripts/start-dev.sh
```

This will:

* Build images
* Start Postgres
* Start API
* Start Worker
* Start Frontend
* Bind ports locally

Default exposed ports (dev override):

* API → [http://localhost:8080](http://localhost:8080)
* Frontend → [http://localhost:3000](http://localhost:3000)
* Postgres → localhost:5433

Stop everything:

```bash
docker compose down
```

---

# 5️⃣ Production Deployment (Hetzner VM)

Start production stack:

```bash
./scripts/start-prod.sh
```

Production characteristics:

* Postgres is NOT exposed publicly.
* API binds to `127.0.0.1:8080`
* Frontend binds to `127.0.0.1:3000`
* Caddy (or reverse proxy) should proxy public traffic.

Backups are stored on disk at:

```
./backups/
```

---

# 6️⃣ Database Migrations

Migrations use `migrate/migrate`.

Run manually:

```bash
docker compose run --rm migrate
```

Migrations:

* Are idempotent
* Must run before API/Worker if schema changed
* Should never be modified retroactively

---

# 7️⃣ Backup Strategy

The `polychain_backup` container:

* Runs every 12 hours `pg_dump`
* Stores compressed custom-format dumps
* Retains last 7 days

Backup location:

```
./backups/backup_YYYYMMDD_HHMMSS.dump
```

Backups are stored on host disk, not inside Docker volumes.

---

# 8️⃣ Restore Procedure

Use the restore script:

```bash
./restore-db.sh backup_YYYYMMDD_HHMMSS.dump
```

The script:

1. Stops API
2. Terminates DB connections
3. Drops database
4. Recreates database
5. Restores dump
6. Restarts API

Always test restore after modifying backup strategy.

---

# 9️⃣ Service Responsibilities

### polychain_api

* Runs in `api` mode
* Stateless
* Exposes REST endpoints
* Depends on `NetworkReader`
* Does NOT perform ingestion

---

### polychain_worker

* Runs in `worker` mode
* Spins per-chain goroutines
* Uses per-chain config
* Respects block gap (`finality depth`)
* Uses exponential backoff and retry
* Writes to DB only

---

### polychain_db

* Stores:

  * Transfers
  * Balances
  * Snapshots

Never expose publicly.

---

### polychain_migrate

* Applies SQL migrations
* Runs once
* Not long-running

---

### polychain_backup

* Performs bi-daily (every 12h) `pg_dump`
* Prunes old backups

---

### polychain_frontend

* Production Next.js build
* Uses build-time `NEXT_PUBLIC_API_URL`

---

# 🔟 Operational Rules

* Never modify database schema manually.
* Always use migrations.
* Never expose Postgres publicly.
* Always test restore after changing backup config.
* Worker must not mutate schema.
* Reader layer must remain stateless.
* RPC retries must remain in RPC client layer.
* Shutdown uses independent timeout context intentionally.

---

# 11️⃣ Clean Reproducibility Test

To validate infrastructure integrity:

On a fresh machine:

```bash
git clone <repo>
cd <repo>
cp .env.example .env
./scripts/start-dev.sh
```

If this works without manual intervention, the system is reproducible.

This test should pass before production deployment.

---

# 12️⃣ Architectural Notes

The system implements a lightweight CQRS pattern:

* Worker = write model
* API = read model
* NetworkReader = abstraction layer
* RPC client = transport + retry layer

Worker and API are isolated at container level.

Each chain is independently configurable via `backend/config.json`. see `backend/config.example.json`
Note: `config.json` is a required artifact for the backend Worker and API containers.
if unprovided, the container is killed via `log.Fatal`

---

# 13️⃣ Future Extensions (Optional)

Not currently required:

* Blue/green deployments
* Horizontal worker scaling
* Distributed locking
* Prometheus metrics
* Multi-node ingestion

Current design supports future evolution if required.

---

# Final Statement

This system is designed to be:

* Deterministic
* Reproducible
* Backup-safe
* Operationally predictable
* Container-native

Changes to infrastructure should preserve these properties.


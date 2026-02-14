#!/usr/bin/env bash

set -e

echo "🚀 Starting ChainLedger (PROD mode)..."

COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"

echo "🔧 Building images..."
docker compose $COMPOSE_FILES build

echo "🗄️ Starting database..."
docker compose $COMPOSE_FILES up -d db

echo "📦 Running migrations..."
docker compose $COMPOSE_FILES run --rm migrate

echo "🔄 Starting API + Worker + Frontend..."
docker compose $COMPOSE_FILES up -d --build

echo "✅ Production environment is up."

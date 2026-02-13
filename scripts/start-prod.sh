#!/usr/bin/env bash

set -e

echo "Starting ChainLedger (PROD mode)..."

docker compose \
  -f docker-compose.yml \
  -f docker-compose.prod.yml \
  up -d db

echo "Running migrations..."
docker compose \
  -f docker-compose.yml \
  -f docker-compose.prod.yml \
  run --rm migrate

docker compose \
  -f docker-compose.yml \
  -f docker-compose.prod.yml \
  up -d

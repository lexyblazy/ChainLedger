#!/usr/bin/env bash

set -e

echo "Starting ChainLedger (DEV mode)..."

docker compose \
  -f docker-compose.yml \
  -f docker-compose.dev.yml \
  up --build

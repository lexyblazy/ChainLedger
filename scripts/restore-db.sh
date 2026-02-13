#!/usr/bin/env bash

set -e

read -p "This will wipe the database. Continue? (y/n): " CONFIRM
if [ "$CONFIRM" != "y" ]; then
  echo "Aborted."
  exit 1
fi


if [ -z "$1" ]; then
  echo "Usage: ./restore.sh <backup_filename>"
  echo "Example: ./restore.sh backup_20260213_143201.dump"
  exit 1
fi

BACKUP_FILE=$1

echo "Stopping api and worker..."
docker compose stop api worker

echo "Terminating active DB connections..."
docker compose exec db sh -c '
psql -U $POSTGRES_USER -d postgres -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname = '\''$POSTGRES_DB'\''
AND pid <> pg_backend_pid();
"'

echo "Dropping database..."
docker compose exec db sh -c 'dropdb -U $POSTGRES_USER $POSTGRES_DB'

echo "Recreating database..."
docker compose exec db sh -c 'createdb -U $POSTGRES_USER $POSTGRES_DB'

echo "Restoring from backup..."
cat backups/$BACKUP_FILE | docker compose exec -T db sh -c \
'pg_restore -U $POSTGRES_USER -d $POSTGRES_DB'


echo "Starting api and worker..."
docker compose start api worker

echo "Restore complete."

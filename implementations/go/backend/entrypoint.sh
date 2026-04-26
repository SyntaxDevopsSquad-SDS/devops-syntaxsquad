#!/bin/sh
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h "$(echo $DATABASE_URL | sed 's|.*@\([^:]*\).*|\1|')" -U "$(echo $DATABASE_URL | sed 's|.*://\([^:]*\).*|\1|')"; do
  sleep 1
done

echo "Running schema..."
psql "$DATABASE_URL" -f /app/schema.sql

exec /app/backend/whoknows

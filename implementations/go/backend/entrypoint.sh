#!/bin/sh
set -e

if [ -z "$DB_PATH" ]; then
    DB_PATH="/data/whoknows.db"
fi

if [ ! -f "$DB_PATH" ]; then
    echo "Initializing database at $DB_PATH"
    mkdir -p "$(dirname "$DB_PATH")"
    sqlite3 "$DB_PATH" < /app/schema.sql
fi

exec /app/backend/whoknows

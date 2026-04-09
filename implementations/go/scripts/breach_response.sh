#!/bin/bash
set -e

# Password breach response script
# Marks affected users to require password reset
# Usage: bash breach_response.sh

DB_PATH="${DB_PATH:-whoknows.db}"

if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file not found at $DB_PATH"
    exit 1
fi

echo "⚠️  Password Breach Response - Marking affected users for forced reset"
echo ""

# List of affected users (from breach)
AFFECTED_USERS=(
    "Viola1998"
    "Jenny1996"
    "William1978"
    "Bente1977"
    "Weena1970"
    "Ulrik2019"
    "Keld1972"
    "Nikolaj1993"
    "Boe1976"
    "Georg1983"
    "Allan1954"
    "Benjamin2015"
    "Victor1994"
    "Ole2026"
    "Otto1954"
    "Ane2025"
    "Flemming1970"
    "Rolf1958"
    "Mark1965"
    "Julius2002"
)

for username in "${AFFECTED_USERS[@]}"; do
    echo "Marking user: $username"
    sqlite3 "$DB_PATH" "UPDATE users SET force_password_reset = 1, password_reset_required_at = datetime('now') WHERE username = '$username';"
done

echo ""
echo "✅ Affected users marked for forced password reset"
echo ""
echo "Status:"
sqlite3 "$DB_PATH" "SELECT username FROM users WHERE force_password_reset = 1;"

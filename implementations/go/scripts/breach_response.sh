#!/bin/bash
set -e

# Password breach response script
# Marks affected users to require password reset
# Usage: bash breach_response.sh

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

# Build SQL statements
SQL_STATEMENTS=""
for username in "${AFFECTED_USERS[@]}"; do
    echo "Marking user: $username"
    SQL_STATEMENTS+="UPDATE users SET force_password_reset = 1, password_reset_required_at = datetime('now') WHERE username = '$username';"$'\n'
done

# Execute via docker exec
echo "Executing via Docker container..."
sudo docker exec whoknows-whoknows-1 sqlite3 /data/whoknows.db << EOF
$SQL_STATEMENTS
SELECT username FROM users WHERE force_password_reset = 1;
EOF

echo ""
echo "✅ Affected users marked for forced password reset"

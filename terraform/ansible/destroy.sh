#!/bin/bash

echo "💥 Destroyer infrastruktur..."
cd "$(dirname "$0")/.."
terraform destroy -auto-approve

echo "✅ Alt er destroyed!"
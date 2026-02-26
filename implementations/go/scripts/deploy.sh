#!/bin/bash

echo "=== Pulling latest changes from GitHub ==="
cd /opt/whoknows
GIT_SSH_COMMAND="ssh -i ~/.ssh/github_key" git pull origin main

if [ $? -ne 0 ]; then
    echo "❌ Git pull fejlede - afbryder deploy"
    exit 1
fi

echo "=== Building new binary ==="
cd /opt/whoknows/implementations/go/backend
go build -o whoknows_new .

if [ $? -ne 0 ]; then
    echo "❌ Build fejlede - afbryder deploy, gamle version kører stadig"
    exit 1
fi

echo "=== Replacing binary ==="
mv whoknows_new whoknows

echo "=== Restarting service ==="
sudo systemctl restart whoknows

sleep 2

echo "=== Tjekker status ==="
if sudo systemctl is-active --quiet whoknows; then
    echo "✅ Deploy lykkedes! App kører"
    sudo systemctl status whoknows
else
    echo "❌ App startede ikke korrekt!"
    sudo journalctl -u whoknows -n 20
    exit 1
fi

echo "=== Nginx status ==="
sudo systemctl status nginx
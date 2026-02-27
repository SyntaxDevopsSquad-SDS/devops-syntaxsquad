#!/bin/bash

echo "=== Installing dependencies ==="
sudo apt update
sudo apt install -y git nginx wget

echo "=== Installing Go 1.26.0 ==="
sudo apt remove -y golang golang-go
sudo apt autoremove -y
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz
rm go1.26.0.linux-amd64.tar.gz

export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

echo "=== Go version ==="
go version

echo "=== Setting up SSH key for GitHub ==="
ssh-keygen -t ed25519 -C "githubVM" -f ~/.ssh/github_key -N ""
echo ""
echo "=== Kopier denne SSH nøgle til GitHub (Settings -> SSH Keys) ==="
cat ~/.ssh/github_key.pub
echo ""
echo "Tryk ENTER når du har tilføjet nøglen til GitHub..."
read

echo "=== Cloning repository ==="
if [ -d "/opt/whoknows" ]; then
    echo "Eksisterende mappe fundet - sletter..."
    sudo rm -rf /opt/whoknows
fi
GIT_SSH_COMMAND="ssh -i ~/.ssh/github_key" git clone git@github.com:SyntaxDevopsSquad-SDS/devops-syntaxsquad.git /opt/whoknows

if [ $? -ne 0 ]; then
    echo "❌ Git clone fejlede - afbryder"
    exit 1
fi

echo "=== Adding swap space for build ==="
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

echo "=== Installing sqlite3 ==="
sudo apt install -y sqlite3

echo "=== Creating database from schema.sql ==="
cd /opt/whoknows/implementations/go/backend
sqlite3 whoknows.db < ../schema.sql

if [ $? -ne 0 ]; then
    echo "❌ Database oprettelse fejlede"
    exit 1
fi

echo "✅ Database oprettet!"

echo "=== Setting up permissions ==="
sudo chown -R www-data:www-data /opt/whoknows/implementations/go/backend
sudo chmod 755 /opt/whoknows/implementations/go/backend
sudo chmod 664 /opt/whoknows/implementations/go/backend/whoknows.db

# Tilføj Go til PATH for alle brugere
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh
source /etc/profile.d/go.sh

echo "=== Building Go app ==="
cd /opt/whoknows/implementations/go/backend
go build -o whoknows .

if [ $? -ne 0 ]; then
    echo "❌ Build fejlede - Python app kører stadig!"
    sudo swapoff /swapfile
    sudo rm /swapfile
    exit 1
fi

echo "=== Removing swap ==="
sudo swapoff /swapfile
sudo rm /swapfile

echo "=== Enter a strong SECRET_KEY for session encryption ==="
read -rsp "SECRET_KEY: " SECRET_KEY
echo ""

if [ -z "$SECRET_KEY" ]; then
    echo "❌ SECRET_KEY cannot be empty"
    exit 1
fi

echo "=== Setting up systemd service ==="
sudo tee /etc/systemd/system/whoknows.service > /dev/null <<EOF
[Unit]
Description=WhoKnows Go App
After=network.target

[Service]
WorkingDirectory=/opt/whoknows/implementations/go/backend
ExecStart=/opt/whoknows/implementations/go/backend/whoknows
Restart=always
RestartSec=3
User=www-data
Environment="SECRET_KEY=${SECRET_KEY}"
Environment="DB_PATH=/opt/whoknows/implementations/go/backend/whoknows.db"

[Install]
WantedBy=multi-user.target
EOF

echo "=== Setting up Nginx ==="
sudo tee /etc/nginx/sites-available/whoknows > /dev/null <<EOF
server {
    listen 80;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;

        # ← Disse er vigtige for sessions/cookies!
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cookie_path / "/; SameSite=Lax";
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/whoknows /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl restart nginx

echo "=== Stopper Python app ==="
pkill -f run_forever.sh
pkill -f app.py
sleep 1

echo "=== Starting Go service ==="
sudo systemctl daemon-reload
sudo systemctl enable whoknows
sudo systemctl start whoknows

sleep 2

if sudo systemctl is-active --quiet whoknows; then
    echo "✅ Go app kører! Python er stoppet"
else
    echo "❌ Go app fejlede - genstarter Python som backup!"
    cd /home/adminuser/devops-syntaxsquad/implementations/python
    nohup bash run_forever.sh &
    exit 1
fi

echo "=== Setup færdig! ==="
sudo systemctl status whoknows
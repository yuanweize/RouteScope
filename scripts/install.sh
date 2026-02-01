#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}==> Installing RouteLens in /opt/routelens...${NC}"

# 1. Build
echo -e "${GREEN}==> Building binary...${NC}"
go build -o routelens ./cmd/server

# 2. User & Group
echo -e "${GREEN}==> Creating user 'routelens'...${NC}"
if ! id "routelens" &>/dev/null; then
    sudo useradd -r -s /bin/false routelens
fi

# 3. Permissions
echo -e "${GREEN}==> Setting permissions...${NC}"
sudo chmod +x routelens

# 4. Set Capabilities (Allow Ping/MTR without root)
echo -e "${GREEN}==> Setting network capabilities...${NC}"
sudo setcap cap_net_raw,cap_net_bind_service+ep routelens

# 5. Environment Config
if [ ! -f env ]; then
    echo -e "${GREEN}==> Creating default env...${NC}"
    cp deploy/env.example env
    # Set default DB path to current dir
    sed -i 's|RS_DB_PATH=.*|RS_DB_PATH=/opt/routelens/routelens.db|' env
fi

# 6. Ownership
echo -e "${GREEN}==> Setting ownership to routelens...${NC}"
sudo chown routelens:routelens /opt/routelens -R

# 7. Install Service
echo -e "${GREEN}==> Installing Systemd service...${NC}"
sudo cp deploy/routelens.service /etc/systemd/system/
sudo systemctl daemon-reload

# 8. Enable & Start
echo -e "${GREEN}==> Enabling and starting service...${NC}"
sudo systemctl enable routelens
sudo systemctl restart routelens

echo -e "${GREEN}==> Installation Complete!${NC}"
echo "Dashboard should be available at http://$(hostname -I | awk '{print $1}'):8080"
echo "Check status with: sudo systemctl status routelens"

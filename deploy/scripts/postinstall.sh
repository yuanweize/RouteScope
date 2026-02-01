#!/bin/bash
# Post-installation script for RouteLens

set -e

# Create routelens user if not exists
if ! id -u routelens &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin routelens
    echo "Created system user 'routelens'"
fi

# Create directories with proper permissions
mkdir -p /var/lib/routelens/data
mkdir -p /var/lib/routelens/data/geoip
chown -R routelens:routelens /var/lib/routelens

# Copy example env if not exists
if [ ! -f /etc/routelens/env ]; then
    cp /etc/routelens/env.example /etc/routelens/env
    chmod 600 /etc/routelens/env
    chown routelens:routelens /etc/routelens/env
fi

# Update systemd service to use correct paths
mkdir -p /opt/routelens
ln -sf /var/lib/routelens/data /opt/routelens/data 2>/dev/null || true

# Reload systemd
systemctl daemon-reload

echo ""
echo "=============================================="
echo "  RouteLens installed successfully!"
echo "=============================================="
echo ""
echo "Next steps:"
echo "  1. Edit /etc/routelens/env with your settings"
echo "  2. Start the service: systemctl start routelens"
echo "  3. Enable on boot: systemctl enable routelens"
echo "  4. Access web UI at http://localhost:8080"
echo ""

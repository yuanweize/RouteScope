#!/bin/bash
# Pre-removal script for RouteLens

set -e

# Stop service if running
if systemctl is-active --quiet routelens; then
    systemctl stop routelens
    echo "Stopped routelens service"
fi

# Disable service
if systemctl is-enabled --quiet routelens 2>/dev/null; then
    systemctl disable routelens
    echo "Disabled routelens service"
fi

echo "RouteLens service stopped and disabled"
echo "Note: Data in /var/lib/routelens is preserved"

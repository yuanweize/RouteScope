# RouteScope (RouteLens)

[![Go Report Card](https://goreportcard.com/badge/github.com/yuanweize/RouteScope)](https://goreportcard.com/report/github.com/yuanweize/RouteScope)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[🇨🇳 中文文档](README_CN.md) | [🇺🇸 English](README.md)

> **A modern network link observation platform for monitoring latency, packet loss, and bandwidth.**

**RouteScope (RouteLens)** empowers users to visualize the "black box" of internet routing by monitoring latency, packet loss, and bandwidth quality between local nodes and remote servers in real-time. By leveraging MTR-style automated tracing and SSH side-channel speed testing, it helps you pinpoint exactly where network degradation occurs.

## 🌟 Key Features

*   **🔍 Field-Tested MTR Tracing**: Visualize packet paths hop-by-hop using Native Go ICMP sockets.
*   **🌍 GeoIP Integration**: Automatically resolve and map IP addresses to physical locations (City/Country/ISP).
*   **🛡️ Stealth Speed Test**: Agentless bandwidth monitoring using SSH side-channels (`/dev/zero` -> SSH -> `/dev/null`), requiring **NO installation** on the target server.
*   **💾 High-Performance Storage**: Built-in SQLite engine with WAL mode and JSON-based series storage for efficient long-term metrics.
*   **📊 Web Dashboard (Coming Soon)**: Interactive World Map and React-based ECharts visualization.

## 🛠️ Architecture

```mermaid
graph TD
    User[User / Administrator] -->|Web UI| FE[React Frontend]
    FE -->|API| BE[Go API Server]
    
    subgraph Core "Probe Engine"
        ICMP[ICMP Pinger]
        MTR[Traceroute Engine]
        SSH[SSH Speed Tester]
    end
    
    BE --> ICMP
    BE --> MTR
    BE --> SSH
    
    ICMP -->|Raw Socket| Network
    MTR -->|Raw Socket| Network
    SSH -->|Encrypted Tunnel| RemoteServer[Remote Target VPS]
    
    BE -->|GORM| DB[(SQLite DB)]
    DB -->|JSON| FE
```

## 📂 Project Structure

```text
.
├── cmd/
│   └── probe_test/      # CLI verification tool for probing logic
├── pkg/
│   ├── prober/          # Core network engine (ICMP, Trace, SSH)
│   ├── storage/         # SQLite persistence layer (GORM)
│   └── geoip/           # MaxMind GeoLite2 wrapper
├── internal/            # Private application logic
└── .github/             # CI/CD workflows
```

## 🚀 Quick Start

### Method 1: Pre-built Binary

Download the latest release for your OS from the [Releases Page](https://github.com/yuanweize/RouteScope/releases).

```bash
# Verify connection
sudo ./routescope-linux-amd64 -mode ping -target 1.1.1.1
```

### Method 2: Build from Source

```bash
# Clone
git clone https://github.com/yuanweize/RouteScope.git
cd RouteScope

# Build
go build -o routescope ./cmd/probe_test

# Run (Traceroute requires root)
sudo ./routescope -mode trace -target 8.8.8.8
```

## License

MIT

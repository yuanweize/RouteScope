<div align="center">

```
██████╗  ██████╗ ██╗   ██╗████████╗███████╗██╗     ███████╗███╗   ██╗███████╗
██╔══██╗██╔═══██╗██║   ██║╚══██╔══╝██╔════╝██║     ██╔════╝████╗  ██║██╔════╝
██████╔╝██║   ██║██║   ██║   ██║   █████╗  ██║     █████╗  ██╔██╗ ██║███████╗
██╔══██╗██║   ██║██║   ██║   ██║   ██╔══╝  ██║     ██╔══╝  ██║╚██╗██║╚════██║
██████╔╝╚██████╔╝╚██████╔╝   ██║   ███████╗███████╗███████╗██║ ╚████║███████║
╚═════╝  ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═══╝╚══════╝
```

**Modern, Agentless Network Observability Platform**

*Trace routes • Measure latency • Visualize paths — all from a single binary*

[![Go Report Card](https://goreportcard.com/badge/github.com/yuanweize/RouteLens)](https://goreportcard.com/report/github.com/yuanweize/RouteLens)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/yuanweize/RouteLens?color=green)](https://github.com/yuanweize/RouteLens/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/yuanweize/RouteLens/release.yml?label=build)](https://github.com/yuanweize/RouteLens/actions)
[![Docker Image](https://img.shields.io/badge/ghcr.io-routelens-blue?logo=docker)](https://github.com/yuanweize/RouteLens/pkgs/container/routelens)

[🇨🇳 中文文档](README_CN.md)

<br>
<img src=".github/assets/webui_screenshot.png" alt="RouteLens Web UI" width="800">

</div>

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 🛰️ **Agentless Monitoring** | Ping, MTR traceroute, and SSH-based bandwidth testing — no agents required on targets |
| 🔄 **Auto-Update** | One-click in-app upgrade mechanism (AdGuard Home style) |
| 🔐 **Secure by Default** | JWT authentication, login rate limiting (5/min), input sanitization |
| 🎨 **Modern UI** | React 19 + Ant Design v5 with automatic dark mode |
| 🌍 **Precision GeoIP** | CLDR country names + ip2region for China with 3000+ city coordinates |
| 📊 **Historical Metrics** | Time-series charts for latency, packet loss, and bandwidth trends |
| 📦 **Single Binary** | Zero dependencies, one-file deployment with systemd support |
| 🗺️ **Smart Map** | Auto-zoom to fit route, auto-refresh with countdown indicator |
| 🎯 **Target Control** | Enable/disable monitoring targets without deletion |
| 🌐 **i18n Ready** | Full Chinese/English support with proper localized place names |

---

## 🚀 Quick Start

### Option 1: Docker (Recommended)

```bash
docker run -d \
  --name routelens \
  --cap-add NET_RAW \
  --cap-add NET_ADMIN \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e RS_JWT_SECRET=your_secure_secret_here \
  ghcr.io/yuanweize/routelens:latest
```

### Option 2: Docker Compose

**Prerequisites:**
- Docker & Docker Compose installed

```bash
# 1. Download configuration
curl -O https://raw.githubusercontent.com/yuanweize/RouteLens/master/compose.yml

# 2. (Optional) Edit compose.yml to set environment variables like JWT secret

# 3. Start service
docker compose up -d

# 4. Visit http://localhost:8080
```

### Option 3: Binary

**Prerequisites:**
- **Linux/macOS**: Requires `mtr` (for route tracing) and `ping`.
  - Ubuntu/Debian: `sudo apt install mtr`
  - CentOS/RHEL: `sudo yum install mtr`
  - macOS: `brew install mtr` (requires running with sudo)
- **Windows**:
  - Requires `mtr` binary in system PATH (e.g. from WinMTR or similar, though Windows support is experimental).
  - Must run terminal as **Administrator**.
  - *Recommendation: Use WSL or Docker on Windows.*

Download from [Releases](https://github.com/yuanweize/RouteLens/releases/latest):

```bash
# Linux
VERSION=$(curl -s https://api.github.com/repos/yuanweize/RouteLens/releases/latest | grep tag_name | cut -d'"' -f4 | tr -d 'v')
curl -LO "https://github.com/yuanweize/RouteLens/releases/latest/download/routelens_${VERSION}_linux_amd64.tar.gz"
tar xzf routelens_${VERSION}_linux_amd64.tar.gz
chmod +x routelens

# Run directly
./routelens --port 8080

# Or install as systemd service
./routelens service install --port 8080
```

---

## 🔧 Initial Setup

1. Open `http://your-server:8080`
2. You'll be redirected to `/setup` on first run
3. Create your admin account
4. Add monitoring targets in the dashboard
5. GeoIP database downloads automatically on first probe

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `RS_JWT_SECRET` | **⚠️ Required for production** - JWT signing key | Random (changes on restart) |
| `RS_HTTP_PORT` | HTTP listen address | `:8080` |
| `RS_DB_PATH` | SQLite database path | `./data/routelens.db` |
| `RS_GEOIP_PATH` | GeoIP database directory | `./data/geoip` |
| `RS_PROBE_INTERVAL` | Probe interval in seconds | `30` |
| `RS_LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

> ⚠️ **Security Note:** In production, always set `RS_JWT_SECRET` to a strong, random value. If not set, a random secret is generated at startup and all sessions will be invalidated on restart.

### Example `.env` file

```env
RS_JWT_SECRET=your-super-secure-random-string-at-least-32-chars
RS_HTTP_PORT=:8080
RS_PROBE_INTERVAL=60
```

---

## 🔄 In-App Updates

RouteLens supports seamless self-updates directly from the web UI:

1. Go to **Settings** → **About & Updates**
2. Click **Check for Updates**
3. If available, click **Install Update**
4. Service restarts automatically with the new version

**Requirements:**
- Process must have write permission to its own binary
- For systemd: service will exit and systemd restarts it

---

## 🛠 Architecture

```mermaid
flowchart LR
  subgraph Backend
    A[Scheduler] --> B[MTR Prober]
    A --> C[ICMP Prober]
    A --> D[SSH Speed Test]
    B & C & D --> E[SQLite]
  end
  
  subgraph Frontend
    F[React 19] --> G[Ant Design v5]
    G --> H[ECharts]
  end
  
  I[Gin API] --> Frontend
  E --> I
```

---

## 📂 Project Structure

```
.
├── cmd/server/       # Application entrypoint
├── internal/
│   ├── api/          # REST API handlers & middleware
│   ├── auth/         # JWT authentication
│   └── monitor/      # Probe scheduler
├── pkg/
│   ├── prober/       # MTR, ICMP, SSH speed test
│   ├── storage/      # SQLite repository
│   └── geoip/        # GeoIP enrichment
└── web/              # React frontend (Vite + TypeScript)
```

---

## 🔨 Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/yuanweize/RouteLens.git
cd RouteLens

# Build frontend
cd web && npm ci && npm run build && cd ..

# Build binary (uses Makefile)
make build          # Build for current platform
make build-linux    # Build for Linux amd64
make build-all      # Build for all platforms
```

### Version Management

Version is managed from a single source of truth: `.github/.release-please-manifest.json`

- **CI builds**: GoReleaser automatically injects version via ldflags
- **Local builds**: `make build` extracts version from manifest
- **Docker builds**: Dockerfile copies manifest for embedding

```bash
# Check current version
make version
```

---

## 🔐 Security

RouteLens includes comprehensive security hardening:

- **JWT Authentication** with cryptographically random secrets
- **Login Rate Limiting** (5 attempts per IP per minute)
- **Input Sanitization** on all probe targets (prevents command injection)
- **Password Validation** (6-72 character limit, bcrypt hashing)
- **Username Validation** (3-32 alphanumeric characters)
- **Generic Error Messages** (internal errors hidden from users)
- **Thread-Safe Operations** (RWMutex protection on shared data)

---

## 📝 License

[MIT License](LICENSE) — Free for personal and commercial use.

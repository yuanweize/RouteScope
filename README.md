# 🛰️ RouteLens

**A Modern, Stealthy, All-in-One Network Observability Platform.**

**现代化、隐身、全栈式的网络链路观测平台。**

[English](#english) | [简体中文](README_CN.md)

![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18+-61DAFB?logo=react&logoColor=000)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)
![Build Status](https://img.shields.io/github/actions/workflow/status/yuanweize/RouteLens/release.yml?branch=master)
![Docker Pulls](https://img.shields.io/docker/pulls/yuanweize/routelens)

---

<a name="english"></a>

## 🚀 Hero

🛰️ **RouteLens**

**Slogan:** *A Modern, Stealthy, All-in-One Network Observability Platform.*

**口号：** 现代化、隐身、全栈式的网络链路观测平台。

> Screenshot placeholder / 截图占位：
> - [ ] Dashboard Overview
> - [ ] World Map Trace View

---

## Introduction / 简介

**EN:** RouteLens helps you pinpoint where your network is slow—ISP bottlenecks, international backbones, or destination datacenters—by tracing the entire path and measuring latency, loss, and throughput.

**中文：** RouteLens 通过全链路追踪与测速，帮助你精确定位网络问题是在本地运营商、国际出口还是目标机房。

**EN:** It is **All-in-One** (Go + React in a single binary) and features **Stealth Mode** for non-invasive bandwidth testing.

**中文：** 项目实现 **All-in-One**（Go + React 单文件交付），并具备 **Stealth Mode**（无 Agent 旁路测速）。

---

## ✨ Key Features / 功能亮点

| Feature | Description |
| --- | --- |
| 🌍 Visual Traceroute | Real-time world map paths (ECharts + GeoIP) with loss hotspot detection. / 实时地图连线，精准定位丢包节点。 |
| 🚀 Multi-Mode Probing | ICMP, HTTP (Download), SSH (Tunnel), Iperf3. / 四种探测模式全覆盖。 |
| 🛡️ Stealth & Safe | Passive probing or SSH tunnel tests to avoid throttling. / 不触发风控的隐蔽测速。 |
| 📦 Zero Dependency | Single binary + built-in Systemd install. / 单二进制交付，内置系统服务安装。 |
| 🔐 Secure Access | Setup wizard + JWT protection. / 初始化向导 + JWT 鉴权。 |

---

## 🛠️ Architecture / 架构逻辑图

```mermaid
flowchart LR
  A[Probe Engine (Go)] --> B[Channel]
  B --> C[(SQLite)]
  C --> D[API Server (Gin)]
  D --> E[Frontend (React)]
```

**中文说明：** 探测引擎产生的链路数据通过通道写入 SQLite，API 层提供查询与触发接口，前端实时渲染图表与地图。

---

## 🚀 Quick Start / 快速开始

### Installation (Linux/macOS)

```bash
# Download
wget https://github.com/yuanweize/RouteLens/releases/latest/download/routelens_linux
chmod +x routelens_linux

# Install as Service
./routelens_linux service install --port 8080
```

**First Run:** open `http://localhost:8080` → `/setup` → set admin password.

**首次运行：** 打开浏览器访问 `http://localhost:8080` → `/setup` → 设置管理员密码。

---

## 📂 Project Structure / 项目结构

```
.
├── cmd/            # Entrypoints (server, tools)
├── internal/       # Core services (API, monitor, auth)
├── pkg/            # Shared libs (prober, storage, geoip)
└── web/            # React frontend (Vite + Arco + ECharts)
```

**中文说明：** cmd 为入口，internal 为核心服务，pkg 为通用库，web 为前端资源。

---

## ⚙️ Configuration / 配置手册

| Env | Description | Default |
| --- | --- | --- |
| RS_PORT | HTTP port (alias). / HTTP 服务端口（别名） | 8080 |
| RS_HTTP_PORT | HTTP port. / HTTP 服务端口 | :8080 |
| RS_DB_PATH | SQLite database path. / SQLite 数据库路径 | ./data/routelens.db |
| RS_JWT_SECRET | JWT signing secret. / JWT 签名密钥 | auto-generated |
| RS_GEOIP_PATH | GeoIP directory (optional). / GeoIP 数据库目录（可选） | empty |
| RS_GEOIP_CITY_DB | GeoIP City DB path. / 城市库路径 | empty |
| RS_GEOIP_ISP_DB | GeoIP ISP DB path. / ISP 库路径 | empty |

---

## 📜 License / 许可证

MIT

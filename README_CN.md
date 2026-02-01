# 🛰️ RouteLens

[🇺🇸 English](README.md)

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/github/actions/workflow/status/yuanweize/RouteLens/release.yml?branch=master)](https://github.com/yuanweize/RouteLens/actions)

---

## 简介

一款现代化、无 Agent 的网络链路观测平台，覆盖路由追踪、延迟/丢包/带宽测量，并提供可视化地图与历史趋势分析。

## ✨ 功能亮点

- 🌍 **实时地图**：ECharts + GeoIP 可视化链路。
- 🚀 **多模式探测**：ICMP、HTTP 下载、SSH 隧道、Iperf3。
- 🕵️ **隐蔽测速**：低噪声、无侵入的带宽测量。
- 📊 **历史趋势**：时序数据展示延迟与丢包曲线。
- 📦 **单文件交付**：一键安装系统服务。
- 🎨 **现代化 UI**：React + Ant Design v5，暗黑模式算法自动适配。

## 📸 截图

- ![Dashboard](docs/images/dashboard.png)
- ![Trace Map](docs/images/map.png)

## 🛠 架构图

```mermaid
flowchart LR
  A[探测引擎 (Go)] --> B[Go Channel]
  B --> C[(SQLite)]
  C --> D[API 服务 (Gin)]
  D --> E[前端 (React + AntD)]
```

## 🚀 快速开始

### 二进制安装

```bash
wget https://github.com/yuanweize/RouteLens/releases/latest/download/routelens_linux
chmod +x routelens_linux
./routelens_linux service install --port 8080
```

访问 `http://localhost:8080` → `/setup` 完成初始化。

### Docker Compose

```yaml
version: '3.8'
services:
  routelens:
    image: yuanweize/routelens:latest
    container_name: routelens
    cap_add:
      - NET_RAW
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    restart: unless-stopped
```

## ⚙️ 配置说明

| 环境变量 | 说明 | 默认值 |
| --- | --- | --- |
| RS_PORT | HTTP 端口（别名） | 8080 |
| RS_HTTP_PORT | 监听地址 | :8080 |
| RS_DB_PATH | SQLite 路径 | ./data/routelens.db |
| RS_JWT_SECRET | JWT 密钥 | 自动生成 |
| RS_GEOIP_PATH | GeoIP 目录 | 空 |
| RS_GEOIP_CITY_DB | GeoIP 城市库 | 空 |
| RS_GEOIP_ISP_DB | GeoIP ISP 库 | 空 |
| RS_PROBE_INTERVAL | 探测间隔（秒） | 30 |

## 📂 项目结构

```
.
├── cmd/          # 入口
├── internal/     # API / 监控 / 鉴权
├── pkg/          # 探测 / 存储 / GeoIP
└── web/          # 前端 (Vite + AntD)
```

## License

MIT

# 🛰️ RouteLens

[English](#english) | [简体中文](#简体中文)

<a name="english"></a>

## English

**RouteLens** is a professional-grade network observability platform that acts like an "X-ray" for your internet connection. It visualizes the entire path from your local device to remote targets, helping you pinpoint network bottlenecks—whether they exist in your local ISP, international backbones, or the destination datacenter.

[![Go Report Card](https://goreportcard.com/badge/github.com/yuanweize/RouteLens)](https://goreportcard.com/report/github.com/yuanweize/RouteLens)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)

### 🚀 Core Features

*   🛰️ **Interactive MTR Visualization**: Real-time traceroute paths rendered on a 3D world map using ECharts. Focus on specific paths to filter out noise.
*   ⚡ **Multi-Mode Probing**:
    *   **ICMP/MTR**: Traditional latency and packet loss tracking.
    *   **SSH Stealth**: Bandwidth testing via SSH side-channels to bypass ISP throttling.
    *   **HTTP Download**: Secure, agent-less bandwidth verification.
    *   **Iperf3 Client**: High-performance benchmarking for server-to-server quality.
*   📉 **Long-term Analytics**: Persistent historical recording of latency, jitter, and bandwidth trends.
*   🛡️ **Modern Security**: Integrated database-backed authentication with a smooth web-based setup wizard.
*   📦 **Single-Binary Delivery**: Built-in system service installation (`./routelens service install`).

### 🛠️ Installation

#### 1. Quick Start (Binary)
Download the latest [Release](https://github.com/yuanweize/RouteLens/releases), then run:
```bash
chmod +x routelens
sudo ./routelens service install --port 8080
```
Visit `http://localhost:8080` to complete the **Setup Wizard**.

#### 2. Docker Compose
```yaml
services:
  routelens:
    image: yuanweize/routelens:latest
    container_name: routelens
    cap_add:
      - NET_RAW
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
    restart: unless-stopped
```

---

<a name="简体中文"></a>

## 简体中文

**RouteLens** 是一款专业级的网络观测平台，被誉为互联网连接的“X光机”。它能够实时可视化从本地到远程目标的完整链路，帮助您精准定位网络瓶颈——无论是本地运营商、国际骨干网（如 CN2/9929）还是目标机房的问题，都一目了然。

### 🚀 核心特性

*   🛰️ **交互式 MTR 可视化**: 基于 ECharts 的 3D 世界地图渲染，实时展示多跳路径。支持路径过滤，拒绝视觉干扰。
*   ⚡ **全能探测引擎**:
    *   **ICMP/MTR**: 经典的延迟与丢包率追踪。
    *   **SSH 隐蔽测速**: 通过 SSH 侧信道进行带宽测试，有效规避运营商流量整形。
    *   **HTTP 下载**: 安全、免客户端的带宽验证方案。
    *   **Iperf3 客户端**: 专业级点对点性能基准测试。
*   📉 **长期趋势分析**: 结构化存储历史数据，直观展示延迟、抖动及带宽的长期趋势图表。
*   🛡️ **现代化安全加固**: 内置数据库鉴权，配合丝滑的 Web 前端初始化向导。
*   📦 **单文件交付**: 原生内置系统服务安装逻辑 (`./routelens service install`)。

### 🛠️ 安装指南

#### 1. 快速开始 (二进制)
下载最新的 [Release](https://github.com/yuanweize/RouteLens/releases)，执行：
```bash
chmod +x routelens
sudo ./routelens service install --port 8080
```
访问 `http://localhost:8080` 即可进入**初始化向导**。

#### 2. Docker Compose 部署
```yaml
services:
  routelens:
    image: yuanweize/routelens:latest
    container_name: routelens
    cap_add:
      - NET_RAW
    ports:
      - "8080:8080"
    volumes:
      - ./data:/root/data
    restart: unless-stopped
```

## ⚙️ Configuration / 配置

| 环境变量 / Env | 描述 / Description | 默认值 / Default |
| :--- | :--- | :--- |
| `RS_PORT` | HTTP 服务端口 | `8080` |
| `RS_DB_PATH` | SQLite 数据库路径 | `./routelens.db` |
| `RS_JWT_SECRET` | JWT 签名密钥 | *(随机生成)* |

## License
MIT

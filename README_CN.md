<div align="center">

```
██████╗  ██████╗ ██╗   ██╗████████╗███████╗██╗     ███████╗███╗   ██╗███████╗
██╔══██╗██╔═══██╗██║   ██║╚══██╔══╝██╔════╝██║     ██╔════╝████╗  ██║██╔════╝
██████╔╝██║   ██║██║   ██║   ██║   █████╗  ██║     █████╗  ██╔██╗ ██║███████╗
██╔══██╗██║   ██║██║   ██║   ██║   ██╔══╝  ██║     ██╔══╝  ██║╚██╗██║╚════██║
██████╔╝╚██████╔╝╚██████╔╝   ██║   ███████╗███████╗███████╗██║ ╚████║███████║
╚═════╝  ╚═════╝  ╚═════╝    ╚═╝   ╚══════╝╚══════╝╚══════╝╚═╝  ╚═══╝╚══════╝
```

**现代化、无 Agent 的网络链路观测平台**

*路由追踪 • 延迟测量 • 路径可视化 — 单文件交付*

[![Go Report Card](https://goreportcard.com/badge/github.com/yuanweize/RouteLens)](https://goreportcard.com/report/github.com/yuanweize/RouteLens)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/yuanweize/RouteLens?color=green)](https://github.com/yuanweize/RouteLens/releases/latest)
[![Build Status](https://img.shields.io/github/actions/workflow/status/yuanweize/RouteLens/release.yml?label=build)](https://github.com/yuanweize/RouteLens/actions)
[![Docker Image](https://img.shields.io/badge/ghcr.io-routelens-blue?logo=docker)](https://github.com/yuanweize/RouteLens/pkgs/container/routelens)

[🇺🇸 English](README.md)

<br>
<img src=".github/assets/webui_screenshot.png" alt="RouteLens Web UI" width="800">

</div>

---

## ✨ 功能亮点

| 功能 | 描述 |
|------|------|
| 🛰️ **无 Agent 监控** | Ping、MTR 路由追踪、SSH 带宽测速 — 目标端无需安装任何软件 |
| 🔄 **应用内更新** | 一键升级机制（类似 AdGuard Home） |
| 🔐 **默认安全** | JWT 认证、登录速率限制（5次/分钟）、输入过滤 |
| 🎨 **现代化 UI** | React 19 + Ant Design v5 + 自动暗色模式 |
| 🌍 **高精度 GeoIP** | CLDR 国家名翻译 + ip2region 中国 IP 精确到城市，内置 3000+ 城市坐标 |
| 📊 **历史指标** | 延迟、丢包、带宽时序曲线 |
| 📦 **单文件交付** | 零依赖，支持 systemd 服务 |
| 🗺️ **智能地图** | 自动缩放适配路由，自动刷新倒计时指示 |
| 🎯 **目标控制** | 禁用/启用监控目标，无需删除 |
| 🌐 **国际化** | 完整中英文支持，地名自动本地化显示 |

---

## 🚀 快速开始

### 方式一：Docker（推荐）

```bash
docker run -d \
  --name routelens \
  --cap-add NET_RAW \
  --cap-add NET_ADMIN \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  -e RS_JWT_SECRET=你的安全密钥 \
  ghcr.io/yuanweize/routelens:latest
```

### 方式二：Docker Compose

**Prerequisites (前置要求):**
- Docker & Docker Compose installed

```bash
# 1. 下载配置文件
curl -O https://raw.githubusercontent.com/yuanweize/RouteLens/master/compose.yml

# 2. (可选) 修改 compose.yml 中的环境变量，如设置 JWT 密钥

# 3. 启动服务
docker compose up -d

# 4. 访问 http://ip:8080
```

### 方式三：二进制部署

**Prerequisites (前置要求):**
- **Linux/macOS**: 需要安装 `mtr` (路由追踪) 和 `ping` (通常自带)。
  - Ubuntu/Debian: `sudo apt install mtr`
  - CentOS/RHEL: `sudo yum install mtr`
  - macOS: `brew install mtr` (需要 sudo 运行)
- **Windows**:
  - 需要下载 [WinMTR](https://sourceforge.net/projects/winmtr/) 或类似工具并将 `mtr.exe` 放入系统 PATH 中（目前 Windows 支持尚不完善，建议使用 WSL 或 Docker）。
  - 需要以 **管理员身份** 运行终端。

从 [Releases](https://github.com/yuanweize/RouteLens/releases/latest) 下载：

```bash
# Linux
VERSION=$(curl -s https://api.github.com/repos/yuanweize/RouteLens/releases/latest | grep tag_name | cut -d'"' -f4 | tr -d 'v')
curl -LO "https://github.com/yuanweize/RouteLens/releases/latest/download/routelens_${VERSION}_linux_amd64.tar.gz"
tar xzf routelens_${VERSION}_linux_amd64.tar.gz
chmod +x routelens

# 直接运行
./routelens --port 8080

# 或安装为 systemd 服务
./routelens service install --port 8080
```

---

## 🔧 初始配置

1. 打开 `http://your-server:8080`
2. 首次运行会跳转到 `/setup` 页面
3. 创建管理员账户
4. 在仪表盘添加监控目标
5. 首次探测时自动下载 GeoIP 数据库

---

## ⚙️ 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `RS_JWT_SECRET` | **⚠️ 生产环境必须设置** - JWT 签名密钥 | 随机生成（重启失效） |
| `RS_HTTP_PORT` | HTTP 监听地址 | `:8080` |
| `RS_DB_PATH` | SQLite 数据库路径 | `./data/routelens.db` |
| `RS_GEOIP_PATH` | GeoIP 数据库目录 | `./data/geoip` |
| `RS_PROBE_INTERVAL` | 探测间隔（秒） | `30` |
| `RS_LOG_LEVEL` | 日志级别（debug/info/warn/error） | `info` |

> ⚠️ **安全提示：** 生产环境务必设置 `RS_JWT_SECRET` 为强随机字符串。未设置时，启动时生成随机密钥，重启后所有会话失效。

---

## 🔄 应用内更新

RouteLens 支持从 Web UI 一键升级：

1. 进入 **设置** → **关于与更新**
2. 点击 **检查更新**
3. 如有新版本，点击 **安装更新**
4. 服务自动重启为新版本

---

## 🔐 安全特性

RouteLens 包含全面安全加固：

- **JWT 认证**：密码学安全的随机密钥
- **登录速率限制**：每 IP 每分钟 5 次
- **输入过滤**：所有探测目标经过验证（防止命令注入）
- **密码验证**：6-72 字符限制，bcrypt 哈希
- **用户名验证**：3-32 字母数字字符
- **通用错误消息**：内部错误对用户隐藏
- **线程安全操作**：RWMutex 保护共享数据

---

## 📂 项目结构

```
.
├── cmd/server/       # 应用入口
├── internal/
│   ├── api/          # REST API 与中间件
│   ├── auth/         # JWT 认证
│   └── monitor/      # 探测调度器
├── pkg/
│   ├── prober/       # MTR、ICMP、SSH 测速
│   ├── storage/      # SQLite 存储
│   └── geoip/        # GeoIP 地理信息
└── web/              # React 前端（Vite + TypeScript）
```

---

## 🔨 开发构建

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/yuanweize/RouteLens.git
cd RouteLens

# 构建前端
cd web && npm ci && npm run build && cd ..

# 构建二进制（使用 Makefile）
make build          # 构建当前平台
make build-linux    # 构建 Linux amd64
make build-all      # 构建所有平台
```

### 版本管理

版本统一由单一来源管理：`.github/.release-please-manifest.json`

- **CI 构建**：GoReleaser 自动通过 ldflags 注入版本
- **本地构建**：`make build` 从 manifest 提取版本
- **Docker 构建**：Dockerfile 复制 manifest 用于嵌入

```bash
# 查看当前版本
make version
```

---

## 📝 许可证

[MIT License](LICENSE) — 可自由用于个人和商业用途。

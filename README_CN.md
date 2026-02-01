# RouteScope (RouteLens) - 网络链路透视镜

[![Go Report Card](https://goreportcard.com/badge/github.com/yuanweize/RouteScope)](https://goreportcard.com/report/github.com/yuanweize/RouteScope)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[🇨🇳 中文文档](README_CN.md) | [🇺🇸 English](README.md)

> **现代化网络链路观测平台，支持延迟、丢包与带宽监控。**
> **A modern network link observation platform for monitoring latency, packet loss, and bandwidth.**

**RouteScope** 通过 PV (Path Visualization) 技术与 SSH 旁路测速机制，帮助用户实时监控从本地到目标服务器的延迟、丢包率与带宽质量。

通过 RouteScope，你可以像医生看 X 光片一样，精准定位网络拥堵是发生在本地 ISP、国际骨干网（如 CN2/9929）还是目标机房，从而彻底告别“网络玄学”。

## 🌟 核心功能

*   **🔍 实时路由追踪 (MTR)**: 基于 Go 原生 Raw Socket 实现的逐跳分析，自动高亮显示丢包节点。
*   **🌍 GeoIP 地理可视化**: 自动解析每一跳 IP 的国家、城市与运营商 (ISP) 信息。
*   **🛡️ 隐蔽旁路测速**: 利用 SSH 协议传输 `/dev/zero` 数据流进行带宽测试，**无需在服务端安装任何 Agent**，安全且不易被流量审查识别。
*   **💾 高性能时序存储**: 内置 SQLite + WAL 模式，单文件存储百万级监控记录，支持自动老化清理。
*   **📊 现代化仪表盘 (开发中)**: 基于 React 的世界地图连线与动态流量波形图。

## 🛠️ 技术架构

```mermaid
graph TD
    User[用户 / 管理员] -->|Web 界面| FE[React 前端]
    FE -->|API 请求| BE[Go 后端服务]
    
    subgraph Core "探测引擎 (Probe Engine)"
        ICMP[ICMP 在线监测]
        MTR[MTR 路由追踪]
        SSH[SSH 带宽测速]
    end
    
    BE --> ICMP
    BE --> MTR
    BE --> SSH
    
    ICMP -->|Raw Socket| Network
    MTR -->|Raw Socket| Network
    SSH -->|加密隧道| RemoteServer[目标 VPS]
    
    BE -->|GORM| DB[(SQLite 数据库)]
    DB -->|JSON 数据| FE
```

## 📂 目录结构

```text
.
├── cmd/
│   └── probe_test/      # 探测逻辑验证 CLI 工具
├── pkg/
│   ├── prober/          # 核心网络探测引擎 (ICMP, Trace, SSH)
│   ├── storage/         # 数据持久化层 (GORM + SQLite)
│   └── geoip/           # GeoIP 解析模块
├── internal/            # 内部业务逻辑
└── .github/             # CI/CD 自动化构建配置
```

## 🚀 快速开始

### 方式 1: 下载二进制 (推荐)

请访问 [Releases 页面](https://github.com/yuanweize/RouteScope/releases) 下载适用于 Linux/macOS/Windows 的最新版本。

```bash
# 不需要安装依赖，直接运行 (需 Root 权限以支持 ICMP)
sudo ./routescope-linux-amd64 -mode ping -target 1.1.1.1
```

### 方式 2: 源码编译

```bash
# 克隆项目
git clone https://github.com/yuanweize/RouteScope.git
cd RouteScope

# 编译 CLI 工具
go build -o routescope ./cmd/probe_test

# 运行路由追踪
sudo ./routescope -mode trace -target 223.5.5.5
```

## 开源协议

MIT

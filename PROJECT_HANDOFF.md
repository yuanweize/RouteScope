# 📋 Project Handoff Report: RouteLens

**Date:** 2026-02-01  
**Status:** Alpha / Core Logic Complete  
**Objective:** Handing over to the next developer/AI for further optimization and production hardening.

---

## 1. Project Status Summary

### Backend (Golang)
- **Framework:** Gin (REST API)
- **Database:** SQLite (GORM) with `modernc.org/sqlite` (CGO-free for easy cross-compilation).
- **Service Management:** Built-in `service install` subcommand using `systemd`.
- **Probing Engine:**
  - **MTR:** Native Go ICMP implementation (requires `NET_RAW` capability).
  - **SSH Speedtest:** Uses `golang.org/x/crypto/ssh` for side-channel speed tests.
  - **HTTP Speedtest:** Direct file download-based bandwidth calculation.
  - **Iperf3:** Wrapper around `iperf3` JSON output.

### Frontend (React)
- **Framework:** Vite + TypeScript.
- **UI:** Arco Design (By ByteDance).
- **Charts:** Apache ECharts with `echarts-for-react`.
- **Assets:** Embedded into the Go binary using `go:embed`.

---

## 2. Technical Architecture

### Data Flow
1. `MonitorService` (run in background) periodically triggers probes stored in the `Target` table.
2. Results are written to the `MonitorRecords` table as time-series data.
3. The API server serves historical data via `/api/v1/history` and real-time triggers via `/api/v1/probe`.

### Prober Logic (`pkg/prober/`)
- Each probe mode (SSH, HTTP, Iperf) is isolated.
- Bandwidth data is NO LONGER mocked. If a prober fails or is not supported (e.g., ICMP-only), speed metrics return 0.

---

## 3. Known Issues & Tech Debt (Must Read!)

> [!WARNING]
> **UI Zebra Stripes (Dark Mode)**
> While most global tokens are fixed, some Arco components (like `Table`) might still render white backgrounds in dark mode if the parent container doesn't properly inherit the `arco-theme` attribute.
> *Location:* `web/src/index.css` and `web/src/App.tsx`.

> [!CAUTION]
> **ECharts Disposal & Reactivity**
> When switching targets in the Dashboard, the ECharts instances in `MapChart.tsx` and `MetricsChart.tsx` sometimes "overlap" or fail to clear old data if the `notMerge` flag isn't respected by the React hook.
> *Tip:* Check the `useEffect` dependencies in `Dashboard/index.tsx`.

> [!IMPORTANT]
> **SSH Key Permissions**
> Automated deployment scripts (`scp`) will fail if the local `.pem` key has permissions wider than `0600`. The binary installation also assumes `root` privileges for systemd.

---

## 4. Recommended Next Steps

1. **Frontend Polish:**
   - Refactor `web/src/components/MapChart.tsx` to handle dynamic GeoIP coordinates from the backend (currently using simulated paths for UI demo).
   - Resolve the Z-index issue for the "Quick Probe" floating loading indicator.
2. **Backend Hardening:**
   - Implement `api/v1/setup` rate limiting to prevent brute-forcing the initialization wizard.
   - Add multi-user support (currently only a single `admin` user is supported in the repository logic).
3. **Deployment Improvements:**
   - Add a `--config` flag to allow specifying a YAML config file instead of relying solely on env vars and CLI flags.

4. **🧹 Housekeeping & Cleanup:**
   Once the next developer/AI has successfully fixed the UI/UX issues and verified the deployment, please **audit and remove** all unnecessary junk files. This includes:
   * Old shell scripts (e.g., `scripts/install.sh`, `deploy/*.sh`).
   * Temporary build artifacts (e.g., `build/`, `dist/` outside of embed).
   * Unused config examples or temporary keys.
   * Keep the repository clean and lightweight.

---

## 5. Contact & Context
This project was developed through a series of intensive phases (1-13.5) focusing on modernizing a legacy MTR tool into a full-stack observability suite.

**End of Handoff.**

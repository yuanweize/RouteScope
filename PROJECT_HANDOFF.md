# Project Handoff

## Status
**Cleaned & Fixed** — Phase Rescue 2.0 complete, documentation polished.

## Scope Summary
- Backend: MTR last-hop latency/loss analysis, GeoIP auto-download (P3TERX mirror), and trace truncation signal.
- Frontend: Ant Design v5 UI, hop table default expanded, N/A formatting, truncated badge.
- Docs: Mermaid diagrams fixed with quoted labels, new badges, and updated installation guidance.

## Quick Start
```bash
./routelens service install --port 8080
```
GeoIP auto-downloads on first run to ./data/geoip.

## Deployment Notes
- Linux service path: /etc/systemd/system/routelens.service
- Working directory: /opt/routelens
- Binary: /opt/routelens/routelens

## Validation Checklist
- [x] GeoIP auto-download log appears on first start
- [x] MTR last-hop latency shown as Avg Latency
- [x] Hop details table expanded by default
- [x] Mermaid diagrams render without errors

## CI/CD Architecture
- Configuration Location: Release Please config lives in .github/release-please-config.json and .github/.release-please-manifest.json (not in repo root).
- Workflow: Commit -> Auto PR -> Merge -> Auto Release -> Auto Build -> Auto Docker.
- Do Not Touch: Do not disable .github/workflows/release.yml; it publishes release binaries.

## Open Items
- None

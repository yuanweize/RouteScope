# Changelog

## [1.2.9](https://github.com/yuanweize/RouteLens/compare/v1.2.0...v1.2.9) (2026-02-02)


### Features

* **self-update:** AdGuard Home 风格的应用内一键更新
* **ci:** Standardize ldflags version injection across all build workflows
* **i18n:** Complete Settings page translations (Chinese/English)
* **update-check:** Multi-strategy version detection with fallback (GitHub API → Raw Manifest → selfupdate)


### Bug Fixes

* **critical:** Fix ldflags injection path (main.version instead of api.Version)
* **frontend:** Remove duplicate Settings component, unify to Ant Design
* **target:** Add enable/disable functionality for monitoring targets
* **ssh-scheduler:** Fix dead code in SSH speed test scheduling


### Chores

* Clean up orphaned releases and tags
* Consolidate version management with release-please


## [1.3.0](https://github.com/yuanweize/RouteLens/compare/v1.2.0...v1.3.0) (2026-02-01) [DEPRECATED]

> This release was not properly published. Changes merged into v1.2.9.

### Features

* **self-update:** AdGuard Home 风格的应用内一键更新 ([c3d2f4c](https://github.com/yuanweize/RouteLens/commit/c3d2f4c79ae70232a3530f2d109c1f381b49070c))


### Bug Fixes

* **critical:** 修复 SSH 测速调度逻辑 - 死代码复活 ([f6df1ec](https://github.com/yuanweize/RouteLens/commit/f6df1eca6d7abfa1fcc02d571aa3ceae21adf683))
* make system/info public API, ensure proper deployment ([7db9dbf](https://github.com/yuanweize/RouteLens/commit/7db9dbf1cd30fddc2fdd8ba57dd2239e02f7ef4f))
* **security:** add cache headers for frontend assets ([de211d8](https://github.com/yuanweize/RouteLens/commit/de211d8c13a14efe14e3caeaad36721ee8f8c17c))
* **security:** update vulnerable dependencies ([96a5c76](https://github.com/yuanweize/RouteLens/commit/96a5c76e7cf21eaf50e6b7832e9471db08b28afd))

## [1.2.0](https://github.com/yuanweize/RouteLens/compare/v1.1.0...v1.2.0) (2026-02-01)


### Features

* **about:** Add Downloads section with release links for all platforms ([1cfa64b](https://github.com/yuanweize/RouteLens/commit/1cfa64b1ce53715638cc440d22ac061ca4d0719c))
* **i18n:** Phase Polish - i18n and observability overhaul ([4823514](https://github.com/yuanweize/RouteLens/commit/4823514b3e0e74dca31411c9c3efdece4f9783b1))
* **observability:** 拒绝'哑巴'系统 - 全面日志增强 ([0ef0012](https://github.com/yuanweize/RouteLens/commit/0ef0012767eac9e3062ca254ee123a580448fb6c))
* **release:** introduce GoReleaser for industrial-grade releases ([cbe5048](https://github.com/yuanweize/RouteLens/commit/cbe5048bc894e670c365e03236b8c4fadaf3b1ed))
* **ui:** Professional About page redesign with hero, tech stack, and quick guide ([ec7ba4e](https://github.com/yuanweize/RouteLens/commit/ec7ba4ec97fcb0bdf9d002609b689a5043717f3d))
* **ux:** Phase User-Experience Fix - live logs and dynamic map i18n ([a4e6f77](https://github.com/yuanweize/RouteLens/commit/a4e6f77d55fe4be9de430ad269205645d3578b26))


### Bug Fixes

* GeoIP precision, input validation, map lines, UI styling ([de8ca85](https://github.com/yuanweize/RouteLens/commit/de8ca856522d5ac2f3cb285f6cd2689bded1c54e))
* release-please token and docker multi-arch build ([6d32d62](https://github.com/yuanweize/RouteLens/commit/6d32d62a486ba9f290e27ebd4638ac6cd4c760bb))


### Performance Improvements

* disable CGO in Docker and split Vite bundles ([4232602](https://github.com/yuanweize/RouteLens/commit/4232602292c6b6ec62eee0adedd7f8895d63ccaa))

## 1.1.0
- Auto GeoIP injection (P3TERX mirror)
- True latency mode (last-hop MTR analysis)
- Ant Design v5 UI refresh

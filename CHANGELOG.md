# Changelog

## [2.2.2](https://github.com/yuanweize/RouteLens/compare/v2.2.1...v2.2.2) (2026-02-02)


### Bug Fixes

* add version.json copy step for CI builds ([b2c6c34](https://github.com/yuanweize/RouteLens/commit/b2c6c348cdc83e8aac0c730bd34f5d27fe83ddab))
* unignore .release-please-manifest.json in .dockerignore ([a05ac2d](https://github.com/yuanweize/RouteLens/commit/a05ac2d072ba81eb3e0224e6e6e1d997d27a869f))

## [2.1.0](https://github.com/yuanweize/RouteLens/compare/v2.0.0...v2.1.0) (2025-01)


### Features

* **geoip:** Integrated ip2region for high-precision China IP lookup (city-level + ISP)
* **geoip:** Embedded 3149 China city coordinates from government data source
* **map:** Auto-zoom to fit all route points on the map
* **map:** Dynamic zoom calculation based on geographic bounding box
* **i18n:** Localized location names on map (Chinese/English based on UI language)
* **i18n:** Localized tooltip labels (Hop/跳数, Latency/延迟, Precision/精度)


### Bug Fixes

* **map:** Fixed missing coordinates for China IPs causing broken map visualization
* **map:** Fixed map default zoom too small to see route details
* **i18n:** Fixed mixed Chinese/English location names when switching language


---

## [2.0.0](https://github.com/yuanweize/RouteLens/compare/v1.3.2...v2.0.0) (2026-02-02)


### ⚠ BREAKING CHANGES

* v2.0.0 - Security hardening and code quality improvements
* Password validation now requires 6-72 characters
* Username validation now requires 3-32 alphanumeric characters


### Features

* **about:** dynamic GitHub releases download links ([a7bc0fb](https://github.com/yuanweize/RouteLens/commit/a7bc0fba4a6772786c696a6cfad27fa9241ef593))
* add database management, settings, and time range selector ([18dbc87](https://github.com/yuanweize/RouteLens/commit/18dbc8772caafab3ec2d8e22c4f4e3c04559a70e))
* add GeoIP database management and SSH key cleanup ([80b24b2](https://github.com/yuanweize/RouteLens/commit/80b24b2f9cf4982763cd5ad321f4113b57a5dd2e))
* enhance GeoIP status display with database metadata ([7251991](https://github.com/yuanweize/RouteLens/commit/72519917951914fb64856ae1fa9717e861d9aacd))
* v1.4.0 - database management, time range selector, improved UI ([44cd139](https://github.com/yuanweize/RouteLens/commit/44cd139218b9bcbc99df687bd9546af8b7ed7ca0))
* v2.0.0 - Security hardening and code quality improvements ([60099d5](https://github.com/yuanweize/RouteLens/commit/60099d5f6737ee97f6fbf2a39ea2a55f0a8e1794))


### Bug Fixes

* display IPv4 + IPv6 for GeoIP ip_version 6 ([fc14ebd](https://github.com/yuanweize/RouteLens/commit/fc14ebde770f72aab9472b556a65e25659b8e331))
* **targets:** parse probe_config when editing to populate SSH fields ([e657096](https://github.com/yuanweize/RouteLens/commit/e6570965bd98fe04b29967a439a0846aa2a1f637))
* **ui:** show error messages on login page for 401/429 errors ([5461e72](https://github.com/yuanweize/RouteLens/commit/5461e7207610a31a797d1ff41a8536289fc2714c))


### Security

* **auth:** Proper bcrypt error handling with logging
* **auth:** Password validation (6-72 character limit)
* **auth:** Username validation (3-32 alphanumeric characters)
* **api:** Target validation in handleProbe using regex pattern
* **api:** Target validation in traceroute before execution
* **api:** Generic error messages (internal errors hidden from users with server-side logging)
* **monitor:** Thread-safe operations with RWMutex on targets slice


## [1.3.2](https://github.com/yuanweize/RouteLens/compare/v1.3.1...v1.3.2) (2026-02-02)


### Bug Fixes

* **api:** password update UNIQUE constraint error ([ff2a9d6](https://github.com/yuanweize/RouteLens/commit/ff2a9d6e1cd02b0e02f2fc443c4113ac225451a9))
* **core:** audit fixes for hardcoded values and GORM misuse ([c695ce4](https://github.com/yuanweize/RouteLens/commit/c695ce46c869f57661dccae1fcf7bf1af2bd332a))

## [1.3.1](https://github.com/yuanweize/RouteLens/compare/v1.3.0...v1.3.1) (2026-02-02)


### Bug Fixes

* **security:** critical security hardening for v1.3.1 ([7488591](https://github.com/yuanweize/RouteLens/commit/7488591b590f862594beabe92dd5ef0fea7a4c61))

## [1.3.0](https://github.com/yuanweize/RouteLens/compare/v1.2.9...v1.3.0) (2026-02-01)


### Features

* **about:** Add Downloads section with release links for all platforms ([1cfa64b](https://github.com/yuanweize/RouteLens/commit/1cfa64b1ce53715638cc440d22ac061ca4d0719c))
* add LICENSE file and update license references in README ([c3c7c11](https://github.com/yuanweize/RouteLens/commit/c3c7c112757e97201100e95a3580bc196ac49f3f))
* add Targets page to display target information ([170f207](https://github.com/yuanweize/RouteLens/commit/170f207555c1fe9fe0283409a13320b603e238a9))
* Complete UI/UX Overhaul & Dynamic Target Management ([1ef2f0e](https://github.com/yuanweize/RouteLens/commit/1ef2f0e69a0978f80a6dd896acf52b313ca02671))
* downgrade React and related types to version 18 for compatibility ([170f207](https://github.com/yuanweize/RouteLens/commit/170f207555c1fe9fe0283409a13320b603e238a9))
* enhance MTR functionality and integrate GeoIP support in monitoring service ([9062c21](https://github.com/yuanweize/RouteLens/commit/9062c21d7981b85b72a5bc4de47aaf228dd65134))
* enhance MTR functionality with latency/loss analysis, GeoIP auto-download, and UI improvements ([9b370bd](https://github.com/yuanweize/RouteLens/commit/9b370bd4bee080606bcc07d44a2592372f6bb41f))
* **i18n:** Phase Polish - i18n and observability overhaul ([4823514](https://github.com/yuanweize/RouteLens/commit/4823514b3e0e74dca31411c9c3efdece4f9783b1))
* implement AppContext for global state management including theme and target selection ([170f207](https://github.com/yuanweize/RouteLens/commit/170f207555c1fe9fe0283409a13320b603e238a9))
* **observability:** 拒绝'哑巴'系统 - 全面日志增强 ([0ef0012](https://github.com/yuanweize/RouteLens/commit/0ef0012767eac9e3062ca254ee123a580448fb6c))
* Phase Cleanup - Target enable/disable, i18n, remove sentimental text ([1b5150e](https://github.com/yuanweize/RouteLens/commit/1b5150ec0d4aeb9e987bcc09a2fb4468403817e1))
* **release:** introduce GoReleaser for industrial-grade releases ([cbe5048](https://github.com/yuanweize/RouteLens/commit/cbe5048bc894e670c365e03236b8c4fadaf3b1ed))
* **self-update:** AdGuard Home 风格的应用内一键更新 ([c3d2f4c](https://github.com/yuanweize/RouteLens/commit/c3d2f4c79ae70232a3530f2d109c1f381b49070c))
* **ui:** add self-update UI with Ant Design ([3662326](https://github.com/yuanweize/RouteLens/commit/3662326e1c211290d75e48e58981b879fd3411b3))
* **ui:** Professional About page redesign with hero, tech stack, and quick guide ([ec7ba4e](https://github.com/yuanweize/RouteLens/commit/ec7ba4ec97fcb0bdf9d002609b689a5043717f3d))
* update CI/CD workflows for Go and Docker, enhance Dockerfile, and modify README for GHCR support ([0adc569](https://github.com/yuanweize/RouteLens/commit/0adc569397fc5e8c224c8e91d93e6d877a01f498))
* **ux:** Phase User-Experience Fix - live logs and dynamic map i18n ([a4e6f77](https://github.com/yuanweize/RouteLens/commit/a4e6f77d55fe4be9de430ad269205645d3578b26))


### Bug Fixes

* adjust metrics and map charts to reflect updated data structure ([170f207](https://github.com/yuanweize/RouteLens/commit/170f207555c1fe9fe0283409a13320b603e238a9))
* CI cross-compilation and update deployment paths ([7fadc20](https://github.com/yuanweize/RouteLens/commit/7fadc20c409039f0d0216795f36879b15acee8e0))
* CI/CD cleanup - reset release-please manifest, complete i18n ([03591fc](https://github.com/yuanweize/RouteLens/commit/03591fc773af113e58be134b8d477d227c799649))
* CI/CD cross-compilation for Darwin/Windows with CGO ([96c0a98](https://github.com/yuanweize/RouteLens/commit/96c0a98cf54d1d0153868b42656f3da23c90de56))
* **critical:** 修复 SSH 测速调度逻辑 - 死代码复活 ([f6df1ec](https://github.com/yuanweize/RouteLens/commit/f6df1eca6d7abfa1fcc02d571aa3ceae21adf683))
* GeoIP precision, input validation, map lines, UI styling ([de8ca85](https://github.com/yuanweize/RouteLens/commit/de8ca856522d5ac2f3cb285f6cd2689bded1c54e))
* handle read deadline error in traceroute execution ([1b559a1](https://github.com/yuanweize/RouteLens/commit/1b559a1d5d8e7ce2387727c4f2faf6864d8c7e4c))
* make system/info public API, ensure proper deployment ([7db9dbf](https://github.com/yuanweize/RouteLens/commit/7db9dbf1cd30fddc2fdd8ba57dd2239e02f7ef4f))
* Redirect loop, Mermaid syntax, and add Docker Compose ([2efe76c](https://github.com/yuanweize/RouteLens/commit/2efe76c3fdb7edc2a0a6c290a7fc9ac7f06afd46))
* release-please token and docker multi-arch build ([6d32d62](https://github.com/yuanweize/RouteLens/commit/6d32d62a486ba9f290e27ebd4638ac6cd4c760bb))
* **security:** add cache headers for frontend assets ([de211d8](https://github.com/yuanweize/RouteLens/commit/de211d8c13a14efe14e3caeaad36721ee8f8c17c))
* **security:** update vulnerable dependencies ([96a5c76](https://github.com/yuanweize/RouteLens/commit/96a5c76e7cf21eaf50e6b7832e9471db08b28afd))
* **ui:** dynamic version display and auto update check ([a43e0bd](https://github.com/yuanweize/RouteLens/commit/a43e0bd7234ea5fcdba6efe17b9f73703dda5d2a))
* update API calls to handle new target structure and add latest trace retrieval ([170f207](https://github.com/yuanweize/RouteLens/commit/170f207555c1fe9fe0283409a13320b603e238a9))
* update check API fallback to GitHub API when no binary found ([1d3bb49](https://github.com/yuanweize/RouteLens/commit/1d3bb49278995c8ac8c05379e7e27e908a1ea41d))


### Performance Improvements

* disable CGO in Docker and split Vite bundles ([4232602](https://github.com/yuanweize/RouteLens/commit/4232602292c6b6ec62eee0adedd7f8895d63ccaa))

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

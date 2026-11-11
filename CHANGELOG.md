# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Token caching** — authentication tokens are now cached per target and reused
  until shortly before expiry, drastically reducing login load on BIG-IP devices.
  Concurrent scrapes for the same target trigger only a single login
  (`internal/utils/cache.go`).
- Unit tests for the token cache, including concurrency and expiry behavior
  (`internal/utils/cache_test.go`).
- `/health` endpoint wired up for liveness checks.
- Build-time version information (`version`, `commit`, `buildDate`) injected via
  `-ldflags` and logged on startup.
- Tooling and community files: `Makefile`, `Dockerfile`, `.dockerignore`,
  `docker-compose.yml`, `prometheus.yml`, systemd unit, `CONTRIBUTING.md`,
  `CODE_OF_CONDUCT.md`, `SECURITY.md`, GitHub Actions (CI + release),
  `.goreleaser.yaml`, and issue/PR templates.

### Changed
- `NewBigIPClient` now uses the cached `GetToken` helper instead of logging in
  on every call.
- Clarified the (intentionally empty) `Collector.Describe` implementation: this
  is a multi-target exporter, so metrics are not known ahead of time and the
  collector is "unchecked" by design.
- Simplified `main()` startup (removed the redundant goroutine + `select{}`).

### Fixed
- Corrected the `oneMinAvgClientSideTraffic.totConns` JSON key (previously
  `totConn`), which prevented that metric from being populated.

### Notes
- Disk metrics keep the `_MB` suffix to match the raw units reported by the F5
  iControl REST API.

## [0.1.0] - Initial



### Added
- Initial release.
- Collectors for Virtual Servers, Pools (and members), Compute/CPU/Memory,
  Disk, SSL Certificates, HA/Sync-Failover, and Global Traffic.
- Multi-target `/probe` endpoint with concurrent scraping.
- Token-based authentication against the BIG-IP `tmos` login provider.
- Multi-arch Docker images (amd64/arm64) automatically built and published to
  GitHub Container Registry (`ghcr.io/haameed/f5_bigip_exporter`) on every release.

[Unreleased]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.7.0...HEAD
[0.1.0]: https://github.com/Haameed/f5_bigip_exporter/releases/tag/v0.1.0
[0.2.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.1.0...v0.2.0

## [0.2.0]

### Changed
- **BREAKING**: Renamed the project from `bigip_exporter` to `f5_bigip_exporter`
  to avoid naming collisions with existing BIG-IP exporters in the Prometheus
  ecosystem. The Go module path, binary name, and Docker image path have all
  changed accordingly:
  - Module: `github.com/Haameed/f5_bigip_exporter`
  - Image: `ghcr.io/haameed/f5_bigip_exporter`
- **BREAKING**: Default listen port changed from `9142` to `11000` to use a
  dedicated, non-conflicting port. Update your scrape configs and deployments.
- TMM (Traffic Management Microkernel) metrics: per-TMM CPU usage ratios
  (5s/1m/5m averages), memory (total/used bytes), and NPU count
  (`/mgmt/tm/sys/tmm-info/stats`). TMM CPU is the primary indicator of traffic
  processing load on BIG-IP.

## [0.3.0]

### Changed
- Node metrics (`/mgmt/tm/ltm/node/stats`): per-node sessions, server-side
  traffic/connections/packets, total requests, and monitor/availability/enabled
  state. Nodes are tracked independently of pool membership.
- Client-SSL profile metrics (`/mgmt/tm/ltm/profile/client-ssl/stats`):
  handshake failures/successes, fatal alerts, insecure handshake accepts,
  active/native connections, encrypted/decrypted throughput, session cache
  stats, and per-version TLS/SSL protocol usage (via a `protocol` label).
- Hardware sensor metrics (`/mgmt/tm/sys/hardware`): fan status, power-supply
  status, temperature (with high limit), and humidity. Includes a
  `bigip_hardware_available` gauge that distinguishes hardware appliances from
  Virtual Edition / vCMP guests — the probe gracefully reports unavailable
  instead of failing on VE.
- `bigip_device_info` metric (constant `1` with metadata in labels:
  marketing name, platform, chassis serial, base MAC, BIOS revision) following
  the standard Prometheus info-metric pattern for Grafana joins.

## [0.4.0]

### Added
- Comprehensive Grafana dashboard (`grafana/f5_bigip_exporter.json`) covering
  all metric families, with a templated data source for easy import.
- Unified alerting guide (`alerts/README.md`) with recommended Prometheus and
  Grafana alerts, including a volume-independent SSL handshake failure-ratio
  alert suitable for high-traffic environments.

## [0.5.0]

### Added
- HTTP profile metrics (`/mgmt/tm/ltm/profile/http/stats`): responses by status
  class (2xx–5xx), requests by method, and requests by HTTP version — enabling
  application-level error-rate monitoring.
- TCP profile metrics (`/mgmt/tm/ltm/profile/tcp/stats`): open connections,
  accepts, connection failures, retransmits, bad segments, and resets as
  network-quality indicators.
- Device metrics (`/mgmt/tm/cm/device/stats`): per-device HA failover state and
  load factor.
- Cipher-strength metrics for client-ssl (`bigip_clientssl_cipher_uses_total`
  with `strength`/`cipher` labels) for weak-cipher detection.
- Dashboard panels and alerts for HTTP 5xx error ratio, TCP retransmits, device
  failover changes, and weak-cipher usage.

### Notes
- Client RTT and server response time are not exposed via iControl REST stats;
  HTTP response-code ratios and TCP retransmit rates provide equivalent
  application/network health signals instead.

## [0.6.0]

### Added
- **Interface metrics** (`/mgmt/tm/net/interface`,
  `/mgmt/tm/net/interface/<name>/stats`): `bigip_interface_status` plus
  throughput/packet/drop/error counters
  (`bigip_interface_counters_{bits,pkts}_{in,out}`,
  `bigip_interface_counters_{drops,errors}_all`), with rich labels for bundle,
  media, MTU, STP, and sFlow settings (`pkg/probe/interface.go`).
  `bigip_interface_status` distinguishes six states: `1`=up, `0`=down,
  `2`=disabled, `3`=uninit, `4`=miss, `-1`=unknown — so an administratively
  disabled or not-yet-initialized interface is never confused with a genuine
  link failure.
- New "🔌 Interfaces" row and `Interface` template variable added to the
  Grafana dashboard (`grafana/bigip_exporter_dashboard.json`), plus a matching
  "Interfaces" alert section in `alerts/README.md`.

### Fixed
- Fixed a `structtag` lint warning in `pkg/probe/interface.go` (unexported
  struct fields carrying `json` tags) by exporting the `InterfaceStatsEntry`
  fields.

## [0.7.0]

### Added
- **License metrics** (`/mgmt/tm/sys/license`, `pkg/probe/license.go`):
  registration key, platform ID, licensed version, and appliance serial
  number (`bigip_license_info`), plus — where reported by the device —
  license start/"licensed on"/last service-check dates, the daily renewal
  notification period, days until expiry, and an expired flag. Not every
  BIG-IP version/license type exposes an expiration date via iControl REST
  ([#1](https://github.com/Haameed/f5_bigip_exporter/issues/1)); in that case
  the probe doesn't fail the scrape — it logs a warning and exposes
  `bigip_license_available=0` instead, so the gap is still visible as a metric.
- New "🔑 License" row added to the Grafana dashboard
  (`grafana/bigip_exporter_dashboard.json`): availability/status stats, a
  days-remaining gauge, an expiration-date stat, a license info table
  (registration key/platform/version/serial), and an important-dates table
  (start/licensed-on/last service check).
- Matching "License" alert rules (expired, expiring ≤7/≤30 days, and an
  expiration-data-not-reported visibility alert) added to `alerts/README.md`.

[0.7.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.4.0...v0.5.0


# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0]

### Changed — ⚠️ BREAKING: Prometheus metric-naming compliance pass

This release brings every metric exposed by the exporter in line with the
[Prometheus metric and label naming best practices](https://prometheus.io/docs/practices/naming/).
**Every dashboard, alert rule, and recording rule that references the metrics
below must be updated.** There is no compatibility shim — metric names and,
in several cases, metric *values* have changed.

- **Renamed metrics to `snake_case`** (no more `camelCase`/`PascalCase`
  segments), across `pkg/probe/traffic.go`, `vs.go`, `interface.go`,
  `disk.go`, and `compute.go`. Examples:
  `bigip_HttpRequests` → `bigip_http_requests_total`,
  `bigip_FiveMinAvg_client_side_traffic_bits_in` →
  `bigip_five_min_avg_client_side_traffic_bits_in`,
  `bigip_virtual_server_clientside_bits_in` →
  `bigip_virtual_server_client_side_received_bits_total`,
  `bigip_interface_counters_bits_in` → `bigip_interface_received_bits_total`.
- **Added the missing `_total` suffix to every cumulative counter** (e.g.
  `bigip_node_total_requests` → `bigip_node_requests_total`,
  `bigip_pool_total_requests` → `bigip_pool_requests_total`,
  `bigip_virtual_server_total_requests` → `bigip_virtual_server_requests_total`,
  `bigip_client_side_traffic_total_connections` →
  `bigip_client_side_connections_total`, and ~30 more).
- **Fixed Gauge/Counter type mismatches**: several cumulative values in
  `pkg/probe/vs.go` (syncookie accepts/rejects/syncookies, message-router
  counters, client/ephemeral-side evicted/slow-killed connections, total
  requests) were previously registered as `GaugeValue` even though they are
  monotonically increasing counters. They are now `CounterValue` and carry a
  `_total` suffix.
- **Converted percentages to base-unit ratios (0–1 instead of 0–100)**, per
  the Prometheus convention that `ratio` values must be 0–1:
  - `bigip_cpu_*_percent` (24 metrics in `pkg/probe/compute.go`) →
    `bigip_cpu_*_ratio`, values divided by 100.
  - `bigip_hardware_humidity_percent` → `bigip_hardware_humidity_ratio`,
    value divided by 100.
  - `bigip_tmm_cpu_usage_ratio`, `bigip_tmm_{five_sec,one_min,five_min}_avg_usage_ratio`,
    and the virtual-server equivalents were already named `_ratio` but were
    still returning 0–100 values; they now correctly return 0–1.
  - **Action required:** any alert/dashboard threshold comparing these
    metrics against `85`, `90`, etc. must be rewritten against `0.85`, `0.90`.
- **Converted disk size metrics from megabytes to bytes** (the Prometheus
  base unit): `bigip_disk_total_size_MB` → `bigip_disk_size_bytes`,
  `bigip_disk_vg_free_size_MB` → `bigip_disk_vg_free_bytes`,
  `bigip_disk_vg_inused_size_MB` → `bigip_disk_vg_in_use_bytes`,
  `bigip_disk_vg_reserved_size_MB` → `bigip_disk_vg_reserved_bytes`. Values
  are now 1,048,576× larger (MiB→bytes).
- **Renamed "info"-style metrics to the `_info` suffix**, following the
  standard Prometheus info-metric pattern (constant value `1`, metadata in
  labels): `bigip_virtual_server_cmp_enable_mode` → `..._cmp_enable_mode_info`,
  `bigip_virtual_server_cmp_enabled` → `..._cmp_enabled_info`,
  `bigip_virtual_server_status_availability_state` → `..._availability_state_info`,
  `bigip_virtual_server_status_enabled_state` → `..._enabled_state_info`,
  `bigip_virtual_server_syncookie_status` → `..._syncookie_status_info`.
- **Converted connection-duration metrics from milliseconds to seconds**
  (the Prometheus base unit for time): `bigip_virtual_server_cs_max_conn_duration`
  → `bigip_virtual_server_client_side_connection_duration_max_seconds` (and
  `_mean_seconds` / `_min_seconds`), values divided by 1000.
- **Converted `camelCase` interface labels to `snake_case`**
  (`bundleSpeed` → `bundle_speed`, `flowControl` → `flow_control`,
  `forceGigabitFiber` → `force_gigabit_fiber`, `stpAutoEdgePort` →
  `stp_auto_edge_port`, etc. — 16 labels total on `bigip_interface_status`).
- Updated `alerts/README.md` and `grafana/bigip_exporter_dashboard.json` to
  use every renamed metric/label and rescaled threshold (ratio panels now
  use Grafana's `percentunit` field type with 0–1 thresholds instead of
  `percent` with 0–100 thresholds).

### Migration notes
- Grep your rule files / dashboards for the old names listed above (or the
  full list of renames applied in this release) and update them before
  upgrading the exporter in production, or you will silently lose
  alerting/graphing coverage for the affected metrics.
- If you depend on absolute values in recording rules for the rescaled
  metrics (CPU/humidity ratios, disk sizes, connection durations), update
  the downstream math (divide-by-100 / divide-by-1000 constants are no
  longer needed — they're applied by the exporter now).

## [0.7.1]

### Added
- **Token caching** — authentication tokens are now cached per target and reused
  until shortly before expiry, drastically reducing login load on BIG-IP devices.
  Concurrent scrapes for the same target trigger only a single login
  (`internal/utils/cache.go`).
- Unit tests for the token cache, including concurrency and expiry behavior
  (`internal/utils/cache_test.go`).
- `/health` endpoint wired up for liveness checks.
- Build-time version information (`version`, `commit`, `buildDate`) injected via
  `-ldflags` and logged on startup.
- Tooling and community files: `Makefile`, `Dockerfile`, `.dockerignore`,
  `docker-compose.yml`, `prometheus.yml`, systemd unit, `CONTRIBUTING.md`,
  `CODE_OF_CONDUCT.md`, `SECURITY.md`, GitHub Actions (CI + release),
  `.goreleaser.yaml`, and issue/PR templates.

### Changed
- `NewBigIPClient` now uses the cached `GetToken` helper instead of logging in
  on every call.
- Clarified the (intentionally empty) `Collector.Describe` implementation: this
  is a multi-target exporter, so metrics are not known ahead of time and the
  collector is "unchecked" by design.
- Simplified `main()` startup (removed the redundant goroutine + `select{}`).

### Fixed
- Corrected the `oneMinAvgClientSideTraffic.totConns` JSON key (previously
  `totConn`), which prevented that metric from being populated.

## [0.1.0] - Initial



### Added
- Initial release.
- Collectors for Virtual Servers, Pools (and members), Compute/CPU/Memory,
  Disk, SSL Certificates, HA/Sync-Failover, and Global Traffic.
- Multi-target `/probe` endpoint with concurrent scraping.
- Token-based authentication against the BIG-IP `tmos` login provider.
- Multi-arch Docker images (amd64/arm64) automatically built and published to
  GitHub Container Registry (`ghcr.io/haameed/f5_bigip_exporter`) on every release.

[0.1.0]: https://github.com/Haameed/f5_bigip_exporter/releases/tag/v0.1.0
[0.2.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.1.0...v0.2.0

## [0.2.0]

### Changed
- **BREAKING**: Renamed the project from `bigip_exporter` to `f5_bigip_exporter`
  to avoid naming collisions with existing BIG-IP exporters in the Prometheus
  ecosystem. The Go module path, binary name, and Docker image path have all
  changed accordingly:
  - Module: `github.com/Haameed/f5_bigip_exporter`
  - Image: `ghcr.io/haameed/f5_bigip_exporter`
- **BREAKING**: Default listen port changed from `9142` to `11000` to use a
  dedicated, non-conflicting port. Update your scrape configs and deployments.
- TMM (Traffic Management Microkernel) metrics: per-TMM CPU usage ratios
  (5s/1m/5m averages), memory (total/used bytes), and NPU count
  (`/mgmt/tm/sys/tmm-info/stats`). TMM CPU is the primary indicator of traffic
  processing load on BIG-IP.

## [0.3.0]

### Changed
- Node metrics (`/mgmt/tm/ltm/node/stats`): per-node sessions, server-side
  traffic/connections/packets, total requests, and monitor/availability/enabled
  state. Nodes are tracked independently of pool membership.
- Client-SSL profile metrics (`/mgmt/tm/ltm/profile/client-ssl/stats`):
  handshake failures/successes, fatal alerts, insecure handshake accepts,
  active/native connections, encrypted/decrypted throughput, session cache
  stats, and per-version TLS/SSL protocol usage (via a `protocol` label).
- Hardware sensor metrics (`/mgmt/tm/sys/hardware`): fan status, power-supply
  status, temperature (with high limit), and humidity. Includes a
  `bigip_hardware_available` gauge that distinguishes hardware appliances from
  Virtual Edition / vCMP guests — the probe gracefully reports unavailable
  instead of failing on VE.
- `bigip_device_info` metric (constant `1` with metadata in labels:
  marketing name, platform, chassis serial, base MAC, BIOS revision) following
  the standard Prometheus info-metric pattern for Grafana joins.

## [0.4.0]

### Added
- Comprehensive Grafana dashboard (`grafana/f5_bigip_exporter.json`) covering
  all metric families, with a templated data source for easy import.
- Unified alerting guide (`alerts/README.md`) with recommended Prometheus and
  Grafana alerts, including a volume-independent SSL handshake failure-ratio
  alert suitable for high-traffic environments.

## [0.5.0]

### Added
- HTTP profile metrics (`/mgmt/tm/ltm/profile/http/stats`): responses by status
  class (2xx–5xx), requests by method, and requests by HTTP version — enabling
  application-level error-rate monitoring.
- TCP profile metrics (`/mgmt/tm/ltm/profile/tcp/stats`): open connections,
  accepts, connection failures, retransmits, bad segments, and resets as
  network-quality indicators.
- Device metrics (`/mgmt/tm/cm/device/stats`): per-device HA failover state and
  load factor.
- Cipher-strength metrics for client-ssl (`bigip_clientssl_cipher_uses_total`
  with `strength`/`cipher` labels) for weak-cipher detection.
- Dashboard panels and alerts for HTTP 5xx error ratio, TCP retransmits, device
  failover changes, and weak-cipher usage.

### Notes
- Client RTT and server response time are not exposed via iControl REST stats;
  HTTP response-code ratios and TCP retransmit rates provide equivalent
  application/network health signals instead.

## [0.6.0]

### Added
- **Interface metrics** (`/mgmt/tm/net/interface`,
  `/mgmt/tm/net/interface/<name>/stats`): `bigip_interface_status` plus
  throughput/packet/drop/error counters
  (`bigip_interface_counters_{bits,pkts}_{in,out}`,
  `bigip_interface_counters_{drops,errors}_all`), with rich labels for bundle,
  media, MTU, STP, and sFlow settings (`pkg/probe/interface.go`).
  `bigip_interface_status` distinguishes six states: `1`=up, `0`=down,
  `2`=disabled, `3`=uninit, `4`=miss, `-1`=unknown — so an administratively
  disabled or not-yet-initialized interface is never confused with a genuine
  link failure.
- New "🔌 Interfaces" row and `Interface` template variable added to the
  Grafana dashboard (`grafana/bigip_exporter_dashboard.json`), plus a matching
  "Interfaces" alert section in `alerts/README.md`.

### Fixed
- Fixed a `structtag` lint warning in `pkg/probe/interface.go` (unexported
  struct fields carrying `json` tags) by exporting the `InterfaceStatsEntry`
  fields.

## [0.7.0]

### Added
- **License metrics** (`/mgmt/tm/sys/license`, `pkg/probe/license.go`):
  registration key, platform ID, licensed version, and appliance serial
  number (`bigip_license_info`), plus — where reported by the device —
  license start/"licensed on"/last service-check dates, the daily renewal
  notification period, days until expiry, and an expired flag. Not every
  BIG-IP version/license type exposes an expiration date via iControl REST
  ([#1](https://github.com/Haameed/f5_bigip_exporter/issues/1)); in that case
  the probe doesn't fail the scrape — it logs a warning and exposes
  `bigip_license_available=0` instead, so the gap is still visible as a metric.
- New "🔑 License" row added to the Grafana dashboard
  (`grafana/bigip_exporter_dashboard.json`): availability/status stats, a
  days-remaining gauge, an expiration-date stat, a license info table
  (registration key/platform/version/serial), and an important-dates table
  (start/licensed-on/last service check).
- Matching "License" alert rules (expired, expiring ≤7/≤30 days, and an
  expiration-data-not-reported visibility alert) added to `alerts/README.md`.

[Unreleased]: https://github.com/Haameed/f5_bigip_exporter/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.7.1...v1.0.0
[0.7.1]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.4.0...v0.5.0


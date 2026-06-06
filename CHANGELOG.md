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

[Unreleased]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.1.0...HEAD
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

[0.5.0]: https://github.com/Haameed/f5_bigip_exporter/compare/v0.4.0...v0.5.0
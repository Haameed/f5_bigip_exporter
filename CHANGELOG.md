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

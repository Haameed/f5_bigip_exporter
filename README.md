# BIG-IP Exporter

[![CI](https://github.com/Haameed/f5_bigip_exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/Haameed/f5_bigip_exporter/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Haameed/f5_bigip_exporter?sort=semver)](https://github.com/Haameed/f5_bigip_exporter/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/Haameed/f5_bigip_exporter)](https://goreportcard.com/report/github.com/Haameed/f5_bigip_exporter)
[![Go Reference](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Exporter-E6522C?logo=prometheus&logoColor=white)](https://prometheus.io/)
[![Docker Image](https://img.shields.io/badge/ghcr.io-f5__bigip__exporter-blue?logo=docker&logoColor=white)](https://github.com/Haameed/f5_bigip_exporter/pkgs/container/f5_bigip_exporter)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)

A **Prometheus exporter** for [F5 BIG-IP](https://www.f5.com/products/big-ip-services) devices, built on top of the **iControl REST API**.

It follows the standard [multi-target exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/) (`/probe?target=...`), allowing a single exporter instance to monitor **many BIG-IP devices concurrently** — without running an agent on each device.

---

## Table of Contents

- [Features](#features)
- [Compatibility](#compatibility)
- [How It Works](#how-it-works)
- [Metrics](#metrics)
- [Installation](#installation)
- [Configuration](#configuration)
- [Command-Line Flags](#command-line-flags)
- [Running the Exporter](#running-the-exporter)
- [Prometheus Configuration](#prometheus-configuration)
- [HTTP Endpoints](#http-endpoints)
- [Security Considerations](#security-considerations)
- [Performance & Token Caching](#performance--token-caching)
- [Development](#development)
- [Contributing](#contributing)
- [Alerting](#alerting)
- [Grafana Dashboard](#grafana-dashboard)
- [License](#license)

---

## Features

- 🔌 **Multi-target & agentless** — monitor multiple BIG-IP devices from one exporter (blackbox-style `/probe`).
- ⚡ **Concurrent scraping** — each subsystem (and even pool members / virtual servers) is collected in parallel via goroutines.
- 🔑 **Token-based authentication** — uses the BIG-IP `tmos` login provider; credentials never appear in URLs.
- 📊 **Rich metric coverage**:
  - **Virtual Servers** — connections, traffic, syncookies, availability/enabled state, usage ratios.
  - **Pools & Pool Members** — active/available members, server-side traffic, connection queues, priority groups, per-member up/down state.
  - **System / Compute** — CPU cores, per-core utilization (5s/1m/5m averages + raw ticks), total/used memory.
  - **Disk** — logical disk size, volume-group free/in-use/reserved.
  - **SSL Certificates** — expiration timestamp & days remaining, expired flag, key size, bundle status.
  - **HA / Sync-Failover** — sync status, failover status & color, traffic-group state.
  - **Global Traffic** — client/server-side throughput, packet errors, denies, TMAuth sessions, HTTP requests.
  - **TMM** — core engine CPU/memory load.
  - **Nodes** — per-node traffic, connections, and health state.
  - **Client-SSL** — handshake success/failure, TLS version usage, session cache.
  - **Hardware** — fan/PSU/temperature/humidity sensors and a device-info metric (appliance only; gracefully skipped on VE).
  - **Interfaces** — per-interface link status (up/down/disabled/uninit/missing/unknown), throughput, packets, drops, and errors.
  - **License** — registration key/platform/version info, license start/expiration dates, days until expiry, and expired flag. Gracefully skipped (with a warning logged and `bigip_license_available=0`) on BIG-IP versions/license types that don't report an expiration date via iControl REST.
- 🏷️ Consistent `target` label on every metric for easy multi-device dashboards.
- 🩺 Built-in `probe_success` and `probe_duration_seconds` metrics per scrape.
- 🔑 **Token-based authentication with caching** — uses the BIG-IP `tmos` login
  provider; tokens are cached and reused until shortly before they expire, so
  repeated scrapes don't hammer the device with new logins.
- **HTTP/TCP profiles** — response codes (2xx–5xx), HTTP versions, TCP retransmits & failures.
- **Devices** — per-device HA failover state.

---

## Compatibility

| Component | Supported / Tested |
|---|---|
| **BIG-IP / TMOS** | 13.x and later (anything exposing the iControl REST endpoints used by the collectors). Verified against **17.1.3** / **17.1.3.1**; appliance, Virtual Edition (VE), and vCMP guests all work — hardware-only and license-expiration metrics are automatically omitted where the platform doesn't report them, without failing the scrape. |
| **Go** | 1.25+ (see [`go.mod`](go.mod)) |
| **OS / Arch (exporter binary & image)** | Linux, macOS, Windows — `amd64` and `arm64` (built via [GoReleaser](.goreleaser.yaml) on every tagged release) |
| **Prometheus** | Any version supporting the [multi-target `/probe` exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/) (relabeling-based scrape config) |

> Running against an older or untested TMOS version? Please [open an
> issue](https://github.com/Haameed/f5_bigip_exporter/issues) with your
> version and any errors you see — reports help grow this table.

---

## How It Works

```
                          ┌──────────────────────────┐
 Prometheus  ── /probe ──▶│      BIG-IP Exporter      │
 (target=...)             │                           │
                          │  1. Resolve target+creds  │
                          │  2. Obtain auth token     │
                          │  3. Fan-out probes (HTTP) │──┐ iControl REST
                          │  4. Aggregate metrics     │  │ (HTTPS)
                          └──────────────────────────┘  │
                                       ▲                 ▼
                                       │        ┌──────────────────┐
                                       └────────│   F5 BIG-IP(s)    │
                                                └──────────────────┘
```

On each `/probe` request:

1. The `target` query parameter is parsed and matched against the configured credentials.
2. The exporter looks for a valid cached auth token for the target. If none
   exists (or it has expired), it authenticates against
   `/mgmt/shared/authn/login` to obtain a fresh `X-F5-Auth-Token`. Concurrent
   requests for the same target trigger only a single login.
3. All collectors (Virtual Servers, Pools, Compute, Disk, Certificates, Sync/HA, Traffic) run **concurrently**.
4. Metrics are merged and returned in the Prometheus exposition format.

---

## Metrics

All metrics are prefixed with `bigip_` and carry a `target` label, and follow
the [Prometheus metric and label naming
conventions](https://prometheus.io/docs/practices/naming/) (base units,
`_total` for counters, `_ratio` for 0–1 fractions, `_info` for constant
metadata metrics, `snake_case` names and labels throughout).

> **Upgrading from < v1.0.0?** Several metric names and a handful of metric
> *values* (CPU/humidity percentages, disk sizes, connection durations)
> changed in v1.0.0 to comply with these conventions. See the
> [CHANGELOG](CHANGELOG.md#100) before upgrading — your dashboards and
> alert rules will need updating.

### Probe (per scrape)

| Metric | Type | Description |
|--------|------|-------------|
| `probe_success` | gauge | `1` if the probe succeeded, `0` otherwise |
| `probe_duration_seconds` | gauge | Time taken to complete the probe |

### Virtual Servers (`bigip_virtual_server_*`)

Labels: `target`, `vs_name`, `partition` (state metrics add `availability_state` / `enabled_state` / `syncookie_status`).

| Metric (suffix) | Type | Description |
|-----------------|------|-------------|
| `client_side_received_bits_total` / `client_side_sent_bits_total` | counter | Client-side throughput in bits |
| `client_side_current_connections` | gauge | Current client-side connections |
| `client_side_max_connections` | gauge | Maximum client-side connections |
| `client_side_received_packets_total` / `client_side_sent_packets_total` | counter | Client-side packets |
| `client_side_connections_total` | counter | Total client-side connections |
| `client_side_evicted_connections_total` / `client_side_slow_killed_connections_total` | counter | Evicted / slow-killed connections |
| `ephemeral_*` | gauge/counter | Same metrics for ephemeral (FastL4) connections |
| `client_side_connection_duration_max_seconds` / `..._mean_seconds` / `..._min_seconds` | gauge | Client-side connection durations, in seconds |
| `requests_total` | counter | Total requests |
| `one_min_avg_usage_ratio` / `five_sec_avg_usage_ratio` / `five_min_avg_usage_ratio` | gauge | CPU usage ratios (0-1) |
| `availability_state_info` / `enabled_state_info` | gauge | Availability / enabled state (constant `1`, state in label) |
| `syncookie_accepts_total` / `syncookie_rejects_total` / `syncookie_syncache_current_entries` / ... | counter/gauge | Syncookie accepts, rejects, syncache, etc. |
| `mr_messages_in_total` / `mr_requests_in_total` / `mr_responses_out_total` ... | counter | Message-router counters |

### Pools (`bigip_pool_*`)

Labels: `target`, `pool`, `partition`. Member metric adds `member`, `session`.

| Metric (suffix) | Type | Description |
|-----------------|------|-------------|
| `active_members` / `available_members` / `total_members` | gauge | Member counts |
| `min_active_members` | gauge | Configured minimum active members |
| `current_sessions` | gauge | Current sessions |
| `serverside_bits_in_total` / `serverside_bits_out_total` | counter | Server-side throughput |
| `serverside_current_connections` / `serverside_max_connections` | gauge | Server-side connections |
| `serverside_connections_total` | counter | Total server-side connections |
| `serverside_packets_in_total` / `serverside_packets_out_total` | counter | Server-side packets |
| `requests_total` | counter | Total requests |
| `availability_state` | gauge | `0`=offline, `1`=unknown, `2`=available |
| `enabled_state` | gauge | `0`=disabled, `1`=enabled |
| `connq_*` / `connq_all_*` | gauge/counter | Connection-queue depth, serviced, age stats |
| `current/highest/lowest_priority_group` | gauge | Priority-group info |
| `mr_*` | counter | Message-router counters |
| `member_state` | gauge | Per-member `1`=up, `0`=down |

### System / Compute (`bigip_system_*`, `bigip_cpu_*`)

Labels: `target`, `host_id` (CPU metrics add `cpu_id`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_system_total_cpu_count` | gauge | Total CPU cores |
| `bigip_system_active_cpu_count` | gauge | Active CPU cores |
| `bigip_system_memory_total_bytes` | gauge | Total memory |
| `bigip_system_memory_used_bytes` | gauge | Used memory |
| `bigip_cpu_{five_sec,one_min,five_min}_avg_{user,system,idle,iowait,irq,softirq,niced,stolen}_ratio` | gauge | Per-core CPU utilization averages (0-1 ratio) |
| `bigip_cpu_{user,system,idle,iowait,irq,softirq,niced,stolen}_ticks_total` | counter | Raw per-core CPU ticks |

### Disk (`bigip_disk_*`)

Labels: `target`, `name`, `mode`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_disk_size_bytes` | gauge | Total disk size, in bytes |
| `bigip_disk_vg_free_bytes` | gauge | Volume-group free size, in bytes |
| `bigip_disk_vg_in_use_bytes` | gauge | Volume-group in-use size, in bytes |
| `bigip_disk_vg_reserved_bytes` | gauge | Volume-group reserved size, in bytes |

### SSL Certificates (`bigip_certificate_*`)

Labels: `target`, `name`, `partition` (plus `key_type`, `expiration_date` on some).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_certificate_expiration_timestamp_seconds` | gauge | Unix expiry timestamp |
| `bigip_certificate_expiration_days` | gauge | Days until expiry |
| `bigip_certificate_expired` | gauge | `1`=expired, `0`=valid |
| `bigip_certificate_key_size_bits` | gauge | Key size in bits |
| `bigip_certificate_is_bundle` | gauge | `1`=bundle, `0`=single |
| `bigip_certificate_version` | gauge | Certificate version |
| `bigip_certificate_size_bytes` | gauge | File size in bytes |

### HA / Sync-Failover (`bigip_ha_*`)

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_ha_sync_status` | gauge | `0`=red, `1`=yellow, `2`=green, `3`=blue, `4`=gray |
| `bigip_ha_failover_status` | gauge | `0`=unknown, `1`=offline, `2`=forced_offline, `3`=standby, `4`=active |
| `bigip_ha_failover_color_status` | gauge | Failover health color (same scale as sync) |
| `bigip_ha_traffic_group_status` | gauge | `0`=unknown, `1`=standby, `2`=active |

### Global Traffic (`bigip_*_side_traffic_*`, `bigip_tmauth_*`, …)

Label: `target`. Includes client/server-side throughput, packets, connections (current/max/total/evicted/slow-killed), 5s/1m/5m averages, `bigip_dropped_packets_total`, `bigip_http_requests_total`, packet errors, various deny counters, and TMAuth session/result metrics.

### TMM (`bigip_tmm_*`)

Labels: `target`, `tmm_id`. TMM is the core traffic-processing engine of BIG-IP.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_tmm_cpu_usage_ratio` | gauge | TMM CPU usage ratio |
| `bigip_tmm_five_sec_avg_usage_ratio` | gauge | CPU usage averaged over 5 seconds |
| `bigip_tmm_one_min_avg_usage_ratio` | gauge | CPU usage averaged over 1 minute |
| `bigip_tmm_five_min_avg_usage_ratio` | gauge | CPU usage averaged over 5 minutes |
| `bigip_tmm_memory_total_bytes` | gauge | Total memory allocated to TMM |
| `bigip_tmm_memory_used_bytes` | gauge | Memory used by TMM |
| `bigip_tmm_npus` | gauge | Number of Network Processing Units |

### Nodes (`bigip_node_*`)

Labels: `target`, `node`, `partition`, `address`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_node_current_sessions` | gauge | Current sessions |
| `bigip_node_serverside_bits_in_total` / `_out_total` | counter | Server-side throughput |
| `bigip_node_serverside_current_connections` | gauge | Current connections |
| `bigip_node_serverside_max_connections` | gauge | Max connections |
| `bigip_node_serverside_connections_total` | counter | Total connections |
| `bigip_node_serverside_packets_in_total` / `_out_total` | counter | Server-side packets |
| `bigip_node_requests_total` | counter | Total requests |
| `bigip_node_monitor_status` | gauge | `0`=down, `1`=up, `2`=unchecked |
| `bigip_node_availability_state` | gauge | `0`=offline, `1`=unknown, `2`=available |
| `bigip_node_enabled_state` | gauge | `0`=disabled, `1`=enabled |

### Client-SSL Profiles (`bigip_clientssl_*`)

Labels: `target`, `profile`, `partition`. Protocol-usage metric adds a
`protocol` label.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_clientssl_handshake_failures_total` | counter | Failed SSL handshakes |
| `bigip_clientssl_secure_handshakes_total` | counter | Successful secure handshakes (basis for SSL TPS) |
| `bigip_clientssl_insecure_handshake_accepts_total` | counter | Insecure handshakes accepted |
| `bigip_clientssl_insecure_handshake_rejects_total` | counter | Insecure handshakes rejected |
| `bigip_clientssl_fatal_alerts_total` | counter | Fatal TLS alerts |
| `bigip_clientssl_current_active_handshakes` | gauge | Handshakes currently in progress |
| `bigip_clientssl_current_native_connections` | gauge | Current native SSL connections |
| `bigip_clientssl_max_native_connections` | gauge | Max native SSL connections |
| `bigip_clientssl_encrypted_bytes_in_total` / `_out_total` | counter | Encrypted throughput |
| `bigip_clientssl_decrypted_bytes_in_total` / `_out_total` | counter | Decrypted throughput |
| `bigip_clientssl_session_cache_hits_total` | counter | SSL session cache hits |
| `bigip_clientssl_session_cache_lookups_total` | counter | SSL session cache lookups |
| `bigip_clientssl_protocol_uses_total` | counter | Handshakes per TLS/SSL version (label `protocol`: `SSLv2`, `SSLv3`, `TLSv1.0`, `TLSv1.1`, `TLSv1.2`, `TLSv1.3`) |
| `bigip_clientssl_cipher_uses_total` | counter | Cipher usage by category (labels: `strength` = `weak`/`strong`, `cipher`) |

### Hardware (`bigip_hardware_*`, `bigip_device_info`)

Physical sensor metrics. **Only available on hardware appliances** — on Virtual
Edition or vCMP guests, `bigip_hardware_available` is `0` and sensor metrics are
omitted (the probe does not fail).

Labels: `target` (plus sensor-specific labels).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_hardware_available` | gauge | `1`=hardware appliance, `0`=VE/vCMP guest |
| `bigip_hardware_fan_status` | gauge | Fan status `1`=up, `0`=not up (label: `index`) |
| `bigip_hardware_power_supply_status` | gauge | PSU status `1`=up, `0`=not up (labels: `index`, `current_status`) |
| `bigip_hardware_temperature_celsius` | gauge | Temperature in °C (labels: `index`, `location`) |
| `bigip_hardware_temperature_high_limit_celsius` | gauge | High temperature limit in °C |
| `bigip_hardware_humidity_ratio` | gauge | Chassis relative humidity (0-1 ratio) |
| `bigip_device_info` | gauge | Constant `1`; device metadata in labels: `marketing_name`, `platform`, `chassis_serial`, `base_mac`, `bios_revision` |

#### Using the info metric in Grafana

`bigip_device_info` follows the standard Prometheus "info metric" pattern. Join
it onto other metrics to enrich them with device metadata:

```promql
bigip_tmm_cpu_usage_ratio
  * on(target) group_left(marketing_name, platform) bigip_device_info
```

### HTTP Profiles (`bigip_http_*`)

Labels: `target`, `profile`, `partition`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_http_profile_requests_total` | counter | Total HTTP requests handled |
| `bigip_http_responses_total` | counter | Responses by status class (label `code`: `2xx`/`3xx`/`4xx`/`5xx`) |
| `bigip_http_requests_by_method_total` | counter | Requests by method (label `method`: `GET`/`POST`) |
| `bigip_http_requests_by_version_total` | counter | Requests by protocol version (label `version`: `0.9`–`3.0`) |

### TCP Profiles (`bigip_tcp_*`)

Labels: `target`, `profile`, `partition`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_tcp_open_connections` | gauge | Current open TCP connections |
| `bigip_tcp_accepts_total` | counter | Accepted TCP connections |
| `bigip_tcp_connection_failures_total` | counter | TCP connection failures |
| `bigip_tcp_retransmits_total` | counter | TCP retransmits (network-quality indicator) |
| `bigip_tcp_received_bad_segments_total` | counter | Bad TCP segments received |
| `bigip_tcp_received_resets_total` | counter | TCP RST packets received |

### Devices (`bigip_device_*`)

Labels: `target`, `device`, `hostname`, `management_ip`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_device_failover_state` | gauge | `0`=unknown, `1`=standby, `2`=active |
| `bigip_device_current_load_factor` | gauge | Current HA load factor |

### Interfaces (`bigip_interface_*`)

Labels: `target`, `name` (status metric adds `bundle`, `bundle_speed`,
`flow_control`, `force_gigabit_fiber`, `forward_error_correction`,
`link_traps_enabled`, `lldp_admin`, `mac_address`, `media_active`, `media_fixed`,
`media_max`, `media_sfp`, `mtu`, `port_fwd_mode`, `prefer_port`, `qinq_ethertype`,
`sflow_poll_interval`, `sflow_poll_interval_global`, `stp`, `stp_auto_edge_port`,
`stp_edge_port`, `stp_link_type`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_interface_status` | gauge | `1`=up, `0`=down, `2`=disabled, `3`=uninit, `4`=miss, `-1`=unknown |
| `bigip_interface_received_bits_total` / `bigip_interface_sent_bits_total` | counter | Interface throughput in bits |
| `bigip_interface_received_packets_total` / `bigip_interface_sent_packets_total` | counter | Interface packet counts |
| `bigip_interface_dropped_packets_total` | counter | Packets dropped on the interface |
| `bigip_interface_errors_total` | counter | Errors on the interface |

> ⚠️ Because `disabled`, `uninit`, and `miss` are distinct from `down`, a
> naive `bigip_interface_status == 0` alert only fires on genuine link
> failures — administratively disabled or not-yet-initialized interfaces
> won't trigger false positives.

### License (`bigip_license_*`)

Labels: `target` (`bigip_license_info` adds `registration_key`, `platform_id`,
`licensed_version`, `appliance_serial_number`; `bigip_license_expired` adds
`expiration_date`).

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_license_available` | gauge | `1` if the device reports a license expiration date, `0` otherwise |
| `bigip_license_info` | gauge | Constant `1`; registration/platform/version metadata via labels |
| `bigip_license_start_timestamp_seconds` | gauge | Unix timestamp when the license became valid |
| `bigip_licensed_on_timestamp_seconds` | gauge | Unix timestamp when the device was licensed |
| `bigip_license_service_check_timestamp_seconds` | gauge | Unix timestamp of the last license service check |
| `bigip_license_daily_renew_notification_period_days` | gauge | Days before expiration that daily renewal notifications start |
| `bigip_license_expiration_timestamp_seconds` | gauge | Unix timestamp when the license expires |
| `bigip_license_expiration_days` | gauge | Days until the license expires (`0` if already expired) |
| `bigip_license_expired` | gauge | `1`=expired, `0`=valid |

> ⚠️ Not every BIG-IP version/license type reports a license expiration date
> via iControl REST (`/mgmt/tm/sys/license`) — see
> [issue #1](https://github.com/Haameed/f5_bigip_exporter/issues/1). When it's
> missing, the exporter logs a warning and exposes only
> `bigip_license_available=0`, instead of failing the scrape.

#### Useful queries


```promql
# SSL handshake failure rate
rate(bigip_clientssl_handshake_failures_total[5m])

# SSL TPS (successful handshakes per second)
rate(bigip_clientssl_secure_handshakes_total[5m])

# Alert on legacy/insecure protocol usage
rate(bigip_clientssl_protocol_uses_total{protocol=~"SSLv.*|TLSv1.0|TLSv1.1"}[5m]) > 0
```

---

## Installation

### Build from Source

Requires **Go 1.25+**.

```bash
git clone https://github.com/Haameed/f5_bigip_exporter.git
cd f5_bigip_exporter
go build -o f5_bigip_exporter ./cmd/f5_bigip_exporter
```

### Docker

#### Pull from GitHub Container Registry

Pre-built multi-arch images (amd64 / arm64) are published on every release:

```bash
docker pull ghcr.io/haameed/f5_bigip_exporter:latest
# or a specific version:
docker pull ghcr.io/haameed/f5_bigip_exporter:v0.7.0
# Run it:
docker run -d --name f5_bigip_exporter \
  -p 11000:11000 \
  -v "$(pwd)/bigip-config.yaml:/etc/f5_bigip_exporter/bigip-config.yaml:ro" \
  ghcr.io/haameed/f5_bigip_exporter:latest -insecure
```

A minimal multi-stage `Dockerfile`:

```dockerfile
# ---- build ----
FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/f5_bigip_exporter ./cmd/f5_bigip_exporter

# ---- runtime ----
FROM gcr.io/distroless/static-debian12
COPY --from=build /bin/f5_bigip_exporter /bin/f5_bigip_exporter
EXPOSE 11000
ENTRYPOINT ["/bin/f5_bigip_exporter"]
CMD ["-config", "/etc/f5_bigip_exporter/bigip-config.yaml"]
```

```bash
docker build -t f5_bigip_exporter .
docker run -d --name f5_bigip_exporter \
  -p 11000:11000 \
  -v "$(pwd)/bigip-config.yaml:/etc/f5_bigip_exporter/bigip-config.yaml:ro" \
  f5_bigip_exporter -insecure
```

---

## Configuration

Credentials are provided via a YAML file. Each key is the **target URL** (scheme + host), mapped to its `username` / `password`.

Create `bigip-config.yaml`:

```yaml
https://192.168.100.10:
  username: yourusername
  password: yourpassword

https://192.168.100.11:
  username: yourusername
  password: yourpassword
```

> ℹ️ A sample file is provided as [`config-example.yml`](config-example.yml).
> The `target` you pass to `/probe?target=...` **must exactly match** a key in this file
> (same scheme, host, and port).

> ⚠️ Token authentication requires the **`https`** scheme.

---

## Command-Line Flags

| Flag             | Default              | Description |
|------------------|----------------------|-------------|
| `-config`        | `bigip-config.yaml`  | Path to the credentials YAML file |
| `-listen`        | `:11000`              | Address the HTTP server listens on |
| `-scrape-timeout`| `30`                 | Maximum seconds allowed for a single scrape |
| `-https-timeout` | `10`                 | TLS handshake timeout in seconds |
| `-insecure`      | `false`              | Skip TLS certificate verification (useful for self-signed BIG-IP certs) |

---

## Running the Exporter

```bash
./f5_bigip_exporter -config bigip-config.yaml -insecure
```

Test a probe manually:

```bash
curl 'http://localhost:11000/probe?target=https://192.168.100.10'
```

You should see Prometheus-formatted metrics, ending with `probe_success 1`.

---

## Prometheus Configuration

Use the multi-target relabeling pattern so the `target` becomes a query parameter and the device address is preserved in the `instance` label:

```yaml
scrape_configs:
  - job_name: 'bigip'
    metrics_path: /probe
    static_configs:
      - targets:
          - https://192.168.100.10
          - https://192.168.100.11
    relabel_configs:
      # Pass the target as a query parameter to the exporter
      - source_labels: [__address__]
        target_label: __param_target
      # Preserve the real device address as the instance label
      - source_labels: [__param_target]
        target_label: instance
      # Point the actual scrape at the exporter
      - target_label: __address__
        replacement: localhost:11000
```

---

## HTTP Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /probe?target=<url>` | Scrape metrics for a single BIG-IP target |
| `GET /metrics` | The exporter's own (process/Go) metrics |
| `GET /health` | Liveness/health check — returns `200 OK` |

---

## Security Considerations

- **Credentials at rest** — the config file contains plaintext credentials. Restrict its permissions (`chmod 600`) and consider mounting it as a read-only secret in containerized environments.
- **Least privilege** — use a dedicated BIG-IP user with read-only / auditor-level permissions.
- **TLS verification** — `-insecure` disables certificate validation. Prefer trusting the BIG-IP CA and leaving verification enabled in production.
- **Network exposure** — the `/probe` endpoint accepts arbitrary `target` values that match the config map. Keep the exporter on a trusted management network.

---

## Performance & Token Caching

To minimize load on your BIG-IP devices, the exporter **caches authentication
tokens per target** and reuses them until shortly before they expire (a small
safety margin is applied so scrapes never fail on a token that expires
mid-request).

This means:

- A login (`/mgmt/shared/authn/login`) only happens on the **first** scrape of a
  target, and then again **after the token expires** — not on every scrape.
- When many scrapes arrive for the same target at once, only **one** login is
  performed; the others wait for and share the result.

No configuration is required — caching is automatic.

## Development

### Project Layout

```
.
├── cmd/f5_bigip_exporter      # main entrypoint
├── internal
│   ├── config              # flag parsing + YAML credentials loading
│   └── utils               # F5 token authentication
└── pkg
    ├── http                # iControl REST client (token-based)
    └── probe               # collectors: vs, pools, compute, disk, certs, syncgroup, traffic, interface, license
```


### Run Tests
The token-caching layer is covered by unit tests in
`internal/utils/cache_test.go`, including a concurrency test that verifies only
a single login is issued under load.

```bash
go test ./...
```

### Adding a New Collector

1. Create a new file under `pkg/probe/` exposing a function with the signature:
   ```go
   func GetMyProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool)
   ```
2. Register it in the `allProbes` slice in `pkg/probe/probe.go`.
3. The framework runs it concurrently and aggregates its metrics automatically.

---

## Contributing

**Everyone is welcome to participate and contribute!** 🎉

- 🐛 Found a bug? [Open an issue](https://github.com/Haameed/f5_bigip_exporter/issues).
- 💡 Have a feature idea or a new metric to expose? Open an issue to discuss it.
- 🔧 Submit Pull Requests — please run `go fmt ./...`, `go vet ./...`, and `go test ./...` before opening.

When contributing metrics, follow the
[Prometheus metric naming best practices](https://prometheus.io/docs/practices/naming/)
(base units, `_total` for counters, descriptive `HELP` text).

---

## Alerting

A complete set of recommended alerts (for both Prometheus and Grafana) is
documented in [`alerts/README.md`](alerts/README.md), covering availability,
system load, TMM, HA, virtual servers, pools/nodes, client-SSL, certificates,
hardware sensors, and interfaces.

---

## Grafana Dashboard

A ready-to-import dashboard covering every metric family exposed by the
exporter — Overview, System (CPU/Memory/Disk), TMM, High Availability,
Global Traffic, Virtual Servers, Pools, Nodes, Client-SSL, Certificates,
Hardware, HTTP/TCP Profiles, Devices, Interfaces, and License — is provided
at [`grafana/bigip_exporter_dashboard.json`](grafana/bigip_exporter_dashboard.json).

<!--
TODO: add a dashboard screenshot here once available, e.g.:
![Grafana dashboard overview](docs/images/dashboard-overview.png)
-->

**To import it:**

1. In Grafana, go to **Dashboards → New → Import**.
2. Upload `grafana/bigip_exporter_dashboard.json` (or paste its contents).
3. Select your Prometheus data source when prompted (the `DS_PROMETHEUS`
   variable).
4. Pick a target from the `Target` dropdown at the top of the dashboard to
   filter all panels to that BIG-IP device.

---

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Hamed Maleki
# BIG-IP Exporter

[![Go Reference](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Prometheus](https://img.shields.io/badge/Prometheus-Exporter-E6522C?logo=prometheus&logoColor=white)](https://prometheus.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](#contributing)

A **Prometheus exporter** for [F5 BIG-IP](https://www.f5.com/products/big-ip-services) devices, built on top of the **iControl REST API**.

It follows the standard [multi-target exporter pattern](https://prometheus.io/docs/guides/multi-target-exporter/) (`/probe?target=...`), allowing a single exporter instance to monitor **many BIG-IP devices concurrently** — without running an agent on each device.

---

## Table of Contents

- [Features](#features)
- [How It Works](#how-it-works)
- [Metrics](#metrics)
- [Installation](#installation)
- [Configuration](#configuration)
- [Command-Line Flags](#command-line-flags)
- [Running the Exporter](#running-the-exporter)
- [Prometheus Configuration](#prometheus-configuration)
- [HTTP Endpoints](#http-endpoints)
- [Security Considerations](#security-considerations)
- [Development](#development)
- [Contributing](#contributing)
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
- 🏷️ Consistent `target` label on every metric for easy multi-device dashboards.
- 🩺 Built-in `probe_success` and `probe_duration_seconds` metrics per scrape.
- 🔑 **Token-based authentication with caching** — uses the BIG-IP `tmos` login
  provider; tokens are cached and reused until shortly before they expire, so
  repeated scrapes don't hammer the device with new logins.

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

All metrics are prefixed with `bigip_` and carry a `target` label.

### Probe (per scrape)

| Metric | Type | Description |
|--------|------|-------------|
| `probe_success` | gauge | `1` if the probe succeeded, `0` otherwise |
| `probe_duration_seconds` | gauge | Time taken to complete the probe |

### Virtual Servers (`bigip_virtual_server_*`)

Labels: `target`, `vs_name`, `partition` (state metrics add `availability_state` / `enabled_state` / `syncookie_status`).

| Metric (suffix) | Type | Description |
|-----------------|------|-------------|
| `clientside_bits_in` / `clientside_bits_out` | gauge | Client-side throughput in bits |
| `clientside_current_connections` | gauge | Current client-side connections |
| `clientside_max_connections` | gauge | Maximum client-side connections |
| `clientside_packets_in` / `clientside_packets_out` | gauge | Client-side packets |
| `clientside_total_connections` | gauge | Total client-side connections |
| `clientside_evicted_connections` / `clientside_slow_killed` | gauge | Evicted / slow-killed connections |
| `ephemeral_*` | gauge | Same metrics for ephemeral connections |
| `cs_max_conn_duration` / `cs_mean_conn_duration` / `cs_min_conn_duration` | gauge | Client-side connection durations |
| `total_requests` | gauge | Total requests |
| `one_min_avg_usage_ratio` / `five_sec_avg_usage_ratio` / `five_min_avg_usage_ratio` | gauge | CPU usage ratios |
| `status_availability_state` / `status_enabled_state` | gauge | Availability / enabled state (value `1`, info in label) |
| `syncookie_*` | gauge | Syncookie accepts, rejects, syncache, etc. |
| `mr_msg_in` / `mr_req_in` / `mr_resp_out` ... | gauge | Message-router counters |

### Pools (`bigip_pool_*`)

Labels: `target`, `pool`, `partition`. Member metric adds `member`, `session`.

| Metric (suffix) | Type | Description |
|-----------------|------|-------------|
| `active_members` / `available_members` / `total_members` | gauge | Member counts |
| `min_active_members` | gauge | Configured minimum active members |
| `current_sessions` | gauge | Current sessions |
| `serverside_bits_in_total` / `serverside_bits_out_total` | counter | Server-side throughput |
| `serverside_current_connections` / `serverside_max_connections` | gauge | Server-side connections |
| `serverside_total_connections` | counter | Total server-side connections |
| `serverside_packets_in_total` / `serverside_packets_out_total` | counter | Server-side packets |
| `total_requests` | counter | Total requests |
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
| `bigip_cpu_{five_sec,one_min,five_min}_avg_{user,system,idle,iowait,irq,softirq,niced,stolen}_percent` | gauge | Per-core CPU utilization averages |
| `bigip_cpu_{user,system,idle,iowait,irq,softirq,niced,stolen}_ticks_total` | counter | Raw per-core CPU ticks |

### Disk (`bigip_disk_*`)

Labels: `target`, `name`, `mode`.

| Metric | Type | Description |
|--------|------|-------------|
| `bigip_disk_total_size_MB` | gauge | Total disk size (MB) |
| `bigip_disk_vg_free_size_MB` | gauge | Volume-group free size (MB) |
| `bigip_disk_vg_inused_size_MB` | gauge | Volume-group in-use size (MB) |
| `bigip_disk_vg_reserved_size_MB` | gauge | Volume-group reserved size (MB) |

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

Label: `target`. Includes client/server-side throughput, packets, connections (current/max/total/evicted/slow-killed), 5s/1m/5m averages, `bigip_dropped_packets`, `bigip_HttpRequests`, packet errors, various deny counters, and TMAuth session/result metrics.

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
### Pull from GitHub Container Registry

Pre-built multi-arch images (amd64 / arm64) are published on every release:

```bash
docker pull ghcr.io/haameed/f5_bigip_exporter:latest
# or a specific version:
docker pull ghcr.io/haameed/f5_bigip_exporter:v0.2.0
# Run it:
docker run -d --name f5_bigip_exporter \
  -p 11000:11000 \
  -v "$(pwd)/bigip-config.yaml:/etc/f5_bigip_exporter/bigip-config.yaml:ro" \
  ghcr.io/haameed/f5_bigip_exporter:latest -insecure

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
    └── probe               # collectors: vs, pools, compute, disk, certs, syncgroup, traffic
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

## License

This project is licensed under the **MIT License** — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Hamed Maleki
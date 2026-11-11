# Alerting Guide — F5 BIG-IP Exporter

This document provides a complete set of recommended alerts for the
`f5_bigip_exporter`, suitable for both **Prometheus alert rules** and
**Grafana Alerting** (the queries are identical — only the configuration UI
differs).

> **Thresholds are sensible starting points.** Tune them to your environment,
> especially in high-traffic deployments (see the SSL section for a
> volume-independent approach).

## Table of Contents

- [Severity Levels](#severity-levels)
- [How to Use](#how-to-use)
- [Availability & Probe](#availability--probe)
- [System (CPU / Memory / Disk)](#system-cpu--memory--disk)
- [TMM](#tmm)
- [High Availability](#high-availability)
- [Virtual Servers](#virtual-servers)
- [Pools & Nodes](#pools--nodes)
- [Client-SSL](#client-ssl)
- [Certificates](#certificates)
- [Hardware](#hardware)
- [Interfaces](#interfaces)
- [License](#license)

## Severity Levels

| Severity   | Meaning                                            |
|------------|----------------------------------------------------|
| `critical` | Immediate action required (outage / hardware risk) |
| `warning`  | Needs investigation soon (degradation / capacity)  |

## How to Use

### Prometheus

Copy the `expr` of any alert into a rule file:

```yaml
groups:
  - name: f5_bigip
    rules:
      - alert: F5TmmCpuHigh
        expr: bigip_tmm_one_min_avg_usage_ratio > 0.85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "TMM CPU high on {{ $labels.instance }}"
          description: "TMM {{ $labels.tmm_id }} CPU usage is above 85% for 5 minutes."
```

### Grafana Alerting

Use the same `expr` as the alert query, set the **reduce/threshold** to match
the condition, and set the pending period to the `for` value.

---

## Availability & Probe

### Probe Down — `critical`
The exporter cannot scrape the target.

```promql
probe_success == 0
```
**For:** `2m`

---

## System (CPU / Memory / Disk)

### CPU High — `warning`
```promql
avg by (instance) (1 - bigip_cpu_one_min_avg_idle_ratio) > 0.85
```
**For:** `10m`

### Memory High — `warning`
```promql
(bigip_system_memory_used_bytes / bigip_system_memory_total_bytes) * 100 > 90
```
**For:** `5m`

### Disk High — `warning`
```promql
(bigip_disk_vg_in_use_bytes / bigip_disk_size_bytes) * 100 > 85
```
**For:** `10m`

---

## TMM

> TMM is the core traffic-processing engine. Its CPU is the single most
> important load indicator on BIG-IP.

### TMM CPU High — `warning`
```promql
bigip_tmm_one_min_avg_usage_ratio > 0.85
```
**For:** `5m`
**Description:** `TMM {{ $labels.tmm_id }} CPU usage is above 85% for 5 minutes.`

### TMM Memory High — `warning`
```promql
(bigip_tmm_memory_used_bytes / bigip_tmm_memory_total_bytes) * 100 > 90
```
**For:** `5m`

---

## High Availability

### Sync Status RED — `warning`
```promql
bigip_ha_sync_status == 0
```
**For:** `5m`

### Unexpected Failover (became Standby) — `warning`
> Useful to catch an unplanned active→standby transition.
```promql
bigip_ha_failover_status == 3
```
**For:** `1m`

---

## Virtual Servers

### Virtual Server Offline — `critical`
> `availability_state` label: `offline` indicates the VS is down.
```promql
bigip_virtual_server_availability_state_info{availability_state="offline"} == 1
```
**For:** `2m`

---

## Pools & Nodes

### Pool Has No Active Members — `critical`
```promql
bigip_pool_active_members == 0
```
**For:** `2m`
**Description:** `Pool {{ $labels.pool }} ({{ $labels.partition }}) has 0 active members.`

### Pool Member Down — `warning`
```promql
bigip_pool_member_state == 0
```
**For:** `5m`

### Node Offline — `warning`
> State: `0=offline, 1=unknown, 2=available`. `< 1` matches only offline.
```promql
bigip_node_availability_state < 1
```
**For:** `5m`
**Description:** `Node {{ $labels.node }} ({{ $labels.address }}) is offline.`

### Node Health Monitor Down — `warning`
> Status: `0=down, 1=up, 2=unchecked`. `< 1` matches only down (ignores
> unmonitored nodes).
```promql
bigip_node_monitor_status < 1
```
**For:** `5m`

---

## Client-SSL

### SSL Handshake Failure Ratio High — `warning`
> **Volume-independent.** Instead of an absolute rate (which breaks in
> high-traffic environments), this alerts on the *percentage* of handshakes
> that fail. A small number of failures is always normal; a high ratio is not.

```promql
sum(rate(bigip_clientssl_handshake_failures_total[5m]))
/
(
  sum(rate(bigip_clientssl_secure_handshakes_total[5m]))
  + sum(rate(bigip_clientssl_handshake_failures_total[5m]))
) * 100
> 5
```
**For:** `10m`
**Description:** `More than 5% of SSL handshakes are failing.`

> Tune the `> 5` threshold: `2` (strict), `5` (balanced), `10` (lenient).
> For very high-volume deployments, start at `5` and adjust.

### Legacy / Insecure TLS in Use — `warning`
```promql
sum by (instance, profile, protocol) (
  rate(bigip_clientssl_protocol_uses_total{protocol=~"SSLv2|SSLv3|TLSv1.0|TLSv1.1"}[15m])
) > 0
```
**For:** `15m`
**Description:** `Profile {{ $labels.profile }} is negotiating {{ $labels.protocol }} connections.`

### SSL Fatal Alerts Spike — `warning`
> Volume-dependent; tune to your baseline.
```promql
rate(bigip_clientssl_fatal_alerts_total[5m]) > 10
```
**For:** `5m`

### Weak Cipher in Use — `warning`
> Detects negotiation of insecure ciphers (DES, RC4, NULL, MD5). The data comes
> from the same client-ssl endpoint — no extra load.

```promql
sum by (instance, profile, cipher) (
  rate(bigip_clientssl_cipher_uses_total{strength="weak"}[15m])
) > 0
```
**For:** `15m`
**Description:** `Profile {{ $labels.profile }} is negotiating weak cipher '{{ $labels.cipher }}'.`

---

## Certificates

### Certificate Expired — `critical`
```promql
bigip_certificate_expired == 1
```
**For:** `0m`

### Certificate Expiring Soon (≤ 7 days) — `critical`
```promql
bigip_certificate_expiration_days > 0 and bigip_certificate_expiration_days <= 7
```
**For:** `1h`

### Certificate Expiring (≤ 30 days) — `warning`
```promql
bigip_certificate_expiration_days > 7 and bigip_certificate_expiration_days <= 30
```
**For:** `1h`

---

## Hardware

> These alerts only fire on physical appliances. On Virtual Edition / vCMP
> guests the underlying metrics are absent, so no false positives are produced.

### Fan Failure — `critical`
```promql
bigip_hardware_fan_status < 1
```
**For:** `1m`

### Power Supply Failure — `critical`
```promql
bigip_hardware_power_supply_status < 1
```
**For:** `1m`
**Description:** `PSU {{ $labels.index }} ({{ $labels.current_status }}) is not up.`

### Temperature Near Limit — `critical`
> Relative to each sensor's own configured high limit (model-independent).
```promql
(bigip_hardware_temperature_celsius / bigip_hardware_temperature_high_limit_celsius) * 100 > 90
```
**For:** `5m`
**Description:** `Sensor {{ $labels.location }} is above 90% of its high limit.`

### High Humidity — `warning`
```promql
bigip_hardware_humidity_ratio > 0.80
```
**For:** `10m`

---

## Interfaces

> `bigip_interface_status` distinguishes `down` (genuine link failure) from
> administratively `disabled` and not-yet-`uninit`ialized interfaces, so these
> alerts don't false-positive on interfaces that are intentionally off.

### Interface Down — `critical`
> Status: `1=up, 0=down, 2=disabled, 3=uninit, 4=miss, -1=unknown`. Only
> matches a genuine link failure on an enabled interface.
```promql
bigip_interface_status == 0
```
**For:** `2m`
**Description:** `Interface {{ $labels.name }} on {{ $labels.instance }} is down.`

### Interface Errors — `warning`
```promql
rate(bigip_interface_errors_total[5m]) > 0
```
**For:** `10m`
**Description:** `Interface {{ $labels.name }} on {{ $labels.instance }} is reporting errors.`

### Interface Drops — `warning`
```promql
rate(bigip_interface_dropped_packets_total[5m]) > 0
```
**For:** `10m`
**Description:** `Interface {{ $labels.name }} on {{ $labels.instance }} is dropping packets.`

### Interface Near Saturation — `warning`
> Tune the `> 900e6` threshold (bits/s) to ~90% of the interface's rated
> speed (available via the `bundle_speed` label on `bigip_interface_status`).
```promql
rate(bigip_interface_received_bits_total[5m]) > 900e6
or
rate(bigip_interface_sent_bits_total[5m]) > 900e6
```
**For:** `10m`

---

## License

> These alerts only fire when the device actually reports a license
> expiration date (`bigip_license_available == 1`). Devices/versions that
> don't expose one simply won't match these expressions — no false
> positives.

### License Expired — `critical`
```promql
bigip_license_expired == 1
```
**For:** `0m`

### License Expiring Soon (≤ 7 days) — `critical`
```promql
bigip_license_expiration_days > 0 and bigip_license_expiration_days <= 7
```
**For:** `1h`

### License Expiring (≤ 30 days) — `warning`
```promql
bigip_license_expiration_days > 7 and bigip_license_expiration_days <= 30
```
**For:** `1h`

### License Expiration Data Not Reported — `warning`
> Visibility check: lets you know which devices you can't rely on
> expiration alerting for, so you can track their license renewal manually.
```promql
bigip_license_available == 0
```
**For:** `1h`

---

## Notes


- All alerts include `{{ $labels.instance }}` in their description to identify
  the affected device.
- The `for` durations are tuned to avoid alerting on brief spikes. Lower them
  for faster detection, raise them to reduce noise.
- In high-throughput environments, prefer **ratio-based** alerts (like the SSL
  handshake failure ratio) over absolute-rate thresholds.
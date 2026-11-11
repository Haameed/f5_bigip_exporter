package probe

import (
	"log"
	"strings"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetNodeProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		nodeCurSessions = prometheus.NewDesc(
			"bigip_node_current_sessions",
			"Current sessions for the node",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersideBitsIn = prometheus.NewDesc(
			"bigip_node_serverside_bits_in_total",
			"Total bits received on the server side",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersideBitsOut = prometheus.NewDesc(
			"bigip_node_serverside_bits_out_total",
			"Total bits sent on the server side",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersideCurConns = prometheus.NewDesc(
			"bigip_node_serverside_current_connections",
			"Current server-side connections",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersideMaxConns = prometheus.NewDesc(
			"bigip_node_serverside_max_connections",
			"Maximum server-side connections",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersideTotConns = prometheus.NewDesc(
			"bigip_node_serverside_connections_total",
			"Total server-side connections",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersidePktsIn = prometheus.NewDesc(
			"bigip_node_serverside_packets_in_total",
			"Total packets received on the server side",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeServersidePktsOut = prometheus.NewDesc(
			"bigip_node_serverside_packets_out_total",
			"Total packets sent on the server side",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeTotRequests = prometheus.NewDesc(
			"bigip_node_requests_total",
			"Total requests to the node",
			[]string{"target", "node", "partition", "address"}, nil,
		)
		nodeMonitorStatus = prometheus.NewDesc(
			"bigip_node_monitor_status",
			"Node monitor status (0=down, 1=up, 2=unchecked/unknown)",
			[]string{"target", "node", "partition", "address", "monitor_rule"}, nil,
		)
		nodeAvailabilityState = prometheus.NewDesc(
			"bigip_node_availability_state",
			"Node availability state (0=offline, 1=unknown, 2=available)",
			[]string{"target", "node", "partition", "address", "status_reason"}, nil,
		)
		nodeEnabledState = prometheus.NewDesc(
			"bigip_node_enabled_state",
			"Node enabled state (0=disabled, 1=enabled)",
			[]string{"target", "node", "partition", "address"}, nil,
		)
	)

	type NodeStatsEntry struct {
		Addr                    StringBased `json:"addr"`
		CurSessions             ValueBased  `json:"curSessions"`
		MonitorRule             StringBased `json:"monitorRule"`
		MonitorStatus           StringBased `json:"monitorStatus"`
		TmName                  StringBased `json:"tmName"`
		ServersideBitsIn        ValueBased  `json:"serverside.bitsIn"`
		ServersideBitsOut       ValueBased  `json:"serverside.bitsOut"`
		ServersideCurConns      ValueBased  `json:"serverside.curConns"`
		ServersideMaxConns      ValueBased  `json:"serverside.maxConns"`
		ServersidePktsIn        ValueBased  `json:"serverside.pktsIn"`
		ServersidePktsOut       ValueBased  `json:"serverside.pktsOut"`
		ServersideTotConns      ValueBased  `json:"serverside.totConns"`
		SessionStatus           StringBased `json:"sessionStatus"`
		StatusAvailabilityState StringBased `json:"status.availabilityState"`
		StatusEnabledState      StringBased `json:"status.enabledState"`
		StatusReason            StringBased `json:"status.statusReason"`
		TotRequests             ValueBased  `json:"totRequests"`
	}

	type NodeStatsNested struct {
		Entries NodeStatsEntry `json:"entries"`
	}

	type NodeStats struct {
		NestedStats NodeStatsNested `json:"nestedStats"`
	}

	type NodeStatsResponse struct {
		Kind     string               `json:"kind"`
		SelfLink string               `json:"selfLink"`
		Entries  map[string]NodeStats `json:"entries"`
	}

	var m []prometheus.Metric

	var resp NodeStatsResponse
	if err := c.Get("/mgmt/tm/ltm/node/stats", &resp); err != nil {
		log.Printf("Error getting node stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		stats := entry.NestedStats.Entries

		nodeName, partition := parseNodeName(stats.TmName.Description)
		address := stats.Addr.Description

		m = append(m, prometheus.MustNewConstMetric(
			nodeCurSessions, prometheus.GaugeValue,
			float64(stats.CurSessions.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersideBitsIn, prometheus.CounterValue,
			float64(stats.ServersideBitsIn.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersideBitsOut, prometheus.CounterValue,
			float64(stats.ServersideBitsOut.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersideCurConns, prometheus.GaugeValue,
			float64(stats.ServersideCurConns.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersideMaxConns, prometheus.GaugeValue,
			float64(stats.ServersideMaxConns.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersideTotConns, prometheus.CounterValue,
			float64(stats.ServersideTotConns.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersidePktsIn, prometheus.CounterValue,
			float64(stats.ServersidePktsIn.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeServersidePktsOut, prometheus.CounterValue,
			float64(stats.ServersidePktsOut.Value), target, nodeName, partition, address,
		))
		m = append(m, prometheus.MustNewConstMetric(
			nodeTotRequests, prometheus.CounterValue,
			float64(stats.TotRequests.Value), target, nodeName, partition, address,
		))

		monitorVal := nodeMonitorStatusToValue(stats.MonitorStatus.Description)
		m = append(m, prometheus.MustNewConstMetric(
			nodeMonitorStatus, prometheus.GaugeValue,
			float64(monitorVal), target, nodeName, partition, address, stats.MonitorRule.Description,
		))

		availVal := poolAvailabilityStateToValue(stats.StatusAvailabilityState.Description)
		m = append(m, prometheus.MustNewConstMetric(
			nodeAvailabilityState, prometheus.GaugeValue,
			float64(availVal), target, nodeName, partition, address, stats.StatusReason.Description,
		))

		enabledVal := poolEnabledStateToValue(stats.StatusEnabledState.Description)
		m = append(m, prometheus.MustNewConstMetric(
			nodeEnabledState, prometheus.GaugeValue,
			float64(enabledVal), target, nodeName, partition, address,
		))
	}

	return m, true
}

func parseNodeName(tmName string) (nodeName, partition string) {
	parts := strings.Split(strings.TrimPrefix(tmName, "/"), "/")
	if len(parts) >= 2 {
		return parts[1], parts[0]
	} else if len(parts) == 1 {
		return parts[0], "Common"
	}
	return tmName, "Common"
}

func nodeMonitorStatusToValue(status string) int {
	switch strings.ToLower(status) {
	case "up":
		return 1
	case "down":
		return 0
	default:
		return 2
	}
}

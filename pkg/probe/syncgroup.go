package probe

import (
	"log"
	"strings"

	"github.com/Haameed/bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetSyncGroupProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		syncStatus = prometheus.NewDesc(
			"bigip_ha_sync_status",
			"Device synchronization status (0=red, 1=yellow, 2=green, 3=blue, 4=gray)",
			[]string{"target", "mode"}, nil,
		)

		failoverStatus = prometheus.NewDesc(
			"bigip_ha_failover_status",
			"Device failover status (0=unknown, 1=offline, 2=forced_offline, 3=standby, 4=active)",
			[]string{"target"}, nil,
		)

		failoverColor = prometheus.NewDesc(
			"bigip_ha_failover_color_status",
			"Failover health color (0=red, 1=yellow, 2=green, 3=blue, 4=gray)",
			[]string{"target"}, nil,
		)

		trafficGroupStatus = prometheus.NewDesc(
			"bigip_ha_traffic_group_status",
			"Traffic group failover state (0=unknown, 1=standby, 2=active)",
			[]string{"target", "traffic_group", "device_name", "failover_state", "next_active"}, nil,
		)
	)

	type SyncStatusEntry struct {
		Color   StringBased `json:"color"`
		Mode    StringBased `json:"mode"`
		Status  StringBased `json:"status"`
		Summary StringBased `json:"summary"`
	}

	type SyncStatusNested struct {
		Entries SyncStatusEntry `json:"entries"`
	}

	type SyncStatusStats struct {
		NestedStats SyncStatusNested `json:"nestedStats"`
	}

	type SyncStatusResponse struct {
		Kind     string                     `json:"kind"`
		Entries  map[string]SyncStatusStats `json:"entries"`
		SelfLink string                     `json:"selfLink"`
	}

	type FailoverStatusEntry struct {
		Color             StringBased `json:"color,omitempty"`
		Status            StringBased `json:"status,omitempty"`
		Summary           StringBased `json:"summary,omitempty"`
		LastMessage       StringBased `json:"lastMsg,omitempty"`
		LocalFailoverAddr StringBased `json:"localFailoverAddr,omitempty"`
		PktsReceived      StringBased `json:"pktsReceived,omitempty"`
		RemoteDeviceName  StringBased `json:"remoteDeviceName,omitempty"`
		Transitions       StringBased `json:"transitions,omitempty"`
		Description       StringBased `json:"description,omitempty"`
	}

	type FailoverStatusNested struct {
		Entries FailoverStatusEntry `json:"entries"`
	}

	type FailoverStatusStats struct {
		NestedStats FailoverStatusNested `json:"nestedStats"`
	}

	type FailoverStatusResponse struct {
		Kind     string                         `json:"kind"`
		Entries  map[string]FailoverStatusStats `json:"entries"`
		SelfLink string                         `json:"selfLink"`
	}

	type TrafficGroupEntry struct {
		DeviceName    StringBased `json:"deviceName"`
		FailoverState StringBased `json:"failoverState"`
		TrafficGroup  StringBased `json:"trafficGroup"`
		NextActive    StringBased `json:"nextActive"`
	}

	type TrafficGroupNested struct {
		Entries TrafficGroupEntry `json:"entries"`
	}

	type TrafficGroupStats struct {
		NestedStats TrafficGroupNested `json:"nestedStats"`
	}

	type TrafficGroupResponse struct {
		Kind     string                       `json:"kind"`
		Entries  map[string]TrafficGroupStats `json:"entries"`
		SelfLink string                       `json:"selfLink"`
	}

	var m []prometheus.Metric

	var syncStatusResp SyncStatusResponse
	if err := c.Get("/mgmt/tm/cm/sync-status", &syncStatusResp); err != nil {
		log.Printf("Error getting sync status: %v", err)
	} else {
		for _, entry := range syncStatusResp.Entries {
			stats := entry.NestedStats.Entries
			colorValue := colorToValue(stats.Color.Description)
			m = append(m, prometheus.MustNewConstMetric(
				syncStatus,
				prometheus.GaugeValue,
				float64(colorValue),
				target,
				stats.Mode.Description,
			))
		}
	}

	var failoverStatusResp FailoverStatusResponse
	if err := c.Get("/mgmt/tm/cm/failover-status", &failoverStatusResp); err != nil {
		log.Printf("Error getting failover status: %v", err)
	} else {
		for _, entry := range failoverStatusResp.Entries {
			stats := entry.NestedStats.Entries

			if stats.Status.Description != "" && stats.Summary.Description != "" {
				statusValue := failoverStatusToValue(stats.Status.Description)
				m = append(m, prometheus.MustNewConstMetric(
					failoverStatus,
					prometheus.GaugeValue,
					float64(statusValue),
					target,
				))

			}
			if stats.Color.Description != "" {
				colorValue := colorToValue(stats.Color.Description)

				m = append(m, prometheus.MustNewConstMetric(
					failoverColor,
					prometheus.GaugeValue,
					float64(colorValue),
					target,
				))
			}

		}
	}

	var trafficGroupResp TrafficGroupResponse
	if err := c.Get("/mgmt/tm/cm/traffic-group/stats", &trafficGroupResp); err != nil {
		log.Printf("Error getting traffic group status: %v", err)
	} else {
		for _, entry := range trafficGroupResp.Entries {
			stats := entry.NestedStats.Entries
			stateValue := trafficGroupStateToValue(stats.FailoverState.Description)

			m = append(m, prometheus.MustNewConstMetric(
				trafficGroupStatus,
				prometheus.GaugeValue,
				float64(stateValue),
				target,
				stats.TrafficGroup.Description,
				stats.DeviceName.Description,
				stats.FailoverState.Description,
				stats.NextActive.Description,
			))

		}
	}

	return m, true
}

func colorToValue(color string) int {
	color = strings.ToLower(color)
	switch color {
	case "red":
		return 0
	case "yellow":
		return 1
	case "green":
		return 2
	case "blue":
		return 3
	case "gray":
		return 4
	default:
		return -1
	}
}

func failoverStatusToValue(status string) int {
	status = strings.ToUpper(status)
	switch status {
	case "OFFLINE":
		return 1
	case "FORCED OFFLINE":
		return 2
	case "STANDBY":
		return 3
	case "ACTIVE":
		return 4
	default:
		return 0
	}
}

func trafficGroupStateToValue(state string) int {
	state = strings.ToLower(state)
	switch state {
	case "standby":
		return 1
	case "active":
		return 2
	default:
		return 0
	}
}


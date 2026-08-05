package probe

import (
	"log"
	"strings"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetDeviceProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		deviceFailoverState = prometheus.NewDesc(
			"bigip_device_failover_state",
			"Device failover state (0=unknown, 1=standby, 2=active)",
			[]string{"target", "device", "hostname", "management_ip"}, nil,
		)
		deviceLoadFactor = prometheus.NewDesc(
			"bigip_device_current_load_factor",
			"Current HA load factor of the device",
			[]string{"target", "device", "hostname"}, nil,
		)
	)

	type DeviceStatsEntry struct {
		FailoverState     StringBased `json:"failoverState"`
		CurrentLoadFactor StringBased `json:"currentLoadFactor"`
		Hostname          StringBased `json:"hostname"`
		ManagementIp      StringBased `json:"managementIp"`
		TmName            StringBased `json:"tmName"`
	}
	type DeviceNested struct {
		Entries DeviceStatsEntry `json:"entries"`
	}
	type DeviceStats struct {
		NestedStats DeviceNested `json:"nestedStats"`
	}
	type DeviceResponse struct {
		Entries map[string]DeviceStats `json:"entries"`
	}

	var m []prometheus.Metric
	var resp DeviceResponse
	if err := c.Get("/mgmt/tm/cm/device/stats", &resp); err != nil {
		log.Printf("Error getting device stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		s := entry.NestedStats.Entries
		device, _ := parsePoolName(s.TmName.Description)

		state := deviceFailoverToValue(s.FailoverState.Description)
		m = append(m, prometheus.MustNewConstMetric(
			deviceFailoverState, prometheus.GaugeValue, float64(state),
			target, device, s.Hostname.Description, s.ManagementIp.Description,
		))

		lf := 0.0
		if s.CurrentLoadFactor.Description == "1" {
			lf = 1
		}
		m = append(m, prometheus.MustNewConstMetric(
			deviceLoadFactor, prometheus.GaugeValue, lf,
			target, device, s.Hostname.Description,
		))
	}

	return m, true
}

func deviceFailoverToValue(state string) int {
	switch strings.ToLower(state) {
	case "active":
		return 2
	case "standby":
		return 1
	default:
		return 0
	}
}

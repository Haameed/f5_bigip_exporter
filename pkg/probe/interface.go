package probe

import (
	"log"
	"strconv"
	"strings"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetInterfaceProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		interfaceStatus = prometheus.NewDesc(
			"bigip_interface_status",
			"Status of the interface (0 = down, 1 = up, 2 = disabled, 3 = uninit, 4 = missing/miss, -1 = unknown)",
			[]string{"target", "name", "bundle", "bundle_speed", "flow_control", "force_gigabit_fiber", "forward_error_correction", "link_traps_enabled", "lldp_admin", "mac_address", "media_active", "media_fixed", "media_max", "media_sfp", "mtu", "port_fwd_mode", "prefer_port", "qinq_ethertype", "sflow_poll_interval", "sflow_poll_interval_global", "stp", "stp_auto_edge_port", "stp_edge_port", "stp_link_type"}, nil,
		)
		countersBitsIn = prometheus.NewDesc(
			"bigip_interface_received_bits_total",
			"Total number of bits received on the interface",
			[]string{"target", "name"}, nil,
		)
		countersBitsOut = prometheus.NewDesc(
			"bigip_interface_sent_bits_total",
			"Total number of bits sent on the interface",
			[]string{"target", "name"}, nil,
		)
		countersPktsIn = prometheus.NewDesc(
			"bigip_interface_received_packets_total",
			"Total number of packets received on the interface",
			[]string{"target", "name"}, nil,
		)
		countersPktsOut = prometheus.NewDesc(
			"bigip_interface_sent_packets_total",
			"Total number of packets sent on the interface",
			[]string{"target", "name"}, nil,
		)
		countersDropsAll = prometheus.NewDesc(
			"bigip_interface_dropped_packets_total",
			"Total number of packets dropped on the interface",
			[]string{"target", "name"}, nil,
		)
		countersErrorsAll = prometheus.NewDesc(
			"bigip_interface_errors_total",
			"Total number of errors on the interface",
			[]string{"target", "name"}, nil,
		)
	)

	type InterfaceStatsEntry struct {
		CountersBitsIn    ValueBased  `json:"counters.bitsIn"`
		CountersBitsOut   ValueBased  `json:"counters.bitsOut"`
		CountersPktsIn    ValueBased  `json:"counters.pktsIn"`
		CountersPktsOut   ValueBased  `json:"counters.pktsOut"`
		CountersDropsAll  ValueBased  `json:"counters.dropsAll"`
		CountersErrorsAll ValueBased  `json:"counters.errorsAll"`
		Status            StringBased `json:"status"`
	}

	type InterfaceStatsNested struct {
		Entries InterfaceStatsEntry `json:"entries"`
	}

	type InterfaceStats struct {
		NestedStats InterfaceStatsNested `json:"nestedStats"`
	}

	type InterfaceStatsResponse struct {
		Kind     string                    `json:"kind"`
		SelfLink string                    `json:"selfLink"`
		Entries  map[string]InterfaceStats `json:"entries"`
	}

	type InterfaceInfo struct {
		Kind                   string `json:"kind"`
		Name                   string `json:"name"`
		FullPath               string `json:"fullPath"`
		Generation             int    `json:"generation"`
		SelfLink               string `json:"selfLink"`
		Bundle                 string `json:"bundle"`
		BundleSpeed            string `json:"bundleSpeed"`
		Enabled                bool   `json:"enabled"`
		Disabled               bool   `json:"disabled"`
		FlowControl            string `json:"flowControl"`
		ForceGigabitFiber      string `json:"forceGigabitFiber"`
		ForwardErrorCorrection string `json:"forwardErrorCorrection"`
		IfIndex                int    `json:"ifIndex"`
		LacpPortPriority       int    `json:"lacpPortPriority"`
		LinkTrapsEnabled       string `json:"linkTrapsEnabled"`
		LldpAdmin              string `json:"lldpAdmin"`
		LldpTlvmap             int    `json:"lldpTlvmap"`
		MacAddress             string `json:"macAddress"`
		MediaActive            string `json:"mediaActive"`
		MediaFixed             string `json:"mediaFixed"`
		MediaMax               string `json:"mediaMax"`
		MediaSfp               string `json:"mediaSfp"`
		Mtu                    int    `json:"mtu"`
		PortFwdMode            string `json:"portFwdMode"`
		PreferPort             string `json:"preferPort"`
		QinqEthertype          string `json:"qinqEthertype"`
		Sflow                  struct {
			PollInterval       int    `json:"pollInterval"`
			PollIntervalGlobal string `json:"pollIntervalGlobal"`
		} `json:"sflow"`
		Stp             string `json:"stp"`
		StpAutoEdgePort string `json:"stpAutoEdgePort"`
		StpEdgePort     string `json:"stpEdgePort"`
		StpLinkType     string `json:"stpLinkType"`
	}
	type InterfaceInfoResponse struct {
		Kind     string          `json:"kind"`
		SelfLink string          `json:"selfLink"`
		Items    []InterfaceInfo `json:"items"`
	}

	var m []prometheus.Metric
	var resp InterfaceInfoResponse
	if err := c.Get("/mgmt/tm/net/interface", &resp); err != nil {
		log.Printf("Error getting interfaces list: %v", err)
		return m, false
	}

	var statsResp InterfaceStatsResponse
	if err := c.Get("/mgmt/tm/net/interface/stats", &statsResp); err != nil {
		log.Printf("Error getting interface stats: %v", err)
		return m, false
	}

	statsByName := make(map[string]InterfaceStatsEntry, len(statsResp.Entries))
	for name, stat := range statsResp.Entries {
		// Entry keys are the interface's fullPath (e.g. "/Common/1.1"); use
		// the trailing path segment to match InterfaceInfo.Name.
		parts := strings.Split(name, "/")
		statsByName[parts[len(parts)-1]] = stat.NestedStats.Entries
	}

	for _, info := range resp.Items {
		stats := statsByName[info.Name]

		m = append(m, prometheus.MustNewConstMetric(
			interfaceStatus,
			prometheus.GaugeValue,
			float64(interfaceStatusToValue(stats.Status.Description)),
			target,
			info.Name,
			info.Bundle,
			info.BundleSpeed,
			info.FlowControl,
			info.ForceGigabitFiber,
			info.ForwardErrorCorrection,
			info.LinkTrapsEnabled,
			info.LldpAdmin,
			info.MacAddress,
			info.MediaActive,
			info.MediaFixed,
			info.MediaMax,
			info.MediaSfp,
			strconv.Itoa(info.Mtu),
			info.PortFwdMode,
			info.PreferPort,
			info.QinqEthertype,
			strconv.Itoa(info.Sflow.PollInterval),
			info.Sflow.PollIntervalGlobal,
			info.Stp,
			info.StpAutoEdgePort,
			info.StpEdgePort,
			info.StpLinkType,
		))

		m = append(m, prometheus.MustNewConstMetric(
			countersBitsIn,
			prometheus.CounterValue,
			float64(stats.CountersBitsIn.Value),
			target, info.Name,
		))
		m = append(m, prometheus.MustNewConstMetric(
			countersBitsOut,
			prometheus.CounterValue,
			float64(stats.CountersBitsOut.Value),
			target, info.Name,
		))
		m = append(m, prometheus.MustNewConstMetric(
			countersPktsIn,
			prometheus.CounterValue,
			float64(stats.CountersPktsIn.Value),
			target, info.Name,
		))
		m = append(m, prometheus.MustNewConstMetric(
			countersPktsOut,
			prometheus.CounterValue,
			float64(stats.CountersPktsOut.Value),
			target, info.Name,
		))
		m = append(m, prometheus.MustNewConstMetric(
			countersDropsAll,
			prometheus.CounterValue,
			float64(stats.CountersDropsAll.Value),
			target, info.Name,
		))
		m = append(m, prometheus.MustNewConstMetric(
			countersErrorsAll,
			prometheus.CounterValue,
			float64(stats.CountersErrorsAll.Value),
			target, info.Name,
		))
	}

	return m, true
}

// interfaceStatusToValue maps the BIG-IP interface "status" field to a
// numeric gauge value:
//
//	 0 = down     (interface is enabled but the link is not up - a real failure)
//	 1 = up       (interface is enabled and the link is up)
//	 2 = disabled (interface has been administratively disabled)
//	 3 = uninit   (interface has not been initialized yet, e.g. right after boot)
//	 4 = miss     (interface is missing/not detected, e.g. removed hardware or bad transceiver)
//	-1 = unknown  (any other/unrecognized status string)
//
// Without separate states for "disabled", "uninit" and "miss", those
// conditions (which are either expected/intentional or indicate a hardware
// problem rather than a simple link-down) would be indistinguishable from a
// genuine link failure on an enabled interface, causing false alerts and
// confusion.
func interfaceStatusToValue(status string) int {
	switch strings.ToLower(status) {
	case "up":
		return 1
	case "down":
		return 0
	case "disabled":
		return 2
	case "uninit":
		return 3
	case "miss":
		return 4
	default:
		return -1
	}
}

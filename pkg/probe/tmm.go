package probe

import (
	"log"
	"strconv"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetTMMProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		tmmCPUUsage = prometheus.NewDesc(
			"bigip_tmm_cpu_usage_ratio",
			"TMM (Traffic Management Microkernel) CPU usage ratio",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmFiveSecAvgUsage = prometheus.NewDesc(
			"bigip_tmm_five_sec_avg_usage_ratio",
			"TMM CPU usage ratio averaged over 5 seconds",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmOneMinAvgUsage = prometheus.NewDesc(
			"bigip_tmm_one_min_avg_usage_ratio",
			"TMM CPU usage ratio averaged over 1 minute",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmFiveMinAvgUsage = prometheus.NewDesc(
			"bigip_tmm_five_min_avg_usage_ratio",
			"TMM CPU usage ratio averaged over 5 minutes",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmMemoryTotal = prometheus.NewDesc(
			"bigip_tmm_memory_total_bytes",
			"Total memory allocated to the TMM in bytes",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmMemoryUsed = prometheus.NewDesc(
			"bigip_tmm_memory_used_bytes",
			"Memory used by the TMM in bytes",
			[]string{"target", "tmm_id"}, nil,
		)
		tmmNpus = prometheus.NewDesc(
			"bigip_tmm_npus",
			"Number of Network Processing Units (NPUs) for the TMM",
			[]string{"target", "tmm_id"}, nil,
		)
	)

	type TMMStatsEntry struct {
		Cpu                  ValueBased  `json:"cpu"`
		FiveMinAvgUsageRatio ValueBased  `json:"fiveMinAvgUsageRatio"`
		FiveSecAvgUsageRatio ValueBased  `json:"fiveSecAvgUsageRatio"`
		MemoryTotal          ValueBased  `json:"memoryTotal"`
		MemoryUsed           ValueBased  `json:"memoryUsed"`
		Npus                 ValueBased  `json:"npus"`
		OneMinAvgUsageRatio  ValueBased  `json:"oneMinAvgUsageRatio"`
		Tmid                 ValueBased  `json:"tmid"`
		TmmId                StringBased `json:"tmmId"`
		TmmPid               ValueBased  `json:"tmmPid"`
	}

	type TMMNestedStats struct {
		Entries TMMStatsEntry `json:"entries"`
	}

	type TMMStats struct {
		NestedStats TMMNestedStats `json:"nestedStats"`
	}

	type TMMInfoResponse struct {
		Kind     string              `json:"kind"`
		SelfLink string              `json:"selfLink"`
		Entries  map[string]TMMStats `json:"entries"`
	}

	var m []prometheus.Metric

	var resp TMMInfoResponse
	if err := c.Get("/mgmt/tm/sys/tmm-info/stats", &resp); err != nil {
		log.Printf("Error getting TMM stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		stats := entry.NestedStats.Entries
		tmmID := stats.TmmId.Description
		if tmmID == "" {
			tmmID = strconv.Itoa(stats.Tmid.Value)
		}

		// The iControl REST API reports these "usage ratio" fields on a
		// 0-100 scale (they are actually percentages); divide by 100 to
		// expose true 0-1 ratios, per Prometheus naming conventions.
		m = append(m, prometheus.MustNewConstMetric(
			tmmCPUUsage, prometheus.GaugeValue,
			float64(stats.Cpu.Value)/100, target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmFiveSecAvgUsage, prometheus.GaugeValue,
			float64(stats.FiveSecAvgUsageRatio.Value)/100, target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmOneMinAvgUsage, prometheus.GaugeValue,
			float64(stats.OneMinAvgUsageRatio.Value)/100, target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmFiveMinAvgUsage, prometheus.GaugeValue,
			float64(stats.FiveMinAvgUsageRatio.Value)/100, target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmMemoryTotal, prometheus.GaugeValue,
			float64(stats.MemoryTotal.Value), target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmMemoryUsed, prometheus.GaugeValue,
			float64(stats.MemoryUsed.Value), target, tmmID,
		))
		m = append(m, prometheus.MustNewConstMetric(
			tmmNpus, prometheus.GaugeValue,
			float64(stats.Npus.Value), target, tmmID,
		))
	}

	return m, true
}

package probe

import (
	"log"
	"strconv"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetComputeProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		cpuCount = prometheus.NewDesc(
			"bigip_system_total_cpu_count",
			"Total nubmber of CPU Cores",
			[]string{"target", "host_id"}, nil,
		)
		activeCPUCount = prometheus.NewDesc(
			"bigip_system_active_cpu_count",
			"Total number of active CPU Cores",
			[]string{"target", "host_id"}, nil,
		)
		totalMemory = prometheus.NewDesc(
			"bigip_system_memory_total_bytes",
			"Total allocated memory in bytes",
			[]string{"target", "host_id"}, nil,
		)
		usedMemory = prometheus.NewDesc(
			"bigip_system_memory_used_bytes",
			"Total used memory in bytes",
			[]string{"target", "host_id"}, nil,
		)
		cpuFiveMinAvgIdle = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_idle_percent",
			"CPU idle percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgIowait = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_iowait_percent",
			"CPU iowait percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgIrq = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_irq_percent",
			"CPU IRQ percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgNiced = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_niced_percent",
			"CPU niced percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgSoftirq = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_softirq_percent",
			"CPU soft IRQ percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgStolen = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_stolen_percent",
			"CPU stolen percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgSystem = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_system_percent",
			"CPU system percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveMinAvgUser = prometheus.NewDesc(
			"bigip_cpu_five_min_avg_user_percent",
			"CPU user percentage averaged over 5 minutes",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgIdle = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_idle_percent",
			"CPU idle percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgIowait = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_iowait_percent",
			"CPU iowait percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgIrq = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_irq_percent",
			"CPU IRQ percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgNiced = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_niced_percent",
			"CPU niced percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgRatio = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_ratio",
			"CPU ratio averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgSoftirq = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_softirq_percent",
			"CPU soft IRQ percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgStolen = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_stolen_percent",
			"CPU stolen percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgSystem = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_system_percent",
			"CPU system percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuFiveSecAvgUser = prometheus.NewDesc(
			"bigip_cpu_five_sec_avg_user_percent",
			"CPU user percentage averaged over 5 seconds",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuIdle = prometheus.NewDesc(
			"bigip_cpu_idle_ticks_total",
			"Total CPU idle ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuIowait = prometheus.NewDesc(
			"bigip_cpu_iowait_ticks_total",
			"Total CPU iowait ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuIrq = prometheus.NewDesc(
			"bigip_cpu_irq_ticks_total",
			"Total CPU IRQ ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuNiced = prometheus.NewDesc(
			"bigip_cpu_niced_ticks_total",
			"Total CPU niced ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgIdle = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_idle_percent",
			"CPU idle percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgIowait = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_iowait_percent",
			"CPU iowait percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgIrq = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_irq_percent",
			"CPU IRQ percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgNiced = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_niced_percent",
			"CPU niced percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgSoftirq = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_softirq_percent",
			"CPU soft IRQ percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgStolen = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_stolen_percent",
			"CPU stolen percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgSystem = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_system_percent",
			"CPU system percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuOneMinAvgUser = prometheus.NewDesc(
			"bigip_cpu_one_min_avg_user_percent",
			"CPU user percentage averaged over 1 minute",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuSoftirq = prometheus.NewDesc(
			"bigip_cpu_softirq_ticks_total",
			"Total CPU soft IRQ ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuStolen = prometheus.NewDesc(
			"bigip_cpu_stolen_ticks_total",
			"Total CPU stolen ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuSystem = prometheus.NewDesc(
			"bigip_cpu_system_ticks_total",
			"Total CPU system ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
		cpuUser = prometheus.NewDesc(
			"bigip_cpu_user_ticks_total",
			"Total CPU user ticks",
			[]string{"target", "host_id", "cpu_id"}, nil,
		)
	)

	type ValueBased struct {
		Value int `json:"value"`
	}

	type StringBased struct {
		Description string `json:"description"`
	}

	type CpuStats struct {
		CpuId             ValueBased `json:"cpuId"`
		FiveMinAvgIdle    ValueBased `json:"fiveMinAvgIdle"`
		FiveMinAvgIowait  ValueBased `json:"fiveMinAvgIowait"`
		FiveMinAvgIrq     ValueBased `json:"fiveMinAvgIrq"`
		FiveMinAvgNiced   ValueBased `json:"fiveMinAvgNiced"`
		FiveMinAvgSoftirq ValueBased `json:"fiveMinAvgSoftirq"`
		FiveMinAvgStolen  ValueBased `json:"fiveMinAvgStolen"`
		FiveMinAvgSystem  ValueBased `json:"fiveMinAvgSystem"`
		FiveMinAvgUser    ValueBased `json:"fiveMinAvgUser"`
		FiveSecAvgIdle    ValueBased `json:"fiveSecAvgIdle"`
		FiveSecAvgIowait  ValueBased `json:"fiveSecAvgIowait"`
		FiveSecAvgIrq     ValueBased `json:"fiveSecAvgIrq"`
		FiveSecAvgNiced   ValueBased `json:"fiveSecAvgNiced"`
		FiveSecAvgRatio   ValueBased `json:"fiveSecAvgRatio"`
		FiveSecAvgSoftirq ValueBased `json:"fiveSecAvgSoftirq"`
		FiveSecAvgStolen  ValueBased `json:"fiveSecAvgStolen"`
		FiveSecAvgSystem  ValueBased `json:"fiveSecAvgSystem"`
		FiveSecAvgUser    ValueBased `json:"fiveSecAvgUser"`
		Idle              ValueBased `json:"idle"`
		Iowait            ValueBased `json:"iowait"`
		Irq               ValueBased `json:"irq"`
		Niced             ValueBased `json:"niced"`
		OneMinAvgIdle     ValueBased `json:"oneMinAvgIdle"`
		OneMinAvgIowait   ValueBased `json:"oneMinAvgIowait"`
		OneMinAvgIrq      ValueBased `json:"oneMinAvgIrq"`
		OneMinAvgNiced    ValueBased `json:"oneMinAvgNiced"`
		OneMinAvgSoftirq  ValueBased `json:"oneMinAvgSoftirq"`
		OneMinAvgStolen   ValueBased `json:"oneMinAvgStolen"`
		OneMinAvgSystem   ValueBased `json:"oneMinAvgSystem"`
		OneMinAvgUser     ValueBased `json:"oneMinAvgUser"`
		Softirq           ValueBased `json:"softirq"`
		Stolen            ValueBased `json:"stolen"`
		System            ValueBased `json:"system"`
		User              ValueBased `json:"user"`
	}

	type CpuEntry struct {
		NestedStats struct {
			Entries CpuStats `json:"entries"`
		} `json:"nestedStats"`
	}

	type CpuInfoStats struct {
		NestedStats struct {
			Entries map[string]CpuEntry `json:"entries"`
		} `json:"nestedStats"`
	}

	type StatEntry struct {
		ActiveCpuCount ValueBased   `json:"activeCpuCount"`
		CPUCount       ValueBased   `json:"cpuCount"`
		HostID         StringBased  `json:"hostId"`
		MemoryTotal    ValueBased   `json:"memoryTotal"`
		MemoryUsed     ValueBased   `json:"memoryUsed"`
		CPUInfo        CpuInfoStats `json:"https://localhost/mgmt/tm/sys/hostInfo/0/cpuInfo/stats"`
	}

	type NestedStats struct {
		Entries StatEntry `json:"entries"`
	}

	type CPUStatsEntries struct {
		NestedStats NestedStats `json:"nestedStats"`
	}

	type HostInfoResponse struct {
		Kind     string                     `json:"kind"`
		SelfLink string                     `json:"selfLink"`
		Entries  map[string]CPUStatsEntries `json:"entries"`
	}
	var response HostInfoResponse
	if err := c.Get("/mgmt/tm/sys/host-info/stats", &response); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	var m []prometheus.Metric

	for _, entry := range response.Entries {
		stats := entry.NestedStats.Entries
		m = append(m, prometheus.MustNewConstMetric(cpuCount, prometheus.GaugeValue, float64(stats.CPUCount.Value), target, stats.HostID.Description))
		m = append(m, prometheus.MustNewConstMetric(activeCPUCount, prometheus.GaugeValue, float64(stats.ActiveCpuCount.Value), target, stats.HostID.Description))
		m = append(m, prometheus.MustNewConstMetric(totalMemory, prometheus.GaugeValue, float64(stats.MemoryTotal.Value), target, stats.HostID.Description))
		m = append(m, prometheus.MustNewConstMetric(usedMemory, prometheus.GaugeValue, float64(stats.MemoryUsed.Value), target, stats.HostID.Description))
		cpuEntries := stats.CPUInfo.NestedStats.Entries
		for _, cpuEntry := range cpuEntries {
			cpuStats := cpuEntry.NestedStats.Entries
			cpuID := strconv.Itoa(cpuStats.CpuId.Value)
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgIdle, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgIdle.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgIowait, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgIowait.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgIrq, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgIrq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgNiced, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgNiced.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgSoftirq, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgSoftirq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgStolen, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgStolen.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgSystem, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgSystem.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveMinAvgUser, prometheus.GaugeValue, float64(cpuStats.FiveMinAvgUser.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgIdle, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgIdle.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgIowait, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgIowait.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgIrq, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgIrq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgNiced, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgNiced.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgRatio, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgRatio.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgSoftirq, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgSoftirq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgStolen, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgStolen.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgSystem, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgSystem.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuFiveSecAvgUser, prometheus.GaugeValue, float64(cpuStats.FiveSecAvgUser.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuIdle, prometheus.CounterValue, float64(cpuStats.Idle.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuIowait, prometheus.CounterValue, float64(cpuStats.Iowait.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuIrq, prometheus.CounterValue, float64(cpuStats.Irq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuNiced, prometheus.CounterValue, float64(cpuStats.Niced.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgIdle, prometheus.GaugeValue, float64(cpuStats.OneMinAvgIdle.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgIowait, prometheus.GaugeValue, float64(cpuStats.OneMinAvgIowait.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgIrq, prometheus.GaugeValue, float64(cpuStats.OneMinAvgIrq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgNiced, prometheus.GaugeValue, float64(cpuStats.OneMinAvgNiced.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgSoftirq, prometheus.GaugeValue, float64(cpuStats.OneMinAvgSoftirq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgStolen, prometheus.GaugeValue, float64(cpuStats.OneMinAvgStolen.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgSystem, prometheus.GaugeValue, float64(cpuStats.OneMinAvgSystem.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuOneMinAvgUser, prometheus.GaugeValue, float64(cpuStats.OneMinAvgUser.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuSoftirq, prometheus.CounterValue, float64(cpuStats.Softirq.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuStolen, prometheus.CounterValue, float64(cpuStats.Stolen.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuSystem, prometheus.CounterValue, float64(cpuStats.System.Value), target, stats.HostID.Description, cpuID))
			m = append(m, prometheus.MustNewConstMetric(cpuUser, prometheus.CounterValue, float64(cpuStats.User.Value), target, stats.HostID.Description, cpuID))
		}
	}
	return m, true
}

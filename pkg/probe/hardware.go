package probe

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetHardwareProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		hardwareAvailable = prometheus.NewDesc(
			"bigip_hardware_available",
			"Whether physical hardware sensor data is available (1=hardware appliance, 0=Virtual Edition / vCMP guest / no data)",
			[]string{"target"}, nil,
		)
		fanStatus = prometheus.NewDesc(
			"bigip_hardware_fan_status",
			"Chassis fan status (1=up, 0=not up)",
			[]string{"target", "index"}, nil,
		)
		powerStatus = prometheus.NewDesc(
			"bigip_hardware_power_supply_status",
			"Chassis power supply status (1=up, 0=not up)",
			[]string{"target", "index", "current_status"}, nil,
		)
		temperature = prometheus.NewDesc(
			"bigip_hardware_temperature_celsius",
			"Chassis temperature in degrees Celsius",
			[]string{"target", "index", "location"}, nil,
		)
		temperatureHigh = prometheus.NewDesc(
			"bigip_hardware_temperature_high_limit_celsius",
			"Configured high temperature limit in degrees Celsius",
			[]string{"target", "index", "location"}, nil,
		)
		humidity = prometheus.NewDesc(
			"bigip_hardware_humidity_ratio",
			"Chassis relative humidity ratio",
			[]string{"target", "index"}, nil,
		)
		deviceInfo = prometheus.NewDesc(
			"bigip_device_info",
			"Device hardware information (constant 1; metadata exposed via labels)",
			[]string{"target", "marketing_name", "platform", "chassis_serial", "base_mac", "bios_revision"}, nil,
		)
	)

	type itemEntries struct {
		Index         ValueBased  `json:"index"`
		Status        StringBased `json:"status"`
		CurrentStatus StringBased `json:"currentStatus"`
		Temperature   ValueBased  `json:"temperature"`
		HiLimit       ValueBased  `json:"hiLimit"`
		LoLimit       ValueBased  `json:"loLimit"`
		Location      StringBased `json:"location"`
		Humidity      ValueBased  `json:"humidity"`
	}
	type item struct {
		NestedStats struct {
			Entries itemEntries `json:"entries"`
		} `json:"nestedStats"`
	}

	type platformEntries struct {
		BaseMac       StringBased `json:"baseMac"`
		BiosRev       StringBased `json:"biosRev"`
		MarketingName StringBased `json:"marketingName"`
	}
	type platformItem struct {
		NestedStats struct {
			Entries platformEntries `json:"entries"`
		} `json:"nestedStats"`
	}

	type systemInfoEntries struct {
		ChassisSerial StringBased `json:"bigipChassisSerialNum"`
		Platform      StringBased `json:"platform"`
	}
	type systemInfoItem struct {
		NestedStats struct {
			Entries systemInfoEntries `json:"entries"`
		} `json:"nestedStats"`
	}

	type topEntry struct {
		NestedStats struct {
			Entries map[string]json.RawMessage `json:"entries"`
		} `json:"nestedStats"`
	}
	type hardwareResponse struct {
		Entries map[string]topEntry `json:"entries"`
	}

	var m []prometheus.Metric

	var resp hardwareResponse
	if err := c.Get("/mgmt/tm/sys/hardware", &resp); err != nil {
		// On VE / vCMP guest this endpoint may be unavailable. We do NOT fail
		// the probe — physical sensors simply don't exist there.
		log.Printf("Hardware stats unavailable for %s (likely Virtual Edition / vCMP guest): %v", target, err)
		m = append(m, prometheus.MustNewConstMetric(hardwareAvailable, prometheus.GaugeValue, 0, target))
		return m, true
	}

	var marketingName, platform, chassisSerial, baseMac, biosRev string
	hasHardware := false

	for key, top := range resp.Entries {
		switch {
		case strings.HasSuffix(key, "chassis-fan-status-index"):
			for _, raw := range top.NestedStats.Entries {
				var it item
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				hasHardware = true
				e := it.NestedStats.Entries
				val := 0.0
				if strings.ToLower(e.Status.Description) == "up" {
					val = 1
				}
				m = append(m, prometheus.MustNewConstMetric(
					fanStatus, prometheus.GaugeValue, val,
					target, strconv.Itoa(e.Index.Value),
				))
			}

		case strings.HasSuffix(key, "chassis-power-supply-status-index"):
			for _, raw := range top.NestedStats.Entries {
				var it item
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				hasHardware = true
				e := it.NestedStats.Entries
				val := 0.0
				if strings.ToLower(e.Status.Description) == "up" {
					val = 1
				}
				m = append(m, prometheus.MustNewConstMetric(
					powerStatus, prometheus.GaugeValue, val,
					target, strconv.Itoa(e.Index.Value), e.CurrentStatus.Description,
				))
			}

		case strings.HasSuffix(key, "chassis-temperature-status-index"):
			for _, raw := range top.NestedStats.Entries {
				var it item
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				hasHardware = true
				e := it.NestedStats.Entries
				idx := strconv.Itoa(e.Index.Value)
				m = append(m, prometheus.MustNewConstMetric(
					temperature, prometheus.GaugeValue, float64(e.Temperature.Value),
					target, idx, e.Location.Description,
				))
				m = append(m, prometheus.MustNewConstMetric(
					temperatureHigh, prometheus.GaugeValue, float64(e.HiLimit.Value),
					target, idx, e.Location.Description,
				))
			}

		case strings.HasSuffix(key, "chassis-humidity-status-index"):
			for _, raw := range top.NestedStats.Entries {
				var it item
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				hasHardware = true
				e := it.NestedStats.Entries
				// Humidity is reported by the device as a 0-100 percentage;
				// divide by 100 to expose a true 0-1 ratio.
				m = append(m, prometheus.MustNewConstMetric(
					humidity, prometheus.GaugeValue, float64(e.Humidity.Value)/100,
					target, strconv.Itoa(e.Index.Value),
				))
			}

		case strings.HasSuffix(key, "/platform"):
			for _, raw := range top.NestedStats.Entries {
				var it platformItem
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				marketingName = strings.TrimSpace(it.NestedStats.Entries.MarketingName.Description)
				baseMac = strings.TrimSpace(it.NestedStats.Entries.BaseMac.Description)
				biosRev = strings.TrimSpace(it.NestedStats.Entries.BiosRev.Description)
			}

		case strings.HasSuffix(key, "/system-info"):
			for _, raw := range top.NestedStats.Entries {
				var it systemInfoItem
				if err := json.Unmarshal(raw, &it); err != nil {
					continue
				}
				chassisSerial = strings.TrimSpace(it.NestedStats.Entries.ChassisSerial.Description)
				platform = strings.TrimSpace(it.NestedStats.Entries.Platform.Description)
			}
		}
	}

	availVal := 0.0
	if hasHardware {
		availVal = 1
	}
	m = append(m, prometheus.MustNewConstMetric(hardwareAvailable, prometheus.GaugeValue, availVal, target))

	if marketingName != "" || platform != "" || chassisSerial != "" {
		m = append(m, prometheus.MustNewConstMetric(
			deviceInfo, prometheus.GaugeValue, 1,
			target, marketingName, platform, chassisSerial, baseMac, biosRev,
		))
	}

	return m, true
}

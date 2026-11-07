package probe

import (
	"log"
	"strconv"
	"time"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

// licenseDateLayout matches the date format returned by the BIG-IP iControl
// REST API for license dates, e.g. "2026/12/03".
const licenseDateLayout = "2006/01/02"

// GetLicenseProbe exposes BIG-IP license metrics (registration/platform info
// and, where reported, license start/expiration dates).
//
// Not every BIG-IP version/license type reports an expiration date via
// iControl REST (see
// https://github.com/Haameed/f5_bigip_exporter/issues/1). When that's the
// case we don't fail the probe: we simply log a warning and expose
// bigip_license_available=0 so the absence is still visible in metrics,
// without breaking scrapes for those devices.
func GetLicenseProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		licenseAvailable = prometheus.NewDesc(
			"bigip_license_available",
			"Whether license expiration data is reported by this device/version (1=available, 0=not reported)",
			[]string{"target"}, nil,
		)
		licenseInfo = prometheus.NewDesc(
			"bigip_license_info",
			"License information (constant 1; metadata exposed via labels)",
			[]string{"target", "registration_key", "platform_id", "licensed_version", "appliance_serial_number"}, nil,
		)
		licenseStartTimestamp = prometheus.NewDesc(
			"bigip_license_start_timestamp_seconds",
			"Unix timestamp when the license became valid",
			[]string{"target"}, nil,
		)
		licensedOnTimestamp = prometheus.NewDesc(
			"bigip_licensed_on_timestamp_seconds",
			"Unix timestamp when the device was licensed",
			[]string{"target"}, nil,
		)
		licenseServiceCheckTimestamp = prometheus.NewDesc(
			"bigip_license_service_check_timestamp_seconds",
			"Unix timestamp of the last license service check date",
			[]string{"target"}, nil,
		)
		licenseDailyRenewNotifPeriod = prometheus.NewDesc(
			"bigip_license_daily_renew_notification_period_days",
			"Number of days before expiration that daily renewal notifications start",
			[]string{"target"}, nil,
		)
		licenseExpirationTimestamp = prometheus.NewDesc(
			"bigip_license_expiration_timestamp_seconds",
			"Unix timestamp when the license expires",
			[]string{"target"}, nil,
		)
		licenseExpirationDays = prometheus.NewDesc(
			"bigip_license_expiration_days",
			"Number of days until the license expires (0 if already expired)",
			[]string{"target"}, nil,
		)
		licenseExpired = prometheus.NewDesc(
			"bigip_license_expired",
			"License expiration status (1=expired, 0=valid)",
			[]string{"target", "expiration_date"}, nil,
		)
	)

	type licenseLeafEntries struct {
		ApplianceSerialNumber StringBased `json:"applianceSerialNumber"`
		DailyRenewNotifPeriod StringBased `json:"dailyRenewNotifPeriod"`
		LicenseEndDate        StringBased `json:"licenseEndDate"`
		LicenseStartDate      StringBased `json:"licenseStartDate"`
		LicensedOnDate        StringBased `json:"licensedOnDate"`
		LicensedVersion       StringBased `json:"licensedVersion"`
		PlatformId            StringBased `json:"platformId"`
		RegistrationKey       StringBased `json:"registrationKey"`
		ServiceCheckDate      StringBased `json:"serviceCheckDate"`
	}

	type licenseNested struct {
		Entries licenseLeafEntries `json:"entries"`
	}

	type licenseStats struct {
		NestedStats licenseNested `json:"nestedStats"`
	}

	type licenseResponse struct {
		Kind    string                  `json:"kind"`
		Entries map[string]licenseStats `json:"entries"`
	}

	var m []prometheus.Metric
	var resp licenseResponse
	if err := c.Get("/mgmt/tm/sys/license", &resp); err != nil {
		log.Printf("Warning: license information unavailable for %s: %v", target, err)
		m = append(m, prometheus.MustNewConstMetric(licenseAvailable, prometheus.GaugeValue, 0, target))
		return m, true
	}

	if len(resp.Entries) == 0 {
		log.Printf("Warning: no license entries reported for %s; skipping license metrics", target)
		m = append(m, prometheus.MustNewConstMetric(licenseAvailable, prometheus.GaugeValue, 0, target))
		return m, true
	}

	hasExpirationData := false

	for _, entry := range resp.Entries {
		s := entry.NestedStats.Entries

		if s.RegistrationKey.Description != "" || s.PlatformId.Description != "" ||
			s.LicensedVersion.Description != "" || s.ApplianceSerialNumber.Description != "" {
			m = append(m, prometheus.MustNewConstMetric(
				licenseInfo, prometheus.GaugeValue, 1,
				target, s.RegistrationKey.Description, s.PlatformId.Description,
				s.LicensedVersion.Description, s.ApplianceSerialNumber.Description,
			))
		}

		if ts, ok := parseLicenseDate(s.LicenseStartDate.Description); ok {
			m = append(m, prometheus.MustNewConstMetric(licenseStartTimestamp, prometheus.GaugeValue, float64(ts), target))
		}

		if ts, ok := parseLicenseDate(s.LicensedOnDate.Description); ok {
			m = append(m, prometheus.MustNewConstMetric(licensedOnTimestamp, prometheus.GaugeValue, float64(ts), target))
		}

		if ts, ok := parseLicenseDate(s.ServiceCheckDate.Description); ok {
			m = append(m, prometheus.MustNewConstMetric(licenseServiceCheckTimestamp, prometheus.GaugeValue, float64(ts), target))
		}

		if period, err := strconv.ParseFloat(s.DailyRenewNotifPeriod.Description, 64); err == nil {
			m = append(m, prometheus.MustNewConstMetric(licenseDailyRenewNotifPeriod, prometheus.GaugeValue, period, target))
		}

		// licenseEndDate isn't reported by every BIG-IP version/license type
		// (see the linked issue). Only expose expiration metrics when it's
		// actually present and parseable.
		if ts, ok := parseLicenseDate(s.LicenseEndDate.Description); ok {
			hasExpirationData = true
			currentTime := time.Now().Unix()
			daysUntilExpiry := float64(ts-currentTime) / 86400.0

			isExpired := float64(0)
			if ts < currentTime {
				isExpired = 1
				daysUntilExpiry = 0
			}

			m = append(m, prometheus.MustNewConstMetric(licenseExpirationTimestamp, prometheus.GaugeValue, float64(ts), target))
			m = append(m, prometheus.MustNewConstMetric(licenseExpirationDays, prometheus.GaugeValue, daysUntilExpiry, target))
			m = append(m, prometheus.MustNewConstMetric(licenseExpired, prometheus.GaugeValue, isExpired, target, s.LicenseEndDate.Description))
		}
	}

	if !hasExpirationData {
		log.Printf("Warning: license expiration date not reported by %s (not available on this BIG-IP version/license type); skipping expiration metrics", target)
	}

	availVal := 0.0
	if hasExpirationData {
		availVal = 1
	}
	m = append(m, prometheus.MustNewConstMetric(licenseAvailable, prometheus.GaugeValue, availVal, target))

	return m, true
}

// parseLicenseDate parses a BIG-IP license date string (e.g. "2026/12/03")
// into a Unix timestamp. It returns ok=false for empty or unparseable values,
// so callers can skip emitting metrics for fields the device doesn't report.
func parseLicenseDate(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}
	t, err := time.Parse(licenseDateLayout, s)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}

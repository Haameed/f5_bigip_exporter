package probe

import (
	"log"
	"strings"
	"time"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetCertificateProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		certExpirationTimestamp = prometheus.NewDesc(
			"bigip_certificate_expiration_timestamp_seconds",
			"Unix timestamp when certificate expires",
			[]string{"target", "name", "partition"}, nil,
		)

		certExpirationDays = prometheus.NewDesc(
			"bigip_certificate_expiration_days",
			"Number of days until certificate expires",
			[]string{"target", "name", "partition"}, nil,
		)

		certExpired = prometheus.NewDesc(
			"bigip_certificate_expired",
			"Certificate expiration status (1=expired, 0=valid)",
			[]string{"target", "name", "partition", "expiration_date"}, nil,
		)

		certKeySize = prometheus.NewDesc(
			"bigip_certificate_key_size_bits",
			"Certificate key size in bits",
			[]string{"target", "name", "partition", "key_type"}, nil,
		)

		certIsBundle = prometheus.NewDesc(
			"bigip_certificate_is_bundle",
			"Whether certificate is a bundle (1=bundle, 0=single)",
			[]string{"target", "name", "partition"}, nil,
		)

		certVersion = prometheus.NewDesc(
			"bigip_certificate_version",
			"Certificate version number",
			[]string{"target", "name", "partition"}, nil,
		)

		certSize = prometheus.NewDesc(
			"bigip_certificate_size_bytes",
			"Certificate file size in bytes",
			[]string{"target", "name", "partition"}, nil,
		)
	)

	type CertificateEntry struct {
		Kind                    string `json:"kind"`
		Name                    string `json:"name"`
		Partition               string `json:"partition"`
		FullPath                string `json:"fullPath"`
		Generation              int    `json:"generation"`
		SelfLink                string `json:"selfLink"`
		CertificateKeyCurveName string `json:"certificateKeyCurveName"`
		CertificateKeySize      int    `json:"certificateKeySize"`
		Checksum                string `json:"checksum"`
		CreateTime              string `json:"createTime"`
		CreatedBy               string `json:"createdBy"`
		ExpirationDate          int64  `json:"expirationDate"`
		ExpirationString        string `json:"expirationString"`
		Fingerprint             string `json:"fingerprint"`
		IsBundle                string `json:"isBundle"`
		Issuer                  string `json:"issuer"`
		KeyType                 string `json:"keyType"`
		LastUpdateTime          string `json:"lastUpdateTime"`
		Mode                    int    `json:"mode"`
		Revision                int    `json:"revision"`
		SerialNumber            string `json:"serialNumber"`
		Size                    int    `json:"size"`
		Subject                 string `json:"subject"`
		UpdatedBy               string `json:"updatedBy"`
		Version                 int    `json:"version"`
	}

	type CertificateResponse struct {
		Kind     string             `json:"kind"`
		SelfLink string             `json:"selfLink"`
		Items    []CertificateEntry `json:"items"`
	}

	var m []prometheus.Metric

	var certResp CertificateResponse
	if err := c.Get("/mgmt/tm/sys/file/ssl-cert", &certResp); err != nil {
		log.Printf("Error getting certificate stats: %v", err)
		return m, false
	}

	currentTime := time.Now().Unix()

	for _, cert := range certResp.Items {
		daysUntilExpiry := float64(cert.ExpirationDate-currentTime) / 86400.0

		isExpired := float64(0)
		if cert.ExpirationDate < currentTime {
			isExpired = 1
			daysUntilExpiry = 0
		}

		m = append(m, prometheus.MustNewConstMetric(
			certExpirationTimestamp,
			prometheus.GaugeValue,
			float64(cert.ExpirationDate),
			target, cert.Name, cert.Partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			certExpirationDays,
			prometheus.GaugeValue,
			daysUntilExpiry,
			target, cert.Name, cert.Partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			certExpired,
			prometheus.GaugeValue,
			isExpired,
			target, cert.Name, cert.Partition, cert.ExpirationString,
		))

		m = append(m, prometheus.MustNewConstMetric(
			certKeySize,
			prometheus.GaugeValue,
			float64(cert.CertificateKeySize),
			target, cert.Name, cert.Partition, cert.KeyType,
		))

		isBundleValue := float64(0)
		if strings.ToLower(cert.IsBundle) == "true" {
			isBundleValue = 1
		}
		m = append(m, prometheus.MustNewConstMetric(
			certIsBundle,
			prometheus.GaugeValue,
			isBundleValue,
			target, cert.Name, cert.Partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			certVersion,
			prometheus.GaugeValue,
			float64(cert.Version),
			target, cert.Name, cert.Partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			certSize,
			prometheus.GaugeValue,
			float64(cert.Size),
			target, cert.Name, cert.Partition,
		))

	}

	return m, true
}

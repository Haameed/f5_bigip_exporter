package probe

import (
	"log"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetClientSSLProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		sslHandshakeFailures = prometheus.NewDesc(
			"bigip_clientssl_handshake_failures_total",
			"Total number of failed SSL handshakes",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslSecureHandshakes = prometheus.NewDesc(
			"bigip_clientssl_secure_handshakes_total",
			"Total number of successful secure SSL handshakes",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslInsecureHandshakeAccepts = prometheus.NewDesc(
			"bigip_clientssl_insecure_handshake_accepts_total",
			"Total number of insecure handshakes that were accepted",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslInsecureHandshakeRejects = prometheus.NewDesc(
			"bigip_clientssl_insecure_handshake_rejects_total",
			"Total number of insecure handshakes that were rejected",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslFatalAlerts = prometheus.NewDesc(
			"bigip_clientssl_fatal_alerts_total",
			"Total number of fatal TLS alerts",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslCurrentActiveHandshakes = prometheus.NewDesc(
			"bigip_clientssl_current_active_handshakes",
			"Number of SSL handshakes currently in progress",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslCurNativeConns = prometheus.NewDesc(
			"bigip_clientssl_current_native_connections",
			"Current number of native SSL connections",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslMaxNativeConns = prometheus.NewDesc(
			"bigip_clientssl_max_native_connections",
			"Maximum number of native SSL connections",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslEncryptedBytesIn = prometheus.NewDesc(
			"bigip_clientssl_encrypted_bytes_in_total",
			"Total encrypted bytes received",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslEncryptedBytesOut = prometheus.NewDesc(
			"bigip_clientssl_encrypted_bytes_out_total",
			"Total encrypted bytes sent",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslDecryptedBytesIn = prometheus.NewDesc(
			"bigip_clientssl_decrypted_bytes_in_total",
			"Total decrypted bytes received",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslDecryptedBytesOut = prometheus.NewDesc(
			"bigip_clientssl_decrypted_bytes_out_total",
			"Total decrypted bytes sent",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslSessCacheHits = prometheus.NewDesc(
			"bigip_clientssl_session_cache_hits_total",
			"Total SSL session cache hits",
			[]string{"target", "profile", "partition"}, nil,
		)
		sslSessCacheLookups = prometheus.NewDesc(
			"bigip_clientssl_session_cache_lookups_total",
			"Total SSL session cache lookups",
			[]string{"target", "profile", "partition"}, nil,
		)
		// TLS/SSL protocol version usage — broken out by version via a label.
		sslProtocolUses = prometheus.NewDesc(
			"bigip_clientssl_protocol_uses_total",
			"Total handshakes per SSL/TLS protocol version",
			[]string{"target", "profile", "partition", "protocol"}, nil,
		)
	)

	type ClientSSLStatsEntry struct {
		HandshakeFailures        ValueBased  `json:"common.handshakeFailures"`
		SecureHandshakes         ValueBased  `json:"common.secureHandshakes"`
		InsecureHandshakeAccepts ValueBased  `json:"common.insecureHandshakeAccepts"`
		InsecureHandshakeRejects ValueBased  `json:"common.insecureHandshakeRejects"`
		FatalAlerts              ValueBased  `json:"common.fatalAlerts"`
		CurrentActiveHandshakes  ValueBased  `json:"common.currentActiveHandshakes"`
		CurNativeConns           ValueBased  `json:"common.curNativeConns"`
		MaxNativeConns           ValueBased  `json:"common.maxNativeConns"`
		EncryptedBytesIn         ValueBased  `json:"common.encryptedBytesIn"`
		EncryptedBytesOut        ValueBased  `json:"common.encryptedBytesOut"`
		DecryptedBytesIn         ValueBased  `json:"common.decryptedBytesIn"`
		DecryptedBytesOut        ValueBased  `json:"common.decryptedBytesOut"`
		SessCacheHits            ValueBased  `json:"common.sessCacheHits"`
		SessCacheLookups         ValueBased  `json:"common.sessCacheLookups"`
		ProtocolSSLv2            ValueBased  `json:"common.protocolUses.sslv2"`
		ProtocolSSLv3            ValueBased  `json:"common.protocolUses.sslv3"`
		ProtocolTLSv1            ValueBased  `json:"common.protocolUses.tlsv1"`
		ProtocolTLSv1_1          ValueBased  `json:"common.protocolUses.tlsv1_1"`
		ProtocolTLSv1_2          ValueBased  `json:"common.protocolUses.tlsv1_2"`
		ProtocolTLSv1_3          ValueBased  `json:"common.protocolUses.tlsv1_3"`
		TmName                   StringBased `json:"tmName"`
	}

	type ClientSSLNested struct {
		Entries ClientSSLStatsEntry `json:"entries"`
	}

	type ClientSSLStats struct {
		NestedStats ClientSSLNested `json:"nestedStats"`
	}

	type ClientSSLResponse struct {
		Kind     string                    `json:"kind"`
		SelfLink string                    `json:"selfLink"`
		Entries  map[string]ClientSSLStats `json:"entries"`
	}

	var m []prometheus.Metric

	var resp ClientSSLResponse
	if err := c.Get("/mgmt/tm/ltm/profile/client-ssl/stats", &resp); err != nil {
		log.Printf("Error getting client-ssl profile stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		stats := entry.NestedStats.Entries
		profile, partition := parsePoolName(stats.TmName.Description)

		m = append(m, prometheus.MustNewConstMetric(sslHandshakeFailures, prometheus.CounterValue, float64(stats.HandshakeFailures.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslSecureHandshakes, prometheus.CounterValue, float64(stats.SecureHandshakes.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslInsecureHandshakeAccepts, prometheus.CounterValue, float64(stats.InsecureHandshakeAccepts.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslInsecureHandshakeRejects, prometheus.CounterValue, float64(stats.InsecureHandshakeRejects.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslFatalAlerts, prometheus.CounterValue, float64(stats.FatalAlerts.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslCurrentActiveHandshakes, prometheus.GaugeValue, float64(stats.CurrentActiveHandshakes.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslCurNativeConns, prometheus.GaugeValue, float64(stats.CurNativeConns.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslMaxNativeConns, prometheus.GaugeValue, float64(stats.MaxNativeConns.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslEncryptedBytesIn, prometheus.CounterValue, float64(stats.EncryptedBytesIn.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslEncryptedBytesOut, prometheus.CounterValue, float64(stats.EncryptedBytesOut.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslDecryptedBytesIn, prometheus.CounterValue, float64(stats.DecryptedBytesIn.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslDecryptedBytesOut, prometheus.CounterValue, float64(stats.DecryptedBytesOut.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslSessCacheHits, prometheus.CounterValue, float64(stats.SessCacheHits.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslSessCacheLookups, prometheus.CounterValue, float64(stats.SessCacheLookups.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolSSLv2.Value), target, profile, partition, "SSLv2"))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolSSLv3.Value), target, profile, partition, "SSLv3"))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolTLSv1.Value), target, profile, partition, "TLSv1.0"))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolTLSv1_1.Value), target, profile, partition, "TLSv1.1"))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolTLSv1_2.Value), target, profile, partition, "TLSv1.2"))
		m = append(m, prometheus.MustNewConstMetric(sslProtocolUses, prometheus.CounterValue, float64(stats.ProtocolTLSv1_3.Value), target, profile, partition, "TLSv1.3"))
	}

	return m, true
}

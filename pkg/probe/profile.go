package probe

import (
	"log"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetHTTPProfileProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		httpResponses = prometheus.NewDesc(
			"bigip_http_responses_total",
			"Total HTTP responses by status class",
			[]string{"target", "profile", "partition", "code"}, nil,
		)
		httpRequestsByMethod = prometheus.NewDesc(
			"bigip_http_requests_by_method_total",
			"Total HTTP requests by method",
			[]string{"target", "profile", "partition", "method"}, nil,
		)
		httpRequestsByVersion = prometheus.NewDesc(
			"bigip_http_requests_by_version_total",
			"Total HTTP requests by protocol version",
			[]string{"target", "profile", "partition", "version"}, nil,
		)
		httpTotalRequests = prometheus.NewDesc(
			"bigip_http_profile_requests_total",
			"Total number of HTTP requests handled by the profile",
			[]string{"target", "profile", "partition"}, nil,
		)
	)

	type HTTPStatsEntry struct {
		GetReqs    ValueBased  `json:"getReqs"`
		PostReqs   ValueBased  `json:"postReqs"`
		NumberReqs ValueBased  `json:"numberReqs"`
		Resp2xx    ValueBased  `json:"resp_2xxCnt"`
		Resp3xx    ValueBased  `json:"resp_3xxCnt"`
		Resp4xx    ValueBased  `json:"resp_4xxCnt"`
		Resp5xx    ValueBased  `json:"resp_5xxCnt"`
		V9Reqs     ValueBased  `json:"v9Reqs"`
		V10Reqs    ValueBased  `json:"v10Reqs"`
		V11Reqs    ValueBased  `json:"v11Reqs"`
		V20Reqs    ValueBased  `json:"v20Reqs"`
		V30Reqs    ValueBased  `json:"v30Reqs"`
		TmName     StringBased `json:"tmName"`
	}
	type HTTPNested struct {
		Entries HTTPStatsEntry `json:"entries"`
	}
	type HTTPStats struct {
		NestedStats HTTPNested `json:"nestedStats"`
	}
	type HTTPResponse struct {
		Entries map[string]HTTPStats `json:"entries"`
	}

	var m []prometheus.Metric
	var resp HTTPResponse
	if err := c.Get("/mgmt/tm/ltm/profile/http/stats", &resp); err != nil {
		log.Printf("Error getting http profile stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		s := entry.NestedStats.Entries
		profile, partition := parsePoolName(s.TmName.Description)

		m = append(m, prometheus.MustNewConstMetric(httpTotalRequests, prometheus.CounterValue, float64(s.NumberReqs.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(httpResponses, prometheus.CounterValue, float64(s.Resp2xx.Value), target, profile, partition, "2xx"))
		m = append(m, prometheus.MustNewConstMetric(httpResponses, prometheus.CounterValue, float64(s.Resp3xx.Value), target, profile, partition, "3xx"))
		m = append(m, prometheus.MustNewConstMetric(httpResponses, prometheus.CounterValue, float64(s.Resp4xx.Value), target, profile, partition, "4xx"))
		m = append(m, prometheus.MustNewConstMetric(httpResponses, prometheus.CounterValue, float64(s.Resp5xx.Value), target, profile, partition, "5xx"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByMethod, prometheus.CounterValue, float64(s.GetReqs.Value), target, profile, partition, "GET"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByMethod, prometheus.CounterValue, float64(s.PostReqs.Value), target, profile, partition, "POST"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByVersion, prometheus.CounterValue, float64(s.V9Reqs.Value), target, profile, partition, "0.9"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByVersion, prometheus.CounterValue, float64(s.V10Reqs.Value), target, profile, partition, "1.0"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByVersion, prometheus.CounterValue, float64(s.V11Reqs.Value), target, profile, partition, "1.1"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByVersion, prometheus.CounterValue, float64(s.V20Reqs.Value), target, profile, partition, "2.0"))
		m = append(m, prometheus.MustNewConstMetric(httpRequestsByVersion, prometheus.CounterValue, float64(s.V30Reqs.Value), target, profile, partition, "3.0"))
	}

	return m, true
}

func GetTCPProfileProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		tcpOpenConns = prometheus.NewDesc(
			"bigip_tcp_open_connections",
			"Current open TCP connections for the profile",
			[]string{"target", "profile", "partition"}, nil,
		)
		tcpAccepts = prometheus.NewDesc(
			"bigip_tcp_accepts_total",
			"Total accepted TCP connections",
			[]string{"target", "profile", "partition"}, nil,
		)
		tcpConnfails = prometheus.NewDesc(
			"bigip_tcp_connection_failures_total",
			"Total TCP connection failures",
			[]string{"target", "profile", "partition"}, nil,
		)
		tcpRetransmits = prometheus.NewDesc(
			"bigip_tcp_retransmits_total",
			"Total TCP retransmits (network quality indicator)",
			[]string{"target", "profile", "partition"}, nil,
		)
		tcpRxBadSeg = prometheus.NewDesc(
			"bigip_tcp_received_bad_segments_total",
			"Total bad TCP segments received",
			[]string{"target", "profile", "partition"}, nil,
		)
		tcpRxRst = prometheus.NewDesc(
			"bigip_tcp_received_resets_total",
			"Total TCP RST packets received",
			[]string{"target", "profile", "partition"}, nil,
		)
	)

	type TCPStatsEntry struct {
		Accepts   ValueBased  `json:"accepts"`
		Connfails ValueBased  `json:"connfails"`
		Open      ValueBased  `json:"open"`
		Txrexmits ValueBased  `json:"txrexmits"`
		Rxbadseg  ValueBased  `json:"rxbadseg"`
		Rxrst     ValueBased  `json:"rxrst"`
		TmName    StringBased `json:"tmName"`
	}
	type TCPNested struct {
		Entries TCPStatsEntry `json:"entries"`
	}
	type TCPStats struct {
		NestedStats TCPNested `json:"nestedStats"`
	}
	type TCPResponse struct {
		Entries map[string]TCPStats `json:"entries"`
	}

	var m []prometheus.Metric
	var resp TCPResponse
	if err := c.Get("/mgmt/tm/ltm/profile/tcp/stats", &resp); err != nil {
		log.Printf("Error getting tcp profile stats: %v", err)
		return m, false
	}

	for _, entry := range resp.Entries {
		s := entry.NestedStats.Entries
		profile, partition := parsePoolName(s.TmName.Description)

		m = append(m, prometheus.MustNewConstMetric(tcpOpenConns, prometheus.GaugeValue, float64(s.Open.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(tcpAccepts, prometheus.CounterValue, float64(s.Accepts.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(tcpConnfails, prometheus.CounterValue, float64(s.Connfails.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(tcpRetransmits, prometheus.CounterValue, float64(s.Txrexmits.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(tcpRxBadSeg, prometheus.CounterValue, float64(s.Rxbadseg.Value), target, profile, partition))
		m = append(m, prometheus.MustNewConstMetric(tcpRxRst, prometheus.CounterValue, float64(s.Rxrst.Value), target, profile, partition))
	}

	return m, true
}
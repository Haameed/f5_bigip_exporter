package probe

import (
	"log"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetTrafficProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {

	type TrafficStatsNested struct {
		Entries map[string]ValueBased `json:"entries"`
	}

	type TrafficStats struct {
		NestedStats TrafficStatsNested `json:"nestedStats"`
	}

	type TrafficStatsResponse struct {
		Kind     string                  `json:"kind"`
		SelfLink string                  `json:"selfLink"`
		Entries  map[string]TrafficStats `json:"entries"`
	}

	var (
		ClientSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_client_side_received_bits_total",
			"Client side traffic bits in",
			[]string{"target"}, nil,
		)
		ClientSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_client_side_sent_bits_total",
			"Client side traffic bits out",
			[]string{"target"}, nil,
		)
		ClientSideTrafficCurConns = prometheus.NewDesc(
			"bigip_client_side_traffic_current_connections",
			"Current client side connections",
			[]string{"target"}, nil,
		)
		ClientSideTrafficEvictedConns = prometheus.NewDesc(
			"bigip_client_side_evicted_connections_total",
			"Evicted client side connections",
			[]string{"target"}, nil,
		)
		ClientSideTrafficMaxConns = prometheus.NewDesc(
			"bigip_client_side_traffic_max_connections",
			"Maximum client side connections",
			[]string{"target"}, nil,
		)
		ClientSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_client_side_received_packets_total",
			"Total packets received on client side",
			[]string{"target"}, nil,
		)
		ClientSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_client_side_sent_packets_total",
			"Total packets sent on client side",
			[]string{"target"}, nil,
		)
		ClientSideTrafficSlowKilled = prometheus.NewDesc(
			"bigip_client_side_slow_killed_connections_total",
			"Slowly killed client side connections",
			[]string{"target"}, nil,
		)
		ClientSideTrafficTotConns = prometheus.NewDesc(
			"bigip_client_side_connections_total",
			"Total client side connections",
			[]string{"target"}, nil,
		)
		ConnectionMemoryErrors = prometheus.NewDesc(
			"bigip_connection_memory_errors_total",
			"Connection memory errors",
			[]string{"target"}, nil,
		)
		DroppedPackets = prometheus.NewDesc(
			"bigip_dropped_packets_total",
			"Dropped packets",
			[]string{"target"}, nil,
		)
		FiveMinAvgClientSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_five_min_avg_client_side_traffic_bits_in",
			"5 minute average client side traffic bits in",
			[]string{"target"}, nil,
		)
		FiveMinAvgClientSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_five_min_avg_client_side_traffic_bits_out",
			"5 minute average client side traffic bits out",
			[]string{"target"}, nil,
		)
		FiveMinAvgClientSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_five_min_avg_client_side_traffic_packets_in",
			"5 minute average client side traffic packets in total",
			[]string{"target"}, nil,
		)
		FiveMinAvgClientSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_five_min_avg_client_side_traffic_packets_out",
			"5 minute average client side traffic packets out total",
			[]string{"target"}, nil,
		)
		FiveMinAvgClientSideTrafficTotConns = prometheus.NewDesc(
			"bigip_five_min_avg_client_side_traffic_connections",
			"5 minute average client side traffic total connections",
			[]string{"target"}, nil,
		)
		FiveMinAvgServerSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_five_min_avg_server_side_traffic_bits_in",
			"5 minute average server side traffic bits in",
			[]string{"target"}, nil,
		)
		FiveMinAvgServerSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_five_min_avg_server_side_traffic_bits_out",
			"5 minute average server side traffic bits out",
			[]string{"target"}, nil,
		)
		FiveMinAvgServerSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_five_min_avg_server_side_traffic_packets_in",
			"5 minute average server side traffic packets in total",
			[]string{"target"}, nil,
		)
		FiveMinAvgServerSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_five_min_avg_server_side_traffic_packets_out",
			"5 minute average server side traffic packets out total",
			[]string{"target"}, nil,
		)
		FiveMinAvgServerSideTrafficTotConns = prometheus.NewDesc(
			"bigip_five_min_avg_server_side_traffic_connections",
			"5 minute average server side traffic total connections",
			[]string{"target"}, nil,
		)
		FiveSecAvgClientSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_five_sec_avg_client_side_traffic_bits_in",
			"5 second average client side traffic bits in",
			[]string{"target"}, nil,
		)
		FiveSecAvgClientSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_five_sec_avg_client_side_traffic_bits_out",
			"5 second average client side traffic bits out",
			[]string{"target"}, nil,
		)
		FiveSecAvgClientSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_five_sec_avg_client_side_traffic_packets_in",
			"5 second average client side traffic packets in total",
			[]string{"target"}, nil,
		)
		FiveSecAvgClientSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_five_sec_avg_client_side_traffic_packets_out",
			"5 second average client side traffic packets out total",
			[]string{"target"}, nil,
		)
		FiveSecAvgClientSideTrafficTotConns = prometheus.NewDesc(
			"bigip_five_sec_avg_client_side_traffic_connections",
			"5 second average client side traffic total connections",
			[]string{"target"}, nil,
		)
		FiveSecAvgServerSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_five_sec_avg_server_side_traffic_bits_in",
			"5 second average server side traffic bits in",
			[]string{"target"}, nil,
		)
		FiveSecAvgServerSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_five_sec_avg_server_side_traffic_bits_out",
			"5 second average server side traffic bits out",
			[]string{"target"}, nil,
		)
		FiveSecAvgServerSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_five_sec_avg_server_side_traffic_packets_in",
			"5 second average server side traffic packets in total",
			[]string{"target"}, nil,
		)
		FiveSecAvgServerSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_five_sec_avg_server_side_traffic_packets_out",
			"5 second average server side traffic packets out total",
			[]string{"target"}, nil,
		)
		FiveSecAvgServerSideTrafficTotConns = prometheus.NewDesc(
			"bigip_five_sec_avg_server_side_traffic_connections",
			"5 second average server side traffic total connections",
			[]string{"target"}, nil,
		)
		HardwareSyncookiesDetected = prometheus.NewDesc(
			"bigip_hardware_syncookies_detected_total",
			"Hardware SYN cookies detected",
			[]string{"target"}, nil,
		)
		HardwareSyncookiesGenerated = prometheus.NewDesc(
			"bigip_hardware_syncookies_generated_total",
			"Hardware SYN cookies generated",
			[]string{"target"}, nil,
		)

		HttpRequests = prometheus.NewDesc(
			"bigip_http_requests_total",
			"HTTP requests",
			[]string{"target"}, nil,
		)
		IncomingPacketErrors = prometheus.NewDesc(
			"bigip_incoming_packet_errors_total",
			"Incoming packet errors",
			[]string{"target"}, nil,
		)
		LicenseDeny = prometheus.NewDesc(
			"bigip_license_deny_total",
			"License deny",
			[]string{"target"}, nil,
		)
		MaintenanceModeDeny = prometheus.NewDesc(
			"bigip_maintenance_mode_deny_total",
			"Maintenance mode deny",
			[]string{"target"}, nil,
		)
		MaxConnVirtualAddressDeny = prometheus.NewDesc(
			"bigip_max_conn_virtual_address_deny_total",
			"Maximum connections per virtual address deny",
			[]string{"target"}, nil,
		)
		MaxConnVirtualPathDeny = prometheus.NewDesc(
			"bigip_max_conn_virtual_path_deny_total",
			"Maximum connections per virtual path deny",
			[]string{"target"}, nil,
		)

		NoHandlerDeny = prometheus.NewDesc(
			"bigip_no_handler_deny_total",
			"No handler deny",
			[]string{"target"}, nil,
		)
		NoStagedHandlerDeny = prometheus.NewDesc(
			"bigip_no_staged_handler_deny_total",
			"No staged handler deny",
			[]string{"target"}, nil,
		)
		OneMinAvgClientSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_one_min_avg_client_side_traffic_bits_in",
			"1 minute average client side traffic bits in",
			[]string{"target"}, nil,
		)
		OneMinAvgClientSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_one_min_avg_client_side_traffic_bits_out",
			"1 minute average client side traffic bits out",
			[]string{"target"}, nil,
		)
		OneMinAvgClientSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_one_min_avg_client_side_traffic_packets_in",
			"1 minute average client side traffic packets in total",
			[]string{"target"}, nil,
		)
		OneMinAvgClientSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_one_min_avg_client_side_traffic_packets_out",
			"1 minute average client side traffic packets out total",
			[]string{"target"}, nil,
		)
		OneMinAvgClientSideTrafficTotConns = prometheus.NewDesc(
			"bigip_one_min_avg_client_side_traffic_connections",
			"1 minute average client side traffic total connections",
			[]string{"target"}, nil,
		)
		OneMinAvgServerSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_one_min_avg_server_side_traffic_bits_in",
			"1 minute average server side traffic bits in",
			[]string{"target"}, nil,
		)
		OneMinAvgServerSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_one_min_avg_server_side_traffic_bits_out",
			"1 minute average server side traffic bits out",
			[]string{"target"}, nil,
		)
		OneMinAvgServerSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_one_min_avg_server_side_traffic_packets_in",
			"1 minute average server side traffic packets in total",
			[]string{"target"}, nil,
		)
		OneMinAvgServerSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_one_min_avg_server_side_traffic_packets_out",
			"1 minute average server side traffic packets out total",
			[]string{"target"}, nil,
		)
		OneMinAvgServerSideTrafficTotConns = prometheus.NewDesc(
			"bigip_one_min_avg_server_side_traffic_connections",
			"1 minute average server side traffic total connections",
			[]string{"target"}, nil,
		)
		OutgoingPacketErrors = prometheus.NewDesc(
			"bigip_outgoing_packet_errors_total",
			"Outgoing packet errors",
			[]string{"target"}, nil,
		)
		ServerSideTrafficBitsIn = prometheus.NewDesc(
			"bigip_server_side_received_bits_total",
			"Server side traffic bits in",
			[]string{"target"}, nil,
		)
		ServerSideTrafficBitsOut = prometheus.NewDesc(
			"bigip_server_side_sent_bits_total",
			"Server side traffic bits out",
			[]string{"target"}, nil,
		)

		ServerSideTrafficCurConns = prometheus.NewDesc(
			"bigip_server_side_traffic_current_connections",
			"Current connections on server side",
			[]string{"target"}, nil,
		)
		ServerSideTrafficEvictedConns = prometheus.NewDesc(
			"bigip_server_side_traffic_evicted_connections_total",
			"Evicted connections on server side",
			[]string{"target"}, nil,
		)
		ServerSideTrafficMaxConns = prometheus.NewDesc(
			"bigip_server_side_traffic_max_connections",
			"Maximum connections on server side",
			[]string{"target"}, nil,
		)
		ServerSideTrafficPktsIn = prometheus.NewDesc(
			"bigip_server_side_received_packets_total",
			"Server side traffic packets in total",
			[]string{"target"}, nil,
		)
		ServerSideTrafficPktsOut = prometheus.NewDesc(
			"bigip_server_side_sent_packets_total",
			"Server side traffic packets out total",
			[]string{"target"}, nil,
		)
		ServerSideTrafficSlowKilled = prometheus.NewDesc(
			"bigip_server_side_traffic_slow_killed_connections_total",
			"Server side traffic slow killed",
			[]string{"target"}, nil,
		)
		ServerSideTrafficTotConns = prometheus.NewDesc(
			"bigip_server_side_connections_total",
			"Server side traffic total connections",
			[]string{"target"}, nil,
		)

		TmauthCurSessions = prometheus.NewDesc(
			"bigip_tmauth_current_sessions",
			"Current sessions in TMAuth",
			[]string{"target"}, nil,
		)

		TmauthFailureResults = prometheus.NewDesc(
			"bigip_tmauth_failure_results_total",
			"Failure results in TMAuth",
			[]string{"target"}, nil,
		)
		TmauthMaxSessions = prometheus.NewDesc(
			"bigip_tmauth_maximum_sessions",
			"Maximum sessions in TMAuth",
			[]string{"target"}, nil,
		)
		TmauthSuccessResults = prometheus.NewDesc(
			"bigip_tmauth_success_results_total",
			"Success results in TMAuth",
			[]string{"target"}, nil,
		)
		TmauthErrorResults = prometheus.NewDesc(
			"bigip_tmauth_error_results_total",
			"Error results in TMAuth",
			[]string{"target"}, nil,
		)
		TmauthWantcredentialResults = prometheus.NewDesc(
			"bigip_tmauth_wantcredential_results_total",
			"Want credential results in TMAuth",
			[]string{"target"}, nil,
		)
		TmauthTotSessions = prometheus.NewDesc(
			"bigip_tmauth_sessions_total",
			"Total sessions in TMAuth",
			[]string{"target"}, nil,
		)
		VirtualServerNonSynDeny = prometheus.NewDesc(
			"bigip_virtual_server_non_syn_deny_total",
			"Non-SYN packets denied on virtual server",
			[]string{"target"}, nil,
		)
	)

	var m []prometheus.Metric

	var resp TrafficStatsResponse
	if err := c.Get("/mgmt/tm/sys/traffic/stats", &resp); err != nil {
		log.Printf("Error getting traffic stats for %s: %v", target, err)
		return m, false
	}

	for _, item := range resp.Entries {
		stats := item.NestedStats.Entries
		getValue := func(key string) float64 {
			if v, ok := stats[key]; ok {
				return float64(v.Value)
			}
			return 0
		}
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficBitsIn,
			prometheus.CounterValue,
			getValue("clientSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficBitsOut,
			prometheus.CounterValue,
			getValue("clientSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficCurConns,
			prometheus.GaugeValue,
			getValue("clientSideTraffic.curConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficEvictedConns,
			prometheus.CounterValue,
			getValue("clientSideTraffic.evictedConns"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficMaxConns,
			prometheus.GaugeValue,
			getValue("clientSideTraffic.maxConns"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficPktsIn,
			prometheus.CounterValue,
			getValue("clientSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficPktsOut,
			prometheus.CounterValue,
			getValue("clientSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficSlowKilled,
			prometheus.CounterValue,
			getValue("clientSideTraffic.slowKilled"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ClientSideTrafficTotConns,
			prometheus.CounterValue,
			getValue("clientSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ConnectionMemoryErrors,
			prometheus.CounterValue,
			getValue("connectionMemoryErrors"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			DroppedPackets,
			prometheus.CounterValue,
			getValue("droppedPackets"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgClientSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("fiveMinAvgClientSideTraffic.bitsIn"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgClientSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("fiveMinAvgClientSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgClientSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("fiveMinAvgClientSideTraffic.pktsIn"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgClientSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("fiveMinAvgClientSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgClientSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("fiveMinAvgClientSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgServerSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("fiveMinAvgServerSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgServerSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("fiveMinAvgServerSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgServerSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("fiveMinAvgServerSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgServerSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("fiveMinAvgServerSideTraffic.pktsOut"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			FiveMinAvgServerSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("fiveMinAvgServerSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgClientSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("fiveSecAvgClientSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgClientSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("fiveSecAvgClientSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgClientSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("fiveSecAvgClientSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgClientSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("fiveSecAvgClientSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgClientSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("fiveSecAvgClientSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgServerSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("fiveSecAvgServerSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgServerSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("fiveSecAvgServerSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgServerSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("fiveSecAvgServerSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgServerSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("fiveSecAvgServerSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			FiveSecAvgServerSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("fiveSecAvgServerSideTraffic.totConns"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			HardwareSyncookiesDetected,
			prometheus.CounterValue,
			getValue("hardwareSyncookiesDetected"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			HardwareSyncookiesGenerated,
			prometheus.CounterValue,
			getValue("hardwareSyncookiesGenerated"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			HttpRequests,
			prometheus.CounterValue,
			getValue("httpRequests"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			IncomingPacketErrors,
			prometheus.CounterValue,
			getValue("incomingPacketErrors"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			LicenseDeny,
			prometheus.CounterValue,
			getValue("licenseDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			MaintenanceModeDeny,
			prometheus.CounterValue,
			getValue("maintenanceModeDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			MaxConnVirtualAddressDeny,
			prometheus.CounterValue,
			getValue("maxConnVirtualAddressDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			MaxConnVirtualPathDeny,
			prometheus.CounterValue,
			getValue("maxConnVirtualPathDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			NoHandlerDeny,
			prometheus.CounterValue,
			getValue("noHandlerDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			NoStagedHandlerDeny,
			prometheus.CounterValue,
			getValue("noStagedHandlerDeny"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgClientSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("oneMinAvgClientSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgClientSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("oneMinAvgClientSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgClientSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("oneMinAvgClientSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgClientSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("oneMinAvgClientSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgClientSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("oneMinAvgClientSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgServerSideTrafficBitsIn,
			prometheus.GaugeValue,
			getValue("oneMinAvgServerSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgServerSideTrafficBitsOut,
			prometheus.GaugeValue,
			getValue("oneMinAvgServerSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgServerSideTrafficPktsIn,
			prometheus.GaugeValue,
			getValue("oneMinAvgServerSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgServerSideTrafficPktsOut,
			prometheus.GaugeValue,
			getValue("oneMinAvgServerSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			OneMinAvgServerSideTrafficTotConns,
			prometheus.GaugeValue,
			getValue("oneMinAvgServerSideTraffic.totConns"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			OutgoingPacketErrors,
			prometheus.CounterValue,
			getValue("outgoingPacketErrors"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficBitsIn,
			prometheus.CounterValue,
			getValue("serverSideTraffic.bitsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficBitsOut,
			prometheus.CounterValue,
			getValue("serverSideTraffic.bitsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficCurConns,
			prometheus.GaugeValue,
			getValue("serverSideTraffic.curConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficEvictedConns,
			prometheus.CounterValue,
			getValue("serverSideTraffic.evictedConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficMaxConns,
			prometheus.GaugeValue,
			getValue("serverSideTraffic.maxConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficPktsIn,
			prometheus.CounterValue,
			getValue("serverSideTraffic.pktsIn"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficPktsOut,
			prometheus.CounterValue,
			getValue("serverSideTraffic.pktsOut"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficSlowKilled,
			prometheus.CounterValue,
			getValue("serverSideTraffic.slowKilled"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			ServerSideTrafficTotConns,
			prometheus.CounterValue,
			getValue("serverSideTraffic.totConns"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthCurSessions,
			prometheus.GaugeValue,
			getValue("tmauth.curSessions"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthErrorResults,
			prometheus.CounterValue,
			getValue("tmauth.errorResults"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthFailureResults,
			prometheus.CounterValue,
			getValue("tmauth.failureResults"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthMaxSessions,
			prometheus.GaugeValue,
			getValue("tmauth.maxSessions"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthSuccessResults,
			prometheus.CounterValue,
			getValue("tmauth.successResults"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthTotSessions,
			prometheus.CounterValue,
			getValue("tmauth.totSessions"),
			target,
		))
		m = append(m, prometheus.MustNewConstMetric(
			TmauthWantcredentialResults,
			prometheus.CounterValue,
			getValue("tmauth.wantcredentialResults"),
			target,
		))

		m = append(m, prometheus.MustNewConstMetric(
			VirtualServerNonSynDeny,
			prometheus.CounterValue,
			getValue("virtualServerNonSynDeny"),
			target,
		))

	}
	return m, true
}

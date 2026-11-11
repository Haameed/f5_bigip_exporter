package probe

import (
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetVirtualServersProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		clientSideBitsIn = prometheus.NewDesc(
			"bigip_virtual_server_client_side_received_bits_total",
			"Total number of bits received on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSideBitsOut = prometheus.NewDesc(
			"bigip_virtual_server_client_side_sent_bits_total",
			"Total number of bits sent on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSideCurConns = prometheus.NewDesc(
			"bigip_virtual_server_client_side_current_connections",
			"Number of current connections on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSideEvictedConns = prometheus.NewDesc(
			"bigip_virtual_server_client_side_evicted_connections_total",
			"Total number of evicted connections on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSideMaxConns = prometheus.NewDesc(
			"bigip_virtual_server_client_side_max_connections",
			"Maximum number of connections on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSidePktsIn = prometheus.NewDesc(
			"bigip_virtual_server_client_side_received_packets_total",
			"Total number of packets received on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSidePktsOut = prometheus.NewDesc(
			"bigip_virtual_server_client_side_sent_packets_total",
			"Total number of packets sent to the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		clientSideSlowKilled = prometheus.NewDesc(
			"bigip_virtual_server_client_side_slow_killed_connections_total",
			"Total number of slow killed connections on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)

		clientSideTotConns = prometheus.NewDesc(
			"bigip_virtual_server_client_side_connections_total",
			"Total number of connections on the client side",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		cmpEnableMode = prometheus.NewDesc(
			"bigip_virtual_server_cmp_enable_mode_info",
			"CMP (Clustered Multiprocessing) enable mode (constant 1; mode exposed via label)",
			[]string{"target", "vs_name", "partition", "cmp_enable_mode"}, nil,
		)
		cmpEnabled = prometheus.NewDesc(
			"bigip_virtual_server_cmp_enabled_info",
			"Whether CMP (Clustered Multiprocessing) is enabled (constant 1; state exposed via label)",
			[]string{"target", "vs_name", "partition", "cmp_enabled"}, nil,
		)
		csMaxConnDur = prometheus.NewDesc(
			"bigip_virtual_server_client_side_connection_duration_max_seconds",
			"Maximum connection duration on the client side, in seconds",
			[]string{"target", "vs_name", "partition"}, nil,
		)

		csMeanConnDur = prometheus.NewDesc(
			"bigip_virtual_server_client_side_connection_duration_mean_seconds",
			"Mean connection duration on the client side, in seconds",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		csMinConnDur = prometheus.NewDesc(
			"bigip_virtual_server_client_side_connection_duration_min_seconds",
			"Minimum connection duration on the client side, in seconds",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralBitsIn = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_received_bits_total",
			"Total number of bits received on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralBitsOut = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_sent_bits_total",
			"Total number of bits sent on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralCurConns = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_current_connections",
			"Number of current connections on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralEvictedConns = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_evicted_connections_total",
			"Total number of evicted connections on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralMaxConns = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_max_connections",
			"Maximum number of connections on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralPktsIn = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_received_packets_total",
			"Total number of packets received on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralPktsOut = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_sent_packets_total",
			"Total number of packets sent on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralSlowKilled = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_slow_killed_connections_total",
			"Total number of slow killed connections on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		ephemeralTotConns = prometheus.NewDesc(
			"bigip_virtual_server_ephemeral_connections_total",
			"Total number of connections on the ephemeral (FastL4) node",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		fiveMinAvgUsageRatio = prometheus.NewDesc(
			"bigip_virtual_server_five_min_avg_usage_ratio",
			"Five minute average usage ratio",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		fiveSecAvgUsageRatio = prometheus.NewDesc(
			"bigip_virtual_server_five_sec_avg_usage_ratio",
			"Five second average usage ratio",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrMsgIn = prometheus.NewDesc(
			"bigip_virtual_server_mr_messages_in_total",
			"Total number of message-router messages received",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrMsgOut = prometheus.NewDesc(
			"bigip_virtual_server_mr_messages_out_total",
			"Total number of message-router messages sent",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrReqIn = prometheus.NewDesc(
			"bigip_virtual_server_mr_requests_in_total",
			"Total number of message-router requests received",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrReqOut = prometheus.NewDesc(
			"bigip_virtual_server_mr_requests_out_total",
			"Total number of message-router requests sent",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrRespIn = prometheus.NewDesc(
			"bigip_virtual_server_mr_responses_in_total",
			"Total number of message-router responses received",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		mrRespOut = prometheus.NewDesc(
			"bigip_virtual_server_mr_responses_out_total",
			"Total number of message-router responses sent",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		oneMinAvgUsageRatio = prometheus.NewDesc(
			"bigip_virtual_server_one_min_avg_usage_ratio",
			"One minute average usage ratio",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		statusAvailabilityState = prometheus.NewDesc(
			"bigip_virtual_server_availability_state_info",
			"Availability state of the virtual server (constant 1; state exposed via label)",
			[]string{"target", "vs_name", "partition", "availability_state"}, nil,
		)
		statusEnabledState = prometheus.NewDesc(
			"bigip_virtual_server_enabled_state_info",
			"Enabled state of the virtual server (constant 1; state exposed via label)",
			[]string{"target", "vs_name", "partition", "enabled_state"}, nil,
		)
		syncookieStatus = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_status_info",
			"Syncookie status of the virtual server (constant 1; status exposed via label)",
			[]string{"target", "vs_name", "partition", "syncookie_status"}, nil,
		)
		syncookieAccepts = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_accepts_total",
			"Total number of syncookie accepts",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieHwAccepts = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_hw_accepts_total",
			"Total number of syncookie hardware accepts",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieHwSyncookies = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_hw_syncookies_total",
			"Total number of syncookie hardware syncookies",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieHwsyncookieInstance = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_hw_syncookie_instances",
			"Current number of hardware syncookie instances",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieRejects = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_rejects_total",
			"Total number of syncookie rejects",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieSwsyncookieInstance = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_sw_syncookie_instances",
			"Current number of software syncookie instances",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieSyncacheCurr = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_syncache_current_entries",
			"Current number of syncache entries",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieSyncacheOver = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_syncache_overflows_total",
			"Total number of syncache overflows",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		syncookieSyncookies = prometheus.NewDesc(
			"bigip_virtual_server_syncookie_syncookies_total",
			"Total number of syncookies sent",
			[]string{"target", "vs_name", "partition"}, nil,
		)
		totRequests = prometheus.NewDesc(
			"bigip_virtual_server_requests_total",
			"Total number of requests",
			[]string{"target", "vs_name", "partition"}, nil,
		)
	)

	type SourceAddressTranslation struct {
		Type string `json:"type"`
	}
	type VlansReference struct {
		Link string `json:"link"`
	}

	type PoliciesReference struct {
		Link            string `json:"link"`
		IsSubcollection bool   `json:"isSubcollection"`
	}
	type ProfilesReference struct {
		Link            string `json:"link"`
		IsSubcollection bool   `json:"isSubcollection"`
	}

	type RulesReference struct {
		Link string `json:"link"`
	}

	type VirtaulServer struct {
		Kind                       string                   `json:"kind"`
		Name                       string                   `json:"name"`
		Partition                  string                   `json:"partition"`
		FullPath                   string                   `json:"fullPath"`
		Generation                 int                      `json:"generation"`
		SelfLink                   string                   `json:"selfLink"`
		AddressStatus              string                   `json:"addressStatus"`
		AutoLasthop                string                   `json:"autoLasthop"`
		CmpEnabled                 string                   `json:"cmpEnabled"`
		ConnectionLimit            int                      `json:"connectionLimit"`
		CreationTime               time.Time                `json:"creationTime"`
		Destination                string                   `json:"destination"`
		Enabled                    bool                     `json:"enabled"`
		EvictionProtected          string                   `json:"evictionProtected"`
		GtmScore                   int                      `json:"gtmScore"`
		IPProtocol                 string                   `json:"ipProtocol"`
		LastModifiedTime           time.Time                `json:"lastModifiedTime"`
		Mask                       string                   `json:"mask"`
		Mirror                     string                   `json:"mirror"`
		MobileAppTunnel            string                   `json:"mobileAppTunnel"`
		Nat64                      string                   `json:"nat64"`
		RateLimit                  string                   `json:"rateLimit"`
		RateLimitDstMask           int                      `json:"rateLimitDstMask"`
		RateLimitMode              string                   `json:"rateLimitMode"`
		RateLimitSrcMask           int                      `json:"rateLimitSrcMask"`
		ServersslUseSni            string                   `json:"serversslUseSni"`
		ServiceDownImmediateAction string                   `json:"serviceDownImmediateAction"`
		Source                     string                   `json:"source"`
		SourceAddressTranslation   SourceAddressTranslation `json:"sourceAddressTranslation"`
		SourcePort                 string                   `json:"sourcePort"`
		SynCookieStatus            string                   `json:"synCookieStatus"`
		TranslateAddress           string                   `json:"translateAddress"`
		TranslatePort              string                   `json:"translatePort"`
		VlansEnabled               bool                     `json:"vlansEnabled,omitempty"`
		VsIndex                    int                      `json:"vsIndex"`
		Vlans                      []string                 `json:"vlans,omitempty"`
		VlansReference             []VlansReference         `json:"vlansReference,omitempty"`
		PoliciesReference          PoliciesReference        `json:"policiesReference"`
		ProfilesReference          ProfilesReference        `json:"profilesReference"`
		Description                string                   `json:"description,omitempty"`
		VlansDisabled              bool                     `json:"vlansDisabled,omitempty"`
		Rules                      []string                 `json:"rules,omitempty"`
		RulesReference             []RulesReference         `json:"rulesReference,omitempty"`
	}

	type StatEntry struct {
		ClientSideBitsIn             ValueBased  `json:"clientside.bitsIn"`
		ClientSideBitsOut            ValueBased  `json:"clientside.bitsOut"`
		ClientSideCurConns           ValueBased  `json:"clientside.curConns"`
		ClientSideEvictedConns       ValueBased  `json:"clientside.evictedConns"`
		ClientSideMaxConns           ValueBased  `json:"clientside.maxConns"`
		ClientSidePktsIn             ValueBased  `json:"clientside.pktsIn"`
		ClientSidePktsOut            ValueBased  `json:"clientside.pktsOut"`
		ClientSideSlowKilled         ValueBased  `json:"clientside.slowKilled"`
		ClientSideTotConns           ValueBased  `json:"clientside.totConns"`
		CmpEnableMode                StringBased `json:"cmpEnableMode"`
		CmpEnabled                   StringBased `json:"cmpEnabled"`
		CsMaxConnDur                 ValueBased  `json:"csMaxConnDur"`
		CsMeanConnDur                ValueBased  `json:"csMeanConnDur"`
		CsMinConnDur                 ValueBased  `json:"csMinConnDur"`
		Destination                  ValueBased  `json:"destination"`
		EphemeralBitsIn              ValueBased  `json:"ephemeral.bitsIn"`
		EphemeralBitsOut             ValueBased  `json:"ephemeral.bitsOut"`
		EphemeralCurConns            ValueBased  `json:"ephemeral.curConns"`
		EphemeralEvictedConns        ValueBased  `json:"ephemeral.evictedConns"`
		EphemeralMaxConns            ValueBased  `json:"ephemeral.maxConns"`
		EphemeralPktsIn              ValueBased  `json:"ephemeral.pktsIn"`
		EphemeralPktsOut             ValueBased  `json:"ephemeral.pktsOut"`
		EphemeralSlowKilled          ValueBased  `json:"ephemeral.slowKilled"`
		EphemeralTotConns            ValueBased  `json:"ephemeral.totConns"`
		FiveMinAvgUsageRatio         ValueBased  `json:"fiveMinAvgUsageRatio"`
		FiveSecAvgUsageRatio         ValueBased  `json:"fiveSecAvgUsageRatio"`
		MrMsgIn                      ValueBased  `json:"mr.msgIn"`
		MrMsgOut                     ValueBased  `json:"mr.msgOut"`
		MrReqIn                      ValueBased  `json:"mr.reqIn"`
		MrReqOut                     ValueBased  `json:"mr.reqOut"`
		MrRespIn                     ValueBased  `json:"mr.respIn"`
		MrRespOut                    ValueBased  `json:"mr.respOut"`
		TmName                       ValueBased  `json:"tmName"`
		OneMinAvgUsageRatio          ValueBased  `json:"oneMinAvgUsageRatio"`
		StatusAvailabilityState      StringBased `json:"status.availabilityState"`
		StatusEnabledState           StringBased `json:"status.enabledState"`
		StatusStatusReason           StringBased `json:"status.statusReason"`
		SyncookieStatus              StringBased `json:"syncookieStatus"`
		SyncookieAccepts             ValueBased  `json:"syncookie.accepts"`
		SyncookieHwAccepts           ValueBased  `json:"syncookie.hwAccepts"`
		SyncookieHwSyncookies        ValueBased  `json:"syncookie.hwSyncookies"`
		SyncookieHwsyncookieInstance ValueBased  `json:"syncookie.hwsyncookieInstance"`
		SyncookieRejects             ValueBased  `json:"syncookie.rejects"`
		SyncookieSwsyncookieInstance ValueBased  `json:"syncookie.swsyncookieInstance"`
		SyncookieSyncacheCurr        ValueBased  `json:"syncookie.syncacheCurr"`
		SyncookieSyncacheOver        ValueBased  `json:"syncookie.syncacheOver"`
		SyncookieSyncookies          ValueBased  `json:"syncookie.syncookies"`
		TotRequests                  ValueBased  `json:"totRequests"`
	}
	type NestedStats struct {
		Kind     string    `json:"kind"`
		SelfLink string    `json:"selfLink"`
		Entries  StatEntry `json:"entries"`
	}

	type VirtualServerstatsEntry struct {
		NestedStats NestedStats `json:"nestedStats"`
	}

	type VirtualServerStats struct {
		Kind       string                             `json:"kind"`
		Generation int                                `json:"generation"`
		SelfLink   string                             `json:"selfLink"`
		Entries    map[string]VirtualServerstatsEntry `json:"entries"`
	}
	type VirtualServers struct {
		Kind     string          `json:"kind"`
		SelfLink string          `json:"selfLink"`
		Items    []VirtaulServer `json:"items"`
	}

	type vsInfo struct {
		vs      VirtaulServer
		vsStats VirtualServerStats
	}
	var virtualServerResponse VirtualServers
	if err := c.Get("/mgmt/tm/ltm/virtual", &virtualServerResponse); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	var m []prometheus.Metric
	var wg sync.WaitGroup
	results := make(chan vsInfo, len(virtualServerResponse.Items))

	for _, vs := range virtualServerResponse.Items {
		wg.Add(1)
		go func(vs VirtaulServer) {
			defer wg.Done()
			url, _ := url.Parse(vs.SelfLink)
			path := url.Path
			var virtualServerStats VirtualServerStats
			c.Get(path+"/stats", &virtualServerStats)
			results <- vsInfo{
				vs:      vs,
				vsStats: virtualServerStats,
			}

		}(vs)

	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for item := range results {
		for _, statEntry := range item.vsStats.Entries {
			m = append(m, prometheus.MustNewConstMetric(clientSideBitsIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSideBitsIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideBitsOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSideBitsOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideCurConns, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.ClientSideCurConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideEvictedConns, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSideEvictedConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideMaxConns, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.ClientSideMaxConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSidePktsIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSidePktsIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSidePktsOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSidePktsOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideSlowKilled, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSideSlowKilled.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(clientSideTotConns, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.ClientSideTotConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(cmpEnableMode, prometheus.GaugeValue, 1.0, target, item.vs.Name, item.vs.Partition, statEntry.NestedStats.Entries.CmpEnableMode.Description))
			m = append(m, prometheus.MustNewConstMetric(cmpEnabled, prometheus.GaugeValue, 1.0, target, item.vs.Name, item.vs.Partition, statEntry.NestedStats.Entries.CmpEnabled.Description))
			m = append(m, prometheus.MustNewConstMetric(csMaxConnDur, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.CsMaxConnDur.Value)/1000, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(csMeanConnDur, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.CsMeanConnDur.Value)/1000, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(csMinConnDur, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.CsMinConnDur.Value)/1000, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralBitsIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralBitsIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralBitsOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralBitsOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralCurConns, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.EphemeralCurConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralEvictedConns, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralEvictedConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralMaxConns, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.EphemeralMaxConns.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralPktsIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralPktsIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralPktsOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralPktsOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralSlowKilled, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralSlowKilled.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(ephemeralTotConns, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.EphemeralTotConns.Value), target, item.vs.Name, item.vs.Partition))
			// The iControl REST API reports these "usage ratio" fields on a
			// 0-100 scale (they are actually percentages); divide by 100 to
			// expose true 0-1 ratios, per Prometheus naming conventions.
			m = append(m, prometheus.MustNewConstMetric(fiveMinAvgUsageRatio, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.FiveMinAvgUsageRatio.Value)/100, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(fiveSecAvgUsageRatio, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.FiveSecAvgUsageRatio.Value)/100, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrMsgIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrMsgIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrMsgOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrMsgOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrReqIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrReqIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrReqOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrReqOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrRespIn, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrRespIn.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(mrRespOut, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.MrRespOut.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(oneMinAvgUsageRatio, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.OneMinAvgUsageRatio.Value)/100, target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(statusAvailabilityState, prometheus.GaugeValue, 1.0, target, item.vs.Name, item.vs.Partition, statEntry.NestedStats.Entries.StatusAvailabilityState.Description))
			m = append(m, prometheus.MustNewConstMetric(statusEnabledState, prometheus.GaugeValue, 1.0, target, item.vs.Name, item.vs.Partition, statEntry.NestedStats.Entries.StatusEnabledState.Description))
			m = append(m, prometheus.MustNewConstMetric(syncookieStatus, prometheus.GaugeValue, 1.0, target, item.vs.Name, item.vs.Partition, statEntry.NestedStats.Entries.SyncookieStatus.Description))
			m = append(m, prometheus.MustNewConstMetric(syncookieAccepts, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieAccepts.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieHwAccepts, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieHwAccepts.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieHwSyncookies, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieHwSyncookies.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieHwsyncookieInstance, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.SyncookieHwsyncookieInstance.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieRejects, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieRejects.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieSwsyncookieInstance, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.SyncookieSwsyncookieInstance.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieSyncacheCurr, prometheus.GaugeValue, float64(statEntry.NestedStats.Entries.SyncookieSyncacheCurr.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieSyncacheOver, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieSyncacheOver.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(syncookieSyncookies, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.SyncookieSyncookies.Value), target, item.vs.Name, item.vs.Partition))
			m = append(m, prometheus.MustNewConstMetric(totRequests, prometheus.CounterValue, float64(statEntry.NestedStats.Entries.TotRequests.Value), target, item.vs.Name, item.vs.Partition))
		}
	}
	return m, true
}

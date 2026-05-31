package probe

import (
	"log"
	"strings"
	"sync"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetPoolProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		poolActiveMembers = prometheus.NewDesc(
			"bigip_pool_active_members",
			"Number of active pool members",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolAvailableMembers = prometheus.NewDesc(
			"bigip_pool_available_members",
			"Number of available pool members",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolTotalMembers = prometheus.NewDesc(
			"bigip_pool_total_members",
			"Total number of pool members",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolCurrentSessions = prometheus.NewDesc(
			"bigip_pool_current_sessions",
			"Current sessions for the pool",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersideBitsIn = prometheus.NewDesc(
			"bigip_pool_serverside_bits_in_total",
			"Total bits received on server side",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersideBitsOut = prometheus.NewDesc(
			"bigip_pool_serverside_bits_out_total",
			"Total bits sent on server side",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersideCurrentConns = prometheus.NewDesc(
			"bigip_pool_serverside_current_connections",
			"Current server side connections",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersideMaxConns = prometheus.NewDesc(
			"bigip_pool_serverside_max_connections",
			"Maximum server side connections",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersideTotalConns = prometheus.NewDesc(
			"bigip_pool_serverside_total_connections",
			"Total server side connections",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersidePktsIn = prometheus.NewDesc(
			"bigip_pool_serverside_packets_in_total",
			"Total packets received on server side",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolServersidePktsOut = prometheus.NewDesc(
			"bigip_pool_serverside_packets_out_total",
			"Total packets sent on server side",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolTotalRequests = prometheus.NewDesc(
			"bigip_pool_total_requests",
			"Total requests to the pool",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolAvailabilityState = prometheus.NewDesc(
			"bigip_pool_availability_state",
			"Pool availability state (0=offline, 1=unknown, 2=available)",
			[]string{"target", "pool", "partition", "status_reason"}, nil,
		)
		poolEnabledState = prometheus.NewDesc(
			"bigip_pool_enabled_state",
			"Pool enabled state (0=disabled, 1=enabled)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMinActiveMembers = prometheus.NewDesc(
			"bigip_pool_min_active_members",
			"Minimum active members configured",
			[]string{"target", "pool", "partition"}, nil,
		)

		poolConnQueueDepth = prometheus.NewDesc(
			"bigip_pool_connq_depth",
			"Current connection queue depth",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueServiced = prometheus.NewDesc(
			"bigip_pool_connq_serviced_total",
			"Total connections serviced from queue",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAgeMax = prometheus.NewDesc(
			"bigip_pool_connq_age_max_milliseconds",
			"Maximum time a connection spent in queue",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAgeHead = prometheus.NewDesc(
			"bigip_pool_connq_age_head_milliseconds",
			"Age of the oldest connection in queue",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAgeEma = prometheus.NewDesc(
			"bigip_pool_connq_age_ema_milliseconds",
			"Exponential moving average of connection queue age",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAgeEdm = prometheus.NewDesc(
			"bigip_pool_connq_age_edm_milliseconds",
			"Exponential decay maximum of connection queue age",
			[]string{"target", "pool", "partition"}, nil,
		)

		poolConnQueueAllDepth = prometheus.NewDesc(
			"bigip_pool_connq_all_depth",
			"Current connection queue depth (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAllServiced = prometheus.NewDesc(
			"bigip_pool_connq_all_serviced_total",
			"Total connections serviced from queue (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAllAgeMax = prometheus.NewDesc(
			"bigip_pool_connq_all_age_max_milliseconds",
			"Maximum time a connection spent in queue (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAllAgeHead = prometheus.NewDesc(
			"bigip_pool_connq_all_age_head_milliseconds",
			"Age of the oldest connection in queue (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAllAgeEma = prometheus.NewDesc(
			"bigip_pool_connq_all_age_ema_milliseconds",
			"Exponential moving average of connection queue age (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolConnQueueAllAgeEdm = prometheus.NewDesc(
			"bigip_pool_connq_all_age_edm_milliseconds",
			"Exponential decay maximum of connection queue age (all virtual servers)",
			[]string{"target", "pool", "partition"}, nil,
		)

		poolCurrentPriorityGroup = prometheus.NewDesc(
			"bigip_pool_current_priority_group",
			"Current priority group in use",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolHighestPriorityGroup = prometheus.NewDesc(
			"bigip_pool_highest_priority_group",
			"Highest priority group number",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolLowestPriorityGroup = prometheus.NewDesc(
			"bigip_pool_lowest_priority_group",
			"Lowest priority group number",
			[]string{"target", "pool", "partition"}, nil,
		)

		poolMrMsgIn = prometheus.NewDesc(
			"bigip_pool_mr_messages_in_total",
			"Total messages received by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMrMsgOut = prometheus.NewDesc(
			"bigip_pool_mr_messages_out_total",
			"Total messages sent by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMrReqIn = prometheus.NewDesc(
			"bigip_pool_mr_requests_in_total",
			"Total requests received by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMrReqOut = prometheus.NewDesc(
			"bigip_pool_mr_requests_out_total",
			"Total requests sent by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMrRespIn = prometheus.NewDesc(
			"bigip_pool_mr_responses_in_total",
			"Total responses received by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMrRespOut = prometheus.NewDesc(
			"bigip_pool_mr_responses_out_total",
			"Total responses sent by message router",
			[]string{"target", "pool", "partition"}, nil,
		)
		poolMemberStatus = prometheus.NewDesc(
			"bigip_pool_member_state",
			"State of the pool member (0=down, 1=up)",
			[]string{"target", "pool", "partition", "member", "session"}, nil,
		)
	)

	type PoolStatsEntry struct {
		ActiveMemberCnt         ValueBased  `json:"activeMemberCnt"`
		AvailableMemberCnt      ValueBased  `json:"availableMemberCnt"`
		MemberCnt               ValueBased  `json:"memberCnt"`
		CurSessions             ValueBased  `json:"curSessions"`
		MinActiveMembers        ValueBased  `json:"minActiveMembers"`
		TmName                  StringBased `json:"tmName"`
		MonitorRule             StringBased `json:"monitorRule"`
		ConnQDepth              ValueBased  `json:"connq.depth"`
		ConnQServiced           ValueBased  `json:"connq.serviced"`
		ConnQAgeMax             ValueBased  `json:"connq.ageMax"`
		ConnQAgeHead            ValueBased  `json:"connq.ageHead"`
		ConnQAgeEma             ValueBased  `json:"connq.ageEma"`
		ConnQAgeEdm             ValueBased  `json:"connq.ageEdm"`
		ConnQAllDepth           ValueBased  `json:"connqAll.depth"`
		ConnQAllServiced        ValueBased  `json:"connqAll.serviced"`
		ConnQAllAgeMax          ValueBased  `json:"connqAll.ageMax"`
		ConnQAllAgeHead         ValueBased  `json:"connqAll.ageHead"`
		ConnQAllAgeEma          ValueBased  `json:"connqAll.ageEma"`
		ConnQAllAgeEdm          ValueBased  `json:"connqAll.ageEdm"`
		CurPriogrp              ValueBased  `json:"curPriogrp"`
		HighestPriogrp          ValueBased  `json:"highestPriogrp"`
		LowestPriogrp           ValueBased  `json:"lowestPriogrp"`
		MrMsgIn                 ValueBased  `json:"mr.msgIn"`
		MrMsgOut                ValueBased  `json:"mr.msgOut"`
		MrReqIn                 ValueBased  `json:"mr.reqIn"`
		MrReqOut                ValueBased  `json:"mr.reqOut"`
		MrRespIn                ValueBased  `json:"mr.respIn"`
		MrRespOut               ValueBased  `json:"mr.respOut"`
		ServersideBitsIn        ValueBased  `json:"serverside.bitsIn"`
		ServersideBitsOut       ValueBased  `json:"serverside.bitsOut"`
		ServersideCurConns      ValueBased  `json:"serverside.curConns"`
		ServersideMaxConns      ValueBased  `json:"serverside.maxConns"`
		ServersidePktsIn        ValueBased  `json:"serverside.pktsIn"`
		ServersidePktsOut       ValueBased  `json:"serverside.pktsOut"`
		ServersideTotConns      ValueBased  `json:"serverside.totConns"`
		StatusAvailabilityState StringBased `json:"status.availabilityState"`
		StatusEnabledState      StringBased `json:"status.enabledState"`
		StatusReason            StringBased `json:"status.statusReason"`
		TotRequests             ValueBased  `json:"totRequests"`
	}

	type PoolStatsNested struct {
		Entries PoolStatsEntry `json:"entries"`
	}

	type PoolStats struct {
		NestedStats PoolStatsNested `json:"nestedStats"`
	}

	type PoolStatsResponse struct {
		Kind     string               `json:"kind"`
		SelfLink string               `json:"selfLink"`
		Entries  map[string]PoolStats `json:"entries"`
	}

	type PoolMemberStats struct {
		Kind            string `json:"kind"`
		Name            string `json:"name"`
		Partition       string `json:"partition"`
		FullPath        string `json:"fullPath"`
		Generation      int    `json:"generation"`
		SelfLink        string `json:"selfLink"`
		Address         string `json:"address"`
		ConnectionLimit int    `json:"connectionLimit"`
		DynamicRatio    int    `json:"dynamicRatio"`
		Ephemeral       string `json:"ephemeral"`
		FQDN            struct {
			AutoPopulate string `json:"autopopulate"`
		} `json:"fqdn"`
		InheritProfile string `json:"inheritProfile"`
		Logging        string `json:"logging"`
		Monitor        string `json:"monitor"`
		PriorityGroup  int    `json:"priorityGroup"`
		RateLimit      string `json:"rateLimit"`
		Ratio          int    `json:"ratio"`
		Session        string `json:"session"`
		State          string `json:"state"`
	}
	type PoolMemberResponse struct {
		Kind     string            `json:"kind"`
		SelfLink string            `json:"selfLink"`
		Items    []PoolMemberStats `json:"items"`
	}

	type PoolMemberDetails struct {
		PoolMemberResponse
		Partition string
		PoolName  string
	}

	var m []prometheus.Metric

	var poolStatsResp PoolStatsResponse
	if err := c.Get("/mgmt/tm/ltm/pool/stats", &poolStatsResp); err != nil {
		log.Printf("Error getting pool stats: %v", err)
		return m, false
	}
	var wg sync.WaitGroup

	memberCh := make(chan PoolMemberDetails, len(poolStatsResp.Entries))
	for _, entry := range poolStatsResp.Entries {

		stats := entry.NestedStats.Entries

		// Extract pool name and partition from tmName (format: /partition/poolname)
		poolName, partition := parsePoolName(stats.TmName.Description)
		m = append(m, prometheus.MustNewConstMetric(
			poolActiveMembers,
			prometheus.GaugeValue,
			float64(stats.ActiveMemberCnt.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolAvailableMembers,
			prometheus.GaugeValue,
			float64(stats.AvailableMemberCnt.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolTotalMembers,
			prometheus.GaugeValue,
			float64(stats.MemberCnt.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMinActiveMembers,
			prometheus.GaugeValue,
			float64(stats.MinActiveMembers.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolCurrentSessions,
			prometheus.GaugeValue,
			float64(stats.CurSessions.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolServersideBitsIn,
			prometheus.CounterValue,
			float64(stats.ServersideBitsIn.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersideBitsOut,
			prometheus.CounterValue,
			float64(stats.ServersideBitsOut.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersideCurrentConns,
			prometheus.GaugeValue,
			float64(stats.ServersideCurConns.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersideMaxConns,
			prometheus.GaugeValue,
			float64(stats.ServersideMaxConns.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersideTotalConns,
			prometheus.CounterValue,
			float64(stats.ServersideTotConns.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersidePktsIn,
			prometheus.CounterValue,
			float64(stats.ServersidePktsIn.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolServersidePktsOut,
			prometheus.CounterValue,
			float64(stats.ServersidePktsOut.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolTotalRequests,
			prometheus.CounterValue,
			float64(stats.TotRequests.Value),
			target, poolName, partition,
		))

		availState := poolAvailabilityStateToValue(stats.StatusAvailabilityState.Description)
		m = append(m, prometheus.MustNewConstMetric(
			poolAvailabilityState,
			prometheus.GaugeValue,
			float64(availState),
			target, poolName, partition, stats.StatusReason.Description,
		))
		enabledState := poolEnabledStateToValue(stats.StatusEnabledState.Description)
		m = append(m, prometheus.MustNewConstMetric(
			poolEnabledState,
			prometheus.GaugeValue,
			float64(enabledState),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueDepth,
			prometheus.GaugeValue,
			float64(stats.ConnQDepth.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueServiced,
			prometheus.CounterValue,
			float64(stats.ConnQServiced.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAgeMax,
			prometheus.GaugeValue,
			float64(stats.ConnQAgeMax.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAgeHead,
			prometheus.GaugeValue,
			float64(stats.ConnQAgeHead.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAgeEma,
			prometheus.GaugeValue,
			float64(stats.ConnQAgeEma.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAgeEdm,
			prometheus.GaugeValue,
			float64(stats.ConnQAgeEdm.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllDepth,
			prometheus.GaugeValue,
			float64(stats.ConnQAllDepth.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllServiced,
			prometheus.CounterValue,
			float64(stats.ConnQAllServiced.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllAgeMax,
			prometheus.GaugeValue,
			float64(stats.ConnQAllAgeMax.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllAgeHead,
			prometheus.GaugeValue,
			float64(stats.ConnQAllAgeHead.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllAgeEma,
			prometheus.GaugeValue,
			float64(stats.ConnQAllAgeEma.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolConnQueueAllAgeEdm,
			prometheus.GaugeValue,
			float64(stats.ConnQAllAgeEdm.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolCurrentPriorityGroup,
			prometheus.GaugeValue,
			float64(stats.CurPriogrp.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolHighestPriorityGroup,
			prometheus.GaugeValue,
			float64(stats.HighestPriogrp.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolLowestPriorityGroup,
			prometheus.GaugeValue,
			float64(stats.LowestPriogrp.Value),
			target, poolName, partition,
		))

		m = append(m, prometheus.MustNewConstMetric(
			poolMrMsgIn,
			prometheus.CounterValue,
			float64(stats.MrMsgIn.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMrMsgOut,
			prometheus.CounterValue,
			float64(stats.MrMsgOut.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMrReqIn,
			prometheus.CounterValue,
			float64(stats.MrReqIn.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMrReqOut,
			prometheus.CounterValue,
			float64(stats.MrReqOut.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMrRespIn,
			prometheus.CounterValue,
			float64(stats.MrRespIn.Value),
			target, poolName, partition,
		))
		m = append(m, prometheus.MustNewConstMetric(
			poolMrRespOut,
			prometheus.CounterValue,
			float64(stats.MrRespOut.Value),
			target, poolName, partition,
		))
		wg.Add(1)
		go func(poolName, partition string) {
			defer wg.Done()
			var members PoolMemberResponse
			poolPath := "/mgmt/tm/ltm/pool/~" + partition + "~" + poolName + "/members"
			if err := c.Get(poolPath, &members); err != nil {
				log.Printf("Error getting pool members for pool %s: %v", poolName, err)
				return
			}
			memberCh <- PoolMemberDetails{
				PoolMemberResponse: members,
				Partition:          partition,
				PoolName:           poolName,
			}

		}(poolName, partition)
	}
	go func() {
		wg.Wait()
		close(memberCh)
	}()
	for pool := range memberCh {
		for _, member := range pool.Items {
			memberState := 0
			if strings.ToLower(member.State) == "up" {
				memberState = 1
			}
			m = append(m, prometheus.MustNewConstMetric(
				poolMemberStatus,
				prometheus.GaugeValue,
				float64(memberState),
				target, pool.PoolName, pool.Partition, member.Name, member.Session,
			))
		}
	}

	return m, true

}

func parsePoolName(tmName string) (poolName, partition string) {
	parts := strings.Split(strings.TrimPrefix(tmName, "/"), "/")
	if len(parts) >= 2 {
		return parts[1], parts[0]
	} else if len(parts) == 1 {
		return parts[0], "Common"
	}
	return tmName, "Common"
}

func poolAvailabilityStateToValue(state string) int {
	state = strings.ToLower(state)
	switch state {
	case "available":
		return 2
	case "unknown":
		return 1
	case "offline":
		return 0
	default:
		return 1
	}
}

func poolEnabledStateToValue(state string) int {
	state = strings.ToLower(state)
	switch state {
	case "enabled":
		return 1
	case "disabled":
		return 0
	default:
		return 0
	}
}

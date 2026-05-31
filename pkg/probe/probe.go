package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/Haameed/bigip_exporter/internal/config"
	bigIPHTTP "github.com/Haameed/bigip_exporter/pkg/http"
)

type Collector struct {
	metrics []prometheus.Metric
}
type probeResult struct {
	metrics []prometheus.Metric
	ok      bool
}

type probeFunc func(bigIPHTTP.BigIPHTTP, string) ([]prometheus.Metric, bool)

type probeDetailedFunc struct {
	name     string
	function probeFunc
}

func (p *Collector) Probe(ctx context.Context, target map[string]string, hc *http.Client, savedConfig config.BigIpExporterConfig) (bool, error) {
	tgt, err := url.Parse(target["target"])
	if err != nil {
		return false, fmt.Errorf("url.Parse failed: %v", err)
	}

	if tgt.Scheme != "https" && tgt.Scheme != "http" {
		return false, fmt.Errorf("unsupported scheme %q", tgt.Scheme)
	}

	u := url.URL{
		Scheme: tgt.Scheme,
		Host:   tgt.Host,
	}

	c, err := bigIPHTTP.NewBigIPClient(ctx, u, hc, savedConfig)
	if err != nil {
		return false, err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allProbes = []probeDetailedFunc{
		{"VirtualServers", GetVirtualServersProbe},
		{"Disks", GetDiskProbe},
		{"Compute", GetComputeProbe},
		{"SyncGroup", GetSyncGroupProbe},
		{"Pools", GetPoolProbe},
		{"Certificates", GetCertificateProbe},
		{"Traffic", GetTrafficProbe},
	}

	success := true
	results := make(chan probeResult, len(allProbes))
	for _, aProbe := range allProbes {
		wg.Add(1)
		go func(probe probeDetailedFunc) {
			defer wg.Done()
			m, ok := aProbe.function(c, u.Host)
			results <- probeResult{
				metrics: m,
				ok:      ok,
			}
		}(aProbe)

	}
	go func() {
		wg.Wait()
		close(results)
	}()
	for res := range results {
		mu.Lock()
		p.metrics = append(p.metrics, res.metrics...)
		mu.Unlock()

		if !res.ok {
			success = false
		}
	}

	return success, nil
}

func (p *Collector) Collect(c chan<- prometheus.Metric) {
	for _, m := range p.metrics {
		c <- m
	}
}

func (p *Collector) Describe(_ chan<- *prometheus.Desc) {
	// TODO: Register metric descriptions here (better practice)

}

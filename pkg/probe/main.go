package probe

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Haameed/f5_bigip_exporter/internal/config"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	params := r.URL.Query()
	target := params.Get("target")

	if target == "" {
		http.Error(w, "target parameter is required", http.StatusBadRequest)
		return
	}

	savedConfig := config.GetConfig()

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(savedConfig.ScrapeTimeout)*time.Second)
	defer cancel()

	registry := prometheus.NewRegistry()
	pc := &Collector{}
	registry.MustRegister(pc)

	success, err := pc.Probe(ctx, map[string]string{"target": target}, &http.Client{}, savedConfig)
	duration := time.Since(start).Seconds()

	if err != nil {
		slog.Error("Probe failed", "target", target, "error", err, "duration", duration)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("Probe completed", "target", target, "success", success, "duration", duration)

	probeSuccess := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_success", Help: "Displays whether or not the probe was a success",
	})
	probeDuration := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "probe_duration_seconds", Help: "Returns how long the probe took to complete in seconds",
	})

	probeSuccess.Set(map[bool]float64{true: 1, false: 0}[success])
	probeDuration.Set(duration)

	registry.MustRegister(probeSuccess, probeDuration)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

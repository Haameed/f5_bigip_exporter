package probe

import (
	"log"

	"github.com/Haameed/f5_bigip_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetDiskProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		totalDiskSize = prometheus.NewDesc(
			"bigip_disk_total_size_MB",
			"Total size of the disk in megabytes",
			[]string{"target", "name", "mode"}, nil,
		)
		vgFreeDiskSize = prometheus.NewDesc(
			"bigip_disk_vg_free_size_MB",
			"Volume group free size of the disk in megabytes",
			[]string{"target", "name", "mode"}, nil,
		)
		vgInUsedDiskSize = prometheus.NewDesc(
			"bigip_disk_vg_inused_size_MB",
			"Volume group in used size of the disk in megabytes",
			[]string{"target", "name", "mode"}, nil,
		)
		vgReservedDiskSize = prometheus.NewDesc(
			"bigip_disk_vg_reserved_size_MB",
			"Volume group reserved size of the disk in megabytes",
			[]string{"target", "name", "mode"}, nil,
		)
	)

	type DiskResponse struct {
		Kind       string `json:"kind"`
		Name       string `json:"name"`
		FullPath   string `json:"fullPath"`
		Generation int    `json:"generation"`
		SelfLink   string `json:"selfLink"`
		Mode       string `json:"mode"`
		Size       int64  `json:"size"`
		VgFree     int64  `json:"vgFree"`
		VgInUsed   int64  `json:"vgInUse"`
		VgReserved int64  `json:"vgReserved"`
	}

	type AllDisksResponse struct {
		Kind     string         `json:"kind"`
		SelfLink string         `json:"selfLink"`
		Items    []DiskResponse `json:"items"`
	}

	var DisksResponse AllDisksResponse
	if err := c.Get("/mgmt/tm/sys/disk/logical-disk", &DisksResponse); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	var m []prometheus.Metric
	for _, disk := range DisksResponse.Items {
		m = append(m, prometheus.MustNewConstMetric(totalDiskSize, prometheus.GaugeValue, float64(disk.Size), target, disk.Name, disk.Mode))
		m = append(m, prometheus.MustNewConstMetric(vgFreeDiskSize, prometheus.GaugeValue, float64(disk.VgFree), target, disk.Name, disk.Mode))
		m = append(m, prometheus.MustNewConstMetric(vgInUsedDiskSize, prometheus.GaugeValue, float64(disk.VgInUsed), target, disk.Name, disk.Mode))
		m = append(m, prometheus.MustNewConstMetric(vgReservedDiskSize, prometheus.GaugeValue, float64(disk.VgReserved), target, disk.Name, disk.Mode))
	}
	return m, true
}

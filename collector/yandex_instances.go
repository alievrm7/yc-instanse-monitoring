package collector

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"

	"yandex_exporter/internal/yandexapi"
)

// InstancesCollector собирает метрики об инстансах
type InstancesCollector struct {
	api     yandexapi.Client
	cloudID string
	info    *prometheus.Desc
}

func NewInstancesCollector(api yandexapi.Client, cloudID string) *InstancesCollector {
	return &InstancesCollector{
		api:     api,
		cloudID: cloudID,
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "instance", "info"),
			"Yandex instance information",
			[]string{"cloud", "name", "ip_internal", "ip_external", "status"},
			nil,
		),
	}
}

func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.info
}

func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) {
	instances, err := c.api.ListInstancesByCloud(c.cloudID)
	if err != nil {
		slog.Error("failed to list instances", "err", err)
		return
	}

	for _, i := range instances {
		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1,
			i.CloudID, i.Name, i.IPInternal, i.IPExternal, i.Status,
		)
	}
}

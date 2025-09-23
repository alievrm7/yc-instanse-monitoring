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

	info  *prometheus.Desc
	count *prometheus.Desc
}

// NewInstancesCollector фабрика
func NewInstancesCollector(api yandexapi.Client, cloudID string) *InstancesCollector {
	return &InstancesCollector{
		api:     api,
		cloudID: cloudID,
		count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "instances", "total"),
			"Number of instances in the cloud",
			nil, nil,
		),
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "instance", "info"),
			"Yandex instance information",
			[]string{"cloud", "name", "ip_internal", "ip_external", "status"},
			nil,
		),
	}
}

// Describe — описывает метрики
func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.count
	ch <- c.info
}

// Collect — собирает метрики
func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) {
	instances, err := c.api.ListInstancesByCloud(c.cloudID)
	if err != nil {
		slog.Error("failed to list instances", "err", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.count, prometheus.GaugeValue, float64(len(instances)))

	for _, i := range instances {
		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1,
			i.CloudID, i.Name, i.IPInternal, i.IPExternal, i.Status,
		)
	}
}

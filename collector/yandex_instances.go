package collector

import (
	"yandex_exporter/internal/yandexapi"

	"github.com/prometheus/client_golang/prometheus"
)

// InstancesCollector собирает метрики об инстансах
type InstancesCollector struct {
	api   yandexapi.Client
	info  *prometheus.Desc
	count *prometheus.Desc
}

// NewInstancesCollector фабрика для регистрации
func NewInstancesCollector(api yandexapi.Client) *InstancesCollector {
	return &InstancesCollector{
		api: api,
		count: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "instances", "total"),
			"Number of instances across all clouds and folders",
			nil, nil,
		),
		info: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "instance", "info"),
			"Yandex instance information",
			[]string{"id", "name", "hostname", "ip", "zone", "status", "cloud", "folder"},
			nil,
		),
	}
}

// Describe описывает метрики
func (c *InstancesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.count
	ch <- c.info
}

// Collect собирает метрики
func (c *InstancesCollector) Collect(ch chan<- prometheus.Metric) {
	instances, err := c.api.ListAllInstances()
	if err != nil {
		// если ошибка, то метрик нет
		return
	}

	// общее количество
	ch <- prometheus.MustNewConstMetric(c.count, prometheus.GaugeValue, float64(len(instances)))

	// по каждому инстансу
	for _, i := range instances {
		ch <- prometheus.MustNewConstMetric(
			c.info,
			prometheus.GaugeValue,
			1,
			i.ID, i.Name, i.Hostname, i.IP, i.Zone, i.Status, i.Cloud, i.Folder,
		)
	}
}

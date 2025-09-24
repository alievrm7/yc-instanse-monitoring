package collector

import (
	"log/slog"

	"yandex_exporter/internal/yandexapi"

	"github.com/prometheus/client_golang/prometheus"
)

type QuotaCollector struct {
	api     yandexapi.Client
	cloudID string
	usage   *prometheus.Desc
	limit   *prometheus.Desc
}

func NewQuotaCollector(api yandexapi.Client, cloudID string) *QuotaCollector {
	return &QuotaCollector{
		api:     api,
		cloudID: cloudID,
		usage: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "quota", "usage"),
			"Yandex Cloud quota usage",
			[]string{"cloud", "service", "quota_id"},
			nil,
		),
		limit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "quota", "limit"),
			"Yandex Cloud quota limits",
			[]string{"cloud", "service", "quota_id"},
			nil,
		),
	}
}

func (c *QuotaCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.usage
	ch <- c.limit
}

func (c *QuotaCollector) Collect(ch chan<- prometheus.Metric) {
	services, err := c.api.ListQuotaServices()
	if err != nil {
		slog.Error("failed to list quota services", "err", err)
		return
	}

	for _, svc := range services {
		quotas, err := c.api.ListQuotaLimits(c.cloudID, svc)
		if err != nil {
			slog.Error("failed to list quota limits", "service", svc, "err", err)
			continue
		}

		for _, q := range quotas {
			if q.Usage != nil {
				ch <- prometheus.MustNewConstMetric(
					c.usage,
					prometheus.GaugeValue,
					*q.Usage,
					c.cloudID, svc, q.QuotaID,
				)
			}
			if q.Limit != nil {
				ch <- prometheus.MustNewConstMetric(
					c.limit,
					prometheus.GaugeValue,
					*q.Limit,
					c.cloudID, svc, q.QuotaID,
				)
			}
		}
	}
}

package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "yandex"

// Collector общий интерфейс
type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

// YandexCollector — просто список коллекторов
type YandexCollector struct {
	collectors []Collector
}

func NewYandexCollector(collectors ...Collector) *YandexCollector {
	return &YandexCollector{collectors: collectors}
}

func (yc *YandexCollector) Describe(ch chan<- *prometheus.Desc) {
	// Каждый коллектор сам описывает метрики через Update
}

func (yc *YandexCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range yc.collectors {
		_ = c.Update(ch)
	}
}

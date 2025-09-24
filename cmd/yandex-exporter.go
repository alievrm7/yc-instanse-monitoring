package main

import (
	"net/http"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"yandex_exporter/collector"
	"yandex_exporter/internal/yandexapi"
)

func main() {
	var (
		metricsPath = kingpin.Flag("web.telemetry-path", "Path to expose metrics.").
				Default("/metrics").String()
		disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Disable exporter self-metrics.").
					Default("false").Bool()
		maxRequests = kingpin.Flag("web.max-requests", "Max parallel scrape requests.").
				Default("40").Int()
		maxProcs = kingpin.Flag("runtime.gomaxprocs", "Go MAXPROCS").
				Envar("GOMAXPROCS").Default("1").Int()
		ycTokenFile = kingpin.Flag("yandex.token-file", "Path to file with YC IAM token").
				Envar("YC_TOKEN_FILE").Required().String()
		ycCloud = kingpin.Flag("yandex.cloud", "Yandex Cloud ID").
			Envar("YC_CLOUD_ID").Required().String()

		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":8080")
	)

	// --- Логгер ---
	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)

	kingpin.Version(version.Print("yandex_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promslogConfig)

	logger.Info("Starting yandex_exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	runtime.GOMAXPROCS(*maxProcs)

	// --- API client ---
	api := yandexapi.NewClient(*ycTokenFile)

	// --- Collectors ---
	instCollector := collector.NewInstancesCollector(api, *ycCloud)
	quotaCollector := collector.NewQuotaCollector(api, *ycCloud)

	// --- Prometheus registry ---
	reg := prometheus.NewRegistry()
	reg.MustRegister(instCollector, quotaCollector)

	if !*disableExporterMetrics {
		reg.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}

	// --- HTTP Handlers ---
	http.Handle(*metricsPath, promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{MaxRequestsInFlight: *maxRequests},
	))

	// --- Run server ---
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

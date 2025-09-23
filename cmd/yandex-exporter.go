package main

import (
	"net/http"
	"os"
	"os/user"
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
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()

		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (process_*, go_*).",
		).Default("false").Bool()

		maxRequests = kingpin.Flag(
			"web.max-requests",
			"Maximum number of parallel scrape requests. Use 0 to disable.",
		).Default("40").Int()

		maxProcs = kingpin.Flag(
			"runtime.gomaxprocs",
			"The target number of CPUs Go will run on (GOMAXPROCS).",
		).Envar("GOMAXPROCS").Default("1").Int()

		ycCloud = kingpin.Flag(
			"yandex.cloud",
			"Yandex Cloud ID",
		).Envar("YC_CLOUD").Required().String()

		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9101")
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
	if u, err := user.Current(); err == nil && u.Uid == "0" {
		logger.Warn("Exporter is running as root — лучше использовать непривилегированного пользователя")
	}
	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	// --- Получаем токен ---
	token, err := yandexapi.GetIAMTokenFromEnv()
	if err != nil {
		logger.Error("failed to get token", "err", err)
		os.Exit(1)
	}

	// --- API client ---
	api := yandexapi.NewClient(token)

	// --- Collectors ---
	instCollector := collector.NewInstancesCollector(api, *ycCloud)
	quotaCollector := collector.NewQuotaCollector(api, *ycCloud)

	// --- Prometheus registry ---
	reg := prometheus.NewRegistry()
	reg.MustRegister(instCollector)
	reg.MustRegister(quotaCollector)

	if !*disableExporterMetrics {
		reg.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}

	// --- HTTP Handlers ---
	http.Handle(*metricsPath, promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			MaxRequestsInFlight: *maxRequests,
		},
	))

	landingConfig := web.LandingConfig{
		Name:        "Yandex Exporter",
		Description: "Prometheus exporter for Yandex.Cloud",
		Version:     version.Info(),
		Links: []web.LandingLinks{
			{Address: *metricsPath, Text: "Metrics"},
		},
	}

	if landingPage, err := web.NewLandingPage(landingConfig); err == nil {
		http.Handle("/", landingPage)
	}

	// --- Run server ---
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

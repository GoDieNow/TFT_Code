package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	metricReporting  *prometheus.GaugeVec
	metricCollection *prometheus.GaugeVec
	metricTime       *prometheus.GaugeVec
	metricCount      *prometheus.GaugeVec
)

func prometheusStart() {

	reg := prometheus.NewPedanticRegistry()

	metricReporting = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: collector + "_Collector",
			Name:      "kafka_send_state",
			Help:      "Reporting information and Kafka topics usage",
		},
		[]string{
			"reason",
			"state",
			"topic",
		},
	)

	metricCollection = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: collector + "_Collector",
			Name:      "Collection",
			Help:      "Collection information and usages data",
		},
		[]string{
			"account",
			"event",
			"reason",
			"state",
			"type",
		},
	)

	metricTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: collector + "_Collector",
			Name:      "collection_time",
			Help:      "Different timing metrics",
		},
		[]string{
			"type",
		},
	)

	metricCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: collector + "_Collector",
			Name:      objects + "_count",
			Help:      "Different VM Counts",
		},
		[]string{
			"type",
		},
	)

	reg.MustRegister(metricReporting, metricCollection, metricTime, metricCount)
	//prometheus.MustRegister(metricReporting, metricCollection)

	l.Trace.Printf("[Prometheus] Starting to serve the metrics.\n")

	go func() {

		if cfg.Prometheus.MetricsExport {

			//http.Handle(cfg.Prometheus.MetricsRoute, promhttp.Handler())
			http.Handle(cfg.Prometheus.MetricsRoute, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

			go log.Fatal(http.ListenAndServe(":"+cfg.Prometheus.MetricsPort, nil))

		}

	}()

}

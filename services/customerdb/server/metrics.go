package main

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	metricTime     *prometheus.GaugeVec
	metricSecurity *prometheus.GaugeVec
)

func prometheusStart() (metricsMap map[string]*prometheus.GaugeVec, register *prometheus.Registry) {

	l.Trace.Printf("[PROMETHEUS] Intializing metrics for the service\n")

	metricsMap = make(map[string]*prometheus.GaugeVec)
	register = prometheus.NewPedanticRegistry()

	metricCache := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "cache_state",
			Help:      "Different cache metrics of fails and usages",
		},
		[]string{
			"state",
			"resource",
		},
	)

	metricCount := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "processed_count",
			Help:      "Different counting metrics for processed tasks",
		},
		[]string{
			"type",
		},
	)

	metricEndpoint := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "api_request",
			Help:      "Different countings metrics for the endpoints",
		},
		[]string{
			"code",
			"method",
			"route",
		},
	)

	metricKafka := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "kafka_state",
			Help:      "Different Kafka metrics of fails and usage of topics",
		},
		[]string{
			"mode",
			"state",
			"topic",
		},
	)

	metricSecurity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "access_control",
			Help:      "Different access control metrics",
		},
		[]string{
			"mode",
			"state",
		},
	)

	metricTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "CYCLOPS",
			Subsystem: strings.Split(service, "-")[0] + "_Service",
			Name:      "processing_time",
			Help:      "Different timing metrics",
		},
		[]string{
			"type",
		},
	)

	register.MustRegister(metricCache, metricCount, metricEndpoint, metricKafka,
		metricSecurity, metricTime)

	metricsMap["cache"] = metricCache
	metricsMap["count"] = metricCount
	metricsMap["api"] = metricEndpoint
	metricsMap["kafka"] = metricKafka
	metricsMap["security"] = metricSecurity
	metricsMap["time"] = metricTime

	return

}

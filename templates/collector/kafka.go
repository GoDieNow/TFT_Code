package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	l "gitlab.com/cyclops-utilities/logging"
)

type kafkaHandlerConf struct {
	out []kafkaPackage
}

type kafkaPackage struct {
	topic     string
	partition int
	channel   chan interface{}
}

// kafkaHandler job is to check the config that it receives and initialize the
// go rutines necesaries to satisfay the configuration it receives.
// Paramenters:
// - kH: kafkaHandlerConf struct with the specific configuration used by the
// service.
func kafkaHandler(kH kafkaHandlerConf) {

	l.Trace.Printf("[KAFKA] Initializing the receivers/senders according to the provided configuration.\n")

	if kH.out != nil {

		for _, p := range kH.out {

			go kafkaSender(p.topic, p.partition, p.channel)

		}

	}
}

// kafkaSender is the abstracted interface handling the sending of data through
// kafka topics.
// Paramenters:
// - t: string containing the kafka-topic in use.
// - p: int containing the kafka-topic partition.
// - c: interface{} channel to receive the data that will be marshalled into
// JSON and then transmitted via kafka.
func kafkaSender(t string, p int, c chan interface{}) {

	l.Trace.Printf("[KAFKA] Initializing kafka sender for topic: %v.\n", t)

	conf := kafka.WriterConfig{
		Brokers:  cfg.Kafka.Brokers,
		Topic:    t,
		Balancer: &kafka.LeastBytes{},
	}

	if cfg.Kafka.TLSEnabled {

		dialer := &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
				PreferServerCipherSuites: true,
				InsecureSkipVerify:       cfg.General.InsecureSkipVerify,
			},
		}

		conf.Dialer = dialer

	}

	w := kafka.NewWriter(conf)
	defer w.Close()

	for {

		v, ok := <-c

		if !ok {

			metricReporting.With(prometheus.Labels{"topic": t, "state": "FAIL", "reason": "Go Channel Problems"}).Inc()

			break

		}

		go func() {

			m, e := json.Marshal(&v)

			if e == nil {

				l.Info.Printf("[KAFKA] Object received through the channel. Starting its processing.\n")

				err := w.WriteMessages(context.Background(),
					kafka.Message{
						Key:   []byte(t + "-" + strconv.Itoa(p)),
						Value: m,
					},
				)

				if err != nil {

					l.Warning.Printf("[KAFKA] There was a problem when sending the record through the stream. Error: %v\n", err)

					metricReporting.With(prometheus.Labels{"topic": t, "state": "FAIL", "reason": "Kafka Stream Problems"}).Inc()

				} else {

					l.Info.Printf("[KAFKA] Object added to the stream succesfully. Topic: %v.\n", t)

					metricReporting.With(prometheus.Labels{"topic": t, "state": "OK", "reason": "Object sent"}).Inc()

				}

			} else {

				l.Warning.Printf("[KAFKA] The information to be sent into the stream cannot be marshalled, please check with the administrator. Error: %v\n", e)

				metricReporting.With(prometheus.Labels{"topic": t, "state": "FAIL", "reason": "JSON Marshalling"}).Inc()

			}

			return

		}()

	}

}

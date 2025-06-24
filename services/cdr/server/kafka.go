package main

import (
	"context"
	"crypto/tls"
	"reflect"
	"strconv"
	"time"

	"github.com/remeh/sizedwaitgroup"
	"github.com/segmentio/encoding/json"

	"github.com/prometheus/client_golang/prometheus"
	kafka "github.com/segmentio/kafka-go"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

type kafkaFunction func(*dbManager.DbParameter, interface{}) error

type kafkaHandlerConf struct {
	in  []kafkaPackage
	out []kafkaPackage
}

type kafkaPackage struct {
	topic     string
	partition int
	model     interface{}
	function  kafkaFunction
	channel   chan interface{}
	saveDB    bool
}

// kafkaHandler job is to check the config that it receives and initialize the
// go rutines necesaries to satisfay the configuration it receives.
// Paramenters:
// - db: DbParameter to direct interaction with the database.
// - monit: statusManager parameter to interact with statusManager subsystem.
// - kH: kafkaHandlerConf struct with the specific configuration used by the
// service.
func kafkaHandler(db *dbManager.DbParameter, monit *statusManager.StatusManager, kH kafkaHandlerConf) {

	l.Trace.Printf("[KAFKA] Initializing the receivers/senders according to the provided configuration.\n")

	if kH.in != nil {

		monit.InitEndpoint("kafka-receiver")

		for _, p := range kH.in {

			go kafkaReceiver(db, monit, p.topic, p.partition, p.model, p.function, p.saveDB)

		}
	}

	if kH.out != nil {

		monit.InitEndpoint("kafka-sender")

		for _, p := range kH.out {

			go kafkaSender(db, monit, p.topic, p.partition, p.channel)

		}

	}
}

// kafkaReceiver is the abstracted interface used to receive JSON data from kafka.
// The functions assumes that by default anything that comes from the kafka topic
// and validates against a specific data model has to be added to the db.
// Besides that it has the option to call an external function to interact with
// the data before putting it in the db.
// Parameters:
// - db: DbParameter to direct interaction with the database.
// - monit: statusManager parameter to interact with statusManager subsystem.
// - t: string containing the kafka-topic in use.
// - p: int containing the kafka-topic partition.
// - m: model to validate data against.
// - f: optional external function for more functionality.
// - saveDB: bool to control is the received data needs to be saved in the db.
func kafkaReceiver(db *dbManager.DbParameter, monit *statusManager.StatusManager, t string, p int, m interface{}, f kafkaFunction, saveDB bool) {

	l.Trace.Printf("[KAFKA] Initializing kafka receiver for topic: %v.\n", t)

	conf := kafka.ReaderConfig{
		Brokers:   cfg.Kafka.Brokers,
		Topic:     t,
		Partition: p,
		MinBytes:  cfg.Kafka.MinBytes,
		MaxBytes:  cfg.Kafka.MaxBytes,
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

	r := kafka.NewReader(conf)
	defer r.Close()

	if e := r.SetOffset(cfg.Kafka.Offset); e != nil {

		l.Warning.Printf("[KAFKA] There was a problem when setting the offset, stopping the kafka handler NOW. Error: %v\n", e)

		db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "FAIL: stream offset"}).Inc()

		return

	}

	swg := sizedwaitgroup.New(8)

	for {

		rm, e := r.ReadMessage(context.Background())

		if e != nil {

			l.Warning.Printf("[KAFKA] Error detected in the kafka stream. Error: %v\n", e)

			db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "FAIL: stream"}).Inc()

			continue

		}

		// Goroutines start
		swg.Add()
		go func() {

			callTime := time.Now()
			monit.APIHit("kafka-receiver", callTime)

			defer swg.Done()

			o := reflect.New(reflect.TypeOf(m)).Interface()

			if e := json.Unmarshal(rm.Value, &o); e == nil {

				l.Info.Printf("[KAFKA] Relevant information detected in the stream: Topic: %v, partition: %v.\n", t, p)

				db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "OK: object received"}).Inc()

				if f != nil {

					l.Info.Printf("[KAFKA] Function for specialized work detected. Starting its processing.\n")

					if err := f(db, o); err != nil {

						l.Warning.Printf("[KAFKA] There was a problem processing the model's specific function. Error: %v\n", err)

						db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "FAIL: linked function"}).Inc()

					} else {

						db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "OK: linked function"}).Inc()

					}

				}

				if saveDB {

					l.Info.Printf("[KAFKA] Saving procesed record in the database.\n")

					if err := db.Db.Create(o).Error; err != nil {

						l.Warning.Printf("[KAFKA] There was a problem adding the record into the database. Error: %v\n", err)

						db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "FAIL: db saving"}).Inc()

					} else {

						db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "OK: object db saved"}).Inc()

					}

				}

			} else {

				l.Warning.Printf("[KAFKA] The information in the stream does not fit the expected model, please check with the administrator. Error: %v\n", e)

				db.Metrics["kafka"].With(prometheus.Labels{"mode": "RECEIVER", "topic": t, "state": "FAIL: stream-rubish"}).Inc()

			}

			monit.APIHitDone("kafka-receiver", callTime)

		}()

	}

}

// kafkaSender is the abstracted interface handling the sending of data through
// kafka topics.
// Paramenters:
// - db: DbParameter to direct interaction with the database.
// - monit: statusManager parameter to interact with statusManager subsystem.
// - t: string containing the kafka-topic in use.
// - p: int containing the kafka-topic partition.
// - c: interface{} channel to receive the data that will be marshalled into
// JSON and then transmitted via kafka.
func kafkaSender(db *dbManager.DbParameter, monit *statusManager.StatusManager, t string, p int, c chan interface{}) {

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

			l.Info.Printf("[KAFKA] The channel has problems or has been closed.\n")

			db.Metrics["kafka"].With(prometheus.Labels{"mode": "SENDER", "topic": t, "state": "FAIL: channel"}).Inc()

			break

		}

		go func() {

			m, e := json.Marshal(&v)

			if e == nil {

				callTime := time.Now()
				monit.APIHit("kafka-sender", callTime)

				l.Info.Printf("[KAFKA] Object received through the channel. Starting its processing.\n")

				err := w.WriteMessages(context.Background(),
					kafka.Message{
						Key:   []byte(t + "-" + strconv.Itoa(p)),
						Value: m,
					},
				)

				if err != nil {

					l.Warning.Printf("[KAFKA] There was a problem when sending the record through the stream. Error: %v\n", err)

					db.Metrics["kafka"].With(prometheus.Labels{"mode": "SENDER", "topic": t, "state": "FAIL: stream"}).Inc()

				} else {

					l.Info.Printf("[KAFKA] Object added to the stream succesfully. Topic: %v.\n", t)

					db.Metrics["kafka"].With(prometheus.Labels{"mode": "SENDER", "topic": t, "state": "OK: object sent"}).Inc()

				}

				monit.APIHitDone("kafka-sender", callTime)

			} else {

				l.Warning.Printf("[KAFKA] The information to be sent into the stream cannot be marshalled, please check with the administrator. Error: %v\n", e)

				db.Metrics["kafka"].With(prometheus.Labels{"mode": "SENDER", "topic": t, "state": "FAIL: JSON Marshalling"}).Inc()

			}

		}()

	}

}

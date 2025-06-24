package main

import (
	"net/http"
	"net/url"
	"strings"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/GoDieNow/TFT_Code/services/cdr/models"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/statusManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/triggerManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/usageManager"
	udrClient "github.com/GoDieNow/TFT_Code/services/udr/client"
	udrModels "github.com/GoDieNow/TFT_Code/services/udr/models"
	l "gitlab.com/cyclops-utilities/logging"
)

// getService is the uService instantiation function. A sample of the elements
// that need to be personalized for a new uService are provided.
// Returns:
// - hadler: return the http.handler with the swagger and (optionally) the CORS
// configured according to the config provided
// - e: in case of any error it's propagated to the invoker
func getService() (handler http.Handler, register *prometheus.Registry, e error) {

	l.Trace.Printf("[MAIN] Intializing service [ %v ] handler\n", strings.Title(service))

	// TABLE0,...,N have to be customized
	db := dbStart(&models.CDRRecord{}, &models.CReport{})
	mon := statusManager.New(db)

	// Prometheus Metrics linked to dbParameter
	db.Metrics, register = prometheusStart()

	// cache linked to the dbParameter
	db.Cache = cacheStart(db.Metrics["cache"])

	// In case of needed, here is where kafka support is started
	// The functions needs to be customized accordingly to the needs of the service
	// And the parts of the service that might need to send messages need to get
	// the channel
	db.Pipe = kafkaStart(db, mon)

	//bp := getBasePath()
	uc := udrClient.Config{
		URL: &url.URL{
			Host:   cfg.General.Services["udr"],
			Path:   udrClient.DefaultBasePath,
			Scheme: "http",
		},
		AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
	}

	// Parts of the service HERE
	t := triggerManager.New(db, mon, uc)
	u := usageManager.New(db, mon)

	// Initiate the http handler, with the objects that are implementing the business logic.
	h, e := restapi.Handler(restapi.Config{
		StatusManagementAPI:  mon,
		TriggerManagementAPI: t,
		UsageManagementAPI:   u,
		Logger:               l.Info.Printf,
		AuthKeycloak:         AuthKeycloak,
		AuthAPIKeyHeader:     AuthAPIKey,
		AuthAPIKeyParam:      AuthAPIKey,
	})

	if e != nil {

		return

	}

	handler = h

	// CORS
	if cfg.General.CORSEnabled {

		l.Trace.Printf("[MAIN] Enabling CORS for the service [ %v ]\n", strings.Title(service))

		handler = cors.New(
			cors.Options{
				Debug:          (cfg.General.LogLevel == "DEBUG") || (cfg.General.LogLevel == "TRACE"),
				AllowedOrigins: cfg.General.CORSOrigins,
				AllowedHeaders: cfg.General.CORSHeaders,
				AllowedMethods: cfg.General.CORSMethods,
			}).Handler(h)

	}

	return

}

// kafkaStart handles the initialization of the kafka service.
// This is a sample function with the most basic usage of the kafka service, it
// should be redefined to match the needs of the service.
// Parameters:
// - db: DbPAramenter reference for interaction with the db.
// - mon: statusManager parameter reference to interact with the statusManager
// subsystem
// Returns:
// - ch: a interface{} channel to be able to send thing through the kafka topic generated.
func kafkaStart(db *dbManager.DbParameter, mon *statusManager.StatusManager) (ch chan interface{}) {

	l.Trace.Printf("[MAIN] Intializing Kafka\n")

	f := func(db *dbManager.DbParameter, model interface{}) (e error) {

		//TODO: Figure out a way to get the Bearer token from the http header somehow
		if e = db.ProcessUDR(*model.(*udrModels.UReport), ""); e != nil {

			l.Warning.Printf("[KAFKA] There was an error while apliying the linked function. Error: %v\n", e)

			db.Metrics["count"].With(prometheus.Labels{"type": "UDR processing FAILED"}).Inc()

			db.Metrics["count"].With(prometheus.Labels{"type": "UDR processing FAILED for Account: " + (model.(*udrModels.UReport)).AccountID}).Inc()

		} else {

			l.Trace.Printf("[KAFKA] Aplication of the linked function done succesfully.\n")

		}

		return

	}

	ch = make(chan interface{}, 500)

	handler := kafkaHandlerConf{
		in: []kafkaPackage{
			{
				topic:    cfg.Kafka.TopicsIn[0],
				model:    udrModels.UReport{},
				function: f,
				saveDB:   false,
			},
		},
		out: []kafkaPackage{
			{
				topic:   cfg.Kafka.TopicsOut[0],
				channel: ch,
			},
		},
	}

	kafkaHandler(db, mon, handler)

	return

}

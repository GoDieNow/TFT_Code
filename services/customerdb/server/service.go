package main

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
	"github.com/GoDieNow/TFT_Code/services/customerdb/restapi"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/customerManager"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/productManager"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/resellerManager"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/statusManager"
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
	db := dbStart(&models.Product{}, &models.Customer{}, &models.Reseller{})
	mon := statusManager.New(db)

	// Prometheus Metrics linked to dbParameter
	db.Metrics, register = prometheusStart()

	bp := getBasePath()

	// Parts of the service HERE
	c := customerManager.New(db, mon, bp)
	p := productManager.New(db, mon, bp)
	r := resellerManager.New(db, mon, bp)

	// Initiate the http handler, with the objects that are implementing the business logic.
	h, e := restapi.Handler(restapi.Config{
		StatusManagementAPI:   mon,
		CustomerManagementAPI: c,
		ProductManagementAPI:  p,
		ResellerManagementAPI: r,
		Logger:                l.Info.Printf,
		AuthKeycloak:          AuthKeycloak,
		AuthAPIKeyHeader:      AuthAPIKey,
		AuthAPIKeyParam:       AuthAPIKey,
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

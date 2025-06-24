package usageManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/cdr/models"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations/usage_management"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// UsageManager is the struct defined to group and contain all the methods
// that interact with the usage endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type UsageManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct UsageManager that grant
// access to the methods to interact with the usage endpoint.
// Parameters:
// - db: a reference to the DbParameter to be able to interact with the db methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - UsageManager: struct to interact with the usage endpoint functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *UsageManager {

	l.Trace.Printf("[UsageManager] Generating new ResellerManager.\n")

	monit.InitEndpoint("usage")

	return &UsageManager{
		db:    db,
		monit: monit,
	}

}

// GetSystemUsage (Swagger func) is the function behind the (GET) endpoint
// /usage
// Its job is to retrieve the usage report given a certain time-window and with
// the posibility of filtering by metric.
func (m *UsageManager) GetSystemUsage(ctx context.Context, params usage_management.GetSystemUsageParams) middleware.Responder {

	l.Trace.Printf("[UsageManager] GetSystemUsage endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("usage", callTime)

	var from, to strfmt.DateTime
	var usage []*models.CReport
	var e error

	metric := ""

	if params.Metric != nil {

		metric = *params.Metric

	}

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	if params.Idlist != nil {

		if usage, e = m.db.GetUsages(*params.Idlist, metric, from, to); e == nil {

			m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage"}).Inc()

			m.monit.APIHitDone("usage", callTime)

			return usage_management.NewGetSystemUsageOK().WithPayload(usage)

		}

	} else {

		if usage, e = m.db.GetUsage("", metric, from, to); e == nil {

			m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage"}).Inc()

			m.monit.APIHitDone("usage", callTime)

			return usage_management.NewGetSystemUsageOK().WithPayload(usage)

		}

	}

	s := "There was an error in the DB operation: " + e.Error()
	returnValueError := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage"}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetSystemUsageInternalServerError().WithPayload(&returnValueError)

}

// GetUsage (Swagger func) is the function behind the (GET) endpoint
// /usage/{id}
// Its job is to retrieve the usage report of the given account during the
// defined time-window, with the posibility of filtering by metric.
func (m *UsageManager) GetUsage(ctx context.Context, params usage_management.GetUsageParams) middleware.Responder {

	l.Trace.Printf("[UsageManager] GetUsage endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("usage", callTime)

	var from, to strfmt.DateTime

	metric := ""

	if params.Metric != nil {

		metric = *params.Metric

	}

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	usage, e := m.db.GetUsage(params.ID, metric, from, to)

	if e == nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage/" + params.ID}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetUsageOK().WithPayload(usage)

	}

	s := "There was an error in the DB operation: " + e.Error()
	returnValueError := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage/" + params.ID}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetUsageInternalServerError().WithPayload(&returnValueError)

}

// GetUsageSummary (Swagger func) is the function behind the (GET) endpoint
// /usage/summary/{id}
// Its job is to retrieve easy to use for the UI usage report of all the products
// linked to the reseller account provided during the defined time-window.
func (m *UsageManager) GetUsageSummary(ctx context.Context, params usage_management.GetUsageSummaryParams) middleware.Responder {

	l.Trace.Printf("[UsageManager] GetUsageSummary endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("usage", callTime)

	var from, to strfmt.DateTime

	if params.From != nil {

		from = *params.From

	} else {

		from = strfmt.DateTime(time.Date(callTime.Year(), callTime.Month(), callTime.Day(), 0, 0, 0, 0, time.UTC))

	}

	if params.To != nil {

		to = *params.To

	} else {

		t := time.Date(callTime.Year(), callTime.Month(), callTime.Day(), 0, 0, 0, 0, time.UTC)
		to = strfmt.DateTime(t.AddDate(0, 0, -1))

	}

	usage, e := m.db.GetUsageSummary(params.ID, from, to)

	if e == nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage/summary/" + params.ID}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetUsageSummaryOK().WithPayload(usage)

	}

	s := "There was an error in the DB operation: " + e.Error()
	returnValueError := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage/summary/" + params.ID}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetUsageSummaryInternalServerError().WithPayload(&returnValueError)

}

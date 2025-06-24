package usageManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/udr/models"
	"github.com/GoDieNow/TFT_Code/services/udr/restapi/operations/usage_management"
	"github.com/GoDieNow/TFT_Code/services/udr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/udr/server/statusManager"
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

	list := ""
	metric := ""

	if params.From != nil {

		from = *params.From

	}

	if params.Idlist != nil {

		list = *params.Idlist

	}

	if params.Metric != nil {

		metric = *params.Metric

	}

	if params.To != nil {

		to = *params.To

	}

	usage, e := m.db.GetUsages(list, metric, from, to)

	if e != nil {

		s := "There was an error retrieving the Usage from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage"}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetSystemUsageInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage"}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetSystemUsageOK().WithPayload(usage)

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

	if params.From != nil {

		from = *params.From

	}

	if params.Metric != nil {

		metric = *params.Metric

	}

	if params.To != nil {

		to = *params.To

	}

	usage, e := m.db.GetUsage(params.ID, metric, from, to)

	if e != nil {

		s := "There was an error retrieving the Usage from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage/" + params.ID}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetUsageInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage/" + params.ID}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetUsageOK().WithPayload(usage)

}

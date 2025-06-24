package usageManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/restapi/operations/usage_management"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

const (
	bigBang   = int64(0)
	endOfTime = int64(32503680000)
)

// UsageManager is the struct defined to group and contain all the methods
// that interact with the usage endpoints.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type UsageManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct UsageManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - UsageManager: struct to interact with UsageManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *UsageManager {

	l.Trace.Printf("[EventManager] Generating new EventManager.\n")

	monit.InitEndpoint("usage")

	return &UsageManager{
		db:    db,
		monit: monit,
	}

}

// GetSystemUsage (Swagger func) is the function behind the (GET) API Endpoint
// /usage
// Its job is to get the usage of all the accounts in the system during the
// provided time-window with the posibility of filtering the results by the
// type of resource.
func (m *UsageManager) GetSystemUsage(ctx context.Context, params usage_management.GetSystemUsageParams) middleware.Responder {

	l.Trace.Printf("[UsageManager] GetSystemUsage endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("usage", callTime)

	from, to := bigBang, endOfTime
	ty := string("")
	rg := string("")

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	if params.Resource != nil {

		ty = *params.Resource

	}

	if params.Region != nil {

		rg = *params.Region

	}

	use, e := m.db.GetSystemUsage(from, to, ty, rg)

	if e != nil {

		s := "Problem retrieving the usage: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage"}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetSystemUsageInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage"}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetSystemUsageOK().WithPayload(use)

}

// GetUsage (Swagger func) is the function behind the (GET) API Endpoint
// /usage/{id}
// Its job is to get the usage of the provided account in the system during the
// provided time-window with the posibility of filtering the results by the
// type of resource.
func (m *UsageManager) GetUsage(ctx context.Context, params usage_management.GetUsageParams) middleware.Responder {

	l.Trace.Printf("[UsageManager] GetUsage endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("usage", callTime)

	from, to := bigBang, endOfTime
	ty := string("")
	rg := string("")

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	if params.Resource != nil {

		ty = *params.Resource

	}

	if params.Region != nil {

		rg = *params.Region

	}

	use, e := m.db.GetUsage(params.ID, ty, rg, from, to)

	if e != nil {

		s := "Problem retrieving the usage from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/usage/" + params.ID}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetUsageInternalServerError().WithPayload(&errorReturn)

	}

	if use != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/usage/" + params.ID}).Inc()

		m.monit.APIHitDone("usage", callTime)

		return usage_management.NewGetUsageOK().WithPayload(use)

	}

	s := "The Usage doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/usage/" + params.ID}).Inc()

	m.monit.APIHitDone("usage", callTime)

	return usage_management.NewGetUsageNotFound().WithPayload(&missingReturn)

}

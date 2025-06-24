package metricsManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/udr/models"
	"github.com/GoDieNow/TFT_Code/services/udr/restapi/operations/metrics_management"
	"github.com/GoDieNow/TFT_Code/services/udr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/udr/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// MetricsManager is the struct defined to group and contain all the methods
// that interact with the Metrics endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type MetricsManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct MetricsManager that grant
// access to the methods to interact with the metrics endpoint.
// Parameters:
// - db: a reference to the DbParameter to be able to interact with the db methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - ResellerManager: struct to interact with the metrics endpoint functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *MetricsManager {

	l.Trace.Printf("[MetricsManager] Generating new MetricsManager.\n")

	monit.InitEndpoint("metrics")

	return &MetricsManager{
		db:    db,
		monit: monit,
	}

}

// GetMetrics (Swagger func) is the function behind the (GET) endpoint
// /metrics
// Its job is to provide a list of the metrics saved in the system that can be
// used for filtering the usage report.
func (m *MetricsManager) GetMetrics(ctx context.Context, params metrics_management.GetMetricsParams) middleware.Responder {

	l.Trace.Printf("[MetricsManager] GetMetrics endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("metrics", callTime)

	metrics, e := m.db.GetMetrics()

	if e == nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/metrics"}).Inc()

		m.monit.APIHitDone("metrics", callTime)

		return metrics_management.NewGetMetricsOK().WithPayload(metrics)

	}

	s := "There was an error retrieving the Metrics from the system: " + e.Error()
	errorReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/metrics"}).Inc()

	m.monit.APIHitDone("metrics", callTime)

	return metrics_management.NewGetMetricsInternalServerError().WithPayload(&errorReturn)

}

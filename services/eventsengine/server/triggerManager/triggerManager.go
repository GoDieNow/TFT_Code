package triggerManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/restapi/operations/trigger_management"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// TriggerManager is the struct defined to group and contain all the methods
// that interact with the trigger subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type TriggerManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

var returnValue models.ErrorResponse

// New is the function to create the struct TriggerManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// Returns:
// - TriggerManager: struct to interact with triggerManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *TriggerManager {

	l.Trace.Printf("[Trigger] Generating new triggerManager.\n")

	monit.InitEndpoint("trigger")

	// Default return string in case something weird happens.
	// It usually means that something went wrong in the dbManager.
	s := "Something unexpected happened, check with the administrator."

	returnValue = models.ErrorResponse{
		ErrorString: &s,
	}

	return &TriggerManager{
		db:    db,
		monit: monit,
	}

}

// ExecSample (Swagger func) is the function behind the /trigger/sample endpoint.
// It is a dummy function for reference.
func (m *TriggerManager) ExecSample(ctx context.Context, params trigger_management.ExecSampleParams) middleware.Responder {

	l.Trace.Printf("[Trigger] ExecSample endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("trigger", callTime)

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/trigger/sample"}).Inc()

	m.monit.APIHitDone("trigger", callTime)

	return trigger_management.NewExecSampleOK()

}

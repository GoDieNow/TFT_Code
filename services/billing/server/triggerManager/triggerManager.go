package triggerManager

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/billing/models"
	"github.com/GoDieNow/TFT_Code/services/billing/restapi/operations/trigger_management"
	"github.com/GoDieNow/TFT_Code/services/billing/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/billing/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// TriggerManager is the struct defined to group and contain all the methods
// that interact with the trigger subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - BasePath: a string with the base path of the system.
type TriggerManager struct {
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
	BasePath string
}

// New is the function to create the struct TriggerManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - TriggerManager: struct to interact with TriggerManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *TriggerManager {

	l.Trace.Printf("[TriggerManager] Generating new triggerManager.\n")

	monit.InitEndpoint("trigger")

	return &TriggerManager{
		db:       db,
		monit:    monit,
		BasePath: bp,
	}

}

// getToken job is to extract the Keycloak Bearer token from the http header.
// Parameters:
// - param: a pointer to the http.Request object from which extract.
// Returns:
// - token: a string containing the extracted token or an empty one in case of
// no token available.
func (m *TriggerManager) getToken(param *http.Request) (token string) {

	if len(param.Header.Get("Authorization")) > 0 {

		token = strings.Fields(param.Header.Get("Authorization"))[1]

	}

	return

}

// PeriodicRun (Swagger func) is the function behind the (GET)
// /trigger/periodicrun endpoint.
// It is a dummy function for reference.
func (m *TriggerManager) PeriodicRun(ctx context.Context, params trigger_management.PeriodicRunParams) middleware.Responder {

	l.Trace.Printf("[Trigger] Periodic endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("trigger", callTime)

	token := m.getToken(params.HTTPRequest)

	if params.Today != nil {

		go m.db.InvoicesGeneration(map[string]interface{}{"today": *params.Today}, token)

	} else {

		go m.db.InvoicesGeneration(nil, token)

	}

	acceptedReturn := models.ItemCreatedResponse{
		APILink: m.BasePath + "/billrun",
		Message: "The autonomous periodic run for billing generation has been started...",
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "GET", "route": "/trigger/periodicrun"}).Inc()

	m.monit.APIHitDone("trigger", callTime)

	return trigger_management.NewPeriodicRunAccepted().WithPayload(&acceptedReturn)

}

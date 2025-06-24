package eventManager

import (
	"context"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/restapi/operations/event_management"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

const (
	bigBang   = int64(0)
	endOfTime = int64(32503680000)
)

// EventManager is the struct defined to group and contain all the methods
// that interact with the events endpoints.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type EventManager struct {
	db      *dbManager.DbParameter
	monit   *statusManager.StatusManager
	filters []string
}

// New is the function to create the struct EventManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - EventManager: struct to interact with EventManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, filters []string) *EventManager {

	l.Trace.Printf("[EventManager] Generating new EventManager.\n")

	monit.InitEndpoint("event")

	return &EventManager{
		db:      db,
		monit:   monit,
		filters: filters,
	}

}

// AddEvent (Swagger func) is the function behind the (POST) API Endpoint
// /event
// Its job is to include in the system the information given by the new event.
func (m *EventManager) AddEvent(ctx context.Context, params event_management.AddEventParams) middleware.Responder {

	l.Trace.Printf("[EventManager] AddEvent endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("event", callTime)

	for _, f := range m.filters {

		if f != "" && strings.Contains(params.Event.ResourceName, f) {

			l.Warning.Printf("Ignoring event: %v", params.Event.ResourceName)

			s := "Event resource is blacklisted, event ignored."
			errorReturn := models.ErrorResponse{
				ErrorString: &s,
			}

			m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/event"}).Inc()

			m.monit.APIHitDone("event", callTime)

			return event_management.NewAddEventInternalServerError().WithPayload(&errorReturn)

		}

	}

	e := m.db.AddEvent(*params.Event)

	if e != nil {

		s := "Error registering the event: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/event"}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewAddEventInternalServerError().WithPayload(&errorReturn)

	}

	createdReturn := models.ItemCreatedResponse{
		Message: "New Event registered in the system.",
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/event"}).Inc()

	m.monit.APIHitDone("event", callTime)

	return event_management.NewAddEventCreated().WithPayload(&createdReturn)

}

// GetState (Swagger func) is the function behind the (GET) API Endpoint
// /event/status/{id}
// Its job is to get the actual state of the provided account.
func (m *EventManager) GetState(ctx context.Context, params event_management.GetStateParams) middleware.Responder {

	l.Trace.Printf("[EventManager] GetState endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("event", callTime)

	state, e := m.db.GetState(params.Account)

	if e != nil {

		s := "Retrieval of the associated states failed: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/event/status/" + params.Account}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewGetStateInternalServerError().WithPayload(&errorReturn)

	}

	if state != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/event/status/" + params.Account}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewGetStateOK().WithPayload(state)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/event/status/" + params.Account}).Inc()

	m.monit.APIHitDone("event", callTime)

	return event_management.NewGetStateNotFound().WithPayload(&missingReturn)

}

// GetHistory (Swagger func) is the function behind the (GET) API Endpoint
// /event/history/{id}
// Its job is to provide the history of changes in the state of the provided
// account.
func (m *EventManager) GetHistory(ctx context.Context, params event_management.GetHistoryParams) middleware.Responder {

	l.Trace.Printf("[EventManager] GetHistory endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("event", callTime)

	ty, rg := string(""), string("")
	from, to := bigBang, endOfTime

	if params.Resource != nil {

		ty = *params.Resource

	}

	if params.Region != nil {

		rg = *params.Region

	}

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	history, e := m.db.GetHistory(params.Account, ty, rg, from, to)

	if e != nil {

		s := "Problem retrieving the history of the event: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/event/history/" + params.Account}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewGetHistoryInternalServerError().WithPayload(&errorReturn)

	}

	if history != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/event/history/" + params.Account}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewGetHistoryOK().WithPayload(history)

	}

	s := "The Usage doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/event/history/" + params.Account}).Inc()

	m.monit.APIHitDone("event", callTime)

	return event_management.NewGetHistoryNotFound().WithPayload(&missingReturn)

}

// ListState (Swagger func) is the function behind the (GET) API Endpoint
// /event/status
// Its job is to get the actual state of the provided account.
func (m *EventManager) ListStates(ctx context.Context, params event_management.ListStatesParams) middleware.Responder {

	l.Trace.Printf("[EventManager] ListState endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("event", callTime)

	var metric, region string

	if params.Resource != nil {

		metric = *params.Resource

	}

	if params.Region != nil {

		region = *params.Region

	}

	state, e := m.db.ListState(metric, region)

	if e != nil {

		s := "List of accounts states failed: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/event/status"}).Inc()

		m.monit.APIHitDone("event", callTime)

		return event_management.NewListStatesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/event/status"}).Inc()

	m.monit.APIHitDone("event", callTime)

	return event_management.NewListStatesOK().WithPayload(state)

}

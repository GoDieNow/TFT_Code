package cycleManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
	"github.com/GoDieNow/TFT_Code/services/planmanager/restapi/operations/cycle_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/planmanager/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

const (
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK
)

// CycleManager is the struct defined to group and contain all the methods
// that interact with the cycle subsystem.
// Parameters:
// - basePath: a string with the base path of the system.
// - db: a DbParameter reference to be able to use the DBManager methods.
// - s.monit. a StatusManager reference to be able to use the status subsystem methods.
type CycleManager struct {
	basePath string
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
}

// New is the function to create the struct CycleManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - CycleManager: struct to interact with CycleManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *CycleManager {

	l.Trace.Printf("[CycleManager] Generating new cycleManager.\n")

	monit.InitEndpoint("cycle")

	return &CycleManager{
		basePath: bp,
		db:       db,
		monit:    monit,
	}

}

// CreateCycle (Swagger func) is the function behind the (POST) endpoint /cycle
// Its job is to add a new Cycle to the system.
func (m *CycleManager) CreateCycle(ctx context.Context, params cycle_management.CreateCycleParams) middleware.Responder {

	l.Trace.Printf("[CycleManager] CreateCycle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("cycle", callTime)

	id, status, e := m.db.CreateCycle(params.Cycle)

	if e != nil {

		s := "Problem creating the new Cycle: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/cycle"}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewCreateCycleInternalServerError().WithPayload(&errorReturn)

	}

	if status == statusDuplicated {

		s := "The Cycle already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/cycle"}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewCreateCycleConflict().WithPayload(&conflictReturn)

	}

	link := m.basePath + "/cycle/" + id

	createReturn := models.ItemCreatedResponse{
		ID:   id,
		Link: link,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/cycle"}).Inc()

	m.monit.APIHitDone("cycle", callTime)

	return cycle_management.NewCreateCycleCreated().WithPayload(&createReturn)

}

// GetCycle (Swagger func) is the function behind the (GET) endpoint /cycle/{id}
// Its job is to get the Cycle linked to the provided id.
func (m *CycleManager) GetCycle(ctx context.Context, params cycle_management.GetCycleParams) middleware.Responder {

	l.Trace.Printf("[CycleManager] GetCycle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("cycle", callTime)

	cycle, e := m.db.GetCycle(params.ID)

	if e != nil {

		s := "Problem retrieving the Cycle from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/cycle/" + params.ID}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewGetCycleInternalServerError().WithPayload(&errorReturn)

	}

	if cycle != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/cycle/" + params.ID}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewGetCycleOK().WithPayload(cycle)

	}

	s := "The Cycle doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/cycle/" + params.ID}).Inc()

	m.monit.APIHitDone("cycle", callTime)

	return cycle_management.NewGetCycleNotFound().WithPayload(&missingReturn)

}

// ListCycles (Swagger func) is the function behind the (GET) endpoint /cycle
// Its job is to get the list of Cycles in the system.
func (m *CycleManager) ListCycles(ctx context.Context, params cycle_management.ListCyclesParams) middleware.Responder {

	l.Trace.Printf("[CycleManager] ListCycles endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("cycle", callTime)

	state := ""

	if params.State != nil {

		state = *params.State

	}

	ty := ""

	if params.Type != nil {

		ty = *params.Type

	}

	cycles, e := m.db.ListCycles(state, ty)

	if e != nil {

		s := "Problem retrieving the Cycles from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/cycle"}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewListCyclesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/cycle"}).Inc()

	m.monit.APIHitDone("cycle", callTime)

	return cycle_management.NewListCyclesOK().WithPayload(cycles)

}

// UpdateCycle (Swagger func) is the function behind the (PUT) endpoint /cycle/{id}
// Its job is to update the liked Cycle to the provided ID with the provided data.
func (m *CycleManager) UpdateCycle(ctx context.Context, params cycle_management.UpdateCycleParams) middleware.Responder {

	l.Trace.Printf("[CycleManager] UpdateCycle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("cycle", callTime)

	state, e := m.db.UpdateCycle(params.ID, params.Cycle)

	if e != nil {

		s := "Problem updating the new Cycle: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/cycle/" + params.ID}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewUpdateCycleInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusMissing {

		s := "The Cycle doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/cycle/" + params.ID}).Inc()

		m.monit.APIHitDone("cycle", callTime)

		return cycle_management.NewUpdateCycleNotFound().WithPayload(&missingReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/cycle/" + params.ID}).Inc()

	m.monit.APIHitDone("cycle", callTime)

	return cycle_management.NewUpdateCycleOK().WithPayload(params.Cycle)

}

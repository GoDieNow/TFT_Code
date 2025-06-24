package planManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
	"github.com/GoDieNow/TFT_Code/services/planmanager/restapi/operations/plan_management"
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

// PlanManager is the struct defined to group and contain all the methods
// that interact with the plan subsystem.
// Parameters:
// - basePath: a string with the base path of the system.
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type PlanManager struct {
	basePath string
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
}

// New is the function to create the struct PlanManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - PlanManager: struct to interact with PlanManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *PlanManager {

	l.Trace.Printf("[PlanManager] Generating new planManager.\n")

	monit.InitEndpoint("plan")

	return &PlanManager{
		basePath: bp,
		db:       db,
		monit:    monit,
	}

}

// CreatePlan (Swagger func) is the function behind the (POST) endpoint /plan
// Its job is to add a new plan in the system.
func (m *PlanManager) CreatePlan(ctx context.Context, params plan_management.CreatePlanParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] CreatePlan endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	id, state, e := m.db.CreatePlan(params.Plan)

	if e != nil {

		s := "Problem creating the new Plan: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/plan"}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewCreatePlanInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusDuplicated {

		s := "The Plan already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/plan"}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewCreatePlanConflict().WithPayload(&conflictReturn)

	}

	link := m.basePath + "/plan/" + id

	createReturn := models.ItemCreatedResponse{
		ID:   id,
		Link: link,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "GET", "route": "/plan"}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewCreatePlanCreated().WithPayload(&createReturn)

}

// GetCompletePlan (Swagger func) is the function behind the (GET) endpoint
// /plan/complete/{id}
// Its job is to retrieve the full information related to a plan provided its ID.
func (m *PlanManager) GetCompletePlan(ctx context.Context, params plan_management.GetCompletePlanParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] GetCompletePlan endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	plan, e := m.db.GetCompletePlan(params.ID)

	if e != nil {

		s := "Problem retrieving the Plan from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/plan/complete/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewGetCompletePlanInternalServerError().WithPayload(&errorReturn)

	}

	if plan != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/plan/complete/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewGetCompletePlanOK().WithPayload(plan)

	}

	s := "The Plan doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/plan/complete/" + params.ID}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewGetCompletePlanNotFound().WithPayload(&missingReturn)

}

// GetPlan (Swagger func) is the function behind the (GET) endpoint /plan/{id}
// Its job is to retrieve the information related to a plan provided its ID.
func (m *PlanManager) GetPlan(ctx context.Context, params plan_management.GetPlanParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] GetPlan endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	plan, e := m.db.GetPlan(params.ID)

	if e != nil {

		s := "Problem retrieving the Plan from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/plan/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewGetPlanInternalServerError().WithPayload(&errorReturn)

	}

	if plan != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/plan/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewGetPlanOK().WithPayload(plan)

	}

	s := "The Plan doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/plan/" + params.ID}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewGetPlanNotFound().WithPayload(&missingReturn)

}

// ListCompletePlans (Swagger func) is the function behind the (GET) endpoint
// /plan/complete
// Its job is to retrieve the full information, including linked skus, related
// to a plan provided its ID.
func (m *PlanManager) ListCompletePlans(ctx context.Context, params plan_management.ListCompletePlansParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] ListCompletePlans endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	plans, e := m.db.ListCompletePlans()

	if e != nil {

		s := "Problem retrieving the Plan from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/plan/complete"}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewListCompletePlansInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/plan/complete"}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewListCompletePlansOK().WithPayload(plans)

}

// ListPlans (Swagger func) is the function behind the (GET) endpoint /plan
// Its job is to retrieve the information related to a plan provided its ID.
func (m *PlanManager) ListPlans(ctx context.Context, params plan_management.ListPlansParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] ListPlans endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	plans, e := m.db.ListPlans()

	if e != nil {

		s := "Problem retrieving the Plan from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/plan"}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewListPlansInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/plan"}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewListPlansOK().WithPayload(plans)

}

// UpdatePlan (Swagger func) is the function behind the (PUT) endpoint /plan/{id}
// Its job is to update the plan linked to the provided ID with the new plan data.
func (m *PlanManager) UpdatePlan(ctx context.Context, params plan_management.UpdatePlanParams) middleware.Responder {

	l.Trace.Printf("[PlanManager] UpdatePlan endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("plan", callTime)

	state, e := m.db.UpdatePlan(params.ID, params.Plan)

	if e != nil {

		s := "Problem retrieving the Plan from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/plan/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewUpdatePlanInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusMissing {

		s := "The Plan doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/plan/" + params.ID}).Inc()

		m.monit.APIHitDone("plan", callTime)

		return plan_management.NewUpdatePlanNotFound().WithPayload(&missingReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/plan/" + params.ID}).Inc()

	m.monit.APIHitDone("plan", callTime)

	return plan_management.NewUpdatePlanOK().WithPayload(params.Plan)

}

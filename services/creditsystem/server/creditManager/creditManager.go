package creditManager

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/restapi/operations/credit_management"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// CreditManager is the struct defined to group and contain all the methods
// that interact with the credit endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type CreditManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct CreditManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - CreditManager: struct to interact with CreditManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *CreditManager {

	l.Trace.Printf("[CreditManager] Generating new CreditManager.\n")

	monit.InitEndpoint("credit")

	return &CreditManager{
		db:    db,
		monit: monit,
	}

}

// AddConsumption (Swagger func) is the function behind the (POST) API Endpoint
// /credit/consumed/{id}
// Its job is to add a system consumption in the credit account provided the
// account ID.
func (m *CreditManager) AddConsumption(ctx context.Context, params credit_management.AddConsumptionParams) middleware.Responder {

	l.Trace.Printf("[CreditManager] AddConsumption endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("credit", time.Now())

	creditStatus, e := m.db.AddConsumption(params.ID, params.Amount, params.Medium)

	if e != nil {

		s := "Problem adding the Comsumption: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewAddConsumptionInternalServerError().WithPayload(&errorReturn)

	}

	if creditStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewAddConsumptionOK().WithPayload(creditStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("credit", callTime)

	return credit_management.NewAddConsumptionNotFound().WithPayload(&missingReturn)

}

// DecreaseCredit (Swagger func) is the function behind the (POST) API Endpoint
// /credit/available/decrease/{id}
// Its job is to decrease the amount of credit for an account in the system
// provided the account ID.
func (m *CreditManager) DecreaseCredit(ctx context.Context, params credit_management.DecreaseCreditParams) middleware.Responder {

	l.Trace.Printf("[CreditManager] DecreaseCredit endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("credit", time.Now())

	creditStatus, e := m.db.DecreaseCredit(params.ID, params.Amount, params.Medium)

	if e != nil {

		s := "Problem decreassing the Credit: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewDecreaseCreditInternalServerError().WithPayload(&errorReturn)

	}

	if creditStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewDecreaseCreditOK().WithPayload(creditStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("credit", callTime)

	return credit_management.NewDecreaseCreditNotFound().WithPayload(&missingReturn)

}

// GetCredit (Swagger func) is the function behind the (GET) API Endpoint
// /account/available/{id}
// Its job is to retrieve the amount of credit available fo an account in the
// system provided the account ID.
func (m *CreditManager) GetCredit(ctx context.Context, params credit_management.GetCreditParams) middleware.Responder {

	l.Trace.Printf("[CreditManager] GetCredit endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("credit", time.Now())

	creditStatus, e := m.db.GetCredit(params.ID)

	if e != nil {

		s := "Problem getting the Credit Status: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewGetCreditInternalServerError().WithPayload(&errorReturn)

	}

	if creditStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewGetCreditOK().WithPayload(creditStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("credit", callTime)

	return credit_management.NewGetCreditNotFound().WithPayload(&missingReturn)

}

// GetHistory (Swagger func) is the function behind the (GET) API Endpoint
// /credit/history/{id}
// Its job is to retrieve history of the credit balance for an account in the
// system provided the account ID.
func (m *CreditManager) GetHistory(ctx context.Context, params credit_management.GetHistoryParams) middleware.Responder {

	l.Trace.Printf("[CreditManager] GetHistory endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("credit", time.Now())

	var filter bool
	var medium string

	if params.FilterSystem != nil {

		filter = *params.FilterSystem

	} else {

		filter = false

	}

	if params.Medium != nil {

		medium = *params.Medium

	}

	creditStatus, e := m.db.GetHistory(params.ID, filter, medium)

	if e != nil {

		s := "Problem getting the Credit History: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewGetHistoryInternalServerError().WithPayload(&errorReturn)

	}

	if creditStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewGetHistoryOK().WithPayload(creditStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("credit", callTime)

	return credit_management.NewGetHistoryNotFound().WithPayload(&missingReturn)

}

// IncreaseCredit (Swagger func) is the function behind the (POST) API Endpoint
// /credit/available/increase/{id}
// Its job is to increase the amount of credit for an account in the system
// provided the account ID.
func (m *CreditManager) IncreaseCredit(ctx context.Context, params credit_management.IncreaseCreditParams) middleware.Responder {

	l.Trace.Printf("[CreditManager] IncreaseCredit endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("credit", time.Now())

	creditStatus, e := m.db.IncreaseCredit(params.ID, params.Amount, params.Medium)

	if e != nil {

		s := "Problem increassing the Credit: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewIncreaseCreditInternalServerError().WithPayload(&errorReturn)

	}

	if creditStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("credit", callTime)

		return credit_management.NewIncreaseCreditOK().WithPayload(creditStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("credit", callTime)

	return credit_management.NewIncreaseCreditNotFound().WithPayload(&missingReturn)

}

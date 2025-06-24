package resellerManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
	"github.com/GoDieNow/TFT_Code/services/customerdb/restapi/operations/reseller_management"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/customerdb/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

const (
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK
)

// ResellerManager is the struct defined to group and contain all the methods
// that interact with the reseller endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - BasePath: a string with the base path of the system.
type ResellerManager struct {
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
	BasePath string
}

// New is the function to create the struct ResellerManager that grant
// access to the methods to interact with the reseller endpoint.
// Parameters:
// - db: a reference to the DbParameter to be able to interact with the db methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// - bp: a string containing the base path of the service.
// Returns:
// - ResellerManager: struct to interact with the reseller endpoint functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *ResellerManager {

	l.Trace.Printf("[ResellerManager] Generating new ResellerManager.\n")

	monit.InitEndpoint("reseller")

	return &ResellerManager{
		db:       db,
		monit:    monit,
		BasePath: bp,
	}

}

// AddReseller (Swagger func) is the function behind the (POST) API Endpoint
// /reseller
// Its function is to add the provided reseller into the db.
func (m *ResellerManager) AddReseller(ctx context.Context, params reseller_management.AddResellerParams) middleware.Responder {

	l.Trace.Printf("[ResellerManager] AddReseller endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("reseller", callTime)

	id, resellerStatus, customerStatus, productStatus := m.db.AddReseller(params.Reseller)

	switch resellerStatus {

	case statusOK:

		if customerStatus == statusOK && productStatus == statusOK {

			acceptedReturn := models.ItemCreatedResponse{
				Message: "Reseller added to the system",
				APILink: m.BasePath + "/reseller/" + id,
			}

			m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/reseller"}).Inc()

			m.monit.APIHitDone("reseller", callTime)

			return reseller_management.NewAddResellerCreated().WithPayload(&acceptedReturn)

		}

	case statusFail:

		s := "It wasn't possible to insert the Reseller in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/reseller"}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewAddResellerInternalServerError().WithPayload(&errorReturn)

	case statusDuplicated:

		s := "The Reseller already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/reseller"}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewAddResellerConflict().WithPayload(&conflictReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		Message: "Reseller added to the system. There might have been problems when adding asociated resellers and/or products",
		APILink: m.BasePath + "/reseller/" + id,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "POST", "route": "/reseller"}).Inc()

	m.monit.APIHitDone("reseller", callTime)

	return reseller_management.NewAddResellerAccepted().WithPayload(&acceptedReturn)

}

// GetReseller (Swagger func) is the function behind the (GET) API Endpoint
// /reseller/{id}
// Its function is to retrieve the information that the system has about the
// reseller whose ID has been provided.
func (m *ResellerManager) GetReseller(ctx context.Context, params reseller_management.GetResellerParams) middleware.Responder {

	l.Trace.Printf("[ResellerManager] GetReseller endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("reseller", callTime)

	reseller, e := m.db.GetReseller(params.ID, m.BasePath)

	if e != nil {

		s := "There was an error retrieving the Reseller from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/reseller/" + params.ID}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewGetResellerInternalServerError().WithPayload(&errorReturn)

	}

	if reseller != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/reseller/" + params.ID}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewGetResellerOK().WithPayload(reseller)

	}

	s := "The Reseller doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/reseller/" + params.ID}).Inc()

	m.monit.APIHitDone("reseller", callTime)

	return reseller_management.NewGetResellerNotFound().WithPayload(&missingReturn)

}

// ListResellers (Swagger func) is the function behind the (GET) API Endpoint
// /reseller
// Its function is to provide a list containing all the resellers in the system.
func (m *ResellerManager) ListResellers(ctx context.Context, params reseller_management.ListResellersParams) middleware.Responder {

	l.Trace.Printf("[ResellerManager] ListResellers endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("reseller", callTime)

	resellers, e := m.db.ListResellers()

	if e != nil {

		s := "There was an error retrieving the Resellers from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/reseller"}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewListResellersInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/reseller"}).Inc()

	m.monit.APIHitDone("reseller", callTime)

	return reseller_management.NewListResellersOK().WithPayload(resellers)

}

// UpdateReseller (Swagger func) is the function behind the (PUT) API Endpoint
// /reseller/{id}
// Its function is to update the reseller whose ID is provided with the new data.
func (m *ResellerManager) UpdateReseller(ctx context.Context, params reseller_management.UpdateResellerParams) middleware.Responder {

	l.Trace.Printf("[ResellerManager] UpdateReseller endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("reseller", callTime)

	reseller := params.Reseller
	reseller.ResellerID = params.ID

	resellerStatus, customerStatus, productStatus := m.db.UpdateReseller(reseller)

	switch resellerStatus {

	case statusOK:

		if customerStatus == statusOK && productStatus == statusOK {

			acceptedReturn := models.ItemCreatedResponse{
				Message: "Reseller updated in the system",
				APILink: m.BasePath + "/reseller/" + params.Reseller.ResellerID,
			}

			m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/reseller/" + params.ID}).Inc()

			m.monit.APIHitDone("reseller", callTime)

			return reseller_management.NewUpdateResellerOK().WithPayload(&acceptedReturn)

		}

	case statusFail:

		s := "It wasn't possible to update the Reseller in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/reseller/" + params.ID}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewUpdateResellerInternalServerError().WithPayload(&errorReturn)

	case statusMissing:

		s := "The Reseller doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/reseller/" + params.ID}).Inc()

		m.monit.APIHitDone("reseller", callTime)

		return reseller_management.NewUpdateResellerNotFound().WithPayload(&missingReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		Message: "Reseller updated in the system. There might have been problems when updating asociated resellers and/or products",
		APILink: m.BasePath + "/reseller/" + params.Reseller.ResellerID,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "PUT", "route": "/reseller/" + params.ID}).Inc()

	m.monit.APIHitDone("reseller", callTime)

	return reseller_management.NewUpdateResellerAccepted().WithPayload(&acceptedReturn)

}

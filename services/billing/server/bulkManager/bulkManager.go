package bulkManager

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/billing/models"
	"github.com/GoDieNow/TFT_Code/services/billing/restapi/operations/bulk_management"
	"github.com/GoDieNow/TFT_Code/services/billing/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/billing/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// BulkManager is the struct defined to group and contain all the methods
// that interact with the bulk subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type BulkManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct BulkManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// Returns:
// - BulkManager: struct to interact with BulkManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *BulkManager {

	l.Trace.Printf("[BulkManager] Generating new bulkManager.\n")

	monit.InitEndpoint("bulk")

	return &BulkManager{
		db:    db,
		monit: monit,
	}

}

func (m *BulkManager) getToken(param *http.Request) (token string) {

	if len(param.Header.Get("Authorization")) > 0 {

		token = strings.Fields(param.Header.Get("Authorization"))[1]

	}

	return

}

// GenerateInvoiceForCustomer (Swagger func) is the function behind
// GetBillRun (Swagger func) is the function behind the (GET) endpoint
// /billrun/{id}
// Its job is to provide the requested billruns from the system.
func (m *BulkManager) GetBillRun(ctx context.Context, params bulk_management.GetBillRunParams) middleware.Responder {

	l.Trace.Printf("[BulkManager] GetBillRun endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bulk", callTime)

	object, e := m.db.GetBillRun(params.ID)

	if e != nil {

		s := "Problem while retrieving the Billrun from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/billrun/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewGetBillRunInternalServerError().WithPayload(&errorReturn)

	}

	if object != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/billrun/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewGetBillRunOK().WithPayload(object)

	}

	s := "The Billrun doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/billrun/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("bulk", callTime)

	return bulk_management.NewGetBillRunNotFound().WithPayload(&missingReturn)

}

// ListBillRuns (Swagger func) is the function behind the (GET) endpoint
// /billrun
// Its job is to provide a list of all the billruns currently in the system.
func (m *BulkManager) ListBillRuns(ctx context.Context, params bulk_management.ListBillRunsParams) middleware.Responder {

	l.Trace.Printf("[BulkManager] ListBillRuns endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bulk", callTime)

	object, e := m.db.ListBillRuns(params.Months)

	if e != nil {

		s := "Problem while retrieving the Billruns from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/billrun"}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewListBillRunsInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/billrun"}).Inc()

	m.monit.APIHitDone("bulk", callTime)

	return bulk_management.NewListBillRunsOK().WithPayload(object)

}

// ListBillRunsByOrganization (Swagger func) is the function behind the (GET)
// endpoint /billrun/organization/{id}
// Its job is to provide a list of all the billruns currently in the system
// linked to the provided organization.
func (m *BulkManager) ListBillRunsByOrganization(ctx context.Context, params bulk_management.ListBillRunsByOrganizationParams) middleware.Responder {

	l.Trace.Printf("[BulkManager] ListBillRuns endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bulk", callTime)

	token := m.getToken(params.HTTPRequest)

	object, e := m.db.ListBillRunsByOrganization(params.ID, token, params.Months)

	if e != nil {

		s := "Problem while retrieving the Billruns from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/billrun/organization/" + params.ID}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewListBillRunsByOrganizationInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/billrun/organization/" + params.ID}).Inc()

	m.monit.APIHitDone("bulk", callTime)

	return bulk_management.NewListBillRunsByOrganizationOK().WithPayload(object)

}

// ReRunAllBillRuns (Swagger func) is the function behind the (PUT) endpoint
// ReRunAllBillRuns (Swagger func) is the function behind the (PUT) endpoint
// /billrun
// Its job is to re-run all the billruns currently not completed in the system.
func (m *BulkManager) ReRunAllBillRuns(ctx context.Context, params bulk_management.ReRunAllBillRunsParams) middleware.Responder {

	l.Trace.Printf("[BulkManager] ReRunAllBillRuns endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bulk", callTime)

	token := m.getToken(params.HTTPRequest)

	state, e := m.db.ReRunBillRun("", params.Months, token)

	if e != nil {

		s := "Problems while trying to re-run the not finished Billruns: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/billrun"}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunAllBillRunsInternalServerError().WithPayload(&errorReturn)

	}

	if state == dbManager.StatusMissing {

		s := "The Billrun doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/billrun"}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunAllBillRunsNotFound().WithPayload(&missingReturn)

	}

	if state == dbManager.StatusOK {

		acceptedReturn := models.ItemCreatedResponse{
			Message: "The request has been added to the queue",
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "PUT", "route": "/billrun"}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunAllBillRunsAccepted().WithPayload(&acceptedReturn)

	}

	s := "This point shouldn't had been reached, contact Diego. Previous error: " + e.Error()
	errorReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/billrun"}).Inc()

	m.monit.APIHitDone("bulk", callTime)

	return bulk_management.NewReRunAllBillRunsInternalServerError().WithPayload(&errorReturn)

}

// ReRunBillRun (Swagger func) is the function behind the (PUT) endpoint
// /billrun/{id}
// Its job is to rerun the provided billrun currently not completed in the system.
func (m *BulkManager) ReRunBillRun(ctx context.Context, params bulk_management.ReRunBillRunParams) middleware.Responder {

	l.Trace.Printf("[BulkManager] ReRunBillRun endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bulk", callTime)

	token := m.getToken(params.HTTPRequest)

	state, e := m.db.ReRunBillRun(params.ID, nil, token)

	if e != nil {

		s := "Problem while trying to re-run the not finished Billrun: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/billrun/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunBillRunInternalServerError().WithPayload(&errorReturn)

	}

	if state == dbManager.StatusMissing {

		s := "The Billrun doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/billrun/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunBillRunNotFound().WithPayload(&missingReturn)

	}

	if state == dbManager.StatusOK {

		acceptedReturn := models.ItemCreatedResponse{
			Message: "The request has been added to the queue",
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "PUT", "route": "/billrun/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("bulk", callTime)

		return bulk_management.NewReRunBillRunAccepted().WithPayload(&acceptedReturn)

	}

	s := "This point shouldn't had been reached, contact Diego. Previous error: " + e.Error()
	errorReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/billrun/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("bulk", callTime)

	return bulk_management.NewReRunBillRunInternalServerError().WithPayload(&errorReturn)

}

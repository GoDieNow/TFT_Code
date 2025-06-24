package invoiceManager

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/billing/models"
	"github.com/GoDieNow/TFT_Code/services/billing/restapi/operations/invoice_management"
	"github.com/GoDieNow/TFT_Code/services/billing/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/billing/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// InvoiceManager is the struct defined to group and contain all the methods
// that interact with the invoice subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - BasePath: a string with the base path of the system.
type InvoiceManager struct {
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
	BasePath string
}

// New is the function to create the struct InvoiceManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - InvoiceManager: struct to interact with InvoiceManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *InvoiceManager {

	l.Trace.Printf("[InvoiceManager] Generating new invoiceManager.\n")

	monit.InitEndpoint("invoice")

	return &InvoiceManager{
		db:       db,
		monit:    monit,
		BasePath: bp,
	}

}

func (m *InvoiceManager) getToken(param *http.Request) (token string) {

	if len(param.Header.Get("Authorization")) > 0 {

		token = strings.Fields(param.Header.Get("Authorization"))[1]

	}

	return

}

// GenerateInvoiceForCustomer (Swagger func) is the function behind
// the (POST) endpoint /invoice/customer/{id}
// It's job is to start the generation of the invoices associated with the
// requested customer in the system within the optional declared time-window.
func (m *InvoiceManager) GenerateInvoiceForCustomer(ctx context.Context, params invoice_management.GenerateInvoiceForCustomerParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] GenerateInvoiceForCustomer endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	var from, to strfmt.DateTime

	token := m.getToken(params.HTTPRequest)

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	billrun, e := m.db.GenerateInvoicesBy("customer", params.ID, token, from, to)

	if e != nil {

		s := "Problem while trying to generate the Invoice: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/invoice/customer/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGenerateInvoiceForCustomerInternalServerError().WithPayload(&errorReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		APILink: m.BasePath + "/billrun/" + billrun,
		Message: "The request has been added to the queue",
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "POST", "route": "/invoice/customer/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewGenerateInvoiceForCustomerAccepted().WithPayload(&acceptedReturn)

}

// GenerateInvoiceForReseller (Swagger func) is the function behind
// the (POST) endpoint /invoice/reseller/{id}
// It's job is to start the generation of the invoices associated with the
// requested reseller in the system within the optional declared time-window.
func (m *InvoiceManager) GenerateInvoiceForReseller(ctx context.Context, params invoice_management.GenerateInvoiceForResellerParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] GenerateInvoiceForReseller endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	var from, to strfmt.DateTime

	token := m.getToken(params.HTTPRequest)

	if params.From != nil {

		from = *params.From

	}

	if params.To != nil {

		to = *params.To

	}

	billrun, e := m.db.GenerateInvoicesBy("reseller", params.ID, token, from, to)

	if e != nil {

		s := "Problem while trying to generate the Invoice: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/invoice/reseller/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGenerateInvoiceForResellerInternalServerError().WithPayload(&errorReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		APILink: m.BasePath + "/billrun/" + billrun,
		Message: "The request has been added to the queue",
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "POST", "route": "/invoice/reseller/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewGenerateInvoiceForResellerAccepted().WithPayload(&acceptedReturn)

}

// GetInvoice (Swagger func) is the function behind the (GET) endpoint
// /invoice/{id}
// It's job is to provide the requested invoice from the system
func (m *InvoiceManager) GetInvoice(ctx context.Context, params invoice_management.GetInvoiceParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] GetInvoice endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.GetInvoice(params.ID)

	if e != nil {

		s := "Problem while retrieving the Invoice from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoiceInternalServerError().WithPayload(&errorReturn)

	}

	if object != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoiceOK().WithPayload(object)
	}

	s := "The Invoice doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/invoice/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewGetInvoiceNotFound().WithPayload(&missingReturn)

}

// GetInvoicesByCustomer (Swagger func) is the function behind
// the (GET) endpoint /invoice/customer/{id}
// It's job is to provide the invoices associated to the requested customer
// in the system.
func (m *InvoiceManager) GetInvoicesByCustomer(ctx context.Context, params invoice_management.GetInvoicesByCustomerParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] GetInvoicesByCustomer endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.GetInvoicesByOrganization(params.ID, params.Months)

	if e != nil {

		s := "Problem while retrieving the Invoice from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice/customer/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoicesByCustomerInternalServerError().WithPayload(&errorReturn)

	}

	if object != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice/customer/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoicesByCustomerOK().WithPayload(object)
	}

	s := "The invoice doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/invoice/customer/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewGetInvoicesByCustomerNotFound().WithPayload(&missingReturn)

}

// GetInvoicesByReseller (Swagger func) is the function behind
// the (GET) endpoint /invoice/reseller/{id}
// It's job is to provide the invoices associated to the requested reseller
// in the system.
func (m *InvoiceManager) GetInvoicesByReseller(ctx context.Context, params invoice_management.GetInvoicesByResellerParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] GetInvoicesByReseller endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.GetInvoicesByOrganization(params.ID, params.Months)

	if e != nil {

		s := "Problem while retrieving the Invoice from the db: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice/reseller/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoicesByResellerInternalServerError().WithPayload(&errorReturn)

	}

	if object != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice/reseller/" + string(params.ID)}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewGetInvoicesByResellerOK().WithPayload(object)
	}

	s := "The Invoice doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/invoice/reseller/" + string(params.ID)}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewGetInvoicesByResellerNotFound().WithPayload(&missingReturn)

}

// ListInvoices (Swagger func) is the function behind the (GET) endpoint
// /invoice
// It's job is to provide all the invoices contained in the system.
func (m *InvoiceManager) ListInvoices(ctx context.Context, params invoice_management.ListInvoicesParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] ListInvoices endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.ListInvoices(params.Model)

	if e != nil {

		s := "Problem while retrieving the Invoices from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice"}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewListInvoicesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice"}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewListInvoicesOK().WithPayload(object)

}

// ListCustomerInvoices (Swagger func) is the function behind the (GET) endpoint
// /invoice/customer
// It's job is to provide all the customers invoices contained in the system.
func (m *InvoiceManager) ListCustomerInvoices(ctx context.Context, params invoice_management.ListCustomerInvoicesParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] ListCustomerInvoices endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.ListInvoices(&models.Invoice{OrganizationType: "customer"})

	if e != nil {

		s := "Problem while retrieving the Customers Invoices from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice/customer"}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewListCustomerInvoicesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice/customer"}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewListCustomerInvoicesOK().WithPayload(object)

}

// ListResellerInvoices (Swagger func) is the function behind the (GET) endpoint
// /invoice/reseller
// It's job is to provide all the resellers invoices contained in the system.
func (m *InvoiceManager) ListResellerInvoices(ctx context.Context, params invoice_management.ListResellerInvoicesParams) middleware.Responder {

	l.Trace.Printf("[InvoiceManager] ListResellerInvoices endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("invoice", callTime)

	object, e := m.db.ListInvoices(&models.Invoice{OrganizationType: "reseller"})

	if e != nil {

		s := "Problem while retrieving the Resellers Invoices from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/invoice/reseller"}).Inc()

		m.monit.APIHitDone("invoice", callTime)

		return invoice_management.NewListResellerInvoicesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/invoice/reseller"}).Inc()

	m.monit.APIHitDone("invoice", callTime)

	return invoice_management.NewListResellerInvoicesOK().WithPayload(object)

}

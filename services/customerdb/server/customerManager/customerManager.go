package customerManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
	"github.com/GoDieNow/TFT_Code/services/customerdb/restapi/operations/customer_management"
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

// CustomerManager is the struct defined to group and contain all the methods
// that interact with the customer endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - BasePath: a string with the base path of the system.
type CustomerManager struct {
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
	BasePath string
}

// New is the function to create the struct CustomerManager that grant
// access to the methods to interact with the customer endpoint.
// Parameters:
// - db: a reference to the DbParameter to be able to interact with the db methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// - bp: a string containing the base path of the service.
// Returns:
// - CustomerManager: struct to interact with customer endpoint functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *CustomerManager {

	l.Trace.Printf("[CustomerManager] Generating new CustomerManager.\n")

	monit.InitEndpoint("customer")

	return &CustomerManager{
		db:       db,
		monit:    monit,
		BasePath: bp,
	}

}

// AddCustomer (Swagger func) is the function behind the (POST) API Endpoint
// /customer
// Its function is to add the provided customer into the db.
func (m *CustomerManager) AddCustomer(ctx context.Context, params customer_management.AddCustomerParams) middleware.Responder {

	l.Trace.Printf("[CustomerManager] AddCustomer endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("customer", callTime)

	id, customerStatus, productStatus := m.db.AddCustomer(params.Customer)

	switch customerStatus {

	case statusOK:

		if productStatus == statusOK {

			acceptedReturn := models.ItemCreatedResponse{
				Message: "Customer added to the system",
				APILink: m.BasePath + "/customer/" + id,
			}

			m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/customer"}).Inc()

			m.monit.APIHitDone("customer", callTime)

			return customer_management.NewAddCustomerCreated().WithPayload(&acceptedReturn)

		}

	case statusFail:

		s := "It wasn't possible to insert the Customer in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/customer"}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewAddCustomerInternalServerError().WithPayload(&errorReturn)

	case statusDuplicated:

		s := "The Customer already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/customer"}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewAddCustomerConflict().WithPayload(&conflictReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		Message: "Customer added to the system. There might have been problems when adding asociated products",
		APILink: m.BasePath + "/customer/" + id,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "POST", "route": "/customer"}).Inc()

	m.monit.APIHitDone("customer", callTime)

	return customer_management.NewAddCustomerAccepted().WithPayload(&acceptedReturn)

}

// GetCustomer (Swagger func) is the function behind the (GET) API Endpoint
// /customer/{id}
// Its function is to retrieve the information that the system has about the
// customer whose ID has been provided.
func (m *CustomerManager) GetCustomer(ctx context.Context, params customer_management.GetCustomerParams) middleware.Responder {

	l.Trace.Printf("[CustomerManager] GetCustomer endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("customer", callTime)

	customer, e := m.db.GetCustomer(params.ID, m.BasePath)

	if e != nil {

		s := "There was an error retrieving the Customer from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/customer/" + params.ID}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewGetCustomerInternalServerError().WithPayload(&errorReturn)

	}

	if customer != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/customer/" + params.ID}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewGetCustomerOK().WithPayload(customer)

	}

	s := "The Customer doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/customer/" + params.ID}).Inc()

	m.monit.APIHitDone("customer", callTime)

	return customer_management.NewGetCustomerNotFound().WithPayload(&missingReturn)

}

// ListCustomers (Swagger func) is the function behind the (GET) API Endpoint
// /customer
// Its function is to provide a list containing all the customers in the system.
func (m *CustomerManager) ListCustomers(ctx context.Context, params customer_management.ListCustomersParams) middleware.Responder {

	l.Trace.Printf("[CustomerManager] ListCustomers endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("customer", callTime)

	customers, e := m.db.ListCustomers()

	if e != nil {

		s := "There was an error retrieving the Customers from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/customer"}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewListCustomersInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/customer"}).Inc()

	m.monit.APIHitDone("customer", callTime)

	return customer_management.NewListCustomersOK().WithPayload(customers)

}

// UpdateCustomer (Swagger func) is the function behind the (PUT) API Endpoint
// /customer/{id}
// Its function is to update the customer whose ID is provided with the new data.
func (m *CustomerManager) UpdateCustomer(ctx context.Context, params customer_management.UpdateCustomerParams) middleware.Responder {

	l.Trace.Printf("[CustomerManager] UpdateCustomer endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("customer", callTime)

	customer := params.Customer
	customer.CustomerID = params.ID

	customerStatus, productStatus := m.db.UpdateCustomer(customer)

	switch customerStatus {

	case statusOK:

		if productStatus == statusOK {

			acceptedReturn := models.ItemCreatedResponse{
				Message: "Customer updated in the system",
				APILink: m.BasePath + "/customer/" + params.Customer.CustomerID,
			}

			m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/customer/" + params.ID}).Inc()

			m.monit.APIHitDone("customer", callTime)

			return customer_management.NewUpdateCustomerOK().WithPayload(&acceptedReturn)

		}

	case statusFail:

		s := "It wasn't possible to update the Customer in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/customer/" + params.ID}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewUpdateCustomerInternalServerError().WithPayload(&errorReturn)

	case statusMissing:

		s := "The Customer doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/customer/" + params.ID}).Inc()

		m.monit.APIHitDone("customer", callTime)

		return customer_management.NewUpdateCustomerNotFound().WithPayload(&missingReturn)

	}

	acceptedReturn := models.ItemCreatedResponse{
		Message: "Customer updated in the system. There might have been problems when updating asociated products",
		APILink: m.BasePath + "/customer/" + params.Customer.CustomerID,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "202", "method": "PUT", "route": "/customer/" + params.ID}).Inc()

	m.monit.APIHitDone("customer", callTime)

	return customer_management.NewUpdateCustomerAccepted().WithPayload(&acceptedReturn)

}

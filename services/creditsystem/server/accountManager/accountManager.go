package accountManager

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/restapi/operations/account_management"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// AccountManager is the struct defined to group and contain all the methods
// that interact with the accounts endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type AccountManager struct {
	db    *dbManager.DbParameter
	monit *statusManager.StatusManager
}

// New is the function to create the struct AccountManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// Returns:
// - AccountManager: struct to interact with AccountManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager) *AccountManager {

	l.Trace.Printf("[AccountManager] Gerenating new AccountManager.\n")

	monit.InitEndpoint("account")

	return &AccountManager{
		db:    db,
		monit: monit,
	}

}

// CreateAccount (Swagger func) is the function behind the (POST) API Endpoint
// /account/create/{id}
// Its job is to create a new credit account in the system with the ID provided.
func (m *AccountManager) CreateAccount(ctx context.Context, params account_management.CreateAccountParams) middleware.Responder {

	l.Trace.Printf("[AccountManager] CreateAccount endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("account", callTime)

	accountStatus, e := m.db.CreateAccount(params.ID)

	if e != nil {

		s := "Problem creating the Account: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewCreateAccountInternalServerError().WithPayload(&errorReturn)

	}

	if accountStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusCreated), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewCreateAccountCreated().WithPayload(accountStatus)

	}

	s := "The Account already exists in the system."
	conflictReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusConflict), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("account", callTime)

	return account_management.NewCreateAccountConflict().WithPayload(&conflictReturn)

}

// DisableAccount (Swagger func) is the function behind the (POST) API Endpoint
// /account/disable/{id}
// Its job is to disable an account existing in the system given its ID.
func (m *AccountManager) DisableAccount(ctx context.Context, params account_management.DisableAccountParams) middleware.Responder {

	l.Trace.Printf("[AccountManager] DisableAccount endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("account", callTime)

	accountStatus, e := m.db.DisableAccount(params.ID)

	if e != nil {

		s := "Problem creating the Account: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewDisableAccountInternalServerError().WithPayload(&errorReturn)

	}

	if accountStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewDisableAccountOK().WithPayload(accountStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("account", callTime)

	return account_management.NewDisableAccountNotFound().WithPayload(&missingReturn)

}

// EnableAccount (Swagger func) is the function behind the (POST) API Endpoint
// /account/enable/{id}
// Its job is to enable an account existing in the system given its ID.
func (m *AccountManager) EnableAccount(ctx context.Context, params account_management.EnableAccountParams) middleware.Responder {

	l.Trace.Printf("[AccountManager] EnableAccount endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("account", callTime)

	accountStatus, e := m.db.EnableAccount(params.ID)

	if e != nil {

		s := "Problem creating the Account: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewEnableAccountInternalServerError().WithPayload(&errorReturn)

	}

	if accountStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewEnableAccountOK().WithPayload(accountStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("account", callTime)

	return account_management.NewEnableAccountNotFound().WithPayload(&missingReturn)

}

// GetAccountStatus (Swagger func) is the function behind the (GET) API Endpoint
// /account/status/{id}
// Its job is to get the status of an account existing in the system given its ID.
func (m *AccountManager) GetAccountStatus(ctx context.Context, params account_management.GetAccountStatusParams) middleware.Responder {

	l.Trace.Printf("[AccountManager] GetAccountStatus endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("account", callTime)

	accountStatus, e := m.db.GetAccountStatus(params.ID)

	if e != nil {

		s := "Problem creating the Account: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewGetAccountStatusInternalServerError().WithPayload(&errorReturn)

	}

	if accountStatus != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewGetAccountStatusOK().WithPayload(accountStatus)

	}

	s := "The Account doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusNotFound), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("account", callTime)

	return account_management.NewGetAccountStatusNotFound().WithPayload(&missingReturn)

}

// ListAccounts (Swagger func) is the function behind the (GET) API Endpoint
// /account/list
// Its job is to list all the account existing in the system.
func (m *AccountManager) ListAccounts(ctx context.Context, params account_management.ListAccountsParams) middleware.Responder {

	l.Trace.Printf("[AccountManager] ListAccounts endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("account", callTime)

	accountsList, e := m.db.ListAccounts()

	if e != nil {

		s := "Problem creating the Account: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusInternalServerError), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

		m.monit.APIHitDone("account", callTime)

		return account_management.NewListAccountsInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": fmt.Sprint(http.StatusOK), "method": params.HTTPRequest.Method, "route": params.HTTPRequest.URL.Path}).Inc()

	m.monit.APIHitDone("account", callTime)

	return account_management.NewListAccountsOK().WithPayload(accountsList)

}

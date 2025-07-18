// Code generated by go-swagger; DO NOT EDIT.

package restapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"

	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations/status_management"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations/trigger_management"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations/usage_management"
)

type contextKey string

const AuthKey contextKey = "Auth"

//go:generate mockery -name StatusManagementAPI -inpkg

/* StatusManagementAPI  */
type StatusManagementAPI interface {
	/* GetStatus Basic status of the system */
	GetStatus(ctx context.Context, params status_management.GetStatusParams) middleware.Responder

	/* ShowStatus Basic status of the system */
	ShowStatus(ctx context.Context, params status_management.ShowStatusParams) middleware.Responder
}

//go:generate mockery -name TriggerManagementAPI -inpkg

/* TriggerManagementAPI  */
type TriggerManagementAPI interface {
	/* ExecTransformation Transformation of UDR to CDR task trigger */
	ExecTransformation(ctx context.Context, params trigger_management.ExecTransformationParams) middleware.Responder
}

//go:generate mockery -name UsageManagementAPI -inpkg

/* UsageManagementAPI  */
type UsageManagementAPI interface {
	/* GetSystemUsage Detailed report covering all accounts within the specified time window */
	GetSystemUsage(ctx context.Context, params usage_management.GetSystemUsageParams) middleware.Responder

	/* GetUsage Detailed report covering of the account associated with the id within the specified time window */
	GetUsage(ctx context.Context, params usage_management.GetUsageParams) middleware.Responder

	/* GetUsageSummary Summary report meant for the UI for the resources linked to the ResellerID provided within the specified time window */
	GetUsageSummary(ctx context.Context, params usage_management.GetUsageSummaryParams) middleware.Responder
}

// Config is configuration for Handler
type Config struct {
	StatusManagementAPI
	TriggerManagementAPI
	UsageManagementAPI
	Logger func(string, ...interface{})
	// InnerMiddleware is for the handler executors. These do not apply to the swagger.json document.
	// The middleware executes after routing but before authentication, binding and validation
	InnerMiddleware func(http.Handler) http.Handler

	// Authorizer is used to authorize a request after the Auth function was called using the "Auth*" functions
	// and the principal was stored in the context in the "AuthKey" context value.
	Authorizer func(*http.Request) error

	// AuthAPIKeyHeader Applies when the "X-API-KEY" header is set
	AuthAPIKeyHeader func(token string) (interface{}, error)

	// AuthAPIKeyParam Applies when the "api_key" query is set
	AuthAPIKeyParam func(token string) (interface{}, error)

	// AuthKeycloak For OAuth2 authentication
	AuthKeycloak func(token string, scopes []string) (interface{}, error)
	// Authenticator to use for all APIKey authentication
	APIKeyAuthenticator func(string, string, security.TokenAuthentication) runtime.Authenticator
	// Authenticator to use for all Bearer authentication
	BasicAuthenticator func(security.UserPassAuthentication) runtime.Authenticator
	// Authenticator to use for all Basic authentication
	BearerAuthenticator func(string, security.ScopedTokenAuthentication) runtime.Authenticator
}

// Handler returns an http.Handler given the handler configuration
// It mounts all the business logic implementers in the right routing.
func Handler(c Config) (http.Handler, error) {
	h, _, err := HandlerAPI(c)
	return h, err
}

// HandlerAPI returns an http.Handler given the handler configuration
// and the corresponding *CDRManagementAPI instance.
// It mounts all the business logic implementers in the right routing.
func HandlerAPI(c Config) (http.Handler, *operations.CDRManagementAPIAPI, error) {
	spec, err := loads.Analyzed(swaggerCopy(SwaggerJSON), "")
	if err != nil {
		return nil, nil, fmt.Errorf("analyze swagger: %v", err)
	}
	api := operations.NewCDRManagementAPIAPI(spec)
	api.ServeError = errors.ServeError
	api.Logger = c.Logger

	if c.APIKeyAuthenticator != nil {
		api.APIKeyAuthenticator = c.APIKeyAuthenticator
	}
	if c.BasicAuthenticator != nil {
		api.BasicAuthenticator = c.BasicAuthenticator
	}
	if c.BearerAuthenticator != nil {
		api.BearerAuthenticator = c.BearerAuthenticator
	}

	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()
	api.APIKeyHeaderAuth = func(token string) (interface{}, error) {
		if c.AuthAPIKeyHeader == nil {
			return token, nil
		}
		return c.AuthAPIKeyHeader(token)
	}

	api.APIKeyParamAuth = func(token string) (interface{}, error) {
		if c.AuthAPIKeyParam == nil {
			return token, nil
		}
		return c.AuthAPIKeyParam(token)
	}

	api.KeycloakAuth = func(token string, scopes []string) (interface{}, error) {
		if c.AuthKeycloak == nil {
			return token, nil
		}
		return c.AuthKeycloak(token, scopes)
	}
	api.APIAuthorizer = authorizer(c.Authorizer)
	api.TriggerManagementExecTransformationHandler = trigger_management.ExecTransformationHandlerFunc(func(params trigger_management.ExecTransformationParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.TriggerManagementAPI.ExecTransformation(ctx, params)
	})
	api.StatusManagementGetStatusHandler = status_management.GetStatusHandlerFunc(func(params status_management.GetStatusParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.StatusManagementAPI.GetStatus(ctx, params)
	})
	api.UsageManagementGetSystemUsageHandler = usage_management.GetSystemUsageHandlerFunc(func(params usage_management.GetSystemUsageParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.UsageManagementAPI.GetSystemUsage(ctx, params)
	})
	api.UsageManagementGetUsageHandler = usage_management.GetUsageHandlerFunc(func(params usage_management.GetUsageParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.UsageManagementAPI.GetUsage(ctx, params)
	})
	api.UsageManagementGetUsageSummaryHandler = usage_management.GetUsageSummaryHandlerFunc(func(params usage_management.GetUsageSummaryParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.UsageManagementAPI.GetUsageSummary(ctx, params)
	})
	api.StatusManagementShowStatusHandler = status_management.ShowStatusHandlerFunc(func(params status_management.ShowStatusParams, principal interface{}) middleware.Responder {
		ctx := params.HTTPRequest.Context()
		ctx = storeAuth(ctx, principal)
		return c.StatusManagementAPI.ShowStatus(ctx, params)
	})
	api.ServerShutdown = func() {}
	return api.Serve(c.InnerMiddleware), api, nil
}

// swaggerCopy copies the swagger json to prevent data races in runtime
func swaggerCopy(orig json.RawMessage) json.RawMessage {
	c := make(json.RawMessage, len(orig))
	copy(c, orig)
	return c
}

// authorizer is a helper function to implement the runtime.Authorizer interface.
type authorizer func(*http.Request) error

func (a authorizer) Authorize(req *http.Request, principal interface{}) error {
	if a == nil {
		return nil
	}
	ctx := storeAuth(req.Context(), principal)
	return a(req.WithContext(ctx))
}

func storeAuth(ctx context.Context, principal interface{}) context.Context {
	return context.WithValue(ctx, AuthKey, principal)
}

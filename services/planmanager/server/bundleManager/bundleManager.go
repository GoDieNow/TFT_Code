package bundleManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
	"github.com/GoDieNow/TFT_Code/services/planmanager/restapi/operations/bundle_management"
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

// BundleManager is the struct defined to group and contain all the methods
// that interact with the bundle subsystem.
// Parameters:
// - basePath: a string with the base path of the system.
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
type BundleManager struct {
	basePath string
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
}

// New is the function to create the struct BundleManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - BundleManager: struct to interact with BundleManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *BundleManager {

	l.Trace.Printf("[BundleManager] Generating new bundleManager.\n")

	monit.InitEndpoint("bundle")

	return &BundleManager{
		basePath: bp,
		db:       db,
		monit:    monit,
	}

}

// CreateSkuBundle (Swagger func) is the function behind the (POST) endpoint /bundle
// Its job is to add a new bundle to the system.
func (m *BundleManager) CreateSkuBundle(ctx context.Context, params bundle_management.CreateSkuBundleParams) middleware.Responder {

	l.Trace.Printf("[BundleManager] CreateSkuBundle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bundle", callTime)

	id, state, e := m.db.CreateSkuBundle(params.Bundle)

	if e != nil {

		s := "Problem creating the new Bundle: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/sku/bundle"}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewCreateSkuBundleInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusDuplicated {

		s := "The Bundle already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/sku/bundle"}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewCreateSkuBundleConflict().WithPayload(&conflictReturn)

	}

	link := m.basePath + "/sku/bundle/" + id

	createdReturn := models.ItemCreatedResponse{
		ID:   id,
		Link: link,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/sku/bundle"}).Inc()

	m.monit.APIHitDone("bundle", callTime)

	return bundle_management.NewCreateSkuBundleCreated().WithPayload(&createdReturn)

}

// GetSkuBundle (Swagger func) is the function behind the (GET) endpoint /bundle/{id}
// Its job is to retrieve the bundle linked to the provided id.
func (m *BundleManager) GetSkuBundle(ctx context.Context, params bundle_management.GetSkuBundleParams) middleware.Responder {

	l.Trace.Printf("[BundleManager] GetSkuBundle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bundle", callTime)

	bundle, e := m.db.GetSkuBundle(params.ID)

	if e != nil {

		s := "Problem retrieving the Bundle from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/bundle/" + params.ID}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewGetSkuBundleInternalServerError().WithPayload(&errorReturn)

	}

	if bundle != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/bundle/" + params.ID}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewGetSkuBundleOK().WithPayload(bundle)

	}

	s := "The Bundle doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/sku/bundle/" + params.ID}).Inc()

	m.monit.APIHitDone("bundle", callTime)

	return bundle_management.NewGetSkuBundleNotFound().WithPayload(&missingReturn)

}

// GetSkuBundleByName (Swagger func) is the function behind the (GET) endpoint
// /bundle/name/{name}
// Its job is to retrieve the bundle provided the name of the bundle.
func (m *BundleManager) GetSkuBundleByName(ctx context.Context, params bundle_management.GetSkuBundleByNameParams) middleware.Responder {

	l.Trace.Printf("[BundleManager] GetSkuBundleByName endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bundle", callTime)

	bundle, e := m.db.GetSkuBundleByName(params.Name)

	if e != nil {

		s := "Problem retrieving the Bundle from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/bundle/name/" + params.Name}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewGetSkuBundleByNameInternalServerError().WithPayload(&errorReturn)

	}

	if bundle != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/bundle/name/" + params.Name}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewGetSkuBundleByNameOK().WithPayload(bundle)

	}

	s := "The Bundle doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/sku/bundle/name/" + params.Name}).Inc()

	m.monit.APIHitDone("bundle", callTime)

	return bundle_management.NewGetSkuBundleByNameNotFound().WithPayload(&missingReturn)

}

// ListSkuBundles (Swagger func) is the function behind the (GET) endpoint /bundle
// Its job is to list all the bundles present in the system.
func (m *BundleManager) ListSkuBundles(ctx context.Context, params bundle_management.ListSkuBundlesParams) middleware.Responder {

	l.Trace.Printf("[BundleManager] ListSkuBundles endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bundle", callTime)

	bundles, e := m.db.ListSkuBundles()

	if e != nil {

		s := "Problem retrieving Bundles in the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/bundle"}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewListSkuBundlesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/bundle"}).Inc()

	m.monit.APIHitDone("bundle", callTime)

	return bundle_management.NewListSkuBundlesOK().WithPayload(bundles)

}

// UpdateSkuBundle (Swagger func) is the function behind the (PUT) endpoint
// /sku/bundle/{id}
// Its job is to update the linked Sku bundle to the provided ID with the provided data.
func (m *BundleManager) UpdateSkuBundle(ctx context.Context, params bundle_management.UpdateSkuBundleParams) middleware.Responder {

	l.Trace.Printf("[BundleManager] UpdateSkuBundle endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("bundle", callTime)

	state, e := m.db.UpdateSkuBundle(params.ID, params.Bundle)

	if e != nil {

		s := "Problem updating Bundle: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/sku/bundle/" + params.ID}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewUpdateSkuBundleInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusMissing {

		s := "The Bundle doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/sku/bundle/" + params.ID}).Inc()

		m.monit.APIHitDone("bundle", callTime)

		return bundle_management.NewUpdateSkuBundleNotFound().WithPayload(&missingReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/sku/bundle/" + params.ID}).Inc()

	m.monit.APIHitDone("bundle", callTime)

	return bundle_management.NewUpdateSkuBundleOK().WithPayload(params.Bundle)

}

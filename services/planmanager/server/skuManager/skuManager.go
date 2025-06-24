package skuManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
	"github.com/GoDieNow/TFT_Code/services/planmanager/restapi/operations/sku_management"
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

// SkuManager is the struct defined to group and contain all the methods
// that interact with the sku subsystem.
// Parameters:
// - basePath: a string with the base path of the system.
// - db: a DbParameter reference to be able to use the DBManager methods.
// - s.monit. a StatusManager reference to be able to use the status subsystem methods.
type SkuManager struct {
	basePath string
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
}

// New is the function to create the struct SkuManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - SkuManager: struct to interact with SkuManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *SkuManager {

	l.Trace.Printf("[SkuManager] Generating new skuManager.\n")

	monit.InitEndpoint("sku")

	return &SkuManager{
		basePath: bp,
		db:       db,
		monit:    monit,
	}

}

// CreateSku (Swagger func) is the function behind the (POST) endpoint /sku
// Its job is to add a new Sku to the system.
func (m *SkuManager) CreateSku(ctx context.Context, params sku_management.CreateSkuParams) middleware.Responder {

	l.Trace.Printf("[SkuManager] CreateSku endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("sku", callTime)

	id, state, e := m.db.CreateSku(params.Sku)

	if e != nil {

		s := "Problem creating the new Sku: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/sku"}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewCreateSkuInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusDuplicated {

		s := "The Sku already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/sku"}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewCreateSkuConflict().WithPayload(&conflictReturn)

	}

	link := m.basePath + "/sku/" + id

	createReturn := models.ItemCreatedResponse{
		ID:   id,
		Link: link,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/sku"}).Inc()

	m.monit.APIHitDone("sku", callTime)

	return sku_management.NewCreateSkuCreated().WithPayload(&createReturn)

}

// GetSku (Swagger func) is the function behind the (GET) endpoint /sku/{id}
// Its job is to get the Sku linked to the provided id.
func (m *SkuManager) GetSku(ctx context.Context, params sku_management.GetSkuParams) middleware.Responder {

	l.Trace.Printf("[SkuManager] GetSku endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("sku", callTime)

	sku, e := m.db.GetSku(params.ID)

	if e != nil {

		s := "Problem retrieving the Sku from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/" + params.ID}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewGetSkuInternalServerError().WithPayload(&errorReturn)

	}

	if sku != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/" + params.ID}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewGetSkuOK().WithPayload(sku)

	}

	s := "The Sku doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/sku/" + params.ID}).Inc()

	m.monit.APIHitDone("sku", callTime)

	return sku_management.NewGetSkuNotFound().WithPayload(&missingReturn)

}

// ListSkus (Swagger func) is the function behind the (GET) endpoint /sku
// Its job is to get the list of Skus in the system.
func (m *SkuManager) ListSkus(ctx context.Context, params sku_management.ListSkusParams) middleware.Responder {

	l.Trace.Printf("[SkuManager] ListSkus endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("sku", callTime)

	skus, e := m.db.ListSkus()

	if e != nil {

		s := "Problem retrieving the Skus from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku"}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewListSkusInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku"}).Inc()

	m.monit.APIHitDone("sku", callTime)

	return sku_management.NewListSkusOK().WithPayload(skus)

}

// UpdateSku (Swagger func) is the function behind the (PUT) endpoint /sku/{id}
// Its job is to update the liked Sku to the provided ID with the provided data.
func (m *SkuManager) UpdateSku(ctx context.Context, params sku_management.UpdateSkuParams) middleware.Responder {

	l.Trace.Printf("[SkuManager] UpdateSku endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("sku", callTime)

	state, e := m.db.UpdateSku(params.ID, params.Sku)

	if e != nil {

		s := "Problem updating the Sku in the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/sku/" + params.ID}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewUpdateSkuInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusMissing {

		s := "The Sku doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/sku/" + params.ID}).Inc()

		m.monit.APIHitDone("sku", callTime)

		return sku_management.NewUpdateSkuNotFound().WithPayload(&missingReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/sku/" + params.ID}).Inc()

	m.monit.APIHitDone("sku", callTime)

	return sku_management.NewUpdateSkuOK().WithPayload(params.Sku)

}

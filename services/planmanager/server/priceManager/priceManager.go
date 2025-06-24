package priceManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
	"github.com/GoDieNow/TFT_Code/services/planmanager/restapi/operations/price_management"
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

// PriceManager is the struct defined to group and contain all the methods
// that interact with the sku subsystem.
// Parameters:
// - basePath: a string with the base path of the system.
// - db: a DbParameter reference to be able to use the DBManager methods.
// - s.monit. a StatusManager reference to be able to use the status subsystem methods.
type PriceManager struct {
	basePath string
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
}

// New is the function to create the struct PriceManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// - bp: a string containing the base path of the service.
// Returns:
// - PriceManager: struct to interact with PriceManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *PriceManager {

	l.Trace.Printf("[PriceManager] Generating new priceManager.\n")

	monit.InitEndpoint("price")

	return &PriceManager{
		basePath: bp,
		db:       db,
		monit:    monit,
	}

}

// CreateSkuPrice (Swagger func) is the function behind the (POST) endpoint
// /sku/price
// Its job is to add a new Sku price to the system.
func (m *PriceManager) CreateSkuPrice(ctx context.Context, params price_management.CreateSkuPriceParams) middleware.Responder {

	l.Trace.Printf("[PriceManager] CreateSkuPrice endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("price", callTime)

	id, state, e := m.db.CreateSkuPrice(params.Price)

	if e != nil {

		s := "Problem creating the new Sku Price: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/sku/price"}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewCreateSkuPriceInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusDuplicated {

		s := "The Sku Price already exists in the system."
		conflictReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/sku/price"}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewCreateSkuPriceConflict().WithPayload(&conflictReturn)

	}

	link := m.basePath + "/sku/price/" + id

	createReturn := models.ItemCreatedResponse{
		ID:   id,
		Link: link,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/sku/price"}).Inc()

	m.monit.APIHitDone("price", callTime)

	return price_management.NewCreateSkuPriceCreated().WithPayload(&createReturn)

}

// GetSkuPrice (Swagger func) is the function behind the () endpoint /sku/price/{id}
// Its job is to get the Sku price linked to the provided id.
func (m *PriceManager) GetSkuPrice(ctx context.Context, params price_management.GetSkuPriceParams) middleware.Responder {

	l.Trace.Printf("[PriceManager] GetSkuPrice endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("price", callTime)

	sku, e := m.db.GetSkuPrice(params.ID)

	if e != nil {

		s := "Problem creating the new Sku Price: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/price/" + params.ID}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewGetSkuPriceInternalServerError().WithPayload(&errorReturn)

	}

	if sku != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/price/" + params.ID}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewGetSkuPriceOK().WithPayload(sku)

	}

	s := "The Sku Price doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/sku/price/" + params.ID}).Inc()

	m.monit.APIHitDone("price", callTime)

	return price_management.NewGetSkuPriceNotFound().WithPayload(&missingReturn)

}

// ListSkuPrices (Swagger func) is the function behind the (GET) endpoint /sku/price
// Its job is to get the list of Sku prices in the system.
func (m *PriceManager) ListSkuPrices(ctx context.Context, params price_management.ListSkuPricesParams) middleware.Responder {

	l.Trace.Printf("[PriceManager] ListSkuPrices endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("price", callTime)

	skus, e := m.db.ListSkuPrices()

	if e != nil {

		s := "Problem creating the new Sku Price: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/sku/price"}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewListSkuPricesInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/sku/price"}).Inc()

	m.monit.APIHitDone("price", callTime)

	return price_management.NewListSkuPricesOK().WithPayload(skus)

}

// UpdateSkuPrice (Swagger func) is the function behind the (PUT) endpoint
// /sku/price/{id}
// Its job is to update the liked Sku price to the provided ID with the provided data.
func (m *PriceManager) UpdateSkuPrice(ctx context.Context, params price_management.UpdateSkuPriceParams) middleware.Responder {

	l.Trace.Printf("[PriceManager] UpdateSkuPrice endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("price", callTime)

	state, e := m.db.UpdateSkuPrice(params.ID, params.Price)

	if e != nil {

		s := "Problem creating the new Sku Price: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/sku/price" + params.ID}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewUpdateSkuPriceInternalServerError().WithPayload(&errorReturn)

	}

	if state == statusMissing {

		s := "The Sku Price doesn't exists in the system."
		missingReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/sku/price/" + params.ID}).Inc()

		m.monit.APIHitDone("price", callTime)

		return price_management.NewUpdateSkuPriceNotFound().WithPayload(&missingReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/sku/price/" + params.ID}).Inc()

	m.monit.APIHitDone("price", callTime)

	return price_management.NewUpdateSkuPriceOK().WithPayload(params.Price)

}

package productManager

import (
	"context"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
	"github.com/GoDieNow/TFT_Code/services/customerdb/restapi/operations/product_management"
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

// ProductManager is the struct defined to group and contain all the methods
// that interact with the product endpoint.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - BasePath: a string with the base path of the system.
type ProductManager struct {
	db       *dbManager.DbParameter
	monit    *statusManager.StatusManager
	BasePath string
}

// New is the function to create the struct ProductManager that grant
// access to the methods to interact with the Product endpoint.
// Parameters:
// - db: a reference to the DbParameter to be able to interact with the db methods.
// - monit: a reference to the StatusManager to be able to interact with the
// status subsystem.
// - bp: a string containing the base path of the service.
// Returns:
// - ProductManager: struct to interact with product endpoint functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, bp string) *ProductManager {

	l.Trace.Printf("[ProductManager] Generating new ProductManager.\n")

	monit.InitEndpoint("product")

	return &ProductManager{
		db:       db,
		monit:    monit,
		BasePath: bp,
	}

}

// AddProduct (Swagger func) is the function behind the (POST) API Endpoint
// /product
// Its function is to add the provided product into the db.
func (m *ProductManager) AddProduct(ctx context.Context, params product_management.AddProductParams) middleware.Responder {

	l.Trace.Printf("[ProductManager] AddProduct endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("product", callTime)

	id, productStatus := m.db.AddProduct(params.Product)

	switch productStatus {

	case statusOK:

		acceptedReturn := models.ItemCreatedResponse{
			Message: "Product added to the system",
			APILink: m.BasePath + "/product/" + id,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "201", "method": "POST", "route": "/product"}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewAddProductCreated().WithPayload(&acceptedReturn)

	case statusFail:

		s := "It wasn't possible to insert the Product in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/product"}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewAddProductInternalServerError().WithPayload(&errorReturn)

	}

	s := "The Product already exists in the system."
	conflictReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "409", "method": "POST", "route": "/product"}).Inc()

	m.monit.APIHitDone("product", callTime)

	return product_management.NewAddProductConflict().WithPayload(&conflictReturn)

}

// GetProduct (Swagger func) is the function behind the (GET) API Endpoint
// /product/{id}
// Its function is to retrieve the information that the system has about the
// product whose ID has been provided.
func (m *ProductManager) GetProduct(ctx context.Context, params product_management.GetProductParams) middleware.Responder {

	l.Trace.Printf("[ProductManager] GetProduct endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("product", callTime)

	product, e := m.db.GetProduct(params.ID)

	if e != nil {

		s := "There was an error retrieving the Product from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "POST", "route": "/product/" + params.ID}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewGetProductInternalServerError().WithPayload(&errorReturn)

	}

	if product != nil {

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/product/" + params.ID}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewGetProductOK().WithPayload(product)

	}

	s := "The Product doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "GET", "route": "/product/" + params.ID}).Inc()

	m.monit.APIHitDone("product", callTime)

	return product_management.NewGetProductNotFound().WithPayload(&missingReturn)

}

// ListProducts (Swagger func) is the function behind the (GET) API Endpoint
// /product
// Its function is to provide a list containing all the products in the system.
func (m *ProductManager) ListProducts(ctx context.Context, params product_management.ListProductsParams) middleware.Responder {

	l.Trace.Printf("[ProductManager] ListProducts endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("product", callTime)

	products, e := m.db.ListProducts()

	if e != nil {

		s := "There was an error retrieving the Products from the system: " + e.Error()
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/product"}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewListProductsInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/product"}).Inc()

	m.monit.APIHitDone("product", callTime)

	return product_management.NewListProductsOK().WithPayload(products)

}

// UpdateProduct (Swagger func) is the function behind the (PUT) API Endpoint
// /product/{id}
// Its function is to update the product whose ID is provided with the new data.
func (m *ProductManager) UpdateProduct(ctx context.Context, params product_management.UpdateProductParams) middleware.Responder {

	l.Trace.Printf("[ProductManager] UpdateProduct endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("product", callTime)

	product := params.Product
	product.ProductID = params.ID

	switch productStatus := m.db.UpdateProduct(product); productStatus {

	case statusOK:

		acceptedReturn := models.ItemCreatedResponse{
			Message: "Product updated in the system",
			APILink: m.BasePath + "/product/" + params.Product.ProductID,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "PUT", "route": "/product/" + params.ID}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewUpdateProductOK().WithPayload(&acceptedReturn)

	case statusFail:

		s := "It wasn't possible to update the product in the system."
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "PUT", "route": "/product/" + params.ID}).Inc()

		m.monit.APIHitDone("product", callTime)

		return product_management.NewUpdateProductInternalServerError().WithPayload(&errorReturn)

	}

	s := "The Product doesn't exists in the system."
	missingReturn := models.ErrorResponse{
		ErrorString: &s,
	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "404", "method": "PUT", "route": "/product/" + params.ID}).Inc()

	m.monit.APIHitDone("product", callTime)

	return product_management.NewUpdateProductNotFound().WithPayload(&missingReturn)

}

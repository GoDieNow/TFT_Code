package dbManager

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/customerdb/models"
	l "gitlab.com/cyclops-utilities/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK
)

// DbParameter is the struct defined to group and contain all the methods
// that interact with the database.
// On it there is the following parameters:
// - connStr: strings with the connection information to the database
// - Db: a gorm.DB pointer to the db to invoke all the db methods
type DbParameter struct {
	connStr string
	Db      *gorm.DB
	Metrics map[string]*prometheus.GaugeVec
}

// New is the function to create the struct DbParameter.
// Parameters:
// - dbConn: strings with the connection information to the database
// - tables: array of interfaces that will contains the models to migrate
// to the database on initialization
// Returns:
// - DbParameter: struct to interact with dbManager functionalities
func New(dbConn string, tables ...interface{}) *DbParameter {

	l.Trace.Printf("[DB] Gerenating new DBParameter.\n")

	var (
		dp  DbParameter
		err error
	)

	dp.connStr = dbConn

	dp.Db, err = gorm.Open(postgres.Open(dbConn), &gorm.Config{})

	if err != nil {

		l.Error.Printf("[DB] Error opening connection. Error: %v\n", err)

		panic(err)

	}

	l.Trace.Printf("[DB] Migrating tables.\n")

	//Database migration, it handles everything
	dp.Db.AutoMigrate(tables...)

	//l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	//dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NewScope(&models.TABLE).TableName() + "', 'TIMESCALE-ROW-INDEX');")

	return &dp

}

// ListResellers function extracts from the db all the resellers in the system.
// It only list the resellers, not their linked customers neither the linked
// products to each customer.
// Returns:
// - r: slice of the reseller model with all the resellers in the system.
func (d *DbParameter) ListResellers() (r []*models.Reseller, e error) {

	l.Trace.Printf("[DB] Attempting to fetch reseller list now.\n")

	if e = d.Db.Find(&r).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] resellers in the db.\n", len(r))

	return

}

// ListCustomers function extracts from the db all the customers in the system.
// It only list the customerss, not their linked products.
// Returns:
// - c: slice of the customers model with all the customers in the system.
func (d *DbParameter) ListCustomers() (c []*models.Customer, e error) {

	l.Trace.Printf("[DB] Attempting to fetch customers list now.\n")

	if e = d.Db.Find(&c).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] customers in the db.\n", len(c))

	return

}

// ListProducts function extracts from the db all the products in the system.
// Returns:
// - p: slice of the products model with all the products in the system.
func (d *DbParameter) ListProducts() (p []*models.Product, e error) {

	l.Trace.Printf("[DB] Attempting to fetch products list now.\n")

	if e = d.Db.Find(&p).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] products in the db.\n", len(p))

	return

}

// GetReseller function extracts from the db the information that the system
// has about the specified reseller.
// Internally this function makes use of GetCustomers() to fill the information
// about all the customers linked to the specified reseller.
// Parameters:
// - id: string containing the id of the reseller to be gotten.
// - host: string containing the basepath of the server for filling out the APILink
// field of the linked customers and products.
// Returns:
// - A reference to the reseller's model containing all the information in the
// system linked to the specified reseller.
func (d *DbParameter) GetReseller(id string, host string) (*models.Reseller, error) {

	l.Trace.Printf("[DB] Attempting to get reseller [ %v ] now.\n", id)

	var r models.Reseller

	e := d.Db.Where(&models.Reseller{ResellerID: id}).First(&r).Error

	if errors.Is(e, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Unable to fetch existing record for reseller [ %v ], check with administrator.\n", id)

	} else {

		l.Info.Printf("[DB] Found existing record for reseller [ %v ] successfully.\n", r.Name)

		r.Customers = d.GetCustomers(id, host)

	}

	return &r, e

}

// GetCustomer function extracts from the db the information that the system
// has about the specified customer.
// Internally this function makes use of GetProducts() to fill the information
// about all the products linked to the specified customer.
// Parameters:
// - id: string containing the id of the customer to be gotten.
// - host: string containing the basepath of the server for filling out the APILink
// field of the linked products.
// Returns:
// - A reference to the customer's model containing all the information in the
// system linked to the specified customer.
func (d *DbParameter) GetCustomer(id string, host string) (*models.Customer, error) {

	l.Trace.Printf("[DB] Attempting to get customer [ %v ] now.\n", id)

	var c models.Customer

	e := d.Db.Where(&models.Customer{CustomerID: id}).First(&c).Error

	if errors.Is(e, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Unable to fetch existing record for customer [ %v ], check with administrator.\n", id)

	} else {

		l.Info.Printf("[DB] Found existing record for customer [ %v ] successfully.\n", c.Name)

		c.Products = d.GetProducts(id, host)

	}

	return &c, e

}

// GetProduct function extracts from the db the information that the system
// has about the specified product.
// Parameters:
// - id: string containing the id of the product to be gotten.
// Returns:
// - A reference to the product's model containing all the information in the
// system linked to the specified product.
func (d *DbParameter) GetProduct(id string) (*models.Product, error) {

	l.Trace.Printf("[DB] Attempting to get product [ %v ] now.\n", id)

	var p models.Product

	e := d.Db.Where(&models.Product{ProductID: id}).First(&p).Error

	if errors.Is(e, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Unable to fetch existing record for product [ %v ], check with administrator.\n", id)

	} else {

		l.Info.Printf("[DB] Found existing record for product [ %v ] successfully.\n", p.Name)

	}

	return &p, e

}

// GetCustomers function extracts from the db the information that the system
// has about all the customers linked to the specified reseller id.
// Internally this function makes use of GetProducts() to fill the information
// about all the products linked to each customer.
// Parameters:
// - id: string containing the id of the customer to be gotten.
// - host: string containing the basepath of the server for filling out the APILink
// field of the customers and the linked products.
// Returns:
// - A slice of the customer's model containing all the information in the
// system linked to the specified reseller id.
func (d *DbParameter) GetCustomers(id string, host string) (c []*models.Customer) {

	l.Trace.Printf("[DB] Attempting to fetch customers for reseller [ %v ] now.\n", id)

	var custs, customers []models.Customer

	if e := d.Db.Where(&models.Customer{ResellerID: id}).Find(&customers).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation while retrieving linked custommers. Error: %v\n", e)

	}

	for _, customer := range customers {

		customer.Products = d.GetProducts(customer.CustomerID, host)
		customer.APILink = host + "/customer/" + customer.CustomerID

		custs = append(custs, customer)

	}

	//This is the only way to avoid the issue with the unproper assigning of
	//&Elemnt when using range over and array
	for id := range custs {

		c = append(c, &custs[id])

	}

	l.Trace.Printf("[DB] Found [ %d ] customers for reseller [ %v ].\n", len(c), id)

	return

}

// GetProducts function extracts from the db the information that the system
// has about all the products linked to the specified customer id.
// Parameters:
// - id: string containing the id of the customer to be gotten.
// - host: string containing the basepath of the server for filling out the APILink
// field of products.
// Returns:
// - A slice of the product's model containing all the information in the
// system linked to the specified customer id.
func (d *DbParameter) GetProducts(id string, host string) (p []*models.Product) {

	l.Trace.Printf("[DB] Attempting to fetch products for customer [ %v ] now.\n", id)

	var products []models.Product

	if e := d.Db.Where(&models.Product{CustomerID: id}).Find(&products).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation while retrieving linked products. Error: %v\n", e)

	}

	for i := range products {

		product := products[i]

		product.APILink = host + "/product/" + product.ProductID

		p = append(p, &product)

	}

	l.Trace.Printf("[DB] Found [ %d ] products for customer [ %v ].\n", len(p), id)

	return

}

// AddReseller function inserts the reseller and its linked customers and products
// into the system.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the reseller or any of its linked
// components in the db.
// Parameters:
// - r: reseller's model containing the information to be imported in the db.
// Returns:
// - rs: integer with the state of adding the reseller itself.
// - cs: integer with the state of adding the linked customers.
// - ps: integer with the state of adding the linked products.
func (d *DbParameter) AddReseller(r *models.Reseller) (id string, rs int, cs int, ps int) {

	l.Trace.Printf("[DB] Attempting to add reseller [ %v ] now.\n", r.ResellerID)

	var r0 models.Reseller

	if e := d.Db.Where(r).First(&r0).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Create(r); e.Error == nil {

			l.Info.Printf("[DB] Inserted new record for reseller [ %v ] successfully.\n", r.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Resellers added"}).Inc()

			rs, cs, ps = statusOK, statusOK, statusOK

			id = (*e.Statement.Model.(*models.Reseller)).ResellerID

			for _, c := range r.Customers {

				c.ResellerID = id

				_, a, b := d.AddCustomer(c)

				if a != statusOK {

					cs = statusFail

				}

				if b != statusOK {

					ps = statusFail

				}

			}

		} else {

			l.Warning.Printf("[DB] Unable to insert the record for reseller [ %v ], check with administrator.\n", r.Name)

			rs = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for reseller [ %v ] already exists, check with administrator.\n", r.Name)

		rs = statusDuplicated

	}

	return

}

// AddCustomer function inserts the customer and its linked products into the system.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the customer or any of its linked
// components in the db.
// Parameters:
// - c: customer's model containing the information to be imported in the db.
// Returns:
// - cs: integer with the state of adding the customer itself.
// - ps: integer with the state of adding the linked products.
func (d *DbParameter) AddCustomer(c *models.Customer) (id string, cs int, ps int) {

	l.Trace.Printf("[DB] Attempting to add customer [ %v ] now.\n", c.CustomerID)

	var c0 models.Customer

	if e := d.Db.Where(c).First(&c0).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Create(c); e.Error == nil {

			l.Info.Printf("[DB] Inserted new record for customer [ %v ] successfully.\n", c.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Customers added"}).Inc()

			cs, ps = statusOK, statusOK

			id = (*e.Statement.Model.(*models.Customer)).CustomerID

			for _, p := range c.Products {

				p.CustomerID = id

				_, a := d.AddProduct(p)

				if a != statusOK {

					ps = statusFail

				}

			}

		} else {

			l.Warning.Printf("[DB] Unable to insert the record for customer [ %v ], check with administrator.\n", c.Name)

			cs = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for customer [ %v ] already exists, check with administrator.\n", c.Name)

		cs = statusDuplicated

	}

	return

}

// AddProduct function inserts the product into the system.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the product in the db.
// Parameters:
// - p: product's model containing the information to be imported in the db.
// Returns:
// - ps: integer with the state of adding the product.
func (d *DbParameter) AddProduct(p *models.Product) (id string, ps int) {

	l.Trace.Printf("[DB] Attempting to add product [ %v ] now.\n", p.ProductID)

	var p0 models.Product

	if e := d.Db.Where(p).First(&p0).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Create(p); e.Error == nil {

			l.Info.Printf("[DB] Inserted new record for prolduct [ %v ] successfully.\n", p.Name)

			id = (*e.Statement.Model.(*models.Product)).ProductID

			d.Metrics["count"].With(prometheus.Labels{"type": "Products added"}).Inc()

			ps = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to insert the record for product [ %v ], check with administrator.\n", p.Name)

			ps = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for product [ %v ] already exists, check with administrator.\n", p.Name)

		ps = statusDuplicated

	}

	return

}

// UpdateReseller function inserts the updated reseller information and its
// linked customers and products into the system, creating new ones in case of
// no previous version found in the database.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the updated reseller or any of its linked
// components in the db.
// Parameters:
// - r: reseller's model containing the information to be imported in the db.
// Returns:
// - rs: integer with the state of updating the reseller itself.
// - cs: integer with the state of updating the linked customers.
// - ps: integer with the state of updating the linked products.
func (d *DbParameter) UpdateReseller(r *models.Reseller) (rs int, cs int, ps int) {

	l.Trace.Printf("[DB] Attempting to update reseller [ %v ] now.\n", r.ResellerID)

	var r0 models.Reseller

	if e := d.Db.Where(models.Reseller{ResellerID: r.ResellerID}).First(&r0).Error; !errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&r0).Updates(*r).Error; e == nil {

			l.Info.Printf("[DB] Updated record for reseller [ %v ] successfully.\n", r.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Resellers updated"}).Inc()

			rs, cs, ps = statusOK, statusOK, statusOK

			for _, c := range r.Customers {

				c.ResellerID = r.ResellerID

				if a, _ := d.UpdateCustomer(c); a != statusOK {

					l.Trace.Printf("[DB] Attempt to update customer [ %v ] failed, considering it as new customer to be added.\n", c.CustomerID)

					_, b, d := d.AddCustomer(c)

					if b != statusOK {

						l.Trace.Printf("[DB] Attempt to add customer [ %v ] failed.\n", c.CustomerID)

						cs = statusFail

					}

					if d != statusOK {

						ps = statusFail

					}

				}
			}

		} else {

			l.Warning.Printf("[DB] Unable to update record for reseller [ %v ], check with administrator.\n", r.Name)

			rs = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for reseller [ %v ] not found, check with administrator.\n", r.Name)

		rs = statusMissing

	}

	return

}

// UpdateCustomer function inserts the updated customer information and its
// linked products into the system, creating new ones in case of no previous
// version found in the database.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the updated customer or any of its linked
// components in the db.
// Parameters:
// - c: customer's model containing the information to be imported in the db.
// Returns:
// - cs: integer with the state of updating the customer itself.
// - ps: integer with the state of updating the linked products.
func (d *DbParameter) UpdateCustomer(c *models.Customer) (cs int, ps int) {

	l.Trace.Printf("[DB] Attempting to update customer [ %v ] now.\n", c.CustomerID)

	var c0 models.Customer

	if e := d.Db.Where(&models.Customer{CustomerID: c.CustomerID}).First(&c0).Error; !errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&c0).Updates(*c).Error; e == nil {

			l.Info.Printf("[DB] Updated record for customer [ %v ] successfully.\n", c.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Customers updated"}).Inc()

			cs, ps = statusOK, statusOK

			for _, p := range c.Products {

				p.CustomerID = c.CustomerID

				if d.UpdateProduct(p) != statusOK {

					l.Trace.Printf("[DB] Attempt to update product [ %v ] failed, considering it as new product to be added.\n", p.ProductID)

					_, a := d.AddProduct(p)

					if a != statusOK {

						l.Trace.Printf("[DB] Attempt to add product [ %v ] failed.\n", p.ProductID)

						cs = statusFail

					}

				}

			}

		} else {

			l.Warning.Printf("[DB] Unable to update record for customer [ %v ], check with administrator.\n", c.Name)

			cs = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for customer [ %v ] not found, check with administrator.\n", c.Name)

		cs = statusMissing

	}

	return

}

// UpdateProduct function inserts the updated product information in the system.
// Since the system is designed to not raised a problem when it fails to add new
// items in the db, the method constains a basic feedback control to know if
// there was a problem when inserting the updated product in the db.
// Parameters:
// - p: product's model containing the information to be imported in the db.
// Returns:
// - ps: integer with the state of updating the product.
func (d *DbParameter) UpdateProduct(p *models.Product) (ps int) {

	l.Trace.Printf("[DB] Attempting to update product [ %v ] now.\n", p.ProductID)

	var p0 models.Product

	if e := d.Db.Where(&models.Product{ProductID: p.ProductID}).First(&p0).Error; !errors.Is(e, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&p0).Updates(*p).Error; e == nil {

			l.Info.Printf("[DB] Updated record for product [ %v ] successfully.\n", p.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Products updated"}).Inc()

			ps = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for product [ %v ], check with administrator.\n", p.Name)

			ps = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for product [ %v ] not found, check with administrator.\n", p.Name)

		ps = statusMissing

	}

	return

}

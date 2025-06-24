package dbManager

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/planmanager/models"
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
// Parameters:
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

	}

	l.Trace.Printf("[DB] Migrating tables.\n")

	//Database migration, it handles everything
	dp.Db.AutoMigrate(tables...)

	l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	//dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NewScope(&models.TABLE).TableName() + "', 'TIMESCALE-ROW-INDEX');")

	return &dp

}

// CreateCycle function is to add a new cycle to the system.
// Parameters:
// - p: a reference to Cycle models containing the new data to be stored
// in the system.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) CreateCycle(p *models.Cycle) (id string, status int, e error) {

	l.Trace.Printf("[DB] Attempting to create a new Cycle now.\n")

	var p0 models.Cycle

	if r := d.Db.Where(p).First(&p0).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(p); r.Error == nil {

			l.Info.Printf("Inserted new record for cycle [ %v ] successfully.\n", *p.State)

			d.Metrics["count"].With(prometheus.Labels{"type": "Cycles added"}).Inc()

			status = statusOK
			id = (*r.Statement.Model.(*models.Cycle)).ID

		} else {

			l.Warning.Printf("Unable to insert the record for cycle [ %v ], check with administrator.\n", *p.State)

			e = r.Error
			status = statusFail

		}

	} else {

		l.Warning.Printf("Record for cycle [ %v ] already exists, check with administrator.\n", *p.State)

		status = statusDuplicated

	}

	return

}

// CreatePlan function is to add a new plan to the system.
// Parameters:
// - p: a reference to Plan models containing the new data to be stored
// in the system.
// Returns:
// - id: a string containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) CreatePlan(p *models.Plan) (id string, status int, e error) {

	l.Trace.Printf("[DB] Attempting to create a new Plan now.\n")

	var p0 models.Plan

	if r := d.Db.Where(p).First(&p0).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(p); r.Error == nil {

			l.Info.Printf("Inserted new record for plan [ %v ] successfully.\n", *p.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Plans added"}).Inc()

			status = statusOK
			id = (*r.Statement.Model.(*models.Plan)).ID

		} else {

			l.Warning.Printf("Unable to insert the record for plan [ %v ], check with administrator.\n", *p.Name)

			e = r.Error
			status = statusFail

		}

	} else {

		l.Warning.Printf("Record for plan [ %v ] already exists, check with administrator.\n", *p.Name)

		status = statusDuplicated

	}

	return

}

// CreateSku function is to add a new sku to the system.
// Parameters:
// - s: a reference to Sku models containing the new data to be stored
// in the system.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) CreateSku(s *models.Sku) (id string, status int, e error) {

	l.Trace.Printf("[DB] Attempting to create a new Sku now.\n")

	var s0 models.Sku

	if r := d.Db.Where(s).First(&s0).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(s); r.Error == nil {

			l.Info.Printf("Inserted new record for sku [ %v ] successfully.\n", *s.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Skus added"}).Inc()

			status = statusOK
			id = (*r.Statement.Model.(*models.Sku)).ID

		} else {

			l.Warning.Printf("Unable to insert the record for sku [ %v ], check with administrator.\n", *s.Name)

			e = r.Error
			status = statusFail

		}

	} else {

		l.Warning.Printf("Record for sku [ %v ] already exists, check with administrator.\n", *s.Name)

		status = statusDuplicated

	}

	return

}

// CreateSkuBundle function is to add a new bundle to the system.
// Parameters:
// - f: a reference to Bundle models containing the new data to be stored
// in the system.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) CreateSkuBundle(f *models.SkuBundle) (id string, status int, e error) {

	l.Trace.Printf("[DB] Attempting to create a new Bundle now.\n")

	var f0 models.SkuBundle

	if r := d.Db.Where(f).First(&f0).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(f); r.Error == nil {

			l.Info.Printf("Inserted new record for bundle [ %v ] successfully.\n", *f.Name)

			d.Metrics["count"].With(prometheus.Labels{"type": "Bundles added"}).Inc()

			status = statusOK
			id = (r.Statement.Model.(*models.SkuBundle)).ID

		} else {

			l.Warning.Printf("Unable to insert the record for bundle [ %v ], check with administrator.\n", *f.Name)

			e = r.Error
			status = statusFail

		}

	} else {

		l.Warning.Printf("Record for bundle [ %v ] already exists, check with administrator. Error: %v\n", *f.Name, r)

		id = f0.ID
		status = statusDuplicated

	}

	return

}

// CreateSkuPrice function is to add a new sku price to the system.
// Parameters:
// - sp: a reference to Sku Price models containing the new data to be stored
// in the system.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) CreateSkuPrice(sp *models.SkuPrice) (id string, status int, e error) {

	l.Trace.Printf("[DB] Attempting to create a new Sku Price now.\n")

	var sp0 models.SkuPrice

	if r := d.Db.Where(sp).First(&sp0).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(sp); r.Error == nil {

			l.Info.Printf("Inserted new record for sku price [ %v ] successfully.\n", sp.SkuName)

			d.Metrics["count"].With(prometheus.Labels{"type": "Sku Prices added"}).Inc()

			status = statusOK
			id = (*r.Statement.Model.(*models.SkuPrice)).ID

		} else {

			l.Warning.Printf("Unable to insert the record for sku price [ %v ], check with administrator.\n", sp.SkuName)

			e = r.Error
			status = statusFail

		}

	} else {

		l.Warning.Printf("Record for sku price [ %v ] already exists, check with administrator.\n", sp.SkuName)

		status = statusDuplicated

	}

	return

}

// GetCompletePlan function is to retrieve a completed plan (includings its
// linked skus) from the system provided its id.
// Parameters:
// - id: a string containing the id of the plan to be retrieved from the system.
// Returns:
// - reference to Plan model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetCompletePlan(id string) (*models.Plan, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Complete Plan [ %v ] now.\n", id)

	var object models.Plan
	var e error

	if e = d.Db.Where(&models.Plan{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Complete plan with id: %v doesn't exist in the system, check with administrator.", id)

	}

	d.Db.Where(&models.SkuPrice{PlanID: &object.ID}).Find(&object.SkuPrices)

	return &object, e

}

// GetCycle function is to retrieve a cycle from the system provided its id.
// Parameters:
// - id: an int64 containing the id of the cycle to be retrieved from the system.
// Returns:
// - reference to Cycle model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetCycle(id string) (*models.Cycle, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Cycle [ %v ] now.\n", id)

	var object models.Cycle
	var e error

	if e = d.Db.Where(&models.Cycle{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Cycle with id: %v doesn't exist in the system, check with administrator.", id)

	}

	return &object, e

}

// GetPlan function is to retrieve a plan from the system provided its id.
// Parameters:
// - id: a string containing the id of the plan to be retrieved from the system.
// Returns:
// - reference to Plan model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetPlan(id string) (*models.Plan, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Plan [ %v ] now.\n", id)

	var object models.Plan
	var e error

	if e = d.Db.Where(&models.Plan{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Plan with id: %v doesn't exist in the system, check with administrator.", id)

	}

	return &object, e

}

// GetSku function is to retrieve a sku from the system provided its id.
// Parameters:
// - id: an int64 containing the id of the sku to be retrieved from the system.
// Returns:
// - reference to Sku model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetSku(id string) (*models.Sku, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Sku [ %v ] now.\n", id)

	var object models.Sku
	var e error

	if e = d.Db.Where(&models.Sku{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Sku with id: %v doesn't exist in the system, check with administrator.", id)

	}

	return &object, e

}

// GetSkuBundle function is to retrieve a bundle from the system provided its id.
// Parameters:
// - id: an int64 containing the id of the bundle to be retrieved from the system.
// Returns:
// - reference to Bundle model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetSkuBundle(id string) (*models.SkuBundle, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Bundle [ %v ] now.\n", id)

	var object models.SkuBundle
	var e error

	if e = d.Db.Where(&models.SkuBundle{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Bundle with id: %v doesn't exist in the system, check with administrator.", id)

	}

	return &object, e

}

// GetSkuBundleByName function is to retrieve a bundle from the system provided its
// name.
// Parameters:
// - name: a string containing the name of the bundle to be retrieved from the system.
// Returns:
// - reference to Bundle model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetSkuBundleByName(name string) (*models.SkuBundle, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Bundle [ %v ] now.\n", name)

	var object models.SkuBundle
	var e error

	if e = d.Db.Where(&models.SkuBundle{Name: &name}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Bundle with name: %v doesn't exist in the system, check with administrator.", name)

	}

	return &object, e

}

// GetSkuPrice function is to retrieve a sku price from the system provided its id.
// Parameters:
// - id: an int64 containing the id of the sku price to be retrieved from the system.
// Returns:
// - reference to Sku Price model containing the requested one stored in the system.
// - error raised in case of problems with the operation.
func (d *DbParameter) GetSkuPrice(id string) (*models.SkuPrice, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the Sku Price [ %v ] now.\n", id)

	var object models.SkuPrice
	var e error

	if e = d.Db.Where(&models.SkuPrice{ID: id}).First(&object).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Sku price with id: %v doesn't exist in the system, check with administrator.", id)

	}

	return &object, e

}

// ListCompletePlans function is to retrieve all the complete plans (with the
// linked skus) contained in the system.
// Returns:
// - p: a reference to Plan models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListCompletePlans() (p []*models.Plan, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Complete Plans in the system now.\n")

	var p0 []*models.Plan

	if e := d.Db.Find(&p0).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	for i := range p0 {

		plan := p0[i]
		d.Db.Where(&models.SkuPrice{PlanID: &p0[i].ID}).Find(&plan.SkuPrices)

		p = append(p, plan)

	}

	l.Trace.Printf("[DB] Found [ %d ] complete plans in the db.\n", len(p))

	return

}

// ListCycles function is to retrieve all the cycles contained in the system.
// Returns:
// - p: a reference to Cycle models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListCycles(state, ty string) (p []*models.Cycle, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Cycles in the system now.\n")

	var c models.Cycle

	if state != "" {

		c.State = &state

	}

	if ty != "" {

		c.ResourceType = &ty

	}

	if e := d.Db.Where(&c).Find(&p).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] cycles in the db.\n", len(p))

	return

}

// ListPlans function is to retrieve all the plans contained in the system.
// Returns:
// - p: a reference to Plan models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListPlans() (p []*models.Plan, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Plans in the system now.\n")

	if e := d.Db.Find(&p).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] plans in the db.\n", len(p))

	return

}

// ListSkuBundles function is to retrieve all the bundles contained in the system.
// Returns:
// - f: a reference to Bundle models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListSkuBundles() (f []*models.SkuBundle, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Bundles in the system now.\n")

	if e := d.Db.Find(&f).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] bundles in the db.\n", len(f))

	return

}

// ListSkuPrices function is to retrieve all the sku prices contained in the system.
// Returns:
// - sp: a reference to Sku Price models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListSkuPrices() (sp []*models.SkuPrice, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Sku Prices in the system now.\n")

	if e := d.Db.Find(&sp).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] sku prices in the db.\n", len(sp))

	return

}

// ListSkus function is to retrieve all the skus contained in the system.
// Returns:
// - s: a reference to Sku models containing the list of them stored in the system.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) ListSkus() (s []*models.Sku, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the Skus in the system now.\n")

	if e := d.Db.Find(&s).Error; e != nil {

		l.Warning.Printf("[DB] Error in DB operation. Error: %v\n", e)

	}

	l.Trace.Printf("[DB] Found [ %d ] skus in the db.\n", len(s))

	return

}

// UpdateCycle function is to update the cycle linked to the provided ID
// with the provided new data.
// Parameters:
// - id: an int64 containing the id of the sku price to be updated in the db.
// - p: a reference to Cycle models containing the new data to be updated.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) UpdateCycle(id string, p *models.Cycle) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to update the Cycle [ %v ] now.\n", id)

	var p0 models.Cycle

	if r := d.Db.Where(&models.Cycle{ID: id}).First(&p0).Error; !errors.Is(r, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&p0).Updates(p).Error; e == nil {

			l.Info.Printf("[DB] Updated record for cycle [ %v ] successfully.\n", p.ID)

			d.Metrics["count"].With(prometheus.Labels{"type": "Cycles updated"}).Inc()

			status = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for cycle [ %v ], check with administrator.\n", p.ID)

			status = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for cycle [ %v ] not found, check with administrator.\n", p.ID)

		status = statusMissing

	}

	return

}

// UpdatePlan function is to update the plan linked to the provided ID
// with the provided new data.
// Parameters:
// - id: a string containing the id of the sku price to be updated in the db.
// - p: a reference to Plan models containing the new data to be updated.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) UpdatePlan(id string, p *models.Plan) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to update the Plan [ %v ] now.\n", id)

	var p0 models.Plan

	if r := d.Db.Where(&models.Plan{ID: id}).First(&p0).Error; !errors.Is(r, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&p0).Updates(p).Error; e == nil {

			l.Info.Printf("[DB] Updated record for plan [ %v ] successfully.\n", p.ID)

			d.Metrics["count"].With(prometheus.Labels{"type": "Plans updated"}).Inc()

			status = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for plan [ %v ], check with administrator.\n", p.ID)

			status = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for plan [ %v ] not found, check with administrator.\n", p.ID)

		status = statusMissing

	}

	return

}

// UpdateSku function is to update the sku linked to the provided ID
// with the provided new data.
// Parameters:
// - id: an int64 containing the id of the sku price to be updated in the db.
// - s: a reference to Sku models containing the new data to be updated.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) UpdateSku(id string, s *models.Sku) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to update the Sku [ %v ] now.\n", id)

	var s0 models.Sku

	if r := d.Db.Where(&models.Sku{ID: id}).First(&s0).Error; !errors.Is(r, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&s0).Updates(s).Error; e == nil {

			l.Info.Printf("[DB] Updated record for sku [ %v ] successfully.\n", s.ID)

			d.Metrics["count"].With(prometheus.Labels{"type": "Skus updated"}).Inc()

			status = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for sku [ %v ], check with administrator.\n", s.ID)

			status = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for sku [ %v ] not found, check with administrator.\n", s.ID)

		status = statusMissing

	}

	return

}

// UpdateSkuBundle function is to update the bundle linked to the provided ID
// with the provided new data.
// Parameters:
// - f: a reference to Bundle models containing the new data to be updated.
// Returns:
// - id: an int64 containing the id of the bundle just added to the db.
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) UpdateSkuBundle(id string, b *models.SkuBundle) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to update the Bundle [ %v ] now.\n", b.ID)

	var b0 models.SkuBundle

	if r := d.Db.Where(&models.SkuBundle{ID: id}).First(&b0).Error; !errors.Is(r, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&b0).Updates(b).Error; e == nil {

			l.Info.Printf("[DB] Updated record for bundle [ %v ] successfully.\n", b.ID)

			d.Metrics["count"].With(prometheus.Labels{"type": "Bundles updated"}).Inc()

			status = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for bundle [ %v ], check with administrator.\n", b.ID)

			status = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for bundle [ %v ] not found, check with administrator.\n", b.ID)

		status = statusMissing

	}

	return

}

// UpdateSkuPrice function is to update the sku price linked to the provided ID
// with the provided new data.
// Parameters:
// - id: an int64 containing the id of the sku price to be updated in the db.
// - sp: a reference to SkuPrice models containing the new data to be updated.
// Returns:
// - status: an int for informing about the status of the operation.
// - e: an error raised in case of problems with the operation.
func (d *DbParameter) UpdateSkuPrice(id string, sp *models.SkuPrice) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to update the Sku Price [ %v ] now.\n", id)

	var sp0 models.SkuPrice

	if r := d.Db.Where(&models.SkuPrice{ID: id}).First(&sp0).Error; !errors.Is(r, gorm.ErrRecordNotFound) {

		if e := d.Db.Model(&sp0).Updates(sp).Error; e == nil {

			l.Info.Printf("[DB] Updated record for sku price [ %v ] successfully.\n", sp.ID)

			d.Metrics["count"].With(prometheus.Labels{"type": "Sku Prices updated"}).Inc()

			status = statusOK

		} else {

			l.Warning.Printf("[DB] Unable to update record for sku price [ %v ], check with administrator.\n", sp.ID)

			status = statusFail

		}

	} else {

		l.Warning.Printf("[DB] Record for sku price [ %v ] not found, check with administrator.\n", sp.ID)

		status = statusMissing

	}

	return

}

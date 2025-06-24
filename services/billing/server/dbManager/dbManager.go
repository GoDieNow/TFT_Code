package dbManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/cyclops-utilities/datamodels"
	"github.com/GoDieNow/TFT_Code/services/billing/models"
	"github.com/GoDieNow/TFT_Code/services/billing/server/cacheManager"
	cdrModels "github.com/GoDieNow/TFT_Code/services/cdr/models"
	cusModels "github.com/GoDieNow/TFT_Code/services/customerdb/models"
	pmModels "github.com/GoDieNow/TFT_Code/services/planmanager/models"
	l "gitlab.com/cyclops-utilities/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Report status code vars
const (
	scaler           = 1e9
	StatusDuplicated = iota
	StatusFail
	StatusMissing
	StatusOK
)

var (
	periodicity = map[string]int{"daily": 1, "weekly": 7, "bi-weekly": 15,
		"monthly": 1, "bi-monthly": 2, "quarterly": 3, "semi-annually": 6, "annually": 0}
	invTotal              float64
	invTime               float64
	preProcFailedInvoices map[strfmt.UUID]int64
)

// DbParameter is the struct defined to group and contain all the methods
// that interact with the database.
// Parameters:
// - Cache: CacheManager pointer for the cache mechanism.
// - connStr: strings with the connection information to the database
// - Db: a gorm.DB pointer to the db to invoke all the db methods
type DbParameter struct {
	Cache       *cacheManager.CacheManager
	connStr     string
	Db          *gorm.DB
	Metrics     map[string]*prometheus.GaugeVec
	workersPool *pool
}

type period struct {
	from strfmt.DateTime
	to   strfmt.DateTime
}

// Workers Pool config struct
type pool struct {
	bufferSize int64
	metadata   chan map[string]interface{}
	poolActive bool
	results    chan workerResult
	sync       *sync.WaitGroup
	workers    int
	mutex      sync.RWMutex
}

type workerResult struct {
	amount       float64
	billrun      strfmt.UUID
	organization string
	status       string
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

	l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	//dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NewScope(&models.TABLE).TableName() + "', 'TIMESCALE-ROW-INDEX');")

	// Workers Pool
	p := pool{
		poolActive: false,
		bufferSize: 100,
	}

	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()

	if maxProcs < numCPU {

		p.workers = maxProcs

	} else {

		p.workers = numCPU
	}

	dp.workersPool = &p

	// Counter for clean billruns
	preProcFailedInvoices = make(map[strfmt.UUID]int64)

	return &dp

}

// createBillRun job is to create the skeleton of a new billrun in the system
// and provide its ID.
// Parameters:
// - size: int64 indicating the amount of invoices to be generated during the
// execution of the billrun.
// Returns:
// - id: a UUID string with the associated ID to the just created billrun.
// - e in case of any error happening.
func (d *DbParameter) createBillRun(size int64, executionType string) (id strfmt.UUID, e error) {

	l.Trace.Printf("[DB] Attempting to create a new billrun.\n")

	var o, origin models.BillRun

	status := "QUEUED"

	o = models.BillRun{
		CreationTimestamp: (strfmt.DateTime)(time.Now()),
		InvoicesCount:     size,
		Status:            &status,
		ExecutionType:     executionType,
	}

	if r := d.Db.Where(&o).First(&origin).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		if r := d.Db.Create(&o); r.Error == nil {

			l.Info.Printf("[DB] Billrun inserted successfully.\n")

			d.Metrics["count"].With(prometheus.Labels{"type": "Billruns created"}).Inc()

			id = r.Statement.Model.(*models.BillRun).ID

		} else {

			l.Warning.Printf("[DB] Something went wrong while trying to create a new billrun in the system. Error: %v.\n", r.Error)

			e = r.Error

		}

	} else {

		l.Warning.Printf("[DB] Billrun with id [ %v ] is already in the system, check with the administrator.", origin.ID)

		id = origin.ID
		e = errors.New("billrun duplication")

	}

	return

}

// createInvoice job is to create the skeleton of a new invoice in the system
// and provide its ID.
// Parameters:
// - p: a preiod object containing the timeframe for the invoice.
// - organization: a string contining the ID of the organization for the invoice.
// Returns:
// - id: a UUID string with the associated ID to the just created invoice.
// - e in case of any error happening.
// -
func (d *DbParameter) createInvoice(p period, organization, orgType string) (id strfmt.UUID, e error) {

	l.Trace.Printf("[DB] Attempting to create a new invoice.\n")

	var o, origin models.Invoice

	status := "NOT_PROCESSED"

	o = models.Invoice{
		Items:            make(datamodels.JSONdb),
		OrganizationID:   organization,
		OrganizationType: orgType,
		PeriodEndDate:    (strfmt.Date)(((time.Time)(p.to)).AddDate(0, 0, -1)),
		PeriodStartDate:  (strfmt.Date)(p.from),
		Status:           &status,
	}

	if r := d.Db.Where(&o).First(&origin).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		o.GenerationTimestamp = (strfmt.DateTime)(time.Now())
		status := "UNPAID"
		o.PaymentStatus = &status
		o.PaymentDeadline = (strfmt.Date)(time.Now().AddDate(0, 0, 15))

		if r := d.Db.Create(&o); r.Error == nil {

			l.Info.Printf("[DB] Invoice created successfully.\n")

			d.Metrics["count"].With(prometheus.Labels{"type": "Invoices created"}).Inc()

			id = r.Statement.Model.(*models.Invoice).ID

		} else {

			l.Warning.Printf("[DB] Something went wrong while trying to save the invoice in the system. Error: %v.\n", r.Error)

			e = r.Error

		}

	} else {

		l.Trace.Printf("[DB] Invoice with id [ %v ] is already in the system.", origin.ID)

		id = origin.ID
		e = errors.New("invoice already in the system")

	}

	return

}

// getFloat job is to check the type of the interface and properly cast its
// value into a float64.
// Parameters:
// - i: interface that should contain a float number.
// Returns:
// - f: a float64 with the value contained in the interface provided.
func (d *DbParameter) getFloat(i interface{}) (f float64) {

	var e error

	if i == nil {

		f = float64(0)

		return

	}

	if v := reflect.ValueOf(i); v.Kind() == reflect.Float64 {

		f = i.(float64)

	} else {

		f, e = i.(json.Number).Float64()

		if e != nil {

			l.Trace.Printf("[DB] GetFloat failed to convert [ %v ]. Error: %v\n", i, e)

		}

	}

	return

}

// getNiceFloat job is to turn the ugly float64 provided into a "nice looking"
// one according to the scale set, so we move from a floating coma number into
// a fixed coma number.
// Parameters:
// - i: the ugly floating coma number.
// Returns:
// - o: the "nice looking" fixed coma float.
func (d *DbParameter) getNiceFloat(i float64) (o float64) {

	min := float64(float64(1) / float64(scaler))

	if diff := math.Abs(i - min); diff < min {

		o = float64(0)

	} else {

		//o = float64(math.Round(i*scaler) / scaler)
		o = float64(i)

	}

	return

}

// contains job is to check if the string is contiained in the StringArray.
// Paramenters:
// - slice: the StringArray to check.
// - s: the probing string.
// Returns:
// - boolean with true if is contained
func (d *DbParameter) contains(slice pq.StringArray, s string) bool {

	for _, v := range slice {

		if v == s {

			return true

		}

	}

	return false

}

// GetBillRun job is to retrieve from the system the BillRunReport associated
// with the provided ID.
// Parameters:
// - id: a UUID string with the associated ID to the billrun requested.
// Returns:
// - the billrunreport associated to the provided ID.
// - e in case of any error happening.
func (d *DbParameter) GetBillRun(id strfmt.UUID) (*models.BillRunReport, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the billrun [ %v ].\n", id)

	var o models.BillRunReport
	var e error
	var oInvoices []*models.InvoiceMetadata

	if e = d.Db.Where(&models.BillRun{ID: id}).First(&models.BillRun{}).Scan(&o).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the billrun [ %v ]. Error: %v\n", id, e)

	} else {

		if e = d.Db.Where(&models.Invoice{BillRunID: id}).Find(&([]*models.Invoice{})).Scan(&oInvoices).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving the associated invoices to the billrun [ %v ]. Error: %v\n", id, e)

		} else {

			o.InvoicesList = oInvoices

			l.Debug.Printf("[DB] Billrun [ %v ] with its corresponding [ %v ] invoices retrieved from the system.\n", id, len(oInvoices))

		}

	}

	return &o, e

}

// GetInvoice job is to retrieve from the system the Invoice associated with
// the provided ID.
// Parameters:
// - id: a UUID string with the associated ID to the invoice requested.
// Returns:
// - the invoice associated to the provided ID.
// - e in case of any error happening.
func (d *DbParameter) GetInvoice(id strfmt.UUID) (*models.Invoice, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the invoice [ %v ].\n", id)

	var o models.Invoice
	var e error

	if e = d.Db.Where(&models.Invoice{ID: id}).First(&o).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the invoice [ %v ]. Error: %v\n", id, e)

	} else {

		l.Debug.Printf("[DB] Invoice [ %v ] retrieved from the system.\n", id)

	}

	return &o, e

}

// GetInvoicesByOrganization job is to retrieve from the system the existent
// invoice(s) related to the organization whose ID is provided.-
// Parameters:
// - id: a string with the organization ID associated to the invoice.
// - months: optional int64 with the amount of months to go back in time looking
// for the invoices associated.
// Returns:
// - o: slice of Invoice containend the requested invoices from the system.
// - e in case of any error happening.
func (d *DbParameter) GetInvoicesByOrganization(id string, months *int64) (o []*models.Invoice, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve the invoices from organization [ %v ].\n", id)

	m := int64(3)

	if months != nil {

		m = *months

	}

	timeWindow := d.getWindow("GenerationTimestamp", m)

	if e = d.Db.Where(timeWindow).Where(&models.Invoice{OrganizationID: id}).Find(&o).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the invoices for the organization [ %v ]. Error: %v\n", id, e)

	} else {

		l.Debug.Printf("[DB] Invoices for organization [ %v ] retrieved from the system.\n", id)

	}

	return

}

// getWindow job is to select the timeframe for the usage retrievals according
// to the data providad today-months<window. Default: today-3months<window
// Parameters:
// - col: string containing the name of the col used in the timeWindow.
// - months: optional int64 with the amount of months to go back in time.
// Returns:
// - s: a string with the needed window correctly formated to be used with GORM.
func (d *DbParameter) getWindow(col string, months int64) (s string) {

	l.Trace.Printf("[DB] Attempting to get a window for a db query.\n")

	if months <= 0 {

		s = fmt.Sprintf("%v > NOW() - INTERVAL '%v month'", d.Db.NamingStrategy.ColumnName("", col), 3)

	} else {

		s = fmt.Sprintf("%v > NOW() - INTERVAL '%v month'", d.Db.NamingStrategy.ColumnName("", col), months)

	}

	return s

}

// GenerateInvoicesBy job is to start the generation of the invoice(s) linked
// with the provided organiztion. Ideally this function will only be run with
// the idea of override the period of the invoice, but the simple re-run of the
// invoice possibility is present.
// Parameters:
// - org: a string with the type of organization: reseller or customer.
// - id: a string with the organization ID whose invoice is going to be generated.
// - token: a string with an optional keycloak bearer token.
// - from: optional datetime value to override the initial time of the invoicing
// period.
// - to: optional datetime value to override the final time of the invoicing
// period.
// Returns:
// - id: a UUID string with the associated ID of billrun in charge of the
// generation of the invoices.
// - e in case of any error happening.
func (d *DbParameter) GenerateInvoicesBy(org, id, token string, from, to strfmt.DateTime) (billrunID string, e error) {

	l.Trace.Printf("[DB] Starting the generation of invoices for customer [ %v ].\n", id)

	var billrun strfmt.UUID

	billrun, e = d.createBillRun(1, "ORGANIZATION: "+id)

	billrunID = (string)(billrun)

	if e != nil {

		return

	}

	metadata := make(map[string]interface{})
	metadata[org] = id
	metadata["billrun"] = billrun
	metadata["regenerate"] = true

	if !((time.Time)(from)).IsZero() && !((time.Time)(to)).IsZero() {

		metadata["period"] = period{
			from: from,
			to:   to,
		}

	}

	go d.InvoicesGeneration(metadata, token)

	return

}

// ListBillRuns job is to provide the list of billruns in the system, by default
// this list only goes back in time 3 months from today, but it can be overrided.
// Parameters:
// - months: optional int64 containing the amount of month to look back in time.
// Returns:
// - o: slice of BillRunList containing the billruns in the system.
// - e in case of any error happening.
func (d *DbParameter) ListBillRuns(months *int64) (o []*models.BillRunList, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the billruns in the system.\n")

	m := int64(3)

	if months != nil {

		m = *months

	}

	timeWindow := d.getWindow("CreationTimestamp", m)

	if e = d.Db.Where(timeWindow).Find(&([]*models.BillRun{})).Scan(&o).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the billruns in the system. Error: %v\n", e)

	} else {

		l.Debug.Printf("[DB] [ %v ] billruns retrieved from the system.\n", len(o))

	}

	return

}

// ListBillRunsByOrganization job is to provide the list of billruns in the system
// that are linked to the specific organization.
// By default this list only goes back in time 3 months from today, but it can
// be overrided.
// Parameters:
// - months: optional int64 containing the amount of month to look back in time.
// Returns:
// - o: slice of BillRunList containing the billruns in the system.
// - e in case of any error happening.
func (d *DbParameter) ListBillRunsByOrganization(organization, token string, months *int64) (o []*models.BillRunList, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the billruns in the system linked to the organization [ %v ].\n", organization)

	var o1, o2 []*models.BillRunList

	m := int64(3)

	executionTypeCus := "ORGANIZATION: "
	executionTypeRes := "ORGANIZATION: "

	c, e := d.Cache.Get(organization, "customer", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the customers. Assuming the id [ %v ] is a reseller. Error: %v\n", organization, e)

		executionTypeRes += organization
		executionTypeCus = ""

	} else {

		customer := c.(cusModels.Customer)

		executionTypeRes += customer.ResellerID
		executionTypeCus += organization

	}

	if months != nil {

		m = *months

	}

	timeWindow := d.getWindow("CreationTimestamp", m)

	if e1 := d.Db.Where(timeWindow).Where(&models.BillRun{ExecutionType: executionTypeCus}).Find(&([]*models.BillRun{})).Scan(&o1).Error; e1 != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the billruns in the system for [ %v ]. Error: %v\n", executionTypeCus, e1)

		e = fmt.Errorf("%v search gave problems. Error: %v", executionTypeCus, e1)

	}

	if e2 := d.Db.Where(timeWindow).Where(&models.BillRun{ExecutionType: executionTypeRes}).Find(&([]*models.BillRun{})).Scan(&o2).Error; e2 != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the billruns in the system for [ %v ]. Error: %v\n", executionTypeRes, e2)

		e = fmt.Errorf("%v search gave problems. Error: %v. Error chain: %v", executionTypeRes, e2, e)

	}

	o = append(o1, o2...)

	l.Debug.Printf("[DB] [ %v ] billruns retrieved from the system.\n", len(o))

	return

}

// ListInvoices job is to provide the list of invoices saved in the system.
// It admits the use of and invoice model to filter the results.
// Parameters:
// - model: optional Invoice model with data set to be used as a filter.
// Returns:
// - o: slice of Invoice containing the invoices matching the filter in the system.
// - e in case of any error happening.
func (d *DbParameter) ListInvoices(model *models.Invoice) (o []*models.Invoice, e error) {

	l.Trace.Printf("[DB] Attempting to retrieve all the invoices in the system.\n")

	if e = d.Db.Where(model).Find(&o).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the invoices from the system. Error: %v\n", e)

	} else {

		l.Debug.Printf("[DB] [ %v ] invoices retrieved from the system.\n", len(o))

	}

	return

}

// getPeriods job is to generate all the possibles invoicing periods that the
// system could be able to safely generate and invoice for based on the day
// marked by today, being the default behaviour to consider today as time.Now().
// Parameters:
// - today: an optional time.Time contained an overrider date.
// Returns:
// - periods: a map[string]period containing all the possibles periods for
// invoicing that can be used considering today as the actual day
// - e in case of any error happening.
func (d *DbParameter) getPeriods(today time.Time) (periods map[string]period) {

	l.Trace.Printf("[DB] Getting the time periods windows for the autonomous invoices.\n")

	now := time.Now()
	periods = make(map[string]period)

	if !today.IsZero() {

		now = today

	}

	for id, v := range periodicity {

		var end, start time.Time

		switch id {

		case "daily":

			end = now
			start = end.AddDate(0, 0, -v)

		case "weekly", "bi-weekly":

			multiplier := int(now.Weekday()) - 1

			if multiplier < 0 {

				multiplier = 6

			}

			end = now.AddDate(0, 0, -multiplier)
			start = end.AddDate(0, 0, -v)

		case "monthly", "bi-monthly", "quarterly", "semi-annually":

			if int(now.Month()) < (v + 1) {

				month := 12 - v + 1

				start = time.Date(now.Year()-1, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
				end = time.Date(now.Year(), time.Month(1), 1, 0, 0, 0, 0, time.UTC)

			} else {

				month := ((int(now.Month()) - 1) % v) * v

				start = time.Date(now.Year(), time.Month(month), 1, 0, 0, 0, 0, time.UTC)
				end = time.Date(now.Year(), time.Month(month+v+1), 1, 0, 0, 0, 0, time.UTC)

			}

		case "annually":

			start = time.Date(now.Year()-1, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(now.Year(), time.Month(1), 1, 0, 0, 0, 0, time.UTC)

		}

		f := strfmt.DateTime(time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC))
		t := strfmt.DateTime(time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.UTC))

		periods[id] = period{from: f, to: t}

	}

	return

}

// InvoicesGeneration job is to initiate a billrun process.
// Specifically its function is to automatically retrieve every reseller and
// customer in the system with their corresponding linked poroducts, select the
// adecuate time for its invoice and start the pool of workers and feed it.
// It also includes the option to override several parts and be used as a way to
// re-run and/or run specific invoice generations.
// Parameters:
// - metadata: the map[string]interface channel used to send the information
// needed for the invoice generation process.
// - token: a string with an optional keycloak bearer token.
func (d *DbParameter) InvoicesGeneration(metadata map[string]interface{}, token string) {

	var today time.Time
	var wg *sync.WaitGroup
	var pipe chan<- map[string]interface{}

	if t, exists := metadata["today"]; exists || metadata == nil {

		l.Trace.Printf("[DB] Starting the autonomous invoice generation pre-processing.\n")

		if t != nil {

			today = (time.Time)(t.(strfmt.DateTime))

		}

		// 1) Generate periods of time according to date and generate a billrun for each
		periods := d.getPeriods(today)

		// 2) List all resellers and customers
		r, e := d.Cache.Get("ALL", "reseller", token)

		if e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving the resellers. Error: %v\n", e)

			return

		}

		resellers := r.([]*cusModels.Reseller)

		c, e := d.Cache.Get("ALL", "customer", token)

		if e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving the customers. Error: %v\n", e)

			return

		}

		customers := c.([]*cusModels.Customer)

		// 3) Separate them by invoicing periods adding the dates from before

		resellersByPeriod := make(map[string][]string)
		cdrByReseller := make(map[string][]string)

		for i := range resellers {

			if !(*resellers[i].Billable) {

				continue

			}

			resellersByPeriod[*resellers[i].BillPeriod] = append(resellersByPeriod[*resellers[i].BillPeriod], resellers[i].ResellerID)

			r, e := d.Cache.Get(resellers[i].ResellerID, "reseller", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the reseller [ %v ]. Error: %v\n", resellers[i].ResellerID, e)

				return

			}

			reseller := r.(cusModels.Reseller)

			for j := range reseller.Customers {

				for k := range reseller.Customers[j].Products {

					cdrByReseller[resellers[i].ResellerID] = append(cdrByReseller[resellers[i].ResellerID], reseller.Customers[j].Products[k].ProductID)

				}

			}

		}

		customersByPeriod := make(map[string][]string)
		cdrByCustomer := make(map[string][]string)

		for i := range customers {

			if !(*customers[i].Billable) {

				continue

			}

			customersByPeriod[*customers[i].BillPeriod] = append(customersByPeriod[*customers[i].BillPeriod], customers[i].CustomerID)

			c, e := d.Cache.Get(customers[i].CustomerID, "customer", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the customer [ %v ]. Error: %v\n", customers[i].CustomerID, e)

				return

			}

			customer := c.(cusModels.Customer)

			for j := range customer.Products {

				cdrByCustomer[customers[i].CustomerID] = append(cdrByCustomer[customers[i].CustomerID], customer.Products[j].ProductID)

			}

		}

		// 4) invocation of ProcessInvoice with each reseller generating the metadata needed

		billrunSize := int64(0)

		for _, n := range resellersByPeriod {

			billrunSize += int64(len(n))

		}

		for _, n := range customersByPeriod {

			billrunSize += int64(len(n))

		}

		billrunID, e := d.createBillRun(billrunSize, "System-wide autonomous generation")

		if e != nil {

			if billrunID != "" {

				l.Warning.Printf("[DB] Something unexpected is actually happening... Will keep working as if nothing.Error: %v\n", e)

			} else {

				l.Warning.Printf("[DB] Something wrong happened when creating a new billrun, check with the administrator. Error: %v\n", e)

				return

			}

		}

		wg, pipe = d.startPool(billrunSize)

		for i, p := range periods {

			for _, res := range resellersByPeriod[i] {

				wg.Add(1)

				invoice, e := d.createInvoice(p, res, "reseller")

				if e != nil {

					l.Warning.Printf("[DB] Something wrong happened when creating a new invoice for org [ %v ], check with the administrator. Error: %v\n", res, e)

					wg.Done()
					preProcFailedInvoices[billrunID]++

					continue

				}

				md := make(map[string]interface{})

				md["type"] = "reseller"
				md["organization"] = res
				md["billrun"] = billrunID
				md["period"] = p
				md["cdr"] = strings.Join(cdrByReseller[res], ",")
				md["token"] = token
				md["invoice"] = invoice

				pipe <- md

			}

			for _, cus := range customersByPeriod[i] {

				wg.Add(1)

				invoice, e := d.createInvoice(p, cus, "customer")

				if e != nil {

					l.Warning.Printf("[DB] Something wrong happened when creating a new invoice for org [ %v ], check with the administrator. Error: %v\n", cus, e)

					wg.Done()
					preProcFailedInvoices[billrunID]++

					continue

				}

				md := make(map[string]interface{})

				md["type"] = "customer"
				md["organization"] = cus
				md["billrun"] = billrunID
				md["period"] = p
				md["cdr"] = strings.Join(cdrByCustomer[cus], ",")
				md["token"] = token
				md["invoice"] = invoice

				pipe <- md

			}

		}

	} else {

		l.Trace.Printf("[DB] Starting the invoice generation pre-processing for specific reseller/customer.\n")

		var periods map[string]period
		var customers []string
		var today time.Time
		var cdrs []string
		var window period
		var ty string
		var e error

		cuscdrs := make(map[string][]string)

		periods = d.getPeriods(today)

		if org, exists := metadata["reseller"]; exists {

			r, e := d.Cache.Get(org, "reseller", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the reseller. Error: %v\n", e)

				return

			}

			o := r.(cusModels.Reseller)
			window = periods[*o.BillPeriod]
			ty = "reseller"

			for i := range o.Customers {

				customers = append(customers, o.Customers[i].CustomerID)
				var cuscdr []string

				for j := range o.Customers[i].Products {

					cdrs = append(cdrs, o.Customers[i].Products[j].ProductID)

					cuscdr = append(cuscdr, o.Customers[i].Products[j].ProductID)

				}

				cuscdrs[o.Customers[i].CustomerID] = cuscdr

			}

		}

		if org, exists := metadata["customer"]; exists {

			c, e := d.Cache.Get(org, "customer", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the customer. Error: %v\n", e)

				return

			}

			o := c.(cusModels.Customer)
			window = periods[*o.BillPeriod]
			ty = "customer"

			for i := range o.Products {

				cdrs = append(cdrs, o.Products[i].ProductID)

			}

		}

		if value, exists := metadata["period"]; exists {

			window = value.(period)

		}

		invoice, e := d.createInvoice(window, metadata[ty].(string), ty)

		if e != nil || invoice == "" {

			l.Warning.Printf("[DB] Something wrong happened when creating a new invoice for org [ %v ], check with the administrator. Error: %v\n", metadata[ty], e)

			return

		}

		wg, pipe = d.startPool(int64(1 + len(customers)))

		wg.Add(1)

		md := make(map[string]interface{})

		md["type"] = ty
		md["organization"] = metadata[ty]
		md["billrun"] = metadata["billrun"]
		md["period"] = window
		md["cdr"] = strings.Join(cdrs, ",")
		md["token"] = token
		md["invoice"] = invoice

		pipe <- md

		if len(customers) != 0 {

			b := models.BillRun{
				ID:            metadata["billrun"].(strfmt.UUID),
				InvoicesCount: int64(1 + len(customers)),
			}

			if e = d.updateBillrun(b); e != nil {

				l.Warning.Printf("[DB] The invoices count of the billrun [ %v ] couldn't be updated. Error: %v.\n", metadata["billrun"], e)

			}

			for _, cus := range customers {

				invoice, e := d.createInvoice(window, cus, "customer")

				if e != nil && invoice == "" {

					l.Warning.Printf("[DB] Something wrong happened when creating a new invoice for org [ %v ], check with the administrator. Error: %v\n", cus, e)

					continue

				}

				wg.Add(1)

				md := make(map[string]interface{})

				md["type"] = "customer"
				md["organization"] = cus
				md["billrun"] = metadata["billrun"]
				md["period"] = window
				md["cdr"] = strings.Join(cuscdrs[cus], ",")
				md["token"] = token
				md["invoice"] = invoice

				pipe <- md

			}

		}

	}

	l.Trace.Printf("[DB] Invoice generation pre-processing completed.\n")

	wg.Done()

	return

}

// startPool job is to initiate the pool of workers and provide to the invoker
// with the associated waitgroup and the feed channel to add task to the pool.
// Parameters:
// - buffer: the expected size of taks the pool will need to handle.
// Returns:
// - the WaitGroup linked to the pool of workers.
// - the map[string]interface channel used to send the information needed for
// the invoice generation process.
func (d *DbParameter) startPool(buffer int64) (*sync.WaitGroup, chan<- map[string]interface{}) {

	l.Trace.Printf("[DB][WorkersPool] Pool start invoked!.\n")

	if d.workersPool.bufferSize < buffer {

		d.workersPool.bufferSize = buffer

	}

	if !d.workersPool.poolActive {

		l.Trace.Printf("[DB][WorkersPool] Pool inactive, generating one.\n")

		d.workersPool.mutex.Lock()

		var wg sync.WaitGroup

		d.workersPool.poolActive = true
		d.workersPool.sync = &wg
		d.workersPool.results = make(chan workerResult, d.workersPool.bufferSize)
		d.workersPool.metadata = make(chan map[string]interface{}, d.workersPool.bufferSize)

		d.workersPool.mutex.Unlock()

		for i := 0; i < d.workersPool.workers; i++ {

			l.Trace.Printf("[DB][WorkersPool] Starting pool worker #%v.\n", i)

			go d.invoiceProcessor(i, d.workersPool.metadata, d.workersPool.results)

		}

		go d.poolResults(d.workersPool.sync, d.workersPool.results)

		wg.Add(1)

		go d.endPool(d.workersPool.sync, d.workersPool.metadata, d.workersPool.results)

	}

	return d.workersPool.sync, d.workersPool.metadata

}

// poolResults job is to process the results of the concurrent invoice
// generation processes and update the linked billrun accordingly.
// Parameters:
// - wg: the WaitGroup linked to the pool of workers.
// - results: a workerResult channel to receive the results of the invoice
// generation process.
func (d *DbParameter) poolResults(wg *sync.WaitGroup, results <-chan workerResult) {

	l.Trace.Printf("[DB][Worker] Results processing worker started.\n")

	for r := range results {

		b := models.BillRun{
			ID:                    r.billrun,
			AmountInvoiced:        r.amount,
			OrganizationsInvolved: pq.StringArray([]string{r.organization}),
			Status:                &r.status,
		}

		if e := d.updateBillrun(b); e != nil {

			l.Warning.Printf("[DB][Worker] Billrun [ %v ][ %v ] couldn't be updated, check with the administrator. Error: %v\n", r.billrun, r.organization, e)

		} else {

			l.Trace.Printf("[DB][Worker] Billrun [ %v ] updated successfully.\n", r.billrun)

		}

		wg.Done()

	}

	l.Trace.Printf("[DB][Worker] Results channel closed, worker task completed.\n")

	return

}

// endPool job is to wait until all the information in the metadata channel
// has been consummed by the workers and then close the channels used by the
// pool of workers.
// Parameters:
// - wg: the WaitGroup linked to the pool of workers.
// - metadata: a map[string]interface channel used to send the information
// needed for the invoice generation process.
// - results: a workerResult channel to report the results of the invoice
// generation process.
func (d *DbParameter) endPool(wg *sync.WaitGroup, metadata chan map[string]interface{}, results chan workerResult) {

	l.Trace.Printf("[DB][WorkersPool] Clean pool mechanism invoked and waiting for workers to finish.\n")

	wg.Wait()

	d.workersPool.mutex.Lock()

	d.workersPool.poolActive = false
	close(metadata)
	close(results)

	d.workersPool.mutex.Unlock()

	l.Trace.Printf("[DB][WorkersPool] Workers finished, pool cleaning process completed.\n")

	return

}

// invoiceProcessor is the core-function of the workers in the pool, its job
// is to generate an invoice based on the information provided based on the
// information provided via the metadata channel and provide the reults of
// the generation back via the resutls channel.
// Parameters:
// - metadata: a map[string]interface channel to receive objects containing all
// the data needed to generate an invoice.
// - results: a workerResult channel to report the end result of the processing.
func (d *DbParameter) invoiceProcessor(worker int, metadata <-chan map[string]interface{}, results chan<- workerResult) {

	l.Trace.Printf("[DB][Worker #%v] Starting an invoce processing worker.\n", worker)

	// meta: org, type, cdrs, billrun, period, regenerate(bool)
MainLoop:
	for m := range metadata {

		l.Trace.Printf("[DB][Worker #%v] Starting processing of organization [ %v ].\n", worker, m["organization"])

		d.Metrics["count"].With(prometheus.Labels{"type": "Invoice processing started"}).Inc()
		countTime := time.Now().UnixNano()

		var invoice models.Invoice
		var items []datamodels.JSONdb

		discount := float64(0)
		invoiceTotal := float64(0)
		invoice.Items = make(datamodels.JSONdb)

		// 0) put everything in vars
		token := m["token"].(string)
		orgID := m["organization"].(string)
		orgType := m["type"].(string)
		orgCDRs := m["cdr"].(string)
		billrunID := m["billrun"].(strfmt.UUID)
		invoicePeriod := m["period"].(period)
		invoiceID := m["invoice"].(strfmt.UUID)

		resultERROR := workerResult{
			amount:       float64(0),
			billrun:      billrunID,
			organization: orgID,
			status:       "ERROR",
		}

		// 1) invoice to processing
		state := "PROCESSING"
		invoice.ID = invoiceID
		invoice.BillRunID = billrunID
		invoice.OrganizationID = orgID
		invoice.Status = &state

		// 6) save/update the complete invoice
		if e := d.updateInvoice(invoice); e != nil {

			l.Warning.Printf("[DB][Worker #%v] Couldn't update the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

			results <- resultERROR

			if e = d.errorInvoice(invoiceID); e != nil {

				l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

			}

			continue MainLoop

		}

		// 2) get the CDRs and skus
		var CDRs []*cdrModels.CReport

		for _, o := range strings.Split(orgCDRs, ",") {

			id := fmt.Sprintf("%v?%v?%v", o, invoicePeriod.from, invoicePeriod.to)

			cdr, e := d.Cache.Get(id, "cdr", token)

			if e != nil {

				l.Warning.Printf("[DB][Worker #%v] Something went wrong while retrieving the associated CDRs. Error: %v\n", worker, e)

				results <- resultERROR

				if e = d.errorInvoice(invoiceID); e != nil {

					l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

				}

				continue MainLoop

			}

			CDRs = append(CDRs, cdr.([]*cdrModels.CReport)...)

		}

		s, e := d.Cache.Get("ALL", "sku", token)

		if e != nil {

			l.Warning.Printf("[DB][Worker #%v] Something went wrong while retrieving the sku list. Error: %v\n", worker, e)

			results <- resultERROR

			if e = d.errorInvoice(invoiceID); e != nil {

				l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

			}

			continue MainLoop

		}

		SKUs := make(map[string]string)

		for _, k := range s.([]*pmModels.Sku) {

			SKUs[*k.Name] = k.ID

		}

		// 3) get the details of the organization and fill the invoice
		if orgType == "reseller" {

			r, e := d.Cache.Get(orgID, "reseller", token)

			if e != nil {

				l.Warning.Printf("[DB][Worker #%v] Something went wrong while retrieving the reseller. Error: %v\n", worker, e)

				results <- resultERROR

				if e = d.errorInvoice(invoiceID); e != nil {

					l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

				}

				continue MainLoop

			}

			o := r.(cusModels.Reseller)

			invoice.BillingContact = o.BillContact
			invoice.Currency = o.BillCurrency
			invoice.OrganizationName = o.Name
			discount = o.Discount

		} else {

			c, e := d.Cache.Get(orgID, "customer", token)

			if e != nil {

				l.Warning.Printf("[DB][Worker #%v] Something went wrong while retrieving the customer. Error: %v\n", worker, e)

				results <- resultERROR

				if e = d.errorInvoice(invoiceID); e != nil {

					l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

				}

				continue MainLoop

			}

			o := c.(cusModels.Customer)

			invoice.BillingContact = o.BillContact
			invoice.Currency = o.BillCurrency
			invoice.OrganizationName = o.Name
			discount = o.Discount

		}

		// 4) process the CDRs
		// we split the load by product/account
		for _, product := range strings.Split(orgCDRs, ",") {

			if product == "" {

				continue

			}

			var costBreakup []datamodels.JSONdb

			grossCost := float64(0)
			acc := make(datamodels.JSONdb)

			acc["ID"] = product
			acc["discountRate"] = discount
			custId, custName, e := d.getCustomerData(product, token)

			if e != nil {

				l.Warning.Printf("[DB][Worker #%v] Something went wrong while retrieving the customer data linked to product [ %v ]. Error: %v\n", worker, product, e)

				results <- resultERROR

				if e = d.errorInvoice(invoiceID); e != nil {

					l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

				}

				continue MainLoop

			}

			acc["customerID"] = custId
			acc["customerName"] = custName

			// let's go sku by sku
			for kind := range SKUs {

				var resourceList []datamodels.JSONdb
				skudata := make(datamodels.JSONdb)
				cost := make(datamodels.JSONdb)

				skudata["skuName"] = kind
				skudata["skuCost"] = float64(0)
				skudata["skuAggregatedDiscount"] = float64(0)
				skudata["skuNet"] = float64(0)

				// go through the cdrs looking for the product id
				for i := range CDRs {

					if CDRs[i].AccountID != product {

						continue

					}

					usage := CDRs[i].Usage

				UsageLoop:
					// once found let's go across the usage
					for j := range usage {

						// if not the type, skip
						if usage[j].ResourceType != kind {

							// if it's server, let's check in detail
							if usage[j].ResourceType != "server" {

								continue

							}

						}

						data := usage[j]
						resource := make(datamodels.JSONdb)

						// In case the resource doesn't have id/name let's use the sku to identify it somehow
						if data.ResourceID != "" {

							resource["resourceID"] = data.ResourceID

						} else {

							resource["resourceID"] = kind

						}

						if data.ResourceName != "" {

							resource["resourceName"] = data.ResourceName

						} else {

							resource["resourceName"] = kind

						}

						resource["metadata"] = data.Metadata
						resource["usage"] = data.UsageBreakup
						resource["skuUnits"] = d.getFloat(float64(1) / float64(3))

						skuCostBreakup := make(datamodels.JSONdb)

						// Update of costs
						if data.Cost["costBreakup"] != nil {

							for _, value := range data.Cost["costBreakup"].([]interface{}) {

								k := value.(map[string]interface{})

								if k["sku"] != kind {

									continue

								}

								skudata["skuNet"] = d.getFloat(skudata["skuNet"]) + d.getFloat(k["sku-net"])
								skudata["skuCost"] = d.getFloat(skudata["skuCost"]) + d.getFloat(k["sku-cost"])
								skuCostBreakup[k["sku-state"].(string)] = d.getFloat(skuCostBreakup[k["sku-state"].(string)]) + d.getFloat(k["sku-net"])

							}

						}

						resource["costBreakup"] = skuCostBreakup

						// if sku not found, skip
						if len(skuCostBreakup) == 0 {

							continue

						}

						// resourceList update...
						for i, duplicated := range resourceList {

							// if the resource is already in the list, create a newone = oldOne + newData
							if duplicated["resourceID"].(string) == resource["resourceID"].(string) &&
								duplicated["resourceName"].(string) == resource["resourceName"].(string) &&
								fmt.Sprintf("%v", duplicated["metadata"]) == fmt.Sprintf("%v", resource["metadata"]) {

								newResource := make(datamodels.JSONdb)

								if data.ResourceID != "" {

									newResource["resourceID"] = data.ResourceID

								} else {

									newResource["resourceID"] = kind

								}

								if data.ResourceName != "" {

									newResource["resourceName"] = data.ResourceName

								} else {

									newResource["resourceName"] = kind

								}

								newResource["metadata"] = data.Metadata
								newResource["skuUnits"] = d.getFloat(float64(1)/float64(3)) + d.getFloat(duplicated["skuUnits"])

								newUse := make(datamodels.JSONdb)
								newCost := make(datamodels.JSONdb)

								oldCost := duplicated["costBreakup"].(datamodels.JSONdb)
								oldUse := duplicated["usage"].(datamodels.JSONdb)

								for x, v := range data.UsageBreakup {

									if oUse, exists := oldUse[x]; exists {

										newUse[x] = d.getFloat(oUse) + d.getFloat(v)

									} else {

										newUse[x] = d.getFloat(v)

									}

								}

								for x, v := range skuCostBreakup {

									if oCost, exists := oldCost[x]; exists {

										newCost[x] = d.getFloat(oCost) + d.getFloat(v)

									} else {

										newUse[x] = d.getFloat(v)

									}

								}

								newResource["usage"] = newUse
								newResource["costBreakup"] = newCost

								resourceList[i] = newResource

								continue UsageLoop

							}

						}

						resourceList = append(resourceList, resource)

					}

				}

				if d.getFloat(skudata["skuNet"]) == float64(0) {

					continue

				}

				skudata["skuAggregatedDiscount"] = d.getNiceFloat(d.getFloat(skudata["skuCost"]) - d.getFloat(skudata["skuNet"]))

				cost["sku"] = skudata
				cost["resourceList"] = resourceList

				grossCost = grossCost + d.getFloat(skudata["skuNet"])
				costBreakup = append(costBreakup, cost)

			}

			acc["grossCost"] = d.getNiceFloat(grossCost)
			acc["discount"] = d.getNiceFloat(grossCost * discount)
			acc["netCost"] = d.getNiceFloat(grossCost - (grossCost * discount))
			acc["costBreakup"] = costBreakup

			items = append(items, acc)

			invoiceTotal = invoiceTotal + d.getNiceFloat(grossCost-(grossCost*discount))

		}

		// 5) complete the invoice
		invoice.AmountInvoiced = invoiceTotal
		invoice.Items["accounts"] = items

		state = "FINISHED"
		invoice.Status = &state

		// 6) save/update the complete invoice
		if e := d.updateInvoice(invoice); e != nil {

			l.Warning.Printf("[DB][Worker #%v] Couldn't save the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

			results <- resultERROR

			if e = d.errorInvoice(invoiceID); e != nil {

				l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

			}

			continue MainLoop

		}

		// 7) update the billrun
		// The invoices with 0 are ok now
		//	if invoiceTotal == float64(0) {

		//		l.Warning.Printf("[DB][Worker #%v] Invoice [ %v ] has invoiced 0, discarding...", worker, invoiceID)

		//		results <- resultERROR

		//		if e = d.errorInvoice(invoiceID); e != nil {

		//			l.Warning.Printf("[DB][Worker #%v] Couldn't set as ERROR the invoice [ %v ] for organization [ %v ], check with the administrator. Error: %v\n", worker, invoiceID, orgID, e)

		//		}

		//	} else {

		results <- workerResult{
			amount:       invoiceTotal,
			billrun:      billrunID,
			organization: orgID,
			status:       "FINISHED",
		}

		//	}

		l.Trace.Printf("[DB][Worker #%v] Finished processing of organization [ %v ].\n", worker, m["organization"])

		d.Metrics["count"].With(prometheus.Labels{"type": "Invoices processing completed"}).Inc()

		invTotal++
		invTime += float64(time.Now().UnixNano() - countTime)

		d.Metrics["time"].With(prometheus.Labels{"type": "Invoice generation average time"}).Set(invTime / invTotal / float64(time.Millisecond))

	}

	l.Trace.Printf("[DB][Worker #%v] No more data to be processed, stopping worker.\n", worker)

	return

}

// getCustomerData job is to retrieve the id and name associated to the customer
// whose product id is provided.
// Parameters:
// - p: string with the product id.
// - token: a string with an optional keycloak bearer token.
// Returns:
// - id: string with the linked customer id.
// - name: string with the linked customer name.
func (d *DbParameter) getCustomerData(p, token string) (id, name string, err error) {

	x, e := d.Cache.Get(p, "product", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the product info for id [ %v ]. Error: %v\n", p, e)

		err = e

		return

	}

	product := x.(cusModels.Product)

	c, e := d.Cache.Get(product.CustomerID, "customer", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the customer info for id [ %v ]. Error: %v\n", product.CustomerID, e)

		err = e

		return

	}

	customer := c.(cusModels.Customer)

	id = customer.CustomerID
	name = customer.Name

	return

}

// updateBillrun job is to update the information of the billrun saved in
// the system with the information provided in the incoming billrun object.
// Parameters:
// - o: billrun object with the date to update in the system.
// Returns:
// - e in case of any error happening.
func (d *DbParameter) updateBillrun(o models.BillRun) (e error) {

	l.Trace.Printf("[DB] Attempting to update the billrun [ %v ].\n", o.ID)

	var origin models.BillRun

	r := d.Db.Where(&models.BillRun{ID: o.ID}).First(&origin).Error

	if errors.Is(r, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Something went wrong, billrun [ %v ] not found in the system.\n", o.ID)

		e = errors.New("billrun not found")

		return

	}

	if r != nil {

		l.Warning.Printf("[DB] Something went wrong. Error: %v\n", r)

		e = r

	} else {

		// The update itself
		if o.Status != nil && *o.Status == "ERROR" {

			if len(o.OrganizationsInvolved) > 0 && !d.contains(origin.InvoicesErrorList, o.OrganizationsInvolved[0]) {

				o.InvoicesErrorList = append(origin.InvoicesErrorList, o.OrganizationsInvolved[0])

			} else {

				o.InvoicesErrorList = origin.InvoicesErrorList

			}

			o.InvoicesErrorCount = int64(len(o.InvoicesErrorList))

		} else {

			status := "PROCESSING"
			o.Status = &status

		}

		if len(o.OrganizationsInvolved) > 0 && !d.contains(origin.OrganizationsInvolved, o.OrganizationsInvolved[0]) {

			o.OrganizationsInvolved = append(origin.OrganizationsInvolved, o.OrganizationsInvolved[0])

		} else {

			o.OrganizationsInvolved = origin.OrganizationsInvolved

		}

		o.InvoicesProcessedCount = int64(len(o.OrganizationsInvolved))
		o.AmountInvoiced += origin.AmountInvoiced

		if (origin.InvoicesCount-preProcFailedInvoices[o.ID] <= o.InvoicesProcessedCount) && (o.Status != nil && (*o.Status != "ERROR" || o.InvoicesErrorCount <= 0)) {

			d.Metrics["count"].With(prometheus.Labels{"type": "Billruns completed"}).Inc()

			status := "FINISHED"
			o.Status = &status

		}

		if e = d.Db.Model(origin).Updates(o).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while updating the billrun [ %v ]. Error: %v\n", o.ID, e)

		} else {

			l.Trace.Printf("[DB] Billrun [ %v ] updated successfully.\n", o.ID)

		}

	}

	return
}

// updateInvoice job is to update the information of the invoice saved in
// the system with the information provided in the incoming invoice object.
// Parameters:
// - o: invoice object with the date to update in the system.
// Returns:
// - e in case of any error happening.
func (d *DbParameter) updateInvoice(o models.Invoice) (e error) {

	l.Trace.Printf("[DB] Attempting to update the invoice [ %v ].\n", o.ID)

	var origin models.Invoice

	r := d.Db.Where(&models.Invoice{ID: o.ID}).First(&origin).Error

	if errors.Is(r, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Something went wrong, invoice [ %v ] not found in the system.\n", o.ID)

		e = errors.New("invoice not found")

		return

	}

	if r != nil {

		l.Warning.Printf("[DB] Something went wrong. Error: %v\n", r)

		e = errors.New("invoice not found")

	} else {

		if e = d.Db.Model(origin).Updates(o).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while updating the invoice [ %v ]. Error: %v\n", o.ID, e)

		} else {

			l.Trace.Printf("[DB] Invoice [ %v ] updated successfully.\n", o.ID)

		}

	}

	return
}

// errorInvoice job is to mark the invoice linked to the provided ID
// with ERROR status.
// Parameters:
// - id: a UUID string with the associated ID to the invoice to ERROR.
// Returns:
// - e in case of any error happening.
func (d *DbParameter) errorInvoice(id strfmt.UUID) (e error) {

	l.Trace.Printf("[DB] Attempting to delete de failed invoice [ %v ].\n", id)

	var origin models.Invoice

	r := d.Db.Where(&models.Invoice{ID: id}).First(&origin).Error

	if errors.Is(r, gorm.ErrRecordNotFound) {

		l.Warning.Printf("[DB] Something went wrong, invoice [ %v ] not found in the system.\n", id)

		e = errors.New("invoice not found")

		return

	}

	if r != nil {

		l.Warning.Printf("[DB] Something went wrong. Error: %v\n", r)

		e = errors.New("invoice not found")

	} else {

		status := "ERROR"

		if e = d.Db.Model(origin).Updates(&models.Invoice{Status: &status}).Error; e != nil {

			l.Warning.Printf("[DB] Something went wrong while ERRORing the invoice [ %v ]. Error: %v\n", id, e)

		} else {

			l.Trace.Printf("[DB] Invoice [ %v ] ERRORed successfully.\n", id)

		}

	}

	return
}

// getCDRS job is to retrieve the list of products linked to an organization ID
// double checking if it's a reseller or a customer, providing this information
// also with the retrieved list.
// Parameters:
// - orgID: a string with the organization ID whose product list is going
// to be retrieved.
// - token: a string with an optional keycloak bearer token.
// Returns:
// - ty: string containing the type of organization
// - cdrs: string containing a coma-separated list with the IDs of the linked
// products to the specified organization.
// - e in case of any error happening.
func (d *DbParameter) getCDRS(orgID, token string) (ty, cdrs string, err error) {

	l.Trace.Printf("[DB] Attempting to get the CDRs for the organization [ %v ].\n", orgID)

	var customers []*cusModels.Customer

	ty = "reseller"

	r, e := d.Cache.Get(orgID, ty, token)

	if e != nil {

		l.Trace.Printf("[DB] Something went wrong while retrieving de organization [ %v ] as a reseller, trying with it as a customer... Error: %v\n", orgID, e)

		ty = "customer"

		c, e := d.Cache.Get(orgID, ty, token)

		if e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving de organization [ %v ] as a customer. Error: %v\n", orgID, e)

			err = e

			return

		}

		customer := c.(cusModels.Customer)

		customers = append(customers, &customer)

	} else {

		reseller := r.(cusModels.Reseller)

		customers = append(customers, reseller.Customers...)

	}

	for i := range customers {

		customer := *customers[i]

		for j := range customer.Products {

			if cdrs != "" {

				cdrs = string(customer.Products[j].ProductID)

			} else {

				cdrs = cdrs + "," + string(customer.Products[j].ProductID)

			}

		}

	}

	return

}

// ReRunBillRun processes the billruns in the system and re-run them in order
// to try to fix the ones that failed.
// Parameters:
// - id: a UUID string with the associated ID to the billrun that should be check.
// - months: int64 indicating the amount of months to go back in time to check
// for failed billruns.
// - token: a string with an optional keycloak bearer token.
// Returns:
// - status: a int indicating the result of the reRun execution
// - e in case of any error happening.
func (d *DbParameter) ReRunBillRun(id strfmt.UUID, months *int64, token string) (status int, e error) {

	l.Trace.Printf("[DB] Attempting to re-run failed invoices on billrun [ %v ].\n", id)

	var pipe chan<- map[string]interface{}
	var invoices []*models.Invoice
	var wg *sync.WaitGroup
	var timeWindow string
	var f, t time.Time

	state := "FINISHED"

	if id == "" {

		m := int64(3)

		if months != nil {

			m = *months

		}

		timeWindow = d.getWindow("GenerationTimestamp", m)

	}

	r := d.Db.Where(timeWindow).Not(&models.Invoice{Status: &state}).Where(&models.Invoice{BillRunID: id}).Find(&invoices).Error

	if errors.Is(r, gorm.ErrRecordNotFound) {

		status = StatusMissing

		e = errors.New("there's nothing to re-run")

		return

	}

	if r != nil {

		status = StatusFail

		e = r

	} else {

		status = StatusOK

		for _, invoice := range invoices {

			wg, pipe = d.startPool(int64(100))

			wg.Add(1)

			f = (time.Time)(invoice.PeriodStartDate)
			from := time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, time.UTC)

			t = (time.Time)(invoice.PeriodEndDate)
			to := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC)

			ty, cdrs, e := d.getCDRS(invoice.OrganizationID, token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving de cdrs associated to the organization [ %v ]. Error: %v\n", invoice.OrganizationID, e)

				continue

			}

			md := make(map[string]interface{})

			md["type"] = ty
			md["organization"] = invoice.OrganizationID
			md["billrun"] = invoice.BillRunID
			md["cdr"] = cdrs
			md["regenerate"] = true
			md["token"] = token
			md["invoice"] = invoice.ID
			md["period"] = period{
				from: (strfmt.DateTime)(from),
				to:   (strfmt.DateTime)(to),
			}

			pipe <- md

		}

		wg.Done()

	}

	return

}

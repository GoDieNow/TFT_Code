package dbManager

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	datamodels "gitlab.com/cyclops-utilities/datamodels"
	"github.com/GoDieNow/TFT_Code/services/cdr/models"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/cacheManager"
	cusModels "github.com/GoDieNow/TFT_Code/services/customerdb/models"
	pmModels "github.com/GoDieNow/TFT_Code/services/planmanager/models"
	udrModels "github.com/GoDieNow/TFT_Code/services/udr/models"
	l "gitlab.com/cyclops-utilities/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	scaler           = 1e9
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK
)

var (
	totalTime  float64
	totalCount int64
)

// DbParameter is the struct defined to group and contain all the methods
// that interact with the database.
// On it there is the following parameters:
// - Cache: CacheManager pointer for the cache mechanism.
// - connStr: strings with the connection information to the database.
// - Db: a gorm.DB pointer to the db to invoke all the db methods.
type DbParameter struct {
	Cache   *cacheManager.CacheManager
	connStr string
	Db      *gorm.DB
	Metrics map[string]*prometheus.GaugeVec
	Pipe    chan interface{}
}

// New is the function to create the struct DbParameter.
// Parameters:
// - dbConn: strings with the connection information to the database.
// - tables: array of interfaces that will contains the models to migrate
// to the database on initialization.
// Returns:
// - DbParameter: struct to interact with dbManager functionalities.
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

	//l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	//dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NewScope(&models.TABLE).TableName() + "', 'TIMESCALE-ROW-INDEX');")

	return &dp

}

// AddRecord job is to save in the system the provided CDR report to be easy to
// retrieve later with the usage endpoints.
// Parameters:
// - cdr: the CDR report model to be added to the system.
// Returns:
// - e: error in case of duplication or failure in the task.
func (d *DbParameter) AddRecord(cdr models.CReport) (e error) {

	l.Trace.Printf("[DB] Attempting to add a new cost report to the system.\n")

	if r := d.Db.Where(&cdr).First(&models.CReport{}).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		e = d.Db.Create(&cdr).Error

		if e != nil {

			l.Trace.Printf("[DB] Something went wrong when adding a new usage report to the system. Error: %v\n", e)

		} else {

			d.Metrics["count"].With(prometheus.Labels{"type": "CDR Reports added"}).Inc()

			l.Trace.Printf("[DB] New cost report added to the system.\n")

		}

	} else {

		l.Warning.Printf("[DB] Cost report already in the system, skipping creation...\n")

	}

	l.Trace.Printf("[DB] Attempting to add CDR Records from reported account [ %v ] to the system.\n", cdr.AccountID)

	for i := range cdr.Usage {

		var cdrrecord, cdrUpdate models.CDRRecord

		cdrrecord.AccountID = cdr.AccountID
		cdrrecord.Metadata = cdr.Usage[i].Metadata
		cdrrecord.ResourceID = cdr.Usage[i].ResourceID
		cdrrecord.ResourceName = cdr.Usage[i].ResourceName
		cdrrecord.ResourceType = cdr.Usage[i].ResourceType
		cdrrecord.TimeFrom = cdr.TimeFrom
		cdrrecord.TimeTo = cdr.TimeTo
		cdrrecord.Unit = cdr.Usage[i].Unit

		if r := d.Db.Where(&cdrrecord).First(&cdrUpdate).Error; errors.Is(r, gorm.ErrRecordNotFound) {

			cdrrecord.Cost = cdr.Usage[i].Cost
			cdrrecord.UsageBreakup = cdr.Usage[i].UsageBreakup

			if e = d.Db.Create(&cdrrecord).Error; e != nil {

				l.Warning.Printf("[DB] Something went wrong when adding a new CDR Record to the system. Error: %v\n", e)

			} else {

				d.Metrics["count"].With(prometheus.Labels{"type": "CDR Records added"}).Inc()

				l.Trace.Printf("[DB] New CDR Record added in the system.\n")

			}

		} else {

			l.Trace.Printf("[DB] CDR record already in the system, updating...\n")

			cdrrecord.Cost = cdr.Usage[i].Cost
			cdrrecord.UsageBreakup = cdr.Usage[i].UsageBreakup

			if e = d.Db.Model(&cdrUpdate).Updates(&cdrrecord).Error; e == nil {

				d.Metrics["count"].With(prometheus.Labels{"type": "CDR Records updated"}).Inc()

				l.Trace.Printf("[DB] CDR Record updated in the system.\n")

			} else {

				l.Trace.Printf("[DB] Something went wrong when updating the existing CDR Record to the system. Error: %v\n", e)

			}

		}

	}

	return

}

// GetReport job is to retrieve the compacted usage records from the system for
// the provided account in the requested time-window with the posibility of
// filter by metric.
// Parameters:
// - id: string containing the id of the account requested.
// - metric: a string with the metric to filter the records.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - a slice of references with the usage contained in the system within the
// requested time-window.
func (d *DbParameter) GetReport(id string, metric string, from strfmt.DateTime, to strfmt.DateTime) []*models.CDRReport {

	l.Trace.Printf("[DB] Attempting to retrieve the compacted usage records.\n")

	var r []*models.CDRReport
	var u models.CDRRecord

	window := d.getWindow(from, to)

	if id != "" {

		u.AccountID = id

	}

	if metric != "" {

		u.ResourceType = metric

	}

	if e := d.Db.Where(window).Where(&u).Find(&models.CDRRecord{}).Scan(&r).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the compacted usage records from the system. Error: %v\n", e)

	} else {

		l.Debug.Printf("[DB] [ %v ] compacted usage records retrieved from the system.\n", len(r))

	}

	return r

}

// GetUsage job is to retrieve the usage report of the system or the specified
// account within the requested time-window with the posibility to filter by
// metric.
// Parameters:
// - id: string containing the id of the account requested.
// - metric: a string with the metric to filter the records.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - a slice of references with the usage report contained in the system within
// the requested time-window.
// - error in case of duplication or failure in the task.
func (d *DbParameter) GetUsage(id, metric string, from, to strfmt.DateTime) ([]*models.CReport, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the usage report from the system.\n")

	var r []*models.CReport
	var r0 models.CReport
	var e error

	window := d.getWindow(from, to)

	if id != "" {

		r0.AccountID = id

	}

	e = d.Db.Where(window).Where(&r0).Find(&r).Error

	if e == nil {

		for idx := range r {

			r[idx].Usage = d.GetReport(r[idx].AccountID, metric, r[idx].TimeFrom, r[idx].TimeTo)

		}

		l.Debug.Printf("[DB] [ %v ] Usage reports retrieved from the system.\n", len(r))

	} else {

		l.Warning.Printf("[DB] Something went wrong while retrieving the usage reports from the system. Error: %v\n", e)

	}

	return r, e

}

// GetUsages job is to retrieve the usage report of the system or the specified
// account accounts within the requested time-window with the posibility to
// filter by metric.
// Parameters:
// - ids: string containing the list of ids of the accounts requested.
// - metric: a string with the metric to filter the records.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - a slice of references with the usage report contained in the system within
// the requested time-window.
// - error in case of duplication or failure in the task.
func (d *DbParameter) GetUsages(ids, metric string, from, to strfmt.DateTime) ([]*models.CReport, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the usage report from the system.\n")

	var total, r []*models.CReport
	var r0 models.CReport
	var e error

	window := d.getWindow(from, to)

	list := strings.Split(ids, ",")

	for _, acc := range list {

		r0.AccountID = acc

		if e = d.Db.Where(window).Where(&r0).Find(&r).Error; e == nil {

			for idx := range r {

				r[idx].Usage = d.GetReport(r[idx].AccountID, metric, r[idx].TimeFrom, r[idx].TimeTo)

			}

			l.Debug.Printf("[DB] [ %v ] Usage reports retrieved from the system.\n", len(r))

		} else {

			l.Warning.Printf("[DB] Something went wrong while retrieving the usage reports from the system. Error: %v\n", e)

		}

		total = append(total, r...)

	}

	return total, e

}

// GetUsageSummary job is to retrieve the usage report of the system of the specified
// account within the requested time-window and process it into an easy to use
// estructure for the UI.
// Parameters:
// - id: string containing the id of the account requested.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - a references with the usage report contained in the system within
// the requested time-window.
// - error in case of duplication or failure in the task.
func (d *DbParameter) GetUsageSummary(id string, from, to strfmt.DateTime) (*models.UISummary, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the usage report from the system.\n")

	var usages []*models.CDRReport
	var report models.UISummary

	interval := float64(28800) //UDR are created every 8h and the usage is provided in GB*s...
	rounder := float64(((time.Time)(to)).Sub((time.Time)(from)).Seconds()) / interval

	report.AccountID = id
	report.TimeFrom = from
	report.TimeTo = to

	// Get the reseller data
	r, err := d.Cache.Get(id, "reseller", "")

	if err != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the reseller info for id [ %v ]. Error: %v\n", id, err)

		return nil, err

	}

	reseller := r.(cusModels.Reseller)

	// Get all the usages linked to the reseller
	for _, customer := range reseller.Customers {

		for _, product := range customer.Products {

			u := d.GetReport(product.ProductID, "", from, to)

			usages = append(usages, u...)

		}

	}

	// Scan the usages and make the numbers
	resourcesMap := make(datamodels.JSONdb)

	for i := range usages {

		masterKey := strings.ToTitle(usages[i].ResourceType)

		if _, exists := resourcesMap[masterKey]; !exists {

			resourcesMap[masterKey] = make(datamodels.JSONdb)

		}

		// First the total
		key := "TotalCount"
		if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key]; !exists {

			resourcesMap[masterKey].(datamodels.JSONdb)[key] = float64(0)

		}

		resourcesMap[masterKey].(datamodels.JSONdb)[key] = resourcesMap[masterKey].(datamodels.JSONdb)[key].(float64) + float64(1/rounder)

		// Second the regions
		if region, exists := usages[i].Metadata["region"]; exists {

			key := "Regions"

			if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key]; !exists {

				resourcesMap[masterKey].(datamodels.JSONdb)[key] = make(datamodels.JSONdb)

			}

			regionKey := region.(string)
			if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[regionKey]; !exists {

				resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[regionKey] = float64(0)

			}

			resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[regionKey] = resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[regionKey].(float64) + float64(1/rounder)

		}

		// Third the flavors in case of
		if flavor, exists := usages[i].Metadata["flavorname"]; exists {

			key := "Flavors"

			if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key]; !exists {

				resourcesMap[masterKey].(datamodels.JSONdb)[key] = make(datamodels.JSONdb)

			}

			flavorKey := flavor.(string)
			if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[flavorKey]; !exists {

				resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[flavorKey] = float64(0)

			}

			resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[flavorKey] = resourcesMap[masterKey].(datamodels.JSONdb)[key].(datamodels.JSONdb)[flavorKey].(float64) + float64(1/rounder)

		}

		// Fourth the sizes in case of
		if size, exists := usages[i].Metadata["size"]; exists {

			sizeF, err := strconv.ParseFloat(size.(string), 64)

			if err != nil {

				l.Warning.Printf("[DB] Failed to turn the size [ %v ] into a float64 value. Error: %v\n", size, err)

			} else {

				key := "TotalAmount"

				if _, exists := resourcesMap[key]; !exists {

					resourcesMap[masterKey].(datamodels.JSONdb)[key] = float64(0)

				}

				resourcesMap[masterKey].(datamodels.JSONdb)[key] = resourcesMap[masterKey].(datamodels.JSONdb)[key].(float64) + float64(sizeF/rounder)

			}

		}

		// UDR usages
		if size, exists := usages[i].UsageBreakup["used"]; exists {

			var sizeF float64
			key := "TotalAmount"

			// size to float64
			if k := reflect.ValueOf(size); k.Kind() == reflect.Float64 {

				sizeF = size.(float64) / interval

			} else {

				v, err := size.(json.Number).Float64()

				if err != nil {

					l.Warning.Printf("[DB] Failed to turn the size [ %v ] into a float64 value. Error: %v\n", size, err)

				} else {

					sizeF = float64(v) / interval

				}

			}

			if _, exists := resourcesMap[masterKey].(datamodels.JSONdb)[key]; !exists {

				resourcesMap[masterKey].(datamodels.JSONdb)[key] = float64(0)

			}

			resourcesMap[masterKey].(datamodels.JSONdb)[key] = resourcesMap[masterKey].(datamodels.JSONdb)[key].(float64) + float64(sizeF/rounder)

		}

	}

	// Reshape of the JSON to easy showing in the front..
	for key, value := range resourcesMap {

		metricData := value.(datamodels.JSONdb)

		if _, exists := metricData["Regions"]; exists {

			var regions []datamodels.JSONdb

			for k, v := range metricData["Regions"].(datamodels.JSONdb) {

				region := make(datamodels.JSONdb)
				region["name"] = k
				region["amount"] = v

				regions = append(regions, region)

			}

			resourcesMap[key].(datamodels.JSONdb)["Regions"] = regions

		}

		if _, exists := metricData["Flavors"]; exists {

			var regions []datamodels.JSONdb

			for k, v := range metricData["Flavors"].(datamodels.JSONdb) {

				region := make(datamodels.JSONdb)
				region["name"] = k
				region["amount"] = v

				regions = append(regions, region)

			}

			resourcesMap[key].(datamodels.JSONdb)["Flavors"] = regions

		}

	}

	// Create the summarey report
	report.UsageBreakup = resourcesMap

	return &report, nil

}

// getWindow job is to select the timeframe for the usage retrievals according
// to the data providad from<window, window<to, or from<window<to.
// Parameters:
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - s: a string with the needed window correctly formated to be used with GORM.
func (d *DbParameter) getWindow(from, to strfmt.DateTime) (s string) {

	l.Trace.Printf("[DB] Attempting to get a window for a db query.\n")

	if !((time.Time)(from)).IsZero() {

		if ((time.Time)(to)).IsZero() {

			s = fmt.Sprintf("%v >= '%v'", d.Db.NamingStrategy.ColumnName("", "TimeFrom"), from)

		} else {

			s = fmt.Sprintf("%v >= '%v' AND %v <= '%v'", d.Db.NamingStrategy.ColumnName("", "TimeFrom"), from, d.Db.NamingStrategy.ColumnName("", "TimeTo"), to)

		}

	} else {

		if !((time.Time)(to)).IsZero() {

			s = fmt.Sprintf("%v <= '%v'", d.Db.NamingStrategy.ColumnName("", "TimeTo"), to)

		}

	}

	return s

}

// ProcessUDR job is to process the UDR Reports and turn then into CDR Reports
// and store them in the system.
// Parameters:
// - report: UDR Report model reference to be processed.
// - token: an optional keycloak token in case it's provided.
// Returns:
// - e: error in case of duplication or failure in the task.
func (d *DbParameter) ProcessUDR(report udrModels.UReport, token string) (e error) {

	l.Trace.Printf("[DB] Attempting to process and transform the UDR report into a CDR report and save it to the system.\n")

	// First we clean all the records so we always have a clean UDR generation
	templateRecord := models.CDRRecord{
		AccountID: report.AccountID,
		TimeFrom:  report.TimeFrom,
		TimeTo:    report.TimeTo,
	}

	l.Info.Printf("[DB] Clean up of previous CDRRecords started for period [ %v ] - [ %v ].\n", report.TimeFrom, report.TimeTo)

	if e := d.Db.Where(&templateRecord).Delete(&models.CDRRecord{}).Error; e != nil {

		l.Warning.Printf("[DB] The clean up of previous CDRecords failed. Error: %v\n", e)

	}

	templateReport := models.CReport{
		AccountID: report.AccountID,
		TimeFrom:  report.TimeFrom,
		TimeTo:    report.TimeTo,
	}

	l.Info.Printf("[DB] Clean up of previous CReports started for period [ %v ] - [ %v ].\n", report.TimeFrom, report.TimeTo)

	if e := d.Db.Where(&templateReport).Delete(&models.CReport{}).Error; e != nil {

		l.Warning.Printf("[DB] The clean up of previous CReports failed. Error: %v\n", e)

	}

	// First we get the metric and accounts in the system.
	var planID string
	var cdr models.CReport
	var usages []*models.CDRReport

	now := time.Now().UnixNano()

	x, e := d.Cache.Get(report.AccountID, "product", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the product info for id [ %v ]. Error: %v\n", report.AccountID, e)

		return

	}

	product := x.(cusModels.Product)

	if product.PlanID != "" {

		planID = product.PlanID

	} else if product.Type != "" {

		planID = product.Type

	} else {

		c, e := d.Cache.Get(product.CustomerID, "customer", token)

		if e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving the customer info for id [ %v ]. Error: %v\n", product.CustomerID, e)

			return e

		}

		customer := c.(cusModels.Customer)

		if customer.PlanID != "" {

			planID = customer.PlanID

		} else {

			r, e := d.Cache.Get(customer.ResellerID, "reseller", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the reseller info for id [ %v ]. Error: %v\n", customer.ResellerID, e)

				return e

			}

			reseller := r.(cusModels.Reseller)

			planID = reseller.PlanID

		}

	}

PlanDefault:
	p, e := d.Cache.Get(planID, "plan", token)

	if e != nil {

		if planID == "DEFAULT" {

			l.Warning.Printf("[DB] Something went wrong while retrieving the default plan id [ %v ]. Error: %v\n", planID, e)

			return

		}

		l.Warning.Printf("[DB] Something went wrong while retrieving the plan id [ %v ]. Re-trying with default plan. Error: %v\n", planID, e)

		planID = "DEFAULT"

		goto PlanDefault

	}

	plan := p.(pmModels.Plan)
	savedPlan := p.(pmModels.Plan)

	// In case the plan is not valid we return to the deault plan (id 0) which is valid ad vitam
	if (time.Now()).After((time.Time)(*plan.OfferedEndDate)) || (time.Now()).Before((time.Time)(*plan.OfferedStartDate)) {

		l.Warning.Printf("[DB] The plan [ %v ] is only valid between [ %v ] and [ %v ]. Falling back to default plan.\n", plan.ID, *plan.OfferedStartDate, *plan.OfferedEndDate)

		planID = "DEFAULT"

		goto PlanDefault

	}

	s, e := d.Cache.Get("ALL", "sku", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the skus list. Error: %v\n", e)

		return

	}

	skus := make(map[string]string)

	for _, k := range s.([]*pmModels.Sku) {

		skus[*k.Name] = k.ID

	}

	// having the customer and the plan...
	// iter the report
	for _, u := range report.Usage {

		costBreakup := make(datamodels.JSONdb)

		var costSkuTotal []datamodels.JSONdb
		var cycles []*pmModels.Cycle
		var bundle pmModels.SkuBundle
		var skuDiscount, skuPrice, usage, multiplier, amount, quantity float64

		// get cycle of resourceType
		c, e := d.Cache.Get(u.ResourceType, "cycle", token)

		if e != nil {

			l.Warning.Printf("[DB] Something went wrong while retrieving the states for the life cycle linked to the ResourceType [ %v ]. Error: %v\n", u.ResourceType, e)

			return e

		}

		cycles = c.([]*pmModels.Cycle)

		// if server, get the skubundle associated
		if _, exist := cycles[0].SkuList[u.ResourceType]; !exist {

			if id, exists := u.Metadata["flavorid"]; exists {

				b, e := d.Cache.Get(id, "bundle", token)

				if e != nil {

					l.Warning.Printf("[DB] Something went wrong while retrieving the skuBundle id [ %v ]. Error: %v\n", u.Metadata["flavorid"], e)

					return e

				}

				bundle = b.(pmModels.SkuBundle)

				// Usecase: flavor is a generic one but then the image is a windows base one, so the license is not included by default
				if id, exists := u.Metadata["imagename"]; exists && strings.Contains(strings.ToLower(id.(string)), "windows") {

					_, license := bundle.SkuPrices["license"]
					vcpu, exists := bundle.SkuPrices["vcpu"]

					if !license && exists {

						var previous pmModels.SkuBundle

						previous.ID = bundle.ID
						previous.Name = bundle.Name
						previous.SkuPrices = make(datamodels.JSONdb)

						for k, v := range bundle.SkuPrices {

							previous.SkuPrices[k] = v

						}

						previous.SkuPrices["license"] = vcpu

						bundle = previous

					}

				}

			} else {

				l.Warning.Printf("[DB] Something is wrong. Error: The flavorid is supposed to exist in the metadata.\n")

				continue

			}

		}

		// METADA OVERRIDERS
		if override, exists := u.Metadata["PlanOverride"]; exists && override.(bool) {

			overridePlan, e := d.Cache.Get("DEFAULT", "plan", token)

			if e != nil {

				l.Warning.Printf("[DB] Something went wrong while retrieving the Overriding plan. Error: %v\n", e)

			} else {

				plan = overridePlan.(pmModels.Plan)

			}

		} else {

			plan = savedPlan

		}

		// loop cycle
		for i := range cycles {

			// if usageBreakup(cycle.state) !exists, continue
			if stateUse, exists := u.UsageBreakup[*(cycles[i].State)]; !exists {

				continue

				// else, get the usage value
			} else {

				// usage to float64
				if k := reflect.ValueOf(stateUse); k.Kind() == reflect.Float64 {

					usage = stateUse.(float64)

				} else {

					v, _ := stateUse.(json.Number).Int64()
					usage = float64(v)

				}

			}

			// and loop actual.cycle.skus
			for sku, value := range cycles[i].SkuList {

				// get the skuDiscount and skuPrice associated
				if id, exists := skus[sku]; exists {

					for j := range plan.SkuPrices {

						if *plan.SkuPrices[j].SkuID == id {

							skuDiscount = float64(plan.SkuPrices[j].Discount)
							skuPrice = float64(*plan.SkuPrices[j].UnitPrice)

							break

						}

					}

				} else {

					l.Warning.Printf("[DB] Something is wrong. Error: [ %v ] is supposed to exist in the system as a sku.\n", u.ResourceType)

					continue

				}

				// multiplier to float64
				if k := reflect.ValueOf(value); k.Kind() == reflect.Float64 {

					multiplier = value.(float64)

				} else {

					v, _ := value.(json.Number).Int64()
					multiplier = float64(v)

				}

				// if resType not server then use actual.cycle.skus as multiplier for usage
				if *cycles[i].ResourceType == sku {

					// Size of buckets to be computed
					if sz, exists := u.Metadata["size"]; exists {

						size, e := strconv.ParseInt(sz.(string), 10, 0)

						if e != nil {

							l.Warning.Printf("[DB] Something is wrong with the size [ %v ] of the metadata. Error: %v.\n", sz, u.ResourceType)

							amount = usage * multiplier

						} else {

							amount = usage * multiplier * float64(size)

						}

					} else {

						amount = usage * multiplier

					}

					// else, resType is server
				} else {

					// check if in bundle for amounts, get bundle value times the usage and use actual.cycle.skus as an extra multiplier
					if q, exists := bundle.SkuPrices[sku]; !exists {

						continue

					} else {

						// quantity to float64
						if k := reflect.ValueOf(q); k.Kind() == reflect.Float64 {

							quantity = q.(float64)

						} else {

							v, _ := q.(json.Number).Int64()
							quantity = float64(v)

						}

						amount = usage * multiplier * quantity

					}

				}

				// create the cost elemt
				costSku := make(datamodels.JSONdb)
				costSku["sku"] = sku
				costSku["sku-state"] = *cycles[i].State

				// get cost
				costSku["sku-cost"] = amount * skuPrice

				// get discount as % of cost
				costSku["sku-discount"] = costSku["sku-cost"].(float64) * skuDiscount

				// get the net cost as cost-discount
				costSku["sku-net"] = costSku["sku-cost"].(float64) - costSku["sku-discount"].(float64)

				// Add the cost to the collection
				costSkuTotal = append(costSkuTotal, costSku)

			}

		}

		// place the costbreakup in place
		costBreakup["costBreakup"] = costSkuTotal

		if costBreakup["costBreakup"] == nil {

			l.Warning.Printf("[DB] Something went wrong in the cost breakup generation. Skipping this element...\n")

			continue

		}

		// Let's create the total cost
		costBreakup["totalFromSku"] = float64(0)
		for _, c := range costBreakup["costBreakup"].([]datamodels.JSONdb) {

			costBreakup["totalFromSku"] = costBreakup["totalFromSku"].(float64) + c["sku-net"].(float64)

		}

		costBreakup["appliedDiscount"] = costBreakup["totalFromSku"].(float64) * float64(plan.Discount)
		costBreakup["netTotal"] = costBreakup["totalFromSku"].(float64) - costBreakup["appliedDiscount"].(float64)

		var use models.CDRReport
		use.Cost = costBreakup
		use.Metadata = u.Metadata
		use.ResourceID = u.ResourceID
		use.ResourceName = u.ResourceName
		use.ResourceType = u.ResourceType
		use.Unit = u.Unit
		use.UsageBreakup = u.UsageBreakup

		usages = append(usages, &use)

	}

	cdr.AccountID = report.AccountID
	cdr.TimeFrom = report.TimeFrom
	cdr.TimeTo = report.TimeTo
	cdr.Usage = usages

	if e = d.AddRecord(cdr); e != nil {

		l.Warning.Printf("[DB] Something went wrong while saving the CDR record in the system. Error: %v\n", e)

		return

	}

	l.Trace.Printf("[DB] UDR processed and transformed into CDR and saved in the system successfully .\n")

	d.Pipe <- cdr

	l.Trace.Printf("[DB] CDR transmited to the Credit Manager successfully .\n")

	totalTime += float64(time.Now().UnixNano() - now)
	totalCount++

	d.Metrics["count"].With(prometheus.Labels{"type": "Total UDRs transformed to CDRs"}).Inc()

	d.Metrics["time"].With(prometheus.Labels{"type": "CDRs average generation time"}).Set(totalTime / float64(totalCount) / float64(time.Millisecond))

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

		o = float64(math.Round(i*scaler) / scaler)

	}

	return

}

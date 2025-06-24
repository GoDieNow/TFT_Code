package dbManager

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	cdrModels "github.com/GoDieNow/TFT_Code/services/cdr/models"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/models"
	"github.com/GoDieNow/TFT_Code/services/creditsystem/server/cacheManager"
	cusModels "github.com/GoDieNow/TFT_Code/services/customerdb/models"
	pmModels "github.com/GoDieNow/TFT_Code/services/planmanager/models"
	l "gitlab.com/cyclops-utilities/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	scaler           = 1e7
	statusDuplicated = iota
	statusFail
	statusMissing
	statusOK
	AC_CREDIT  = "CREDIT"
	AC_CASH    = "CASH"
	AC_BOTH    = "BOTH"
	AC_NONE    = "NONE"
	MED_CASH   = "CASH"
	MED_CREDIT = "CREDIT"
)

var (
	//states = []string{"active", "error", "inactive", "suspended", "terminated"}
	states     = []string{"active", "error", "inactive", "used"}
	totalTime  float64
	totalCount int64
)

// DbParameter is the struct defined to group and contain all the methods
// that interact with the database.
// On it there is the following parameters:
// - Cache: CacheManager pointer for the cache mechanism.
// - connStr: strings with the connection information to the database
// - Db: a gorm.DB pointer to the db to invoke all the db methods
type DbParameter struct {
	Cache   *cacheManager.CacheManager
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
// - DbParameter: struct to interact with dbManager functionalities¬
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

// AddConsumption job is to decrease the credit in the system linked to the
// provided account by a certain amount of credit.
// This function is the one that produces a "cosumed" decrease of credit without
// human intervention.
// Parameters:
// - id: string containing the id of the account requested.
// - amount: float containing the amount to be -- in the system
// Returns:
// - reference to creditStatus containing the credit balance after the operation.
// - error raised in case of problems.
func (d *DbParameter) AddConsumption(id string, amount float64, medium string) (*models.CreditStatus, error) {

	l.Trace.Printf("[DB] Attempting to add a comsuption of %v credit, in the account with id: %v", amount, id)

	var c, c0, cs models.CreditStatus
	var ce models.CreditEvents
	var e error

	if r := d.Db.Where(&models.CreditStatus{AccountID: id}).First(&c).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	} else {

		med := strings.ToUpper(medium)

		if med == models.CreditEventsMediumCREDIT {

			c0.AvailableCredit = c.AvailableCredit - amount

		}

		if med == models.CreditEventsMediumCASH {

			c0.AvailableCash = c.AvailableCash - amount

		}

		c0.LastUpdate = strfmt.DateTime(time.Now())

		ce.AccountID = id
		ce.Delta = -amount
		ce.EventType = func(s string) *string { return &s }(models.EventEventTypeConsumption)
		ce.Timestamp = strfmt.DateTime(time.Now())
		ce.Medium = &med

		if e = d.Db.Model(&c).Updates(c0).Error; e == nil {

			if e = d.Db.Create(&ce).Error; e == nil {

				d.Db.Where(&models.CreditStatus{AccountID: id}).First(&cs)

			} else {

				l.Trace.Printf("[DB] Attempting to add the comsuption of credit event for account in the system failed.")

			}

		} else {

			l.Trace.Printf("[DB] Attempting to update the account in the system failed.")

		}

	}

	return &cs, e

}

// CreateAccount job is to register a new account in the system with the provided
// ID.
// Parameters:
// - id: string containing the id of the account to be created.
// Returns:
// - reference to AccountStatus containing the state of the account after the
// operation.
// - error raised in case of problems.
func (d *DbParameter) CreateAccount(id string) (*models.AccountStatus, error) {

	l.Trace.Printf("[DB] Attempting to create a new account with id: %v", id)

	var a, a0, as models.AccountStatus
	var c models.CreditStatus
	var e error

	if r := d.Db.Where(&models.AccountStatus{AccountID: id}).First(&a).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		a0.AccountID = id
		a0.CreatedAt = strfmt.DateTime(time.Now())

		c.AccountID = id
		c.LastUpdate = strfmt.DateTime(time.Now())

		if e = d.Db.Create(&a0).Error; e == nil {

			if e = d.Db.Create(&c).Error; e == nil {

				d.Db.Where(&models.AccountStatus{AccountID: id}).First(&as)

				d.Metrics["count"].With(prometheus.Labels{"type": "Accounts created"}).Inc()

			} else {

				l.Trace.Printf("[DB] Attempting to create the credit account in the system failed.")

			}

		} else {

			l.Trace.Printf("[DB] Attempting to create the account in the system failed.")

		}

	} else {

		l.Trace.Printf("[DB] Account with id: %v already in the system, check with administrator.", id)

		e = errors.New("account already exist in the system")

	}

	return &as, e

}

// DecreaseCredit job is to decrease the credit in the system linked to the
// provided account by a certain amount of credit.
// Parameters:
// - id: string containing the id of the account requested.
// - amount: float containing the amount to be decreased in the system.
// Returns:
// - reference to CreditStatus containing the credit balance after the operation.
// - error raised in case of problems.
func (d *DbParameter) DecreaseCredit(id string, amount float64, medium string) (*models.CreditStatus, error) {

	l.Trace.Printf("[DB] Attempting to decrease credit by: %v, in the account with id: %v", amount, id)

	var c, c0, cs models.CreditStatus
	var ce models.CreditEvents
	var e error

	if r := d.Db.Where(&models.CreditStatus{AccountID: id}).First(&c).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	} else {

		med := strings.ToUpper(medium)

		if med == models.CreditEventsMediumCREDIT {

			c0.AvailableCredit = c.AvailableCredit - amount

		}

		if med == models.CreditEventsMediumCASH {

			c0.AvailableCash = c.AvailableCash - amount

		}

		c0.LastUpdate = strfmt.DateTime(time.Now())

		ce.AccountID = id
		ce.Delta = -amount
		ce.EventType = func(s string) *string { return &s }(models.EventEventTypeAuthorizedDecrease)
		ce.Timestamp = strfmt.DateTime(time.Now())
		ce.Medium = &med

		if e = d.Db.Model(&c).Updates(c0).Error; e == nil {

			if e = d.Db.Create(&ce).Error; e == nil {

				d.Db.Where(&models.CreditStatus{AccountID: id}).First(&cs)

			} else {

				l.Trace.Printf("[DB] Attempting to add the decrease of credit event for account in the system failed.")

			}

		} else {

			l.Trace.Printf("[DB] Attempting to update the account in the system failed.")

		}

	}

	return &cs, e

}

// DisableAccount job is to mark as disabled in the system the account provided.
// Parameters:
// - id: string containing the id of the account to be disabled.
// Returns:
// - reference to AccountStatus containing the state of the account after the
// operation.
// - error raised in case of problems.
func (d *DbParameter) DisableAccount(id string) (*models.AccountStatus, error) {

	l.Trace.Printf("[DB] Attempting to disable account with id: %v", id)

	var a, as models.AccountStatus
	var e error

	if r := d.Db.Where(&models.AccountStatus{AccountID: id}).First(&a).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	} else {

		enabled := false

		if e = d.Db.Model(&a).Updates(&models.AccountStatus{Enabled: enabled}).Error; e == nil {

			d.Db.Where(&models.AccountStatus{AccountID: id}).First(&as)

		} else {

			l.Trace.Printf("[DB] Attempting to update the account in the system failed.")

		}

	}

	return &as, e

}

// EnableAccount job is to mark as enabled in the system the account provided.
// Parameters:
// - id: string containing the id of the account to be enabled.
// Returns:
// - reference to AccountStatus containing the state of the account after the
// operation.
// - error raised in case of problems.
func (d *DbParameter) EnableAccount(id string) (*models.AccountStatus, error) {

	l.Trace.Printf("[DB] Attempting to enable account with id: %v", id)

	var a, as models.AccountStatus
	var e error

	if r := d.Db.Where(&models.AccountStatus{AccountID: id}).First(&a).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	} else {

		if e = d.Db.Model(&a).Update("Enabled", true).Error; e == nil {

			d.Db.Where(&models.AccountStatus{AccountID: id}).First(&as)

		} else {

			l.Trace.Printf("[DB] Attempting to update the account in the system failed.")

		}

	}

	return &as, e

}

// GetAccountStatus job is to retrieve the actual state of the provided account.
// Parameters:
// - id: string containing the id of the account to be retrieved.
// Returns:
// - reference to AccountStatus containing the state of the account.
// - error raised in case of problems.
func (d *DbParameter) GetAccountStatus(id string) (*models.AccountStatus, error) {

	l.Trace.Printf("[DB] Attempting to get the status of the account with id: %v", id)

	var as models.AccountStatus
	var e error

	if r := d.Db.Where(&models.AccountStatus{AccountID: id}).First(&as).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	}

	return &as, e

}

// GetCredit job is to retrieve the actual state of the credit balance of the
// provided account.
// Parameters:
// - id: string containing the id of the account requested.
// Returns:
// - reference to CreditStatus containing the credit balance.
// - error raised in case of problems.
func (d *DbParameter) GetCredit(id string) (*models.CreditStatus, error) {

	l.Trace.Printf("[DB] Attempting to get the credit balance in the account with id: %v", id)

	var cs models.CreditStatus
	var e error

	if r := d.Db.Where(&models.CreditStatus{AccountID: id}).First(&cs).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	}

	return &cs, e

}

// GetHistory job is to retrieve the history of the credit balance of the account
// provided with the posibility of filter the actions.
// Parameters:
// - id: string containing the id of the account requested.
// - filtered: string specifying the criteria for filtering.
// Returns:
// - reference to CreditHistory containing the credit balance history contained
// in the system.
// - error raised in case of problems.
func (d *DbParameter) GetHistory(id string, filtered bool, medium string) (*models.CreditHistory, error) {

	l.Trace.Printf("[DB] Attempting to get the credit history in the account with id: %v", id)

	var ch models.CreditHistory
	var ce []*models.CreditEvents
	var e0 []*models.Event
	var e error

	ch.AccountID = id

	med := strings.ToUpper(medium)

	if filtered {

		evType := models.EventEventTypeConsumption

		e = d.Db.Where(&models.CreditEvents{AccountID: id, Medium: &med}).Not(&models.CreditEvents{EventType: &evType}).Find(&ce).Error

	} else {

		e = d.Db.Where(&models.CreditEvents{AccountID: id, Medium: &med}).Find(&ce).Error

	}

	if e != nil {

		l.Trace.Printf("[DB] Account with id: %v doesn't have events in the system, check with administrator.", id)

	}

	for _, event := range ce {

		var ev models.Event
		ev.Delta = event.Delta
		ev.EventType = event.EventType
		ev.Timestamp = event.Timestamp
		ev.AuthorizedBy = event.AuthorizedBy
		e0 = append(e0, &ev)

	}

	l.Trace.Printf("[DB] Amount of events for the account with id: %v, is: %v.", id, len(e0))

	ch.Events = e0

	return &ch, e

}

// IncreaseCredit job is to increase the credit in the system linked to the
// provided account by a certain amount of credit.
// Parameters:
// - id: string containing the id of the account requested.
// - amount: float containing the amount to be -- in the system
// Returns:
// - reference to CreditStatus containing the credit balance after the operation.
// - error raised in case of problems.
func (d *DbParameter) IncreaseCredit(id string, amount float64, medium string) (*models.CreditStatus, error) {

	l.Trace.Printf("[DB] Attempting to increase credit by: %v, in the account with id: %v", amount, id)

	var c, c0, cs models.CreditStatus
	var ce models.CreditEvents
	var e error

	if r := d.Db.Where(&models.CreditStatus{AccountID: id}).First(&c).Error; errors.Is(r, gorm.ErrRecordNotFound) {

		l.Trace.Printf("[DB] Account with id: %v doesn't exist in the system, check with administrator.", id)

		e = errors.New("account doesn't exist in the system")

	} else {

		med := strings.ToUpper(medium)

		if med == models.CreditEventsMediumCREDIT {

			c0.AvailableCredit = c.AvailableCredit + amount

		}

		if med == models.CreditEventsMediumCASH {

			c0.AvailableCash = c.AvailableCash + amount

		}

		c0.LastUpdate = strfmt.DateTime(time.Now())

		ce.AccountID = id
		ce.Delta = amount
		ce.EventType = func(s string) *string { return &s }(models.EventEventTypeAuthorizedIncrease)
		ce.Timestamp = strfmt.DateTime(time.Now())
		ce.Medium = &med

		if e = d.Db.Model(&c).Updates(c0).Error; e == nil {

			if e = d.Db.Create(&ce).Error; e == nil {

				d.Db.Where(&models.CreditStatus{AccountID: id}).First(&cs)

			} else {

				l.Trace.Printf("[DB] Attempting to add the increase of credit event for account in the system failed.")

			}

		} else {

			l.Trace.Printf("[DB] Attempting to update the account in the system failed.")

		}

	}

	return &cs, e

}

// ListAccounts job is to retrieve the list of all the accounts safe in the system.
// Returns:
// - slice of references to AccountStatus containing the status of every account
// in the system
func (d *DbParameter) ListAccounts() ([]*models.AccountStatus, error) {

	l.Trace.Printf("[DB] Attempting to list accounts in the system.")

	var as []*models.AccountStatus
	var e error

	if e = d.Db.Find(&as).Error; e != nil {

		l.Trace.Printf("[DB] There is not accounts in the system")

	}

	return as, e

}

// ProcessCDR job is to check every CDR report arriving to the system to add the
// consumptions of each account to the corresponding
// Parameters:
// - report: CDR Report model reference to be processed.
// - usage: a boolean discriminant to set the accounting to usage instead of cost
// Returns:
// - e: error in case of problem while doing a conversion, a cache.get, or adding the comsumption.
func (d *DbParameter) ProcessCDR(report cdrModels.CReport, token string) error {

	l.Trace.Printf("[DB] Starting the processing of the a CDR.\n")

	var planID string

	now := time.Now().UnixNano()

	listAccounts, e := d.ListAccounts()

	if e != nil {

		l.Warning.Printf("[DB] There's was an error retrieving the accounts in the system. Error: %v\n", e)

		return e

	}

	if len(listAccounts) < 1 {

		l.Trace.Printf("[DB] There's no accounts in the system right now, skipping the processing...\n")

		return nil

	}

	p, e := d.Cache.Get(report.AccountID, "product", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the product info for id [ %v ]. Error: %v\n", report.AccountID, e)

		return e

	}

	product := p.(cusModels.Product)

	c, e := d.Cache.Get(product.CustomerID, "customer", token)

	if e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the customer info for id [ %v ]. Error: %v\n", product.CustomerID, e)

		return e

	}

	customer := c.(cusModels.Customer)

	account, e := d.GetAccountStatus(customer.CustomerID)

	if e != nil {

		l.Trace.Printf("[DB] There's no account associated in the customer level ID: [ %v ], moving to the next level...\n", customer.CustomerID)

		acc, e := d.GetAccountStatus(customer.ResellerID)

		if e != nil {

			l.Trace.Printf("[DB] There's no account associated in the reseller level. ID: [ %v ], skipping this product [ %v ].\n", customer.ResellerID, product.ProductID)

			return nil

		}

		account = acc

	}

	if !(account.Enabled) {

		l.Trace.Printf("[DB] Account associated [ %v ] is disabled! Skipping...\n", account.AccountID)

		return nil

	}

	if product.PlanID != "" {

		planID = product.PlanID

	} else {

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
	p, e = d.Cache.Get(planID, "plan", token)

	if e != nil {

		if planID == "DEFAULT" {

			l.Warning.Printf("[DB] Something went wrong while retrieving the default plan id [ %v ]. Error: %v\n", planID, e)

			return e

		}

		l.Warning.Printf("[DB] Something went wrong while retrieving the plan id [ %v ]. Re-trying with default plan. Error: %v\n", planID, e)

		planID = "DEFAULT"

		goto PlanDefault

	}

	plan := p.(pmModels.Plan)

	// In case the plan is not valid we return to the deault plan (id 0) which is valid ad vitam
	if (time.Now()).After((time.Time)(*plan.OfferedEndDate)) || (time.Now()).Before((time.Time)(*plan.OfferedStartDate)) {

		l.Warning.Printf("[DB] The plan [ %v ] is only valid between [ %v ] and [ %v ]. Falling back to default plan.\n", plan.ID, *plan.OfferedStartDate, *plan.OfferedEndDate)

		planID = "DEFAULT"

		goto PlanDefault

	}

	skus := make(map[string]*pmModels.SkuPrice)

	for _, k := range plan.SkuPrices {

		skus[k.SkuName] = k

	}

	var cashDelta, creditDelta float64

	for i := range report.Usage {

		var value interface{}

		sku, exists := skus[report.Usage[i].ResourceType]

		if exists {

			mode := *sku.AccountingMode

			if mode == AC_CREDIT || mode == AC_BOTH {

				for _, j := range states {

					value = report.Usage[i].UsageBreakup[j]

					if value == nil {

						l.Debug.Printf("[DB] The state [ %v ] seem to not be part of the usage UsageBreakup [ %+v ], skipping...", j, report.Usage[i].UsageBreakup)

						continue

					}

					if k := reflect.ValueOf(value); k.Kind() == reflect.Float64 {

						creditDelta += value.(float64) * skus[report.Usage[i].ResourceType].UnitCreditPrice

						continue

					}

					v, e := value.(json.Number).Float64()

					if e != nil {

						l.Warning.Printf("[DB] There was a problem retrieving the float value from the usagebreakup interface. Error: %v.\n", e)

						return e

					}

					creditDelta += v * skus[report.Usage[i].ResourceType].UnitCreditPrice

				}

			}

			if mode == AC_CASH || mode == AC_BOTH {

				value = report.Usage[i].Cost["totalFromSku"]

				if k := reflect.ValueOf(value); k.Kind() == reflect.Float64 {

					cashDelta += value.(float64)

					continue

				}

				v, e := value.(json.Number).Float64()

				if e != nil {

					l.Warning.Printf("[DB] There was a problem retrieving the float value from the cost interface. Error: %v.\n", e)

					return e

				}

				cashDelta += v

			}

		} else {

			states_string := strings.Join(states, " ")

			if report.Usage[i].Cost["costBreakup"] != nil {

				for _, unit := range report.Usage[i].Cost["costBreakup"].([]interface{}) {

					sku := unit.(map[string]interface{})

					model, exists := skus[sku["sku"].(string)]

					if exists {

						mode := *model.AccountingMode

						if mode == AC_CREDIT || mode == AC_BOTH {

							if strings.Contains(states_string, sku["sku-state"].(string)) {

								value = report.Usage[i].UsageBreakup[sku["sku-state"].(string)]

								if value == nil {

									l.Debug.Printf("[DB] The state [ %v ] seem to not be part of the usage UsageBreakup [ %+v ], skipping...", sku["sku-state"], report.Usage[i].UsageBreakup)

									goto OutOfCREDIT

								}

								if k := reflect.ValueOf(value); k.Kind() == reflect.Float64 {

									creditDelta += value.(float64) * skus[sku["sku"].(string)].UnitCreditPrice

									goto OutOfCREDIT

								}

								v, e := value.(json.Number).Float64()

								if e != nil {

									l.Warning.Printf("[DB] There was a problem retrieving the float value from the usagebreakup interface. Error: %v.\n", e)

									return e

								}

								creditDelta += v * skus[sku["sku"].(string)].UnitCreditPrice

							}

						}

					OutOfCREDIT:

						if mode == AC_CASH || mode == AC_BOTH {

							value = sku["sku-cost"]

							if k := reflect.ValueOf(value); k.Kind() == reflect.Float64 {

								cashDelta += value.(float64)

								continue

							}

							v, e := value.(json.Number).Float64()

							if e != nil {

								l.Warning.Printf("[DB] There was a problem retrieving the float value from the cost interface. Error: %v.\n", e)

								return e

							}

							cashDelta += v

						}

					}

				}

			}

		}

	}

	state, e := d.AddConsumption(account.AccountID, d.getNiceFloat(creditDelta), MED_CREDIT)

	if e != nil {

		l.Warning.Printf("[DB] There was a problem while adding the consumption to the account [ %v ]. Error: %v.\n", account.AccountID, e)

		return e
	}

	if state.AvailableCredit < float64(0) {

		l.Warning.Printf("[DB] Account [ %v ] credit is not positive. Credit: [ %v ].\n", account.AccountID, state.AvailableCredit)

	}

	l.Trace.Printf("[DB] Account [ %v ] has been updated with a [ %v ] consumption.\n", account.AccountID, creditDelta)

	state, e = d.AddConsumption(account.AccountID, d.getNiceFloat(cashDelta), MED_CASH)

	if e != nil {

		l.Warning.Printf("[DB] There was a problem while adding the consumption to the account [ %v ]. Error: %v.\n", account.AccountID, e)

		return e
	}

	if state.AvailableCash < float64(0) {

		l.Warning.Printf("[DB] Account [ %v ] cash is not positive. Cash: [ %v ].\n", account.AccountID, state.AvailableCash)

	}

	l.Trace.Printf("[DB] Account [ %v ] has been updated with a [ %v ] consumption.\n", account.AccountID, cashDelta)

	totalTime += float64(time.Now().UnixNano()-now) / float64(time.Millisecond)
	totalCount++

	d.Metrics["count"].With(prometheus.Labels{"type": "CDRs procesed"}).Inc()

	d.Metrics["time"].With(prometheus.Labels{"type": "CDRs average processing time"}).Set(totalTime / float64(totalCount))

	return nil

}

func (d *DbParameter) getNiceFloat(i float64) (o float64) {

	return float64(math.Round(i*scaler) / scaler)

}

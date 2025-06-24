package dbManager

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/GoDieNow/TFT_Code/services/udr/models"
	"github.com/GoDieNow/TFT_Code/services/udr/server/cacheManager"
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

		panic(err)

	}

	l.Trace.Printf("[DB] Migrating tables.\n")

	//Database migration, it handles everything
	dp.Db.AutoMigrate(tables...)

	l.Trace.Printf("[DB] Generating hypertables.\n")

	// Hypertables creation for timescaledb in case of needed
	dp.Db.Exec("SELECT create_hypertable('" + dp.Db.NamingStrategy.TableName("Usage") + "', 'timedate');")

	return &dp

}

// AddMetric tries to insert a given new metric in the system, checking and
// reporting if it's already exists in the system.
// Parameters:
// - metric: string with the new metric to add to the system.
// Returns:
// - error in case of duplication or failure in the task.
func (d *DbParameter) AddMetric(metric string) error {

	l.Trace.Printf("[DB] Attempting to add a new metric [ %v ] to the system.\n", metric)

	m := models.Metric{Metric: metric}

	if e := d.Db.Where(&m).First(&models.Metric{}).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		e := d.Db.Create(&m).Error

		if e != nil {

			l.Warning.Printf("[DB] There were an problem adding the metric [ %v ] to the system. Error: %v\n", metric, e)

		} else {

			d.Metrics["count"].With(prometheus.Labels{"type": "Usage metrics added"}).Inc()

			l.Debug.Printf("[DB] Metric [ %v ] added to the system.\n", metric)

		}

		return e

	}

	l.Debug.Printf("[DB] The metric [ %v ] already exists in the system.\n", metric)

	return nil

}

// AddRecord job is to add a compacted record for a 8h interval in the system.
// Parameters:
// - report: a UDRRecord containing the Usage received and compacted.
// Returns:
// - error in case of duplication or failure in the task.
func (d *DbParameter) AddRecord(report models.UDRRecord) (e error) {

	l.Trace.Printf("[DB] Attempting to add a new compacted UDR Record in the system.\n")

	var reportUpdate models.UDRRecord

	reportCheck := models.UDRRecord{
		AccountID:    report.AccountID,
		Metadata:     report.Metadata,
		TimeTo:       report.TimeTo,
		TimeFrom:     report.TimeFrom,
		ResourceID:   report.ResourceID,
		ResourceName: report.ResourceName,
		ResourceType: report.ResourceType,
	}

	e = d.Db.Where(&reportCheck).First(&reportUpdate).Error

	if errors.Is(e, gorm.ErrRecordNotFound) {

		e = d.Db.Create(&report).Error

		if e == nil {

			d.Metrics["count"].With(prometheus.Labels{"type": "UDR Records added"}).Inc()

			l.Trace.Printf("[DB] New UDR Record added in the system.\n")

			return

		}

		l.Trace.Printf("[DB] Something went wrong when adding a new UDR Record to the system. Error: %v\n", e)

		return

	}

	l.Trace.Printf("[DB] The UDR Record [ %v ] is already in the system, attempt to update it...\n", reportUpdate)

	e = d.Db.Model(&reportUpdate).Updates(&report).Error

	if e == nil {

		d.Metrics["count"].With(prometheus.Labels{"type": "UDR Records updated"}).Inc()

		l.Trace.Printf("[DB] UDR Record updated in the system.\n")

		return

	}

	l.Trace.Printf("[DB] Something went wrong when updating the existing UDR Record to the system. Error: %v\n", e)

	return

}

// AddReport job is to add a new report of compacted usage to the system.
// Parameters:
// - report: a Report containing the data to be added to the system.
// Returns:
// - error in case of duplication or failure in the task.
func (d *DbParameter) AddReport(report models.UReport) (e error) {

	l.Trace.Printf("[DB] Attempting to add a new usage report to the system.\n")

	if e = d.Db.Where(&report).First(&models.UReport{}).Error; errors.Is(e, gorm.ErrRecordNotFound) {

		e = d.Db.Create(&report).Error

		if e == nil {

			d.Metrics["count"].With(prometheus.Labels{"type": "UDR Reports added"}).Inc()

			l.Trace.Printf("[DB] New usage report added in the system.\n")

		} else {

			l.Trace.Printf("[DB] Something went wrong when adding a new usage report to the system. Error: %v\n", e)

		}

		return e

	}

	return nil

}

// GetAccounts job is to retrieve a list of the accounts in the system.
// Returns:
// - Slice of strings containing the accounts in the system with usage data.
func (d *DbParameter) GetAccounts() []string {

	l.Trace.Printf("[DB] Attempting to retrieve the account list in the system.\n")

	type scann struct {
		Account string
	}

	var accounts []string
	var acc []scann

	if e := d.Db.Debug().Raw("SELECT DISTINCT " + d.Db.NamingStrategy.ColumnName("", "Account") + " FROM " + d.Db.NamingStrategy.TableName("Usage")).Scan(&acc).Error; e == nil {

		for i := range acc {

			accounts = append(accounts, acc[i].Account)

		}

		l.Trace.Printf("[DB] [ %v ] accounts were retrieved from the system.\n", len(accounts))

	} else {

		l.Warning.Printf("[DB] Something went wrong while retrieving the accounts in the system. Error: %v\n", e)

	}

	return accounts

}

// GetUDRAccounts job is to retrieve a list of the accounts in the system that
// have a UDR Report ready to be sent.
// Returns:
// - Slice of strings containing the accounts in the system with UDR Reports.
func (d *DbParameter) GetUDRAccounts() []string {

	l.Trace.Printf("[DB] Attempting to retrieve the account list in the system.\n")

	type scann struct {
		AccountID string
	}

	var accounts []string
	var acc []scann

	if e := d.Db.Debug().Raw("SELECT DISTINCT " + d.Db.NamingStrategy.ColumnName("", "AccountID") + " FROM " + d.Db.NamingStrategy.TableName("UDRRecord")).Scan(&acc).Error; e == nil {

		for i := range acc {

			accounts = append(accounts, acc[i].AccountID)

		}

		l.Trace.Printf("[DB] [ %v ] accounts were retrieved from the system.\n", len(accounts))

	} else {

		l.Warning.Printf("[DB] Something went wrong while retrieving the accounts in the system. Error: %v\n", e)

	}

	return accounts

}

// GetInterval job is to provide a proper query string to use in GORM to indicate
// a range of time.
// Parameters:
// - field: a string with the db field to use for the time-window.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - s: the ready for query string with the data provided.
func (d *DbParameter) GetInterval(field string, from strfmt.DateTime, to strfmt.DateTime) (s string) {

	l.Trace.Printf("[DB] Attempting to get an interval for a db query.\n")

	s = fmt.Sprintf("%v >= '%v' AND %v <= '%v'", field, from, field, to)

	return

}

// GetMetrics job is to retrieve the list of metrics registered in the system.
// Returns:
// - a slice of references with the metrics in the system.
// - error in case of duplication or failure in the task.
func (d *DbParameter) GetMetrics() ([]*models.Metric, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the metrics in the system.\n")

	var m []*models.Metric
	var e error

	if e = d.Db.Debug().Find(&m).Error; e == nil {

		l.Debug.Printf("[DB] [%v] metrics retrieved from the system.\n", len(m))

	} else {

		l.Warning.Printf("[DB] Something went wrong while retrieving metrics from the system. Error: %v\n", e)

	}

	return m, e

}

// GetRecords job is to retrieve the non-compacted usage records from the system
// in the requested time-window with the posibility of filter by metric.
// Parameters:
// - metric: a string with the metric to filter the records.
// - from: a datatime reference for the initial border of the time-window.
// - to: a datatime reference for the final border of the time-window.
// Returns:
// - a slice of references with the usage contained in the system within the
// requested time-window.
func (d *DbParameter) GetRecords(metric string, from strfmt.DateTime, to strfmt.DateTime) []*models.Usage {

	l.Trace.Printf("[DB] Attempting to retrieve the non-compacted usage records.\n")

	var u []*models.Usage

	interval := d.GetInterval("timedate", from, to)

	if e := d.Db.Where(interval).Where(&models.Usage{ResourceType: metric}).Find(&u).Error; e != nil {

		l.Warning.Printf("[DB] Something went wrong while retrieving the non-compacted usage records from the system. Error: %v\n", e)

	} else {

		l.Debug.Printf("[DB] [ %v ] non-compacted usage records retrieved from the system.\n", len(u))

	}

	return u

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
func (d *DbParameter) GetReport(id, metric string, from, to strfmt.DateTime) []*models.UDRReport {

	l.Trace.Printf("[DB] Attempting to retrieve the compacted usage records.\n")

	var r []*models.UDRReport
	var u models.UDRRecord

	window := d.getWindow(from, to)

	if id != "" {

		u.AccountID = id

	}

	if metric != "" {

		u.ResourceType = metric

	}

	if e := d.Db.Where(window).Where(&u).Find(&models.UDRRecord{}).Scan(&r).Error; e != nil {

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
func (d *DbParameter) GetUsage(id, metric string, from, to strfmt.DateTime) ([]*models.UReport, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the usage report from the system.\n")

	var r []*models.UReport
	var r0 models.UReport
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
func (d *DbParameter) GetUsages(ids, metric string, from, to strfmt.DateTime) ([]*models.UReport, error) {

	l.Trace.Printf("[DB] Attempting to retrieve the usage report from the system.\n")

	var total, r []*models.UReport
	var r0 models.UReport
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

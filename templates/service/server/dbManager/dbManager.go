package dbManager

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
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

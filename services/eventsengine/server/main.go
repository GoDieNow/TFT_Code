package main

import (
	"crypto/tls"
	"github.com/segmentio/encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/restapi"
	"github.com/GoDieNow/TFT_Code/services/eventsengine/server/dbManager"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	cfg     configuration
	version string
	service string
)

// dbStart handles the initialization of the dbManager returning a pointer to
// the DbParameter to be able to use the dbManager methods.
// Parameters:
// - models: variadic interface{} containing all the models that need to be
// migrated to the db.
// Returns:
// - DbParameter reference.
func dbStart(models ...interface{}) *dbManager.DbParameter {

	connStr := "user=" + cfg.DB.Username + " password=" + cfg.DB.Password +
		" dbname=" + cfg.DB.DbName + " sslmode=" + cfg.DB.SSLMode +
		" host=" + cfg.DB.Host + " port=" + strconv.Itoa(cfg.DB.Port)

	return dbManager.New(connStr, models...)

}

// getBasePath function is to get the base URL of the server.
// Returns:
// - String with the value of the base URL of the server.
func getBasePath() string {

	type jsonBasePath struct {
		BasePath string
	}

	bp := jsonBasePath{}

	e := json.Unmarshal(restapi.SwaggerJSON, &bp)

	if e != nil {

		l.Warning.Printf("[MAIN] Unmarshalling of the basepath failed: %v\n", e)

	}

	return bp.BasePath

}

func init() {

	confFile := flag.String("conf", "./config", "configuration file path (without toml extension)")

	flag.Parse()

	//placeholder code as the default value will ensure this situation will never arise
	if len(*confFile) == 0 {

		fmt.Printf("Usage: %v -conf=/path/to/configuration/file\n", strings.Title(service))

		os.Exit(0)

	}

	// err := gcfg.ReadFileInto(&cfg, *confFile)
	viper.SetConfigName(*confFile) // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".") // path to look for the config file in

	err := viper.ReadInConfig() // Find and read the config file

	if err != nil {

		// TODO(murp) - differentiate between file not found and formatting error in
		// config file)
		fmt.Printf("[MAIN] Failed to parse configuration data: %v\nCorrect usage: %v -conf=/path/to/configuration/file\n", err, strings.Title(service))

		os.Exit(1)

	}

	cfg = parseConfig()

	e := l.InitLogger(cfg.General.LogFile, cfg.General.LogLevel, cfg.General.LogToConsole)

	if e != nil {

		fmt.Printf("[MAIN] Initialization of the logger failed. Error: %v\n", e)

	}

	l.Info.Printf("Cyclops Labs %v Manager version %v initialized\n", strings.Title(service), version)

	dumpConfig(cfg)

}

func main() {

	//Getting the service handler and prometheus register
	h, r, e := getService()

	if e != nil {

		log.Fatal(e)

	}

	if cfg.Prometheus.MetricsExport {

		l.Info.Printf("[MAIN] Starting to serve Prometheus Metrics, access server on https://localhost:%v/%v\n", cfg.Prometheus.MetricsPort, cfg.Prometheus.MetricsRoute)

		go func() {

			http.Handle(cfg.Prometheus.MetricsRoute, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

			log.Fatal(http.ListenAndServe(":"+cfg.Prometheus.MetricsPort, nil))

		}()

	}

	serviceLocation := ":" + strconv.Itoa(cfg.General.ServerPort)

	if cfg.General.HTTPSEnabled {

		c := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
		}

		srv := &http.Server{
			Addr:         serviceLocation,
			Handler:      h,
			TLSConfig:    c,
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		}

		l.Info.Printf("[MAIN] Starting to serve %v, access server on https://localhost%v\n", strings.Title(service), serviceLocation)

		metricTime.With(prometheus.Labels{"type": "Service (HTTPS) start time"}).Set(float64(time.Now().Unix()))

		log.Fatal(srv.ListenAndServeTLS(cfg.General.CertificateFile, cfg.General.CertificateKey))

	} else {

		l.Info.Printf("[MAIN] Starting to serve %v, access server on http://localhost%v\n", strings.Title(service), serviceLocation)

		metricTime.With(prometheus.Labels{"type": "Service (HTTP) start time"}).Set(float64(time.Now().Unix()))

		// Run the standard http server
		log.Fatal(http.ListenAndServe(serviceLocation, h))

	}

}

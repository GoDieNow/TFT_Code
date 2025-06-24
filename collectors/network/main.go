package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/spf13/viper"
	cusClient "github.com/GoDieNow/TFT_Code/services/customerdb/client"
	eeClient "github.com/GoDieNow/TFT_Code/services/eventsengine/client"
	eeModels "github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	udrModels "github.com/GoDieNow/TFT_Code/services/udr/models"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	version       string
	cfg           configuration
	pipeU         chan interface{}
	pipeE         chan interface{}
	reportClient  *eeClient.EventEngineManagementAPI
	zombiesClient *cusClient.CustomerDatabaseManagement
)

// kafkaStart handles the initialization of the kafka service.
// This is a sample function with the most basic usage of the kafka service, it
// should be redefined to match the needs of the service.
// Returns:
// - ch: a interface{} channel to be able to send things through the kafka topic
// generated.
func kafkaStart() (chUDR, chEvents chan interface{}) {

	l.Trace.Printf("[MAIN] Intializing Kafka\n")

	chUDR = make(chan interface{}, 1000)
	chEvents = make(chan interface{}, 1000)

	handler := kafkaHandlerConf{
		out: []kafkaPackage{
			{
				topic:   cfg.Kafka.TopicUDR,
				channel: chUDR,
			},
			{
				topic:   cfg.Kafka.TopicEEngine,
				channel: chEvents,
			},
		},
	}

	kafkaHandler(handler)

	return

}

// report handles the process of sending the event or usage to the respective
// service.
// Parameters:
// - object: an interface{} reference with the event/usage to be sent.
func report(object interface{}) {

	l.Trace.Printf("[REPORT] The reporting process has been started.\n")

	if reflect.TypeOf(object) == reflect.TypeOf(udrModels.Usage{}) {

		l.Trace.Printf("[REPORT] UDR Object detected. Sending through kafka.\n")

		pipeU <- object

		return

	}

	if reflect.TypeOf(object) == reflect.TypeOf(eeModels.Event{}) {

		l.Trace.Printf("[REPORT] Event Object detected. Sending through kafka.\n")

		pipeE <- object

		return

	}

	fail := "the provided object doesn't belong to UDR or EE models"

	l.Warning.Printf("[REPORT] Something went wrong while processing the object, check with the administrator. Error: %v.\n", fail)

	return

}

func init() {

	confFile := flag.String("conf", "./config", "configuration file path (without toml extension)")

	flag.Parse()

	//placeholder code as the default value will ensure this situation will never arise
	if len(*confFile) == 0 {

		fmt.Printf("Usage: Collector-TYPE -conf=/path/to/configuration/file\n")

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
		fmt.Printf("[MAIN] Failed to parse configuration data: %s\nCorrect usage: Collector-TYPE -conf=/path/to/configuration/file\n", err)

		os.Exit(1)

	}

	cfg = parseConfig()

	e := l.InitLogger(cfg.General.LogFile, cfg.General.LogLevel, cfg.General.LogToConsole)

	if e != nil {

		fmt.Printf("[MAIN] Initialization of the logger failed. Error: %v\n", e)

	}

	l.Info.Printf("Cyclops Labs Collector TYPE version %v initialized\n", version)

	dumpConfig(cfg)

	// Let's start the HTTP Server and Gauges for Prometheus
	prometheusStart()

}

func main() {

	// If needed here is the initialization for the kafka sender:
	pipeU, pipeE = kafkaStart()

	// Here we start the client instantiation to send reports to the EventsEngine.
	eeConfig := eeClient.Config{
		URL: &url.URL{
			Host:   cfg.Services["eventsengine"],
			Path:   eeClient.DefaultBasePath,
			Scheme: "http",
		},
		AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
	}

	reportClient = eeClient.New(eeConfig)

	// Here we start the client instantiation to get the canceled customers to check for zombies.
	cusConfig := cusClient.Config{
		URL: &url.URL{
			Host:   cfg.Services["customerdb"],
			Path:   cusClient.DefaultBasePath,
			Scheme: "http",
		},
		AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
	}

	zombiesClient = cusClient.New(cusConfig)

	// Let's lunch the first collection process..
	go collect()

	// cfg.General.Periodicity should be changed to cfg.General.ObjectPeriodicity
	// in the case you need the long (8h) periodicity.
	for range time.NewTicker(time.Duration(cfg.General.Periodicity) * time.Minute).C {

		go collect()

	}

}

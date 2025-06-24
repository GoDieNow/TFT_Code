package main

import (
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	l "gitlab.com/cyclops-utilities/logging"
)

// The following structs: apikey, dbConfig, eventsConfig, generalConfig,
// kafkaConfig, and keycloakConfig are part of the configuration struct which
// acts as the main reference for configuration parameters in the system.
type apiKey struct {
	Enabled bool `json:"enabled"`
	Key     string
	Place   string
	Token   string `json:"token"`
}

type configuration struct {
	APIKey       apiKey
	DB           dbConfig
	Events       eventsConfig
	General      generalConfig
	Kafka        kafkaConfig
	Keycloak     keycloakConfig `json:"keycloak"`
	DefaultPlans map[string]string
	Prometheus   prometheusConfig
}

type dbConfig struct {
	CacheRetention string
	DbName         string
	Host           string
	Password       string
	Port           int
	SSLMode        string
	Username       string
}

type eventsConfig struct {
	Filters []string
}

type generalConfig struct {
	CertificateFile    string `json:"certificate_file"`
	CertificateKey     string `json:"certificate_key"`
	CORSEnabled        bool
	CORSHeaders        []string
	CORSMethods        []string
	CORSOrigins        []string
	HTTPSEnabled       bool
	InsecureSkipVerify bool
	LogFile            string
	LogLevel           string
	LogToConsole       bool
	ServerPort         int
	Services           map[string]string
}

type kafkaConfig struct {
	Brokers    []string
	MaxBytes   int
	MinBytes   int
	Offset     int64
	Partition  int
	TopicsIn   []string
	TopicsOut  []string
	TLSEnabled bool
}

type keycloakConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Enabled      bool   `json:"enabled"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Realm        string `json:"realm"`
	RedirectURL  string `json:"redirect_url"`
	UseHTTP      bool   `json:"use_http"`
}

type planConfig struct {
	Default string
}
type prometheusConfig struct {
	Host          string
	MetricsExport bool
	MetricsPort   string
	MetricsRoute  string
}

// dumpConfig 's job is to dumps the configuration in JSON format to the log
// system. It makes use of the masking function to keep some secrecy in the log.
// Parameters:
// - c: configuration type containing the config present in the system.
func dumpConfig(c configuration) {
	cfgCopy := c

	// deal with configuration params that should be masked
	cfgCopy.APIKey.Token = masked(c.APIKey.Token, 4)
	cfgCopy.DB.Password = masked(c.DB.Password, 4)
	cfgCopy.Keycloak.ClientSecret = masked(c.Keycloak.ClientSecret, 4)

	// mmrshalindent creates a string containing newlines; each line starts with
	// two spaces and two spaces are added for each indent...
	configJSON, _ := json.MarshalIndent(cfgCopy, "  ", "  ")

	l.Info.Printf("[CONFIG] Configuration settings:\n")

	l.Info.Printf("%v\n", string(configJSON))

}

// masked 's job is to return asterisks in place of the characters in a
// string with the exception of the last indicated.
// Parameters:
// - s: string to be masked
// - unmaskedChars: int with the amount (counting from the end of the string) of
// characters to keep unmasked.
// Returns:
// - returnString: the s string passed as parameter masked.
func masked(s string, unmaskedChars int) (returnString string) {

	if len(s) <= unmaskedChars {

		returnString = s

		return

	}

	asteriskString := strings.Repeat("*", (len(s) - unmaskedChars))

	returnString = asteriskString + string(s[len(s)-unmaskedChars:])

	return

}

// parseConfig handles the filling of the config struct with the data Viper gets
// from the configuration file.
// Returns:
// - c: the configuration struct filled with the relevant parsed configuration.
func parseConfig() (c configuration) {

	l.Trace.Printf("[CONFIG] Retrieving configuration.\n")

	c = configuration{

		APIKey: apiKey{
			Enabled: viper.GetBool("apikey.enabled"),
			Key:     viper.GetString("apikey.key"),
			Place:   viper.GetString("apikey.place"),
			Token:   viper.GetString("apikey.token"),
		},

		DB: dbConfig{
			CacheRetention: viper.GetString("database.cacheretention"),
			DbName:         viper.GetString("database.dbname"),
			Host:           viper.GetString("database.host"),
			Password:       viper.GetString("database.password"),
			Port:           viper.GetInt("database.port"),
			SSLMode:        viper.GetString("database.sslmode"),
			Username:       viper.GetString("database.username"),
		},

		Events: eventsConfig{
			Filters: viper.GetStringSlice("events.filters"),
		},

		General: generalConfig{
			CertificateFile:    viper.GetString("general.certificatefile"),
			CertificateKey:     viper.GetString("general.certificatekey"),
			CORSEnabled:        viper.GetBool("general.corsenabled"),
			CORSHeaders:        viper.GetStringSlice("general.corsheaders"),
			CORSMethods:        viper.GetStringSlice("general.corsmethods"),
			CORSOrigins:        viper.GetStringSlice("general.corsorigins"),
			HTTPSEnabled:       viper.GetBool("general.httpsenabled"),
			InsecureSkipVerify: viper.GetBool("general.insecureskipverify"),
			LogFile:            viper.GetString("general.logfile"),
			LogLevel:           viper.GetString("general.loglevel"),
			LogToConsole:       viper.GetBool("general.logtoconsole"),
			ServerPort:         viper.GetInt("general.serverport"),
			Services:           viper.GetStringMapString("general.services"),
		},

		Kafka: kafkaConfig{
			Brokers:    viper.GetStringSlice("kafka.brokers"),
			MaxBytes:   viper.GetInt("kafka.sizemax"),
			MinBytes:   viper.GetInt("kafka.sizemin"),
			Offset:     viper.GetInt64("kafka.offset"),
			Partition:  viper.GetInt("kafka.partition"),
			TopicsIn:   viper.GetStringSlice("kafka." + service + "in"),
			TopicsOut:  viper.GetStringSlice("kafka." + service + "out"),
			TLSEnabled: viper.GetBool("kafka.tlsenabled"),
		},

		Keycloak: keycloakConfig{
			ClientID:     viper.GetString("keycloak.clientid"),
			ClientSecret: viper.GetString("keycloak.clientsecret"),
			Enabled:      viper.GetBool("keycloak.enabled"),
			Host:         viper.GetString("keycloak.host"),
			Port:         viper.GetInt("keycloak.port"),
			Realm:        viper.GetString("keycloak.realm"),
			RedirectURL:  viper.GetString("keycloak.redirecturl"),
			UseHTTP:      viper.GetBool("keycloak.usehttp"),
		},

		DefaultPlans: viper.GetStringMapString("plans"),

		Prometheus: prometheusConfig{
			Host:          viper.GetString("prometheus.host"),
			MetricsExport: viper.GetBool("prometheus.metricsexport"),
			MetricsPort:   viper.GetString("prometheus.metricsport"),
			MetricsRoute:  viper.GetString("prometheus.metricsroute"),
		},
	}

	return

}

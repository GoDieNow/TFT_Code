package main

import (
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
	l "gitlab.com/cyclops-utilities/logging"
)

// The following structs are part of the configuration struct which
// acts as the main reference for configuration parameters in the system.
type apiKey struct {
	Enabled bool
	Key     string
	Place   string
	Token   string
}

type configuration struct {
	APIKey         apiKey
	General        generalConfig
	Heappe         heappeConfig
	Kafka          kafkaConfig
	Keycloak       keycloakConfig
	Lieutenant     lieutenantConfig
	OpenStack      openStackConfig
	Prometheus     prometheusConfig
	RGW            rgwConfig
	NameFilters    []string
	ProjectFilters []string
	Services       map[string]string
}

type generalConfig struct {
	InsecureSkipVerify    bool
	LogFile               string
	LogLevel              string
	LogToConsole          bool
	ObjectsPeriodicity    int
	Periodicity           int
	PrometheusPeriodicity int
}

type heappeConfig struct {
	Username                    string
	Password                    string
	GroupResourceUsageReportURI string
	AuthenticateUserPasswordURI string
}

type kafkaConfig struct {
	Brokers      []string
	MaxBytes     int
	MinBytes     int
	Offset       int64
	Partition    int
	TLSEnabled   bool
	TopicUDR     string
	TopicEEngine string
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

type lieutenantConfig struct {
	Host  string
	Token string
}

type openStackConfig struct {
	Domain   string
	Keystone string
	Password string
	Project  string
	Region   string
	User     string
}

type prometheusConfig struct {
	Host          string
	MetricsExport bool
	MetricsPort   string
	MetricsRoute  string
}

type rgwConfig struct {
	AccessKeyID     string
	AdminPath       string
	Region          string
	SecretAccessKey string
	ServerURL       string
}

// dumpConfig 's job is to dumps the configuration in JSON format to the log
// system. It makes use of the masking function to keep some secrecy in the log.
// Parameters:
// - c: configuration type containing the config present in the system.
func dumpConfig(c configuration) {
	cfgCopy := c

	// deal with configuration params that should be masked
	cfgCopy.APIKey.Token = masked(c.APIKey.Token, 4)
	cfgCopy.Heappe.Username = masked(c.Heappe.Username, 4)
	cfgCopy.Heappe.Password = masked(c.Heappe.Password, 4)
	cfgCopy.Keycloak.ClientSecret = masked(c.Keycloak.ClientSecret, 4)
	cfgCopy.Lieutenant.Token = masked(c.Lieutenant.Token, 4)
	cfgCopy.OpenStack.Password = masked(c.OpenStack.Password, 4)
	cfgCopy.RGW.AccessKeyID = masked(c.RGW.AccessKeyID, 4)
	cfgCopy.RGW.SecretAccessKey = masked(c.RGW.SecretAccessKey, 4)

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

		General: generalConfig{
			InsecureSkipVerify:    viper.GetBool("general.insecureskipverify"),
			LogFile:               viper.GetString("general.logfile"),
			LogLevel:              viper.GetString("general.loglevel"),
			LogToConsole:          viper.GetBool("general.logtoconsole"),
			ObjectsPeriodicity:    viper.GetInt("general.objectsperiodicity"),
			Periodicity:           viper.GetInt("general.periodicity"),
			PrometheusPeriodicity: viper.GetInt("general.prometheusperiodicity"),
		},

		Heappe: heappeConfig{
			Username:                    viper.GetString("heappe.username"),
			Password:                    viper.GetString("heappe.password"),
			GroupResourceUsageReportURI: viper.GetString("heappe.groupResourceUsageReportUri"),
			AuthenticateUserPasswordURI: viper.GetString("heappe.authenticateUserPasswordUri"),
		},

		Kafka: kafkaConfig{
			Brokers:      viper.GetStringSlice("kafka.brokers"),
			MaxBytes:     viper.GetInt("kafka.sizemax"),
			MinBytes:     viper.GetInt("kafka.sizemin"),
			Offset:       viper.GetInt64("kafka.offset"),
			Partition:    viper.GetInt("kafka.partition"),
			TLSEnabled:   viper.GetBool("kafka.tlsenabled"),
			TopicUDR:     viper.GetString("kafka.topicudr"),
			TopicEEngine: viper.GetString("kafka.topiceengine"),
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

		Lieutenant: lieutenantConfig{
			Host:  viper.GetString("lieutenant.host"),
			Token: viper.GetString("lieutenant.token"),
		},

		OpenStack: openStackConfig{
			Domain:   viper.GetString("openstack.domain"),
			Keystone: viper.GetString("openstack.keystone"),
			Password: viper.GetString("openstack.password"),
			Project:  viper.GetString("openstack.project"),
			Region:   viper.GetString("openstack.region"),
			User:     viper.GetString("openstack.user"),
		},

		Prometheus: prometheusConfig{
			Host:          viper.GetString("prometheus.host"),
			MetricsExport: viper.GetBool("prometheus.metricsexport"),
			MetricsPort:   viper.GetString("prometheus.metricsport"),
			MetricsRoute:  viper.GetString("prometheus.metricsroute"),
		},

		RGW: rgwConfig{
			AccessKeyID:     viper.GetString("rgw.accesskey"),
			AdminPath:       viper.GetString("rgw.adminpath"),
			Region:          viper.GetString("rgw.region"),
			SecretAccessKey: viper.GetString("rgw.secretaccesskey"),
			ServerURL:       viper.GetString("rgw.serverurl"),
		},

		NameFilters:    viper.GetStringSlice("events.namefilters"),
		ProjectFilters: viper.GetStringSlice("events.projectfilters"),
		Services:       viper.GetStringMapString("services"),
	}

	return

}

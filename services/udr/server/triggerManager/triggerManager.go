package triggerManager

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	"gitlab.com/cyclops-utilities/datamodels"
	eventsClient "github.com/GoDieNow/TFT_Code/services/eventsengine/client"
	eventsUsage "github.com/GoDieNow/TFT_Code/services/eventsengine/client/usage_management"
	eeModels "github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	"github.com/GoDieNow/TFT_Code/services/udr/models"
	"github.com/GoDieNow/TFT_Code/services/udr/restapi/operations/trigger_management"
	"github.com/GoDieNow/TFT_Code/services/udr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/udr/server/statusManager"
	l "gitlab.com/cyclops-utilities/logging"
)

// TriggerManager is the struct defined to group and contain all the methods
// that interact with the trigger subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - eventsConfig: a EventsEngine client config reference to interact with EventsEngine
type TriggerManager struct {
	pipe         chan interface{}
	db           *dbManager.DbParameter
	monit        *statusManager.StatusManager
	eventsConfig eventsClient.Config
}

type CompactionKey struct {
	Account  string
	ID       string
	Name     string
	Metadata string
	Metric   string
	Unit     string
}

const (
	scaler = 1e7
)

// New is the function to create the struct TriggerManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the TriggerManager methods.
// - eventsClient.Config: a reference to the config needed to start the client
// interface to be able to interact with the EventsEngine service.
// Returns:
// - TriggerManager: struct to interact with triggerManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, c eventsClient.Config, ch chan interface{}) *TriggerManager {

	l.Trace.Printf("[TriggerManager] Generating new TriggerManager.\n")

	monit.InitEndpoint("trigger")

	return &TriggerManager{
		pipe:         ch,
		db:           db,
		monit:        monit,
		eventsConfig: c,
	}

}

func (m *TriggerManager) getClient(param *http.Request) *eventsClient.EventEngineManagementAPI {

	config := m.eventsConfig

	if len(param.Header.Get("Authorization")) > 0 {

		token := strings.Fields(param.Header.Get("Authorization"))[1]

		config.AuthInfo = httptransport.BearerToken(token)

	}

	r := eventsClient.New(config)

	return r

}

func (m *TriggerManager) getNiceFloat(i float64) (o float64) {

	return float64(math.Round(i*scaler) / scaler)

}

// ExecCompactation (Swagger func) is the function behind the (GET) endpoint
// /trigger/compact
// Its job is to generate a compactation of the information in the database
// by account, metrics and elements.
// It also extracts the usage from the EventsEngine and consolidates it with
// the UDRs already present in the system.
func (m *TriggerManager) ExecCompactation(ctx context.Context, params trigger_management.ExecCompactationParams) middleware.Responder {

	l.Trace.Printf("[TriggerManager] Compact endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("trigger", callTime)

	var from, to strfmt.DateTime
	var accumErr []string
	var udrCount int64
	var interval float64

	queue := make(chan string, 1)

	now := time.Now()
	records := make(map[string][]*models.Usage)

	if params.From != nil && params.To != nil {

		l.Trace.Printf("[TriggerManager] Preriod for compactation provided by params: [ %v ] to [ %v ].\n", params.From, params.To)

		from = *params.From
		to = *params.To

	} else {

		if params.FastMode != nil && *params.FastMode {

			switch now.Minute() {

			case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15:

				f := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, time.UTC)
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC))

				from = strfmt.DateTime(f.Add(time.Minute * -60))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: HH(-1):45-HH:00 in day(-1?).\n")

			case 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: HH:00-HH:15.\n")

			case 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: HH:15-HH:30.\n")

			case 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 0:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: HH:30-HH:45.\n")

			}

		} else {

			switch now.Hour() {

			case 0, 1, 2, 3, 4, 5, 6, 7:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day()-1, 16, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: 16:00-24:00 in day -1.\n")

			case 8, 9, 10, 11, 12, 13, 14, 15:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: 00:00-08:00.\n")

			case 16, 17, 18, 19, 20, 21, 22, 23:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Preriod for compactation: 08:00-16:00.\n")

			}

		}

	}

	interval = float64(((time.Time)(to)).Sub((time.Time)(from)).Seconds())

	// First we clean all the records so we always have a clean UDR generation
	templateRecord := models.UDRRecord{
		TimeTo:   to,
		TimeFrom: from,
	}

	l.Info.Printf("[TriggerManager] Clean up of previous UDRRecords started for period [ %v ] - [ %v ].\n", from, to)

	if e := m.db.Db.Where(&templateRecord).Delete(&models.UDRRecord{}).Error; e != nil {

		l.Warning.Printf("[TriggerManager] The clean up of previous UDRecords failed. Error: %v\n", e)

	}

	templateReport := models.UReport{
		TimeFrom: from,
		TimeTo:   to,
	}

	l.Info.Printf("[TriggerManager] Clean up of previous UReports started for period [ %v ] - [ %v ].\n", from, to)

	if e := m.db.Db.Where(&templateReport).Delete(&models.UReport{}).Error; e != nil {

		l.Warning.Printf("[TriggerManager] The clean up of previous UReports failed. Error: %v\n", e)

	}

	// Handling all the error in the same place
	go func() {

		for t := range queue {
			accumErr = append(accumErr, t)
		}

	}()

	// First we get the metric and accounts in the system.
	metrics, _ := m.db.GetMetrics()
	accounts := m.db.GetAccounts()

	// Then, we get all the records by metrics
	for _, metric := range metrics {

		records[metric.Metric] = m.db.GetRecords(metric.Metric, from, to)

	}

	swg := sizedwaitgroup.New(8)
	// Goroutines start
	swg.Add()
	go func() {

		defer swg.Done()

		// And finally, we compact the records by account
		// Using the metadata as matching-key
		for i := range accounts {

			swg.Add()
			go func(acc string) {

				account := acc
				defer swg.Done()

				for _, record := range records {

					accum := make(map[CompactionKey]float64)
					counter := make(map[CompactionKey]float64)
					metas := make(map[CompactionKey]datamodels.JSONdb)

					// In the first part we calculate the usage by metadata
					// And create a map of metadatas/matching keys
					for id := range record {

						if record[id].Account == account {

							key := CompactionKey{
								Account:  record[id].Account,
								ID:       record[id].ResourceID,
								Name:     record[id].ResourceName,
								Metadata: fmt.Sprintf("%v", record[id].Metadata),
								Metric:   record[id].ResourceType,
								Unit:     record[id].Unit,
							}

							accum[key] += record[id].Usage
							metas[key] = record[id].Metadata
							counter[key]++

						}

					}

					// In the second part we add the compacted records in the system
					for key := range accum {

						use := datamodels.JSONdb{"used": accum[key] * interval / counter[key]}
						unit := key.Unit + "(*period)"

						if v, exists := metas[key]["UDRMode"].(string); exists {

							switch v {

							case "avg":

								use = datamodels.JSONdb{"used": accum[key] / counter[key]}
								unit = key.Unit

							case "sum":

								use = datamodels.JSONdb{"used": accum[key]}
								unit = key.Unit

							default:

								use = datamodels.JSONdb{"used": accum[key] * interval / counter[key]}
								unit = key.Unit + "(*period)"

							}

						}

						report := models.UDRRecord{
							AccountID:    key.Account,
							UsageBreakup: use,
							Metadata:     metas[key],
							TimeTo:       to,
							TimeFrom:     from,
							ResourceID:   key.ID,
							ResourceName: key.Name,
							ResourceType: key.Metric,
							Unit:         unit,
						}

						if e := m.db.AddRecord(report); e != nil {

							err := "Acc: " + account + "-" + report.ResourceID + ". Error: " + e.Error()
							queue <- err

							l.Trace.Printf("[TriggerManager] There was a problem adding the record in the system. Error: %v.\n", e)

						} else {

							l.Trace.Printf("[TriggerManager] Compacted record for account [ %v ] added to the system.\n", account)

						}

					}

				}

				r := models.UReport{
					AccountID: account,
					TimeFrom:  from,
					TimeTo:    to,
				}

				if e := m.db.AddReport(r); e != nil {

					err := "Acc(r): " + account + ". Error: " + e.Error()
					queue <- err

					l.Trace.Printf("[TriggerManager] There was a problem adding the Report in the system. Error: %v.\n", e)

				} else {

					udrCount++

					l.Trace.Printf("[TriggerManager] Compacted report for account [ %v ] added to the system.\n", account)

				}

			}(accounts[i])

		}

	}()

	// EE import and compactation
	client := m.getClient(params.HTTPRequest)

	fu := ((time.Time)(from)).Unix()
	tu := ((time.Time)(to)).Unix()

	eeParams := eventsUsage.NewGetSystemUsageParams().WithFrom(&fu).WithTo(&tu)

	r, e := client.UsageManagement.GetSystemUsage(ctx, eeParams)

	if e != nil {

		err := "EE(g)-Error: " + e.Error()
		queue <- err

		l.Warning.Printf("[TriggerManager] There was a problem while importing data from EventsEngine. Error: %v", e)

	} else {

		for i := range r.Payload {

			swg.Add()
			go func(ac *eeModels.Usage) {

				defer swg.Done()

				acc := *ac

				accum := make(map[CompactionKey]datamodels.JSONdb)
				metas := make(map[CompactionKey]datamodels.JSONdb)

				// In the first part we calculate the usage by metadata
				// And create a map of metadatas/matching keys
				for id := range acc.Usage {

					key := CompactionKey{
						Account:  acc.AccountID,
						ID:       acc.Usage[id].ResourceID,
						Name:     acc.Usage[id].ResourceName,
						Metadata: fmt.Sprintf("%v", acc.Usage[id].MetaData),
						Metric:   acc.Usage[id].ResourceType,
						Unit:     acc.Usage[id].Unit,
					}

					accum[key] = make(datamodels.JSONdb)
					metas[key] = acc.Usage[id].MetaData

					for i := range acc.Usage[id].UsageBreakup {

						if _, exists := accum[key][i]; !exists {

							accum[key][i] = int64(0)

						}

						addition, _ := acc.Usage[id].UsageBreakup[i].(json.Number).Int64()

						accum[key][i] = accum[key][i].(int64) + addition

					}

				}

				// In the second part we add the compacted records in the system
				for key := range accum {

					report := models.UDRRecord{
						AccountID:    key.Account,
						UsageBreakup: accum[key],
						Metadata:     metas[key],
						TimeTo:       to,
						TimeFrom:     from,
						ResourceID:   key.ID,
						ResourceName: key.Name,
						ResourceType: key.Metric,
						Unit:         key.Unit,
					}

					if e := m.db.AddRecord(report); e != nil {

						err := "Acc: " + report.AccountID + "-" + report.ResourceID + ". Error: " + e.Error()
						queue <- err

						l.Trace.Printf("[TriggerManager] There was a problem adding the record in the system. Error: %v.\n", e)

					} else {

						l.Trace.Printf("[TriggerManager] Compacted record for account [ %v ] added to the system.\n", report.AccountID)

					}

				}

				r := models.UReport{
					AccountID: acc.AccountID,
					TimeFrom:  from,
					TimeTo:    to,
				}

				if e := m.db.AddReport(r); e != nil {

					err := "Acc(r): " + r.AccountID + ". Error: " + e.Error()
					queue <- err

					l.Trace.Printf("[TriggerManager] There was a problem adding the Report in the system. Error: %v.\n", e)

				} else {

					udrCount++

					l.Trace.Printf("[TriggerManager] Compacted report for account [ %v ] added to the system.\n", r.AccountID)

				}

			}(r.Payload[i])

		}

	}

	swg.Wait()

	// Kafka report
	accs := m.db.GetUDRAccounts()

	for _, acc := range accs {

		usages, e := m.db.GetUsage(acc, "", from, to)

		if e == nil {

			l.Trace.Printf("[TriggerManager] There's [ %v ] compacted usages associated with the account [ %v ].\n", len(usages), acc)

			for _, usage := range usages {

				m.pipe <- *usage

			}

		} else {

			err := "Usage(g)-Acc: " + acc + ". Error: " + e.Error()
			queue <- err

			l.Warning.Printf("[TriggerManager] There was a problem while retrieving the compacted usages to be sent through Kafka. Error: %v", e)

		}

	}

	close(queue)

	m.db.Metrics["time"].With(prometheus.Labels{"type": "UDRs generation time"}).Set(float64(time.Now().UnixNano()-now.UnixNano()) / float64(time.Millisecond))

	m.db.Metrics["count"].With(prometheus.Labels{"type": "UDRs generated"}).Set(float64(udrCount))

	m.db.Metrics["count"].With(prometheus.Labels{"type": "Errors during UDR generation"}).Set(float64(len(accumErr)))

	if len(accumErr) > 0 {

		s := "This are the accumulative errors: " + strings.Join(accumErr, ",\n")
		errorReturn := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/trigger/compact"}).Inc()

		m.monit.APIHitDone("trigger", callTime)

		return trigger_management.NewExecCompactationInternalServerError().WithPayload(&errorReturn)

	}

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/trigger/compact"}).Inc()

	m.monit.APIHitDone("trigger", callTime)

	return trigger_management.NewExecCompactationOK()

}

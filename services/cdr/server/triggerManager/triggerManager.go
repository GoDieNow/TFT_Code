package triggerManager

import (
	"context"
	"net/http"
	"strings"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	"github.com/GoDieNow/TFT_Code/services/cdr/models"
	"github.com/GoDieNow/TFT_Code/services/cdr/restapi/operations/trigger_management"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/dbManager"
	"github.com/GoDieNow/TFT_Code/services/cdr/server/statusManager"
	udrClient "github.com/GoDieNow/TFT_Code/services/udr/client"
	udrUsage "github.com/GoDieNow/TFT_Code/services/udr/client/usage_management"
	udrModels "github.com/GoDieNow/TFT_Code/services/udr/models"
	l "gitlab.com/cyclops-utilities/logging"
)

// TriggerManager is the struct defined to group and contain all the methods
// that interact with the trigger subsystem.
// Parameters:
// - db: a DbParameter reference to be able to use the DBManager methods.
// - monit: a StatusManager reference to be able to use the status subsystem methods.
// - udrConfig: a UDR client config reference to interact with UDR service.
type TriggerManager struct {
	db        *dbManager.DbParameter
	monit     *statusManager.StatusManager
	udrConfig udrClient.Config
}

// New is the function to create the struct TriggerManager.
// Parameters:
// - DbParameter: reference pointing to the DbParameter that allows the interaction
// with the DBManager methods.
// - StatusParameter: reference poining to the StatusManager that allows the
// interaction with the StatusManager methods.
// Returns:
// - TriggerManager: struct to interact with TriggerManager subsystem functionalities.
func New(db *dbManager.DbParameter, monit *statusManager.StatusManager, c udrClient.Config) *TriggerManager {

	l.Trace.Printf("[TriggerManager] Generating new triggerManager.\n")

	monit.InitEndpoint("trigger")

	return &TriggerManager{
		db:        db,
		monit:     monit,
		udrConfig: c,
	}

}

func (m *TriggerManager) getClient(param *http.Request) *udrClient.UDRManagementAPI {

	config := m.udrConfig

	if len(param.Header.Get("Authorization")) > 0 {

		token := strings.Fields(param.Header.Get("Authorization"))[1]

		config.AuthInfo = httptransport.BearerToken(token)

	}

	u := udrClient.New(config)

	return u

}

// ExecTransformation (Swagger func) is the function behind the
// /trigger/execTransformation endpoint.
// It function is to serve as a backup functionality to trigger the manual
// transformation of the UDR into CDRs.
func (m *TriggerManager) ExecTransformation(ctx context.Context, params trigger_management.ExecTransformationParams) middleware.Responder {

	l.Trace.Printf("[Trigger] ExecSample endpoint invoked.\n")

	callTime := time.Now()
	m.monit.APIHit("trigger", callTime)

	var from, to strfmt.DateTime
	var now time.Time
	var cdrCount int64

	if params.From != nil && params.To != nil {

		l.Trace.Printf("[TriggerManager] Preriod for compactation provided by params: [ %v ] to [ %v ].\n", params.From, params.To)

		from = *params.From
		to = *params.To

	} else {

		if params.FastMode != nil && *params.FastMode {

			switch now = time.Now(); now.Minute() {

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

			switch now = time.Now(); now.Hour() {

			case 0, 1, 2, 3, 4, 5, 6, 7:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day()-1, 16, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Period for transformation: 16:00-24:00 in day -1.\n")

			case 8, 9, 10, 11, 12, 13, 14, 15:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Period for transformation: 00:00-08:00.\n")

			case 16, 17, 18, 19, 20, 21, 22, 23:

				from = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.UTC))
				to = strfmt.DateTime(time.Date(now.Year(), now.Month(), now.Day(), 16, 0, 0, 0, time.UTC))

				l.Trace.Printf("[TriggerManager] Period for transformation: 08:00-16:00.\n")

			}

		}

	}

	udrClient := m.getClient(params.HTTPRequest)

	udrParams := udrUsage.NewGetSystemUsageParams().WithFrom(&from).WithTo(&to)

	r, e := udrClient.UsageManagement.GetSystemUsage(ctx, udrParams)

	if e != nil {

		l.Warning.Printf("[TriggerManager] There was a problem while importing data from UDR. Error: %v", e)

		s := "UDRs import failed: " + e.Error()
		returnValueError := models.ErrorResponse{
			ErrorString: &s,
		}

		m.db.Metrics["api"].With(prometheus.Labels{"code": "500", "method": "GET", "route": "/trigger/execTransformation"}).Inc()

		m.monit.APIHitDone("trigger", callTime)

		return trigger_management.NewExecTransformationInternalServerError().WithPayload(&returnValueError)

	}

	swg := sizedwaitgroup.New(8)

	for id := range r.Payload {

		// Goroutines start
		swg.Add()
		go func(udrIN *udrModels.UReport) {

			udr := *udrIN
			defer swg.Done()

			token := ""

			if len(params.HTTPRequest.Header.Get("Authorization")) > 0 {

				token = strings.Fields(params.HTTPRequest.Header.Get("Authorization"))[1]

			}

			e := m.db.ProcessUDR(udr, token)

			if e != nil {

				m.db.Metrics["count"].With(prometheus.Labels{"type": "UDR processing FAILED"}).Inc()

				m.db.Metrics["count"].With(prometheus.Labels{"type": "UDR processing FAILED for Account: " + r.Payload[id].AccountID}).Inc()

				l.Warning.Printf("[TriggerManager] There was a problem while processing the UDR Report, skipping this one. Error: %v", e)

			} else {

				cdrCount++

			}

		}(r.Payload[id])

	}

	swg.Wait()

	m.db.Metrics["time"].With(prometheus.Labels{"type": "CDRs triggered generation time"}).Set(float64(callTime.UnixNano()-now.UnixNano()) / float64(time.Millisecond))

	m.db.Metrics["count"].With(prometheus.Labels{"type": "UDRs transformed to CDRs"}).Set(float64(cdrCount))

	m.db.Metrics["api"].With(prometheus.Labels{"code": "200", "method": "GET", "route": "/trigger/execTransformation"}).Inc()

	m.monit.APIHitDone("trigger", callTime)

	return trigger_management.NewExecTransformationOK()

}

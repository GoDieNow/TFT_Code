package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/prometheus/client_golang/prometheus"
	datamodels "gitlab.com/cyclops-utilities/datamodels"
	eeEvent "github.com/GoDieNow/TFT_Code/services/eventsengine/client/event_management"
	eeModels "github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	collector       = "Network"
	objects         = "ip"
	remotelist      []*eeModels.MinimalState
	collectionStart int64
)

// collect handles the process of retrieving the information from the system.
func collect() {

	l.Trace.Printf("[COLLECTION] The collection process has been started.\n")

	collectionStart = time.Now().UnixNano()

	metricTime.With(prometheus.Labels{"type": "Collection Start Time"}).Set(float64(collectionStart))

	l.Trace.Printf("[COLLECTION] Found ...")

	// Here comes the transformation of the information retrieved into either
	md := make(datamodels.JSONdb)
	evTime := int64(time.Now().Unix())
	// events or usage reports to be sent.
	event := ...
	report(event)


	l.Trace.Printf("[COLLECTION] The collection process has been finished.\n")

	return

}

// getStatus job is to normalize the event state returned by the collectors.
// Parameters:
// - state: string returned by the system.
// Returns:
// - status: normalized state to be returned.
func getStatus(state string) (status string) {

	switch strings.ToUpper(state) {

	case "ACTIVE":

		status = "active"

	case "ATTACHING":

		status = "active"

	case "AVAILABLE":

		status = "active"

	case "BUILD":

		status = "inactive"

	case "CREATING":

		status = "active"

	case "DELETED":

		status = "terminated"

	case "DELETING":

		status = "terminated"

	case "DETACHING":

		status = "active"

	case "DOWN":

		status = "inactive"

	case "ERROR":

		status = "error"

	case "ERROR_DELETING":

		status = "error"

	case "EXTENDING":

		status = "inactive"

	case "HARD_DELETED":

		status = "terminated"

	case "IN-USE":

		status = "active"

	case "MAINTENANCE":

		status = "active"

	case "PAUSED":

		status = "inactive"

	case "RESCUED":

		status = "active"

	case "RESIZE":

		status = "active"

	case "RESIZED":

		status = "active"

	case "RESERVED":

		status = "active"

	case "RETYPING":

		status = "inactive"

	case "SHUTOFF":

		status = "inactive"

	case "SOFT_DELETED":

		status = "terminated"

	case "STOPPED":

		status = "inactive"

	case "SUSPENDED":

		status = "inactive"

	case "TERMINATED":

		status = "terminated"

	case "VERIFY_RESIZE":

		status = "active"

	}

	l.Trace.Printf("[REPORT] State received from the system [ %v ] normalized to [ %v ]", state, status)

	return status

}

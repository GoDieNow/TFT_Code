package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/prometheus/client_golang/prometheus"
	datamodels "gitlab.com/cyclops-utilities/datamodels"
	eeEvent "github.com/GoDieNow/TFT_Code/services/eventsengine/client/event_management"
	eeModels "github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	collector       = "Servers"
	objects         = "vm"
	collectionStart int64
	vmCount         int
	terminatedCount int
	dropCount       int
	remotelist      []*eeModels.MinimalState
	client          *gophercloud.ServiceClient
)

// collect handles the process of retrieving the information from the system.
func collect() {

	l.Trace.Printf("[COLLECTION] The collection process has been started.\n")

	collectionStart = time.Now().UnixNano()

	metricTime.With(prometheus.Labels{"type": "Collection Start Time"}).Set(float64(collectionStart))

	// Here comes the logic to retrieve the information from the system.
	opts := gophercloud.AuthOptions{
		DomainName:       cfg.OpenStack.Domain,
		IdentityEndpoint: cfg.OpenStack.Keystone,
		Password:         cfg.OpenStack.Password,
		Username:         cfg.OpenStack.User,
	}

	if len(cfg.OpenStack.Project) > 0 {

		opts.TenantName = cfg.OpenStack.Project

	}

	provider, e := openstack.AuthenticatedClient(opts)

	if e != nil {

		l.Error.Printf("[COLLECTION] Error authenticating against OpenStack. Error: %v", e)

		os.Exit(1)

	}

	client, e = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: cfg.OpenStack.Region,
	})

	if e != nil {

		l.Error.Printf("[COLLECTION] Error creating compute collector client. Error: %v", e)

		os.Exit(1)

	}

	serveropts := servers.ListOpts{
		AllTenants: true, //set this to true to list VMs from all tenants if policy allows it
	}

	l.Trace.Printf("[COLLECTION] Querying events engine service for list of known and not terminated servers")

	resourceType := "server"
	eeParams := eeEvent.NewListStatesParams().WithResource(&resourceType).WithRegion(&cfg.OpenStack.Region)

	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
	r, e := reportClient.EventManagement.ListStates(ctx, eeParams)

	// Clears the remotelist between runs
	remotelist = nil

	if e != nil {

		l.Warning.Printf("[COLLECTION] Something went wrong while retrieving the usage from the system, check with the administrator. Error: %v.\n", e)

	} else {

		remotelist = r.Payload

	}

	metricCount.With(prometheus.Labels{"type": "Total VMs from EventEngine"}).Set(float64(len(remotelist)))

	l.Trace.Printf("[COLLECTION] (BEFORE) Existing count of servers at remote [ %v ].\n", len(remotelist))

	eeCount := len(remotelist)

	pager := servers.List(client, serveropts)

	vmCount = 0

	e = pager.EachPage(extractPage)

	if e != nil {

		l.Error.Printf("[COLLECTION] Error processing the lists of active resources. Error: %v\n", e)

		os.Exit(1)

	}

	metricCount.With(prometheus.Labels{"type": "Total VMs reported by OS API"}).Set(float64(vmCount))

	metricCount.With(prometheus.Labels{"type": "Total VMs reported by OS API (to terminated state)"}).Set(float64(terminatedCount))

	metricCount.With(prometheus.Labels{"type": "Total VMs DROPPED due to unknown flavor"}).Set(float64(dropCount))

	l.Trace.Printf("[COLLECTION] (AFTER) Remaining count of servers at remote which were left unprocessed [ %v ].\n", len(remotelist))

	//now for all remaining servers send terminated status
	for _, object := range remotelist {

		l.Trace.Printf("[COLLECTION] Sending terminated event for server [ %v ] for project [ %v ] with ID [ %v ].\n", object.ResourceID, object.ResourceName, object.Account)

		evTime := int64(time.Now().Unix())
		evLast := getStatus("terminated")

		// events reports to be sent.
		event := eeModels.Event{
			Account:      object.Account,
			EventTime:    &evTime,
			LastEvent:    &evLast,
			MetaData:     object.MetaData,
			Region:       cfg.OpenStack.Region,
			ResourceID:   object.ResourceID,
			ResourceName: object.ResourceName,
			ResourceType: "server",
		}

		report(event)

	}

	metricCount.With(prometheus.Labels{"type": "Total VMs forcefully TERMINATED"}).Set(float64(len(remotelist)))

	metricTime.With(prometheus.Labels{"type": "Collection Processing Time"}).Set(float64(time.Now().UnixNano()-collectionStart) / float64(time.Millisecond))

	l.Warning.Printf("[COLLECTION] Completed.\n - OS Report: %v\n - EE Report: %v\n - Droped: %v\n - OS Terminated: %v\n - Forced Termination: %v\n - Processing Time: %v[ms]\n", vmCount, eeCount, dropCount, terminatedCount, len(remotelist), float64(time.Now().UnixNano()-collectionStart)/float64(time.Millisecond))

	l.Trace.Printf("[COLLECTION] The collection process has been finished.\n")

	return

}

// extractPage is the handler function invoked to process each page collected
// from the server list.
// Parameters:
// - page: Pagination.Page reference of the page to be processed.
// Returns:
// - ok: a bool to mark the state of the processing.
// - e: an error reference raised in case of something goes wrong.
func extractPage(page pagination.Page) (ok bool, e error) {

	var serverList []servers.Server

	serverList, e = servers.ExtractServers(page)

	if e != nil {

		return

	}

allProjectsLoop:
	for _, s := range serverList {

		// Filter by project id:
		for _, filter := range cfg.ProjectFilters {

			if strings.Contains(s.TenantID, filter) && filter != "" {

				l.Debug.Printf("[COLLECTION] The Project [ %v ] matches filter [ %v ] and won't be further processed.", s.TenantID, filter)

				continue allProjectsLoop

			}

		}

		// Filter by project name:
		for _, filter := range cfg.NameFilters {

			if strings.Contains(strings.ToLower(s.Name), strings.ToLower(filter)) && filter != "" {

				l.Debug.Printf("[COLLECTION] The Project [ %v ] matches filter [ %v ] and won't be further processed.", s.Name, filter)

				continue allProjectsLoop

			}

		}

		vmCount++

		// "s" will be a servers.Server
		var imageid, flavorid, imagename, imageosflavor, flavorname string

		for k, val := range s.Image {

			switch v := val.(type) {

			case string:

				if strings.Compare(k, "id") == 0 {

					imageid = v

				}

			}

		}

		for k, val := range s.Flavor {

			switch v := val.(type) {

			case string:

				if strings.Compare(k, "id") == 0 {

					flavorid = v

				}

			}

		}

		l.Trace.Printf("%+v, %v", client, imageid)

		imagename, imageosflavor, e := getFromImageCache(client, imageid)

		if e != nil {

			l.Error.Printf("[COLLECTION] Error while getting the image id [ %+v ]. Error: %v\n", imageid, e)

		}

		flavorname, e = getFromFlavorCache(client, flavorid)

		if e != nil {

			l.Error.Printf("[COLLECTION] Error while getting the flavor id [ %+v ]. Error: %v\n", flavorid, e)

		}

		if len(flavorname) == 0 {

			l.Warning.Printf("[COLLECTION] Found VM - Name:[%s], TenantID:[%s], Status:[%s], ID:[%s], ImageID:[%s], ImageName:[%s], ImageOSFlavor:[%s], FlavorId:[%s], FlavorName:[%s] :: with missing FlavorName, skipping record!",
				s.Name, s.TenantID, s.Status, s.ID, imageid, imagename, imageosflavor, flavorid, flavorname)

			dropCount++

			continue

		}

		l.Info.Printf("[COLLECTION] Found VM - Name:[%s], TenantID:[%s], Status:[%s], ID:[%s], ImageID:[%s], ImageName:[%s], ImageOSFlavor:[%s], FlavorId:[%s], FlavorName:[%s]",
			s.Name, s.TenantID, s.Status, s.ID, imageid, imagename, imageosflavor, flavorid, flavorname)

		// Potential problem with these filters are if clients create their
		// VMs with the filter strings those will not be billed.
		// It will be better to actually filter out all resources within a given
		// tenant, so if rally or tempest testrun create resources exclusively
		// belonging to a specific tenant then that tenant can be filtered out.
		// TODO: TBD with SWITCH!

		// Here comes the transformation of the information retrieved into either
		metadata := make(datamodels.JSONdb)
		metadata["imageid"] = imageid
		metadata["imagename"] = imagename
		metadata["imageosflavor"] = imageosflavor
		metadata["flavorid"] = flavorid
		metadata["flavorname"] = flavorname
		metadata["region"] = cfg.OpenStack.Region

		// TODO: MAke more generic and customizable via config file
		if value, exists := s.Metadata["schedule_frequency"]; exists && value == "never" {

			metadata["PlanOverride"] = true

		}

		evTime := int64(time.Now().Unix())
		evLast := getStatus(s.Status)

		if evLast == "terminated" {

			terminatedCount++

		}

		// events reports to be sent.
		event := eeModels.Event{
			Account:      s.TenantID,
			EventTime:    &evTime,
			LastEvent:    &evLast,
			MetaData:     metadata,
			Region:       cfg.OpenStack.Region,
			ResourceID:   s.ID,
			ResourceName: s.Name,
			ResourceType: "server",
		}

		report(event)

		//if this object exists in remote list then lets remove it
		for i, object := range remotelist {

			if strings.Compare(object.Account, s.TenantID) == 0 &&
				strings.Compare(object.ResourceID, s.ID) == 0 &&
				strings.Compare(object.ResourceName, s.Name) == 0 {

				l.Debug.Printf("[COLLECTION] Event send cleaned from the processing list..\n")

				remotelist = append(remotelist[:i], remotelist[i+1:]...)

				break

			}

		}

	}

	ok = true

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

	case "SHELVED_OFFLOADED":

		status = "terminated"

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

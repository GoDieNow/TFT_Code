package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/prometheus/client_golang/prometheus"
	datamodels "gitlab.com/cyclops-utilities/datamodels"
	eeEvent "github.com/GoDieNow/TFT_Code/services/eventsengine/client/event_management"
	eeModels "github.com/GoDieNow/TFT_Code/services/eventsengine/models"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	collector       = "Blockstorage"
	objects         = "volume"
	remotelist      []*eeModels.MinimalState
	collectionStart int64
)

// collect handles the process of retrieving the information from the system.
func collect() {

	l.Trace.Printf("[COLLECTION] The collection process has been started.\n")

	collectionStart = time.Now().UnixNano()

	metricTime.With(prometheus.Labels{"type": "Collection Start Time"}).Set(float64(collectionStart))

	// Here comes the logic to retrieve the information from the system
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

		l.Error.Printf("[COLLECTION] Error authenticating against OpenStack. Error: %v\n", e)

		os.Exit(1)

	}

	listOpts := projects.ListOpts{
		Enabled: gophercloud.Enabled,
	}

	authClient, e := openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})

	if e != nil {

		l.Error.Printf("[COLLECTION] Error creating the collector client. Error: %v\n", e)

		os.Exit(1)

	}

	allPages, e := projects.List(authClient, listOpts).AllPages()

	if e != nil {

		l.Error.Printf("[COLLECTION] Error listing all the pages. Error: %v\n", e)

		os.Exit(1)

	}

	allProjects, e := projects.ExtractProjects(allPages)

	if e != nil {

		l.Error.Printf("[COLLECTION] Error listing all the projects. Error: %v\n", e)

		os.Exit(1)

	}

	l.Trace.Printf("[COLLECTION] Querying events engine service for list of known and not terminated volumes")

	resourceType := "blockstorage"
	eeParams := eeEvent.NewListStatesParams().WithResource(&resourceType).WithRegion(&cfg.OpenStack.Region)

	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
	r, e := reportClient.EventManagement.ListStates(ctx, eeParams)

	// Clears the remotelist between runs
	remotelist = nil

	if e != nil {

		l.Warning.Printf("[COLLECTION] Something went wrong while retrieving the usage from EventsEngine, check with the administrator. Error: %v.\n", e)

	} else {

		remotelist = r.Payload

	}

	resourceType = "blockstorage_ssd"
	eeParams = eeEvent.NewListStatesParams().WithResource(&resourceType).WithRegion(&cfg.OpenStack.Region)

	ctx, _ = context.WithTimeout(context.Background(), 300*time.Second)
	r, e = reportClient.EventManagement.ListStates(ctx, eeParams)

	if e != nil {

		l.Warning.Printf("[COLLECTION] Something went wrong while retrieving the usage(ssd) from EventsEngine, check with the administrator. Error: %v.\n", e)

	} else {

		remotelist = append(remotelist, r.Payload...)

	}

	eeCount := len(remotelist)
	apiCount := 0
	ssdCount := 0

	l.Trace.Printf("[COLLECTION] (BEFORE) Existing count of blockstorage elements at remote [ %v ].\n", len(remotelist))

allProjectsLoop:
	for _, project := range allProjects {

		l.Trace.Printf("[COLLECTION] Found project [ %v ] with ID [ %v ]. Proceeding to get list of volumes.\n", project.Name, project.ID)

		volumeClient, e := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{
			Region: cfg.OpenStack.Region,
		})

		if e != nil {

			l.Error.Printf("[COLLECTION] Error creating block storage client. Error: %v\n", e)

			continue

		}

		// Filter by project id:
		for _, filter := range cfg.ProjectFilters {

			if strings.Contains(project.ID, filter) && filter != "" {

				l.Debug.Printf("[COLLECTION] The Project [ %v ] matches filter [ %v ] and won't be further processed.", project.ID, filter)

				continue allProjectsLoop

			}

		}

		// Filter by project name:
		for _, filter := range cfg.NameFilters {

			if strings.Contains(strings.ToLower(project.Name), strings.ToLower(filter)) && filter != "" {

				l.Debug.Printf("[COLLECTION] The Project [ %v ] matches filter [ %v ] and won't be further processed.", project.Name, filter)

				continue allProjectsLoop

			}

		}

		opts := volumes.ListOpts{
			AllTenants: true,
			TenantID:   project.ID,
		}

		pager := volumes.List(volumeClient, opts)

		e = pager.EachPage(func(page pagination.Page) (bool, error) {

			vList, e := volumes.ExtractVolumes(page)

			if e != nil {

				l.Error.Printf("[COLLECTION] Error processing the lists of active resources. Error: %v\n", e)

				return false, e

			}

			apiCount += len(vList)

			for _, v := range vList {

				// "v" will be a volumes.Volume
				l.Trace.Printf("[COLLECTION] Found volume [ %v ] with ID [ %v ]. Parameters: Status [ %v ], Type [ %v ], Size [ %v ] Zone [ %v ] and Replication [ %v ].\n",
					strings.TrimSpace(v.Name), v.ID, v.Status, v.VolumeType, v.Size, v.AvailabilityZone, v.ReplicationStatus)

				// Here comes the transformation of the information retrieved into either
				md := make(datamodels.JSONdb)
				md["size"] = strconv.Itoa(v.Size)
				md["availabilityzone"] = v.AvailabilityZone
				md["replication"] = v.ReplicationStatus
				md["volumetype"] = v.VolumeType
				md["region"] = cfg.OpenStack.Region

				evTime := int64(time.Now().Unix())
				evLast := getStatus(v.Status)

				objectType := "blockstorage"

				if strings.Contains(v.VolumeType, "ssd") {

					objectType = "blockstorage_ssd"
					ssdCount++

				}

				// events or usage reports to be sent.
				event := eeModels.Event{
					Account:      project.ID,
					EventTime:    &evTime,
					LastEvent:    &evLast,
					MetaData:     md,
					Region:       cfg.OpenStack.Region,
					ResourceID:   v.ID,
					ResourceName: strings.TrimSpace(v.Name),
					ResourceType: objectType,
				}

				report(event)

				//if this object exists in remote list then lets remove it
				for i, object := range remotelist {

					metadata := object.MetaData

					if strings.Compare(object.Account, project.ID) == 0 &&
						strings.Compare(object.ResourceID, v.ID) == 0 &&
						strings.Compare(object.ResourceName, v.Name) == 0 &&
						strings.Compare(metadata["size"].(string), strconv.Itoa(v.Size)) == 0 {

						l.Debug.Printf("[COLLECTION] Event send cleaned from the processing list..\n")

						remotelist = append(remotelist[:i], remotelist[i+1:]...)

						break

					}

				}

			}

			return true, nil

		})

	}

	l.Trace.Printf("[COLLECTION] (AFTER) Remaining count of blockstorage elements at remote which were left unprocessed [ %v ].\n", len(remotelist))

	for _, object := range remotelist {

		l.Debug.Printf("[COLLECTION] Sending termination for zombie data in the system..\n")

		evTime := int64(time.Now().Unix())
		evLast := getStatus("terminated")

		objectType := "blockstorage"

		if strings.Contains(object.MetaData["volumetype"].(string), "ssd") {

			objectType = "blockstorage_ssd"

		}

		// events or usage reports to be sent.
		event := eeModels.Event{
			Account:      object.Account,
			EventTime:    &evTime,
			LastEvent:    &evLast,
			MetaData:     object.MetaData,
			Region:       cfg.OpenStack.Region,
			ResourceID:   object.ResourceID,
			ResourceName: object.ResourceName,
			ResourceType: objectType,
		}

		report(event)

	}

	metricCount.With(prometheus.Labels{"type": "Total Storage Blocks reported by OS API"}).Set(float64(apiCount))

	metricCount.With(prometheus.Labels{"type": "Total Storage Blocks (only SSD) reported by OS API"}).Set(float64(ssdCount))

	metricCount.With(prometheus.Labels{"type": "Total Storage Blocks from EventEngine"}).Set(float64(eeCount))

	metricCount.With(prometheus.Labels{"type": "Total Storage Blocks forcefully TERMINATED"}).Set(float64(len(remotelist)))

	metricTime.With(prometheus.Labels{"type": "Collection Processing Time"}).Set(float64(time.Now().UnixNano()-collectionStart) / float64(time.Millisecond))

	l.Warning.Printf("[COLLECTION] Completed.\n - OS Report: %v\n - EE Report: %v\n - Forced Termination: %v\n - Processing Time: %v[ms]\n", apiCount, eeCount, len(remotelist), float64(time.Now().UnixNano()-collectionStart)/float64(time.Millisecond))

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

		status = "terminated"

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

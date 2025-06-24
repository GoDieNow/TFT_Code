package main

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	rgw "github.com/myENA/radosgwadmin"
	rcl "github.com/myENA/restclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/remeh/sizedwaitgroup"
	datamodels "gitlab.com/cyclops-utilities/datamodels"
	udrModels "github.com/GoDieNow/TFT_Code/services/udr/models"
	l "gitlab.com/cyclops-utilities/logging"
)

var (
	collector       = "Objects"
	objects         = "object"
	collectionStart int64
)

// collect handles the process of retrieving the information from the system.
func collect() {

	l.Trace.Printf("[COLLECTION] The collection process has been started.\n")

	collectionStart = time.Now().UnixNano()

	metricTime.With(prometheus.Labels{"type": "Collection Start Time"}).Set(float64(collectionStart))

	// Here comes the logic to retrieve the information from the system.
	rcfg := &rgw.Config{
		AdminPath:   cfg.RGW.AdminPath,
		AccessKeyID: cfg.RGW.AccessKeyID,
		ClientConfig: rcl.ClientConfig{
			ClientTimeout: rcl.Duration(time.Second * 30),
		},
		SecretAccessKey: cfg.RGW.SecretAccessKey,
		ServerURL:       cfg.RGW.ServerURL,
	}

	a, e := rgw.NewAdminAPI(rcfg)

	if e != nil {

		l.Warning.Printf("[COLLECTION] Could not authenticate with RadosGW. Not obtaining data for this period. Error: %v\n", e.Error())

		return

	}

	users, e := a.MListUsers(context.Background())

	if e != nil {

		l.Warning.Printf("[COLLECTION] Error reading project lists from RGW. Not obtaining data for this period. Error: %v\n", e.Error())

		return

	}

	apiCount := 0
	dropCount := 0

	swg := sizedwaitgroup.New(8)

	// fmt.Printf("Users = %v\n", users)
	for _, user := range users {

		// Goroutines start
		swg.Add()
		go func(u string) {

			defer swg.Done()

			// filter out users which are not Openstack projects
			if filterOpenstackProject(u) {

				// get all buckets for given user
				bucketList, e := a.BucketList(context.Background(), u)

				if e != nil {

					l.Warning.Printf("[COLLECTION] Error reading bucket list for user [ %v ] from RGW. Error: %v\n", u, e.Error())

				}

				// l.Info.Printf("bucketlist = %v\n", bucketList)

				totalBuckets := len(bucketList)
				aggregateBucketSize := float64(0.0)

				apiCount += len(bucketList)

				for _, b := range bucketList {

					bucketStats, e := a.BucketStats(context.Background(), u, b)

					if e != nil {

						l.Warning.Printf("[COLLECTION] Error reading bucket list for user [ %v ] from RGW. Error: %v\n", u, e.Error())

					}

					if len(bucketStats) == 0 {

						l.Warning.Printf("[COLLECTION] The length of bucket stats = 0 for the bucket [ %v ]\n", b)

					} else {

						d := datamodels.JSONdb{
							"region": cfg.RGW.Region,
							"bucket": b,
						}

						// the Usage struct contains support for a number of different RGWs...
						// RGWMain, RGWShadow and a couple of others; here we just take
						// RGWMain as the base - if it's null, we don't publish anything to UDR
						// RGWMain contains SizeKbActual and SizeKb - SizeKBActual is used here;
						// this is the usage as perceived by the user, SizeKb is a bit larger with
						// the delta dependent on the basic minimum object size allocation
						if bucketStats[0].Usage.RGWMain != nil {

							usage := float64(bucketStats[0].Usage.RGWMain.SizeKb)

							if usage != 0 {

								// Here comes the transformation of the information retrieved into either
								// events or usage reports to be sent.
								usageReport := udrModels.Usage{
									Account:      u,
									Metadata:     d,
									ResourceType: "objectstorage",
									Time:         time.Now().Unix(),
									Unit:         "GB",
									Usage:        float64((usage / float64(1024)) / float64(1024)),
								}

								report(usageReport)

								aggregateBucketSize += usageReport.Usage

							}

						} else {

							l.Warning.Printf("[COLLECTION] There's no information for RGWMain for the bucket [ %v ]. Ignoring...\n", b)

							dropCount++

						}
					}
				}

				l.Info.Printf("[COLLECTION] Wrote data for user [ %v ]. AggregateBucketSize [ %v ] distributed over [ %v ] buckets.\n", u, aggregateBucketSize, totalBuckets)

			}

		}(user)

	}

	swg.Wait()

	metricCount.With(prometheus.Labels{"type": "Total Objects reported by OS API"}).Set(float64(apiCount))

	metricCount.With(prometheus.Labels{"type": "Total Objects DROPPED due to missing information"}).Set(float64(dropCount))

	metricTime.With(prometheus.Labels{"type": "Collection Processing Time"}).Set(float64(time.Now().UnixNano()-collectionStart) / float64(time.Millisecond))

	l.Warning.Printf("[COLLECTION] Completed.\n OS Report: %v\n, Dropped: %v\n, Processing Time: %v[ms]\n", apiCount, dropCount, float64(time.Now().UnixNano()-collectionStart)/float64(time.Millisecond))

	l.Trace.Printf("[COLLECTION] The collection process has been finished.\n")

	return

}

// filterOpenstackProject checks if the user name looks like an Openstack
// project.
// Parameters:
// - p: string representing the user.
// Returns:
// - a bool with the result of the check.
func filterOpenstackProject(p string) bool {

	_, e := hex.DecodeString(p)

	return (len(p) == 32) && (e == nil)
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

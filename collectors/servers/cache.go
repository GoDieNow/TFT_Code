package main

import (
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	l "gitlab.com/cyclops-utilities/logging"
)

type FlavorIDCacheData struct {
	FlavorName    string
	LastRefreshed int64
}

type ImageIDCacheData struct {
	ImageName     string
	OSFlavor      string
	LastRefreshed int64
}

var (
	FlavorCache map[string]FlavorIDCacheData
	ImageCache  map[string]ImageIDCacheData
)

func getFromFlavorCache(client *gophercloud.ServiceClient, flavorid string) (flavorname string, e error) {

	var flavor *flavors.Flavor

	if FlavorCache[flavorid].LastRefreshed != 0 {

		//the cache entry exists
		l.Trace.Printf("[CACHE] Flavor cache entry found for key: %v.\n", flavorid)

		if time.Now().Unix()-FlavorCache[flavorid].LastRefreshed > 86400 {

			//invalidate the entry and get a fresh one
			l.Trace.Printf("[CACHE] Cache invalidation for key [ %v ], refreshing entry now.\n", flavorid)

			flavor, e = flavors.Get(client, flavorid).Extract()

			if e != nil {

				l.Warning.Printf("[CACHE] Error reading remote API response for flavor-id data. Error: %v\n", e)

				return

			}

			flavorname = flavor.Name

			FlavorCache[flavorid] = FlavorIDCacheData{
				flavor.Name,
				time.Now().Unix(),
			}

			l.Trace.Printf("[CACHE] Cache entry updated for key [ %v ]. New value returned.\n", flavorid)

			return

		}

		l.Trace.Printf("[CACHE] Cache entry value for key [ %v ] returned without refreshing.\n", flavorid)

		flavorname = FlavorCache[flavorid].FlavorName

		return

	}

	l.Trace.Printf("[CACHE] Cache entry miss for key: %v\n.", flavorid)

	flavor, e = flavors.Get(client, flavorid).Extract()

	if e != nil {

		l.Warning.Printf("[CACHE] Error reading remote API response for flavor-id data. Error: %v ", e)

		return

	}

	flavorname = flavor.Name

	l.Trace.Printf("[CACHE] Going to add a new cache entry added for key: %v\n", flavorid)

	FlavorCache[flavorid] = FlavorIDCacheData{
		flavor.Name,
		time.Now().Unix(),
	}

	l.Trace.Printf("[CACHE] Cache entry added for key [ %v ]. New value returned.\n", flavorid)

	return

}

func getFromImageCache(client *gophercloud.ServiceClient, imageid string) (imagename string, osflavor string, e error) {

	var image *images.Image

	if ImageCache[imageid].LastRefreshed != 0 {

		//the cache entry exists
		l.Trace.Printf("[CACHE] Image cache entry found for image key: %v.\n", imageid)

		if time.Now().Unix()-ImageCache[imageid].LastRefreshed > 86400 {

			//invalidate the entry and get a fresh one
			l.Trace.Printf("[CACHE] Cache invalidation for image key [ %v ], refreshing entry now.\n", imageid)

			image, e = images.Get(client, imageid).Extract()

			if e != nil {

				l.Warning.Printf("[CACHE] Error reading remote API response for image data. Error: %v.\n", e)

				return

			}

			imagename = image.Name

			osflavor_test, exists := image.Metadata["os_flavor"] //image operating system flavor

			if exists {

				osflavor = osflavor_test.(string)

			}

			ImageCache[imageid] = ImageIDCacheData{
				image.Name,
				osflavor,
				time.Now().Unix(),
			}

			l.Trace.Printf("[CACHE] Cache entry updated for image key [ %v ]. New value returned.\n", imageid)

			return

		}

		l.Trace.Printf("[CACHE] Cache entry value for image key [ %v ] returned without refreshing.\n", imageid)

		imagename = ImageCache[imageid].ImageName
		osflavor = ImageCache[imageid].OSFlavor

		return

	}

	l.Trace.Printf("[CACHE] Cache entry miss for image key: %v.\n", imageid)

	image, e = images.Get(client, imageid).Extract()

	if e != nil {

		l.Warning.Printf("[CACHE] Error reading remote API response for image data. Error: %v.\n", e)

		return

	}

	imagename = image.Name

	osflavor_test, exists := image.Metadata["os_flavor"] //image operating system flavor

	if exists {

		osflavor = osflavor_test.(string)

	}

	l.Trace.Printf("[CACHE] Going to add a new cache entry added for image key: " + imageid)

	ImageCache[imageid] = ImageIDCacheData{
		image.Name,
		osflavor,
		time.Now().Unix(),
	}

	l.Trace.Printf("[CACHE] Cache entry added for image key: " + imageid + ". New value returned.")

	return

}

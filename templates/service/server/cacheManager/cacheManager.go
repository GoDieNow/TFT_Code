package cacheManager

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	l "gitlab.com/cyclops-utilities/logging"
)

// cacheFetchFunction it will need the id to retrieve and as and option a Bearer
// Keycloak token can be provided when called.
// It shall return the object and errors in case of problems.
type cacheFetchFunction func(interface{}, string) (interface{}, error)

// cache struct is the main "object" which handles the cache and its items.
// It also contains a map with the conversions of interfaces to strings.
type cache struct {
	////conversionDictionary map[string]string
	data     map[string]*cacheItem
	fetchers map[string]cacheFetchFunction
	mutex    sync.RWMutex
}

// cacheItem struct referes to the data of each element saved in the cache.
type cacheItem struct {
	fetcher    cacheFetchFunction
	lastUpdate time.Time
	value      interface{}
}

// CacheManager is the struct defined to group and contain all the methods
// that interact with the caching mechanism.
type CacheManager struct {
	APIKey   string
	duration time.Duration
	metrics  *prometheus.GaugeVec
	store    cache
}

// New is the function to create the struct CacheManager.
// Parameters:
// - t: a time.Duration with the max duration alive of the cache elements.
// - k: string containing the APIKey/token in case of need.
// Returns:
// - CacheManager: struct to interact with CacheManager subsystem functionalities.
func New(metrics *prometheus.GaugeVec, t time.Duration, k string) *CacheManager {

	l.Trace.Printf("[CACHE] Initializing the cache service.\n")

	c := CacheManager{
		APIKey:   k,
		duration: t,
		metrics:  metrics,
		store: cache{
			data:     make(map[string]*cacheItem),
			fetchers: make(map[string]cacheFetchFunction),
		},
	}

	return &c

}

// Add function job is to insert a new model in the cache.
// What it does is link the model with a fetching function and, if wanted, with
// a plain text name, so later in order to retrieve things from the cache they
// can be refereced either by the struct model or the plain text name.
// Paramenters:
// - plainName: a case insensitive name/alias to retrieve the data.
// - fetcher: the cacheFetchFunction used to retrieve the data.
func (c *CacheManager) Add(plainName string, fetcher cacheFetchFunction) {

	l.Trace.Printf("[CACHE] Adding a new object fetcher in the cache.\n")

	key := strings.ToUpper(plainName)

	c.store.mutex.Lock()

	c.store.fetchers[key] = fetcher

	c.store.mutex.Unlock()

	c.metrics.With(prometheus.Labels{"state": "OK", "resource": "Models in Cache"}).Inc()

	return

}

// fetch function job is to retrieve a new and updated copy of the remote object.
// Paramenters:
// - item: a string used as key-value in the cache storage to identify the item
// that is going to be updated.
// - token: Keycloak Bearer token, completely optional.
// Returns:
// - e: error in case of something went wrong while setting up the new association.
func (c *CacheManager) fetch(item string, token string) (e error) {

	l.Trace.Printf("[CACHE] Fetching the item [ %v ] from the remote location.\n", item)

	c.store.mutex.RLock()

	object := c.store.data[item]

	c.store.mutex.RUnlock()

	id := strings.SplitN(item, "-", 2)[1]

	uValue, e := object.fetcher(id, token)

	if e == nil {

		l.Trace.Printf("[CACHE] Item [ %v ] retrieved from the remote location and saved in the cache.\n", item)

		object.value = uValue
		object.lastUpdate = time.Now()

	} else {

		l.Warning.Printf("[CACHE] Something went wrong while retrieving the item. Error: %v\n", e)

	}

	return

}

// fetch function job is to create the new item in the cache and retrieve a new
// and updated initial copy of the remote object to be saved in the cache.
// Paramenters:
// - item: a string used as key-value in the cache storage to identify the item
// that is going to be updated.
// - token: Keycloak Bearer token, completely optional.
// Returns:
// - e: error in case of something went wrong while setting up the new association.
func (c *CacheManager) init(item string, token string) (e error) {

	l.Trace.Printf("[CACHE] Fetching the item [ %v ] from the remote location.\n", item)

	key := strings.Split(item, "-")[0]
	id := strings.SplitN(item, "-", 2)[1]

	uValue, e := c.store.fetchers[key](id, token)

	if e == nil {

		l.Trace.Printf("[CACHE] Item [ %v ] retrieved from the remote location and saved in the cache.\n", item)

		i := cacheItem{
			fetcher:    c.store.fetchers[key],
			lastUpdate: time.Now(),
			value:      uValue,
		}

		c.store.mutex.Lock()

		c.store.data[item] = &i

		c.store.mutex.Unlock()

	} else {

		l.Warning.Printf("[CACHE] Something went wrong while retrieving the item. Error: %v\n", e)

	}

	return

}

// key is a function to ensure that the creation of the the item key for the
// cache is consistent across all the functions.
// Paramenters:
// - id: the reference id of the object to be retrieved
// - model: the alias text used to identify the source of objects.
// Returns:
// - s: the key string
func (c *CacheManager) key(id interface{}, model string) (s string) {

	s = fmt.Sprintf("%v-%v", strings.ToUpper(model), id)

	return

}

// Get function job is to retrieve an object from the cache or fetch it from the
// source and upgrade the copy in the cache in case the expiration time has been
// exceeded.
// Paramenters:
// - id: the reference id of the object.
// - model: the text alias set to reference the model.
// - token: Keycloak Bearer token, completely optional.
// Returns:
// - The object associated with the request
// - An error raised in case something went wrong while retrieving the object.
func (c *CacheManager) Get(id interface{}, model string, token string) (interface{}, error) {

	l.Trace.Printf("[CACHE] Retrieving object [ %v, %v ] from the cache.\n", id, model)

	item := c.key(id, model)

	c.store.mutex.RLock()

	object, exists := c.store.data[item]

	c.store.mutex.RUnlock()

	if !exists {

		l.Trace.Printf("[CACHE] Object [ %v ] first time requested, including in the cache.\n", item)

		if e := c.init(item, token); e != nil {

			l.Warning.Printf("[CACHE] Something went wrong while adding the new item [ %v ] to the cache. Error: %v\n", item, e)

			c.metrics.With(prometheus.Labels{"state": "FAIL", "resource": "Total objects cached"}).Inc()

			return nil, e

		}

		l.Trace.Printf("[CACHE] Object [ %v ] retrieved from the cache.\n", item)

		c.store.mutex.RLock()

		o := c.store.data[item].value

		c.store.mutex.RUnlock()

		c.metrics.With(prometheus.Labels{"state": "OK", "resource": "Total objects cached"}).Inc()

		return o, nil

	}

	l.Trace.Printf("[CACHE] Object [ %v ] exists in the cache.\n", item)

	c.store.mutex.RLock()

	diff := (time.Now()).Sub(c.store.data[item].lastUpdate)

	c.store.mutex.RUnlock()

	if diff <= c.duration {

		l.Trace.Printf("[CACHE] Object [ %v ] cache hasn't expired yet.\n", item)

		c.metrics.With(prometheus.Labels{"state": "OK", "resource": "Total objects retrieved from cache"}).Inc()

		return object.value, nil

	}

	l.Warning.Printf("[CACHE] Object [ %v ] cache has expired. Starting the upgrade.\n", item)

	if e := c.fetch(item, token); e != nil {

		l.Warning.Printf("[CACHE] Something went wrong while fetching the updated data for the object [ %v ] to the cache. Error: %v\n", item, e)

		c.metrics.With(prometheus.Labels{"state": "FAIL", "resource": "Total objects refreshed"}).Inc()

		return nil, e

	}

	l.Trace.Printf("[CACHE] Object [ %v ] updated retrieved from the cache.\n", item)

	c.store.mutex.RLock()

	o := c.store.data[item].value

	c.store.mutex.RUnlock()

	c.metrics.With(prometheus.Labels{"state": "OK", "resource": "Total objects refreshed"}).Inc()
	c.metrics.With(prometheus.Labels{"state": "OK", "resource": "Total objects retrieved from cache"}).Inc()

	return o, nil

}

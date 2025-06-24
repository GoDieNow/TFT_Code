package main

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/GoDieNow/TFT_Code/services/billing/server/cacheManager"
	cdrClient "github.com/GoDieNow/TFT_Code/services/cdr/client"
	cdrUsage "github.com/GoDieNow/TFT_Code/services/cdr/client/usage_management"
	cusClient "github.com/GoDieNow/TFT_Code/services/customerdb/client"
	cusCustomer "github.com/GoDieNow/TFT_Code/services/customerdb/client/customer_management"
	cusProduct "github.com/GoDieNow/TFT_Code/services/customerdb/client/product_management"
	cusReseller "github.com/GoDieNow/TFT_Code/services/customerdb/client/reseller_management"
	pmClient "github.com/GoDieNow/TFT_Code/services/planmanager/client"
	pmBundle "github.com/GoDieNow/TFT_Code/services/planmanager/client/bundle_management"
	pmCycle "github.com/GoDieNow/TFT_Code/services/planmanager/client/cycle_management"
	pmPlan "github.com/GoDieNow/TFT_Code/services/planmanager/client/plan_management"
	pmSku "github.com/GoDieNow/TFT_Code/services/planmanager/client/sku_management"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
	l "gitlab.com/cyclops-utilities/logging"
)

// cacheStart handles the initialization of the cache mechanism service.
// Returns:
// - A cacheManager reference struct already initialized and ready to be used.
func cacheStart(metrics *prometheus.GaugeVec) *cacheManager.CacheManager {

	l.Trace.Printf("[CACHE][INIT] Intializing cache mechanism.\n")

	cacheDuration, _ := time.ParseDuration(cfg.DB.CacheRetention)

	c := cacheManager.New(metrics, cacheDuration, cfg.APIKey.Token)

	resellerFunction := func(id interface{}, token string) (interface{}, error) {

		config := cusClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["customerdb"],
				Path:   cusClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := cusClient.New(config)
		ctx := context.Background()

		i := id.(string)

		if i != "ALL" {

			params := cusReseller.NewGetResellerParams().WithID(i)

			r, e := client.ResellerManagement.GetReseller(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the reseller with id [ %v ]. Error: %v", i, e)

				return nil, e

			}

			return *r.Payload, nil

		}

		params := cusReseller.NewListResellersParams()

		r, e := client.ResellerManagement.ListResellers(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the list of resellers. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	customerFunction := func(id interface{}, token string) (interface{}, error) {

		config := cusClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["customerdb"],
				Path:   cusClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := cusClient.New(config)
		ctx := context.Background()

		i := id.(string)

		if i != "ALL" {

			params := cusCustomer.NewGetCustomerParams().WithID(i)

			r, e := client.CustomerManagement.GetCustomer(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the customer with id [ %v ]. Error: %v", i, e)

				return nil, e

			}

			return *r.Payload, nil

		}

		params := cusCustomer.NewListCustomersParams()

		r, e := client.CustomerManagement.ListCustomers(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the list of customers. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	productFunction := func(id interface{}, token string) (interface{}, error) {

		config := cusClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["customerdb"],
				Path:   cusClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := cusClient.New(config)
		ctx := context.Background()

		i := id.(string)

		if i != "ALL" {

			params := cusProduct.NewGetProductParams().WithID(i)

			r, e := client.ProductManagement.GetProduct(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the product with id [ %v ]. Error: %v", i, e)

				return nil, e

			}

			return *r.Payload, nil

		}

		params := cusProduct.NewListProductsParams()

		r, e := client.ProductManagement.ListProducts(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][CUSDB-FUNCTION] There was a problem while retrieving the list of products. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	// id == id0[,id1,...,idN]?from?to
	cdrFunction := func(id interface{}, token string) (interface{}, error) {

		config := cdrClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["cdr"],
				Path:   cdrClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		idSplit := strings.SplitN(id.(string), "?", 3)

		i := idSplit[0]

		from, e := time.Parse(time.RFC3339Nano, idSplit[1])
		if e != nil {

			l.Warning.Printf("[CACHE][CDR-FUNCTION] There was a problem while parsing the datetime [ %v ]. Error: %v", idSplit[1], e)

			return nil, e

		}

		f := (strfmt.DateTime)(from)

		to, e := time.Parse(time.RFC3339Nano, idSplit[2])
		if e != nil {

			l.Warning.Printf("[CACHE][CDR-FUNCTION] There was a problem while parsing the datetime [ %v ]. Error: %v", idSplit[2], e)

			return nil, e

		}

		t := (strfmt.DateTime)(to)

		client := cdrClient.New(config)
		ctx := context.Background()

		if strings.Contains(i, ",") {

			params := cdrUsage.NewGetSystemUsageParams().WithIdlist(&i).WithFrom(&f).WithTo(&t)

			r, e := client.UsageManagement.GetSystemUsage(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][CDR-FUNCTION] There was a problem while retrieving all the CDRs from the system. Error: %v", e)

				return nil, e

			}

			return r.Payload, nil

		}

		if i != "ALL" {

			params := cdrUsage.NewGetUsageParams().WithID(i).WithFrom(&f).WithTo(&t)

			r, e := client.UsageManagement.GetUsage(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][CDR-FUNCTION] There was a problem while retrieving the CDRs under the id [ %v ]. Error: %v", id, e)

				return nil, e

			}

			return r.Payload, nil

		}

		params := cdrUsage.NewGetSystemUsageParams().WithFrom(&f).WithTo(&t)

		r, e := client.UsageManagement.GetSystemUsage(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][CDR-FUNCTION] There was a problem while retrieving all the CDRs from the system. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	skuFunction := func(id interface{}, token string) (interface{}, error) {

		config := pmClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["planmanager"],
				Path:   pmClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := pmClient.New(config)
		ctx := context.Background()

		if id.(string) != "ALL" {

			params := pmSku.NewGetSkuParams().WithID(id.(string))

			r, e := client.SkuManagement.GetSku(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][SKU-FUNCTION] There was a problem while retrieving the sku [ %v ]. Error: %v", id, e)

				return nil, e

			}

			return r.Payload, nil

		}

		params := pmSku.NewListSkusParams()

		r, e := client.SkuManagement.ListSkus(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][SKU-FUNCTION] There was a problem while retrieving the skus list. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	planFunction := func(id interface{}, token string) (interface{}, error) {

		config := pmClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["planmanager"],
				Path:   pmClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := pmClient.New(config)
		ctx := context.Background()

		if id.(string) != "ALL" {

			var planID string

			if i, exists := cfg.DefaultPlans[strings.ToLower(id.(string))]; exists {

				planID = i

			} else {

				planID = id.(string)

			}

			params := pmPlan.NewGetCompletePlanParams().WithID(planID)

			r, e := client.PlanManagement.GetCompletePlan(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][PLAN-FUNCTION] There was a problem while retrieving the plan [ %v ]. Error: %v", id, e)

				return nil, e

			}

			return *r.Payload, nil

		}

		params := pmPlan.NewListCompletePlansParams()

		r, e := client.PlanManagement.ListCompletePlans(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][PLAN-FUNCTION] There was a problem while retrieving the plan list. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	bundleFunction := func(id interface{}, token string) (interface{}, error) {

		config := pmClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["planmanager"],
				Path:   pmClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := pmClient.New(config)
		ctx := context.Background()

		if id.(string) != "ALL" {

			params := pmBundle.NewGetSkuBundleParams().WithID(id.(string))

			r, e := client.BundleManagement.GetSkuBundle(ctx, params)

			if e != nil {

				l.Warning.Printf("[CACHE][BUNDLE-FUNCTION] There was a problem while retrieving the skubundle [ %v ]. Error: %v", id, e)

				return nil, e

			}

			return *r.Payload, nil

		}

		params := pmBundle.NewListSkuBundlesParams()

		r, e := client.BundleManagement.ListSkuBundles(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][BUNDLE-FUNCTION] There was a problem while retrieving the skubundle list. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	cycleFunction := func(id interface{}, token string) (interface{}, error) {

		config := pmClient.Config{
			URL: &url.URL{
				Host:   cfg.General.Services["planmanager"],
				Path:   pmClient.DefaultBasePath,
				Scheme: "http",
			},
			AuthInfo: httptransport.APIKeyAuth(cfg.APIKey.Key, cfg.APIKey.Place, cfg.APIKey.Token),
		}

		if token != "" {

			config.AuthInfo = httptransport.BearerToken(token)

		}

		client := pmClient.New(config)
		ctx := context.Background()

		var params *pmCycle.ListCyclesParams

		if id.(string) == "ALL" {

			params = pmCycle.NewListCyclesParams()

		} else {

			ty := id.(string)

			params = pmCycle.NewListCyclesParams().WithType(&ty)

		}

		r, e := client.CycleManagement.ListCycles(ctx, params)

		if e != nil {

			l.Warning.Printf("[CACHE][CYCLE-FUNCTION] There was a problem while retrieving the cycle list. Error: %v", e)

			return nil, e

		}

		return r.Payload, nil

	}

	c.Add("reseller", resellerFunction)
	l.Trace.Printf("[CACHE][INIT] Reseller fetcher added to the cache.\n")

	c.Add("customer", customerFunction)
	l.Trace.Printf("[CACHE][INIT] Customer fetcher added to the cache.\n")

	c.Add("product", productFunction)
	l.Trace.Printf("[CACHE][INIT] Product fetcher added to the cache.\n")

	c.Add("cdr", cdrFunction)
	l.Trace.Printf("[CACHE][INIT] CDR usage fetcher added to the cache.\n")

	c.Add("sku", skuFunction)
	l.Trace.Printf("[CACHE][INIT] SKU fetcher added to the cache.\n")

	c.Add("plan", planFunction)
	l.Trace.Printf("[CACHE][INIT] Plan fetcher added to the cache.\n")

	c.Add("bundle", bundleFunction)
	l.Trace.Printf("[CACHE][INIT] SkuBundle fetcher added to the cache.\n")

	c.Add("cycle", cycleFunction)
	l.Trace.Printf("[CACHE][INIT] Life Cycle fetcher added to the cache.\n")

	return c

}

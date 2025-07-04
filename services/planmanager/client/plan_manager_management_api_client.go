// Code generated by go-swagger; DO NOT EDIT.

package client

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"net/url"

	"github.com/go-openapi/runtime"
	rtclient "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/GoDieNow/TFT_Code/services/planmanager/client/bundle_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/cycle_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/plan_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/price_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/sku_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/status_management"
	"github.com/GoDieNow/TFT_Code/services/planmanager/client/trigger_management"
)

const (
	// DefaultHost is the default Host
	// found in Meta (info) section of spec file
	DefaultHost string = "localhost:8000"
	// DefaultBasePath is the default BasePath
	// found in Meta (info) section of spec file
	DefaultBasePath string = "/api/v1.0"
)

// DefaultSchemes are the default schemes found in Meta (info) section of spec file
var DefaultSchemes = []string{"http", "https"}

type Config struct {
	// URL is the base URL of the upstream server
	URL *url.URL
	// Transport is an inner transport for the client
	Transport http.RoundTripper
	// AuthInfo is for authentication
	AuthInfo runtime.ClientAuthInfoWriter
}

// New creates a new plan manager management API HTTP client.
func New(c Config) *PlanManagerManagementAPI {
	var (
		host     = DefaultHost
		basePath = DefaultBasePath
		schemes  = DefaultSchemes
	)

	if c.URL != nil {
		host = c.URL.Host
		basePath = c.URL.Path
		schemes = []string{c.URL.Scheme}
	}

	transport := rtclient.New(host, basePath, schemes)
	if c.Transport != nil {
		transport.Transport = c.Transport
	}

	cli := new(PlanManagerManagementAPI)
	cli.Transport = transport
	cli.BundleManagement = bundle_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.CycleManagement = cycle_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.PlanManagement = plan_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.PriceManagement = price_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.SkuManagement = sku_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.StatusManagement = status_management.New(transport, strfmt.Default, c.AuthInfo)
	cli.TriggerManagement = trigger_management.New(transport, strfmt.Default, c.AuthInfo)
	return cli
}

// PlanManagerManagementAPI is a client for plan manager management API
type PlanManagerManagementAPI struct {
	BundleManagement  *bundle_management.Client
	CycleManagement   *cycle_management.Client
	PlanManagement    *plan_management.Client
	PriceManagement   *price_management.Client
	SkuManagement     *sku_management.Client
	StatusManagement  *status_management.Client
	TriggerManagement *trigger_management.Client
	Transport         runtime.ClientTransport
}

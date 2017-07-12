package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/tylerb/graceful"

	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target ../lib/apiservers/service --name  --spec ../lib/apiservers/service/swagger.json --exclude-main

func configureFlags(api *operations.VicMachineAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.VicMachineAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.TxtProducer = runtime.TextProducer()

	// Applies when the Authorization header is set with the Basic scheme
	api.BasicAuth = handlers.BasicAuth

	// GET /container
	api.GetHandler = operations.GetHandlerFunc(func(params operations.GetParams) middleware.Responder {
		return middleware.NotImplemented("operation .Get has not yet been implemented")
	})

	// GET /container/version
	api.GetVersionHandler = &handlers.VersionGet{}

	// POST /container/{target}
	api.PostTargetHandler = operations.PostTargetHandlerFunc(func(params operations.PostTargetParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTarget has not yet been implemented")
	})

	// GET /container/{target}/vch
	api.GetTargetVchHandler = &handlers.VCHListGet{}

	// POST /container/{target}/vch
	api.PostTargetVchHandler = &handlers.VCHCreate{}

	// GET /container/{target}/vch/{vch-id}
	api.GetTargetVchVchIDHandler = &handlers.VCHGet{}

	// PUT /container/{target}/vch/{vch-id}
	api.PutTargetVchVchIDHandler = operations.PutTargetVchVchIDHandlerFunc(func(params operations.PutTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PutTargetVchVchID has not yet been implemented")
	})

	// PATCH /container/{target}/vch/{vch-id}
	api.PatchTargetVchVchIDHandler = operations.PatchTargetVchVchIDHandlerFunc(func(params operations.PatchTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PatchTargetVchVchID has not yet been implemented")
	})

	// POST /container/{target}/vch/{vch-id}
	api.PostTargetVchVchIDHandler = operations.PostTargetVchVchIDHandlerFunc(func(params operations.PostTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetVchVchID has not yet been implemented")
	})

	// DELETE /container/{target}/vch/{vch-id}
	api.DeleteTargetVchVchIDHandler = operations.DeleteTargetVchVchIDHandlerFunc(func(params operations.DeleteTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .DeleteTargetVchVchID has not yet been implemented")
	})

	// POST /container/{target}/datacenter/{datacenter}
	api.PostTargetDatacenterDatacenterHandler = operations.PostTargetDatacenterDatacenterHandlerFunc(func(params operations.PostTargetDatacenterDatacenterParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetDatacenterDatacenter has not yet been implemented")
	})

	// GET /container/{target}/datacenter/{datacenter}/vch
	api.GetTargetDatacenterDatacenterVchHandler = &handlers.VCHDatacenterListGet{}

	// POST /container/target/datacenter/{datacenter}/vch
	api.PostTargetDatacenterDatacenterVchHandler = operations.PostTargetDatacenterDatacenterVchHandlerFunc(func(params operations.PostTargetDatacenterDatacenterVchParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetDatacenterDatacenterVch has not yet been implemented")
	})

	// GET /container/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.GetTargetDatacenterDatacenterVchVchIDHandler = operations.GetTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.GetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .GetTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// PUT /container/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PutTargetDatacenterDatacenterVchVchIDHandler = operations.PutTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PutTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PutTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// PATCH /container/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PatchTargetDatacenterDatacenterVchVchIDHandler = operations.PatchTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PatchTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PatchTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// POST /container/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PostTargetDatacenterDatacenterVchVchIDHandler = operations.PostTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PostTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// DELETE /container/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.DeleteTargetDatacenterDatacenterVchVchIDHandler = operations.DeleteTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.DeleteTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .DeleteTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}

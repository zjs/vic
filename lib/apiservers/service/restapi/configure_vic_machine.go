// Copyright 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package restapi

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"github.com/rs/cors"
	"github.com/tylerb/graceful"

	"github.com/vmware/vic/lib/apiservers/service/restapi/handlers"
	"github.com/vmware/vic/lib/apiservers/service/restapi/operations"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target ../lib/apiservers/service --name  --spec ../lib/apiservers/service/swagger.json --exclude-main

var logging = struct {
	Directory string `long:"log-directory" description:"the directory to write server logs" default:"/var/log/vic-machine-server/" env:"LOG_DIRECTORY"`
}{}

var logger *log.Logger

func configureFlags(api *operations.VicMachineAPI) {
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			ShortDescription: "Logging Options",
			LongDescription: "",
			Options: &logging,
		},
	}
}

func configureAPI(api *operations.VicMachineAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	output := logging.Directory + "/vic-machine-server.log"
	file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", output, ":", err)
	}

	logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

	api.Logger = logger.Printf


	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	api.TxtProducer = runtime.TextProducer()

	// Applies when the Authorization header is set with the Basic scheme
	api.BasicAuth = handlers.BasicAuth

	api.SessionAuth = handlers.SessionAuth

	// GET /container
	api.GetHandler = operations.GetHandlerFunc(func(params operations.GetParams) middleware.Responder {
		return middleware.NotImplemented("operation .Get has not yet been implemented")
	})

	// GET /container/version
	api.GetVersionHandler = &handlers.VersionGet{}

	// POST /container/target/{target}
	api.PostTargetTargetHandler = operations.PostTargetTargetHandlerFunc(func(params operations.PostTargetTargetParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetTarget has not yet been implemented")
	})

	// GET /container/target/{target}/vch
	api.GetTargetTargetVchHandler = &handlers.VCHListGet{}

	// POST /container/target/{target}/vch
	api.PostTargetTargetVchHandler = &handlers.VCHCreate{}

	// GET /container/target/{target}/vch/{vch-id}
	api.GetTargetTargetVchVchIDHandler = &handlers.VCHGet{}

	// GET /container/target/{target}/vch/{vch-id}/certificate
	api.GetTargetTargetVchVchIDCertificateHandler = &handlers.VCHCertGet{}

	// GET /container/target/{target}/vch/{vch-id}/log
	api.GetTargetTargetVchVchIDLogHandler = &handlers.VCHLogGet{}

	// GET /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}/certificate
	api.GetTargetTargetDatacenterDatacenterVchVchIDCertificateHandler = &handlers.VCHDatacenterCertGet{}

	// GET /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}/log
	api.GetTargetTargetDatacenterDatacenterVchVchIDLogHandler = &handlers.VCHDatacenterLogGet{}

	// PUT /container/target/{target}/vch/{vch-id}
	api.PutTargetTargetVchVchIDHandler = operations.PutTargetTargetVchVchIDHandlerFunc(func(params operations.PutTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PutTargetTargetVchVchID has not yet been implemented")
	})

	// PATCH /container/target/{target}/vch/{vch-id}
	api.PatchTargetTargetVchVchIDHandler = operations.PatchTargetTargetVchVchIDHandlerFunc(func(params operations.PatchTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PatchTargetTargetVchVchID has not yet been implemented")
	})

	// POST /container/target/{target}/vch/{vch-id}
	api.PostTargetTargetVchVchIDHandler = operations.PostTargetTargetVchVchIDHandlerFunc(func(params operations.PostTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetTargetVchVchID has not yet been implemented")
	})

	// DELETE /container/target/{target}/vch/{vch-id}
	api.DeleteTargetTargetVchVchIDHandler = operations.DeleteTargetTargetVchVchIDHandlerFunc(func(params operations.DeleteTargetTargetVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .DeleteTargetTargetVchVchID has not yet been implemented")
	})

	// POST /container/target/{target}/datacenter/{datacenter}
	api.PostTargetTargetDatacenterDatacenterHandler = operations.PostTargetTargetDatacenterDatacenterHandlerFunc(func(params operations.PostTargetTargetDatacenterDatacenterParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetTargetDatacenterDatacenter has not yet been implemented")
	})

	// GET /container/target/{target}/datacenter/{datacenter}/vch
	api.GetTargetTargetDatacenterDatacenterVchHandler = &handlers.VCHDatacenterListGet{}

	// POST /container/target/{target}/datacenter/{datacenter}/vch
	api.PostTargetTargetDatacenterDatacenterVchHandler = &handlers.VCHDatacenterCreate{}

	// GET /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.GetTargetTargetDatacenterDatacenterVchVchIDHandler = &handlers.VCHDatacenterGet{}

	// PUT /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PutTargetTargetDatacenterDatacenterVchVchIDHandler = operations.PutTargetTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PutTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PutTargetTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// PATCH /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PatchTargetTargetDatacenterDatacenterVchVchIDHandler = operations.PatchTargetTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PatchTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PatchTargetTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// POST /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.PostTargetTargetDatacenterDatacenterVchVchIDHandler = operations.PostTargetTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.PostTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .PostTargetTargetDatacenterDatacenterVchVchID has not yet been implemented")
	})

	// DELETE /container/target/{target}/datacenter/{datacenter}/vch/{vch-id}
	api.DeleteTargetTargetDatacenterDatacenterVchVchIDHandler = operations.DeleteTargetTargetDatacenterDatacenterVchVchIDHandlerFunc(func(params operations.DeleteTargetTargetDatacenterDatacenterVchVchIDParams, principal interface{}) middleware.Responder {
		return middleware.NotImplemented("operation .DeleteTargetTargetDatacenterDatacenterVchVchID has not yet been implemented")
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
	// These settings have security implications. These settings should not be changed without appropriate review.
	// For more information, see the relevant section of the design document:
	// https://github.com/vmware/vic/blob/7f575392df99642c5edd8f539a74fe9c89155b00/doc/design/vic-machine/service.md#cross-origin-requests--cross-site-request-forgery
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "User-Agent", "X-VMWARE-TICKET"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: false,
	})

	return addLogging(c.Handler(handler))
}

func addLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mrw := NewMyResponseWriter(w)
		next.ServeHTTP(mrw, r)
		logger.Println("request:", mrw.status, r.Method, r.URL)
	})
}

// https://gist.github.com/ciaranarcher/abccf50cb37645ca27fa
// Maybe https://github.com/felixge/httpsnoop#why-this-package-exists is better?
type MyResponseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func NewMyResponseWriter(w http.ResponseWriter) *MyResponseWriter {
	return &MyResponseWriter{ResponseWriter: w}
}

func (w *MyResponseWriter) Status() int {
	return w.status
}

func (w *MyResponseWriter) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *MyResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.status = code
	w.wroteHeader = true
}

// Filename: cmd/api/dependencies/handlers/routes.go

package handlers

import (
  "net/http"
  "github.com/julienschmidt/httprouter"
  "github.com/spector-asael/banking-api/cmd/api/dependencies/middleware"
  // "expvar"
)

func (a *HandlerDependencies) Routes() http.Handler  {

	middlewareInstance := & middleware.MiddlewareDependencies{
		Config: a.Config,
		Logger: a.Logger,
	}
  // setup a new router
  router := httprouter.New()
  // router.NotFound = http.HandlerFunc(a.notFoundResponse)
  // router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
  // setup routes

  // Persons routes
  router.HandlerFunc(http.MethodGet, "/api/persons", a.getAllPersonsHandler)                   // Get all persons
  router.HandlerFunc(http.MethodPost, "/api/persons", a.createPersonHandler)                   // Create a new person
  router.HandlerFunc(http.MethodGet, "/api/persons/:ssid", a.getPersonBySSIDHandler)          // Get a person by SSID
  router.HandlerFunc(http.MethodPatch, "/api/persons/:ssid", a.updatePersonHandler)           // Update a person by SSID
  router.HandlerFunc(http.MethodDelete, "/api/persons/:ssid", a.deletePersonHandler) 
    
  gzipRequestMiddleware := middlewareInstance.GzipRequestMiddleware(router)
  gzipResponseMiddleware := middlewareInstance.GzipResponseMiddleware(gzipRequestMiddleware)
	loggingMiddleware := middlewareInstance.LoggingMiddleware(gzipResponseMiddleware)
	rateLimitMiddleware := middlewareInstance.RateLimit(loggingMiddleware)
  corsMiddleware := middlewareInstance.EnableCORS(rateLimitMiddleware)
  metricsMiddleware := middlewareInstance.MetricsMiddleware(corsMiddleware)
	panicMiddleware := middlewareInstance.RecoverPanic(metricsMiddleware)


  // Request sent first to recoverPanic() then sent to rateLimit()
  // finally it is sent to the router.
	
  return panicMiddleware
  
}

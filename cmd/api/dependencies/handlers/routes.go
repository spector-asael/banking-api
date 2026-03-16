// Filename: cmd/api/dependencies/handlers/routes.go

package handlers

import (
  "net/http"
  "github.com/julienschmidt/httprouter"
  "github.com/spector-asael/banking-api/cmd/api/dependencies/middleware"
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
   /*
   router.HandlerFunc(http.MethodGet, "/v1/balance", a.checkBalanceHandler)
   router.HandlerFunc(http.MethodPost, "/v1/deposit", a.depositHandler)
   router.HandlerFunc(http.MethodPost, "/v1/history", a.checkHistoryHandler)
   router.HandlerFunc(http.MethodDelete, "/v1/delete", a.deleteDepositHandler)
   router.HandlerFunc(http.MethodPatch, "/v1/update", a.updateDepositHandler)
   router.HandlerFunc(http.MethodPost, "/v1/transfer", TransferHandler)
   router.HandlerFunc(http.MethodGet, "/shutdown", shutdownTestHandler)
   */

   	router.HandlerFunc(http.MethodGet, "/test", a.testHandler)
	
	loggingMiddleware := middlewareInstance.LoggingMiddleware(router)
	rateLimitMiddleware := middlewareInstance.RateLimit(loggingMiddleware)
	panicMiddleware := middlewareInstance.RecoverPanic(rateLimitMiddleware)
		// Request sent first to recoverPanic() then sent to rateLimit()
		// finally it is sent to the router.
	
   	return panicMiddleware
  
}

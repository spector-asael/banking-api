// Filename: cmd/api/dependencies/handlers/routes.go

package handlers

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/spector-asael/banking-api/cmd/api/dependencies/middleware"
)

func (a *HandlerDependencies) Routes() http.Handler {
	if a.Logger != nil {
		a.Logger.Info("Routes() called: building handler chain")
	}

	middlewareInstance := &middleware.MiddlewareDependencies{
		Config:  a.Config,
		Logger:  a.Logger,
		Helpers: &a.Helper,
		Models:  &a.Models,
	}
	// setup a new router
	router := httprouter.New()
	// router.NotFound = http.HandlerFunc(a.notFoundResponse)
	// router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	// setup routes
	// Persons routes
	// Persons routes
	router.HandlerFunc(http.MethodGet, "/api/persons", middlewareInstance.RequirePermission("persons:read", a.getAllPersonsHandler))
	router.HandlerFunc(http.MethodPost, "/api/persons", middlewareInstance.RequirePermission("persons:write", a.createPersonHandler))
	router.HandlerFunc(http.MethodGet, "/api/persons/:ssid", middlewareInstance.RequirePermission("persons:read", a.getPersonBySSIDHandler))
	router.HandlerFunc(http.MethodPatch, "/api/persons/:ssid", middlewareInstance.RequirePermission("persons:write", a.updatePersonHandler))
	router.HandlerFunc(http.MethodDelete, "/api/persons/:ssid", middlewareInstance.RequirePermission("persons:write", a.deletePersonHandler))

	// Customers routes
	router.HandlerFunc(http.MethodGet, "/api/customers", middlewareInstance.RequirePermission("customers:read", a.getAllCustomersHandler))
	router.HandlerFunc(http.MethodPost, "/api/customers", middlewareInstance.RequirePermission("customers:write", a.createCustomerHandler))
	router.HandlerFunc(http.MethodGet, "/api/customers/:id", middlewareInstance.RequirePermission("customers:read", a.getCustomerByIDHandler))
	router.HandlerFunc(http.MethodPatch, "/api/customers/:id/kyc-status", middlewareInstance.RequirePermission("customers:write", a.updateCustomerKYCStatusHandler))
	router.HandlerFunc(http.MethodDelete, "/api/customers/:id", middlewareInstance.RequirePermission("customers:write", a.deleteCustomerHandler))

	// Employees routes
	router.HandlerFunc(http.MethodGet, "/api/employees", middlewareInstance.RequirePermission("employees:read", a.getAllEmployeesHandler))
	router.HandlerFunc(http.MethodGet, "/api/employees/:id", middlewareInstance.RequirePermission("employees:read", a.getEmployeeByIDHandler))
	router.HandlerFunc(http.MethodDelete, "/api/employees/:id", middlewareInstance.RequirePermission("employees:write", a.deleteEmployeeHandler))
	router.HandlerFunc(http.MethodPost, "/api/employees", middlewareInstance.RequirePermission("employees:write", a.createEmployeeHandler))
	router.HandlerFunc(http.MethodPatch, "/api/employees/:id", middlewareInstance.RequirePermission("employees:write", a.updateEmployeeHandler))

	// User routes
	router.HandlerFunc(http.MethodGet, "/api/users", middlewareInstance.RequirePermission("users:read", a.GetAllUsersHandler))
	router.HandlerFunc(http.MethodGet, "/api/users/:id", middlewareInstance.RequirePermission("users:read", a.GetUserByIDHandler))
	router.HandlerFunc(http.MethodPatch, "/api/users/:id", middlewareInstance.RequirePermission("users:write", a.UpdateUserHandler))

	// NOTE: Left unprotected so users can activate their account won't be activated/authenticated yet.
	router.HandlerFunc(http.MethodPut, "/api/users/activated", a.ActivateUserHandler)

	// Accounts routes
	router.HandlerFunc(http.MethodGet, "/api/accounts", middlewareInstance.RequirePermission("accounts:read", a.getAllAccountsHandler))
	router.HandlerFunc(http.MethodPost, "/api/accounts", middlewareInstance.RequirePermission("accounts:write", a.createAccountHandler))
	router.HandlerFunc(http.MethodGet, "/api/accounts/:id", middlewareInstance.RequirePermission("accounts:read", a.getAccountByIDHandler))
	router.HandlerFunc(http.MethodPatch, "/api/accounts/:id", middlewareInstance.RequirePermission("accounts:write", a.updateAccountHandler))
	router.HandlerFunc(http.MethodDelete, "/api/accounts/:id", middlewareInstance.RequirePermission("accounts:write", a.deleteAccountHandler))

	// Authentication (Left unprotected so users can log in)
	router.HandlerFunc(http.MethodPost, "/api/authentication/token", a.createAuthenticationTokenHandler)

	// Transaction routes (Deposits and Transfers use transactions:write)
	router.HandlerFunc(http.MethodPost, "/api/deposits", middlewareInstance.RequirePermission("transactions:write", a.HandleDeposit))
	router.HandlerFunc(http.MethodPost, "/api/transfers", middlewareInstance.RequirePermission("transactions:write", a.HandleTransfer))

	// Withdrawal route
	router.HandlerFunc(http.MethodPost, "/api/withdrawals", middlewareInstance.RequirePermission("withdrawals:write", a.HandleWithdrawal))

	// Loans routes
	router.HandlerFunc(http.MethodPost, "/api/loans", middlewareInstance.RequirePermission("loans:write", a.CreateLoanHandler))
	router.HandlerFunc(http.MethodPost, "/api/loans/payments", middlewareInstance.RequirePermission("loans:write", a.CreateLoanPaymentHandler))

	// Metrics (Public for now)
	router.Handler(http.MethodGet, "/api/metrics", expvar.Handler())

	authenticateMiddleware := middlewareInstance.Authenticate(router)
	gzipRequestMiddleware := middlewareInstance.GzipRequestMiddleware(authenticateMiddleware)
	gzipResponseMiddleware := middlewareInstance.GzipResponseMiddleware(gzipRequestMiddleware)
	rateLimitMiddleware := middlewareInstance.RateLimit(gzipResponseMiddleware)
	loggingMiddleware := middlewareInstance.LoggingMiddleware(rateLimitMiddleware)
	metricsMiddleware := middlewareInstance.MetricsMiddleware(loggingMiddleware)
	panicMiddleware := middlewareInstance.RecoverPanic(metricsMiddleware)

	return middlewareInstance.EnableCORS(panicMiddleware)

}

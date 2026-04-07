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
	router.HandlerFunc(http.MethodGet, "/api/persons", middlewareInstance.RequireActivatedUser(a.getAllPersonsHandler))         // Get all persons
	router.HandlerFunc(http.MethodPost, "/api/persons", middlewareInstance.RequireActivatedUser(a.createPersonHandler))         // Create a new person
	router.HandlerFunc(http.MethodGet, "/api/persons/:ssid", middlewareInstance.RequireActivatedUser(a.getPersonBySSIDHandler)) // Get a person by SSID
	router.HandlerFunc(http.MethodPatch, "/api/persons/:ssid", middlewareInstance.RequireActivatedUser(a.updatePersonHandler))  // Update a person by SSID
	router.HandlerFunc(http.MethodDelete, "/api/persons/:ssid", middlewareInstance.RequireActivatedUser(a.deletePersonHandler))

	// Customers routes
	router.HandlerFunc(http.MethodGet, "/api/customers", middlewareInstance.RequireActivatedUser(a.getAllCustomersHandler))                          // Get all customers
	router.HandlerFunc(http.MethodPost, "/api/customers", middlewareInstance.RequireActivatedUser(a.createCustomerHandler))                          // Create a new customer
	router.HandlerFunc(http.MethodGet, "/api/customers/:id", middlewareInstance.RequireActivatedUser(a.getCustomerByIDHandler))                      // Get a customer by ID
	router.HandlerFunc(http.MethodPatch, "/api/customers/:id/kyc-status", middlewareInstance.RequireActivatedUser(a.updateCustomerKYCStatusHandler)) // Update KYC status
	router.HandlerFunc(http.MethodDelete, "/api/customers/:id", middlewareInstance.RequireActivatedUser(a.deleteCustomerHandler))                    // Delete a customer

	// Employees routes
	router.HandlerFunc(http.MethodGet, "/api/employees", middlewareInstance.RequireActivatedUser(a.getAllEmployeesHandler))       // Get all employees
	router.HandlerFunc(http.MethodGet, "/api/employees/:id", middlewareInstance.RequireActivatedUser(a.getEmployeeByIDHandler))   // Get an employee by ID
	router.HandlerFunc(http.MethodDelete, "/api/employees/:id", middlewareInstance.RequireActivatedUser(a.deleteEmployeeHandler)) // Delete an employee
	router.HandlerFunc(http.MethodPost, "/api/employees", middlewareInstance.RequireActivatedUser(a.createEmployeeHandler))       // Create a new employee
	router.HandlerFunc(http.MethodPatch, "/api/employees/:id", middlewareInstance.RequireActivatedUser(a.updateEmployeeHandler))  // Update employee status

	// User routes
	router.HandlerFunc(http.MethodGet, "/api/users", middlewareInstance.RequireActivatedUser(a.GetAllUsersHandler))      // Get all users
	router.HandlerFunc(http.MethodGet, "/api/users/:id", middlewareInstance.RequireActivatedUser(a.GetUserByIDHandler))  // Get a user by ID
	router.HandlerFunc(http.MethodPatch, "/api/users/:id", middlewareInstance.RequireActivatedUser(a.UpdateUserHandler)) // Update a user
	router.HandlerFunc(http.MethodPut, "/api/users/activated", a.ActivateUserHandler)                                    // Activate a user account

	// Accounts routes
	router.HandlerFunc(http.MethodGet, "/api/accounts", middlewareInstance.RequireActivatedUser(a.getAllAccountsHandler))       // Get all accounts
	router.HandlerFunc(http.MethodPost, "/api/accounts", middlewareInstance.RequireActivatedUser(a.createAccountHandler))       // Create a new account
	router.HandlerFunc(http.MethodGet, "/api/accounts/:id", middlewareInstance.RequireActivatedUser(a.getAccountByIDHandler))   // Get an account by ID
	router.HandlerFunc(http.MethodPatch, "/api/accounts/:id", middlewareInstance.RequireActivatedUser(a.updateAccountHandler))  // Update an account
	router.HandlerFunc(http.MethodDelete, "/api/accounts/:id", middlewareInstance.RequireActivatedUser(a.deleteAccountHandler)) // Delete an account

	router.HandlerFunc(http.MethodPost, "/api/authentication/token", a.createAuthenticationTokenHandler)

	// Deposit route
	router.HandlerFunc(http.MethodPost, "/api/deposits", middlewareInstance.RequireActivatedUser(a.HandleDeposit)) // Make a deposit

	router.HandlerFunc(http.MethodPost, "/api/withdrawals", middlewareInstance.RequireActivatedUser(a.HandleWithdrawal)) // Make a withdrawal

	// Transfer route
	router.HandlerFunc(http.MethodPost, "/api/transfers", middlewareInstance.RequireActivatedUser(a.HandleTransfer)) // Make a transfer

	router.HandlerFunc(http.MethodPost, "/api/loans", middlewareInstance.RequireActivatedUser(a.CreateLoanHandler))
	router.HandlerFunc(http.MethodPost, "/api/loans/payments", middlewareInstance.RequireActivatedUser(a.CreateLoanPaymentHandler))

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

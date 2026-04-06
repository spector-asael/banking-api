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
		Config: a.Config,
		Logger: a.Logger,
	}
	// setup a new router
	router := httprouter.New()
	// router.NotFound = http.HandlerFunc(a.notFoundResponse)
	// router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	// setup routes
	// Persons routes
	router.HandlerFunc(http.MethodGet, "/api/persons", a.getAllPersonsHandler)         // Get all persons
	router.HandlerFunc(http.MethodPost, "/api/persons", a.createPersonHandler)         // Create a new person
	router.HandlerFunc(http.MethodGet, "/api/persons/:ssid", a.getPersonBySSIDHandler) // Get a person by SSID
	router.HandlerFunc(http.MethodPatch, "/api/persons/:ssid", a.updatePersonHandler)  // Update a person by SSID
	router.HandlerFunc(http.MethodDelete, "/api/persons/:ssid", a.deletePersonHandler)

	// Customers routes
	router.HandlerFunc(http.MethodGet, "/api/customers", a.getAllCustomersHandler)                          // Get all customers
	router.HandlerFunc(http.MethodPost, "/api/customers", a.createCustomerHandler)                          // Create a new customer
	router.HandlerFunc(http.MethodGet, "/api/customers/:id", a.getCustomerByIDHandler)                      // Get a customer by ID
	router.HandlerFunc(http.MethodPatch, "/api/customers/:id/kyc-status", a.updateCustomerKYCStatusHandler) // Update KYC status
	router.HandlerFunc(http.MethodDelete, "/api/customers/:id", a.deleteCustomerHandler)                    // Delete a customer

	// Employees routes
	router.HandlerFunc(http.MethodGet, "/api/employees", a.getAllEmployeesHandler)       // Get all employees
	router.HandlerFunc(http.MethodGet, "/api/employees/:id", a.getEmployeeByIDHandler)   // Get an employee by ID
	router.HandlerFunc(http.MethodDelete, "/api/employees/:id", a.deleteEmployeeHandler) // Delete an employee
	router.HandlerFunc(http.MethodPost, "/api/employees", a.createEmployeeHandler)       // Create a new employee
	router.HandlerFunc(http.MethodPatch, "/api/employees/:id", a.updateEmployeeHandler)  // Update employee status

	// User routes
	router.HandlerFunc(http.MethodGet, "/api/users", a.GetAllUsersHandler)      // Get all users
	router.HandlerFunc(http.MethodGet, "/api/users/:id", a.GetUserByIDHandler)  // Get a user by ID
	router.HandlerFunc(http.MethodPatch, "/api/users/:id", a.UpdateUserHandler) // Update a user

	// Accounts routes
	router.HandlerFunc(http.MethodGet, "/api/accounts", a.getAllAccountsHandler)       // Get all accounts
	router.HandlerFunc(http.MethodPost, "/api/accounts", a.createAccountHandler)       // Create a new account
	router.HandlerFunc(http.MethodGet, "/api/accounts/:id", a.getAccountByIDHandler)   // Get an account by ID
	router.HandlerFunc(http.MethodPatch, "/api/accounts/:id", a.updateAccountHandler)  // Update an account
	router.HandlerFunc(http.MethodDelete, "/api/accounts/:id", a.deleteAccountHandler) // Delete an account

	// Deposit route
	router.HandlerFunc(http.MethodPost, "/api/deposits", a.HandleDeposit) // Make a deposit

	router.HandlerFunc(http.MethodPost, "/api/withdrawals", a.HandleWithdrawal) // Make a withdrawal

	// Transfer route
	router.HandlerFunc(http.MethodPost, "/api/transfers", a.HandleTransfer) // Make a transfer

	router.HandlerFunc(http.MethodPost, "/api/loans", a.CreateLoanHandler)
	router.HandlerFunc(http.MethodPost, "/api/loans/payments", a.CreateLoanPaymentHandler)

	router.Handler(http.MethodGet, "/api/metrics", expvar.Handler())

	gzipRequestMiddleware := middlewareInstance.GzipRequestMiddleware(router)
	gzipResponseMiddleware := middlewareInstance.GzipResponseMiddleware(gzipRequestMiddleware)
	rateLimitMiddleware := middlewareInstance.RateLimit(gzipResponseMiddleware)
	loggingMiddleware := middlewareInstance.LoggingMiddleware(rateLimitMiddleware)
	metricsMiddleware := middlewareInstance.MetricsMiddleware(loggingMiddleware)
	panicMiddleware := middlewareInstance.RecoverPanic(metricsMiddleware)
	// CORS should be the true outermost middleware
	return middlewareInstance.EnableCORS(panicMiddleware)

}
